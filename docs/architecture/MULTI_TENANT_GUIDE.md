# 多租户（Multi-Tenancy）完整指南

> 面向理解与使用：本文解释「多租户是什么、有什么作用、什么场景用、本平台是否支持、如何落地」，并给出一个端到端的完整案例。
> 技术实现细节见同目录《IAM_MULTI_TENANT.md》。

---

## 一、什么是多租户

### 1.1 概念

**多租户（Multi-Tenancy）** 是一种软件架构模式：**一套应用实例同时服务多个「租户」（Tenant），每个租户的数据彼此隔离，互不可见，但共享同一套代码、同一套部署。**

- **租户（Tenant）**：一个独立的组织、团队、公司或客户，是数据隔离的最小边界。
- **单租户（Single-Tenant）**：每个客户一套独立部署（独立代码/独立数据库），隔离最强但成本高。
- **多租户（Multi-Tenant）**：所有客户共用一套部署，靠「租户标识」在数据层做隔离，成本低但需要可靠的隔离机制。

一句话：**多租户 = 一套系统，多个客户各自使用，数据互相看不到。**

### 1.2 三种隔离模型

| 模型 | 数据隔离方式 | 隔离强度 | 成本 | 适用 |
|---|---|---|---|---|
| **共享库共享表（行级隔离）** | 同一张表里用 `tenant_id` 列区分 | 低（依赖代码正确性） | 最低 | 中小型 SaaS |
| **共享库独立 Schema** | 每个租户一个 schema | 中 | 中 | 租户较多、需要更强隔离 |
| **独立数据库** | 每个租户一个库 | 最强 | 最高 | 金融/合规强要求 |

**本平台采用第一种：共享库共享表 + `tenant_id` 行级隔离**（成本最低，适合运维平台这类场景，配合代码层的自动过滤/自动填充保证隔离正确性）。

### 1.3 本平台对「租户」的定义

在本平台中，一个「租户」对应一个**独立的组织/团队**，拥有：

- 自己的**用户**（`user` 表带 `tenant_id`）
- 自己的**角色与权限**（`sys_role` 表带 `tenant_id`）
- 自己的**K8s 集群、应用、流水线、审计日志**等业务数据（各业务表带 `tenant_id`）

租户之间完全隔离，只有「超级管理员」可以跨租户查看。

---

## 二、多租户的作用与价值

| 作用 | 说明 |
|---|---|
| **成本降低** | 一套部署服务所有客户，不用为每个客户单独部署/运维一套系统 |
| **数据安全隔离** | 租户 A 看不到租户 B 的集群、应用、日志、审计，防止数据泄露 |
| **SaaS 化能力** | 平台可以按租户售卖/授权，快速接入新客户，是 SaaS 商业模式的基础 |
| **管理简化** | 一次升级、一次运维，所有租户同时受益 |
| **权限自治** | 每个租户可管理自己的用户、角色、资源，超管只需管租户本身 |

---

## 三、使用场景

1. **SaaS 运维平台**（本平台的典型场景）：平台运营方服务多家企业客户，每家企业是独立租户，各自管理自己的 K8s 集群、CI/CD 流水线、监控告警。
2. **集团/多子公司**：集团总部（超管）+ 各子公司（租户），子公司间业务隔离，总部可全局查看。
3. **多团队/多项目隔离**：同一公司内，研发、测试、生产团队按租户隔离，避免误操作互相影响。
4. **多环境隔离**：开发/测试/生产作为不同租户，同一套平台管理，环境间数据互不干扰。
5. **对外提供托管服务**：平台作为产品卖给第三方，每个客户一个租户。

---

## 四、当前架构能否实现多租户 —— 能，且已实现

本平台**已经完整实现了多租户**，从数据模型、隔离机制、权限模型、跨租户能力四个层面都具备。

### 4.1 租户数据模型

```
tenant（租户） ──1:N──> user（用户） ──M:N──> sys_role（角色） ──M:N──> sys_permission（权限）
```

| 表 | 租户列 | 说明 |
|---|---|---|
| `tenant` | `id` | 租户主表（name 名称、code 编码、status 状态） |
| `user` | `tenant_id` | 用户归属租户，通过 `db.Base` 嵌入 |
| `sys_role` | `tenant_id` | 角色租户隔离（默认租户 1 的角色作为模板） |
| `sys_user_role` | `tenant_id` | 用户-角色关联 |
| `sys_role_permission` | — | 角色权限映射 |
| `audit_log` | `tenant_id` | 审计日志租户隔离 |
| 各业务表（集群/应用/流水线等） | `tenant_id` | 业务数据隔离 |

> `tenant_id` 默认值为 `1`，即「默认租户」。`DefaultTenantID = 1`。

### 4.2 隔离机制（读写双保险）

**写：INSERT 自动填充 tenant_id**

所有嵌入 `db.Base` 的模型，在 `Create` 时由 GORM 回调 `fillTenantID` 自动把当前租户 ID 写入 `tenant_id` 列（仅当该字段为零值时，不覆盖超管的显式跨租户写入）。

```go
// pkg/tenant/scope.go
db.Callback().Create().Before("gorm:create").Register("tenant:fill_tenant_id", fillTenantID)
```

**读：SELECT/UPDATE/DELETE 自动追加 WHERE tenant_id = ?**

请求进入时，`TenantScope` 中间件根据 JWT 里的 `tenant_id` 构造「租户隔离 DB」，之后所有查询自动带上 `WHERE tenant_id = 当前租户`。

```go
// middlewares/tenant.go → pkg/tenant/scope.go
scopedDB := tenant.NewScopedDB(global.DB, tid)   // 自动 WHERE tenant_id = tid
c.Set("db", scopedDB)
```

**为什么必须两个都做**：如果只过滤查询不填充写入，非默认租户新建的数据会因为 `tenant_id` 零值落到 MySQL 默认值 `1`（默认租户），该租户自己反而查不到自己刚建的数据。

### 4.3 权限模型

| 角色 | 范围 | 说明 |
|---|---|---|
| `super_admin` | 全平台 | 跨租户，可看所有租户数据 |
| `platform_admin` | 单租户 | 管理本租户的平台配置/用户/角色 |
| `devops` / `developer` / `tester` / `viewer` | 单租户 | 运维/开发/测试/只读 |

权限按「域 × 级别」控制：

- **域（Scope）**：`platform`（平台）、`cluster`（集群）、`cicd`（CI/CD）
- **级别（AccessLevel）**：`none` / `read` / `write` / `admin`

### 4.4 跨租户能力（超管）

- 超级管理员 `is_super_admin == true`，走 `global.DB`（无 tenant 过滤），可查看全平台数据。
- 审计日志支持超管传 `?tenant_id=3` 筛选特定租户；普通用户传该参数无效（仍被限制在本租户）。

### 4.5 结论

**可以，且已经落地。** 新增业务表时，只要让模型嵌入 `db.Base`，并走 `NewServicesFromContext` 取 DB，就自动获得多租户隔离能力，无需手写 `WHERE tenant_id`。

---

## 五、完整案例（端到端）

### 5.1 场景

某运维平台服务两家客户：**甲公司（租户 A）** 和 **乙公司（租户 B）**。要求：两家公司各自管理自己的 K8s 集群和流水线，互不可见；平台方超管可全局查看。

### 5.2 步骤

**① 超管创建两个租户**

```http
POST /api/v1/platform/tenants
Authorization: Bearer <super_admin_token>

{"name": "甲公司", "code": "company-a"}
{"name": "乙公司", "code": "company-b"}
```

系统在创建租户后，自动调用 `SeedTenantRBAC(tenantID)`，把默认租户的角色+权限模板克隆到新租户（每个租户都有独立的 devops/developer/viewer 等角色）。

**② 租户 A 管理员创建本租户用户**

```http
POST /api/v1/users
Authorization: Bearer <tenant_a_admin_token>

{"username": "alice", "password": "***", "role": "devops"}
```

`alice` 落库时 `user.tenant_id = 租户A的ID`（写入自动填充）。

**③ alice 登录，JWT 携带租户信息**

```
POST /api/v1/auth/login  →  JWT 中包含 tenant_id=租户A
```

**④ alice 创建自己的集群**

```http
POST /api/v1/clusters
Authorization: Bearer <alice_token>

{"name": "甲公司的生产集群", "kubeconfig": "..."}
```

写入时 `fillTenantID` 自动给这条集群记录打上 `tenant_id = 租户A`。

**⑤ 隔离验证**

- alice（租户 A）查询集群列表 → 只看到「甲公司的生产集群」。
- bob（租户 B）查询集群列表 → **看不到** 甲公司的任何集群（Scoped DB 自动 `WHERE tenant_id = 租户B`）。
- 即使 bob 手动构造 `?tenant_id=租户A` 参数，也依然被限制在本租户（普通用户无法越权）。

**⑥ 超管跨租户查看**

超管登录后（`is_super_admin=true`），审计日志、集群列表等走 `global.DB`，能看到甲、乙两家的全部数据，并可用 `tenant_id` 参数筛选。

**⑦ 审计留痕**

alice 的每次操作都写入 `audit_log`，并带上 `tenant_id = 租户A`，超管可在审计中心按租户追溯。

### 5.3 隔离效果总结

| 角色 | 能看到的租户数据 |
|---|---|
| 甲公司用户 alice | 只有租户 A |
| 乙公司用户 bob | 只有租户 B |
| 平台超管 | 全部租户（可筛选） |

---

## 六、操作手册（如何接入新租户）

1. **超管创建租户**：`POST /api/v1/platform/tenants`（name + 唯一 code）。
2. **系统自动初始化**：`SeedTenantRBAC` 克隆默认角色权限到新租户。
3. **租户管理员创建用户**：在租户管理页/用户接口创建用户，用户自动归属当前租户。
4. **分配角色**：给用户绑定 `devops`/`developer`/`viewer` 等本租户角色。
5. **用户使用**：登录后所有操作自动隔离到本租户，无需额外配置。

### 关键 API

| 操作 | 路径 |
|---|---|
| 租户列表/创建 | `GET/POST /api/v1/platform/tenants` |
| 租户更新/删除 | `PUT/DELETE /api/v1/platform/tenants/:id` |
| 用户管理 | `/api/v1/users` |
| 审计日志（超管可跨租户） | `GET /api/v1/platform/audit/logs?tenant_id=N` |
| 平台设置（需超管/平台管理员） | `PUT /api/v1/platform/settings` |

---

## 七、开发新业务模块时如何保持多租户

1. 模型嵌入 `db.Base`（自动获得 `tenant_id` 列）。
2. 控制器用 `middlewares.NewServicesFromContext(ctx)` 获取带租户的 Services。
3. 不要在 SQL 里手写 `WHERE tenant_id`（Scoped DB 已自动加）。
4. 跨租户场景（如超管）显式用 `NewBackgroundServices()`。

---

## 八、后续可增强项

- 前端超管租户下拉筛选（后端 API 已支持 `tenant_id` 参数）
- `sys_user_cluster` 增加 `tenant_id` 列（跨租户集群分配隔离）
- LDAP 登录时指定租户（当前默认落租户 1）
- 平台设置按租户实例化（当前全局单行）
