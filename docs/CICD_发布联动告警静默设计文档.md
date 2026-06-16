# CI/CD 发布联动告警静默架构设计文档

> 当 CI/CD 流水线进入部署阶段时，自动创建告警静默规则，避免滚动更新期间产生的**误告警轰炸**（如 Pod 重启、CPU 飙升、副本数不足等），部署完成后自动解除静默。

---

## 一、核心问题与动机

### 1.1 问题场景

在生产环境中，**每次发布都会触发一系列预期内的告警**：

| 发布阶段 | 预期内的告警 | 影响 |
|---------|------------|------|
| 镜像更新触发 | `Pod频繁重启`（旧 Pod 被 Terminate） | 误告警 |
| 新 Pod 启动中 | `Pod处于非就绪状态`、`Deployment副本数不足` | 误告警 |
| 流量切换中 | `API延迟过高`（短暂 P95 飙升） | 误告警 |
| 旧 Pod 优雅退出 | `K8s Job失败`（preStop hook 超时） | 误告警 |

这些告警在 **10-30 秒的滚动更新窗口内** 是完全正常的，但如果不静默，运维群会被瞬间轰炸 5-10 条无意义消息。

### 1.2 解决方案

```
┌──────────────────────────────────────────────────────────────────────────┐
│                       CI/CD + 告警静默联动                                 │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   流水线 Deploy 阶段开始                                                   │
│         │                                                                │
│         ▼                                                                │
│   ┌─────────────────────┐    ┌──────────────────────────┐               │
│   │ 自动创建静默规则     │───▶│ 匹配条件：                │               │
│   │ (namespace+workload) │    │  namespace = xxx         │               │
│   │                     │    │  deployment = yyy        │               │
│   └─────────────────────┘    │  severity =~ warning|..  │               │
│         │                    └──────────────────────────┘               │
│         ▼                                                                │
│   ┌─────────────────────┐                                               │
│   │ 滚动更新执行中...    │  ← 此期间所有匹配告警被静默                      │
│   │ (Pod 重启/切换)      │                                               │
│   └─────────────────────┘                                               │
│         │                                                                │
│         ▼                                                                │
│   ┌─────────────────────┐                                               │
│   │ Deploy 完成/失败     │                                               │
│   └─────────────────────┘                                               │
│         │                                                                │
│         ├── 成功：等待缓冲期（默认 2 分钟）后自动解除静默                     │
│         └── 失败：立即解除静默（让真正的问题告警及时触达）                     │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 二、整体架构

### 2.1 模块交互图

```
┌──────────────┐         ┌──────────────────┐         ┌────────────────────┐
│  CI/CD 模块   │────────▶│  静默联动服务     │────────▶│  告警静默规则       │
│  (Pipeline)  │  事件    │ (DeploySilence)  │  CRUD   │ (MonitorSilence)   │
└──────────────┘         └──────────────────┘         └────────────────────┘
       │                        │                            │
       │ Deploy阶段开始          │ CreateSilence              │ 告警评估引擎
       │ Deploy阶段完成          │ DeleteSilence              │ 匹配静默规则
       ▼                        ▼                            ▼
┌──────────────┐         ┌──────────────────┐         ┌────────────────────┐
│  Stage 状态机 │         │  审计日志记录     │         │  AlertEvalWorker   │
│  deploying   │         │  谁/何时/静默了   │         │  跳过匹配的告警     │
│  → success   │         │  什么             │         │                    │
│  → failed    │         └──────────────────┘         └────────────────────┘
└──────────────┘
```

### 2.2 数据流

```
                    Pipeline Deploy 触发
                           │
              ┌────────────┼────────────────┐
              │            │                │
              ▼            ▼                ▼
     ①创建静默规则    ②执行镜像更新    ③记录审计日志
    (namespace匹配)   (Rollout)       (deploy_silence)
              │            │                │
              │            ▼                │
              │     等待 Rollout 完成        │
              │            │                │
              │     ┌──────┴──────┐         │
              │     │             │         │
              │   成功           失败        │
              │     │             │         │
              ▼     ▼             ▼         ▼
    ④缓冲期后    ⑤解除静默     ⑤立即解除   ⑥更新审计
    自动解除                    静默
```

---

## 三、详细设计

### 3.1 流水线模型扩展

在 `CicdPipeline` 模型中新增配置字段：

```go
// CicdPipeline 新增字段
type CicdPipeline struct {
    // ... 现有字段 ...
    
    // 发布联动告警静默
    EnableDeploySilence  bool   `gorm:"column:enable_deploy_silence;default:0" json:"enable_deploy_silence"`   // 是否启用发布静默
    SilenceBufferMinutes int    `gorm:"column:silence_buffer_minutes;default:2" json:"silence_buffer_minutes"` // 部署完成后缓冲时间(分钟)
    SilenceSeverities    string `gorm:"column:silence_severities;default:'warning'" json:"silence_severities"` // 静默的告警级别(逗号分隔)
}
```

**配置说明：**

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `enable_deploy_silence` | false | 是否启用部署静默联动（生产环境建议开启） |
| `silence_buffer_minutes` | 2 | 部署成功后保持静默的缓冲时间（等待新 Pod 稳定） |
| `silence_severities` | "warning" | 静默哪些级别的告警（critical 建议不静默） |

### 3.2 静默联动服务

```go
// DeploySilenceService 发布联动静默服务
type DeploySilenceService struct {
    dao *dao.Dao
}

// CreateDeploySilence 部署开始时创建静默规则
func (s *DeploySilenceService) CreateDeploySilence(ctx context.Context, pipeline *CicdPipeline, runID int64) (int64, error) {
    // 构建匹配条件：namespace + workload name
    matchers := []map[string]string{
        {"label": "namespace", "op": "=", "value": pipeline.TargetNamespace},
        {"label": "deployment", "op": "=", "value": pipeline.TargetWorkloadName},
    }
    // 可选：限定静默级别
    if pipeline.SilenceSeverities != "" {
        matchers = append(matchers, map[string]string{
            "label": "severity", "op": "=~", "value": pipeline.SilenceSeverities,
        })
    }
    
    matchersJSON, _ := json.Marshal(matchers)
    
    rule := &MonitorSilenceRule{
        Name:     fmt.Sprintf("[自动] %s 发布静默 #%d", pipeline.Name, runID),
        Type:     "silence",
        Matchers: string(matchersJSON),
        Duration: fmt.Sprintf("%dm", 15 + pipeline.SilenceBufferMinutes), // 最大15分钟发布 + 缓冲
        Comment:  fmt.Sprintf("CI/CD 流水线 %s 正在部署，自动静默匹配告警", pipeline.Name),
        Enabled:  true,
        CreatedBy: pipeline.CreatedUserID,
    }
    
    return s.dao.SilenceRuleCreate(ctx, rule)
}

// RemoveDeploySilence 部署完成时移除静默规则
func (s *DeploySilenceService) RemoveDeploySilence(ctx context.Context, silenceID int64, success bool, bufferMinutes int) {
    if !success {
        // 部署失败：立即删除静默，让真正的问题告警触达
        s.dao.SilenceRuleDelete(ctx, silenceID)
        return
    }
    // 部署成功：延迟 buffer 时间后删除（等待新 Pod 稳定）
    go func() {
        time.Sleep(time.Duration(bufferMinutes) * time.Minute)
        s.dao.SilenceRuleDelete(context.Background(), silenceID)
    }()
}
```

### 3.3 集成点：部署阶段执行流程

在现有的 `executeDeploy` 函数中增加联动逻辑：

```go
func (s *StageService) executeDeploy(ctx context.Context, stage *CicdPipelineStage, pipeline *CicdPipeline) error {
    var silenceID int64
    
    // ① 部署前：自动创建静默规则
    if pipeline.EnableDeploySilence {
        silenceSvc := NewDeploySilenceService()
        sid, err := silenceSvc.CreateDeploySilence(ctx, pipeline, stage.RunID)
        if err != nil {
            // 静默创建失败不阻塞部署
            global.Logger.Warn("[部署静默] 创建静默规则失败，继续部署", zap.Error(err))
        } else {
            silenceID = sid
            global.Logger.Info("[部署静默] 已创建静默规则",
                zap.Int64("silence_id", sid),
                zap.String("pipeline", pipeline.Name),
                zap.String("namespace", pipeline.TargetNamespace),
            )
        }
    }
    
    // ② 执行部署（现有逻辑不变）
    err := s.doRollingUpdate(ctx, stage, pipeline)
    
    // ③ 部署后：解除静默
    if silenceID > 0 {
        silenceSvc := NewDeploySilenceService()
        silenceSvc.RemoveDeploySilence(ctx, silenceID, err == nil, pipeline.SilenceBufferMinutes)
    }
    
    return err
}
```

### 3.4 告警评估引擎匹配

现有 `AlertEvalWorker` 已支持静默规则匹配（在 `sendNotification` 中检查 `MonitorSilenceRule`），无需额外修改。联动创建的静默规则会自动被评估引擎识别。

---

## 四、静默规则匹配策略

### 4.1 匹配条件生成

根据流水线的部署配置，自动生成精确的匹配条件：

```
┌────────────────────────────────────────────────────────────────┐
│              静默匹配条件生成规则                                 │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  Pipeline.TargetNamespace     → namespace = "production"       │
│  Pipeline.TargetWorkloadName  → deployment = "my-service"      │
│  Pipeline.SilenceSeverities   → severity =~ "warning|info"     │
│                                                                │
│  组合结果（AND 逻辑）：                                          │
│  [                                                             │
│    {"label":"namespace", "op":"=", "value":"production"},      │
│    {"label":"deployment","op":"=~","value":"my-service.*"},    │
│    {"label":"severity",  "op":"=~","value":"warning"}          │
│  ]                                                             │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### 4.2 不同场景的匹配策略

| 场景 | 匹配条件 | 静默范围 |
|------|---------|---------|
| 单服务部署 | namespace + deployment | 仅该服务的告警 |
| 命名空间级发布 | namespace only | 该命名空间所有告警 |
| 全量发布 | severity=warning | 所有 warning 告警 |
| 精确控制 | namespace + deployment + severity | 仅该服务的指定级别 |

### 4.3 安全策略

```
                    安全规则（硬性约束）
    ┌──────────────────────────────────────────────┐
    │                                              │
    │  1. critical 告警默认不静默                    │
    │     （可通过 silence_severities 配置覆盖）      │
    │                                              │
    │  2. 静默最长时限 = 30 分钟                     │
    │     （超时自动失效，防止遗忘）                   │
    │                                              │
    │  3. 部署失败 → 立即解除                        │
    │     （让真正的问题告警尽快触达）                  │
    │                                              │
    │  4. 静默规则带 [自动] 标记                     │
    │     （审计可追溯，区分手动 vs 自动）             │
    │                                              │
    └──────────────────────────────────────────────┘
```

---

## 五、API 接口设计

### 5.1 流水线配置（复用现有接口）

在创建/更新流水线时，传入新增字段：

```bash
# 更新流水线，开启部署静默
curl -X PUT http://<地址>:8080/api/v1/cicd/pipeline/5 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enable_deploy_silence": true,
    "silence_buffer_minutes": 3,
    "silence_severities": "warning,info"
  }'
```

### 5.2 手动创建发布静默（可选）

对于不走 CI/CD 的手动发布场景，提供独立 API：

```bash
# 手动触发部署静默
POST /api/v1/monitoring/deploy-silence

{
    "namespace": "production",
    "workload_name": "my-service",
    "workload_kind": "Deployment",
    "duration": "10m",
    "severities": "warning",
    "comment": "手动发布 v2.1.0，预计 5 分钟"
}
```

### 5.3 查看活跃的部署静默

```bash
# 列出所有由 CI/CD 自动创建的静默规则
GET /api/v1/monitoring/deploy-silence?active=true

# 响应
{
    "code": 0,
    "data": {
        "items": [
            {
                "id": 15,
                "name": "[自动] my-pipeline 发布静默 #42",
                "namespace": "production",
                "workload": "my-service",
                "created_at": "2025-05-27T14:30:00Z",
                "expires_at": "2025-05-27T14:47:00Z",
                "pipeline_id": 5,
                "run_id": 42,
                "status": "active"
            }
        ]
    }
}
```

### 5.4 提前解除静默

```bash
# 手动提前解除（如确认部署已稳定）
DELETE /api/v1/monitoring/deploy-silence/15
```

---

## 六、时序图

```
┌────────┐     ┌──────────┐     ┌──────────────┐     ┌───────────────┐
│ Jenkins │     │ Platform │     │ DeploySilence│     │ AlertEvalWorker│
│ 回调    │     │ 部署服务  │     │ 联动服务     │     │ 告警评估引擎   │
└───┬────┘     └────┬─────┘     └──────┬───────┘     └───────┬───────┘
    │                │                   │                     │
    │ 构建成功回调    │                   │                     │
    │───────────────▶│                   │                     │
    │                │                   │                     │
    │                │ ①创建静默规则      │                     │
    │                │──────────────────▶│                     │
    │                │                   │ 写入 DB              │
    │                │    返回 silenceID  │                     │
    │                │◀──────────────────│                     │
    │                │                   │                     │
    │                │ ②执行镜像更新      │                     │
    │                │──────────────────────────────────────────┤
    │                │                   │                     │
    │                │                   │    ③评估告警时        │
    │                │                   │    检测到匹配的静默   │
    │                │                   │    → 跳过通知发送     │
    │                │                   │                     │
    │                │ Rollout 完成       │                     │
    │                │                   │                     │
    │                │ ④删除静默规则       │                     │
    │                │──────────────────▶│                     │
    │                │                   │ 从 DB 删除/标记过期   │
    │                │                   │                     │
    │                │                   │    ⑤后续告警正常      │
    │                │                   │    评估并通知        │
    │                │                   │                     │
```

---

## 七、配置指南

### 7.1 流水线创建时开启（推荐）

在前端创建流水线向导的「部署配置」步骤中，新增开关：

```
┌─────────────────────────────────────────────────────┐
│            部署配置                                   │
├─────────────────────────────────────────────────────┤
│                                                     │
│  目标集群:    [生产集群 ▼]                            │
│  命名空间:    [production]                           │
│  工作负载:    [Deployment / my-service]              │
│  容器名称:    [app]                                  │
│                                                     │
│  ── 高级选项 ──                                      │
│                                                     │
│  ☑ 发布时自动静默告警                                 │
│     静默级别:  [warning ▼] (critical 建议不静默)      │
│     缓冲时间:  [2] 分钟 (部署完成后保持静默的时间)     │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### 7.2 生产环境推荐配置

| 环境 | `enable_deploy_silence` | `silence_severities` | `silence_buffer_minutes` |
|------|------------------------|---------------------|--------------------------|
| 开发环境 | false | - | - |
| 测试环境 | false | - | - |
| 预发环境 | true | "warning,info" | 1 |
| **生产环境** | **true** | **"warning"** | **3** |

> **生产建议**：仅静默 warning 级别，critical 告警始终保持通知（如镜像拉取失败、OOM 等真正的部署问题）。

### 7.3 手动发布场景

对于不走 CI/CD 的 kubectl 手动发布，可以提前手动创建静默：

```bash
# 发布前创建静默（10分钟有效期）
curl -X POST http://<地址>:8080/api/v1/monitoring/deploy-silence \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "production",
    "workload_name": "my-service",
    "duration": "10m",
    "severities": "warning",
    "comment": "手动发布 v2.1.0"
  }'

# 发布操作...
kubectl set image deployment/my-service app=registry/my-service:v2.1.0 -n production

# 确认稳定后手动解除（或等待自动过期）
curl -X DELETE http://<地址>:8080/api/v1/monitoring/deploy-silence/15 \
  -H "Authorization: Bearer $TOKEN"
```

---

## 八、与现有系统集成点

### 8.1 与通知路由策略的关系

```
告警评估引擎处理顺序：
  ① 评估 PromQL → 产生告警事件
  ② 检查静默规则 → 如果命中静默，跳过通知 ← 发布联动在这里生效
  ③ 检查抑制规则 → 高优先级压制低优先级
  ④ 匹配路由策略 → 确定发到哪个群
  ⑤ 发送通知
```

### 8.2 与审计日志的关系

所有自动创建/删除静默的操作都记录到审计日志：

```json
{
    "action": "deploy_silence_create",
    "operator": "system/cicd",
    "pipeline_id": 5,
    "pipeline_name": "my-pipeline",
    "run_id": 42,
    "silence_id": 15,
    "namespace": "production",
    "workload": "my-service",
    "duration": "17m",
    "timestamp": "2025-05-27T14:30:00Z"
}
```

### 8.3 与钉钉通知的关系

部署通知消息中会提示静默状态：

```markdown
### ⏳ 部署进行中

**流水线**: my-pipeline
**环境**: 🚀 生产环境
**命名空间**: production
**工作负载**: Deployment/my-service
**镜像**: registry/my-service:v2.1.0

🔇 **告警静默已启用** — 部署期间 warning 级别告警将被自动屏蔽
```

---

## 九、异常处理与边界情况

| 异常场景 | 处理策略 |
|---------|---------|
| 静默创建失败 | 不阻塞部署，记录 Warn 日志，告警正常发送 |
| 部署超时（>15min） | 静默规则设置 max duration=30min，超时自动失效 |
| 平台重启 | 带 duration 的静默规则按时间自动失效，无需额外处理 |
| 并行部署同一服务 | 每次部署创建独立静默规则，后完成的先删除自己的 |
| 回滚操作 | 回滚也是镜像更新，同样触发联动静默 |
| 手动强制解除 | 提供 DELETE API，运维可随时手动解除 |

---

## 十、API 速查表

| 功能 | 方法 | 路径 |
|------|------|------|
| 手动创建部署静默 | POST | `/api/v1/monitoring/deploy-silence` |
| 查看活跃静默列表 | GET | `/api/v1/monitoring/deploy-silence?active=true` |
| 手动解除静默 | DELETE | `/api/v1/monitoring/deploy-silence/:id` |
| 流水线开启联动 | PUT | `/api/v1/cicd/pipeline/:id`（新增字段） |

---

## 十一、总结

### 架构优势

- **零侵入**：复用现有 `MonitorSilenceRule` 模型，无需新建表
- **自动化**：流水线部署自动触发，无需人工介入
- **安全兜底**：最长 30 分钟自动失效 + critical 不静默
- **可追溯**：所有操作记录审计日志 + 钉钉通知提示静默状态
- **灵活配置**：按流水线粒度开关，支持手动发布场景

### 与现有系统的完美衔接

```
CI/CD 14阶段闭环
     └── Deploy 阶段
          └── 联动告警静默（本设计）
               └── 复用 MonitorSilenceRule
                    └── AlertEvalWorker 自动匹配
                         └── 多群路由策略正常工作（静默期后）
```
