# CICD 多集群与混合云发布能力总览

> 本文档全面梳理 K8sOperation 平台在 **CI/CD 全链路**、**多集群治理** 和 **混合云发布** 方面的设计与实现。
> 对标 Rancher / KubeSphere 企业级能力。

---

## 一、平台定位

K8sOperation 是一个企业级 Kubernetes 多集群管理平台，核心定位：

| 定位 | 说明 |
|------|------|
| **多集群统一治理** | 公有云 / 私有云 / 边缘集群统一纳管，一套平台管理所有集群 |
| **CI/CD 发布控制** | 全链路流水线 + 多环境分级 + 人工审批 + 自动部署 |
| **三层 RBAC 权限** | 平台级 → 集群级 → 命名空间级，精细化权限管控 |
| **全面资源管理** | Deployment / StatefulSet / DaemonSet / Job / CronJob / Ingress / Service 等 |

---

## 二、架构全景

```
                          ┌─────────────────────────────────┐
                          │         开发者 Push 代码          │
                          └──────────────┬──────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                         Jenkins 流水线（14 阶段标准化）                           │
│                                                                                  │
│  ① 拉取代码 → ② 安装依赖 → ③ 编译构建 → ④ 单元测试 → ⑤ 代码扫描               │
│  → ⑥ SonarQube 分析 → ⑦ 质量门禁 → ⑧ 构建制品 → ⑨ 上传制品库                  │
│  → ⑩ 打包镜像 → ⑪ 推送镜像仓库 → ⑫ 审批（可选） → ⑬ 部署                      │
│                                                                                  │
└──────────────────────────────────┬──────────────────────────────────────────────┘
                                   │ Jenkins 回调（HMAC 签名验证）
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          K8sOperation 平台后端                                   │
│                                                                                  │
│   ┌────────────────┐      ┌──────────────┐      ┌──────────────────┐           │
│   │  Pipeline 模式  │      │ Release 模式  │      │  环境管理         │           │
│   │  单集群自动部署  │      │ 多集群分发     │      │  dev/test/staging │           │
│   │  构建→审批→部署  │      │ Redis Stream  │      │  /prod           │           │
│   └───────┬────────┘      └──────┬───────┘      └──────────────────┘           │
│           │                      │                                               │
│           ▼                      ▼                                               │
│   ┌──────────────────────────────────────────────────────────────┐              │
│   │              ClusterClientFactory（集群客户端工厂）             │              │
│   │   缓存 + singleflight + TTL 抖动 + 快速失败 + 优雅降级        │              │
│   └──────┬───────────────┬───────────────┬───────────────────────┘              │
│          │               │               │                                       │
└──────────┼───────────────┼───────────────┼───────────────────────────────────────┘
           ▼               ▼               ▼
   ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
   │  公有云 EKS   │ │  公有云 AKS   │ │  私有云 K8s   │ │  边缘 K3s    │
   │  (AWS)       │ │  (Azure)     │ │  (Rancher)   │ │  (IoT)       │
   └──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

---

## 三、多集群治理层

### 3.1 集群接入（混合云支持）

平台通过 **kubeconfig** 统一接入各类 K8s 集群，无论底层基础设施是什么：

| 云环境 | 集群类型 | 接入方式 |
|--------|---------|---------|
| **公有云** | AWS EKS / Azure AKS / GCP GKE | 导入 kubeconfig |
| **私有云** | OpenShift / Rancher RKE | 导入 kubeconfig |
| **本地** | kubeadm / K3s / Kind | 导入 kubeconfig |
| **边缘** | K3s / MicroK8s | 导入 kubeconfig |

**核心实现文件**：

| 文件 | 职责 |
|------|------|
| `internal/app/services/k8s_cluster.go` | 集群 CRUD、客户端构建、健康检查 |
| `internal/app/services/k8s_cluster_factory.go` | 多集群客户端工厂（缓存 + singleflight） |
| `internal/app/dao/k8s_cluster.go` | 集群数据持久化（AES 加密 kubeconfig） |
| `pkg/utils/crypto.go` | AES-256 加密/解密 kubeconfig |
| `middlewares/cluster.go` | ClusterMiddleware - 请求级集群路由 |
| `global/k8s.go` | 管理集群全局变量（与业务集群隔离） |

### 3.2 安全架构

```
┌───────────────────────────────────────────────────┐
│                   安全层                            │
├───────────────────────────────────────────────────┤
│ 1. kubeconfig AES-256 加密存储（数据库落盘）         │
│ 2. TLS 动态信任（支持自签名证书集群）                │
│ 3. 管理集群与业务集群客户端隔离                      │
│ 4. 请求级集群路由（ClusterMiddleware）               │
│ 5. 三层 RBAC 权限模型                               │
│    ├── 平台级：超级管理员 / 普通用户                  │
│    ├── 集群级：集群管理员 / 只读用户                  │
│    └── 命名空间级：资源操作权限                       │
└───────────────────────────────────────────────────┘
```

### 3.3 集群客户端工厂

参考 Rancher / KubeSphere 设计的高可用客户端管理：

| 特性 | 实现 |
|------|------|
| **缓存优先** | 命中缓存直接返回，避免重复连接 |
| **singleflight** | 防止并发创建同一集群客户端（雷群效应） |
| **TTL + 抖动** | 缓存过期时间随机化，防止缓存雪崩 |
| **快速失败** | 单个集群连接超时（5s）不影响其他集群 |
| **健康状态** | 连接结果写入数据库（OK / Bad / Pending） |

---

## 四、CI/CD 流水线层

### 4.1 14 阶段标准化流水线

平台实现了企业级标准的 **14 阶段全链路**流水线，支持 Go / Java / Frontend / Python 四种语言：

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                        14 阶段标准化 CI/CD 流水线                              │
├───────┬───────┬───────┬───────┬───────┬───────┬────────────────────────────┤
│ 阶段  │ 名称   │ 类型   │ 必需   │ 可配置  │ 可阻断  │ 说明                     │
├───────┼───────┼───────┼───────┼───────┼───────┼────────────────────────────┤
│  ①   │ 拉取代码  │ CI │ ✅ │ 分支   │ ✅ │ Git clone               │
│  ②   │ 安装依赖  │ CI │ ✅ │ -     │ ✅ │ go mod / mvn / npm       │
│  ③   │ 编译构建  │ CI │ ✅ │ -     │ ✅ │ go build / mvn / npm build│
│  ④   │ 单元测试  │ CI │ 可选│ SKIP_TESTS │ ✅ │ go test / mvn test   │
│  ⑤   │ 代码扫描  │ CI │ 可选│ ENABLE_SONAR │ ❌ │ 静态分析           │
│  ⑥   │ SonarQube │ CI │ 可选│ SONAR_*│ ✅ │ sonar-scanner / mvn sonar│
│  ⑦   │ 质量门禁  │ CI │ 可选│ QUALITY_GATE │ ✅ │ waitForQualityGate() │
│  ⑧   │ 构建制品  │ CI │ 可选│ -     │ ✅ │ 打包二进制/JAR/dist       │
│  ⑨   │ 上传制品  │ CI │ 可选│ -     │ ❌ │ curl 上传制品库 API       │
│  ⑩   │ 打包镜像  │ CD │ ✅ │ Dockerfile │ ✅ │ docker build          │
│  ⑪   │ 推送镜像  │ CD │ ✅ │ 仓库地址 │ ✅ │ docker push             │
│  ⑫   │ 审批     │ CD │ 可选│ 审批人  │ ✅ │ 钉钉通知 + 人工审批       │
│  ⑬   │ 部署     │ CD │ ✅ │ 集群/NS │ ✅ │ Patch 镜像 + waitRollout │
│  ⑭   │ 通知     │ CD │ ✅ │ Webhook │ ❌ │ 钉钉/企微结果通知         │
└───────┴───────┴───────┴───────┴───────┴───────┴────────────────────────────┘
```

### 4.2 多语言 Jenkins 模板

| 语言 | 模板文件 | 特点 |
|------|---------|------|
| **Go** | `configs/jenkins-templates/go-pipeline.groovy` | go mod + go build + sonar-scanner |
| **Java** | `configs/jenkins-templates/java-spring-pipeline.groovy` | Maven + mvn sonar:sonar |
| **Frontend** | `configs/jenkins-templates/frontend-pipeline.groovy` | npm + Nginx 镜像 |
| **Python** | `configs/jenkins-templates/python-pipeline.groovy` | pip + sonar-scanner |

### 4.3 流水线数据模型

**核心实现文件**：

| 文件 | 职责 |
|------|------|
| `internal/app/models/cicd_pipeline.go` | 流水线模型（含部署配置） |
| `internal/app/services/cicd_pipeline.go` | 流水线 CRUD + 触发 + 回调处理 |
| `internal/app/services/cicd_stage.go` | 阶段管理 + 阶段状态推进 |
| `internal/app/services/cicd_notify.go` | 钉钉通知（构建/审批/部署） |
| `internal/app/requests/cicd_pipeline.go` | 请求参数定义 |

**流水线关键字段**：

```go
type CicdPipeline struct {
    // 部署配置
    AutoDeploy         bool   // 是否自动部署
    TargetClusterID    int64  // 目标集群 ID（指向哪个 K8s 集群）
    TargetNamespace    string // 目标命名空间
    TargetWorkloadKind string // 工作负载类型（Deployment/StatefulSet/DaemonSet）
    TargetWorkloadName string // 工作负载名称
    TargetContainer    string // 容器名称
    DeployEnv          string // 部署环境（dev/test/staging/prod）
    RequireApproval    bool   // 是否需要审批
    EnableSonar        bool   // 是否启用 SonarQube 代码扫描
}
```

---

## 五、发布模式

平台支持两种发布模式，覆盖从简单到复杂的发布场景：

### 5.1 Pipeline 模式（单集群自动部署）

适用场景：单个应用部署到单个集群（最常见的 CI/CD 场景）

```
┌────────────┐     ┌──────────┐     ┌──────────┐     ┌──────────────┐
│ Jenkins 构建 │ ──▶ │ HMAC 回调 │ ──▶ │ 审批（可选）│ ──▶ │ 自动部署到    │
│ 成功        │     │ 平台后端   │     │ 钉钉通知   │     │ 目标集群      │
└────────────┘     └──────────┘     └──────────┘     └──────────────┘
```

**执行流程**：
1. Jenkins 构建完成后通过 HMAC 签名回调平台
2. 平台验证签名、更新构建状态
3. 若 `require_approval=true`，发送钉钉审批通知
4. 审批通过后（或无需审批），调用 `autoDeployToK8sWithResult()`
5. 根据 `TargetClusterID` 获取集群客户端
6. Patch 工作负载镜像 → waitRollout → 钉钉通知结果

**核心代码链路**：
```
PipelineCallback() → autoDeployToK8sWithResult() → executeAutoDeployAsync()
  → K8sClusterInit(clusterID) → PatchDeploymentImage() → waitRollout()
```

### 5.2 Release 模式（多集群分发）

适用场景：同一镜像需要同时部署到多个集群（生产多活、灾备、混合云）

```
┌────────────┐     ┌──────────────────────────────────────────┐
│ 创建发布单   │ ──▶ │ BuildCicdReleaseTasks(clusterIDs)         │
│ ClusterIDs: │     │ 为每个集群生成独立 Task                     │
│ [1, 2, 3]   │     └──────────────────┬───────────────────────┘
└────────────┘                         │
                                       ▼
                    ┌──────────────────────────────────────┐
                    │       Redis Stream 消息队列            │
                    │       cicd:deploy:stream              │
                    │  ┌──────┐ ┌──────┐ ┌──────┐          │
                    │  │Task1 │ │Task2 │ │Task3 │          │
                    │  │集群1 │ │集群2 │ │集群3 │          │
                    │  └──────┘ └──────┘ └──────┘          │
                    └────────────────┬─────────────────────┘
                                     │
                    ┌────────────────┼────────────────┐
                    ▼                ▼                ▼
             ┌──────────┐     ┌──────────┐     ┌──────────┐
             │ Worker 1  │     │ Worker 2  │     │ Worker 3  │
             │ 执行 Task1│     │ 执行 Task2│     │ 执行 Task3│
             └──────────┘     └──────────┘     └──────────┘
                    │                │                │
                    ▼                ▼                ▼
             ┌──────────┐     ┌──────────┐     ┌──────────┐
             │ K8s 集群1 │     │ K8s 集群2 │     │ K8s 集群3 │
             │ Patch+Wait│     │ Patch+Wait│     │ Patch+Wait│
             └──────────┘     └──────────┘     └──────────┘
```

**核心实现文件**：

| 文件 | 职责 |
|------|------|
| `internal/app/requests/cicd_release.go` | 发布单请求（含 `ClusterIDs []int64`） |
| `internal/app/services/cicd_release.go` | 发布单创建 + Redis 入队 + 状态管理 |
| `internal/app/builder/cicd_task_builder.go` | 为每个集群生成独立 Task |
| `internal/app/services/cicd_executor.go` | 部署执行器（Patch + waitRollout） |
| `internal/app/worker/cicd_worker.go` | Worker 进程（消费 Redis Stream） |

**发布单数据模型**：

```go
// 创建发布单请求
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

---

## 六、环境分级管理

### 6.1 四级环境模型

```
┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐
│   🔧 DEV  │ ──▶│  🧪 TEST │ ──▶│  📦 STAGING│ ──▶│  🚀 PROD │
│   开发环境 │    │  测试环境  │    │  预发环境   │    │  生产环境  │
└──────────┘    └──────────┘    └──────────┘    └──────────┘
```

### 6.2 各环境管控策略

| 维度 | DEV 开发 | TEST 测试 | STAGING 预发 | PROD 生产 |
|------|---------|----------|-------------|----------|
| **触发方式** | Git Push 自动 | MR 合并自动 | 手动触发 | 手动触发 |
| **代码分支** | `feature/*` | `develop` | `release` | `release` / `main` |
| **单元测试** | 可跳过 | 必须执行 | 必须执行 | 必须执行 |
| **代码扫描** | 关闭 | 开启（不阻断） | 开启（质量门禁） | 严格质量门禁 |
| **人工审批** | 不需要 | 不需要 | 可选 | **强制审批** |
| **审批人** | - | - | 技术负责人 | 技术 + 运维负责人 |
| **部署方式** | 自动部署 | 自动部署 | 自动部署 | 审批后自动部署 |
| **回滚支持** | 不需要 | 手动回滚 | 一键回滚 | 一键回滚 + 自动告警 |
| **通知范围** | 无 | 构建结果 | 全流程通知 | 全流程 + 值班群 |
| **发布窗口** | 随时 | 工作时间 | 工作时间 | 工作日 10:00-17:00 |

### 6.3 环境数据模型

**核心实现文件**：`internal/app/models/cicd_environment.go`

```go
type CicdEnvironment struct {
    Name            string // 环境名称（dev/staging/prod）
    DisplayName     string // 显示名称
    ClusterID       int64  // 关联集群 ID
    Namespace       string // 默认命名空间
    RequireApproval bool   // 是否需要审批
    ApprovalUsers   JSONMap // 审批人员列表
}
```

每个环境可以绑定**不同的集群**，支持典型的混合云部署拓扑：

```
DEV 环境     → 本地 K3s 集群（开发团队自用）
TEST 环境    → 公司内网 K8s 集群
STAGING 环境 → 阿里云 ACK 集群（模拟生产）
PROD 环境    → AWS EKS 集群（线上服务）
```

---

## 七、审批闭环

### 7.1 审批流程

```
构建成功 → 是否需要审批？
    │
    ├── 否 → 直接自动部署
    │
    └── 是 → 发送钉钉审批通知
              │
              ├── 审批通过 → 执行自动部署 → 钉钉通知结果
              │
              └── 审批拒绝 → 记录拒绝原因 → 钉钉通知
```

### 7.2 审批数据模型

**数据库表**：`cicd_deploy_approval`

| 字段 | 说明 |
|------|------|
| `pipeline_id` | 关联流水线 |
| `release_id` | 关联发布单 |
| `env` | 目标环境 |
| `risk_level` | 风险等级（low / medium / high） |
| `risk_warnings` | 风险提示列表 |
| `status` | 审批状态（pending / approved / rejected / expired） |
| `applicant_id` | 申请人 |
| `approver_id` | 审批人 |
| `approve_comment` | 审批意见 |

### 7.3 钉钉通知类型

| 通知类型 | 触发时机 | 内容 |
|---------|---------|------|
| 构建结果 | 构建成功/失败 | 流水线名、分支、耗时、状态 |
| 审批提醒 | 需要审批时 | 流水线名、环境、镜像、操作链接 |
| 部署结果 | 部署成功/失败 | 集群、命名空间、工作负载、镜像、耗时 |

---

## 八、已实现能力矩阵

| 能力维度 | 功能点 | 状态 | 说明 |
|---------|--------|------|------|
| **多集群治理** | 集群注册 & 接入 | ✅ 已实现 | kubeconfig AES 加密存储 |
| | 集群健康检查 | ✅ 已实现 | OK / Bad / Pending 状态机 |
| | 客户端缓存工厂 | ✅ 已实现 | singleflight + TTL 抖动 |
| | TLS 动态信任 | ✅ 已实现 | 支持自签名证书集群 |
| | 管理集群隔离 | ✅ 已实现 | ManagementKubeClient 独立 |
| **CI/CD 流水线** | 14 阶段标准化 | ✅ 已实现 | 拉取 → 部署 全链路 |
| | 多语言支持 | ✅ 已实现 | Go / Java / Frontend / Python |
| | SonarQube 集成 | ✅ 已实现 | 代码扫描 + 质量门禁 |
| | 制品库 | ✅ 已实现 | 构建产物上传 & 管理 |
| | Jenkins 模板化 | ✅ 已实现 | 按语言自动选择模板 |
| **发布管理** | Pipeline 单集群部署 | ✅ 已实现 | 构建 → 审批 → 部署 |
| | Release 多集群分发 | ✅ 已实现 | ClusterIDs + Redis Stream |
| | 滚动更新 | ✅ 已实现 | Patch + waitRollout |
| | 多工作负载类型 | ✅ 已实现 | Deployment / StatefulSet / DaemonSet |
| **环境管理** | 四级环境 | ✅ 已实现 | dev / test / staging / prod |
| | 环境-集群绑定 | ✅ 已实现 | 每个环境独立集群 |
| | 环境级审批策略 | ✅ 已实现 | 生产强制审批 |
| **审批闭环** | 人工审批 | ✅ 已实现 | 钉钉通知 + 风险等级 |
| | 审批过期 | ✅ 已实现 | 超时自动过期 |
| | 操作审计 | ✅ 已实现 | 完整操作留痕 |
| **通知** | 钉钉通知 | ✅ 已实现 | 构建/审批/部署全流程 |
| | Jenkins 链接 | ✅ 已实现 | 通知中嵌入控制台链接 |
| **发布策略** | 滚动更新（Rolling） | ✅ 已实现 | 默认策略 |
| | 金丝雀发布 | 🔜 规划中 | 灰度流量切换 |
| | 蓝绿部署 | 🔜 规划中 | 双版本切换 |

---

## 九、典型部署拓扑

### 9.1 单集群（入门）

```
┌──────────────┐      ┌──────────────┐
│ Jenkins      │ ───▶ │ K8sOperation │ ───▶ ┌──────────────┐
│ 构建服务器    │      │ 平台         │      │ K8s 集群      │
└──────────────┘      └──────────────┘      │ (all-in-one) │
                                             └──────────────┘
```

### 9.2 多集群（企业级）

```
                                             ┌──────────────┐
                                        ┌──▶ │ DEV 集群     │
┌──────────────┐      ┌──────────────┐  │    │ (本地 K3s)   │
│ Jenkins      │ ───▶ │ K8sOperation │  │    └──────────────┘
│ 构建服务器    │      │ 平台         │ ─┤    ┌──────────────┐
└──────────────┘      └──────────────┘  ├──▶ │ STAGING 集群  │
                                        │    │ (阿里云 ACK)  │
                                        │    └──────────────┘
                                        │    ┌──────────────┐
                                        └──▶ │ PROD 集群     │
                                             │ (AWS EKS)    │
                                             └──────────────┘
```

### 9.3 混合云多活（大规模）

```
                                             ┌──────────────┐
                                        ┌──▶ │ AWS EKS      │ (北美区)
                                        │    └──────────────┘
┌──────────────┐      ┌──────────────┐  │    ┌──────────────┐
│ Jenkins      │ ───▶ │ K8sOperation │ ─┼──▶ │ 阿里云 ACK    │ (亚太区)
│ 构建服务器    │      │ + Redis      │  │    └──────────────┘
└──────────────┘      │ + Worker x N │  │    ┌──────────────┐
                      └──────────────┘  ├──▶ │ Azure AKS    │ (欧洲区)
                                        │    └──────────────┘
                                        │    ┌──────────────┐
                                        └──▶ │ 私有云 K8s    │ (灾备)
                                             └──────────────┘
```

---

## 十、配置示例

### 10.1 创建生产流水线

```json
{
  "name": "user-service-prod",
  "description": "用户服务 - 生产环境发布",
  "language_type": "java",
  "git_repo": "https://gitee.com/company/user-service.git",
  "git_branch": "release",
  "deploy_env": "prod",
  "auto_deploy": true,
  "require_approval": true,
  "target_cluster_id": 1,
  "target_namespace": "production",
  "target_workload_kind": "Deployment",
  "target_workload_name": "user-service",
  "target_container": "user-service",
  "env_vars": [
    { "name": "SKIP_TESTS", "value": "false" },
    { "name": "ENABLE_SONAR", "value": "true" },
    { "name": "SONAR_QUALITY_GATE", "value": "true" }
  ]
}
```

### 10.2 创建多集群发布单

```json
{
  "app_name": "user-service",
  "namespace": "production",
  "workload_kind": "Deployment",
  "workload_name": "user-service",
  "container_name": "user-service",
  "strategy": "rolling",
  "timeout_sec": 300,
  "concurrency": 3,
  "image_repo": "registry.company.com/user-service",
  "image_tag": "v1.2.3",
  "cluster_ids": [1, 2, 3]
}
```

### 10.3 平台全局配置

```yaml
# config.yaml
Jenkins:
  URL: "http://jenkins.internal:8080/"
  Username: "admin"
  APIToken: "xxxxxxxxxx"
  TriggerTimeout: 60
  CallbackURL: "http://k8s-platform:10537"
  PlatformURL: "http://k8s-web.internal:30851"
  HMACSecret: "your-hmac-secret-key"
  DingTalkWebhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
```

---

## 十一、相关文档索引

| 文档 | 说明 |
|------|------|
| [CICD_多集群部署架构设计.md](./CICD_多集群部署架构设计.md) | Worker + Redis Stream 技术实现细节 |
| [CICD_平台架构设计文档.md](./CICD_平台架构设计文档.md) | CICD 平台整体架构 |
| [CI_CD_流水线阶段化与通知闭环.md](./CI_CD_流水线阶段化与通知闭环.md) | 14 阶段设计 + 钉钉通知 |
| [SonarQube_代码质量检测设计文档.md](./SonarQube_代码质量检测设计文档.md) | 代码扫描 + 质量门禁 |
| [生产环境CICD发布设计规范.md](./生产环境CICD发布设计规范.md) | 生产环境发布规范 |
| [模板化CICD快速接入指南.md](./模板化CICD快速接入指南.md) | 快速接入教程 |
| [CICD发布与模板化架构设计.md](./CICD发布与模板化架构设计.md) | 模板化设计 |
| [Java项目CICD完整接入指南.md](./Java项目CICD完整接入指南.md) | Java 项目接入 |
| [平台整体架构总览.md](./平台整体架构总览.md) | 平台整体架构 |
