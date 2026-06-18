# CI/CD 发布联动告警静默 — 完整实现架构与验证手册

> 本文档覆盖**架构设计全景、代码实现细节、数据库变更、API 接口、前端交互**及**端到端功能验证操作手册**。

---

## 一、功能概述

### 1.1 解决的问题

每次发布（镜像滚动更新）期间，K8s 会短暂出现以下**预期内**的监控告警：

| 阶段 | 告警类型 | 实际情况 |
|------|---------|---------|
| 旧 Pod 被终止 | Pod 频繁重启 | 正常滚动更新 |
| 新 Pod 启动中 | Pod 处于非就绪状态 | Ready 探针未通过 |
| 副本数波动 | Deployment 副本数不足 | 新旧切换中 |
| 流量切换 | API P95 延迟飙升 | 短暂连接中断 |

**如果不做处理**：每次发布运维群会被 5-10 条无意义告警轰炸。

### 1.2 解决方案

```
部署开始 → 自动创建精确匹配的告警静默规则
   ↓
滚动更新执行中... (所有匹配告警被自动屏蔽)
   ↓
部署完成
   ├── 成功 → 保留 2 分钟缓冲后自动解除
   └── 失败 → 立即解除静默（让真正问题暴露）
```

---

## 二、整体架构

### 2.1 模块关系图

```
┌──────────────────┐         ┌──────────────────────────┐
│   前端 (Vue 3)    │         │    后端 (Go + Gin)        │
│   PipelineCreate │────────▶│    Pipeline API           │
│   配置静默参数     │  HTTP   │    (Create/Update)       │
└──────────────────┘         └─────────┬────────────────┘
                                       │ 保存到 DB
                                       ▼
┌──────────────────┐         ┌──────────────────────────┐
│  AlertEvalWorker │◀────────│    CicdPipeline 表       │
│  告警评估引擎     │  查询    │    (含静默配置字段)       │
│  匹配静默规则     │  静默    └─────────┬────────────────┘
└──────────────────┘          规则表     │
        ▲                               │ 触发部署
        │                               ▼
┌──────────────────┐         ┌──────────────────────────┐
│ MonitorSilence   │◀────────│ DeploySilenceService     │
│ Rule 表          │  CRUD   │ 静默生命周期管理          │
│ (复用现有模型)    │         │ - CreateDeploySilence    │
└──────────────────┘         │ - ReleaseDeploySilence   │
                             └──────────────────────────┘
```

### 2.2 核心设计原则

| 原则 | 说明 |
|------|------|
| **零新建表** | 复用现有 `monitor_silence_rules` 表 |
| **零阻塞** | 静默创建失败不影响部署流程 |
| **自动兜底** | 最长 30 分钟自动失效 |
| **精准匹配** | namespace + workload + severity 三维精确匹配 |
| **critical 保护** | 默认不静默 critical 级别告警 |

---

## 三、数据库变更

### 3.1 CicdPipeline 表新增字段

GORM AutoMigrate 会自动添加以下字段，也可手动执行 SQL：

```sql
ALTER TABLE cicd_pipeline 
ADD COLUMN enable_deploy_silence TINYINT(1) DEFAULT 0 COMMENT '是否启用发布联动静默',
ADD COLUMN silence_buffer_minutes INT DEFAULT 10 COMMENT '静默缓冲时间(分钟)',
ADD COLUMN silence_severities VARCHAR(100) DEFAULT 'warning,info' COMMENT '静默的告警级别';
```

### 3.2 字段说明

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `enable_deploy_silence` | bool | false | 是否开启发布静默 |
| `silence_buffer_minutes` | int | 10 | 部署超时后额外缓冲时间 |
| `silence_severities` | string | "warning,info" | 要静默的告警级别（逗号分隔） |

### 3.3 静默规则存储（复用 monitor_silence_rules 表）

自动创建的静默规则示例：

```json
{
  "name": "[发布静默] my-pipeline - production/my-service",
  "type": "silence",
  "matchers": "[{\"label\":\"namespace\",\"op\":\"=\",\"value\":\"production\"},{\"label\":\"workload\",\"op\":\"=~\",\"value\":\"my-service\"},{\"label\":\"severity\",\"op\":\"=~\",\"value\":\"warning|info\"}]",
  "starts_at": 1748332800,
  "ends_at": 1748333700,
  "duration": "15m",
  "enabled": true,
  "comment": "CI/CD 发布自动静默 | 流水线: my-pipeline | 目标: production/my-service | 预计 15 分钟后自动解除"
}
```

---

## 四、后端代码实现详解

### 4.1 文件清单

| 文件路径 | 修改类型 | 说明 |
|---------|---------|------|
| `internal/app/models/cicd_pipeline.go` | 修改 | 新增 3 个配置字段 |
| `internal/app/services/cicd_deploy_silence.go` | **新建** | 静默生命周期管理核心服务 |
| `internal/app/services/cicd_stage.go` | 修改 | `executeDeployAsync` 集成静默 |
| `internal/app/services/cicd_pipeline.go` | 修改 | `executeLegacyDeployAsync` 集成 + 字段映射 |
| `internal/app/requests/cicd_pipeline.go` | 修改 | 请求结构体新增字段 |
| `internal/app/controllers/api/v1/cicd/pipeline_controller.go` | 修改 | 新增 DeploySilenceStatus |
| `internal/app/routers/kube_cicd/cicd_router.go` | 修改 | 注册路由 |

### 4.2 核心服务 — cicd_deploy_silence.go

```go
// 创建发布静默规则
func (s *Services) CreateDeploySilence(ctx, pipeline, namespace, workloadName) (*DeploySilenceInfo, error)

// 释放静默规则
func (s *Services) ReleaseDeploySilence(ctx, silenceRuleID int64, success bool)

// 查询活跃静默
func (s *Services) GetActiveDeploySilences(ctx, pipelineID int64) ([]MonitorSilenceRule, error)

// 清理过期静默
func (s *Services) CleanExpiredDeploySilences(ctx) (int64, error)
```

**静默时长计算公式**：
```
totalMinutes = min(5 + buffer_minutes, 30)
```
- 5 分钟 = Deployment Rollout 超时时间
- buffer_minutes = 用户配置的缓冲时间
- 30 分钟 = 绝对上限（安全兜底）

### 4.3 部署流程集成点

两条部署路径都已集成：

**路径 A: 阶段化部署 (executeDeployAsync)**
```
cicd_stage.go → executeDeployAsync
    ├── 部署前: CreateDeploySilence()
    ├── 部署失败: ReleaseDeploySilence(false)  → 立即解除
    └── 部署成功: ReleaseDeploySilence(true)   → 延续 2 分钟
```

**路径 B: 旧配置部署 (executeLegacyDeployAsync)**
```
cicd_pipeline.go → executeLegacyDeployAsync
    ├── 部署前: CreateDeploySilence()
    ├── 部署失败: ReleaseDeploySilence(false)
    └── 部署成功: ReleaseDeploySilence(true)
```

### 4.4 请求/响应结构

**创建流水线 - 新增字段：**
```go
type PipelineCreateRequest struct {
    // ... 现有字段 ...
    EnableDeploySilence  bool   `json:"enable_deploy_silence"`
    SilenceBufferMinutes int    `json:"silence_buffer_minutes"`
    SilenceSeverities    string `json:"silence_severities"`
}
```

**更新流水线 - 指针类型（支持 partial update）：**
```go
type PipelineUpdateRequest struct {
    // ... 现有字段 ...
    EnableDeploySilence  *bool   `json:"enable_deploy_silence"`
    SilenceBufferMinutes *int    `json:"silence_buffer_minutes"`
    SilenceSeverities    *string `json:"silence_severities"`
}
```

---

## 五、API 接口

### 5.1 配置接口（复用流水线 CRUD）

#### 创建流水线时开启静默

```bash
POST /api/v1/k8s/cicd/pipeline/create
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "my-service-pipeline",
  "git_repo": "https://git.example.com/team/my-service.git",
  "git_branch": "main",
  "auto_deploy": true,
  "target_cluster_id": 1,
  "target_namespace": "production",
  "target_workload_kind": "Deployment",
  "target_workload_name": "my-service",
  "target_container": "app",
  "enable_deploy_silence": true,
  "silence_buffer_minutes": 10,
  "silence_severities": "warning,info"
}
```

#### 更新流水线静默配置

```bash
POST /api/v1/k8s/cicd/pipeline/update
Content-Type: application/json
Authorization: Bearer <token>

{
  "id": 5,
  "enable_deploy_silence": true,
  "silence_buffer_minutes": 5,
  "silence_severities": "warning"
}
```

### 5.2 状态查询接口

#### 获取流水线发布静默状态

```bash
GET /api/v1/k8s/cicd/pipeline/deploy-silence-status?pipeline_id=5
Authorization: Bearer <token>
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "active": true,
    "count": 1,
    "rules": [
      {
        "id": 15,
        "name": "[发布静默] my-pipeline - production/my-service",
        "type": "silence",
        "matchers": "[...]",
        "starts_at": 1748332800,
        "ends_at": 1748333700,
        "enabled": true,
        "comment": "CI/CD 发布自动静默 | 流水线: my-pipeline | ..."
      }
    ]
  }
}
```

---

## 六、前端实现

### 6.1 界面位置

在 `PipelineCreate.vue` 的「自动部署配置」区块中，位于 `auto_deploy` 开关之后：

```
部署配置区块
├── 自动部署开关
├── ★ 发布静默告警开关 (新增)
│     ├── 静默缓冲时间 输入框 (1-25 分钟)
│     └── 静默告警级别 下拉选择
├── 目标集群
├── 目标命名空间
├── 工作负载类型
├── 工作负载名称
└── 容器名称
```

### 6.2 UI 交互

| 控件 | 类型 | 说明 |
|------|------|------|
| 发布静默告警 | Toggle 开关 | 带「推荐开启」标签 |
| 静默缓冲时间 | Number Input | 范围 1-25，默认 10 |
| 静默告警级别 | Select | 4 个选项（见下方） |

**告警级别选项：**
- `warning,info`：warning + info（推荐）
- `warning`：仅 warning
- `info`：仅 info
- `critical,warning,info`：全部（含 critical）

### 6.3 数据流

```
前端 pipelineData → API 请求 → 后端 Request 结构体 → GORM 写入 DB
                                                           ↓
部署执行时读取 Pipeline 配置 → 判断 enable_deploy_silence → 创建静默规则
```

---

## 七、端到端验证操作手册

### 7.1 前置条件

- [x] 后端代码编译通过 (`go build ./...`)
- [ ] 数据库已执行字段变更（GORM AutoMigrate 或手动 SQL）
- [ ] 至少有一个可用集群和部署目标
- [ ] 监控告警规则已配置（有 namespace/severity 标签）

### 7.2 验证步骤

---

#### 测试 1：数据库字段验证

**目的**：确认新字段已正确添加到数据库

```sql
-- 查看表结构
DESCRIBE cicd_pipeline;

-- 验证字段存在
SELECT column_name, column_type, column_default 
FROM information_schema.columns 
WHERE table_name = 'cicd_pipeline' 
AND column_name IN ('enable_deploy_silence', 'silence_buffer_minutes', 'silence_severities');
```

**期望结果**：
```
+------------------------+-------------+-----------+
| column_name            | column_type | default   |
+------------------------+-------------+-----------+
| enable_deploy_silence  | tinyint(1)  | 0         |
| silence_buffer_minutes | int         | 10        |
| silence_severities     | varchar(100)| warning,info |
+------------------------+-------------+-----------+
```

---

#### 测试 2：API 创建流水线（含静默配置）

**目的**：确认 API 正确接收和保存静默配置

```powershell
# PowerShell 请求
$headers = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer <your_token>"
}

$body = @{
    name = "silence-test-pipeline"
    description = "测试发布静默功能"
    git_repo = "https://github.com/test/repo.git"
    git_branch = "main"
    jenkins_job = "k8s-builder-go"
    language_type = "go"
    auto_deploy = $true
    target_cluster_id = 1
    target_namespace = "default"
    target_workload_kind = "Deployment"
    target_workload_name = "test-app"
    target_container = "app"
    enable_deploy_silence = $true
    silence_buffer_minutes = 5
    silence_severities = "warning,info"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/k8s/cicd/pipeline/create" `
    -Method POST -Headers $headers -Body $body
```

**期望结果**：
```json
{"code": 0, "data": {"pipeline_id": <ID>}}
```

**数据库验证**：
```sql
SELECT id, name, enable_deploy_silence, silence_buffer_minutes, silence_severities 
FROM cicd_pipeline WHERE name = 'silence-test-pipeline';
```

---

#### 测试 3：API 更新流水线静默配置

**目的**：确认 partial update 工作正常

```powershell
$body = @{
    id = <pipeline_id>
    enable_deploy_silence = $true
    silence_buffer_minutes = 8
    silence_severities = "warning"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/k8s/cicd/pipeline/update" `
    -Method POST -Headers $headers -Body $body
```

**期望结果**：`silence_buffer_minutes` 更新为 8，`silence_severities` 更新为 "warning"

---

#### 测试 4：触发部署 → 验证静默规则自动创建

**目的**：核心功能验证 — 部署开始时自动创建静默规则

**操作步骤**：
1. 确保流水线 `enable_deploy_silence = true`
2. 触发流水线运行（通过 API 或前端点击「运行」）
3. 在部署阶段执行期间，检查静默规则表

```powershell
# 触发流水线
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/k8s/cicd/pipeline/trigger" `
    -Method POST -Headers $headers `
    -Body (@{ id = <pipeline_id> } | ConvertTo-Json)
```

**验证静默规则已创建**：
```sql
SELECT id, name, type, matchers, starts_at, ends_at, enabled, comment
FROM monitor_silence_rules 
WHERE name LIKE '[发布静默]%' 
ORDER BY id DESC LIMIT 5;
```

**期望结果**：
- `name` 包含 `[发布静默]` 前缀
- `matchers` 包含正确的 namespace 和 workload 匹配
- `enabled = 1`
- `ends_at - starts_at` ≈ (5 + buffer_minutes) × 60

---

#### 测试 5：查询发布静默状态 API

**目的**：验证状态查询接口

```powershell
# 部署进行中时查询
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/k8s/cicd/pipeline/deploy-silence-status?pipeline_id=<ID>" `
    -Method GET -Headers $headers
```

**期望结果（部署中）**：
```json
{
  "code": 0,
  "data": {
    "active": true,
    "count": 1,
    "rules": [...]
  }
}
```

**期望结果（无部署）**：
```json
{
  "code": 0,
  "data": {
    "active": false,
    "count": 0,
    "rules": []
  }
}
```

---

#### 测试 6：部署成功 → 验证静默延续 2 分钟后解除

**目的**：验证成功路径的静默释放逻辑

**操作**：等待部署成功完成后检查静默规则

```sql
-- 部署成功后查看规则变化
SELECT id, name, ends_at, comment, enabled
FROM monitor_silence_rules 
WHERE name LIKE '[发布静默]%' 
ORDER BY id DESC LIMIT 1;
```

**期望结果**：
- `ends_at` 被更新为「当前时间 + 2 分钟」
- `comment` 变为 "部署成功，静默将于 HH:MM:SS 自动解除"
- `enabled` 仍为 1（2 分钟后自然失效）

---

#### 测试 7：部署失败 → 验证静默立即解除

**目的**：验证失败路径的静默释放逻辑

**模拟失败方式**：
- 配置一个不存在的容器名称
- 或配置一个无法拉取的镜像

```sql
-- 强制让部署失败（更新目标容器为不存在的名称）
UPDATE cicd_pipeline SET target_container = 'nonexistent-container' WHERE id = <ID>;
```

触发部署后检查：

```sql
SELECT id, name, ends_at, comment, enabled
FROM monitor_silence_rules 
WHERE name LIKE '[发布静默]%' 
ORDER BY id DESC LIMIT 1;
```

**期望结果**：
- `enabled = 0`（立即禁用）
- `ends_at` 被更新为当前时间
- `comment` = "部署失败，静默已立即解除"

---

#### 测试 8：静默未开启时不创建规则

**目的**：验证开关关闭时无副作用

```sql
UPDATE cicd_pipeline SET enable_deploy_silence = 0 WHERE id = <ID>;
```

触发部署后检查：

```sql
-- 应该没有新规则产生
SELECT COUNT(*) FROM monitor_silence_rules 
WHERE name LIKE '[发布静默]%' 
AND starts_at > UNIX_TIMESTAMP(NOW() - INTERVAL 5 MINUTE);
```

**期望结果**：count = 0

---

#### 测试 9：静默创建失败不阻塞部署

**目的**：验证容错性

**模拟方法**：临时断开数据库连接或 mock DB 错误（可通过查看日志验证）

**验证日志中包含**：
```
[WARN] 创建发布静默规则失败（不影响部署）: <error>
```

且部署仍然正常执行完成。

---

#### 测试 10：前端 UI 验证

**目的**：验证前端配置界面

1. **访问流水线创建页面** → 进入「自动部署」步骤
2. **开启「自动部署」开关** → 应出现部署配置面板
3. **开启「发布静默告警」开关** → 应出现暖黄色配置面板
4. **配置缓冲时间** → 输入范围 1-25
5. **选择告警级别** → 下拉选项正确
6. **提交创建** → 检查 Network 请求中包含 3 个静默字段
7. **编辑已有流水线** → 静默配置正确回显

---

#### 测试 11：30 分钟自动失效兜底

**目的**：验证最大时限保护

```sql
-- 创建一个 buffer=30 的流水线
UPDATE cicd_pipeline 
SET enable_deploy_silence = 1, silence_buffer_minutes = 30 
WHERE id = <ID>;
```

触发部署后验证：
```sql
SELECT (ends_at - starts_at) / 60 as duration_minutes
FROM monitor_silence_rules 
WHERE name LIKE '[发布静默]%' 
ORDER BY id DESC LIMIT 1;
```

**期望结果**：`duration_minutes = 30`（被截断到上限，而非 35）

---

#### 测试 12：日志验证

**目的**：验证关键日志输出

**应用日志中应包含**：

| 场景 | 日志关键词 |
|------|-----------|
| 创建成功 | `[发布静默] 静默规则已创建 rule_id=XX pipeline=XX` |
| 部署成功释放 | `[发布静默] 部署成功，静默规则将延续2分钟后解除` |
| 部署失败释放 | `[发布静默] 部署失败，静默规则已立即解除` |
| 创建失败 | `[WARN] 创建发布静默规则失败` |

**部署阶段实时日志中应包含**：
```
[INFO] 发布静默已生效，规则ID: 15，级别: warning,info
...
[INFO] 发布静默将在2分钟后自动解除
```

---

### 7.3 回归验证清单

确保发布静默功能不影响现有功能：

| 验证项 | 方法 | 预期 |
|--------|------|------|
| 不开启静默的流水线正常部署 | 触发 `enable_deploy_silence=false` 的流水线 | 正常部署，无静默规则产生 |
| 手动创建的静默规则不受影响 | 通过监控页面创建静默规则 | 正常工作 |
| 现有告警评估不受影响 | 检查 AlertEvalWorker 日志 | 非部署期间告警正常发送 |
| 流水线列表接口正常 | GET /pipeline/list | 返回含新字段的完整数据 |
| 流水线详情接口正常 | GET /pipeline/detail?id=X | 含 3 个新字段 |
| 批量创建流水线正常 | POST /pipeline/batch-create | 支持传入静默配置 |

---

## 八、运维操作指南

### 8.1 生产环境推荐配置

| 环境 | enable_deploy_silence | silence_severities | silence_buffer_minutes |
|------|----------------------|-------------------|----------------------|
| 开发 | false | - | - |
| 测试 | false | - | - |
| 预发 | true | warning,info | 5 |
| **生产** | **true** | **warning** | **10** |

> **生产环境建议**：仅静默 warning 级别，保留 critical 通知（OOM、镜像拉取失败等真问题）。

### 8.2 批量开启静默

```sql
-- 为所有生产环境流水线开启静默
UPDATE cicd_pipeline 
SET enable_deploy_silence = 1, 
    silence_buffer_minutes = 10, 
    silence_severities = 'warning,info'
WHERE deploy_env = 'prod' AND is_del = 0;
```

### 8.3 紧急手动解除静默

如果需要立即解除所有自动创建的静默：

```sql
-- 禁用所有活跃的发布静默规则
UPDATE monitor_silence_rules 
SET enabled = 0, ends_at = UNIX_TIMESTAMP(NOW())
WHERE name LIKE '[发布静默]%' AND enabled = 1 AND is_del = 0;
```

### 8.4 清理历史静默规则

```sql
-- 软删除 7 天前的过期发布静默
UPDATE monitor_silence_rules 
SET is_del = 1 
WHERE name LIKE '[发布静默]%' 
AND ends_at < UNIX_TIMESTAMP(NOW() - INTERVAL 7 DAY);
```

### 8.5 监控静默规则健康

```sql
-- 查看当前活跃的发布静默（不应该超过流水线并发数）
SELECT COUNT(*) as active_count, 
       GROUP_CONCAT(name SEPARATOR '\n') as rules
FROM monitor_silence_rules 
WHERE name LIKE '[发布静默]%' 
AND enabled = 1 AND is_del = 0 
AND ends_at > UNIX_TIMESTAMP(NOW());
```

---

## 九、故障排查指南

### 9.1 静默规则未创建

**排查步骤**：
1. 确认流水线 `enable_deploy_silence = true`
2. 检查应用日志中是否有 WARN 级别的静默创建失败信息
3. 确认数据库连接正常
4. 确认 `monitor_silence_rules` 表存在且可写入

### 9.2 告警仍然在发送（静默未生效）

**排查步骤**：
1. 确认静默规则的 `matchers` 与告警标签匹配
2. 检查告警的 `severity` 是否在配置的 `silence_severities` 中
3. 确认 AlertEvalWorker 的静默匹配逻辑正确（按 `namespace` + `workload` + `severity` 三维匹配）
4. 确认 `enabled = 1` 且 `ends_at > now`

### 9.3 静默未自动解除

**排查步骤**：
1. 检查部署是否正常结束（成功或失败）
2. 查看日志中是否有 `ReleaseDeploySilence` 相关输出
3. 确认 `ends_at` 字段是否被正确更新
4. 兜底：30 分钟后一定会自然过期

---

## 十、代码结构速查

```
internal/app/
├── models/
│   └── cicd_pipeline.go          # 模型定义（3个新字段）
├── requests/
│   └── cicd_pipeline.go          # API 请求结构体
├── services/
│   ├── cicd_deploy_silence.go    # ★ 核心：静默生命周期管理
│   ├── cicd_stage.go             # 阶段化部署集成点
│   └── cicd_pipeline.go          # 旧配置部署集成点 + 字段映射
├── controllers/api/v1/cicd/
│   └── pipeline_controller.go    # DeploySilenceStatus API
└── routers/kube_cicd/
    └── cicd_router.go            # 路由注册

k8s-web/src/views/cicd/
└── PipelineCreate.vue            # 前端配置 UI
```

---

## 十一、安全与风控

| 安全项 | 机制 |
|--------|------|
| 不静默 critical | 默认 `silence_severities = "warning,info"` |
| 最长 30 分钟 | `totalMinutes = min(5+buffer, 30)` |
| 失败立即解除 | `ReleaseDeploySilence(ctx, id, false)` |
| 创建失败不阻塞 | `if silenceErr != nil { logs.WriteString(WARN) }` |
| 规则可追溯 | `[发布静默]` 前缀 + comment 包含流水线信息 |
| 手动可覆盖 | SQL 或 API 可随时解除 |

---

## 十二、总结

本功能通过**零新建表、最小侵入**的方式，在 CI/CD 部署流程中自动联动告警静默，实现了：

1. **自动化** — 配置一次，每次部署自动生效
2. **精准匹配** — namespace + workload + severity 三维定向
3. **安全兜底** — 30 分钟上限 + critical 不静默 + 失败立即解除
4. **可观测** — 日志记录 + 状态查询 API + 规则可审计
5. **零阻塞** — 静默服务异常不影响部署流程
