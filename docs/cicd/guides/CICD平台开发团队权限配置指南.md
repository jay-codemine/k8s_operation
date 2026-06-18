# CI/CD 发布平台 — 开发团队权限配置指南

> 本文档说明如何将 K8sOperation CI/CD 平台开放给开发团队使用，包括权限模型设计、角色分配方案和操作步骤。

---

## 目录

- [权限模型概览](#权限模型概览)
- [内置角色矩阵](#内置角色矩阵)
- [典型团队配置方案](#典型团队配置方案)
- [具体操作步骤](#具体操作步骤)
- [CI/CD 发布权限细化](#cicd-发布权限细化)
- [集群权限细化（命名空间隔离）](#集群权限细化命名空间隔离)
- [安全最佳实践](#安全最佳实践)
- [常见场景 Q&A](#常见场景-qa)
- [API 接口参考](#api-接口参考)

---

## 权限模型概览

平台采用 **三域六角色** 权限设计：

```
┌─────────────────────────────────────────────────────────────┐
│                    三大功能域（Scope）                         │
├─────────────────┬────────────────┬──────────────────────────┤
│  平台域          │  集群域         │  发布域（CI/CD）          │
│  (platform)     │  (cluster)     │  (cicd)                  │
├─────────────────┼────────────────┼──────────────────────────┤
│ 用户管理         │ 集群管理        │ 流水线管理                │
│ 角色权限         │ 工作负载        │ 构建/触发/停止            │
│ 系统设置         │ 服务与路由      │ 部署审批                  │
│ 审计日志         │ 配置管理        │ 制品与镜像                │
│                 │ 存储管理        │ 代码扫描                  │
│                 │ 节点管理        │ 发布记录                  │
│                 │ 监控与日志      │                          │
└─────────────────┴────────────────┴──────────────────────────┘
```

每个域有 **4 个权限级别**：

| 级别 | 标识 | 含义 |
|------|------|------|
| 不可见 | `none` | 功能完全不可见 |
| 只读 | `read` | 只能查看，不能操作 |
| 读写 | `write` | 可创建/编辑/触发，不能删除系统级资源 |
| 全权 | `admin` | 含删除/批量操作/管理他人资源 |

---

## 内置角色矩阵

系统已预置 **6 个角色**（初始化 SQL 自动创建，不可删除）：

| 角色 | 角色标识 | 平台域 | 集群域 | 发布域 | 典型人群 |
|------|---------|--------|--------|--------|---------|
| 超级管理员 | `super_admin` | admin | admin | admin | CTO / 架构师 |
| 平台管理员 | `platform_admin` | admin | read | read | 运维主管 |
| 运维工程师 | `devops` | read | admin | admin | SRE / DevOps |
| **开发工程师** | `developer` | none | **write** | **write** | 后端/前端开发 |
| **测试工程师** | `tester` | none | read | **write** | QA / 测试 |
| 观察者 | `viewer` | read | read | read | 产品 / 项目经理 |

### 开发人员的权限明细

`developer` 角色在 **CI/CD 发布域** 拥有 `write` 级别，具体能做：

| 操作 | 是否允许 | 说明 |
|------|---------|------|
| 查看所有流水线 | ✅ | read 级别以上即可 |
| 创建流水线 | ✅ | write 级别 |
| 编辑自己的流水线 | ✅ | write 级别 |
| 触发构建 | ✅ | write 级别 |
| 停止构建 | ✅ | write 级别（自己触发的） |
| 查看构建日志 | ✅ | read 级别以上 |
| 审批部署 | ✅ | write 级别 |
| 查看发布记录 | ✅ | read 级别以上 |
| 删除流水线 | ❌ | 需要 admin 级别 |
| 修改他人的流水线 | ❌ | 需要 admin 级别 |
| 管理 Jenkins 配置 | ❌ | 需要 admin 级别 |

---

## 典型团队配置方案

### 方案一：小团队（5-15 人）

```
┌─────────────── 团队结构 ──────────────────────┐
│                                               │
│  技术负责人 (1人) → super_admin               │
│      ↓                                        │
│  运维 (1-2人) → devops                        │
│      ↓                                        │
│  后端开发 (3-5人) → developer                 │
│  前端开发 (2-3人) → developer                 │
│  测试 (1-2人) → tester                        │
│  产品经理 (1人) → viewer                      │
│                                               │
└───────────────────────────────────────────────┘
```

**配置步骤：**
1. admin 账户（超级管理员）由技术负责人持有
2. 为每位成员创建账号，分配对应角色
3. 开发人员可自行创建流水线、触发构建
4. 生产环境部署需审批（`require_approval: true`）

### 方案二：中大型团队（多项目组）

```
┌─────────── 平台层 ────────────────────────────────────────────┐
│  平台管理员 → 管理用户/角色/系统设置                             │
│  运维团队   → 管理集群/监控/告警                                │
└───────────────────────────────────────────────────────────────┘
         │
    ┌────┼────────────────────────┐
    ↓    ↓                        ↓
┌──────────┐  ┌──────────┐  ┌──────────┐
│ 项目A     │  │ 项目B     │  │ 项目C     │
│ ns: app-a │  │ ns: app-b │  │ ns: app-c │
│           │  │           │  │           │
│ 开发x3    │  │ 开发x5    │  │ 开发x4    │
│ 测试x1    │  │ 测试x2    │  │ 测试x1    │
└──────────┘  └──────────┘  └──────────┘
```

**配置步骤：**
1. 所有开发 → `developer` 角色（CI/CD write）
2. 通过 **集群权限** 细化：每个项目组只能访问自己的 namespace
3. 流水线按命名空间隔离：项目 A 的流水线只部署到 `app-a` namespace
4. 生产环境统一由运维审批

---

## 具体操作步骤

### Step 1: 创建开发账户

**方式一：通过平台 UI**

> 系统设置 → 用户管理 → 新建用户

**方式二：通过 API**

```bash
# 创建用户
curl -X POST http://your-platform/api/v1/user/create \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "dev_zhangsan",
    "password": "Dev@2024!",
    "email": "zhangsan@company.com",
    "phone": "13800138000"
  }'
```

### Step 2: 分配角色

```bash
# 将用户分配为 developer 角色
# 角色 ID: 4 = developer
curl -X POST http://your-platform/api/v1/rbac/user-role/assign \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "role_ids": [4]
  }'
```

### Step 3: 分配集群权限（可选，精细控制）

```bash
# 给开发者分配指定集群 + 指定命名空间的权限
curl -X POST http://your-platform/api/v1/rbac/cluster-permission/create \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "cluster_id": 1,
    "role_type": "developer",
    "access_level": "write",
    "namespaces": ["dev", "test", "app-a"]
  }'
```

### Step 4: 创建流水线（开发自行操作）

开发人员登录后，可以自行创建流水线：

```bash
# 开发者创建自己项目的流水线
curl -X POST http://your-platform/api/v1/k8s/cicd/pipeline/create \
  -H "Authorization: Bearer <dev-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "app-a-backend",
    "git_repo": "https://gitee.com/company/app-a.git",
    "git_branch": "main",
    "language_type": "go",
    "auto_deploy": true,
    "target_cluster_id": 1,
    "target_namespace": "app-a",
    "target_workload_name": "app-a-backend",
    "target_container": "app-a-backend",
    "deploy_env": "dev",
    "require_approval": false
  }'
```

---

## CI/CD 发布权限细化

### 按环境分级审批

| 环境 | 谁能触发 | 是否需要审批 | 谁来审批 |
|------|---------|------------|---------|
| dev（开发） | developer | 否 | — |
| test（测试） | developer / tester | 否 | — |
| staging（预发） | developer | **是** | devops |
| prod（生产） | devops | **是** | super_admin / devops 互审 |

**实现方式：** 在创建流水线时设置：

```json
{
  "deploy_env": "prod",
  "require_approval": true
}
```

平台会在构建成功后自动创建审批记录，支持：
- 平台内审批
- 飞书卡片审批（已集成）
- 钉钉通知

### 按项目隔离流水线

通过命名空间隔离实现项目间权限分离：

```
项目 A 团队 → 只能看到 target_namespace="app-a" 的流水线
项目 B 团队 → 只能看到 target_namespace="app-b" 的流水线
```

> 当前平台所有 developer 能看到所有流水线（read 级别），如需完全隔离，可通过自定义角色实现。

---

## 集群权限细化（命名空间隔离）

### 场景：不同团队隔离到不同命名空间

```bash
# 给项目 A 开发者 → 只能操作 dev、test 和 app-a 命名空间
curl -X POST http://your-platform/api/v1/rbac/cluster-permission/create \
  -d '{
    "user_id": 2,
    "cluster_id": 1,
    "access_level": "write",
    "namespaces": ["dev", "test", "app-a"]
  }'

# 给项目 B 开发者 → 只能操作 app-b 命名空间
curl -X POST http://your-platform/api/v1/rbac/cluster-permission/create \
  -d '{
    "user_id": 5,
    "cluster_id": 1,
    "access_level": "write",
    "namespaces": ["dev", "test", "app-b"]
  }'
```

### 场景：批量分配多集群权限

```bash
# 运维同时管理开发集群和生产集群
curl -X POST http://your-platform/api/v1/rbac/cluster-permission/batch \
  -d '{
    "user_id": 3,
    "cluster_ids": [1, 2, 3],
    "access_level": "admin"
  }'
```

---

## 安全最佳实践

### 1. 生产环境必须开启审批

```json
{
  "deploy_env": "prod",
  "require_approval": true
}
```

### 2. 开发者不应拥有生产集群写权限

```
开发者：
  - 开发集群: write（可部署）
  - 测试集群: write（可部署）
  - 生产集群: read（只能看，不能操作）
```

### 3. 密钥管理

| 场景 | 建议 |
|------|------|
| Jenkins Token | 只有 admin 配置，开发者不可见 |
| 镜像仓库密码 | 写在 Secret 中，开发者无需知道 |
| KubeConfig | 由平台统一管理，开发者通过平台操作 |

### 4. 审批权限分离

- 不能自己审批自己触发的构建
- 生产部署至少需要一位 devops 或 super_admin 审批
- 审批支持飞书/钉钉通知，避免阻塞

### 5. 自定义角色（高级）

如果内置角色不满足需求，可以创建自定义角色：

```bash
# 创建"实习生"角色 —— CI/CD 只读 + 集群只读
curl -X POST http://your-platform/api/v1/rbac/role/create \
  -d '{
    "name": "intern",
    "display_name": "实习生",
    "description": "仅可查看，不可操作",
    "role_type": "custom",
    "scope_platform": "none",
    "scope_cluster": "read",
    "scope_cicd": "read",
    "color": "#bfbfbf",
    "icon": "user"
  }'
```

---

## 常见场景 Q&A

### Q1: 开发者能自己创建流水线吗？

**可以。** `developer` 角色的 `scope_cicd = write`，允许创建、编辑和触发流水线。

### Q2: 开发者能删除别人的流水线吗？

**不能。** 删除需要 `admin` 级别（devops 或 super_admin 才行）。

### Q3: 如何让开发者只能部署到开发环境？

两种方式：
1. **集群权限隔离**：只给开发者分配开发集群的 write 权限
2. **流水线审批**：生产环境的流水线设置 `require_approval: true`

### Q4: 测试人员需要什么权限？

`tester` 角色已预置：
- 集群域 `read`：可以查看 Pod 状态、查看日志
- 发布域 `write`：可以触发测试环境的构建

### Q5: 新入职的开发怎么快速配置？

```bash
# 1. 创建账号
# 2. 分配 developer 角色
# 3. 分配所在项目的命名空间权限
# 一步到位的 API 调用：

# 创建用户
curl -X POST /api/v1/user/create -d '{"username":"new_dev","password":"Init@2024!"}'

# 分配角色（developer = ID 4）
curl -X POST /api/v1/rbac/user-role/assign -d '{"user_id":NEW_ID,"role_ids":[4]}'

# 分配集群权限
curl -X POST /api/v1/rbac/cluster-permission/create -d '{
  "user_id": NEW_ID, "cluster_id": 1,
  "access_level": "write", "namespaces": ["dev","test","team-ns"]
}'
```

### Q6: 如何查看某用户的完整权限？

```bash
# 获取用户完整 RBAC 信息（含三域 Scope + 集群权限）
curl -X GET http://your-platform/api/v1/rbac/user/permissions \
  -H "Authorization: Bearer <user-token>"
```

返回示例：
```json
{
  "user_id": 2,
  "username": "dev_zhangsan",
  "is_super_admin": false,
  "roles": [
    {
      "name": "developer",
      "display_name": "开发工程师",
      "scope_platform": "none",
      "scope_cluster": "write",
      "scope_cicd": "write"
    }
  ],
  "scopes": {
    "platform": "none",
    "cluster": "write",
    "cicd": "write"
  },
  "cluster_permissions": [
    {
      "cluster_id": 1,
      "cluster_name": "dev-cluster",
      "access_level": "write",
      "ns_list": ["dev", "test", "app-a"]
    }
  ]
}
```

---

## API 接口参考

### 角色管理

| 方法 | 路径 | 说明 | 权限要求 |
|------|------|------|---------|
| GET | `/api/v1/rbac/role/list` | 角色列表（分页） | platform:admin |
| GET | `/api/v1/rbac/role/all` | 所有角色（下拉） | platform:admin |
| POST | `/api/v1/rbac/role/create` | 创建角色 | platform:admin |
| POST | `/api/v1/rbac/role/update` | 更新角色 | platform:admin |
| POST | `/api/v1/rbac/role/delete` | 删除角色 | platform:admin |
| GET | `/api/v1/rbac/role/permissions` | 角色权限详情 | platform:admin |
| POST | `/api/v1/rbac/role/permissions/update` | 更新角色权限 | platform:admin |
| GET | `/api/v1/rbac/role/users` | 角色绑定用户列表 | platform:admin |

### 用户角色管理

| 方法 | 路径 | 说明 | 权限要求 |
|------|------|------|---------|
| POST | `/api/v1/rbac/user-role/assign` | 分配用户角色 | platform:admin |
| GET | `/api/v1/rbac/user-role/list` | 用户角色列表 | platform:admin |

### 集群权限管理

| 方法 | 路径 | 说明 | 权限要求 |
|------|------|------|---------|
| POST | `/api/v1/rbac/cluster-permission/create` | 创建集群权限 | platform:admin |
| POST | `/api/v1/rbac/cluster-permission/update` | 更新集群权限 | platform:admin |
| POST | `/api/v1/rbac/cluster-permission/delete` | 删除集群权限 | platform:admin |
| POST | `/api/v1/rbac/cluster-permission/batch` | 批量分配 | platform:admin |
| GET | `/api/v1/rbac/cluster-permission/list` | 集群权限列表 | platform:admin |

### 权限检查

| 方法 | 路径 | 说明 | 权限要求 |
|------|------|------|---------|
| GET | `/api/v1/rbac/user/permissions` | 当前用户完整权限 | 登录即可 |
| GET | `/api/v1/rbac/check/cluster` | 检查集群权限 | 登录即可 |

---

## 推荐的上线检查清单

在开放给开发团队之前，请确认：

- [ ] 初始化 SQL 已执行（含 6 个内置角色 + 13 个权限 + admin 账户）
- [ ] 修改了 admin 默认密码（admin123 → 强密码）
- [ ] 为每位团队成员创建了账号
- [ ] 为开发者分配了 `developer` 角色
- [ ] 为测试人员分配了 `tester` 角色
- [ ] 为运维分配了 `devops` 角色
- [ ] 配置了集群权限（命名空间隔离）
- [ ] 生产环境流水线开启了审批（`require_approval: true`）
- [ ] Jenkins 连接正常（构建能触发）
- [ ] 目标 K8s 集群的 Deployment/Service 已提前创建
- [ ] 钉钉/飞书通知已配置（审批通知）

完成以上步骤，你的 CI/CD 平台就可以安全地交给开发团队使用了！
