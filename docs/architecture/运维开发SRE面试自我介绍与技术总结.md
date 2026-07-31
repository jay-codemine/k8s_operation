# 运维开发 / SRE 面试自我介绍与技术总结

> 基于 K8sOperation 全栈运维平台实战经验

---

## 一、一句话定位

> 独立设计并开发了一个**企业级 K8s 全栈运维平台**，集成多集群管理 + CI/CD 发布编排 + AI 智能运维，覆盖从代码提交到生产发布的全链路自动化。

---

## 二、项目概述（30 秒电梯演讲）

我做的 K8sOperation 是一个**一站式 K8s 运维平台**，核心解决三个痛点：

1. **多集群割裂**：用 Hub 模式统一管理多个 K8s 集群，一个控制平面操作所有集群，RBAC 细粒度权限隔离
2. **CI/CD 混乱**：用 4 个通用 Jenkins Job 替代 N 个项目各自维护 Jenkinsfile，构建到部署全自动
3. **运维门槛高**：AI 助手让用户用自然语言操作 K8s，AI 告警分析 + 日志诊断 + 定时巡检

技术栈：**Go + Gin + client-go + GORM + Vue3 + MySQL + Redis + Jenkins + ArgoCD + DeepSeek LLM**

---

## 三、技术亮点（面试重点展开）

### 3.1 多集群管理 —— 考点：K8s 架构、client-go、安全

**怎么做的：**
- 每个目标集群的 kubeconfig **AES-256-GCM 加密**存在 DB，`json:"-"` 永不返回前端，只在 DAO 层解密
- `ClusterClientFactory`：按集群 ID 懒加载 `*kubernetes.Clientset`，正缓存 30min + 负缓存 20s（故障快速拒绝），`singleflight` 防止并发初始化同一个集群
- 中间件 `ClusterMiddleware`：解析 `X-Cluster-ID` Header → RBAC 检查 → 工厂获取客户端 → 注入 Gin Context → Controller 直接用

**为什么这么做：**
- 一个集群挂了不影响其他集群（负缓存 + 故障隔离）
- 缓存用 `modified_at` 做版本号，kubeconfig 变更自动失效
- `singleflight` 防止 100 个并发请求同时初始化同一个集群

**面试延伸：**
> 如果集群数量到 100+，怎么优化？
> — 工厂缓存已支持 30min TTL + 随机抖动防雪崩；进一步可以按 region 分组部署多实例，每实例管理就近集群；配置下发可以考虑用 etcd 做 config watch 替代 DB 轮询

### 3.2 CI/CD 发布系统 —— 考点：设计模式、分布式、可靠性

**一个模板服务 100 个项目：**
- 不像传统 Jenkins 一个项目一个 Job，我用 **4 个通用 Groovy 模板**（Go/Java/前端/Python），所有差异通过 Jenkins 参数传入
- 前端 3 步向导创建流水线 → 自动触发首次构建 → Jenkins 回调 → 平台自动部署

**双保险状态同步：**
- **回调（主）**：Jenkins 通过 HMAC-SHA256 签名回调平台
- **轮询（备）**：`PipelinePollWorker` 每 10s 轮询 Jenkins，处理回调丢失、构建卡死、孤儿流水线
- 回调处理全部幂等（`CallbackReceived=1` 跳过重入）

**不可变镜像晋级：**
- 构建产物用 `image@sha256:digest` 固定，同一 Digest 从 dev → test → staging → prod
- 不是每个环境重新构建，是同一个镜像逐级放行

**面试延伸：**
> 为什么不用 Jenkinsfile 而用平台管理？
> — Jenkinsfile 散落在各仓库，运维不可见、不可控。平台集中管理后：① 批量改所有流水线的构建参数 ② 统一质量门禁策略 ③ 构建日志和部署状态全留存 ④ 非开发人员（测试/运营）也能看懂和执行发布

> 如何保证回调不丢？
> — 回调层：HMAC 签名 + 幂等。轮询层：PollWorker 10s 间隔兜底。如果平台宕机：Jenkins 重试 HTTP 回调（模板内置 retry），恢复后 PollWorker 立即扫到遗漏的构建

### 3.3 金丝雀发布 —— 考点：K8s、Prometheus、渐进式交付

**实现：**
1. 创建金丝雀 Deployment（`name-canary`），configurable 副本数 + 流量比
2. 后台协程每 30s 通过 Prometheus 评估分析规则（错误率、延迟、成功率）
3. 决策：失败→自动回滚 / 成功+autoPromote→全量 / 成功-manual→等审批

**面试延伸：**
> 和 Argo Rollouts 的区别？
> — Argo Rollouts 是 CRD 驱动的标准方案。我选择自研的原因：① 与平台审批流 + 钉钉通知深度整合 ② 分析规则 UI 表单化，不需要写 YAML ③ 多集群环境一个控制平面统一调度金丝雀

### 3.4 审批流 + 飞书闭环 —— 考点：系统设计、安全

**三级审批：**
1. 平台内审批（管理员 approve/reject，前端页操作）
2. 飞书交互卡片（构建完成后自动发卡片，一键审批，支持多级审批）
3. 发布单级审批（环境定义决定是否需审批，多级审批链）

**安全设计：**
- 禁止自审（requester ≠ approver）
- 审批 30min 过期
- 飞书回调 HMAC 签名（`timestamp + "\n" + secret`）
- 回滚发布单**跳过审批**（紧急处理），但记录完整审计

### 3.5 AI 智能运维 —— 考点：LLM 应用、安全设计

**Function Calling 架构：**
- 30+ K8s 工具：只读工具（直接执行）、写入工具（需审批）、高危工具（删除/驱逐，强制审批）
- 智能路由：根据用户消息关键词选取工具子集（查 Pod 不发删除工具，省 Token）
- FC 循环上限 5 轮，防止死循环
- AI 模块完全解耦：LLM 挂了、Key 过期不影响平台核心功能

**AIOps：**
- 告警分析：自动关联 PromQL + 告警上下文，LLM 输出根因 + 优先级 + kubectl 命令
- 日志诊断：Loki 查询 → 错误行优先采样 → LLM 分析异常模式
- 定时巡检（每 6h）：收集全平台健康数据 → 扣分制评分 → LLM 综合评估

**多提供商架构：**
- Registry 模式：DeepSeek / 通义千问 / 智谱 GLM 随意切换，按模型缓存客户端
- APIKey 永不暴露前端，`ListProviders` 只返回元数据

**面试延伸：**
> AI 操作 K8s 安全吗？
> — 4 级风险控制：read→直接执行 / write→需审批 / danger→需审批 / critical→强制审批。工具执行复用平台的 Service 层，没有 AI 专属特权路径。所有审批操作全量审计。

### 3.6 Kubeconfig 安全 —— 考点：加密、密钥管理

- AES-256-GCM 加密，密钥来自 config.yaml `Security.KubeConfigEncryptKey`
- 存入格式：`"ENC:" + base64(ciphertext)`
- 兼容三种格式：ENC 加密 / 明文 YAML / 旧版 base64（`DecodeKubeconfigSmart`）
- DAO 层是唯一解密点，`json:"-"` 永不序列化返回前端
- `mustFromDB()` 返回明文 `*rest.Config` 直接建客户端，明文不落地到任何日志或中间变量

---

## 四、系统架构一句话总结

```
前端 Vue3 → Nginx → Go/Gin → ClusterMiddleware(X-Cluster-ID+RBAC)
                                ↓
                         ClusterClientFactory(缓存+singleflight+故障隔离)
                                ↓
                         client-go → 目标 K8s API Server

CI/CD: 平台管理配置 → 参数化 Jenkins Job → 回调+轮询 → 自动部署/金丝雀/回滚
AI:   用户自然语言 → DeepSeek Function Calling → 工具执行/审批 → K8s 操作
```

---

## 五、项目量化指标（面试用）

| 指标 | 数据 |
|------|------|
| 支持语言类型 | 4 种（Go/Java/前端/Python） |
| Jenkins Job 数量 | 仅 4 个通用 Job（vs 传统 N 个项目 N 个 Job） |
| 部署模式 | 3 种（Push 直接部署 / Pull GitOps / 金丝雀渐进） |
| AI 工具数量 | 30+ K8s 操作工具 |
| 支持 LLM | 所有 OpenAI 兼容 API（DeepSeek/千问/GLM/...） |
| 审批通道 | 3 条（平台内 / 飞书卡片 / 发布单级） |
| 回滚机制 | 4 层（阶段级 / 发布单级 / 自动回滚 / 镜像 Digest） |
| 代码规模 | 后端 150+ 文件 / 前端 50+ 页面组件 |
| 中间件 | 12 层（JWT/Tenant/Cluster/Audit/RateLimit/Prometheus/...） |

---

## 六、面试常见追问预案

**Q1: 为什么不用 Rancher/OpenShift 等现成方案？**
> 业务需要与钉钉/飞书、审批流、AI 助手深度整合，现成方案二次开发成本更高。自研让我们可以完全控制数据流、安全模型和 UX。

**Q2: 单体应用怎么拆分微服务？**
> 当前是模块化单体（controller → service → dao 清晰分层）。拆分路径：先把 AI 模块和 CICD Worker 独立部署（已解耦），再按子域（集群管理、工作负载、CI/CD、监控）拆分。共享 DB 先不变，逐步按表归属拆。

**Q3: 怎么保证高可用？**
> ① 多实例部署 + Nginx upstream ② Redis 会话共享 ③ PipelinePollWorker 单实例（可加分布式锁迁移到多实例）④ 集群客户端独立故障隔离，一个集群挂了不影响主控平面

**Q4: 最难的技术点？**
> ClusterClientFactory 的并发安全设计。需要同时处理：缓存 TTL / 版本失效 / 负缓存 / singleflight 合并 / 连接超时。用 channel + select 实现超时控制，用 `sync.RWMutex` 保护缓存 map，singleflight 的 key 包含 `clusterID:version` 防止过期缓存复用。

**Q5: 你在这个项目中最大的成长？**
> 从"写功能"到"设计系统"。多集群管理让我深入理解了 K8s 的认证模型和 client-go 的机制，CI/CD 让我掌握了分布式系统的可靠性和幂等设计，AI 助手让我实践了 LLM 应用的工程落地。

---

## 七、个人技术能力总结

| 方向 | 具体能力 |
|------|---------|
| **K8s** | client-go 二次开发、多集群管理、RBAC、资源编排、金丝雀发布 |
| **Go 后端** | Gin 框架、GORM、中间件链、并发安全、单飞模式、Redis Stream |
| **CI/CD** | Jenkins Pipeline (Groovy)、Argo Workflows、ArgoCD、Harbor、Kaniko、Docker |
| **前端** | Vue3 + Vite + Pinia + ArcoDesign 组件库 |
| **AI/LLM** | OpenAI Function Calling、多提供商注册表、Prompt Engineering、安全设计 |
| **可观测性** | Prometheus + Grafana + Loki + 钉钉/飞书通知 |
| **安全** | AES-256-GCM 加密、HMAC 签名、JWT 认证、RBAC 权限、审计日志 |
| **系统设计** | 工厂模式、缓存策略、故障隔离、幂等设计、模块解耦 |
