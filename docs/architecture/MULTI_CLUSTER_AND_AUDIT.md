# 多集群管理与审计系统架构设计

## 目录
- [一、多集群管理](#一多集群管理)
  - [1.1 核心设计理念](#11-核心设计理念)
  - [1.2 数据模型](#12-数据模型)
  - [1.3 Kubeconfig 加密存储](#13-kubeconfig-加密存储)
  - [1.4 ClusterClientFactory 客户端工厂](#14-clusterclientfactory-客户端工厂)
  - [1.5 ClusterMiddleware 请求级集群切换](#15-clustermiddleware-请求级集群切换)
  - [1.6 路由架构](#16-路由架构)
  - [1.7 完整请求链路](#17-完整请求链路)
  - [1.8 健康检查](#18-健康检查)
  - [1.9 架构图示](#19-架构图示)
  - [1.10 关键设计决策](#110-关键设计决策)
- [二、审计系统](#二审计系统)
  - [2.1 核心设计理念](#21-核心设计理念)
  - [2.2 数据模型](#22-数据模型)
  - [2.3 中间件链路](#23-中间件链路)
  - [2.4 敏感数据脱敏](#24-敏感数据脱敏)
  - [2.5 异步写入机制](#25-异步写入机制)
  - [2.6 查询与过滤](#26-查询与过滤)
  - [2.7 保留策略与自动清理](#27-保留策略与自动清理)
  - [2.8 路由解析](#28-路由解析)
  - [2.9 架构图示](#29-架构图示)
  - [2.10 已知问题与改进方向](#210-已知问题与改进方向)

---

## 一、多集群管理

### 1.1 核心设计理念

**所有集群平等对待，不区分 in-cluster 与 external cluster。**

平台不预设"管理集群"和"被管理集群"的区别——每个集群都通过上传 kubeconfig 注册到 `kube_cluster` 表中（AES-256-GCM 加密存储），所有集群统一走相同的客户端构建和缓存流程。

唯一的特殊概念是 `DefaultClusterID`（config.yaml 配置），启动时连接该集群并将客户端存入 `global.ManagementKubeClient` 作为后台任务的默认连接。

### 1.2 数据模型

**文件**: `internal/domain/k8s/models.go:18-34`

```go
type Cluster struct {
    ID             uint32  // 主键
    ClusterName    string  // 唯一名称
    ClusterVersion string  // 如 "v1.36.1"
    KubeConfig     string  // AES-256-GCM 加密的 kubeconfig（json:"-"，永不出现在 API 响应中）
    Status         uint8   // 0=正常 1=异常 2=待确认
    CreatedAt      uint64
    ModifiedAt     uint64  // 用作缓存的版本戳
    DeletedAt      uint64
    LastCheckAt    uint64  // 健康检查时间戳
    LastError      string  // 最近连通性错误
    IsDel          uint8   // 软删除标记
}
```

**关键设计**: `ModifiedAt` 是客户端工厂缓存的**版本戳**。仅当 kubeconfig 或集群名称变化时更新；健康检查更新 `LastCheckAt`/`LastError`/`Status` 不触碰 `ModifiedAt`，避免每次健康检查都使缓存失效。

### 1.3 Kubeconfig 加密存储

#### 加密流程

**文件**: `pkg/utils/crypto.go:22-98`, `pkg/utils/kubeconfig.go:50-82`

```
明文 kubeconfig
  → SHA-256(key) 生成 32 字节 AES-256 密钥
  → 随机 nonce
  → AES-256-GCM 加密（带认证标签）
  → nonce + 密文 → Base64
  → 前缀 "ENC:"（区分加密/非加密数据）
```

#### 解密兼容

`DecodeKubeconfigSmart()` 支持三种格式：
| 格式 | 前缀 | 处理方式 |
|---|---|---|
| 加密 | `ENC:` | AES-256-GCM 解密 |
| 明文 | `apiVersion:` 或 `{` | 直接使用 |
| Base64（遗留）| 其他 | Base64 解码尝试，失败返回原文 |

#### 密钥来源

**文件**: `initialize/setting.go:142-160`

读取 `config.yaml` 中 `Security.KubeConfigEncryptKey`，默认值 `"k8s-operation-default-secret-key"`。生产环境**必须**修改。

#### 写入路径

**文件**: `internal/domain/k8s/cluster_crud.go:17-20, 50-65`

- 新建集群: `Cluster.SetKubeConfig(plaintext)` → encrypt + set
- 更新集群: 同新建，重置健康状态 `Status=0, LastError=""`
- 数据迁移: `Cluster.EncryptKubeConfigIfNeeded(tx)` → 只加密未加密数据（`IsKubeConfigEncrypted()` 检查）

### 1.4 ClusterClientFactory 客户端工厂

**文件**: `internal/domain/k8s/cluster_factory.go`

这是整个多集群管理的**核心引擎**。单例创建，所有请求共享。

#### 结构

```
ClusterClientFactory
├── provider       ClusterClientProvider   // 获取集群数据、构建客户端
├── m              map[uint32]*cachedClients  // clusterID → 缓存
├── g              singleflight.Group      // 去重并发构建
├── failures       map[uint32]failureRecord    // 故障负缓存
├── baseTTL        30 分钟（可配）
├── jitterRange    3 分钟（随机分散过期时间）
├── connectTimeout 15 秒（可配）
├── failureTTL     20 秒（故障快速失败窗口）
└── logger         *logger.Logger
```

#### Get() 方法完整流程

```
Get(ctx, clusterID)
  │
  ├─ 1. provider.GetCluster(ctx, clusterID)
  │     → 返回加密的 Cluster（只读 ModifiedAt 做版本对比）
  │
  ├─ 2. 检查缓存
  │     ├─ 命中（version 匹配 + 未过期）→ 直接返回，无解密开销
  │     └─ 未命中 → 继续
  │
  ├─ 3. 检查 negative cache
  │     └─ failureTTL 内→ 直接返回错误，不重试
  │
  ├─ 4. singleflight 去重
  │     key = "{clusterID}:{version}"
  │     并发请求共享一次构建
  │
  ├─ 5. provider.BuildClientsForCluster(ctx, clusterID)
  │     │
  │     ├─ 解密 kubeconfig（KubeConfig.Decrypt()）
  │     ├─ 解析 YAML → rest.Config
  │     ├─ TuneRESTConfig(cfg)   // QPS=50, Burst=100, Timeout=30s, Insecure=true
  │     ├─ kubernetes.NewForConfig(cfg)
  │     ├─ ServerVersion() 连通性验证 ← 失败则整体失败
  │     ├─ dynamic.NewForConfig(cfg)
  │     └─ metricsclient.NewForConfig(cfg)  // 可选，失败不阻塞
  │
  ├─ 6. 成功 → 缓存 K8sClients（TTL: 30min + rand(0~3min)）
  │              清除 failure record
  │
  └─ 7. 失败 → Invalidate(clusterID), 标记 failure
               更新 DB 健康状态（LastError, Status）
```

#### 缓存策略总结

| 场景 | 行为 |
|---|---|
| 正常命中 | 直接返回，0 解密开销 |
| 缓存过期 | singleflight 重建，并发共享 |
| 集群刚故障 | failureTTL 20s 内快速返回错误 |
| kubeconfig 变更 | ModifiedAt 变化 → 版本不匹配 → 自动重建 |
| 健康状态变更 | 不影响 ModifiedAt → 缓存保持有效 |

#### Provider 接口

**文件**: `internal/domain/k8s/interfaces.go:1-13`

```go
type ClusterClientProvider interface {
    GetCluster(ctx, clusterID) (*Cluster, error)               // 加密路径，版本检查用
    BuildClientsForCluster(ctx, clusterID) (*K8sClients, error) // 解密+构建
}
```

`Services` 结构体实现了该接口，工厂创建时注入自身。

### 1.5 ClusterMiddleware 请求级集群切换

**文件**: `middlewares/cluster.go:104-220`

每个目标集群资源请求到达时，中间件完成认证→鉴权→客户端获取→上下文注入。

#### 执行流程

```
1. 提取 cluster_id（Header X-Cluster-ID 或 query 参数）
2. 解析为 uint32
3. 鉴权：inferClusterAction(c) 推断操作类型
   CheckClusterPermission(userID, clusterID, action) RBAC 检查
4. 获取客户端：factory.Get(ctx, clusterID)
5. 错误映射：
   ├─ 404 → 集群不存在
   ├─ 403 → 无权限
   ├─ 503 → 连接失败/超时（通用错误，不泄漏内部信息）
   └─ 500 → 其他错误
6. 注入上下文：ctx.Set("k8s_clients", clients)
                 ctx.Set("cluster_id", clusterID)
```

#### 操作推断

**文件**: `middlewares/cluster.go:250-287`

基于 URL 路径和 HTTP 方法推断 RBAC 操作：

| 检测规则 | RBAC Action |
|---|---|
| 路径含 `terminal`/`exec` | `exec` |
| 路径含 `delete`/`drain`/`evict` | `delete` |
| 路径含 `create` | `create` |
| 路径含 `update`/`patch`/`apply`/`scale`/`restart`/`rollback` | `update` |
| GET 请求 | `view` |
| POST 请求 | `create` |
| PUT/PATCH 请求 | `update` |
| DELETE 请求 | `delete` |

### 1.6 路由架构

**文件**: `initialize/router.go:222-427`

三类路由：

| 路由类别 | 路径 | 中间件 | 说明 |
|---|---|---|---|
| 平台集群管理 | `/api/v1/k8s/cluster/*` | JWT only | 集群 CRUD，操作 kube_cluster 表 |
| 目标集群资源 | `/api/v1/k8s/pod/*` 等 | JWT + ClusterMiddleware | 通过 X-Cluster-ID 切换集群 |
| CICD 路由 | `/api/v1/k8s/cicd/*` | JWT only | 内部用 ResolveClusterClients() |

### 1.7 完整请求链路

以 `GET /api/v1/k8s/pod/list?cluster_id=5` 为例：

```
┌─────────────────────────────────────────────────────────────┐
│ 客户端: GET /api/v1/k8s/pod/list?cluster_id=5              │
│ Header: Authorization: Bearer <jwt>, X-Cluster-ID: 5       │
└──────────────────────────┬──────────────────────────────────┘
                           │
    ┌──────────────────────────────────────────────────────┐
    │ 1. CORS middleware     → 允许 X-Cluster-ID 头         │
    │ 2. Prometheus          → 请求计数                     │
    │ 3. RateLimit           → IP 级别限流                   │
    │ 4. Logger              → 请求日志                     │
    │ 5. Recovery            → panic 恢复                   │
    │ 6. LicenseGate         → 许可证检查（可选）            │
    └──────────────────────────────────────────────────────┘
                           │
    ┌──────────────────────────────────────────────────────┐
    │ 7. AuthJWT middleware                                  │
    │    → 解析 token → ctx.Set("user_id", 123)              │
    │    → ctx.Set("tenant_id", 1)                          │
    └──────────────────────────────────────────────────────┘
                           │
    ┌──────────────────────────────────────────────────────┐
    │ 8. TenantScope middleware                              │
    │    → 创建租户隔离 DB 连接 ScopedDB                      │
    │    → ctx.Set("scoped_db", db)                         │
    └──────────────────────────────────────────────────────┘
                           │
    ┌──────────────────────────────────────────────────────┐
    │ 9. ClusterMiddleware(factory)                          │
    │    a. 提取 cluster_id=5                                │
    │    b. 获取 user_id=123                                 │
    │    c. 推断 action="view" (GET方法)                     │
    │    d. CheckClusterPermission(123, 5, "view")           │
    │       → 查询 sys_user_cluster 表                       │
    │       → 验证 access_level ≥ read                      │
    │    e. factory.Get(ctx, 5):                             │
    │       ├─ GetCluster(5) → version=1689782400            │
    │       ├─ 缓存检查: map[5] 存在? version 匹配? 未过期?    │
    │       ├─ 未命中 → singleflight build                   │
    │       │   ├─ 解密 kubeconfig                           │
    │       │   ├─ parse → rest.Config                       │
    │       │   ├─ TuneRESTConfig                            │
    │       │   ├─ kubernetes.NewForConfig                   │
    │       │   ├─ ServerVersion() 验证                      │
    │       │   └─ 缓存 (30min + jitter)                     │
    │       └─ 返回 *K8sClients                              │
    │    f. ctx.Set("k8s_clients", clients)                  │
    │    g. ctx.Set("cluster_id", 5)                         │
    └──────────────────────────────────────────────────────┘
                           │
    ┌──────────────────────────────────────────────────────┐
    │ 10. Pod Controller                                     │
    │     → 从 ctx 取 k8s_clients                            │
    │     → clients.Kube.CoreV1().Pods(namespace).List()     │
    │     → 返回 Pod 列表                                     │
    └──────────────────────────────────────────────────────┘
```

### 1.8 健康检查

**文件**: `internal/app/services/platform_health.go`

周期检查所有集群连通性（在 Dashboard 的集群健康页面触发）：

```
CheckClusterConnectivity(clusterID)
  ├─ factory.ResetFailure(clusterID)   // 清除负缓存
  ├─ factory.GetClient(ctx, clusterID) // 强制重新连接
  ├─ 成功 → 更新 Status=0, LastError=""
  └─ 失败 → 更新 Status=1, LastError=错误信息
          → 不更新 ModifiedAt（保护缓存）
```

并发检查：每个集群独立的 goroutine，10 秒超时，panic 恢复。

### 1.9 架构图示

```
                         ┌──────────────────┐
                         │  global.Kube*    │
                         │  (管理集群客户端)   │
                         │  启动时连接         │
                         └────────┬─────────┘
                                  │
┌─────────────┐          ┌───────┴─────────┐
│  Controller  │          │  Bootstrap      │
│  (HTTP 层)   │          │  InitAll()      │
└──────┬───────┘          └─────────────────┘
       │
       │ ServicesFromContext(ctx)  ← TenantScope 中间件注入
       │
┌──────┴──────────────────────────────────────────────────┐
│                  ClusterClientFactory                     │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │  cache: map[clusterID]*cachedClients                │ │
│  │  ├── 1: {version: 123, expires: 30m}              │ │
│  │  ├── 5: {version: 456, expires: 28m}              │ │
│  │  └── 8: {version: 789, expires: 25m}              │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │  singleflight.Group                                 │ │
│  │  去重并发构建 {clusterID}:{version}                  │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │  failures: map[clusterID]failureRecord              │ │
│  │  负缓存 20s → 快速失败，避免雪崩                      │ │
│  └────────────────────────────────────────────────────┘ │
└────────────┬─────────────────────────────────────────────┘
             │
    ┌────────┴────────────┐
    │                     │
    ▼                     ▼
┌───────────┐     ┌───────────────┐
│ 集群 1     │     │ 集群 5         │
│ (管理集群)  │     │ (本地 kind)    │
│ K8s v1.36 │     │ K8s v1.36     │
│ 172.18.0.2│     │ 127.0.0.1:    │
│           │     │   28659       │
└───────────┘     └───────────────┘
```

### 1.10 关键设计决策

| 决策 | 原因 |
|---|---|
| 不区分 in-cluster/external | 统一模型，减少分支逻辑 |
| 单例工厂 + 共享缓存 | 多请求复用连接，singleflight 防雪崩 |
| 缓存 key = clusterID + ModifiedAt | kubeconfig 变更自动失效 |
| 健康检查不更新 ModifiedAt | 避免健康状态变更导致频繁缓存重建 |
| negative cache 20s | 故障集群快速失败，不阻塞其他请求 |
| ENC: 前缀区分加密数据 | 向后兼容遗留明文数据 |
| 中间件级鉴权 | 在到达 Controller 前拦截，不污染业务逻辑 |

---

## 二、审计系统

### 2.1 核心设计理念

**全局中间件拦截 + 异步写入 + 自动保留清理。**

审计中间件注册在 Gin Engine 顶层，对所有请求生效（按需跳过 GET 请求和高频端点）。写操作在独立 goroutine 中异步完成，不阻塞请求响应。

### 2.2 数据模型

**文件**: `internal/domain/audit/models.go:6-36`

`audit_log` 表（MySQL）：

| 字段 | 类型 | 索引 | 说明 |
|---|---|---|---|
| `id` | bigint PK | PRIMARY | 自增 |
| `user_id` | bigint | idx_audit_user | 0=未认证 |
| `username` | varchar(191) | | 操作时的用户名快照 |
| `user_ip` | varchar(50) | | 客户端 IP |
| `user_agent` | varchar(500) | | 浏览器 UA |
| `action` | varchar(100) | idx_audit_action | create/update/delete/login/logout/exec 等 |
| `action_display` | varchar(191) | | 中文显示名（如"创建 Deployment"） |
| `module` | varchar(100) | idx_audit_module | auth/cluster/workload/network/cicd/rbac 等 |
| `target_type` | varchar(100) | idx_audit_target | 资源类型 |
| `target_id` | varchar(100) | idx_audit_target | 资源 ID |
| `target_name` | varchar(191) | | 从 URL 提取 |
| `request_uri` | varchar(500) | | 请求路径 |
| `request_method` | varchar(10) | | HTTP 方法 |
| `request_body` | text | | 脱敏后，最大 4KB |
| `response_code` | int | | HTTP 状态码 |
| `response_message` | varchar(500) | | 预留字段（未用） |
| `detail` | json | | 预留字段（未用） |
| `extra` | json | | 预留字段（未用） |
| `cluster_id` | *int64 | idx_audit_cluster | X-Cluster-ID |
| `cluster_name` | varchar(191) | | 预留字段（未用） |
| `namespace` | varchar(100) | | URL 参数 |
| `pipeline_id` | *int64 | idx_audit_pipeline | 预留字段（未用） |
| `pipeline_name` | varchar(191) | | 预留字段（未用） |
| `project_id` | *int64 | | 预留字段（未用） |
| `project_name` | varchar(191) | | 预留字段（未用） |
| `status` | varchar(50) | idx_audit_status | success / failed |
| `error_message` | varchar(1000) | | 失败时 = "HTTP {code}" |
| `duration_ms` | int | | 请求耗时（毫秒） |
| `created_at` | int64 | idx_audit_created | Unix 时间戳 |
| `tenant_id` | int | idx_tenant_id | DB 层面存在，Go 模型未映射 |

### 2.3 中间件链路

**文件**: `middlewares/audit.go`

#### 注册位置

**文件**: `initialize/server.go:95`

```go
s.Use(middlewares.Audit(nil)) // 引擎顶层，全局生效
```

#### 跳过规则

```go
defaultConfig := AuditConfig{
    RecordGET: false,     // GET 请求不记录
    ExcludePaths: []string{
        "/api/v1/platform/health",   // 健康检查
        "/api/v1/monitoring/",       // 监控轮询
        "/swagger/",                 // API 文档
        "/api/v1/ai/chat",           // AI 对话（高频）
    },
    MaxBodySize: 4096,    // 请求体最大 4KB
}
```

跳过条件：
- GET 请求且 `RecordGET==false`（默认）
- OPTIONS 请求（CORS 预检）
- 路径匹配 ExcludePaths 前缀

#### 请求体捕获

```
对非 GET + 非 multipart 请求:
  ├─ 读取请求体（最多 MaxBodySize 字节）
  ├─ io.NopCloser 恢复请求体（下游可继续读）
  ├─ sanitizeBody() 脱敏
  └─ 截断 → "请求体(太长已截断)...[truncated]"
```

### 2.4 敏感数据脱敏

**文件**: `middlewares/audit.go:342-366`

```go
func sanitizeBody(body []byte) []byte {
    // 对 JSON 请求体递归 redact 以下 key：
    // password, secret, token, kube_config, kubeconfig,
    // access_key_secret, access_key_id, webhook
    // → 值替换为 "***"
}
```

仅对 `application/json` 有效。非 JSON 请求体原样存储。multipart 请求体记录为 `"[multipart/form-data upload]"`。

### 2.5 异步写入机制

**文件**: `middlewares/audit.go:132-206`

```
c.Next()  // 先执行业务逻辑
    │
    ▼
go writeAuditLog(...)  // 独立 goroutine 异步写入，不阻塞响应
    │
    ├─ defer recover()          // panic 恢复
    ├─ global.DB == nil → return // DB 未就绪则放弃
    ├─ 提取 user_id & username
    ├─ parseRouteInfo(path, method) → module, action, targetType, actionDisplay
    ├─ 解析 cluster_id, namespace
    ├─ status = "success" (code<400) / "failed" (code≥400)
    ├─ 构造 AuditLog 结构体
    └─ NewBackgroundServices().AuditLogRecord(ctx, log)
       → 直接使用 global.DB 写入（无租户作用域）
```

### 2.6 查询与过滤

**文件**: `internal/domain/audit/models.go:51-66`, `internal/infra/persistence/audit_repo.go:54-134`

#### 支持的过滤维度

| 维度 | 匹配方式 | 说明 |
|---|---|---|
| user_id | 精确 | |
| username | LIKE 模糊 | |
| action | 精确 | create/update/delete/login/logout/exec/approve/reject/deploy/scale/view |
| module | 精确 | auth/cluster/workload/network/config/storage/cicd/rbac/platform/ai/monitoring/image |
| target_type | 精确 | 资源类型 |
| status | 精确 | success / failed |
| cluster_id | 精确 | |
| keyword | LIKE 模糊 | 同时搜索 target_name, action_display, request_uri, error_message |
| start_time / end_time | 范围 | created_at 时间范围 |
| sort_field | 白名单 | 仅: created_at, duration_ms, user_id, action, module, status |
| sort_order | | asc / desc |

#### API 端点

```
GET    /api/v1/platform/audit/logs        → 列表查询
GET    /api/v1/platform/audit/logs/:id    → 单条详情
GET    /api/v1/platform/audit/statistics  → 统计概览
GET    /api/v1/platform/audit/retention   → 获取保留策略
PUT    /api/v1/platform/audit/retention   → 更新保留策略
POST   /api/v1/platform/audit/cleanup     → 手动触发清理
GET    /api/v1/platform/audit/export      → 导出 CSV
```

#### 统计概览

`QueryStatistics()` 一次查询返回：
- `total_today` — 今日操作数
- `total_week` — 本周操作数
- `total_all` — 总计
- `success_rate` — 成功率
- `top_users` — 近一周 Top 5 活跃用户
- `top_modules` — 近一周 Top 5 活跃模块
- `action_summary` — 今日操作类型分布
- `hourly_counts` — 今日每小时操作数曲线

### 2.7 保留策略与自动清理

**文件**: `internal/domain/audit/service.go:42-81`, `internal/app/worker/audit_cleanup.go`

#### 策略存储

保留策略存在 `platform_settings` 表中：
```json
{"category": "security", "key": "audit_retention", "value": "30"}
```
`value=0` 表示永久保留。

#### 自动清理 Worker

```
启动: 应用启动后 5 分钟首次执行
周期: 每 24 小时
超时: 60 秒
流程:
  1. 读取保留策略
  2. 若永久或 days==0 → 跳过
  3. DELETE FROM audit_log WHERE created_at < (now - retention_days)
```

#### 手动清理

前端 AuditLog 页面 → 保留策略设置 → 手动清理按钮 → 立即触发 Cleanup。

### 2.8 路由解析

**文件**: `middlewares/audit.go:209-316`

`parseRouteInfo(path, method)` 使用路径子串匹配推导结构化审计信息：

```
/api/v1/auth/login               → module=auth,     action=login,    action_display=用户登录
/api/v1/k8s/pod/list             → module=workload, action=view,     action_display=查看 Pod 列表
/api/v1/k8s/deployment/create    → module=workload, action=create,   action_display=创建 Deployment
/api/v1/k8s/cluster/create       → module=cluster,  action=create,   action_display=创建集群
/api/v1/k8s/cicd/pipeline/create → module=cicd,     action=create,   action_display=创建流水线
/api/v1/k8s/rbac/role/update     → module=rbac,     action=update,   action_display=更新角色
/api/v1/platform/settings/update → module=platform,  action=update,   action_display=更新平台设置
```

`methodToAction` 映射：
```
POST           → "create"
PUT / PATCH    → "update"
DELETE         → "delete"
everything else → "view"
```

### 2.9 架构图示

```
                        HTTP 请求进入
                             │
              ┌──────────────┴──────────────────────────────┐
              │          Gin Engine                           │
              │                                               │
              │  ┌──────────────────────────────────────┐     │
              │  │ middlewares.Audit(nil)                │     │
              │  │   1. 记录 start 时间                   │     │
              │  │   2. 读取 body (最多 4KB)             │     │
              │  │   3. sanitizeBody 脱敏                │     │
              │  │   4. 恢复 body                        │     │
              │  │   5. c.Next() → 执行业务逻辑           │     │
              │  │   6. 提取响应码、耗时                  │     │
              │  │   7. 判断 shouldSkip                  │     │
              │  │   8. go writeAuditLog(...)  异步写入   │     │
              │  └──────────────────────────────────────┘     │
              │                                               │
              │  ┌──────────────────────────────────────┐     │
              │  │ AuthJWT → TenantScope → Router        │     │
              │  └──────────────────────────────────────┘     │
              └──────────────────────────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              │         异步写入               │
              │                               │
              │  go writeAuditLog():          │
              │    ├─ parseRouteInfo()       │
              │    │  └─ 推导 module/action   │
              │    ├─ 构造 AuditLog            │
              │    ├─ NewBackgroundServices() │
              │    └─ AuditLogRecord()        │
              │       └─ repo.Save()          │
              │                              │
              └──────────────┬───────────────┘
                             │
                    ┌────────┴─────────┐
                    │  MySQL            │
                    │  audit_log 表     │
                    └──────────────────┘

              ┌──────────────────────────────────────┐
              │  自动清理 Worker                       │
              │                                       │
              │  每 24h 执行:                          │
              │    1. 读取保留策略                      │
              │    2. DELETE WHERE created_at < cutoff │
              └──────────────────────────────────────┘
```

### 2.10 已知问题与改进方向

#### 当前问题

| 问题 | 状态 | 修复内容 |
|---|---|---|
| Go 模型缺 `TenantID` 字段 | ✅ 已修复 | `internal/domain/audit/models.go:8` — 添加 `TenantID uint32` 字段，gorm tag 与 DB 列匹配 |
| 异步写入用 `global.DB` 无租户上下文 | ✅ 已修复 | `middlewares/audit.go` — 在 `c.Next()` 后提取 `tenant_id`，传入 `writeAuditLog` 并写入 AuditLog |
| `ResponseMessage` 未填充 | ✅ 已修复 | 从 `c.Errors` 收集错误信息，写入 `response_message` 列 |
| `ClusterName` 未填充 | ✅ 已修复 | `writeAuditLog` 中通过 `Services.GetCluster()` 异步查询集群名称 |
| `PipelineName` 未填充 | 待处理 | 需从 CICD 上下文提取 pipeline_id 后查询 |
| `Detail`, `Extra` 字段未使用 | 按需 | Controller 需要在设置响应后手动追加业务上下文 |

#### 扩展方向

1. **PipelineName/ProjectName 填充**: 在 CICD 路由中设置 pipeline_id/project_id 到 gin.Context，审计中间件自动查询名称
2. **超管跨租户审计查看**: 读取路径判断 `is_super_admin`，跳过租户过滤
3. **敏感数据脱敏增强**: 支持 XML/protobuf 格式，支持 nested key 递归匹配
4. **审计告警**: 基于规则引擎检测异常模式（如短时间大量删除操作）
5. **结构化 Detail/Extra**: Controller 在设置响应后可以附加业务上下文
6. **审计事件总线**: 将关键写操作事件发送到外部系统（Kafka/Webhook）
