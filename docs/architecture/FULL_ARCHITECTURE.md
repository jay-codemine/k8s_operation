# 全链路架构图——前端到后端

## 一、完整请求生命周期（以"创建集群"为例）

```
用户点击 [创建集群] 按钮
         │
         ▼
┌─────────────────────────────────────────────────────────┐
│  前端 (Vue 3 + Vite)                                    │
│  k8s-web/src/views/ClusterCreate.vue                   │
│                                                         │
│  const resp = await clusterApi.create({                 │
│    cluster_name: '生产集群',                             │
│    cluster_version: 'v1.36',                            │
│    kube_config: '...'                                   │
│  })                                                     │
│                                                         │
│  POST /api/v1/k8s/cluster/create                        │
└────────────────────┬────────────────────────────────────┘
                     │  HTTP POST (JSON + JWT Token)
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Vite Proxy (开发) / Nginx (生产)                       │
│  /api/* → http://localhost:8080                        │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Gin HTTP Server (:8080)                                │
│                                                         │
│  ┌──────────────────────────────────────────────────┐  │
│  │ 中间件链                                         │  │
│  │ 1. AuthJWT()        → 验证 JWT, 查用户, 设 ctx  │  │
│  │ 2. TenantScope()    → 注入租户隔离 DB           │  │
│  │ 3. Audit()          → 异步写审计日志             │  │
│  │ 4. ClusterMiddleware→ 解析集群 ID, 注入 K8sClients│  │
│  └──────────────────────────────────────────────────┘  │
│                         │                               │
│                         ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │ Router: /api/v1/k8s/cluster                      │  │
│  │   POST /create → ClusterController.Create()      │  │
│  └──────────────────────┬───────────────────────────┘  │
└─────────────────────────┬──────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│  Controller 层 (internal/app/controllers/api/v1/)      │
│                                                         │
│  func (c *ClusterController) Create(ctx *gin.Context) { │
│    // 1. 参数绑定                                        │
│    param := requests.NewK8sClusterCreateRequest()      │
│    valid.Validate(ctx, param, ...)                     │
│                                                         │
│    // 2. 获取租户隔离 Services                           │
│    svc := middlewares.NewServicesFromContext(ctx)       │
│                                                         │
│    // 3. 委托应用服务                                     │
│    err := svc.K8sClusterCreate(ctx, param)              │
│                                                         │
│    // 4. 返回响应                                        │
│    resp.Success(nil)                                    │
│  }                                                      │
└────────────────────┬────────────────────────────────────┘
                     │  svc.K8sClusterCreate(ctx, param)
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Services 层 (internal/app/services/)                   │
│                                                         │
│  func (s *Services) K8sClusterCreate(ctx, param) {     │
│    // 1. 请求参数 → 领域参数                              │
│    return s.clusterSvc().Create(                        │
│      ctx,                                               │
│      param.ClusterName,                                 │
│      param.ClusterVersion,                              │
│      param.KubeConfig,                                  │
│    )                                                    │
│  }                                                      │
│                                                         │
│  // Services 持有: db, logger, eventBus, tenantID      │
│  // clusterSvc() 创建 Domain Service，注入依赖            │
│  func (s *Services) clusterSvc() *dm.ClusterService {  │
│    return dm.NewClusterService(                         │
│      persistence.NewClusterRepository(s.db),   ← GORM  │
│      s.logger,                                          │
│      s.eventBus,                               ← Event │
│    )                                                    │
│  }                                                      │
└────────────────────┬────────────────────────────────────┘
                     │  s.clusterSvc().Create(ctx, ...)
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Domain Service 层 (internal/domain/k8s/service.go)    │
│                                                         │
│  func (s *ClusterService) Create(ctx, ...) error {     │
│    // 1. 值对象验证                                      │
│    name, err := NewClusterName(clusterName)             │
│    if err != nil { return err }                         │
│                                                         │
│    // 2. 领域规则: KubeConfig 加密                       │
│    kcCfg, err := NewKubeConfigFromPlain(kubeConfigPlain)│
│    if err != nil { return err }                         │
│                                                         │
│    // 3. 通过 Repository 持久化                          │
│    kc := &Cluster{                                      │
│      ClusterName: name.String(),                        │
│      KubeConfig:  kcCfg.Encrypted(),                   │
│      Status:      0,                                    │
│    }                                                    │
│    if err := s.repo.Save(ctx, kc); err != nil {         │
│      s.logger.Error("创建集群失败", zap.Error(err))      │
│      return err                                         │
│    }                                                    │
│                                                         │
│    // 4. 发布领域事件                                    │
│    s.publish(NewClusterCreated(kc.ID, clusterName))    │
│    return nil                                           │
│  }                                                      │
│                                                         │
│  // ClusterService 持有: repo, logger, publisher       │
└────────────────────┬────────────────────────────────────┘
                     │  s.repo.Save(ctx, kc)
                     ▼
┌─────────────────────────────────────────────────────────┐
│  Repository 接口 (internal/domain/k8s/repository.go)   │
│                                                         │
│  type ClusterRepository interface {                     │
│    Save(ctx context.Context, c *Cluster) error          │
│    FindByID(ctx context.Context, id uint32) (*Cluster, error)│
│  }                                                      │
│  // 纯接口，不依赖任何实现                                │
└────────────────────┬────────────────────────────────────┘
                     │  运行时绑定
                     ▼
┌─────────────────────────────────────────────────────────┐
│  GORM 实现 (internal/infra/persistence/cluster_repo.go) │
│                                                         │
│  type clusterRepo struct { db *gorm.DB }                │
│  func (r *clusterRepo) Save(ctx, c *Cluster) error {   │
│    return r.db.WithContext(ctx).Create(c).Error         │
│  }                                                      │
│                      │                                  │
│                      ▼                                  │
│              ┌──────────────┐                           │
│              │   MySQL      │                           │
│              │ kube_cluster │                           │
│              └──────────────┘                           │
└─────────────────────────────────────────────────────────┘

同时，领域事件异步传播:
                     s.publish(NewClusterCreated(...))
                              │
                              ▼
                    ┌─────────────────┐
                    │    EventBus     │
                    │ Subscribe/Pub   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        [审计日志]     [通知服务]     [监控更新]
        bootstrap/event_handlers.go
```

## 二、分层职责对照表

| 层 | 文件位置 | 职责 | 依赖方向 |
|----|----------|------|---------|
| **前端** | `k8s-web/src/` | UI 渲染、用户交互、API 调用 | → Vite Proxy |
| **Router** | `internal/app/routers/` | URL 映射、中间件绑定 | → Controller |
| **Middleware** | `middlewares/` | 认证、租户隔离、审计、集群解析 | → Services / DB |
| **Controller** | `internal/app/controllers/` | 参数绑定、调用 Services、返回响应 | → Services |
| **Services** | `internal/app/services/` | 编排领域服务、事务、创建 Domain Service 实例 | → Domain Service |
| **Domain Service** | `internal/domain/*/service.go` | 业务规则、值对象验证、聚合根约束 | → Repository(接口) |
| **Repository** | `internal/domain/*/repository.go` | 持久化接口定义 | 无依赖（纯接口） |
| **GORM 实现** | `internal/infra/persistence/` | 数据库操作 | → MySQL |
| **EventBus** | `internal/domain/events/` | 事件发布/订阅 | → 处理器函数 |
| **Adapter** | `internal/infra/adapter/` | 跨域接口实现 | → Domain Service |

## 三、数据流向

```
请求流向 (Request):
  Browser → Vite Proxy → Gin Router → Middleware → Controller
          → Services → Domain Service → Repository(接口) → GORM → MySQL

响应流向 (Response):
  MySQL → GORM → Repository → Domain Service → Services
        → Controller → JSON → Browser

事件流向 (Async):
  Domain Service → EventBus.Publish → EventBus.Subscribe → Handler → Logger/DB/External

依赖注入流向 (DI):
  main.go → bootstrap → Services → persistence.NewXxxRepository(global.DB)
                                   → events.NewEventBus(global.Logger)
                                   → adapter.NewXxx(domainSvc)
```

## 四、以 CMDB 为例的完整文件链路

```
新增 CMDB 域，用户创建一个资产：

前端                             后端
─────                            ─────
AssetCreate.vue                  Router
  POST /api/v1/cmdb/asset/create   → /cmdb/asset
    { hostname, ip, ... }            → AssetController.Create()
                                       │
                                       │ param 绑定
                                       ▼
                                     svc.AssetCreate(req)
                                       │
                                       │ Services hook
                                       ▼
                                     s.cmdbSvc().Create(ctx, req)
                                       │
                                       │ 注入 repo + logger + eventBus
                                       ▼
                                     domain/cmdb/service.go
                                       │
                                       ├─ NewHostname(hostname)  ← 值对象验证
                                       ├─ s.repo.Save(ctx, asset)  ← Repository
                                       └─ s.publish(NewAssetCreated(...)) ← 事件
                                          │
                          ┌───────────────┴───────────────┐
                          ▼                               ▼
                    infra/persistence             domain/events/
                    cmdb_repo.go                  EventBus
                    → MySQL                       → Handler
```

## 五、关键设计决策

| 决策 | 选择 | 原因 |
|------|------|------|
| 单体 vs 微服务 | 单体 | 团队规模、运维成本、事务简化 |
| Repository 接口在哪 | domain 包 | 依赖倒置——domain 不依赖 infra |
| 值对象验证在哪 | Domain Service | 业务规则属于领域层 |
| 跨域事务在哪 | Services 层 | 只有 Services 可以编排多个域 |
| 事件总线实现 | 内存 EventBus | 够用，不需要 Kafka 复杂度 |
| 租户隔离 | ScopedDB WHERE + 显式 tenantID | 部分表无 tenant_id，混合策略 |
