# IAM 多租户架构设计

## 1. 租户模型

### 1.1 数据模型

```
sys_tenant ──1:N──> sys_user ──M:N──> sys_role ──M:N──> sys_permission
```

| 表 | 租户列 | 说明 |
|---|---|---|
| `sys_tenant` | `id` | 租户主表 |
| `sys_user` | `tenant_id` | 用户归属租户，通过 `db.Base` 嵌入 |
| `sys_role` | `tenant_id` | 角色租户隔离（默认租户的角色为模板） |
| `sys_user_role` | `tenant_id` | 用户-角色关联（通过角色间接关联租户） |
| `sys_role_permission` | — | 角色权限映射（依赖所属角色） |
| `sys_permission` | — | 全局权限目录，不区分租户 |
| `audit_log` | `tenant_id` | 审计日志租户隔离 |

### 1.2 租户隔离机制

**GORM 回调自动注入**（`pkg/tenant/tenant.go`）：

```go
// INSERT 时自动填充 tenant_id
db.Callback().Create().Before("gorm:create").Register("fillTenantID", fillTenantID)

// SELECT/UPDATE/DELETE 时自动追加 WHERE tenant_id = ?
scopedDB := tenant.NewScopedDB(global.DB, tenantID)
```

**TenantScope 中间件**（`middlewares/tenant.go`）：

```go
func TenantScope() gin.HandlerFunc {
    return func(c *gin.Context) {
        tenantID, _ := c.Get("tenant_id")  // AuthJWT 中间件设置
        if tid, ok := tenantID.(uint32); ok && tid > 0 {
            c.Set("db", tenant.NewScopedDB(global.DB, tid))
        }
        c.Next()
    }
}
```

**Services 构造函数**（`internal/app/services/services.go`）：

| 构造函数 | DB | tenantID | 使用场景 |
|---|---|---|---|
| `NewServices()` | `global.DB` | 0 | 启动期 |
| `NewServicesWithDB(db)` | Scoped DB | 从 DB 派生 | HTTP 租户请求 |
| `NewBackgroundServices()` | `global.DB` | 0 | 后台跨租户 |

### 1.3 租户初始化流程

1. 超管创建租户 → `sys_tenant` 插入
2. `SeedTenantRBAC(tenantID)` 克隆默认租户的角色+权限到新租户
3. 租户管理员创建用户 → `sys_user` 写入对应 `tenant_id`

## 2. 权限模型

### 2.1 角色分层

| 角色 | 范围 | 说明 |
|---|---|---|
| `super_admin` | 全平台 | 跨租户，可查看所有租户数据 |
| `platform_admin` | 单个租户 | 管理租户内的平台配置/用户/角色 |
| `devops` | 单个租户 | 运维操作（K8s 资源管理） |
| `developer` | 单个租户 | 开发操作（应用部署/日志查看） |
| `tester` | 单个租户 | 测试操作（测试环境操作） |
| `viewer` | 单个租户 | 只读 |

### 2.2 权限域（Scope）

```go
ScopePlatform = "platform"  // 平台管理权限
ScopeCluster  = "cluster"   // 集群操作权限
ScopeCICD     = "cicd"      // CI/CD 操作权限
```

每个域支持 4 个访问级别：`none` / `read` / `write` / `admin`。

### 2.3 鉴权流程

```
请求 → AuthJWT(解析JWT，设置user_id/tenant_id/is_super_admin)
     → TenantScope(按tenant_id注入scoped DB)
     → 业务 Handler
       ├─ 租户操作：NewServicesFromContext(ctx) → Scoped DB
       ├─ 超管跨租户：NewBackgroundServices() → global.DB + tenant_id筛选
       └─ 平台写操作：requirePlatformWrite(ctx) → 校验 is_super_admin || ScopePlatform.admin
```

## 3. 跨租户审计

### 3.1 超管查看全部租户审计

超管登录后 `ctx.GetBool("is_super_admin") == true`，审计列表/详情/统计/导出自动切换为跨租户服务：

```go
func auditLogServices(ctx *gin.Context) *services.Services {
    if ctx.GetBool("is_super_admin") {
        return services.NewBackgroundServices() // global.DB，无 tenant 过滤
    }
    return middlewares.NewServicesFromContext(ctx) // Scoped DB
}
```

前端传 `?tenant_id=3` 可筛选特定租户日志（后端 `AuditLogQuery.TenantID` 字段 + `audit_repo` 条件过滤）。不传则返回全量。

### 3.2 普通用户隔离

普通用户 `is_super_admin == false` → `NewServicesFromContext` → Scoped DB（自动 `WHERE tenant_id=X`），额外传 `tenant_id` 参数无法越权。

### 3.3 登录审计完整性

登录是公开路由，不经过 `AuthJWT`。修复后登录 handler 在 `c.Next()` 前显式注入上下文：

```go
ctx.Set("user_id", user.ID)
ctx.Set("current_user_name", user.Username)
ctx.Set("tenant_id", user.TenantID)
```

审计中间件（全局 `s.Use(Audit(nil))`）在 `c.Next()` 之后读取这些值，登录事件完整记录。

## 4. 平台写操作鉴权

| 操作 | 权限要求 | 实现 |
|---|---|---|
| `PUT /platform/settings` | 超管 / 平台管理员 | `requirePlatformWrite(ctx)` |
| `POST /platform/settings/reset` | 超管 / 平台管理员 | `requirePlatformWrite(ctx)` |
| `PUT /platform/audit/retention` | 超管 / 平台管理员 | `requirePlatformWrite(ctx)` |
| `POST /platform/audit/cleanup` | 超管 / 平台管理员 | `requirePlatformWrite(ctx)` |
| `GET /platform/settings` | 任意登录用户 | 无校验 |
| `GET /platform/audit/retention` | 任意登录用户 | 无校验 |

## 5. JIT 用户归属

| 场景 | tenantID 来源 | 默认行为 |
|---|---|---|
| 本地注册 | `Create(name, password, tenantID)` | 传入 0 → 默认租户 1 |
| LDAP 首次登录 (JIT) | `syncLDAPUser` → `CreateFull(..., s.tenantID)` | 登录时 `s.tenantID=0` → 默认租户 1 |
| 批量导入 | `BatchImport(..., s.tenantID)` | 租户 Scoped 上下文 → 当前租户 |
| Web 注册 | `Register(username, password)` → `Create(name, password, 0)` | 默认租户 1 |

## 6. 本次修复对照表

### 2026-08-12 IAM 多租户改造

| 问题 | 修复 | 影响文件 |
|---|---|---|
| 超管无法跨租户查看审计 | `auditLogServices` 按 `is_super_admin` 切换服务；`AuditLogQuery.TenantID` 筛选字段 | `audit_log_controller.go`, `models.go`, `audit_repo.go` |
| 登录审计 tenant_id=0 | login handler 显式 `ctx.Set("tenant_id", ...)` | `login.go` |
| 平台设置无权限校验 | `requirePlatformWrite` 守卫 Update/Reset | `platform_settings_controller.go`, `authz.go` |
| 审计保留策略无权限校验 | `requirePlatformWrite` 守卫 UpdateRetention/Cleanup | `audit_log_controller.go` |
| CreateFull 未设置租户 | 新增 `tenantID` 参数，默认 1 | `service.go` (user domain) |
| 批量导入未传递租户 | `s.tenantID` 传入 `BatchImport` | `user.go` (services) |
| LDAP JIT 未设置租户 | `s.tenantID` 传入 `CreateFull` | `ldap.go` |

### 2026-08-11 审计租户隔离修复（前一轮）

| 问题 | 修复 |
|---|---|
| `AuditLog` 无 `TenantID` 字段 | 添加字段 + GORM tag |
| 异步写入无租户上下文 | 提取 `c.Get("tenant_id")` 传递 |
| `ResponseMessage` 未填充 | 从 `c.Errors` 提取 |
| `ClusterName` 未填充 | `GetCluster` 异步查询 |

## 7. 后续可选项

- 前端超管租户下拉筛选（后端 API 已支持 `tenant_id` 参数）
- 审计保留策略迁入 `Settings` 域（消除 `domain/audit/service.go` 中的跨域 `sys_settings` 直查）
- `sys_user_cluster` 增加 `tenant_id` 列（跨租户集群分配隔离）
- LDAP 租户选择（登录时指定租户而非默认 1）
- 平台设置实例化（当前全局单行，未来可按租户存储部分设置）
