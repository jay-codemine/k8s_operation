# AI 智能助手 — 功能与架构设计文档

> 版本：v2.0 | 更新日期：2026-04-13

---

## 一、功能概述

AI 助手是 K8s 管理平台的核心智能化模块，基于 **多模型提供商 + Function Calling** 实现。支持 **OpenAI / DeepSeek / 智谱GLM / 通义千问 / Moonshot** 等多家大模型，用户可自由切换。通过自然语言对话即可查询、管理 Kubernetes 集群资源，系统自动识别操作意图、选择工具函数执行，并对高危操作强制要求人工审批，实现 **“对话即操作”** 的智能运维体验。

### 核心能力

| 能力 | 说明 |
|------|------|
| 多模型提供商 | 支持 OpenAI / DeepSeek / 智谱GLM / 通义千问 / Moonshot 等，用户可实时切换 |
| 智能对话 | 支持普通对话、流式对话 (SSE)、快捷问答三种模式 |
| 意图识别 | GPT 自动分析用户输入，返回结构化意图 JSON |
| Function Calling | GPT 自动选择 **40+** 个工具函数直接操作平台 |
| 风险分级 | 工具按 `read / write / danger / critical` 四级风险分级 |
| 自动执行 | 只读查询类操作直接执行并返回真实数据 |
| 审批拦截 | 写操作和高危操作自动创建审批请求，管理员审批后异步执行 |
| 会话管理 | 支持多轮对话、历史消息、会话归档 |
| 审批管理 | 审批列表、我的审批、通过/拒绝/取消、操作日志 |

---

## 二、系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         前端 (Vue 3 + Vite)                         │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  AiAssistant.vue (全局悬浮聊天组件)                           │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐ │   │
│  │  │ 对话窗口  │  │ 消息列表  │  │ 审批徽标  │  │ 会话管理     │ │   │
│  │  └──────────┘  └──────────┘  └──────────┘  └─────────────┘ │   │
│  └──────────────────────────┬───────────────────────────────────┘   │
│                             │ api/ai.js (Axios)                     │
└─────────────────────────────┼───────────────────────────────────────┘
                              │ HTTP / SSE
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     后端 (Go + Gin + GORM)                          │
│                                                                     │
│  ┌─── Router Layer ─────────────────────────────────────────────┐   │
│  │  /api/v1/ai/*  (JWT 认证保护)                                 │   │
│  └──────────────────────────┬────────────────────────────────────┘   │
│                             ▼                                       │
│  ┌─── Controller Layer ─────────────────────────────────────────┐   │
│  │  ai_controller.go       → Chat / ChatStream / QuickAsk       │   │
│  │  approval_controller.go → List / Approve / Reject / Cancel   │   │
│  └──────────────────────────┬────────────────────────────────────┘   │
│                             ▼                                       │
│  ┌─── Service Layer (核心引擎) ─────────────────────────────────┐   │
│  │  ai_assistant.go  → 对话引擎 + 会话管理 + 审批流程            │   │
│  │  ai_tools.go      → 40+ 工具定义 + 风险等级注册表             │   │
│  │  ai_executor.go   → 工具执行器，调用真实 K8s/平台 API         │   │
│  └──────────┬────────────────────────┬───────────────────────────┘   │
│             ▼                        ▼                              │
│  ┌─── DAO Layer ────┐     ┌─── OpenAI Client ───────┐             │
│  │  ai_assistant.go  │     │  pkg/openai/client.go    │             │
│  │  ├ Conversation   │     │  ├ Chat (普通)           │             │
│  │  ├ Message        │     │  ├ ChatWithTools (FC)    │             │
│  │  ├ ApprovalReq    │     │  ├ ChatStream (SSE)     │             │
│  │  └ ApprovalLog    │     │  ├ ContinueWithResults  │             │
│  └────────┬──────────┘     │  └ AnalyzeIntent        │             │
│           ▼                └──────────┬──────────────┘             │
│     ┌──────────┐                      ▼                            │
│     │  MySQL   │             ┌──────────────────┐                  │
│     │  4张AI表  │             │  OpenAI GPT API  │                  │
│     └──────────┘             │  (gpt-4o-mini)   │                  │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 多模型提供商架构

```
┌─────────────────────────────────────────────────────────┐
│              前端模型选择器 (Model Picker)                   │
│  ┌───────────┐ ┌───────────┐ ┌─────────┐ ┌─────────┐ │
│  │ 🟢 OpenAI  │ │ 🔵 DeepSeek│ │ 🟣 智谱  │ │ 🟠 通义  │ │
│  │ GPT-4o    │ │ deepseek  │ │ GLM-4   │ │ Qwen    │ │
│  │ GPT-4o-mi │ │ reasoner  │ │ GLM-4+  │ │ Qwen+   │ │
│  └───────────┘ └───────────┘ └─────────┘ └─────────┘ │
└───────────────────────┬─────────────────────────────────┘
                        │ provider_id + model_id
                        ▼
┌─────────────────────────────────────────────────────────┐
│           Registry (提供商注册中心)                        │
│  ┌─────────────────────────────────────────────────┐  │
│  │  providers map[providerID] → providerEntry         │  │
│  │    ├ config (APIKey, BaseURL, Temperature)      │  │
│  │    ├ models map[modelID] → AIModelConfig         │  │
│  │    └ clients map[modelID] → *Client (懒加载缓存)   │  │
│  └─────────────────────────────────────────────────┘  │
│  GetClient(providerID, modelID, prompt) → *Client     │
│  ListProviders() → []ProviderInfo (前端展示用)          │
└───────────────┬─────────────┬─────────────┬─────────────┘
                │             │             │
                ▼             ▼             ▼
      api.openai.com  api.deepseek.com  open.bigmodel.cn ...
      (均兼容 OpenAI API 协议，复用 sashabaranov/go-openai SDK)
```

### 2.3 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + Arco Design + Vite |
| 后端框架 | Go + Gin v1.11 + GORM v1.30 |
| AI SDK | `sashabaranov/go-openai` (社区版，兼容所有 OpenAI API 协议提供商) |
| 模型管理 | `pkg/openai/registry.go` Provider Registry 模式 |
| 数据库 | MySQL 8.x |
| 认证 | JWT Bearer Token |
| 配置 | `configs/config.yaml` → `AIAssistant` 配置块 |

---

## 三、核心对话流程

### 3.1 Function Calling 对话引擎

```
用户消息
    │
    ▼
┌─────────────────────────────────────────────────┐
│  1. 获取/创建会话 (ai_conversations)             │
│  2. 保存用户消息 (ai_messages)                    │
│  3. 构建历史上下文 (最近 N 轮消息)                │
│  4. 构建工具列表 BuildAllTools() → 40+ 工具       │
│  5. 调用 GPT ChatWithTools()                     │
└──────────────────────┬──────────────────────────┘
                       ▼
            ┌─── GPT 返回 ───┐
            │                 │
      无 tool_calls     有 tool_calls
            │                 │
            ▼                 ▼
       直接返回文本     逐个处理工具调用
                              │
                 ┌────────────┼────────────┐
                 ▼            ▼            ▼
              read 级      write 级    danger/critical 级
            (直接执行)    (创建审批)      (强制审批)
                 │            │            │
                 ▼            └────┬───────┘
          ExecuteToolCall()        ▼
          调用真实 K8s API   createToolApproval()
                 │          写入 ai_approval_requests
                 ▼                 │
          工具结果反馈 GPT          ▼
                 │         告知 GPT "已提交审批"
                 ▼                 │
          GPT 整理数据      GPT 生成审批提示回复
          生成用户友好回复          │
                 │                 │
                 └────────┬────────┘
                          ▼
                    保存 AI 回复消息
                    返回给前端
```

### 3.2 普通查询示例

```
用户: "帮我看看 default 命名空间有哪些 Pod"

→ GPT 自动选择工具: list_pods(namespace="default", cluster_id=2)
→ 风险等级: read → 直接执行
→ 调用 KubePodList() 获取真实 Pod 数据
→ 工具结果反馈给 GPT
→ GPT 整理输出:

  "default 命名空间下共有 5 个 Pod：
   1. nginx-7d5b8c9f-abc12  Running  (节点: node-1)
   2. redis-6f8b2a1c-def34  Running  (节点: node-2)
   ..."
```

### 3.3 高危操作审批示例

```
用户: "删除 default 命名空间的 nginx deployment"

→ GPT 选择工具: delete_deployment(namespace="default", name="nginx", cluster_id=2)
→ 风险等级: danger → NeedApproval = true
→ 创建审批请求 (ai_approval_requests)
→ 记录审批日志 (ai_approval_logs)
→ GPT 生成回复:

  "⚠️ 检测到需要审批的操作:
   - 删除 Deployment（风险等级: danger，审批ID: 1）
   请等待管理员审批后自动执行，或在「审批管理」页面查看进度。"

→ 管理员在审批页面点击「通过」
→ AIApprovalApprove() 更新状态 = approved
→ 异步执行 executeApprovedTool()
→ 调用 delete_deployment() 执行真实删除
→ 记录执行结果到 ai_approval_requests.execute_result
```

---

## 四、工具函数清单

### 4.1 查询类工具 (read, 24个, 直接执行)

| 工具名 | 功能描述 |
|--------|----------|
| `list_pods` | 查询指定命名空间下的 Pod 列表 |
| `get_pod_detail` | 获取指定 Pod 的详细信息 |
| `get_pod_logs` | 查看 Pod 日志 |
| `list_deployments` | 查询 Deployment 列表 |
| `get_deployment_detail` | 获取 Deployment 详情（副本数/镜像/状态） |
| `list_services` | 查询 Service 列表 |
| `get_service_detail` | 获取 Service 详情 |
| `list_namespaces` | 查询所有命名空间 |
| `list_nodes` | 查询所有节点 |
| `get_node_detail` | 获取节点详情（CPU/内存/标签/污点） |
| `list_ingresses` | 查询 Ingress 列表 |
| `list_configmaps` | 查询 ConfigMap 列表 |
| `list_secrets` | 查询 Secret 列表 |
| `list_pvcs` | 查询 PVC 列表 |
| `list_clusters` | 查询平台管理的所有 K8s 集群 |
| `list_pipelines` | 查询 CI/CD 流水线列表 |
| `get_pipeline_detail` | 获取流水线详情 |
| `list_statefulsets` | 查询 StatefulSet 列表 |
| `list_daemonsets` | 查询 DaemonSet 列表 |
| `list_jobs` | 查询 Job 列表 |
| `list_cronjobs` | 查询 CronJob 列表 |
| `get_node_metrics` | 查询节点指标 |
| `get_events` | 查询 K8s 事件列表（排查问题） |

### 4.2 写操作工具 (write, 12个, 需审批确认)

| 工具名 | 功能描述 |
|--------|----------|
| `scale_deployment` | 调整 Deployment 副本数（扩缩容） |
| `restart_deployment` | 重启 Deployment（触发滚动更新） |
| `rollback_deployment` | 回滚 Deployment 到指定版本 |
| `update_deployment_image` | 更新 Deployment 容器镜像 |
| `create_deployment` | 创建 Deployment |
| `create_service` | 创建 Service |
| `create_namespace` | 创建 Namespace |
| `create_configmap` | 创建 ConfigMap |
| `create_ingress` | 创建 Ingress |
| `create_pvc` | 创建 PVC |
| `trigger_pipeline` | 触发 CI/CD 流水线构建 |
| `cordon_node` | 节点调度开关 (cordon/uncordon) |

### 4.3 高危操作工具 (danger/critical, 11个, 强制审批)

| 工具名 | 风险等级 | 功能描述 |
|--------|----------|----------|
| `delete_pod` | danger | 删除指定 Pod |
| `delete_deployment` | danger | 删除 Deployment 及所有 Pod |
| `delete_service` | danger | 删除 Service |
| `delete_configmap` | danger | 删除 ConfigMap |
| `delete_ingress` | danger | 删除 Ingress |
| `delete_pvc` | danger | 删除 PVC（可能丢数据） |
| `delete_statefulset` | danger | 删除 StatefulSet |
| `delete_daemonset` | danger | 删除 DaemonSet |
| `delete_pipeline` | danger | 删除流水线 |
| `delete_namespace` | **critical** | 删除整个命名空间（极高危） |
| `drain_node` | **critical** | 排空节点（驱逐所有 Pod） |

---

## 五、API 接口清单

所有接口前缀: `/api/v1/ai/`，需 JWT Bearer Token 认证。

### 5.1 AI 对话

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/status` | AI 助手状态检查（是否启用、默认提供商、模型数） |
| `GET` | `/models` | 获取可用 AI 提供商和模型列表 |
| `POST` | `/chat` | 普通对话（支持 provider_id + model_id 指定模型） |
| `POST` | `/chat/stream` | 流式对话（SSE 模式） |
| `POST` | `/quick-ask` | 快捷问答（无会话上下文） |
| `POST` | `/intent` | 意图分析（返回结构化意图 JSON） |

#### 请求示例 — `/chat`

```json
POST /api/v1/ai/chat
Authorization: Bearer <token>

{
  "conversation_id": 0,
  "message": "帮我看看 default 命名空间有哪些 Pod",
  "provider_id": "deepseek",
  "model_id": "deepseek-chat"
}
```

#### 响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "conversation_id": 1,
    "reply": "default 命名空间下共有 5 个 Pod...",
    "need_approval": false,
    "tools_called": ["list_pods"],
    "pending_tools": []
  }
}
```

#### 高危操作响应示例

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "conversation_id": 1,
    "reply": "⚠️ 检测到需要审批的操作...",
    "need_approval": true,
    "approval_id": 1,
    "tools_called": [],
    "pending_tools": [
      {
        "tool_name": "delete_deployment",
        "approval_id": 1,
        "risk_level": "danger",
        "summary": "删除 Deployment"
      }
    ]
  }
}
```

### 5.2 会话管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/conversations` | 获取用户会话列表 |
| `GET` | `/conversations/:id/messages` | 获取会话消息历史 |
| `DELETE` | `/conversations/:id` | 归档（软删除）会话 |

### 5.3 审批管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/approvals` | 审批列表（管理员视角） |
| `GET` | `/approvals/mine` | 我的审批申请 |
| `GET` | `/approvals/pending-count` | 待审批数量（前端徽标） |
| `GET` | `/approvals/:id` | 审批详情 |
| `POST` | `/approvals/:id/approve` | 通过审批 → 异步执行工具 |
| `POST` | `/approvals/:id/reject` | 拒绝审批 |
| `POST` | `/approvals/:id/cancel` | 取消审批 |

---

## 六、数据库表设计

### 6.1 ai_conversations — AI 会话

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint32 PK | 会话 ID |
| user_id | uint32 | 关联用户 |
| title | varchar(200) | 会话标题 |
| status | uint8 | 1=活跃 2=归档 |
| created_at | uint32 | 创建时间 |
| modified_at | uint32 | 更新时间 |

### 6.2 ai_messages — 聊天消息

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint32 PK | 消息 ID |
| conversation_id | uint32 FK | 关联会话 |
| role | varchar(20) | system / user / assistant |
| content | text | 消息内容 |
| intent_json | text | 意图识别结果 JSON |
| token_used | int | Token 消耗量 |
| created_at | uint32 | 创建时间 |

### 6.3 ai_approval_requests — 高危操作审批

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint32 PK | 审批 ID |
| conversation_id | uint32 | 关联 AI 会话 |
| request_user_id | uint32 | 发起人 |
| approver_user_id | uint32 | 审批人 |
| intent | varchar(50) | 操作意图 (delete/drain/scale) |
| resource | varchar(100) | 资源类型 (deployment/namespace) |
| resource_name | varchar(200) | 资源名称 |
| namespace | varchar(100) | 命名空间 |
| cluster_id | uint32 | 目标集群 |
| risk_level | varchar(20) | 风险等级 |
| operation_json | text | 完整操作参数 JSON |
| tool_name | varchar(100) | Function Calling 工具名 |
| tool_args_json | text | 工具调用参数 JSON |
| tool_call_id | varchar(100) | OpenAI tool_call_id |
| execute_result | text | 执行结果 |
| executed | bool | 是否已执行 |
| summary | varchar(500) | 操作摘要 |
| status | uint8 | 1=待审批 2=通过 3=拒绝 4=过期 5=取消 |
| approve_comment | varchar(500) | 审批备注 |
| expire_at | uint32 | 过期时间戳 |
| created_at | uint32 | 创建时间 |
| modified_at | uint32 | 更新时间 |

### 6.4 ai_approval_logs — 审批操作日志

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint32 PK | 日志 ID |
| approval_id | uint32 FK | 关联审批请求 |
| user_id | uint32 | 操作人 (0=系统自动) |
| action | varchar(50) | create / approve / reject / cancel / execute |
| comment | varchar(500) | 操作说明 |
| created_at | uint32 | 创建时间 |

---

## 七、核心文件索引

```
k8s_operation/
├── configs/
│   └── config.yaml                          # AIAssistant 多提供商配置块
├── pkg/openai/
│   ├── client.go                            # OpenAI 客户端封装 (Chat/Stream/FC/Intent)
│   └── registry.go                          # 多模型提供商注册中心 (Provider Registry)
├── internal/app/
│   ├── models/
│   │   └── ai_assistant.go                  # 4 张表模型定义 + CRUD 方法
│   ├── dao/
│   │   └── ai_assistant.go                  # DAO 层数据访问
│   ├── services/
│   │   ├── ai_assistant.go                  # 对话引擎 + 会话管理 + 审批流程
│   │   ├── ai_tools.go                      # 40+ 工具定义 + 风险等级注册表
│   │   └── ai_executor.go                   # 工具执行器 (调用真实 K8s API)
│   ├── controllers/api/v1/ai/
│   │   ├── ai_controller.go                 # AI 对话控制器
│   │   └── approval_controller.go           # 审批管理控制器
│   └── routers/ai_assistant/
│       └── router.go                        # AI 路由注册
├── internal/errorcode/
│   └── ai_assistant.go                      # AI 错误码定义 (800001-800012)
├── initialize/
│   ├── db.go                                # AutoMigrate 注册 AI 表
│   └── router.go                            # 路由注入 (protected 分组)
└── k8s-web/src/
    ├── api/
    │   └── ai.js                            # AI API 调用封装
    └── components/
        └── AiAssistant.vue                  # 全局悬浮聊天组件
```

---

## 八、配置说明

`configs/config.yaml` 中的 AI 配置块:

```yaml
AIAssistant:
  Enabled: true
  DefaultProvider: "openai"           # 默认提供商 ID
  SystemPrompt: "你是 K8s 管理平台的 AI 助手..."
  ApprovalExpire: 30
  MaxHistoryRound: 20

  # 多提供商配置（均兼容 OpenAI API 协议）
  Providers:
    - ID: "openai"
      Name: "OpenAI"
      Icon: "openai"
      APIKey: "sk-xxx"
      BaseURL: ""                     # 留空走官方 api.openai.com
      MaxTokens: 2048
      Temperature: 0.7
      Models:
        - ID: "gpt-4o-mini"
          Name: "GPT-4o Mini"
          Description: "快速响应，性价比高"
          Capability: "chat"

    - ID: "deepseek"
      Name: "DeepSeek"
      Icon: "deepseek"
      APIKey: "sk-xxx"
      BaseURL: "https://api.deepseek.com/v1"
      Models:
        - ID: "deepseek-chat"
          Name: "DeepSeek Chat"
          Description: "国产最强通用对话"
          Capability: "chat"

    - ID: "zhipu"
      Name: "智谱 AI"
      Icon: "zhipu"
      APIKey: "xxx.xxx"
      BaseURL: "https://open.bigmodel.cn/api/paas/v4"
      Models:
        - ID: "glm-4-flash"
          Name: "GLM-4 Flash"
          Capability: "chat"

    # 更多提供商: 通义千问、Moonshot...
```

---

## 九、安全设计

| 安全措施 | 说明 |
|----------|------|
| JWT 认证 | 所有 AI 接口在 `protected` 路由组下，需有效 Token |
| 风险分级 | 40+ 工具按 read/write/danger/critical 四级分级 |
| 强制审批 | write/danger/critical 操作自动创建审批，不直接执行 |
| 审批过期 | 审批请求默认 30 分钟过期 |
| 审批日志 | 每次审批操作都记录日志 (ai_approval_logs) |
| 异步执行 | 审批通过后，工具调用在后台异步执行，避免阻塞 |
| 静默容错 | 前端 pending-count 轮询使用 `_silent` 模式，错误不弹 toast |
| Function Calling 循环限制 | 最多 5 轮工具调用，防止死循环 |

---

## 十、AI 助手与平台的解耦设计

### 10.1 核心原则：AI 是独立模块，故障不影响平台

AI 助手与 K8s 管理平台主体之间是**完全解耦**的关系。AI 助手只是平台的一个「智能入口」，即使 AI 服务完全宕机，用户仍可通过平台界面手动完成所有操作。

```
┌────────────────────────────────────────────────────────┐
│                 K8s 管理平台（主体）                      │
│                                                        │
│  集群管理 / 资源管理 / CI/CD / RBAC / 监控 ...          │  ← 用户手动操作
│                                                        │
│  Service 层: KubePodList / DeployScale / ...            │
│                                                        │
└────────────────────────┬───────────────────────────────┘
                         │ 调用相同的 Service/API
                         │ （AI 只是另一个"用户"）
┌────────────────────────┴───────────────────────────────┐
│                AI 助手（独立模块）                        │
│                                                        │
│  对话引擎 / Function Calling / 审批流程                  │  ← AI 操作
│  独立 4 张表 / 独立路由 /api/v1/ai/*                     │
│  独立配置开关: AIAssistant.Enabled                       │
└────────────────────────────────────────────────────────┘
```

### 10.2 隔离维度

| 隔离维度 | 说明 |
|----------|------|
| **配置开关** | `config.yaml` 中 `Enabled: false` 即可完全关闭 AI，平台零影响 |
| **独立数据表** | AI 使用 4 张独立表（ai_conversations / ai_messages / ai_approval_requests / ai_approval_logs），不动平台任何表 |
| **独立路由** | AI 路由全在 `/api/v1/ai/*` 下，和平台路由互不干扰 |
| **独立依赖** | AI 依赖 OpenAI/DeepSeek 等外部 API；外部 API 挂了只影响 AI 对话，平台功能完全正常 |
| **无侵入调用** | AI 操作集群时，调用的是和用户手动操作 **完全相同** 的 Service 层接口，不引入额外副作用 |
| **独立前端组件** | AI 前端是一个悬浮组件（AiAssistant.vue），不影响平台任何页面的渲染和交互 |

### 10.3 故障隔离场景

| 故障场景 | 对平台的影响 |
|----------|--------------|
| OpenAI API Key 过期 | 无影响 — AI 对话返回错误，平台所有功能正常 |
| DeepSeek / 智谱等国产模型宕机 | 无影响 — 用户可切换其他可用模型，平台功能正常 |
| AI 配置被关闭（Enabled: false） | 无影响 — 前端 AI 按钮显示"离线"，其他页面正常 |
| AI 数据库表损坏 | 无影响 — AI 4 张表与平台业务表完全独立 |
| Function Calling 死循环 | 无影响 — 最多 5 轮自动中断，不阻塞后端主线程 |
| 审批流程出错 | 无影响 — 只影响 AI 发起的审批，手动操作不经过 AI 审批 |

---

## 十一、AI 助手能力详解

### 11.1 问答能力分类

AI 助手支持**任意问答**，不限于 K8s 运维领域：

| 问答类型 | 说明 | 示例 |
|----------|------|------|
| **通用知识问答** | 像 ChatGPT 一样回答任何知识性问题 | "什么是微服务架构？"、"帮我写个 Nginx 配置" |
| **K8s 运维咨询** | 回答 Kubernetes 相关的运维和概念问题 | "Pod 的生命周期是什么？"、"如何排查 OOMKilled？" |
| **平台查询操作** | 通过 Function Calling 调用工具获取平台真实数据 | "查看所有 Pod"、"列出集群节点" |
| **平台写操作** | 通过 Function Calling 执行变更操作（需审批） | "扩容 nginx 到 5 副本"、"删除这个 Deployment" |
| **代码辅助** | 帮助编写 YAML、Dockerfile、脚本等 | "帮我写一个 Redis Deployment 的 YAML" |
| **故障诊断** | 结合平台真实数据进行问题分析 | "为什么 Pod 一直 CrashLoopBackOff？" |

### 11.2 AI 智能判断机制

AI 会根据用户输入自动判断处理方式：

```
用户提问
   │
   ▼
AI 分析意图
   │
   ├─ 通用问题 → 直接用大模型知识回答（不调用任何工具）
   │  例: "什么是 K8s？"、"帮我写个 Dockerfile"
   │
   ├─ 查询操作 → 调用 read 工具获取真实数据 → 整理后回答
   │  例: "查看 default 下的 Pod" → list_pods()
   │
   ├─ 写操作   → 创建审批请求 → 等待管理员确认
   │  例: "扩容 nginx" → scale_deployment() → 审批
   │
   └─ 高危操作 → 强制审批 + 风险提示
      例: "删除 namespace" → delete_namespace() → 审批
```

### 11.3 多模型支持

用户可在对话界面实时切换 AI 模型，不同模型各有优势：

| 提供商 | 模型 | 特点 |
|--------|------|------|
| OpenAI | GPT-4o / GPT-4o Mini / GPT-3.5 | 综合能力最强，Function Calling 支持最佳 |
| DeepSeek | deepseek-chat / deepseek-reasoner | 国产最强，推理能力突出，性价比高 |
| 智谱 AI | GLM-4 Flash / GLM-4 Plus | 中文理解能力强，响应快 |
| 通义千问 | Qwen-Turbo / Qwen-Plus / Qwen-Max | 阿里云生态，中文优化 |
| Moonshot | moonshot-v1-8k / moonshot-v1-32k | 长上下文窗口，适合复杂分析 |

> **提示：** 所有国产大模型均兼容 OpenAI API 协议，底层复用同一个 SDK（`sashabaranov/go-openai`），只需在 `config.yaml` 配置对应的 APIKey 和 BaseURL 即可启用。

### 11.4 工具调用完整流程示例

#### 示例 1：通用问答（不调用工具）

```
用户: "K8s 的 Service 有几种类型？"

AI 回复:
  Kubernetes Service 有 4 种类型：
  1. ClusterIP（默认）— 只在集群内部可访问
  2. NodePort — 通过节点端口对外暴露
  3. LoadBalancer — 使用云厂商负载均衡器
  4. ExternalName — DNS CNAME 映射到外部服务
```

#### 示例 2：查询操作（read 工具，直接执行）

```
用户: "帮我看看 production 命名空间的 Deployment 状态"

→ AI 自动调用: list_deployments(namespace="production", cluster_id=2)
→ 风险等级: read → 直接执行
→ 获取真实数据后整理回复:

AI 回复:
  production 命名空间下共有 3 个 Deployment：
  1. api-server    3/3 Ready  镜像: api:v2.1.0
  2. web-frontend  2/2 Ready  镜像: web:v1.8.3
  3. worker        1/1 Ready  镜像: worker:v1.5.0
  所有 Deployment 运行正常 ✅
```

#### 示例 3：写操作（write 工具，需要审批）

```
用户: "把 api-server 扩容到 5 个副本"

→ AI 自动调用: scale_deployment(name="api-server", namespace="production", replicas=5)
→ 风险等级: write → NeedApproval = true
→ 创建审批请求

AI 回复:
  ⚠️ 检测到需要审批的操作：
  - 扩容 Deployment api-server 到 5 副本（风险等级: write，审批ID: 15）
  请等待管理员审批后自动执行。

→ 管理员点击「通过」后，系统自动执行扩容
```

#### 示例 4：代码辅助

```
用户: "帮我写一个 Redis 的 Deployment YAML"

AI 回复:
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: redis
    namespace: default
  spec:
    replicas: 1
    selector:
      matchLabels:
        app: redis
    template:
      metadata:
        labels:
          app: redis
      spec:
        containers:
        - name: redis
          image: redis:7-alpine
          ports:
          - containerPort: 6379
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "256Mi"
              cpu: "200m"
```

---

## 十二、意图识别与判断机制深度设计

### 12.1 核心设计原则

**AI 助手如何区分「常规问题」和「平台操作」？**

答案：**不是后端写规则判断的，而是完全由大模型自主决定的。** 这是 OpenAI **Function Calling** 协议的天然能力。

后端在每次对话时，将以下三要素一起发送给大模型 API：

| 要素 | 内容 | 作用 |
|------|------|------|
| System Prompt | "你是 K8s 管理平台的智能操作助手..." | 告诉大模型它的角色和行为准则 |
| 历史对话 | 最近 20 轮消息上下文 | 让大模型理解对话连贯性 |
| 工具定义 | 40+ 个工具的 name + description + parameters | 让大模型知道它"能做什么" |

大模型收到后，基于自身的语义理解能力自动决定：
- 不匹配任何工具 → 用知识直接回答（常规问题）
- 匹配到某个工具 → 返回 `tool_calls`（平台操作）

### 12.2 三层判断架构

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          用户输入一条消息                                  │
│            例: "什么是K8s?" / "查看 Pod" / "删除 nginx"                  │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  第一层：大模型意图识别 (AI 自主判断)                                       │
│                                                                          │
│  后端将以下内容一起发给大模型 API:                                           │
│  ① System Prompt: "你是 K8s 管理平台的智能操作助手..."                      │
│  ② 历史对话上下文 (最近 20 轮)                                              │
│  ③ 用户消息: "xxx"                                                       │
│  ④ 40+ 工具定义 (tools):                                                  │
│     [list_pods, scale_deployment, delete_pod, ...]                        │
│     每个工具都有 name + description + parameters                           │
│                                                                          │
│  大模型内部语义理解:                                                        │
│  "什么是K8s?"          → 我的知识就能回答   → 不调用工具 (ToolCalls=0)       │
│  "帮我写个 Dockerfile"  → 不需要平台数据    → 不调用工具 (ToolCalls=0)       │
│  "查看 default 的 Pod" → 需要真实数据      → 调用 list_pods (ToolCalls>0)  │
│  "扩容 nginx 到 5"     → 需要操作平台      → 调用 scale_deployment         │
│  "删除 test namespace" → 需要操作平台      → 调用 delete_namespace         │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
                  ┌────────────┴────────────┐
                  ▼                         ▼
         ToolCalls == 0               ToolCalls > 0
         (常规问题)                   (平台操作)
                  │                         │
                  ▼                         ▼
         直接返回文本回复    ┌─────────────────────────────────────┐
                            │  第二层：后端风险管控 (toolRegistry)   │
                            │                                     │
                            │  后端根据工具名查表:                   │
                            │  toolRegistry["list_pods"]           │
                            │    → RiskLevel: read                │
                            │    → NeedApproval: false → 直接执行   │
                            │  toolRegistry["scale_deployment"]    │
                            │    → RiskLevel: write               │
                            │    → NeedApproval: true  → 创建审批   │
                            │  toolRegistry["delete_namespace"]    │
                            │    → RiskLevel: critical            │
                            │    → NeedApproval: true  → 强制审批   │
                            └─────────────┬───────────────────────┘
                                          │
                                          ▼
                            ┌─────────────────────────────────────┐
                            │  第三层：人工审批 (管理员最终确认)      │
                            │                                     │
                            │  管理员在审批页面:                     │
                            │  - 查看操作详情 (工具名/参数/风险等级)  │
                            │  - 点击「通过」→ 系统异步执行操作       │
                            │  - 点击「拒绝」→ 操作取消              │
                            │  - 30 分钟未处理 → 自动过期            │
                            └─────────────────────────────────────┘
```

### 12.3 第一层：大模型意图识别的两个关键要素

#### 12.3.1 System Prompt 引导

系统提示词定义在 `internal/app/services/ai_assistant.go` 中，告诉大模型行为规则：

```go
const defaultSystemPrompt = `你是 K8s 管理平台的智能操作助手。你可以通过工具函数直接操作平台：
1. 查询和管理 Kubernetes 集群资源（Pod、Deployment、Service、Node、Namespace 等）
2. 执行资源操作（创建、删除、扩缩容、重启、回滚、镜像更新）
3. 管理 CI/CD 流水线（查询、触发构建）
4. 节点管理（查询、cordon/uncordon、drain）
5. 集群级资源管理（ConfigMap、Secret、Ingress、PVC 等）

重要规则：
- 用简洁专业的中文回答
- 当用户查询资源时，优先调用工具函数获取真实数据，而不是编造
- 用户没有指定 cluster_id 时，先调用 list_clusters 获取集群列表
- 查询类操作直接执行，写操作和删除操作会需要人工审批
- 将工具返回的数据整理成对用户友好的格式进行回答`
```

关键引导策略：

| Prompt 规则 | 作用 |
|-------------|------|
| "优先调用工具函数获取真实数据，而不是编造" | 防止大模型对平台数据"胡编"，确保用真实 API 数据 |
| "用户没有指定 cluster_id 时，先调用 list_clusters" | 引导大模型在缺少上下文时先查询再操作 |
| "查询类操作直接执行，写操作和删除操作会需要人工审批" | 让大模型在回复中预告审批流程，用户体验更好 |

#### 12.3.2 工具描述语义匹配

每个工具都有精确的 `Description`，大模型通过语义匹配来决定是否调用：

```go
// 定义在 internal/app/services/ai_tools.go → BuildAllTools()
{
    Name: "list_pods",
    Description: "查询指定命名空间下的 Pod 列表",  // ← 用户说"看看Pod"就会匹配
    Parameters: ...
},
{
    Name: "scale_deployment",
    Description: "调整 Deployment 的副本数量(扩缩容)",  // ← 用户说"扩容"就会匹配
    Parameters: ...
},
{
    Name: "delete_namespace",
    Description: "删除整个命名空间(极高危！会删除命名空间下所有资源)",  // ← 用户说"删除namespace"匹配
    Parameters: ...
},
```

大模型的匹配逻辑：

| 用户输入 | 大模型分析 | 匹配结果 |
|----------|-----------|----------|
| "什么是 K8s？" | 通用知识，无需平台数据 | 不调用工具，直接回答 |
| "帮我写个 Nginx 配置" | 代码生成，无需平台数据 | 不调用工具，直接回答 |
| "查看 default 的 Pod" | 语义匹配→"查询 Pod 列表" | 调用 `list_pods` |
| "nginx 有几个副本？" | 语义匹配→"获取 Deployment 详情" | 调用 `get_deployment_detail` |
| "扩容 nginx 到 5" | 语义匹配→"调整副本数量" | 调用 `scale_deployment` |
| "为什么 Pod 在 CrashLoop？" | 需要诊断→先查日志和事件 | 调用 `get_pod_logs` + `get_events` |

### 12.4 第二层：后端风险管控 (toolRegistry)

大模型决定调用工具后，后端有独立的**安全管控层**，不信任大模型的判断：

```go
// 定义在 internal/app/services/ai_tools.go
var toolRegistry = map[string]ToolMeta{
    // read 级：直接执行
    "list_pods":         {RiskLevel: "read",     NeedApproval: false},
    "get_pod_logs":      {RiskLevel: "read",     NeedApproval: false},
    // write 级：需审批
    "scale_deployment":  {RiskLevel: "write",    NeedApproval: true},
    "restart_deployment":{RiskLevel: "write",    NeedApproval: true},
    // danger 级：强制审批
    "delete_pod":        {RiskLevel: "danger",   NeedApproval: true},
    "delete_deployment": {RiskLevel: "danger",   NeedApproval: true},
    // critical 级：强制审批
    "delete_namespace":  {RiskLevel: "critical", NeedApproval: true},
    "drain_node":        {RiskLevel: "critical", NeedApproval: true},
}
```

后端处理逻辑（`ai_assistant.go` 核心代码）：

```go
for _, tc := range result.ToolCalls {
    toolName := tc.Function
    meta, _ := GetToolMeta(toolName)

    if meta.NeedApproval {
        // 写操作/高危操作 → 不直接执行，创建审批请求
        approval, _ := s.createToolApproval(ctx, convID, userID, tc, meta)
        // 告知大模型："该操作已提交审批，等待管理员确认"
    } else {
        // 只读操作 → 直接执行，将结果反馈给大模型
        result, _ := s.ExecuteToolCall(ctx, factory, toolName, tc.Args)
        // 大模型根据真实数据整理成用户友好的回复
    }
}
```

### 12.5 第三层：人工审批

当工具触发审批后，操作不会立即执行：

1. 系统自动创建 `ai_approval_requests` 记录
2. 前端审批管理页面实时显示待审批列表
3. 管理员查看操作详情（工具名、参数、风险等级）
4. 管理员点击「通过」→ 系统在后台异步执行工具调用
5. 管理员点击「拒绝」→ 操作取消，不执行
6. 30 分钟未处理 → 审批自动过期

### 12.6 为什么不用规则引擎/关键词匹配？

| 方案 | 优点 | 缺点 |
|------|------|------|
| **关键词匹配** (if "查看" → 查询) | 简单、可控 | 无法处理复杂表述："nginx 怎么了"、"帮我看看为什么挂了" |
| **正则/NLP 规则引擎** | 相对精准 | 维护成本极高，难以覆盖自然语言的多样性 |
| **Function Calling (当前方案)** | 大模型原生支持，语义理解能力强 | 依赖大模型质量，需配合后端安全管控 |

当前方案的优势：

1. **零维护成本** — 不需要手写任何意图识别规则，大模型自动理解
2. **自然语言支持** — 用户可以用任意表述方式提问
3. **多工具组合** — 大模型可以一次调用多个工具（如：先查日志再查事件来诊断问题）
4. **双重安全** — 即使大模型判断错误，后端 toolRegistry 也会拦截危险操作

### 12.7 意图识别准确性保障

| 保障措施 | 说明 |
|----------|------|
| **精确的工具描述** | 每个工具的 Description 用中文精确描述功能，帮助大模型准确匹配 |
| **参数 Schema** | 每个工具定义了严格的 JSON Schema（必填/选填参数），大模型会按规范提取参数 |
| **System Prompt 引导** | 明确告知大模型：查询用工具获取真实数据、缺少参数先查询、写操作需审批 |
| **历史上下文** | 保留最近 20 轮对话，大模型可以根据上下文理解指代关系（"刚才那个 Pod"） |
| **循环上限** | 最多 5 轮 Function Calling 循环，防止大模型陷入死循环 |
| **未注册工具默认高危** | `IsHighRisk()` 对未注册的工具名返回 true，防止大模型"发明"新工具绕过管控 |
