# 请求完整链路文档

## 一、全景图

```
浏览器 http://localhost:18082
      │
      ▼
┌─────────────────────────────────────────────────────┐
│  Nginx (k8soperation-web pod :80)                   │
│  /api/* → proxy_pass → backend:8080                 │
│  /assets/* → 静态文件                                │
│  /* → try_files $uri /index.html (SPA fallback)     │
└─────────────────────────────────────────────────────┘
      │ /api/v1/k8s/cicd/pipeline/update
      ▼
┌─────────────────────────────────────────────────────┐
│ ① Router (路由注册)                                   │
│    internal/app/routers/kube_cicd/cicd_router.go     │
│    pipeline.POST("/update",                          │
│      RequireCICDPermission("cicd:pipeline:edit"),    │
│      r.pipelineCtrl.Update)                          │
├─────────────────────────────────────────────────────┤
│ ② Middleware 链                                      │
│    AuthJWT         → 解析 token, 注入 user/tenant    │
│    TenantScope     → 租户隔离 DB 注入 context         │
│    RequireCICDPermission → 细粒度权限检查             │
│    Logger/Audit    → 请求日志/审计                     │
├─────────────────────────────────────────────────────┤
│ ③ Controller (HTTP 适配层)                           │
│    internal/app/controllers/api/v1/cicd/             │
│    pipeline_controller.go                            │
│    → 绑定参数 → 校验 → 调 Services → 统一响应         │
├─────────────────────────────────────────────────────┤
│ ④ Services (应用编排层)                               │
│    internal/app/services/cicd_pipeline.go            │
│    → 跨域编排、事务、DTO 转换、权限校验入口            │
├─────────────────────────────────────────────────────┤
│ ⑤ Domain Service (领域服务层)                        │
│    internal/domain/cicd/service.go                   │
│    → 业务规则、值对象、状态机                          │
├─────────────────────────────────────────────────────┤
│ ⑥ Repository 接口 (纯契约)                            │
│    internal/domain/cicd/repository.go                │
│    → 纯 interface, 不依赖 GORM                       │
├─────────────────────────────────────────────────────┤
│ ⑦ infra/persistence (GORM 实现)                      │
│    internal/infra/persistence/cicd_repo.go           │
│    → SQL 查询, WHERE tenant_id=?                     │
└─────────────────────────────────────────────────────┘
      │
      ▼
┌─────────────────────────────────────────────────────┐
│  MySQL (1.117.227.207:30306)                        │
│  Database: k8s-platform                              │
└─────────────────────────────────────────────────────┘
```

---

## 二、逐层详解

### 2.1 Router — 路由注册

每一组路由一个文件，注入中间件和 Controller：

```go
// internal/app/routers/kube_cicd/cicd_router.go

func (r *CicdRouter) Inject(rg *gin.RouterGroup) {
    pipeline := rg.Group("/pipeline")

    // GET - 读操作，无需额外权限
    pipeline.GET("/list", r.pipelineCtrl.List)          // 列表
    pipeline.GET("/detail", r.pipelineCtrl.Detail)      // 详情
    pipeline.GET("/logs", r.pipelineCtrl.Logs)          // 日志
    pipeline.GET("/status", r.pipelineCtrl.Status)      // 状态
    pipeline.GET("/history", r.pipelineCtrl.History)    // 历史

    // POST - 写操作，需要细粒度权限
    pipeline.POST("/create", RequireCICDPermission("cicd:pipeline:create"), r.pipelineCtrl.Create)
    pipeline.POST("/update", RequireCICDPermission("cicd:pipeline:edit"),   r.pipelineCtrl.Update)
    pipeline.POST("/delete", RequireCICDPermission("cicd:pipeline:delete"), r.pipelineCtrl.Delete)
    pipeline.POST("/run",    RequireCICDPermission("cicd:pipeline:run"),    r.pipelineCtrl.Run)
    pipeline.POST("/stop",   RequireCICDPermission("cicd:build:cancel"),    r.pipelineCtrl.Stop)
}

// 主入口注册所有路由组:
// internal/app/routers/router.go
func Register(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    // ... 按域注册
    kubeCicd := kube_cicd.NewCicdRouterWithFactory(factory)
    kubeCicd.Inject(v1.Group("/k8s/cicd"))
}
```

### 2.2 Middleware — 中间件链

```
请求 → [Recovery] → [CORS] → [Logger] → [AuthJWT] → [TenantScope] → [RequireCICDPermission] → Controller
```

每个中间件的职责：

| 中间件 | 文件 | 作用 |
|--------|------|------|
| Recovery | gin 内置 | panic 恢复，防止进程崩溃 |
| CORS | middlewares/cors.go | 跨域许可 |
| Logger | middlewares/logger.go | 请求耗时、状态码日志 |
| AuthJWT | middlewares/auth.go | 解析 JWT → 注入 `user_id`/`tenant_id`/`is_super_admin` |
| TenantScope | middlewares/tenant.go | 构造租户隔离 DB → `ctx.Set("db", scopedDB)` |
| RequireCICDPermission | middlewares/auth.go | 检查 `HasUserPermission(userID, "cicd:pipeline:edit")` |

**AuthJWT 流程**:
```go
// middlewares/auth.go
func AuthJWT() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        token := extractToken(ctx)        // 从 Header 取 Bearer token
        claims := parseJWT(token)         // 解析 user_id, tenant_id
        ctx.Set("user_id", claims.UserID)
        ctx.Set("tenant_id", claims.TenantID)

        // 检查是否为超级管理员（写入上下文）
        isSuperAdmin := NewServicesFromContext(ctx).IsSuperAdmin(claims.UserID)
        ctx.Set("is_super_admin", isSuperAdmin)

        ctx.Next()
    }
}
```

**RequireCICDPermission 流程**:
```go
// middlewares/auth.go
func RequireCICDPermission(permissionName string) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        userID := ctx.GetInt64("user_id")

        // 走完整权限检查链 → Controller → Services → Domain → Repo → MySQL
        if !NewServicesFromContext(ctx).HasUserPermission(userID, permissionName) {
            rsp.ToErrorResponse(ErrorRBACAccessDenied)  // 403
            ctx.Abort()
            return
        }
        ctx.Next()
    }
}
```

### 2.3 Controller — HTTP 适配层

只做三件事：**绑参 → 调 Services → 返回统一格式**

```go
// internal/app/controllers/api/v1/cicd/pipeline_controller.go

func (c *PipelineController) Update(ctx *gin.Context) {
    // 1. 绑定参数
    param := requests.NewPipelineUpdateRequest()
    if ok := valid.Validate(ctx, param, requests.ValidPipelineUpdateRequest); !ok {
        return  // 参数校验失败，validator 已自动返回错误
    }

    // 2. 从 context 获取 Services（含租户隔离 DB）
    svc := middlewares.NewServicesFromContext(ctx)

    // 3. 委托给 Services 层
    err := svc.PipelineUpdate(ctx.Request.Context(), param)
    if err != nil {
        rsp.ToErrorResponse(ErrorPipelineUpdateFail.WithDetails(err.Error()))
        return
    }

    // 4. 统一成功响应
    rsp.Success(gin.H{"message": "更新成功"})
}
```

### 2.4 Services — 应用编排层

跨域编排、事务管理、DTO 转换。**只有这一层可以协调多个领域服务**。

```go
// internal/app/services/cicd_pipeline.go

func (s *Services) PipelineUpdate(ctx context.Context, req *PipelineUpdateRequest) error {
    // 1. 调用领域服务 — 检查是否存在
    pipeline, err := s.cicdSvc().PipelineGetByID(ctx, req.ID)
    if err != nil {
        return errors.New("流水线不存在")
    }

    // 2. 调用领域服务 — 检查重名
    if req.Name != "" && req.Name != pipeline.Name {
        _, err := s.cicdSvc().PipelineGetByName(ctx, req.Name)
        if err == nil {
            return errors.New("流水线名称已存在")
        }
    }

    // 3. 构建更新字段（DTO → map）
    updates := make(map[string]interface{})
    if req.Name != ""        { updates["name"] = req.Name }
    if req.Description != "" { updates["description"] = req.Description }
    if req.GitRepo != ""     { updates["git_repo"] = req.GitRepo }

    // 4. 委托给领域服务执行更新
    return s.cicdSvc().PipelineUpdate(ctx, req.ID, updates)
}

// cicdSvc 工厂钩子 —— 每次调用创建新的领域服务实例
func (s *Services) cicdSvc() *dmcicd.CicdService {
    svc := dmcicd.NewCicdService(
        global.DB,                          // *gorm.DB
        persistence.NewCicdRepository(global.DB),  // Repository 实现
        s.eventBus,                         // 领域事件总线
    )
    svc.SetTenantID(s.tenantID)             // 注入当前请求的租户 ID
    return svc
}
```

### 2.5 Domain Service — 领域服务层

业务规则、值对象验证、聚合根约束。**不 import GORM / gin / HTTP**。

```go
// internal/domain/cicd/service.go

type CicdService struct {
    db       *gorm.DB           // 只用于直接查询（非 CRUD）
    repo     CicdRepository     // 仓储接口（核心）
    eventBus *events.EventBus
    tenantID uint32
}

func (s *CicdService) PipelineUpdate(ctx context.Context, id int64, updates map[string]interface{}) error {
    // 1. 领域规则：运行时流水线不允许修改名称
    pipeline, err := s.repo.PipelineGetByID(ctx, id, s.tenantID)
    if err != nil {
        return err
    }
    if pipeline.Status == PipelineStatusRunning {
        if _, ok := updates["name"]; ok {
            return errors.New("运行中的流水线不能修改名称")
        }
    }

    // 2. 值对象校验
    if name, ok := updates["name"]; ok {
        pipelineName := NewPipelineName(name.(string))
        if err := pipelineName.Validate(); err != nil {
            return err
        }
    }

    // 3. 委托仓储执行持久化
    return s.repo.PipelineUpdate(ctx, id, s.tenantID, updates)
}
```

### 2.6 Repository 接口 — 纯领域契约

只定义接口，**不 import GORM**。

```go
// internal/domain/cicd/repository.go

type CicdRepository interface {
    // Pipeline CRUD
    PipelineCreate(ctx context.Context, pipeline *Pipeline) error
    PipelineGetByID(ctx context.Context, id int64, tenantID uint32) (*Pipeline, error)
    PipelineGetByName(ctx context.Context, name string, tenantID uint32) (*Pipeline, error)
    PipelineList(ctx context.Context, filter PipelineListFilter) ([]*Pipeline, int64, error)
    PipelineUpdate(ctx context.Context, id int64, tenantID uint32, updates map[string]interface{}) error
    PipelineDelete(ctx context.Context, id int64, tenantID uint32) error

    // Pipeline Run
    PipelineRunCreate(ctx context.Context, run *PipelineRun) error
    PipelineRunGetLatest(ctx context.Context, pipelineID int64) (*PipelineRun, error)
    PipelineRunUpdateStatus(ctx context.Context, id int64, status int) error
    // ... 更多接口
}
```

### 2.7 infra/persistence — GORM 实现

```go
// internal/infra/persistence/cicd_repo.go

type cicdRepoImpl struct {
    db       *gorm.DB
    tenantID uint32
}

func (r *cicdRepoImpl) PipelineUpdate(ctx context.Context, id int64, tenantID uint32, updates map[string]interface{}) error {
    return r.db.WithContext(ctx).
        Table("cicd_pipeline").
        Where("id = ? AND tenant_id = ?", id, tenantID).
        Updates(updates).Error
}

func (r *cicdRepoImpl) PipelineGetByID(ctx context.Context, id int64, tenantID uint32) (*Pipeline, error) {
    var pipeline Pipeline
    err := r.db.WithContext(ctx).
        Where("id = ? AND tenant_id = ?", id, tenantID).
        First(&pipeline).Error
    return &pipeline, err
}
```

---

## 三、权限检查完整链路

```
HTTP Request
    │
    ▼
AuthJWT 中间件
    ├─ 解析 JWT → user_id=3, tenant_id=1
    ├─ ctx.Set("user_id", 3)
    ├─ ctx.Set("tenant_id", 1)
    └─ isSuperAdmin = Services.IsSuperAdmin(3) → true
       ctx.Set("is_super_admin", true)
    │
    ▼
TenantScope 中间件
    ├─ scopedDB := tenant.NewScopedDB(global.DB, 1)  // WHERE tenant_id=1
    ├─ ctx.Set("db", scopedDB)
    └─ ctx.Next()
    │
    ▼
RequireCICDPermission("cicd:pipeline:edit") 中间件
    ├─ userID := ctx.GetInt64("user_id")  // 3
    └─ Services.HasUserPermission(3, "cicd:pipeline:edit")
        └─ rbacSvc().HasUserPermission(3, "cicd:pipeline:edit")
            └─ repo.HasUserPermission(ctx, 3, "cicd:pipeline:edit")
                └─ rbacRepoImpl.HasUserPermission(ctx, 3, "cicd:pipeline:edit")
                    │
                    ├─ ① r.IsSuperAdmin(ctx, 3)  // 使用 r.tenantID
                    │     → SELECT count(*) FROM sys_user_role
                    │       JOIN sys_role ON ...
                    │       WHERE user_id=3 AND role_type='super_admin'
                    │       AND tenant_id=1
                    │     → count > 0 → true → 直接返回 true ✅
                    │
                    └─ ② (如果不是超级管理员)
                          → SELECT count(*) FROM sys_role_permission
                            JOIN sys_user_role ON ...
                            JOIN sys_permission ON ...
                            WHERE user_id=3 AND name='cicd:pipeline:edit'
                            AND sys_user_role.tenant_id=1
                            AND sys_role_permission.tenant_id=1
    │
    ▼
Controller.Update(ctx)
```

---

## 四、跨域通信模式

### 4.1 直接依赖（类型引用）
```go
// rbac/service.go
import "k8soperation/internal/domain/k8s"
func (s *RbacService) GetUserClusters(userID int64) ([]*k8s.Cluster, error) { ... }
```

### 4.2 依赖倒置接口（推荐）
```go
// rbac/interfaces.go — rbac 域定义自己需要的接口
type ClusterLister interface {
    ListAllClusters(ctx context.Context) ([]ClusterInfo, error)
}

// infra/adapter/cluster_lister.go — k8s 域的实现
type ClusterListerAdapter struct { svc *k8s.ClusterService }
func (a *ClusterListerAdapter) ListAllClusters(ctx) ([]rbac.ClusterInfo, error) { ... }

// app/services/rbac.go — 组装时注入
svc.WithClusterLister(adapter.NewClusterLister(s.clusterSvc()))
```

### 4.3 领域事件（异步解耦）
```go
// k8s/service.go
s.publisher.Publish(NewClusterCreated(clusterID, clusterName))

// bootstrap/event_handlers.go
bus.Subscribe("k8s.cluster.created", func(e DomainEvent) {
    // 审计日志、通知、缓存失效
})
```

### 4.4 Services 层编排（跨域事务）
```go
// app/services/tenant.go
func (s *Services) TenantCreate(ctx, name, code string) {
    s.db.Transaction(func(tx *gorm.DB) error {
        tenant := s.tenantSvc().CreateWithTx(tx, name, code)
        rbac.SeedTenantRBAC(tx, tenant.ID)
        return nil
    })
}
```

---

## 五、10 个领域及其文件清单

| 域 | models | repository | service | values | events | aggregate | interfaces |
|----|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| k8s | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| user | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — |
| cicd | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | — |
| rbac | ✅ | ✅ | ✅ | ✅ | — | ✅ | ✅ |
| tenant | ✅ | ✅ | ✅ | — | ✅ | ✅ | — |
| audit | ✅ | ✅ | ✅ | — | — | — | — |
| settings | ✅ | ✅ | ✅ | — | — | — | — |
| monitor | ✅ | ✅ | ✅ | ✅ | — | — | ✅ |
| image | ✅ | ✅ | ✅ | — | — | — | — |
| ai | ✅ | ✅ | ✅ | ✅ | — | — | — |
| appstore | ✅ | ✅ | ✅ | ✅ | — | — | — |

### 每个域的标准注册链路

```
1. domain/{name}/models.go           ← 实体 + 常量 + 查询 DTO
2. domain/{name}/repository.go       ← 纯接口
3. domain/{name}/service.go          ← 领域服务（注入 repo + 业务逻辑）
4. domain/{name}/values.go           ← 值对象（推荐）
5. infra/persistence/{name}_repo.go  ← GORM 实现
6. app/services/{name}.go            ← Services 钩子 + 编排
7. app/controllers/api/v1/{name}/    ← Controller
8. app/routers/{name}/               ← 路由注册
```

---

## 六、新增一个领域的步骤

以新增 `notification` 域为例：

```
□ 1. internal/domain/notification/models.go        // 实体定义
□ 2. internal/domain/notification/repository.go    // 仓储接口
□ 3. internal/domain/notification/service.go        // 领域服务
□ 4. internal/domain/notification/values.go         // 值对象（如 NotificationChannel）
□ 5. internal/infra/persistence/notification_repo.go // GORM 实现
□ 6. internal/app/services/notification.go          // Services 钩子
□ 7. internal/app/controllers/api/v1/notification/   // Controller
□ 8. internal/app/routers/notification/              // 路由注册
□ 9. internal/app/routers/router.go                 // 注册新路由组
```

---

## 七、项目目录结构

```
k8s_operation/
├── cmd/
│   └── k8soperation/main.go         # 入口：初始化 DB、注册路由、启动 HTTP
├── internal/
│   ├── app/
│   │   ├── controllers/api/v1/      # Controller (HTTP 适配层)
│   │   ├── middlewares/              # 中间件：Auth/CORS/Logger/Tenant/Permission
│   │   ├── models/                   # 防腐层类型别名
│   │   ├── requests/                 # 请求参数 DTO + 校验规则
│   │   ├── routers/                  # 按域分组的 Gin 路由
│   │   └── services/                 # Services 层（编排 + 事务 + DTO）
│   ├── domain/                       # 领域层（纯 Go，不依赖框架）
│   │   ├── k8s/                      # 每个域独立包
│   │   ├── user/
│   │   ├── cicd/
│   │   ├── rbac/
│   │   ├── tenant/
│   │   ├── audit/
│   │   ├── settings/
│   │   ├── monitor/
│   │   ├── image/
│   │   ├── ai/
│   │   ├── appstore/
│   │   └── events/                   # 领域事件定义 + EventBus
│   ├── infra/
│   │   ├── adapter/                  # 跨域适配器（依赖倒置）
│   │   └── persistence/              # Repository 的 GORM 实现
│   └── pkg/                          # 内部公共工具
│       ├── k8s/                       # client-go 封装
│       ├── tenant/                    # 租户隔离（ScopedDB）
│       └── ...
├── pkg/                               # 可复用公共库
│   ├── config/                        # Viper 配置加载
│   ├── errors/                        # 统一错误码
│   ├── logger/                        # Zap 日志
│   ├── metrics/                       # Prometheus 指标
│   ├── response/                      # 统一响应格式 {code, msg, data}
│   └── tenant/                        # ScopedDB 工具
├── configs/
│   ├── config.yaml                    # 默认配置
│   └── jenkins-templates/             # Jenkins Pipeline 模板
├── scripts/migrations/                # SQL 迁移脚本
├── k8s-web/                           # Vue 3 前端
├── deploy/                            # K8s 部署 YAML
└── docs/
    └── architecture/
        ├── DDD_ARCHITECTURE.md        # DDD 架构总览
        └── REQUEST_LIFECYCLE.md       # 本文档
```
