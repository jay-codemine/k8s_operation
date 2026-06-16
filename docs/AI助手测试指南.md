# K8s AI 助手测试指南

本文档说明 AI 助手的三种提问模式及测试方法，帮助快速验证各功能是否正常。

---

## 架构概览

```
用户提问 → needToolCalling() 本地意图预判
           ├── 不含平台关键词 → 普通 Chat（无工具，快速回复）
           └── 含平台关键词 → selectRelevantTools() 智能筛选
                              ├── 查询类 → 只读工具子集（~20个）
                              ├── 写操作 → 写工具子集 + list_clusters
                              └── 删除类 → 删除工具子集 + list_clusters
                                           ↓
                              ChatWithTools() → Function Calling
                                           ↓
                              ├── 只读工具 → 直接执行，返回结果
                              ├── 写操作工具（NeedApproval） → 创建审批请求
                              └── 高危工具（NeedApproval） → 创建审批请求
```

---

## 一、常规问答（不触发工具）

**特点**：AI 直接用自身知识回答，不调用平台工具，速度最快（约 20-30s）

### 测试用例

| # | 提问内容 | 预期行为 |
|---|---------|---------|
| 1 | `你好` | 返回问候语 + 功能介绍 |
| 2 | `hello` | 同上（英文问候） |
| 3 | `Kubernetes 中 Pod 和 Deployment 有什么区别？` | 返回 K8s 知识对比 |
| 4 | `什么是 Service Mesh？` | 返回概念解释 |
| 5 | `如何排查 Pod CrashLoopBackOff？` | 返回排查步骤 |
| 6 | `解释一下 K8s 的 RBAC 权限模型` | 返回架构知识 |
| 7 | `Helm 和 Kustomize 哪个更好？` | 返回对比分析 |
| 8 | `CI/CD 最佳实践有哪些？` | 返回最佳实践列表 |

### 验证要点

- ✅ `need_approval` = `false`
- ✅ `tools_called` = 空数组或无此字段
- ✅ 响应时间 < 60s
- ✅ 回复内容为 Markdown 格式的专业回答

---

## 二、平台操作（触发 Function Calling 查询）

**特点**：AI 调用工具函数查询平台真实数据，涉及只读操作，速度较慢（约 30-90s）

### 测试用例

| # | 提问内容 | 触发工具 | 预期结果 |
|---|---------|---------|---------|
| 1 | `查看所有集群` | `list_clusters` | 返回集群列表表格 |
| 2 | `查看 default 命名空间的 Pod` | `list_pods` | 返回 Pod 列表 |
| 3 | `查看集群1的所有节点` | `list_nodes` | 返回节点列表 |
| 4 | `查看所有命名空间` | `list_namespaces` | 返回 Namespace 列表 |
| 5 | `查看 default 下的 Deployment` | `list_deployments` | 返回 Deployment 列表 |
| 6 | `查看 default 下的 Service` | `list_services` | 返回 Service 列表 |
| 7 | `查看 default 的事件` | `get_events` | 返回 Event 列表 |
| 8 | `查看流水线列表` | `list_pipelines` | 返回 CI/CD 流水线列表 |
| 9 | `查看集群1 default 下 nginx 的详细信息` | `get_pod_detail` | 返回 Pod 详情 |

### 验证要点

- ✅ `need_approval` = `false`
- ✅ `tools_called` 包含对应的工具名
- ✅ 回复中包含从平台查到的真实数据
- ✅ 数据以 Markdown 表格或列表格式展示

---

## 三、高危操作（触发审批流程）

**特点**：AI 识别为危险/写操作，创建审批请求等待管理员确认，**不会直接执行**

### 3.1 写操作（risk_level: write）

| # | 提问内容 | 触发工具 | 风险等级 |
|---|---------|---------|---------|
| 1 | `把 default 下的 nginx 扩容到 5 个副本` | `scale_deployment` | write |
| 2 | `重启 default 下的 web-app deployment` | `restart_deployment` | write |
| 3 | `回滚 default 下的 nginx deployment` | `rollback_deployment` | write |
| 4 | `更新 default 下 nginx 的镜像为 nginx:1.25` | `update_deployment_image` | write |
| 5 | `触发流水线 1 构建` | `trigger_pipeline` | write |
| 6 | `把节点 node-1 设为不可调度` | `cordon_node` | write |

### 3.2 高危删除操作（risk_level: danger / critical）

| # | 提问内容 | 触发工具 | 风险等级 |
|---|---------|---------|---------|
| 1 | `删除 default 命名空间下的 nginx pod` | `delete_pod` | danger |
| 2 | `删除 default 下的 nginx deployment` | `delete_deployment` | danger |
| 3 | `删除 default 下的 nginx service` | `delete_service` | danger |
| 4 | `删除 test 命名空间` | `delete_namespace` | **critical** |
| 5 | `排空节点 node-1` | `drain_node` | **critical** |
| 6 | `删除 default 下的 app-config configmap` | `delete_configmap` | danger |
| 7 | `删除 default 下的 data-pvc` | `delete_pvc` | danger |

### 验证要点

- ✅ `need_approval` = `true`
- ✅ `approval_id` > 0（生成了审批记录）
- ✅ `pending_tools` 中包含正确的工具名和风险等级
- ✅ 回复中提示"已提交审批"相关文字
- ✅ **操作未被实际执行**（审批通过前）

---

## 四、审批流程验证

### 4.1 查看审批详情

```
GET /api/v1/ai/approvals/{id}
```

验证字段：
- `intent` = 工具名（如 `delete_pod`）
- `risk_level` = 正确的风险等级
- `resource_name` / `namespace` / `cluster_id` 正确解析
- `status` = 1（pending）
- `executed` = false

### 4.2 拒绝审批

```
POST /api/v1/ai/approvals/{id}/reject
Body: {"comment": "不允许此操作"}
```

验证：
- ✅ 状态变为 3（rejected）
- ✅ `executed` = false
- ✅ 审批日志新增 `reject` 记录

### 4.3 通过审批

```
POST /api/v1/ai/approvals/{id}/approve
Body: {"comment": "同意执行"}
```

验证：
- ✅ 状态变为 2（approved）
- ✅ `executed` = true（自动异步执行）
- ✅ `execute_result` 有内容（执行结果 JSON）
- ✅ 审批日志新增 `approve` + `execute` 两条记录

### 4.4 取消审批（申请人取消）

```
POST /api/v1/ai/approvals/{id}/cancel
```

验证：
- ✅ 状态变为 5（canceled）
- ✅ `executed` = false

---

## 五、审批状态码说明

| 状态值 | 含义 | 说明 |
|-------|------|------|
| 1 | pending | 待审批 |
| 2 | approved | 已通过（自动执行） |
| 3 | rejected | 已拒绝 |
| 4 | expired | 已过期（默认30分钟） |
| 5 | canceled | 已取消（申请人取消） |

---

## 六、风险等级说明

| 等级 | 标识 | 说明 | 操作类型 |
|-----|------|------|---------|
| 只读 | `read` | 无风险，直接执行 | 查询资源列表/详情 |
| 写操作 | `write` | 中风险，需审批 | 扩缩容、重启、回滚、创建 |
| 危险 | `danger` | 高风险，需审批 | 删除 Pod/Deployment/Service 等 |
| 极危 | `critical` | 极高风险，需审批 | 删除 Namespace、排空节点 |

---

## 七、常见问题排查

### Q1: 请求超时（timeout exceeded）
- **原因**：NVIDIA NIM + MiniMax M2.7 模型响应较慢，带工具的请求可能需要 30-90s
- **解决**：前端超时已设为 200s，后端 180s；如仍超时，检查网络或换用更快的模型

### Q2: AI 不调用工具，只回复文字
- **原因**：SystemPrompt 中缺少工具使用指令
- **解决**：确保 `ai_assistant.go` 中 `toolUsageInstruction` 正确追加到 SystemPrompt

### Q3: 模型收到乱码（??????）
- **原因**：前端发送中文时编码不正确
- **解决**：确保请求 Content-Type 包含 `charset=utf-8`

### Q4: Function Calling 返回 400 错误
- **原因**：历史消息中 assistant 的 tool_calls 丢失，API 要求 tool 消息前必须有对应的 assistant tool_calls
- **解决**：`buildHistoryMessages()` 已跳过 tool 角色消息

---

## 八、API 快速参考

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/ai/chat` | POST | AI 对话（Function Calling） |
| `/api/v1/ai/chat/stream` | POST | 流式对话（SSE） |
| `/api/v1/ai/quick-ask` | POST | 快捷问答（无工具） |
| `/api/v1/ai/status` | GET | AI 状态检查 |
| `/api/v1/ai/models` | GET | 可用模型列表 |
| `/api/v1/ai/conversations` | GET | 会话列表 |
| `/api/v1/ai/conversations/{id}/messages` | GET | 会话消息历史 |
| `/api/v1/ai/approvals` | GET | 审批列表 |
| `/api/v1/ai/approvals/{id}` | GET | 审批详情 |
| `/api/v1/ai/approvals/{id}/approve` | POST | 通过审批 |
| `/api/v1/ai/approvals/{id}/reject` | POST | 拒绝审批 |
| `/api/v1/ai/approvals/{id}/cancel` | POST | 取消审批 |
| `/api/v1/ai/approvals/pending-count` | GET | 待审批数量 |
| `/api/v1/ai/logs` | GET | AI 日志查询 |
