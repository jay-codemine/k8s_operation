# DDD 生产级项目骨架设计文档

> 本文档沉淀当前项目经过多轮 DDD 改造后验证过的分层架构，作为**新项目开发的标准骨架模板**。
> 目标是：拿到本骨架后，能直接按"领域 → 仓储 → 应用服务 → 控制器/路由 → 中间件 → 事件"的顺序落地一个新业务领域，无需重新设计分层与横切关注点。

---

## 1. 架构总览

### 1.1 分层结构

```
┌─────────────────────────────────────────────────────────────┐
│  接口层 Interface        routers + controllers（HTTP handler） │
│                         对外提供 REST API + Swagger 文档        │
├─────────────────────────────────────────────────────────────┤
│  应用层 Application      services / requests / models /        │
│                         builder / worker                      │
│                         编排领域服务，实现用例，事务边界        │
├─────────────────────────────────────────────────────────────┤
│  领域层 Domain           models / values / repository(接口)    │
│                         service(领域服务) / events(领域事件)    │
│                         纯业务规则，不依赖框架与基础设施         │
├─────────────────────────────────────────────────────────────┤
│  基础设施层 Infra        persistence(仓储实现) / adapter(外部)  │
│                         实现领域定义的仓储接口，对接 DB/外部服务 │
├─────────────────────────────────────────────────────────────┤
│  中间件 Middlewares      鉴权/租户隔离/审计/限流/日志/恢复...   │
├─────────────────────────────────────────────────────────────┤
│  公共组件 pkg            可复用的技术组件（与业务无关）          │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 依赖规则

- **只能向内依赖**：Interface → Application → Domain；Infra 实现 Domain 定义的接口。
- **领域层零依赖框架**：`internal/domain` 只依赖标准库 + `pkg` 中的纯工具（如 `pkg/db.Base`、`pkg/logger` 接口），不 import `gin` / `gorm` 的 HTTP 层概念。
- **依赖倒置**：领域层定义 `Repository` 接口，基础设施层提供实现，组装在应用层完成。

---

## 2. 完整目录骨架（可直接复制）

```
<project>/
├── cmd/                              # 程序入口
│   └── <app>/
│       └── main.go                   # 唯一入口：调用 bootstrap.InitAll() + 启动 server
│
├── configs/                          # 配置文件
│   ├── config.yaml                   # 主配置（本地/默认）
│   ├── config.yaml.example           # 配置模板（提交到 git）
│   └── <domain>/                     # 领域相关配置（如 jenkins-templates 等）
│
├── global/                           # 全局单例（配置/DB/Logger/Redis 等）
│   ├── setting.go                    # ServerSetting/AppSetting/DatabaseSetting/...
│   ├── db.go                         # global.DB
│   ├── redis.go                      # global.RedisCli
│   ├── logger.go                     # global.Logger / BizLogger / AILogger
│   └── ...
│
├── initialize/                       # 初始化器（每个组件一个 Setup 函数）
│   ├── setting.go                    # 读配置 → global.*
│   ├── logger.go                     # 初始化日志
│   ├── db.go                         # 初始化 DB + 注册租户回调
│   ├── redis_client.go               # 初始化 Redis
│   ├── session.go                    # 初始化 Session
│   ├── router.go                     # 注册路由
│   ├── server.go                     # 组装 HTTP server
│   └── ...
│
├── internal/                         # 业务核心（不对外暴露）
│   ├── bootstrap/                    # 启动引导：组装顺序、Worker 启动、事件注册
│   │   ├── bootstrap.go              # InitAll() / StopCicdWorker() / FlushLoggers()
│   │   └── event_handlers.go         # 全局领域事件处理器注册
│   │
│   ├── domain/                       # ★ 领域层
│   │   ├── aggregate.go              # AggregateRoot / DomainEvent 标记接口
│   │   ├── events/                   # 事件总线（EventBus）
│   │   │   └── events.go
│   │   ├── user/                     # 示例：用户领域
│   │   │   ├── models.go             # 实体/聚合（含 TableName/AggregateID）
│   │   │   ├── values.go             # 值对象
│   │   │   ├── repository.go         # 仓储接口
│   │   │   ├── service.go            # 领域服务（业务规则）
│   │   │   └── events.go             # 领域事件定义
│   │   ├── tenant/                   # 租户领域
│   │   ├── rbac/                     # 权限领域
│   │   └── ...
│   │
│   ├── app/                          # ★ 应用层
│   │   ├── services/                 # 应用服务（用例编排）
│   │   │   ├── services.go           # Services 聚合根（组装所有子服务）
│   │   │   ├── user.go               # UserCreate/UserDelete/...
│   │   │   ├── rbac.go
│   │   │   └── ...
│   │   ├── controllers/              # 控制器（HTTP handler）
│   │   │   └── api/v1/<domain>/      # 每个领域一个包
│   │   ├── requests/                 # 请求 DTO + 校验规则
│   │   ├── routers/                  # 路由注册（每个领域一个 router）
│   │   ├── models/                   # 应用层读模型/视图模型/持久化模型
│   │   ├── builder/                  # 复杂对象组装器
│   │   ├── infra/                    # 应用层基础设施（如 Redis Stream）
│   │   └── worker/                   # 后台 Worker
│   │
│   ├── infra/                        # ★ 基础设施层
│   │   ├── persistence/              # 仓储实现（GORM）
│   │   │   ├── user_repo.go          # 实现 domain/user.UserRepository
│   │   │   └── ...
│   │   └── adapter/                  # 外部服务适配器
│   │       ├── cluster_lister.go
│   │       └── user_lookup.go
│   │
│   ├── server/                       # HTTP 服务器生命周期
│   │   ├── http.go
│   │   └── shutdown.go               # 优雅退出
│   ├── errorcode/                    # 业务错误码
│   └── health/                       # 健康检查
│
├── middlewares/                      # ★ 中间件（HTTP 横切）
│   ├── auth.go                       # JWT 鉴权
│   ├── tenant.go                     # 多租户隔离（注入 scoped DB）
│   ├── audit.go                      # 审计日志
│   ├── logger.go                     # 请求日志
│   ├── recovery.go                   # panic 恢复
│   ├── ratelimit.go                  # 限流
│   ├── prometheus.go                 # 指标采集
│   └── ...
│
├── pkg/                              # ★ 公共组件（与业务无关，可独立复用）
│   ├── db/base.go                    # Base 基础字段（ID/租户/时间/软删除）
│   ├── tenant/scope.go               # 多租户 GORM 隔离
│   ├── setting/                      # 配置读取（viper 封装）
│   ├── logger/                       # 日志
│   ├── cache/                        # 缓存
│   ├── jwt/                          # JWT 工具
│   ├── utils/                        # 通用工具（含密码哈希）
│   ├── valid/                        # 参数校验
│   ├── app/response/                 # 统一响应 envelope
│   ├── app/pagination.go             # 分页
│   └── ...
│
├── deploy/                           # 部署清单（k8s yaml / dockerfile）
├── scripts/                          # 运维脚本
└── docs/                             # 文档（含本文档）
```

---

## 3. 分层详解

### 3.1 配置层（configs/ + global/ + initialize/setting.go）

**职责**：启动时把 YAML 配置分段反序列化到 `global` 包中的全局只读结构体。

```go
// initialize/setting.go（简化）
func SetupSetting() error {
    s, _ := setting.NewSetting()                    // viper 封装
    s.ReadSection("Server", &global.ServerSetting)
    s.ReadSection("App",    &global.AppSetting)
    s.ReadSection("Database", &global.DatabaseSetting)
    s.ReadSection("Cache",  &global.CacheSetting)
    // ...
    return nil
}
```

**约定**：
- 配置按顶层 key 分段（`Server`/`Database`/`Cache`/`App`/`Security`/`Jenkins`/...）。
- 每个段对应 `global` 包中的一个结构体，运行期**只读**。
- 敏感信息（密码/token）提交 `config.yaml.example`，真实 `config.yaml` 不入库。
- 运行期需要动态修改的配置（如监控数据源）应落地到 DB，`config.yaml` 仅作首次启动引导值（见 `bootstrap.BootstrapMonitorDatasource`）。

### 3.2 领域层（internal/domain/）

这是 DDD 的核心，**每个领域一个子包**，包内固定五类文件：

| 文件 | 职责 | 示例（user 域） |
|---|---|---|
| `models.go` | 实体/聚合，GORM 模型 | `User` 内嵌 `*db.Base` |
| `values.go` | 值对象 | 枚举、不可变对象 |
| `repository.go` | 仓储接口（只有接口） | `UserRepository` |
| `service.go` | 领域服务（业务规则） | `UserService.Create/Register` |
| `events.go` | 领域事件定义 | `UserRegistered` |

**聚合根标记**：

```go
// internal/domain/aggregate.go
type AggregateRoot interface { AggregateID() int64 }
type DomainEvent   interface { EventName() string }
```

**实体示例**：

```go
// internal/domain/user/models.go
type User struct {
    Username string `json:"username" gorm:"column:username"`
    Password string `json:"-"       gorm:"column:password"`
    Role     string `json:"role"     gorm:"column:role;default:user"`
    Email    string `json:"email"    gorm:"column:email"`
    Status   int8   `json:"status"   gorm:"column:status;default:1"`
    *db.Base
}
func (u *User) TableName() string { return "user" }
func (u *User) AggregateID() int64 { return int64(u.ID) }
```

**仓储接口（依赖倒置）**：

```go
// internal/domain/user/repository.go
type UserRepository interface {
    Save(ctx context.Context, user *User) error
    Update(ctx context.Context, id uint32, values interface{}) error
    Delete(ctx context.Context, id uint32) error
    FindByID(ctx context.Context, id int64) (*User, error)
    FindByName(ctx context.Context, username string) (*User, error)
    Query(ctx context.Context, username, role, status string, page, limit int) ([]*User, int64, error)
}
```

**领域服务**：封装业务规则（密码哈希、默认租户、发布事件），通过**构造函数注入**依赖：

```go
// internal/domain/user/service.go（简化）
type UserService struct {
    repo      UserRepository
    logger    *logger.Logger
    publisher events.EventPublisher
}
func NewUserService(repo UserRepository, logger *logger.Logger, publisher events.EventPublisher) *UserService {
    return &UserService{repo: repo, logger: logger, publisher: publisher}
}

func (s *UserService) Register(username, password string) error {
    // ... 业务规则 ...
    user, err := s.Create(username, password, 0)
    if err != nil { return err }
    s.publish(NewUserRegistered(user.ID, username))  // 发布领域事件
    return nil
}
```

> 领域服务**不直接写 DB SQL**，通过 `repo` 接口操作；**不依赖 HTTP 框架**。

### 3.3 应用层（internal/app/）

**应用服务**：编排领域服务、实现完整用例、控制事务边界、处理跨域协作。核心是 `Services` 聚合根，把 DB、Redis、事件总线、日志、以及各子服务聚合在一起。

```go
// internal/app/services/services.go（简化）
type Services struct {
    db       *gorm.DB
    stream   *infra.RedisStream
    logger   *logger.Logger
    eventBus *events.EventBus
    tenantID uint32
}

func NewServices() *Services             { /* 全局：global.DB */ }
func NewServicesWithDB(db *gorm.DB) *Services { /* 租户隔离 DB（HTTP 请求） */ }
func NewBackgroundServices() *Services   { /* 后台任务：跨租户 global.DB */ }
```

**应用服务方法**（用例）：

```go
// internal/app/services/user.go（示意）
func (s *Services) UserCreate(param *requests.UserCreateRequest) (*domainUser.User, error) {
    svc := domainUser.NewUserService(
        persistence.NewUserRepository(s.db),
        s.logger, s.eventBus,
    )
    return svc.Create(param.Name, param.Password, s.tenantID)
}
```

**三个子层**：

- `controllers/`：HTTP handler，只做「绑定参数 → 校验 → 调服务 → 返回响应」。
- `requests/`：请求 DTO + 校验规则（`valid.Validate`）。
- `routers/`：每个领域一个 `NewXxxRouter().Inject(routerGroup)`。
- `models/`：应用层读模型/持久化模型（与领域模型区分，用于跨域查询结果、视图模型）。
- `worker/`：后台异步任务（CICD 部署、轮询、告警评估、审计清理等）。
- `builder/`：复杂对象组装（避免构造函数参数爆炸）。

**控制器示例**：

```go
// internal/app/controllers/api/v1/user/user.go
func (c *UserController) Create(ctx *gin.Context) {
    param := requests.NewUserCreateRequest()
    resp := response.NewResponse(ctx)

    if ok := valid.Validate(ctx, param, requests.ValidUserCreateRequest); !ok {
        return
    }

    svc := middlewares.NewServicesFromContext(ctx)   // 拿到租户隔离的 Services
    user, err := svc.UserCreate(param)
    if err != nil {
        resp.ToErrorResponse(errorcode.ErrorUserCreateFail)
        return
    }
    resp.Success(gin.H{"id": user.ID, "username": user.Username})
}
```

### 3.4 基础设施层（internal/infra/）

- `persistence/`：实现领域层定义的仓储接口，负责 GORM 细节（分页、软删除、租户过滤）。
- `adapter/`：对接外部服务（第三方 API 的防腐层）。

```go
// internal/infra/persistence/user_repo.go
type userRepo struct { db *gorm.DB }
func NewUserRepository(db *gorm.DB) user.UserRepository { return &userRepo{db: db} }
func (r *userRepo) Save(ctx context.Context, u *user.User) error {
    return r.db.WithContext(ctx).Create(u).Error
}
```

> 关键：仓储实现通过 `NewUserRepository(db)` 暴露，返回**接口**类型 `user.UserRepository`，调用方（应用服务）只依赖接口。

### 3.5 接口层（对外提供服务）

对外以 **HTTP REST API** 为主，通过 `routers` + `controllers` 暴露，并用 Swagger 注解自动生成文档。

```go
// internal/app/routers/user/user.go
func (r *UserRouter) Inject(router *gin.RouterGroup) {
    uc := v1.NewUserController()
    g := router.Group("/user")
    g.POST("/create", uc.Create)
    g.GET("/list",   uc.List)
}
```

对外服务的完整形态（详见第 7 节）：
1. **同步 REST API**：`/api/v1/...`，统一 `{code, msg, data}` envelope。
2. **后台 Worker**：异步长任务（构建/部署/巡检），不阻塞 HTTP 请求。
3. **领域事件**：进程内 EventBus，解耦领域间副作用。
4. **回调接收**：第三方回调（如 Jenkins/GitOps webhook）作为公开路由。

### 3.6 中间件层（middlewares/）

HTTP 横切关注点，统一在 `initialize/server.go` 注册，形成请求处理链：

```
Recovery → Logger → Prometheus → CORS → Auth(JWT) → TenantScope → Audit → 具体路由
```

每个中间件一个文件，职责单一：

| 中间件 | 职责 |
|---|---|
| `recovery.go` | panic 恢复，返回 500 |
| `logger.go` | 请求日志 |
| `auth.go` | JWT 解析，注入 `user_id`/`tenant_id`/`is_super_admin` |
| `tenant.go` | 用 `tenant_id` 构造 scoped DB 注入 context |
| `audit.go` | 记录写操作的审计日志 |
| `ratelimit.go` | 限流 |
| `prometheus.go` | HTTP 指标采集 |
| `license.go` | License 校验 |

**多租户中间件是关键**（见 4.1）：

```go
// middlewares/tenant.go（核心）
func TenantScope() gin.HandlerFunc {
    return func(c *gin.Context) {
        tid := c.GetUint("tenant_id")
        scopedDB := tenant.NewScopedDB(global.DB, tid)  // 所有查询自动 WHERE tenant_id=tid
        c.Set("db", scopedDB)
        c.Next()
    }
}
func NewServicesFromContext(c *gin.Context) *services.Services {
    return services.NewServicesWithDB(GetTenantDB(c))
}
```

### 3.7 公共组件层（pkg/）

与业务无关、可独立复用的技术组件：

| 组件 | 说明 |
|---|---|
| `pkg/db/base.go` | `Base` 内嵌结构（ID/TenantID/时间戳/软删除） |
| `pkg/tenant/scope.go` | 多租户 GORM 隔离 |
| `pkg/setting/` | viper 配置读取封装 |
| `pkg/logger/` | 日志（系统/业务/AI 三类） |
| `pkg/cache/` | 缓存封装 |
| `pkg/jwt/` | JWT 生成/解析 |
| `pkg/utils/` | 通用工具（含 `HashPassword`） |
| `pkg/valid/` | 参数校验 |
| `pkg/app/response/` | 统一响应 envelope |
| `pkg/app/pagination.go` | 分页 |

---

## 4. 横切关注点

### 4.1 数据库与多租户

**Base 模型**：所有需要持久化的实体内嵌 `*db.Base`，自动获得 ID、租户 ID、时间戳、软删除。

```go
// pkg/db/base.go
type Base struct {
    ID         uint32 `gorm:"primary_key"`
    TenantID   uint32 `gorm:"column:tenant_id;default:1;index"`
    CreatedAt  uint32
    ModifiedAt uint32
    DeletedAt  uint32
    IsDel      uint8  `json:"is_del"`  // 软删除标记
}
```

**多租户隔离机制**（三件套）：

1. **Scoped DB**（`pkg/tenant.NewScopedDB`）：基于 GORM `Session` + 持久化 `WHERE tenant_id = X`，后续所有查询自动隔离。
2. **INSERT 自动填充**（`pkg/tenant.RegisterCallbacks` 的 `fillTenantID` 回调）：把当前租户 ID 写入待插入记录的 `tenant_id` 字段（仅当字段为零值时生效，超管跨租户写入不被覆盖）。
3. **中间件注入**（`middlewares/tenant.go`）：从 JWT 拿 `tenant_id` → 构造 scoped DB → 注入 context。

**三种 Services 实例**：

| 构造函数 | 用途 | DB |
|---|---|---|
| `NewServices()` | 全局单例 | `global.DB` |
| `NewServicesWithDB(db)` | HTTP 请求（租户隔离） | scoped DB |
| `NewBackgroundServices()` | 后台 Worker/启动引导（跨租户） | `global.DB` |

### 4.2 领域事件

**进程内 EventBus**（同步发布订阅，适合单体内部解耦）：

```go
// internal/domain/events/events.go
type EventBus struct {
    handlers map[string][]EventHandler
}
func (b *EventBus) Subscribe(eventName string, handler EventHandler)
func (b *EventBus) Publish(event DomainEvent)  // 带 panic 恢复，单个 handler 异常不影响其他
```

**使用流程**：

1. 领域层定义事件（内嵌 `events.BaseEvent`）：
```go
// internal/domain/user/events.go
type UserRegistered struct {
    events.BaseEvent
    UserID   uint32
    Username string
}
func NewUserRegistered(id uint32, username string) UserRegistered {
    return UserRegistered{BaseEvent: events.NewBaseEvent("user.registered"), UserID: id, Username: username}
}
```

2. 领域服务发布事件：`s.publish(NewUserRegistered(user.ID, username))`。

3. 启动时注册全局处理器：
```go
// internal/bootstrap/event_handlers.go
bus.Subscribe("user.registered", func(e events.DomainEvent) {
    global.Logger.Infof("[DomainEvent] 用户注册: %s", e.EventName())
})
```

> 扩展建议：需要跨进程/异步时，可将 `EventBus.Publish` 替换为消息队列（Redis Stream / Kafka）实现，接口不变。

### 4.3 错误码

- 定义在 `internal/errorcode`，启动时 `errorcode.Register()` 一次性注册。
- 统一错误结构：`Code / Msg / Details / StatusCode`。
- 控制器通过 `resp.ToErrorResponse(errorcode.ErrorXxxFail)` 返回。

### 4.4 日志

- `global.Logger`（系统）、`global.BizLogger`（业务）、`global.AILogger`（AI 调用）三类。
- 通过 `initialize.SetupLogger()` 初始化，`bootstrap.FlushLoggers()` 在退出时落盘。

### 4.5 请求校验与响应

- 请求 DTO 定义在 `internal/app/requests/`，每个 DTO 配套校验规则函数。
- 控制器入口 `valid.Validate(ctx, param, requests.ValidXxxRequest)`。
- 统一响应 envelope：`{code:0, msg:"OK", data:{...}}`；列表用 `SuccessList`。

---

## 5. 启动引导流程

```
cmd/main.go
  └─ bootstrap.InitAll()
       ├─ initialize.SetupSetting()       # 1. 读配置
       ├─ errorcode.Register()            # 2. 注册错误码
       ├─ initialize.SetupValidator()     # 3. 校验器
       ├─ initialize.SetupLogger()        # 4. 日志
       ├─ initialize.SetupDB()            # 5. DB + 租户回调
       ├─ initialize.SetupSession()       # 6. Session（Redis）
       ├─ initialize.SetupRedis()         # 7. Redis
       ├─ initialize.SetupK8sBootstrap()  # 8. K8s（失败不阻塞）
       ├─ initialize.LogDocsReady()       # 9. Swagger
       ├─ bootstrap.StartCicdWorker()     # 10. 启动 Worker
       ├─ ... 启动审计清理/告警/AIOps Worker
       └─ bootstrap.registerEventHandlers() # 11. 注册领域事件
  └─ server 启动 + 优雅退出（shutdown）
```

**启动原则**：
- **核心依赖失败必须终止**：DB、Redis、配置、日志。
- **非核心依赖失败降级**：K8s、Worker、可选第三方（Jenkins/ArgoCD）失败只告警不阻塞。

---

## 6. 如何基于骨架开发一个新领域（实战步骤）

以新增「订单 Order」领域为例：

### Step 1：定义领域模型与仓储接口

```
internal/domain/order/
├── models.go       # Order 实体（内嵌 *db.Base）
├── repository.go   # OrderRepository 接口
├── service.go      # OrderService 领域服务
└── events.go       # OrderCreated 事件
```

### Step 2：实现仓储

```
internal/infra/persistence/order_repo.go
└── type orderRepo struct { db *gorm.DB }
    func NewOrderRepository(db *gorm.DB) order.OrderRepository
```

### Step 3：编写应用服务用例

```
internal/app/services/order.go
└── func (s *Services) OrderCreate(param *requests.OrderCreateRequest) (*order.Order, error)
     内部组装：NewOrderService(NewOrderRepository(s.db), s.logger, s.eventBus)
```

### Step 4：定义请求 DTO + 校验

```
internal/app/requests/order.go
└── OrderCreateRequest + ValidOrderCreateRequest
```

### Step 5：编写控制器

```
internal/app/controllers/api/v1/order/order.go
└── OrderController.Create(ctx)：绑定 → 校验 → svc.OrderCreate → 响应
```

### Step 6：注册路由

```
internal/app/routers/order/order.go
└── NewOrderRouter().Inject(router.Group("/order"))
```

### Step 7：接入启动流程

- 在 `initialize/router.go` 中 `NewOrderRouter().Inject(apiGroup)`。
- 如需领域事件副作用，在 `bootstrap/event_handlers.go` 中 `Subscribe("order.created", ...)`。
- 如需后台任务，在 `internal/app/worker/` 新增 Worker，`bootstrap.InitAll()` 中启动。

---

## 7. 对外提供服务的方式

| 方式 | 场景 | 实现位置 |
|---|---|---|
| **同步 REST API** | 前端/第三方调用，查询与命令 | `routers` + `controllers` |
| **后台 Worker** | 异步长任务（构建、部署、巡检、轮询） | `app/worker` + Redis 队列 |
| **领域事件** | 领域间副作用解耦 | `domain/events` + EventBus |
| **Webhook 回调接收** | Jenkins/GitOps 等第三方回调 | 公开路由（不经过鉴权） |
| **对外调用适配** | 调用 Jenkins/ArgoCD/Prometheus/OpenAI/LDAP | `pkg` 客户端 + `infra/adapter` |

---

## 8. 命名与规范约定

1. **包名单数、小写**：`user`、`order`、`rbac`。
2. **领域文件五件套**：`models.go` / `values.go` / `repository.go` / `service.go` / `events.go`。
3. **仓储**：接口名 `XxxRepository`，实现结构体小写 `xxxRepo`，构造函数 `NewXxxRepository(db) XxxRepository`（返回接口）。
4. **服务**：领域服务 `XxxService` + `NewXxxService(...)`；应用服务挂 `Services` 结构体，方法名 `XxxDoSomething`。
5. **控制器**：`XxxController` + `NewXxxController()`，方法对应一个 endpoint。
6. **请求**：`XxxCreateRequest` / `XxxUpdateRequest` / `XxxQueryRequest`，校验函数 `ValidXxxRequest`。
7. **表名**：领域实体 `TableName()` 显式指定；蛇形命名（如 `cicd_pipeline`）。
8. **软删除**：统一用 `IsDel` 标记，查询默认加 `is_del = 0`。
9. **错误码**：集中在 `internal/errorcode`，命名 `ErrorXxxFail`。
10. **跨层通信**：领域层 ↔ 应用层用领域对象；控制器 ↔ 应用层用 `requests` DTO；对外输出用 `response` envelope。

---

## 9. 骨架的适用边界

**适合**：单体优先、需要清晰分层与多租户隔离的中大型后端服务（平台类/管理系统/中台）。

**当前骨架的取舍**（可作为后续演进方向）：
- 事件为进程内同步，跨进程/异步需引入消息队列（接口已预留）。
- 多租户用「共享表 + tenant_id 列」隔离，超大规模租户数可考虑 schema 隔离。
- 对外仅 REST，如需 gRPC 可在 `internal/server` 并行挂载 gRPC server，复用应用服务层。
- 领域层与应用层模型目前存在一定重合（`domain/*/models.go` 与 `app/models/`），新领域应优先在 `domain` 定义实体，`app/models` 仅放跨域读模型与视图模型。
