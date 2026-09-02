# 多集群管理与多集群发布指南

> 本文聚焦说明平台**如何纳管多个 K8s 集群**，以及**如何把应用发布到多个集群**，并给出关键实现文件清单。
> 更底层的实现细节见《MULTI_CLUSTER_AND_AUDIT.md》与《CICD_多集群与混合云发布能力总览.md》。

---

## 一、总览

多集群能力分两个独立层次：

| 层次 | 解决的问题 | 核心机制 |
|---|---|---|
| **多集群管理** | 如何接入、切换、操作多个集群 | kubeconfig 加密注册 + 客户端工厂 + 请求级切换 |
| **多集群发布** | 如何把应用发布到（多个）集群 | Pipeline 单集群 + Release 多集群分发 |

---

## 二、多集群管理

核心链路：**kubeconfig 加密注册 → 客户端工厂缓存 → 中间件按请求切换**

### 2.1 集群接入

- 每个集群通过**上传 kubeconfig** 注册，存到 `kube_cluster` 表。
- kubeconfig 用 **AES-256-GCM 加密**后落库（`pkg/utils/crypto.go`），带 `ENC:` 前缀，`json:"-"` 永不出现在 API 响应中。
- 设计理念：**不区分 in-cluster 与 external cluster**，所有集群统一模型。

### 2.2 客户端工厂（多集群的核心引擎）

`ClusterClientFactory`（`internal/domain/k8s/cluster_factory.go`）单例共享，负责把 `clusterID` 变成可用的 K8s 客户端：

| 特性 | 作用 |
|---|---|
| 缓存（30min + 抖动） | 命中直接返回，0 解密开销 |
| singleflight 去重 | 并发请求同一集群只构建一次，防雷群效应 |
| 负缓存（20s） | 故障集群快速失败，不拖垮其他请求 |
| 版本戳失效 | kubeconfig 变更（ModifiedAt 变化）自动重建客户端 |

构建流程：`Get(clusterID)` → 查缓存 → singleflight → 解密 kubeconfig → 解析 rest.Config → `ServerVersion()` 连通性验证 → 缓存客户端。

### 2.3 请求级集群切换

`ClusterMiddleware`（`middlewares/cluster.go`）：每个目标集群资源请求到达时：

```
提取 X-Cluster-ID (header/query) → RBAC 鉴权 → factory.Get(clusterID) → 注入 context
```

路由架构（`initialize/router.go`）：

| 路由类别 | 路径 | 说明 |
|---|---|---|
| 平台集群管理 | `/api/v1/k8s/cluster/*` | 集群 CRUD，操作 kube_cluster 表 |
| 目标集群资源 | `/api/v1/k8s/pod/*` 等 | 通过 X-Cluster-ID 切换集群 |
| CICD 路由 | `/api/v1/k8s/cicd/*` | 内部用 ResolveClusterClients() |

### 2.4 三层 RBAC 权限

平台级（超管/普通用户）→ 集群级（集群管理员/只读）→ 命名空间级，通过 `sys_user_cluster` 表做用户-集群授权。

### 2.5 健康检查

周期检查所有集群连通性，更新 `Status`（OK/Bad/Pending）与 `LastError`，不触碰 `ModifiedAt`（保护客户端缓存）。

---

## 三、多集群发布

支持**两种发布模式**：

### 3.1 Pipeline 模式（单集群自动部署）

适用「一个应用 → 一个集群」的常见场景：

```
Jenkins 构建成功 → HMAC 回调平台 → 审批(可选) → 按 TargetClusterID 部署到目标集群
```

流水线模型存 `TargetClusterID`，回调后 `autoDeployToK8sWithResult()` 根据该 ID 取客户端 → Patch 镜像 → waitRollout → 通知结果。

### 3.2 Release 模式（多集群分发）⭐

适用「一个镜像 → 同时部署到多个集群」（生产多活、灾备、混合云）：

```
创建发布单(ClusterIDs=[1,2,3])
  → 为每个集群生成独立 Task
  → 投入 Redis Stream 消息队列 (cicd:deploy:stream)
  → 多个 Worker 并发消费
  → 各自集群 Patch 镜像 + waitRollout
```

发布单请求模型：

```go
type CicdReleaseCreateRequest struct {
    AppName       string  // 应用名称
    Namespace     string  // 命名空间
    WorkloadKind  string  // 工作负载类型
    WorkloadName  string  // 工作负载名称
    ContainerName string  // 容器名称
    Strategy      string  // 发布策略（rolling）
    TimeoutSec    uint32  // 超时时间
    Concurrency   uint32  // 并发数
    ImageRepo     string  // 镜像仓库
    ImageTag      string  // 镜像标签
    ClusterIDs    []int64 // 目标集群 ID 列表（一次发布到多个集群）
}
```

### 3.3 环境分级（配合多集群）

`dev / test / staging / prod` 四级环境，**每个环境绑定不同集群**：

```
DEV     → 本地 K3s
TEST    → 公司内网 K8s
STAGING → 阿里云 ACK
PROD    → AWS EKS（强制审批）
```

---

## 四、关键实现文件清单

### 多集群管理

| 文件 | 职责 |
|---|---|
| `internal/domain/k8s/models.go` | 集群数据模型（Cluster） |
| `internal/domain/k8s/cluster_factory.go` | 客户端工厂（缓存 + singleflight） |
| `internal/domain/k8s/cluster_crud.go` | 集群 CRUD + kubeconfig 加密 |
| `internal/app/services/k8s_cluster.go` | 集群服务 + 健康检查 |
| `middlewares/cluster.go` | ClusterMiddleware 请求级切换 |
| `pkg/utils/crypto.go` | AES-256 加解密 |
| `pkg/utils/kubeconfig.go` | kubeconfig 解析/加密兼容 |

### 多集群发布

| 文件 | 职责 |
|---|---|
| `internal/app/requests/cicd_release.go` | 发布单请求（含 ClusterIDs） |
| `internal/app/services/cicd_release.go` | 发布单创建 + Redis 入队 + 状态管理 |
| `internal/app/builder/cicd_task_builder.go` | 按 ClusterIDs 拆分 Task |
| `internal/app/services/cicd_executor.go` | 部署执行器（Patch + waitRollout） |
| `internal/app/worker/cicd_worker.go` | Worker 消费 Redis Stream |
| `internal/app/services/cicd_pipeline.go` | 流水线 CRUD + 回调 + 单集群部署 |

---

## 五、典型部署拓扑

### 单集群（入门）

```
Jenkins → K8sOperation 平台 → K8s 集群(all-in-one)
```

### 多集群（企业级）

```
                  ┌──▶ DEV 集群（本地 K3s）
Jenkins → 平台 ───┼──▶ STAGING 集群（阿里云 ACK）
                  └──▶ PROD 集群（AWS EKS）
```

### 混合云多活（大规模）

```
                  ┌──▶ AWS EKS（北美区）
Jenkins → 平台 ───┼──▶ 阿里云 ACK（亚太区）
 + Redis          ├──▶ Azure AKS（欧洲区）
 + Worker x N     └──▶ 私有云 K8s（灾备）
```

---

## 六、小结

| 问题 | 答案 |
|---|---|
| 多集群怎么管理 | kubeconfig AES 加密注册 + ClusterClientFactory（缓存/singleflight）+ ClusterMiddleware 按 X-Cluster-ID 切换 + 三层 RBAC |
| 多集群怎么发布 | 单集群用 Pipeline 模式（TargetClusterID）；多集群用 Release 模式（ClusterIDs 列表 + Redis Stream + Worker 并发分发） |
