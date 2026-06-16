# RBAC 三域六角色权限设计文档

> 版本：v2.0 | 更新时间：2026-05-18

---

## 一、设计概述

本系统采用 **三域六角色** RBAC 权限模型，将平台功能划分为三大功能域，预设六个系统角色，每个角色在每个域内拥有独立的权限级别。

### 核心思想

- **域（Scope）**：按业务边界划分功能区域，避免权限爆炸
- **角色（Role）**：预设覆盖 90% 场景的系统角色，支持自定义扩展
- **级别（Level）**：每域四级递增权限，简单直观
- **集群级细控**：在角色级别之上，支持按集群 + 命名空间精确控制

### 设计目标

| 目标 | 说明 |
|------|------|
| 简洁 | 从 26 个细粒度操作点简化为 3 域 × 4 级别 = 12 个维度 |
| 直觉 | 运维、开发、测试人员凭角色名即可理解权限边界 |
| 兼容 | 平滑兼容旧 `can_view/can_create/...` 布尔字段 |
| 可扩展 | 支持自定义角色，支持集群/命名空间级别覆盖 |

---

## 二、三大功能域

| 域 | 标识 | 图标 | 覆盖范围 |
|----|------|------|----------|
| **平台域** | `platform` | 🏛 | 用户管理、角色管理、系统设置、审计日志、数据源、告警配置 |
| **集群域** | `cluster` | ☸ | 集群管理、命名空间、工作负载、Service/Ingress、ConfigMap/Secret、存储、节点、监控、日志、Web终端 |
| **发布域** | `cicd` | 🚀 | 流水线、构建任务、部署发布、制品管理、镜像仓库、代码扫描、发布审批 |

---

## 三、四级权限级别

每个域内的权限级别为单调递增关系：

```
none (0) < read (1) < write (2) < admin (3)
```

| 级别 | 标识 | 说明 |
|------|------|------|
| 不可见 | `none` | 该域功能完全不可见 |
| 只读 | `read` | 可查看，不可修改 |
| 读写 | `write` | 可创建/更新/触发，不可删除系统级资源 |
| 全权 | `admin` | 含删除、批量操作、管理他人资源 |

**级别比较函数**（后端 Go）：

```go
func AccessLevelGte(a, b string) bool {
    return accessLevelOrder(a) >= accessLevelOrder(b)
}
```

---

## 四、六大系统角色

| # | 角色 | 标识 | 平台域 | 集群域 | 发布域 | 典型职责 |
|---|------|------|--------|--------|--------|----------|
| 1 | 超级管理员 | `super_admin` | admin | admin | admin | 系统所有者，拥有全部权限 |
| 2 | 平台管理员 | `platform_admin` | admin | read | read | 管理用户/角色/系统设置，集群和CI/CD只读 |
| 3 | 运维工程师 | `devops` | read | admin | admin | 集群全权+CI/CD全权，平台域只读 |
| 4 | 开发工程师 | `developer` | none | write | write | 集群域读写(指定NS)+CI/CD读写(自己的流水线) |
| 5 | 测试工程师 | `tester` | none | read | write | 集群域只读(指定NS)+CI/CD读写(测试环境流水线) |
| 6 | 观察者 | `viewer` | read | read | read | 全域只读，无任何修改权限 |

> 支持自定义角色：通过管理界面创建新角色，自由配置三域级别。

---

## 五、数据库表设计

### 5.1 sys_role（角色表）

```sql
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id`             BIGINT PRIMARY KEY AUTO_INCREMENT,
  `name`           VARCHAR(50) NOT NULL UNIQUE COMMENT '角色标识',
  `display_name`   VARCHAR(100) COMMENT '显示名称',
  `description`    VARCHAR(500) COMMENT '描述',
  `role_type`      VARCHAR(30) NOT NULL COMMENT '角色类型',
  `scope_platform` VARCHAR(10) DEFAULT 'none' COMMENT '平台域级别: none/read/write/admin',
  `scope_cluster`  VARCHAR(10) DEFAULT 'none' COMMENT '集群域级别: none/read/write/admin',
  `scope_cicd`     VARCHAR(10) DEFAULT 'none' COMMENT '发布域级别: none/read/write/admin',
  `is_system`      TINYINT(1) DEFAULT 0 COMMENT '是否系统内置(不可删除)',
  `color`          VARCHAR(20) DEFAULT '#1890ff' COMMENT '角色颜色标识',
  `icon`           VARCHAR(50) DEFAULT 'user' COMMENT '图标',
  `sort_order`     INT DEFAULT 0 COMMENT '排序',
  `created_at`     BIGINT UNSIGNED,
  `modified_at`    BIGINT UNSIGNED,
  `deleted_at`     BIGINT UNSIGNED DEFAULT 0,
  `is_del`         TINYINT(1) DEFAULT 0
);
```

### 5.2 sys_permission（权限定义表）

```sql
CREATE TABLE IF NOT EXISTS `sys_permission` (
  `id`            BIGINT PRIMARY KEY AUTO_INCREMENT,
  `name`          VARCHAR(100) NOT NULL UNIQUE COMMENT '权限标识',
  `display_name`  VARCHAR(100) COMMENT '显示名称',
  `description`   VARCHAR(500) COMMENT '描述',
  `scope`         VARCHAR(20) DEFAULT 'cluster' COMMENT '所属功能域: platform/cluster/cicd',
  `resource_type` VARCHAR(50) COMMENT '模块标识',
  `action`        VARCHAR(30) COMMENT '操作类型',
  `parent_id`     BIGINT DEFAULT 0 COMMENT '父权限ID',
  `path`          VARCHAR(200) COMMENT '权限路径(树形展示)',
  `sort_order`    INT DEFAULT 0,
  `created_at`    BIGINT UNSIGNED,
  `modified_at`   BIGINT UNSIGNED
);
```

**预置权限（13条模块级权限）**：

| ID | 标识 | 域 | 模块 |
|----|------|-----|------|
| 1 | platform:user | platform | 用户管理 |
| 2 | platform:role | platform | 角色权限 |
| 3 | platform:settings | platform | 系统设置 |
| 4 | platform:audit | platform | 审计日志 |
| 5 | cluster:manage | cluster | 集群管理 |
| 6 | cluster:workload | cluster | 工作负载 |
| 7 | cluster:network | cluster | 服务与路由 |
| 8 | cluster:config | cluster | 配置管理 |
| 9 | cluster:storage | cluster | 存储管理 |
| 10 | cluster:node | cluster | 节点管理 |
| 11 | cluster:monitor | cluster | 监控与日志 |
| 12 | cicd:pipeline | cicd | 流水线管理 |
| 13 | cicd:artifact | cicd | 制品与镜像 |

### 5.3 sys_user_role（用户角色关联表）

```sql
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `id`         BIGINT PRIMARY KEY AUTO_INCREMENT,
  `user_id`    BIGINT NOT NULL,
  `role_id`    BIGINT NOT NULL,
  `created_at` BIGINT UNSIGNED,
  `created_by` BIGINT DEFAULT 0,
  INDEX idx_user_id (`user_id`),
  INDEX idx_role_id (`role_id`)
);
```

### 5.4 sys_role_permission（角色权限关联表）

```sql
CREATE TABLE IF NOT EXISTS `sys_role_permission` (
  `id`            BIGINT PRIMARY KEY AUTO_INCREMENT,
  `role_id`       BIGINT NOT NULL,
  `permission_id` BIGINT NOT NULL,
  `created_at`    BIGINT UNSIGNED,
  INDEX idx_role_id (`role_id`),
  INDEX idx_perm_id (`permission_id`)
);
```

### 5.5 sys_user_cluster（集群级细粒度权限表）

```sql
CREATE TABLE IF NOT EXISTS `sys_user_cluster` (
  `id`           BIGINT PRIMARY KEY AUTO_INCREMENT,
  `user_id`      BIGINT NOT NULL,
  `cluster_id`   BIGINT NOT NULL,
  `role_type`    VARCHAR(30) COMMENT '在该集群的角色类型',
  `access_level` VARCHAR(10) DEFAULT 'read' COMMENT '权限级别: none/read/write/admin',
  `namespaces`   TEXT COMMENT '可访问命名空间(JSON数组,空=全部)',
  `can_view`     TINYINT(1) DEFAULT 1 COMMENT 'deprecated',
  `can_create`   TINYINT(1) DEFAULT 0 COMMENT 'deprecated',
  `can_update`   TINYINT(1) DEFAULT 0 COMMENT 'deprecated',
  `can_delete`   TINYINT(1) DEFAULT 0 COMMENT 'deprecated',
  `can_exec`     TINYINT(1) DEFAULT 0 COMMENT 'deprecated',
  `expire_at`    BIGINT UNSIGNED DEFAULT 0 COMMENT '过期时间(0=永不过期)',
  `created_at`   BIGINT UNSIGNED,
  `modified_at`  BIGINT UNSIGNED,
  `created_by`   BIGINT DEFAULT 0,
  INDEX idx_user_id (`user_id`),
  INDEX idx_cluster_id (`cluster_id`)
);
```

---

## 六、权限判定流程

### 6.1 整体流程图

```
┌─────────────────────────────────────────────────────┐
│              用户发起请求                              │
└─────────────────────┬───────────────────────────────┘
                      ▼
           ┌─────────────────────┐
           │  是否超级管理员？     │
           └─────────┬───────────┘
                     │
              ┌──────┴──────┐
              │ Yes         │ No
              ▼             ▼
         ✅ 直接放行    ┌─────────────────────┐
                        │ 获取用户所有角色      │
                        └─────────┬───────────┘
                                  ▼
                       ┌─────────────────────────┐
                       │ 取所有角色在目标域的最高级别│
                       │ (scope_platform/cluster/ │
                       │  scope_cicd 取 MAX)      │
                       └─────────┬───────────────┘
                                 ▼
                      ┌─────────────────────────┐
                      │ 有效级别 >= 所需最低级别？ │
                      └─────────┬───────────────┘
                                │
                         ┌──────┴──────┐
                         │ Yes         │ No
                         ▼             ▼
                    ✅ 允许     ┌───────────────────┐
                               │ 回退: 检查集群级    │
                               │ 细粒度权限表        │
                               └─────────┬─────────┘
                                         ▼
                               ┌───────────────────┐
                               │ access_level >= ?  │
                               └─────────┬─────────┘
                                         │
                                  ┌──────┴──────┐
                                  │ Yes         │ No
                                  ▼             ▼
                             ✅ 允许       ❌ 拒绝
```

### 6.2 后端检查入口

**核心方法**（`internal/app/services/rbac.go`）：

```go
// CheckScopePermission 检查用户是否满足指定域的最低权限级别
func (s *Services) CheckScopePermission(userID int64, scope, minLevel string) bool {
    if s.dao.IsSuperAdmin(userID) {
        return true
    }
    roles, err := s.dao.UserRoleList(userID)
    for _, role := range roles {
        if models.AccessLevelGte(role.GetEffectiveScope(scope), minLevel) {
            return true
        }
    }
    return false
}
```

**获取完整权限信息**（API: `GET /api/v1/rbac/user/permissions`）：

```go
// 返回结构
type UserWithRBACInfo struct {
    UserID             int64                            `json:"user_id"`
    Username           string                           `json:"username"`
    IsSuperAdmin       bool                             `json:"is_super_admin"`
    Roles              []*models.SysRole                `json:"roles"`
    ClusterPermissions []*models.ClusterPermissionDetail `json:"cluster_permissions"`
    Scopes             *UserScopes                      `json:"scopes"`
}

type UserScopes struct {
    Platform string `json:"platform"` // none/read/write/admin
    Cluster  string `json:"cluster"`
    CICD     string `json:"cicd"`
}
```

### 6.3 前端路由守卫

路由权限配置（`k8s-web/src/stores/permission.js`）：

```javascript
export const ROUTE_SCOPES = {
  // 🏛 平台域
  '/security/users':      { scope: 'platform', minLevel: 'admin' },
  '/security/roles':      { scope: 'platform', minLevel: 'admin' },
  '/security/audit':      { scope: 'platform', minLevel: 'read' },

  // ☸ 集群域
  '/clusters':            { scope: 'cluster', minLevel: 'read' },
  '/environments':        { scope: 'cluster', minLevel: 'read' },
  '/monitoring':          { scope: 'cluster', minLevel: 'read' },

  // 🚀 发布域
  '/cicd/pipelines':       { scope: 'cicd', minLevel: 'read' },
  '/cicd/pipelines/create': { scope: 'cicd', minLevel: 'write' },
  '/cicd/templates':       { scope: 'cicd', minLevel: 'admin' },
  '/cicd/artifacts':       { scope: 'cicd', minLevel: 'read' },
}
```

**核心检查函数**：

```javascript
function hasScopeAccess(scope, minLevel = 'read') {
  if (state.isSuperAdmin) return true
  const userLevel = state.scopes[scope] || 'none'
  return levelGte(userLevel, minLevel)
}

function canAccessMenu(path) {
  if (state.isSuperAdmin) return true
  const routeScope = ROUTE_SCOPES[path]
  if (!routeScope) return true // 未配置则默认允许
  return hasScopeAccess(routeScope.scope, routeScope.minLevel)
}
```

---

## 七、集群级细粒度控制

在角色维度之外，`sys_user_cluster` 表提供更精细的集群级权限控制：

### 7.1 access_level 字段

- 优先使用 `access_level`（新字段），兼容旧 `can_view/can_create/...` 布尔字段
- 旧数据自动推导逻辑：

```go
func (c *SysUserCluster) GetAccessLevel() string {
    if c.AccessLevel != "" && c.AccessLevel != "none" {
        return c.AccessLevel
    }
    // 兼容旧数据
    if c.CanDelete       { return "admin" }
    if c.CanCreate || c.CanUpdate || c.CanExec { return "write" }
    if c.CanView         { return "read" }
    return "none"
}
```

### 7.2 命名空间限制

- `namespaces` 字段为 JSON 数组，如 `["default","production"]`
- 空数组或 `["*"]` 表示允许访问所有命名空间
- 前端 `filterNamespaces()` 方法据此过滤列表展示

### 7.3 权限过期

- `expire_at` 为 Unix 时间戳，`0` 表示永不过期
- 可为临时用户设置有时限的集群访问权

---

## 八、关键代码文件索引

| 文件 | 说明 |
|------|------|
| `internal/app/models/sys_rbac.go` | RBAC 数据模型、常量定义、工具函数 |
| `internal/app/services/rbac.go` | RBAC 业务逻辑层（权限检查、信息聚合） |
| `internal/app/dao/rbac.go` | 数据访问层（DB 查询） |
| `internal/app/controllers/api/v1/rbac/rbac.go` | API 控制器 |
| `k8s-web/src/stores/permission.js` | 前端权限状态管理（三域判断核心） |
| `k8s-web/src/router/index.js` | 路由守卫（beforeEach） |
| `k8s-web/src/api/rbac.js` | 前端 RBAC API 调用 |
| `initialize/db.go` | 启动时 AutoMigrate + 数据回填 |
| `docs/sql/k8s_platform_full_init.sql` | 全量 SQL 初始化（含角色/权限种子数据） |

---

## 九、API 接口

### 9.1 获取用户权限

```
GET /api/v1/rbac/user/permissions
Authorization: Bearer <token>
```

**响应示例**：

```json
{
  "code": 0,
  "data": {
    "user_id": 1,
    "username": "admin",
    "is_super_admin": true,
    "scopes": {
      "platform": "admin",
      "cluster": "admin",
      "cicd": "admin"
    },
    "roles": [
      {
        "id": 1,
        "name": "super_admin",
        "display_name": "超级管理员",
        "role_type": "super_admin",
        "scope_platform": "admin",
        "scope_cluster": "admin",
        "scope_cicd": "admin"
      }
    ],
    "cluster_permissions": [
      {
        "cluster_id": 1,
        "cluster_name": "prod-cluster",
        "access_level": "admin",
        "namespaces": "[]"
      }
    ]
  }
}
```

### 9.2 角色管理接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/rbac/roles` | 获取角色列表 |
| POST | `/api/v1/rbac/roles` | 创建角色 |
| PUT | `/api/v1/rbac/roles/:id` | 更新角色 |
| DELETE | `/api/v1/rbac/roles/:id` | 删除角色（系统角色不可删除） |
| POST | `/api/v1/rbac/users/:id/roles` | 分配用户角色 |
| DELETE | `/api/v1/rbac/users/:id/roles/:roleId` | 移除用户角色 |

### 9.3 集群权限接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/rbac/users/:id/clusters` | 获取用户集群权限 |
| POST | `/api/v1/rbac/users/:id/clusters` | 分配集群权限 |
| PUT | `/api/v1/rbac/users/:id/clusters/:clusterId` | 更新集群权限 |
| DELETE | `/api/v1/rbac/users/:id/clusters/:clusterId` | 移除集群权限 |

---

## 十、存量升级与兼容

### 10.1 自动迁移

后端启动时通过 GORM AutoMigrate 自动为现有表补齐新字段：

```go
// initialize/db.go
global.DB.AutoMigrate(
    &models.SysRole{},        // 补齐 scope_platform/scope_cluster/scope_cicd
    &models.SysPermission{},  // 补齐 scope
    &models.SysUserCluster{}, // 补齐 access_level
)
```

### 10.2 数据回填

启动时 `backfillRBACScopes()` 自动回填存量角色的 scope 值：

- 仅当三域均为默认 `none` 时才触发（不覆盖已配置值）
- `cluster_admin` 旧类型自动映射为 `devops`
- `sys_user_cluster.access_level` 从旧布尔字段推导

### 10.3 SQL 脚本

全量初始化脚本 `docs/sql/k8s_platform_full_init.sql` 包含：

1. 幂等 ALTER 补丁（`IF NOT EXISTS` 逻辑）
2. 种子数据（6 角色 + 13 权限 + 关联）
3. 存量回填 SQL（UPDATE WHERE scope=none）

---

## 十一、设计扩展点

| 扩展方向 | 说明 |
|----------|------|
| 自定义角色 | 管理员可通过 UI 创建角色并自由组合三域级别 |
| 命名空间级权限 | `sys_user_cluster.namespaces` 支持 JSON 数组精确控制 |
| 时效性权限 | `expire_at` 支持临时授权 |
| 操作审计 | 所有权限变更均记录审计日志 |
| 多集群联邦 | 每个集群可独立分配 `access_level`，不受全局角色限制 |

---

## 十二、FAQ

**Q: 用户拥有多个角色，权限如何计算？**

A: 取所有角色在每个域的最高级别（取 MAX），即「就高不就低」原则。

**Q: 超级管理员的权限从哪判断？**

A: 通过 `sys_user_role` 关联到 `role_type = 'super_admin'` 的角色即可。后端 `IsSuperAdmin()` 方法直接走 JOIN 查询。

**Q: 旧版 can_view/can_create 字段还有效吗？**

A: 兼容保留。`GetAccessLevel()` 方法优先读 `access_level`，若为空则自动从布尔字段推导。新数据不再写布尔字段。

**Q: 路由没有配 ROUTE_SCOPES 会怎样？**

A: 未配置的路由默认允许所有已登录用户访问（`if (!routeScope) return true`）。

**Q: 如何添加新的功能域？**

A: 在 `sys_role` 表新增 `scope_xxx` 列，在 `models/sys_rbac.go` 添加常量和 `GetEffectiveScope()` 分支，前端 `permission.js` 同步新增域即可。
