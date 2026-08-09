# DDD 领域驱动架构设计文档

## 一、架构定位

### 1.1 当前架构：DDD-Lite 单体应用

```
本项目是 DDD 领域驱动设计的单体应用（Modular Monolith），不是微服务架构。
所有领域运行在同一进程内，通过接口调用通信，不经过网络。
```

| 维度 | 本项目（DDD 单体） | 微服务架构 |
|------|-------------------|-----------|
| 部署单元 | 1 个二进制 | N 个独立服务 |
| 通信方式 | 进程内函数调用 | HTTP/gRPC/消息队列 |
| 数据存储 | 共享 MySQL + Redis | 每服务独立数据库 |
| 事务 | GORM 本地事务 | 分布式事务（Saga/TCC） |
| 运维复杂度 | 低 | 高 |
| 领域隔离 | 包级隔离 + 接口契约 | 进程级隔离 |

### 1.2 何时演进到微服务

当出现以下信号时，可从单体中拆分独立微服务：
- 某领域团队需要独立部署节奏
- 某领域需要独立扩缩容
- 某领域需要独立技术栈或数据库
- 跨领域调用出现性能瓶颈

---

## 二、内部领域架构

### 2.1 领域分层

```
┌─────────────────────────────────────────────┐
│  Controller (HTTP 层)                       │
│  middlewares.NewServicesFromContext(ctx)     │
├─────────────────────────────────────────────┤
│  Services (应用服务层)                       │
│  编排领域服务、事务、DTO转换、权限校验       │
│  s.userSvc() / s.rbacSvc() / s.cicdSvc()   │
├─────────────────────────────────────────────┤
│  Domain Service (领域服务层)                 │
│  业务规则、值对象验证、聚合根约束            │
│  通过 Repository 接口访问持久化              │
├─────────────────────────────────────────────┤
│  Repository 接口 (domain/{name}/repository.go)│
│  纯接口，不依赖 GORM                        │
├─────────────────────────────────────────────┤
│  infra/persistence (GORM 实现)              │
│  infra/adapter (跨域适配器)                  │
└─────────────────────────────────────────────┘
```

### 2.2 领域包结构

```
internal/domain/{name}/
├── models.go           # 实体定义（GORM 标签）
├── repository.go       # 仓储接口（纯领域契约）
├── service.go          # 领域服务（业务逻辑）
├── values.go           # 值对象（不可变、自验证）
├── events.go           # 领域事件类型（可选）
├── aggregate.go        # 聚合根 + 状态机（可选）
└── interfaces.go       # 跨域接口契约（可选）
```

### 2.3 完整领域定义清单

| 文件 | 必要性 | 说明 |
|------|:------|------|
| `models.go` | 必须 | 实体、常量、查询 DTO |
| `repository.go` | 必须 | 仓储接口，定义 CRUD 契约 |
| `service.go` | 必须 | 领域服务，注入 repo + 业务逻辑 |
| `values.go` | 推荐 | 值对象，字段级验证 |
| `events.go` | 可选 | 领域事件，解耦跨域通信 |
| `aggregate.go` | 可选 | 聚合根 + 子实体状态机 |
| `interfaces.go` | 可选 | 跨域依赖倒置接口 |

### 2.4 每个领域需要注册的完整链路

```
1. domain/{name}/models.go           ← 实体定义
2. domain/{name}/repository.go       ← 仓储接口
3. domain/{name}/service.go          ← 领域服务(NEW)
4. infra/persistence/{name}_repo.go  ← GORM 实现(NEW)
5. infra/adapter/{name}_xxx.go       ← 跨域适配器(NEW/如需)
6. app/services/{name}.go            ← Services hook
7. app/services/services.go          ← 无需改（hook 在单独文件）
```

---

## 三、内部领域通信

### 3.1 通信模式总览

```
┌──────────┐   接口调用    ┌──────────┐
│  k8s 域  │──────────────→│  rbac 域  │
│          │  直接 import  │          │
└──────────┘              └──────────┘
      ↑                        │
      │ 依赖倒置接口            │ 领域事件
      │                 ┌──────↓─────┐
      └─────────────────│  EventBus   │
                        └─────────────┘
```

### 3.2 模式一：直接依赖（简单场景）

**场景**: A 域需要 B 域的类型定义

```go
// rbac/service.go
import "k8soperation/internal/domain/k8s"

func (s *RbacService) GetUserAccessibleClusters(userID int64) ([]*k8s.Cluster, error) {
    // 直接返回 k8s.Cluster 类型
}
```

**适用**: 类型引用、读操作、不涉及业务规则

**风险**: 紧耦合，B 域变更影响 A 域

### 3.3 模式二：依赖倒置接口（推荐）

**场景**: A 域需要 B 域的数据，但不想直接依赖 B 域

```go
// rbac/interfaces.go — rbac 域定义自己需要的接口
type ClusterLister interface {
    ListAllClusters(ctx context.Context) ([]ClusterInfo, error)
}

type ClusterInfo struct {  // 跨域 DTO，不是 k8s.Cluster
    ID          uint32
    ClusterName string
    Status      uint8
}
```

```go
// infra/adapter/cluster_lister.go — k8s 域的实现
type ClusterListerAdapter struct { svc *k8s.ClusterService }

func (a *ClusterListerAdapter) ListAllClusters(ctx context.Context) ([]rbac.ClusterInfo, error) {
    clusters, _, err := a.svc.List(ctx, "", 1, 1000)
    // 转换 k8s.Cluster → rbac.ClusterInfo
    return result, err
}
```

```go
// app/services/rbac.go — 组装
svc.WithClusterLister(adapter.NewClusterLister(s.clusterSvc()))
```

**适用**: 跨域数据查询、需要解耦

**优点**: rbac 不依赖 k8s 包，可替换实现

### 3.4 模式三：领域事件（异步解耦）

**场景**: A 域发生事情，B 域需要响应

```go
// k8s/service.go — 发布事件
func (s *ClusterService) Create(...) error {
    // ...业务逻辑
    s.publisher.Publish(NewClusterCreated(clusterID, clusterName))
}

// bootstrap/event_handlers.go — 注册处理器
bus.Subscribe("k8s.cluster.created", func(e events.DomainEvent) {
    // 审计日志、通知、缓存失效等
})
```

**适用**: 跨域副作用（审计、通知、统计）

**优点**: 完全解耦，发布者不知道订阅者

### 3.5 模式四：Services 层编排（跨域事务）

**场景**: 需要原子化操作多个域

```go
// app/services/tenant.go
func (s *Services) TenantCreate(...) {
    s.db.Transaction(func(tx *gorm.DB) error {
        tenant, _ := s.tenantSvc().CreateWithTx(tx, name, code)
        rbac.SeedTenantRBAC(tx, tenantID)  // RBAC 初始化
        return nil
    })
}
```

**适用**: 跨域事务、复杂编排

**注意**: 只在 Services 层使用，Domain 层不做编排

---

## 四、外部领域通信（未来微服务拆分）

### 4.1 当前状态

```
目前所有"外部"领域都在同一进程内。
拆分后，当前接口调用将变为网络调用。
```

### 4.2 拆分路径

```
当前:                                  拆分后:
┌─────────────────────┐        ┌──────────┐  HTTP   ┌──────────┐
│  Monolith            │        │  k8s-svc │←──────→│ rbac-svc │
│  ┌──────┐ ┌───────┐ │        │  :8081   │  gRPC  │  :8082   │
│  │ k8s  │ │ rbac  │ │   →   └──────────┘        └──────────┘
│  └──────┘ └───────┘ │
└─────────────────────┘
```

### 4.3 拆分准备

本项目的 Repository 接口和跨域适配器已经为微服务拆分做好了准备：

| 当前 | 拆分后 |
|------|--------|
| `domain/k8s/repository.go` | 接口不变，GORM 实现替换为 HTTP 实现 |
| `domain/rbac/interfaces.go` | 接口不变，Adapter 从本地调用改为 RPC 调用 |
| `infra/persistence/*_repo.go` | 移动到 k8s-svc，通过 HTTP/gRPC 暴露 |
| `domain/events/EventBus` | 替换为 Kafka/RabbitMQ |

---

## 五、完整新增领域检查清单

```
□ 1. domain/{name}/models.go       实体 + 常量 + DTO
□ 2. domain/{name}/repository.go   仓储接口
□ 3. domain/{name}/service.go      领域服务
□ 4. domain/{name}/values.go       值对象（推荐）
□ 5. domain/{name}/events.go       领域事件（可选）
□ 6. domain/{name}/aggregate.go    聚合根（可选）
□ 7. domain/{name}/interfaces.go   跨域接口（可选）
□ 8. infra/persistence/{name}_repo.go  GORM 实现
□ 9. infra/adapter/{name}_xxx.go   跨域适配器（可选）
□ 10. app/services/{name}.go       Services hook
□ 11. app/models/{name}.go         类型别名（防腐层）
```

---

## 六、当前项目覆盖率

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

---

## 七、关键设计原则

1. **领域层不依赖基础设施**: domain/ 不 import `global`、`gin`、`gorm`(仅 models 标签例外)
2. **持久化通过接口**: Repository 接口在 domain 定义，实现在 infra
3. **跨域通信优先接口**: 需要其他域数据时，定义接口（依赖倒置），而非直接 import
4. **Service 层是唯一编排点**: 只有 Services 层可以协调多个领域服务
5. **每个域独立演进**: 新增域只需加文件，不影响其他域
