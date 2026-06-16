# AIOps 智能运维 — 架构设计与验证指南

> 版本：v1.0 | 更新日期：2026-05-27

---

## 一、功能概述

AIOps 智能运维是 K8s 管理平台的核心 AI 赋能模块，基于 **大模型 + 平台健康数据 + Loki 日志 + Prometheus 告警** 构建。通过 AI 对运维数据的深度分析，实现从"被动响应"到"智能预判"的运维模式转变。

### 三大核心能力

| 能力 | 说明 | 触发方式 |
|------|------|----------|
| **AI 告警分析** | 基于告警事件 + PromQL 规则 + 上下文，AI 自动定位根因、评估影响、给出处置建议 | 手动（在告警事件页点击分析） |
| **AI 日志诊断** | 集成 Loki，自动查询 + 采样错误日志，AI 分析异常模式并给出修复方案 | 手动（指定 namespace/pod/query） |
| **智能巡检** | 自动收集集群/节点/工作负载/告警数据，AI 生成健康报告（评分+等级+建议） | 自动（每6小时）+ 手动触发 |

### 设计理念

```
传统运维:  告警 → 人工排查 → 人工决策 → 人工处置
AIOps:     告警 → AI根因分析 → AI建议方案 → 人工确认执行
                    ↑
         定时巡检 → AI趋势预测 → 预防性建议（主动运维）
```

---

## 二、系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       前端 (Vue 3 + Composition API)                          │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │  AIOps.vue (智能运维页面)                                               │  │
│  │  ┌───────────────┐ ┌───────────────┐ ┌────────────────┐              │  │
│  │  │ 仪表盘(统计卡)  │ │ 功能操作卡片   │ │ 巡检报告列表    │              │  │
│  │  └───────────────┘ └───────────────┘ └────────────────┘              │  │
│  │  ┌───────────────┐ ┌───────────────┐ ┌────────────────┐              │  │
│  │  │ 告警分析弹窗   │ │ 日志诊断弹窗   │ │ 分析记录表格    │              │  │
│  │  └───────────────┘ └───────────────┘ └────────────────┘              │  │
│  └──────────────────────────┬────────────────────────────────────────────┘  │
│                              │ api/platform/aiops.js (Axios)                  │
└──────────────────────────────┼───────────────────────────────────────────────┘
                               │ HTTP REST
                               ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                       后端 (Go + Gin + GORM)                                 │
│                                                                              │
│  ┌─── Router Layer ──────────────────────────────────────────────────────┐  │
│  │  /api/v1/ai/ops/*  (JWT 认证保护)                                      │  │
│  └──────────────────────────┬─────────────────────────────────────────────┘  │
│                              ▼                                                │
│  ┌─── Controller Layer ──────────────────────────────────────────────────┐  │
│  │  aiops_controller.go                                                    │  │
│  │  ├ AnalyzeAlert   → POST /ai/ops/alert/analyze                         │  │
│  │  ├ DiagnoseLogs   → POST /ai/ops/log/diagnose                          │  │
│  │  ├ RunInspection  → POST /ai/ops/inspection/run                         │  │
│  │  ├ GetInspection  → GET  /ai/ops/inspection/list, /:id                  │  │
│  │  ├ GetDashboard   → GET  /ai/ops/dashboard                              │  │
│  │  └ GetRecords     → GET  /ai/ops/records                                │  │
│  └──────────────────────────┬─────────────────────────────────────────────┘  │
│                              ▼                                                │
│  ┌─── Service Layer (核心引擎) ──────────────────────────────────────────┐  │
│  │  aiops.go (816行)                                                        │  │
│  │  ├ AnalyzeAlert()       → 查事件→构建Prompt→调AI→解析→保存记录            │  │
│  │  ├ DiagnoseLogs()       → 构建LogQL→查Loki→采样→AI分析→保存              │  │
│  │  ├ RunInspection()      → 创建报告→异步执行巡检流程                       │  │
│  │  ├ executeInspection()  → 收集健康→算评分→AI报告→更新状态                 │  │
│  │  ├ GetDashboardStats()  → 统计仪表盘数据                                  │  │
│  │  └ Prompt Engineering   → 3套专业 System Prompt                          │  │
│  └─────────┬──────────────────────┬──────────────────────┬────────────────┘  │
│            ▼                      ▼                      ▼                   │
│  ┌─── 数据源 ───────┐  ┌─── AI Client ──────┐  ┌─── Worker ────────┐      │
│  │ PlatformHealth    │  │ pkg/openai/client   │  │ InspectionWorker  │      │
│  │ Loki Client       │  │ Registry (多提供商)  │  │ (6h 定时巡检)      │      │
│  │ MySQL (GORM)      │  │ Chat / ChatStream   │  │ stopCh 优雅退出    │      │
│  └────────┬──────────┘  └────────┬────────────┘  └───────────────────┘      │
│           ▼                      ▼                                           │
│  ┌──────────────────────────────────────────────────────┐                   │
│  │  MySQL         Prometheus       Loki       AI API    │                   │
│  │  (2张AIOps表)  (告警规则/事件)  (容器日志)  (多模型)   │                   │
│  └──────────────────────────────────────────────────────┘                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流图

```
              ┌──────────────────── 定时触发 (每6h) ────────────────────┐
              │                                                          │
              ▼                                                          │
┌─── InspectionWorker ─────┐                                           │
│  1. 检查 AI 是否启用       │                                           │
│  2. 调用 RunInspection()  │◄──────── 手动触发 (前端按钮)               │
└──────────────┬────────────┘                                           │
               ▼                                                         │
┌─── executeInspection() ──────────────────────────────────────────────┐│
│                                                                       ││
│  ① PlatformHealthService.GetFullHealth()                              ││
│     └→ 集群连通性 + 节点状态 + 工作负载 + Pod + 事件                     ││
│                                                                       ││
│  ② 查询 MySQL 告警事件表 (firing 状态)                                  ││
│                                                                       ││
│  ③ 构建 InspectionSummary                                             ││
│     └→ 集群/节点/工作负载/Pod/告警 数量统计                               ││
│                                                                       ││
│  ④ calculateHealthScore() → 加权评分算法                               ││
│     └→ 集群(-20)/节点(-10)/工作负载(-2)/告警(-3,-5)                     ││
│                                                                       ││
│  ⑤ AI 生成巡检报告 (inspectionSystemPrompt)                            ││
│     └→ 整体评估 + 问题发现 + 优化建议 + 趋势预测                        ││
│                                                                       ││
│  ⑥ 更新巡检报告 → completed 状态                                       ││
└───────────────────────────────────────────────────────────────────────┘│
                                                                         │
              ┌─────────────────────────────────────────────────────────┘
              │
              └─── Bootstrap 注入启动 / StopCicdWorker 注入停止
```

### 2.3 技术栈

| 层级 | 技术 |
|------|------|
| 前端框架 | Vue 3 + Composition API + Vite |
| 后端框架 | Go 1.21+ / Gin v1.11 / GORM v1.30 |
| AI SDK | `sashabaranov/go-openai`（兼容 OpenAI API 协议） |
| 模型管理 | `pkg/openai/registry.go` Provider Registry 多提供商模式 |
| 日志系统 | Grafana Loki + `pkg/loki/client.go` |
| 告警系统 | Prometheus + 内置告警评估引擎 |
| 定时任务 | Go goroutine + time.Ticker + graceful shutdown |
| 数据库 | MySQL 8.x (GORM AutoMigrate) |
| 认证 | JWT Bearer Token |

---

## 三、文件结构

```
k8s_operation/
├── internal/app/
│   ├── models/
│   │   └── aiops.go                       # 数据模型（2张表 + 常量）
│   ├── services/
│   │   └── aiops.go                       # 核心服务（816行，AI分析+巡检+Prompt）
│   ├── controllers/api/v1/ai/
│   │   └── aiops_controller.go            # HTTP 控制器（7个端点）
│   ├── routers/ai_assistant/
│   │   └── router.go                      # 路由注册（ops 子组）
│   └── worker/
│       └── aiops_inspection_worker.go     # 定时巡检 Worker
├── internal/bootstrap/
│   └── bootstrap.go                       # Worker 启动/停止 + AutoMigrate
└── k8s-web/src/
    ├── api/platform/
    │   └── aiops.js                       # 前端 API 封装
    ├── views/platform/
    │   └── AIOps.vue                      # 智能运维页面
    ├── router/
    │   └── index.js                       # 前端路由
    └── components/
        └── Layout.vue                     # 侧边栏菜单入口
```

---

## 四、数据库设计

### 4.1 aiops_analysis_record（AI 分析记录表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 自增主键 |
| type | varchar(30) | 类型: alert_analysis / log_diagnosis / inspection |
| ref_id | bigint | 关联 ID（告警事件ID / 巡检报告ID） |
| title | varchar(300) | 分析标题 |
| input | text | 输入数据（Prompt 内容） |
| result | longtext | AI 分析结果（Markdown） |
| severity | varchar(20) | AI 判定严重级别 |
| suggestions | text | AI 建议 JSON |
| model | varchar(100) | 使用的 AI 模型 |
| tokens_used | int | 消耗 Token 数 |
| latency_ms | bigint | 分析耗时(ms) |
| status | varchar(20) | success / failed / timeout |
| error | varchar(500) | 错误信息 |
| user_id | bigint | 发起人 |
| created_at | bigint | 创建时间(Unix) |

### 4.2 aiops_inspection_report（巡检报告表）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 自增主键 |
| type | varchar(30) | scheduled / manual |
| scope | varchar(30) | full / cluster / namespace |
| scope_id | varchar(100) | 范围 ID |
| health_score | int | 健康评分 0-100 |
| level | varchar(20) | healthy / warning / critical |
| summary | text | 巡检摘要 |
| details | longtext | 巡检详情 JSON |
| ai_analysis | longtext | AI 综合分析（Markdown） |
| findings | int | 发现问题数 |
| suggestions_count | int | AI 建议数 |
| duration | bigint | 巡检耗时(ms) |
| status | varchar(20) | running / completed / failed |
| error | varchar(500) | 错误信息 |
| triggered_by | bigint | 触发人(0=系统) |
| created_at | bigint | 创建时间 |
| completed_at | bigint | 完成时间 |

---

## 五、API 接口文档

### 5.1 接口总览

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/ai/ops/alert/analyze` | AI 告警分析 |
| POST | `/api/v1/ai/ops/log/diagnose` | AI 日志诊断 |
| POST | `/api/v1/ai/ops/inspection/run` | 手动触发巡检 |
| GET | `/api/v1/ai/ops/inspection/list` | 巡检报告列表 |
| GET | `/api/v1/ai/ops/inspection/:id` | 巡检报告详情 |
| GET | `/api/v1/ai/ops/dashboard` | 仪表盘数据 |
| GET | `/api/v1/ai/ops/records` | 分析记录列表 |

### 5.2 AI 告警分析

```http
POST /api/v1/ai/ops/alert/analyze
Authorization: Bearer <token>
Content-Type: application/json

{
  "event_id": 42,
  "provider_id": "",   // 可选，默认使用系统配置
  "model_id": ""       // 可选
}
```

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "event_id": 42,
    "rule_name": "Pod CPU 使用率过高",
    "severity": "critical",
    "analysis": "## 🔍 根因分析\n...(Markdown)\n## 📋 处置建议\n- ...",
    "root_cause": "容器 CPU limits 设置过低导致 Throttling",
    "impact": "影响 production 命名空间下 order-service 的请求响应时间",
    "suggestions": ["调整 CPU limits 至 2000m", "检查是否存在内存泄漏"],
    "priority": "critical",
    "latency_ms": 3420
  }
}
```

### 5.3 AI 日志诊断

```http
POST /api/v1/ai/ops/log/diagnose
Authorization: Bearer <token>
Content-Type: application/json

{
  "namespace": "production",
  "pod": "order-service",
  "container": "",
  "time_range": "15m",
  "query": "",         // 可选，自动根据 namespace/pod 构建 LogQL
  "provider_id": "",
  "model_id": ""
}
```

**响应示例：**
```json
{
  "code": 0,
  "data": {
    "query": "{namespace=\"production\",pod=~\"order-service.*\"}",
    "log_lines": 1230,
    "error_count": 47,
    "analysis": "## 📊 异常模式\n...(Markdown)",
    "pattern": "大量 OOM Killed 和 Connection Timeout",
    "root_cause": "数据库连接池耗尽",
    "suggestions": ["增加连接池 maxPoolSize", "检查慢查询"],
    "severity": "warning",
    "latency_ms": 5120
  }
}
```

### 5.4 智能巡检

```http
POST /api/v1/ai/ops/inspection/run
Authorization: Bearer <token>
```

**响应（立即返回 running 状态，异步执行巡检）：**
```json
{
  "code": 0,
  "data": {
    "id": 15,
    "type": "manual",
    "scope": "full",
    "status": "running",
    "triggered_by": 1
  }
}
```

**巡检完成后查询详情：**
```http
GET /api/v1/ai/ops/inspection/15
```

```json
{
  "code": 0,
  "data": {
    "id": 15,
    "type": "manual",
    "health_score": 85,
    "level": "healthy",
    "summary": "健康评分 85/100 | 集群 2/2 | 节点 5/6 | 告警 3 条",
    "ai_analysis": "## 📊 整体评估\n平台整体运行良好...\n## 🔍 问题发现\n...",
    "findings": 4,
    "suggestions_count": 6,
    "duration": 12340,
    "status": "completed",
    "completed_at": 1748361600
  }
}
```

### 5.5 仪表盘数据

```http
GET /api/v1/ai/ops/dashboard
```

```json
{
  "code": 0,
  "data": {
    "today_analysis": 12,
    "total_analysis": 156,
    "week_analysis": 45,
    "firing_alerts": 3,
    "last_health_score": 85,
    "last_health_level": "healthy",
    "last_inspection_at": 1748361600
  }
}
```

---

## 六、核心算法

### 6.1 健康评分算法

```go
func calculateHealthScore(summary *InspectionSummary) int {
    score := 100

    // 集群健康: 每个异常集群 -20 分
    unhealthyClusters := summary.ClustersTotal - summary.ClustersHealthy
    score -= unhealthyClusters * 20

    // 节点健康: 每个 NotReady 节点 -10 分
    unhealthyNodes := summary.NodesTotal - summary.NodesReady
    score -= unhealthyNodes * 10

    // 工作负载: 每个异常工作负载 -2 分
    unhealthyWL := summary.WorkloadsTotal - summary.WorkloadsHealthy
    score -= unhealthyWL * 2

    // 告警: 每条 firing 告警 -3 分，critical -5 分额外
    score -= summary.AlertsFiring * 3
    score -= summary.AlertsCritical * 5

    if score < 0 { score = 0 }
    return score
}
```

**评分等级映射：**

| 评分范围 | 等级 | 含义 |
|----------|------|------|
| 80-100 | `healthy` | 平台健康 |
| 60-79 | `warning` | 存在风险 |
| 0-59 | `critical` | 需要立即处理 |

### 6.2 Prompt Engineering

系统使用 3 套专业 System Prompt，分别针对不同场景：

| Prompt | 角色设定 | 输出格式 |
|--------|----------|----------|
| alertAnalysisSystemPrompt | K8s 运维 AIOps 专家 | 根因分析 + 影响范围 + 优先级 + 处置建议 + kubectl 命令 |
| logDiagnosisSystemPrompt | K8s 日志分析专家 | 异常模式 + 错误分析 + 根因判断 + 修复建议 + 风险等级 |
| inspectionSystemPrompt | K8s 平台巡检专家 | 整体评估 + 问题发现 + 优化建议 + 趋势预测 + 结论 |

---

## 七、验证指南

### 7.1 环境前提

| 组件 | 要求 | 说明 |
|------|------|------|
| MySQL | 已连接 | GORM AutoMigrate 自动建表 |
| AI 配置 | `config.yaml` 中 `ai_assistant.enabled: true` | 需配置至少一个 AI 提供商 |
| Loki（可选） | 监控数据源中已配置 Loki | 日志诊断功能依赖 |
| Prometheus（可选） | 已配置告警规则且有 firing 告警 | 告警分析功能依赖 |

### 7.2 启动验证

1. **启动后端**
```bash
cd D:\k8s-go\k8s_operation
go run cmd/k8soperation/main.go
```

观察日志输出：
```
[AIOps-InspectionWorker] 启动智能巡检引擎  {"interval": "6h0m0s"}
```

2. **确认数据库建表**
```sql
SHOW TABLES LIKE 'aiops_%';
-- 应看到：aiops_analysis_record, aiops_inspection_report
```

3. **启动前端**
```bash
cd D:\k8s-go\k8s_operation\k8s-web
npm run dev
```

4. **访问页面**
```
http://localhost:5173/platform/aiops
```

### 7.3 功能验证步骤

#### 验证一：仪表盘数据

```bash
# 获取仪表盘统计
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/ai/ops/dashboard
```

预期：返回 `today_analysis`、`firing_alerts` 等统计数据。

#### 验证二：手动触发巡检

```bash
# 触发巡检
curl -X POST -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/ai/ops/inspection/run
```

预期：
- 立即返回 `status: "running"` 的巡检报告
- 后台异步执行，几秒后查询报告变为 `completed`
- 包含 health_score、ai_analysis（Markdown 格式）

```bash
# 查询巡检列表
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/ai/ops/inspection/list
```

#### 验证三：AI 告警分析

前提：需要有 firing 状态的告警事件（可手动创建测试数据）

```bash
# 分析告警事件 ID=1
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"event_id": 1}' \
  http://localhost:8080/api/v1/ai/ops/alert/analyze
```

预期：返回 AI 生成的根因分析、影响范围、处置建议。

#### 验证四：AI 日志诊断

前提：需要已配置 Loki 数据源

```bash
# 诊断 default 命名空间的 nginx 日志
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"namespace": "default", "pod": "nginx", "time_range": "1h"}' \
  http://localhost:8080/api/v1/ai/ops/log/diagnose
```

预期：返回日志行数、错误数、AI 分析结果。

#### 验证五：前端页面

1. 登录系统，左侧导航栏「平台」→「智能运维」
2. 查看顶部 4 个统计卡片（今日分析 / 健康评分 / 活跃告警 / 累计分析）
3. 在"告警分析"卡片输入告警事件 ID，点击分析
4. 在"日志诊断"卡片输入 namespace / pod，点击诊断
5. 点击"立即巡检"按钮触发巡检
6. 查看巡检报告列表和详情弹窗

### 7.4 定时巡检验证

启动后等待 60 秒（首次巡检延迟），观察日志：

```
[AIOps-InspectionWorker] 开始定时巡检
[AIOps] 巡检完成  {"report_id": 1, "score": 92, "level": "healthy", "findings": 1, "duration_ms": 8432}
```

之后每 6 小时自动执行。

---

## 八、配置说明

### 8.1 AI 配置（config.yaml）

```yaml
ai_assistant:
  enabled: true
  providers:
    - id: "deepseek"
      name: "DeepSeek"
      api_key: "sk-xxxxx"
      base_url: "https://api.deepseek.com/v1"
      models:
        - id: "deepseek-chat"
          name: "DeepSeek Chat"
          max_tokens: 4096
```

### 8.2 Loki 配置（监控数据源）

在「监控中心」→「数据源管理」中添加 Loki 类型数据源：
- URL: `http://loki:3100`
- 类型: loki

### 8.3 巡检频率调整

修改 `aiops_inspection_worker.go`:
```go
func NewAIOpsInspectionWorker() *AIOpsInspectionWorker {
    return &AIOpsInspectionWorker{
        interval: 6 * time.Hour, // 修改此值调整巡检频率
        ...
    }
}
```

---

## 九、扩展规划

### 9.1 短期（v1.1）

| 功能 | 说明 |
|------|------|
| 告警页面一键分析 | 在告警事件列表添加"AI 分析"按钮 |
| 巡检报告导出 | 导出为 PDF / Markdown |
| 分析缓存 | 同一告警 1 小时内不重复分析 |
| Token 用量统计 | 统计各功能的 Token 消耗 |

### 9.2 中期（v1.5）

| 功能 | 说明 |
|------|------|
| 异常自愈 | AI 建议 + 一键执行（kubectl 命令） |
| 告警关联分析 | 多条告警关联性分析（同一根因） |
| 自定义巡检范围 | 按集群/命名空间巡检 |
| 巡检报告通知 | 巡检完成后推送钉钉/企微 |

### 9.3 长期（v2.0）

| 功能 | 说明 |
|------|------|
| RAG 知识库 | 接入历史事件/Runbook/内部文档 |
| 时序预测 | 基于 Prometheus 时序数据预测容量 |
| 变更影响分析 | 发布前自动评估风险 |
| 多集群对比巡检 | 跨集群健康状态对比 |

---

## 十、故障排查

### 10.1 常见问题

| 问题 | 原因 | 解决 |
|------|------|------|
| 巡检报告 status 一直是 running | AI 调用超时或 panic | 检查日志中 `[AIOps]` 前缀的错误 |
| "AI 功能未启用" | config.yaml 中 AI 未开启 | 确认 `ai_assistant.enabled: true` |
| "Loki 未配置" | 监控数据源中无 Loki | 在数据源管理中添加 Loki |
| "告警事件不存在" | event_id 无效 | 先在告警事件页确认 ID |
| 健康评分为 0 | 大量集群/节点异常 | 检查集群连通性 |

### 10.2 日志关键字

```bash
# 过滤 AIOps 相关日志
grep "[AIOps" storage/logs/app.log

# 关键日志标记
[AIOps-InspectionWorker]  → 巡检 Worker 生命周期
[AIOps] 巡检完成           → 巡检执行结果
[AIOps] 告警分析失败       → 告警分析错误
[AIOps] 日志诊断失败       → 日志诊断错误
[AIOps] 保存分析记录失败   → 数据库写入错误
```

---

## 十一、性能指标

| 指标 | 目标值 | 说明 |
|------|--------|------|
| 告警分析延迟 | < 10s | 取决于 AI API 响应速度 |
| 日志诊断延迟 | < 15s | Loki 查询 + AI 分析 |
| 巡检执行耗时 | < 30s | 健康数据收集 + AI 报告生成 |
| 巡检 Worker 内存 | < 50MB | 无状态，按需创建 |
| 最大并发分析 | 10 | Gin 默认并发，AI API 限流 |

---

## 十二、安全设计

1. **认证保护**：所有 AIOps 接口位于 JWT 认证中间件之后
2. **敏感数据脱敏**：日志样本不超过 80 行，避免向 AI 泄露大量数据
3. **操作审计**：每次分析记录用户 ID 和时间
4. **AI 输出限制**：Prompt 中约束输出格式，避免执行建议被直接执行
5. **资源保护**：巡检超时 5 分钟自动取消，防止 goroutine 泄漏
