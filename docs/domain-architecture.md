# K8s Operation 领域驱动架构设计

## 目录

1. [什么是领域](#1-什么是领域)
2. [当前架构 vs 领域架构](#2-当前架构-vs-领域架构)
3. [本项目领域划分](#3-本项目领域划分)
4. [领域内部架构](#4-领域内部架构)
5. [领域之间如何通信](#5-领域之间如何通信)
6. [完整请求流转](#6-完整请求流转)
7. [对比总结](#7-对比总结)
8. [迁移路径](#8-迁移路径)
9. [企业级演进](#9-企业级演进)

---

## 1. 什么是领域

**领域 = 业务中一块独立的知识和职责范围。** 开发人员如果不深入理解这个范围，就看不懂代码在干什么。

### 判断标准

一个合格的领域满足三个条件：

1. **独立的语言体系** — 有自己的术语，外行听不懂
2. **完整的内聚逻辑** — 领域的业务流程是自包含的
3. **明确的边界** — 对外只暴露有限的接口

### 例子

| 领域 | 术语 | 核心职责 |
|------|------|----------|
| K8s 资源管理 | Pod、Deployment、Node、Namespace、Label、Taint | 操作 K8s 集群资源 |
| CICD | Pipeline、Stage、Artifact、Release、Rollout | 流水线构建与发布 |
| 监控告警 | Metric、Alert、Threshold、Datasource、Silence | 指标采集与告警管理 |

同一个人在不同领域语境下思维模式完全不同。**领域划分就是把这些不同语境显式地隔离出来。**

---

## 2. 当前架构 vs 领域架构

### 当前：按技术分层（水平切）

```
internal/app/
  controllers/     ← 所有 Controller 混在一起
    cicd/
    deployment/
    pod/
    monitoring/
  services/        ← 所有 Service 混在 Services struct 上（~30k 行，200+ 方法）
    cicd_pipeline.go
    k8s_deployment.go
    monitoring.go
    ...（60+ 文件）
  models/          ← 所有表混在一起
  dao/             ← 所有数据访问混在一起
```

**问题：**

- 找代码要翻 3 层目录："创建 Pipeline 的逻辑在哪？"
- 改 CICD 可能影响 Services struct 上 200+ 方法
- 所有方法挂一个 struct，Mock 测试困难
- 跨域调用和域内调用没有区别，边界模糊

### 目标：按业务领域切（垂直切）

```
internal/domain/
  k8s/             ← K8s 资源管理（Pod / Deploy / Svc / Node...）
    handler/       ← HTTP 适配层
    service/       ← 业务逻辑层（原来 16 个 k8s_*.go 文件）
    repository/    ← 数据访问层（只操作 K8s 自己的表）
  cicd/            ← CICD 流水线（Pipeline / Stage / Release...）
    handler/
    service/       ← 原来 23 个 cicd_*.go 文件
    repository/
  monitor/         ← 监控告警（Prometheus / Loki / AlertRule...）
    handler/
    service/
    repository/
  auth/            ← 认证授权
  image/           ← 镜像管理
  appstore/        ← 应用商城
  platform/        ← 平台管理（健康检查、设置、审计）
  shared/          ← 跨领域共享
    auth.go        ← JWT 校验
    tenant.go      ← 租户隔离
    k8s_client.go  ← ClusterClientFactory
```

**收益：**

| | 当前 | 领域化后 |
|---|---|---|
| 找代码 | 翻 3 层目录 | 打开 `domain/cicd/` 一目了然 |
| 改代码 | 影响 Services 上 200+ 方法 | 只影响本域 Service 接口 |
| 测试 | Mock 整个 Services/DAO | Mock 本域 5-10 个方法的 interface |
| 新人上手 | 需理解整个 Services 结构 | 只需理解分配的领域 |
| 插件化 | 不支持 | 配置文件 enable/disable 按需加载 |
| 拆微服务 | 几乎不可能 | 抽出一个 domain 就是独立服务 |

---

## 3. 本项目领域划分

### 领域全景图

```
                        ┌────────────────┐
                        │   API Gateway  │
                        └───────┬────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│  auth        │       │  platform     │       │  monitor      │
│  认证授权     │       │  平台管理      │       │  监控告警      │
│              │       │              │       │              │
│  Login       │       │  HealthCheck │       │  Prometheus  │
│  Register    │       │  Settings    │       │  Loki        │
│  LDAP        │       │  AuditLog    │       │  AlertRule   │
│  User CRUD   │       │  License     │       │  HealthScore │
└──────┬───────┘       └──────────────┘       └──────────────┘
        │
        │ 所有域都依赖 auth（拿用户/租户信息）
        │
        ▼
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│  cicd         │       │  k8s          │       │  image        │
│  交付流水线    │       │  资源管理      │       │  镜像仓库      │
│              │       │              │       │              │
│  Pipeline   │──────►│  Deployment  │       │  Registry    │
│  Stage       │ 依赖   │  Pod         │       │  ACR         │
│  Release     │       │  Service     │       │  Harbor      │
│  Artifact    │       │  Node        │       │              │
│  Rollout     │       │  Namespace   │       │              │
│  GitOps      │       │  HPA/VPA     │       │              │
└──────────────┘       └──────────────┘       └──────────────┘

┌──────────────┐
│  appstore     │  应用商城（独立，几乎不依赖其他域）
│  QuickOnboard│
└──────────────┘
```

### 依赖关系

```
        ┌─────────┐
        │  auth   │ ← 最底层，所有域依赖它
        └────┬────┘
             │
   ┌─────────┼─────────┐
   │         │         │
   ▼         ▼         ▼
┌──────┐ ┌──────┐ ┌──────────┐
│ cicd │ │ plat │ │ monitor  │
└──┬───┘ └──────┘ └──────────┘
   │
   │ 依赖
   ▼
┌──────┐
│ k8s  │ ← CICD 部署时需要操作 K8s 资源
└──────┘

┌───────┐  ┌──────────┐
│ image │  │ appstore │ ← 独立，不依赖其他域
└───────┘  └──────────┘

核心原则：依赖方向永远单向。下层域不知道上层域的存在。
```

### 各域核心职责

#### Auth 域（认证授权）

```
职责：用户登录/注册/LDAP认证、JWT签发、Token刷新、用户CRUD
数据表：user、ldap_config
接口暴露：
  Login(username, password) → token
  ValidateToken(token) → userInfo
  GetUser(id) → User
```

#### K8s 域（资源管理）

```
职责：操作 K8s 集群资源（Pod/Deploy/Service/Node/Namespace/PV...）
数据表：kube_cluster（集群连接信息）
外部依赖：K8s API（通过 ClusterClientFactory 连接目标集群）
接口暴露：
  UpdateDeploymentImage(clusterID, ns, name, image) → error
  ScaleDeployment(clusterID, ns, name, replicas) → error
  ListPods(clusterID, ns) → []PodInfo
  GetNodeMetrics(clusterID) → []NodeMetrics
```

#### CICD 域（交付流水线）

```
职责：Pipeline/Stage/Release 全生命周期管理、构建触发、部署策略（滚动/金丝雀）
数据表：cicd_pipeline、cicd_stage、cicd_release、cicd_artifact、cicd_environment...
外部依赖：Jenkins/GitLab API、Git 仓库
依赖其他域：K8s 域（更新部署镜像）、Audit（记录审计日志）
接口暴露：
  CreatePipeline(req) → Pipeline
  TriggerBuild(pipelineID) → BuildResult
  CreateRelease(pipelineID, version) → Release
  GetReleaseStatus(releaseID) → Status
```

#### Monitor 域（可观测性）

```
职责：Prometheus 指标查询、Loki 日志查询、告警规则 CRUD、告警事件管理
数据表：monitor_datasource、monitor_alert_rule、monitor_alert_event...
外部依赖：Prometheus API、Loki API
接口暴露：
  GetClusterOverview() → Overview
  GetNodeMetrics() → []NodeMetric
  QueryRange(query, start, end) → []DataPoint
```

#### Platform 域（平台管理）

```
职责：集群健康检查、平台设置、租户管理、审计日志
数据表：platform_settings、audit_log、tenant...
接口暴露：
  CheckClusterHealth(clusterID) → HealthStatus
  GetSettings() → Settings
  LogAudit(entry) → error
```

#### Image 域（镜像仓库）

```
职责：镜像仓库注册、镜像列表/搜索/同步
数据表：image_registry
接口暴露：
  ListImages(registryID, keyword) → []Image
  SyncImage(registryID, imageName) → error
```

#### AppStore 域（应用商城）

```
职责：应用模板管理、一键部署（QuickOnboard）
数据表：app_template、app_instance
依赖其他域：K8s 域（创建资源）、Helm（渲染模板）
接口暴露：
  ListTemplates() → []Template
  InstallApp(templateID, clusterID, values) → AppInstance
```

---

## 4. 领域内部架构

### 不是 MVC，是「清洁分层」

MVC 是为服务端渲染 HTML 设计的（Controller 处理后交给 View 渲染页面）。本项目是 REST API 只有 JSON 响应，**没有 V**。

领域内部是 **Handler → Service → Repository** 三层：

```
                     ┌────────────────────────────┐
                     │      一个领域的内部结构       │
                     │                            │
HTTP Request ───────►│  handler/ (HTTP 适配层)     │
                     │  ├─ 参数绑定 & 校验          │
                     │  ├─ 调用 service interface   │
                     │  └─ 返回 JSON               │
                     │         │                  │
                     │         ▼                  │
                     │  service/ (业务逻辑层)       │
                     │  ├─ 业务规则、流程编排         │
                     │  ├─ 调用自己的 repository     │
                     │  ├─ 调用其他域的 interface    │
                     │  └─ 返回结果                 │
                     │         │                  │
                     │         ▼                  │
                     │  repository/ (数据持久层)    │
                     │  ├─ 只做 SQL / 缓存 / API   │
                     │  ├─ 不包含业务判断            │
                     │  └─ 只操作本域的表            │
                     └────────────────────────────┘
```

### 每层的职责边界

#### Handler 层 — HTTP 适配

```go
// domain/cicd/handler/release.go
type ReleaseHandler struct {
    svc ReleaseService  // 依赖 interface，不是具体实现
}

func (h *ReleaseHandler) Create(c *gin.Context) {
    // ✅ 允许：参数绑定、参数校验
    var req CreateReleaseRequest
    c.ShouldBindJSON(&req)

    // ✅ 允许：从 context 获取中间件注入的值
    tenantID := c.GetUint32("tenant_id")
    userID := c.GetInt64("user_id")

    // ✅ 允许：调用 Service（只调一个方法）
    result, err := h.svc.CreateRelease(c.Request.Context(), tenantID, userID, req)

    // ✅ 允许：返回 JSON
    c.JSON(200, gin.H{"code": 0, "data": result})

    // ❌ 禁止：任何业务判断
    // ❌ 禁止：直接调 Repository
    // ❌ 禁止：跨域调用
}
```

**Handler 唯一职责：把 HTTP 请求翻译成 Service 调用，再把结果翻译成 HTTP 响应。**

#### Service 层 — 业务核心

```go
// domain/cicd/service/release.go
type ReleaseService struct {
    repo    ReleaseRepository       // 自己的 Repository (struct)
    k8sSvc  k8s.K8sService          // 跨域依赖 (interface)
    auditSvc audit.AuditService     // 跨域依赖 (interface)
}

func (s *ReleaseService) CreateRelease(
    ctx context.Context, tenantID uint32, userID int64, req CreateReleaseReq,
) (*Release, error) {

    // ============ 领域内调用 ============
    // 直接调自己的 repo，不需要接口

    pipeline, err := s.repo.FindPipeline(ctx, tenantID, req.PipelineID)
    if err != nil || pipeline.Status != "ready" {
        return nil, ErrPipelineNotReady
    }

    release := &Release{
        TenantID:   tenantID,
        PipelineID: req.PipelineID,
        Version:    req.Version,
        Status:     "pending",
        CreatedBy:  userID,
    }
    s.repo.CreateRelease(ctx, release)

    // ============ 领域外调用 ============
    // 通过对方暴露的 interface 调用

    err = s.k8sSvc.UpdateDeploymentImage(ctx, k8s.UpdateImageReq{
        ClusterID:  pipeline.ClusterID,
        Namespace:  pipeline.Namespace,
        Deployment: pipeline.WorkloadName,
        Image:      pipeline.Image + ":" + req.Version,
    })
    if err != nil {
        s.repo.UpdateReleaseStatus(ctx, release.ID, "failed")
        return nil, err
    }

    // ============ 领域外调用 ============
    s.auditSvc.Log(ctx, &audit.Entry{
        Action: "cicd.release.created",
        Target: fmt.Sprintf("pipeline:%d", pipeline.ID),
        UserID: userID,
    })

    s.repo.UpdateReleaseStatus(ctx, release.ID, "deployed")
    return release, nil
}
```

**Service 职责：业务规则、流程编排、跨域协调。这是整个系统的核心。**

#### Repository 层 — 数据持久

```go
// domain/cicd/repository/pipeline_repo.go
type PipelineRepo struct {
    db *gorm.DB
}

func (r *PipelineRepo) FindPipeline(
    ctx context.Context, tenantID uint32, id int64,
) (*Pipeline, error) {
    var p Pipeline
    err := r.db.WithContext(ctx).
        Where("id = ? AND tenant_id = ? AND is_del = 0", id, tenantID).
        First(&p).Error
    return &p, err
}

// ❌ 禁止：业务判断（如 if pipeline.Status == "ready"）
// ❌ 禁止：跨域调用
// ❌ 禁止：写日志
// ✅ 允许：纯数据操作（CRUD）
```

---

## 5. 领域之间如何通信

### 核心原则

> **接口定义在使用方，实现在提供方。依赖方向永远单向。**

```
┌─────────────────────────────────┐
│  CICD 域（使用方）               │
│                                 │
│  // 这是 CICD 域定义的接口        │
│  // "我需要的 K8s 能力"           │
│  type K8sClient interface {      │
│    UpdateImage(...) error   ─────┼─── CICD 说：我需要能更新镜像
│    ScaleDeploy(...) error   ─────┼─── CICD 说：我需要能扩缩容
│  }                              │
│                                 │
│  type ReleaseService struct {   │
│    k8sCli K8sClient  // 接口     │
│  }                              │
└─────────────────┬───────────────┘
                  │
                  │ 运行时注入实现
                  │
┌─────────────────▼───────────────┐
│  K8s 域（提供方）               │
│                                 │
│  // K8s 域实现 CICD 定义的接口   │
│  // "我能更新镜像"               │
│  type Service struct {}         │
│  func (s *Service) UpdateImage(...) error { ... }
│  func (s *Service) ScaleDeploy(...) error { ... }
│                                 │
│  // K8s 域不知道 CICD 存在        │
└─────────────────────────────────┘
```

### 为什么接口定义在 CICD 域？

1. CICD 域只需要 2 个 K8s 方法，而不是 K8s 域的全部 50 个方法
2. CICD 域不关心实现细节（用的是哪个 K8s 集群、怎么连的）
3. 未来换实现（改调 K8s API → 调 OpenShift API），CICD 域零改动

### 注入方式：组合到顶层

```go
// 启动期 (initialize/router.go 或 main.go)
func buildDomainServices(db *gorm.DB, factory *ClusterClientFactory) *DomainServices {

    // 先创建不依赖其他域的底层域
    k8sSvc := k8s.NewService(db, factory)
    auditSvc := audit.NewService(db)

    // 再创建上层域，注入依赖
    cicdSvc := cicd.NewService(cicd.Deps{
        Repo:   cicd.NewRepo(db),
        K8s:    k8sSvc,     // ← K8s 域注入
        Audit:  auditSvc,   // ← Audit 域注入
    })

    return &DomainServices{
        Auth:   auth.NewService(db),
        K8s:    k8sSvc,
        CICD:   cicdSvc,
        Audit:  auditSvc,
    }
}
```

### 领域内外对比

| | 领域内 | 领域外 |
|---|---|---|
| 调用方式 | `s.repo.xxx()` | `s.otherDomain.xxx()` |
| 依赖对象 | 本域 Repository (struct) | 其他域 Interface |
| 操作的数据 | 只操作本域的表 | 不碰别人的表，只调接口 |
| 改动影响 | 改了表结构只影响本域 | Interface 不变就不影响调用方 |
| 举例 | `s.repo.FindPipeline()` | `s.k8s.UpdateDeploymentImage()` |

---

## 6. 完整请求流转

以「运维点击"触发发布"按钮」为例，追踪一条请求的完整生命周期：

```
前端 Vue
  │  POST /api/v1/k8s/cicd/release
  │  Authorization: Bearer <jwt>
  │  X-Cluster-ID: 10
  │  Body: { pipeline_id: 42, version: "v2.3.1" }
  ▼

═══════════════════════════════════════════════════════════
Step 1 — Gin Router 路由匹配
═══════════════════════════════════════════════════════════
  │  ReleaseHandler.Create → router: cicd.NewReleaseHandler(svc).Inject(router)
  │
  │  中间件链（按顺序执行）：
  │  AuthJWT → 解密 JWT，注入 user_id / tenant_id / current_user
  │  TenantScope → 切换租户 DB scope
  │  ClusterMiddleware → 验证 X-Cluster-ID，注入 clientset
  ▼

═══════════════════════════════════════════════════════════
Step 2 — Handler 层（HTTP 适配）
═══════════════════════════════════════════════════════════
  文件：domain/cicd/handler/release.go
  │
  │  func (h *ReleaseHandler) Create(c *gin.Context) {
  │      var req CreateReleaseRequest
  │      c.ShouldBindJSON(&req)                    // 参数绑定
  │
  │      tenantID := c.GetUint32("tenant_id")       // 从 context 拿
  │      userID   := c.GetInt64("user_id")
  │
  │      result, err := h.svc.CreateRelease(        // 只调一个方法
  │          c.Request.Context(), tenantID, userID, req)
  │
  │      c.JSON(200, gin.H{"code": 0, "data": result})
  │  }
  ▼

═══════════════════════════════════════════════════════════
Step 3 — Service 层（业务核心）
═══════════════════════════════════════════════════════════
  文件：domain/cicd/service/release.go
  │
  │  func (s *ReleaseService) CreateRelease(...) (*Release, error) {
  │
  │    // --- 领域内：查 Pipeline 是否存在 ---
  │    pipeline, _ := s.repo.FindPipeline(ctx, tenantID, req.PipelineID)
  │    if pipeline.Status != "ready" { return nil, ErrNotReady }
  │         ↓
  │    SELECT * FROM cicd_pipeline WHERE id = 42 AND tenant_id = 100
  │
  │    // --- 领域内：创建 Release 记录 ---
  │    release := &Release{...}
  │    s.repo.CreateRelease(ctx, release)
  │         ↓
  │    INSERT INTO cicd_release (...) VALUES (...)
  │
  │    // --- 跨域调用：更新 Deployment 镜像 ---
  │    s.k8sSvc.UpdateDeploymentImage(ctx, k8s.UpdateImageReq{
  │        ClusterID:  req.ClusterID,
  │        Namespace:  pipeline.Namespace,
  │        Deployment: pipeline.WorkloadName,
  │        Image:      "harbor/app:v2.3.1",
  │    })
  │         ↓
  │    → 进入 K8s 域的 Service 层
  │    → K8s 域查 kube_cluster 表拿 kubeconfig
  │    → K8s 域调 K8s API: PATCH /apis/apps/v1/deployments/xxx
  │    → 返回结果
  │
  │    // --- 跨域调用：写审计日志 ---
  │    s.auditSvc.Log(ctx, &audit.Entry{Fired: "cicd.release.created", ...})
  │         ↓
  │    → 进入 Audit 域的 Service 层
  │    → INSERT INTO audit_log (...) VALUES (...)
  │
  │    // --- 领域内：更新 Release 状态 ---
  │    s.repo.UpdateReleaseStatus(ctx, release.ID, "deployed")
  │         ↓
  │    UPDATE cicd_release SET status = 'deployed' WHERE id = 999
  │
  │    return release, nil
  │  }
  ▼

═══════════════════════════════════════════════════════════
Step 4 — Handler 层 — JSON 响应
═══════════════════════════════════════════════════════════
  │
  │  c.JSON(200, gin.H{"code": 0, "data": {
  │      "release_id": 999,
  │      "status": "deployed",
  │      "version": "v2.3.1"
  │  }})
  ▼

前端收到响应 → 更新 UI
```

### 调用链路总结

```
HTTP 请求
  │
  ├─ Middleware 链  → 认证 / 租户隔离 / 集群校验
  │
  ▼
Handler (HTTP 适配)
  │  职责: 参数绑定 → 调 Service → JSON 响应
  │  不包含: 任何业务判断
  ▼
Service (领域核心)
  │  职责: 业务规则 + 编排流程 + 跨域协调
  │  ├─ s.repo.xxx()       ← 领域内，直接调
  │  ├─ s.k8sSvc.xxx()     ← 领域外，通过 interface
  │  └─ s.auditSvc.xxx()   ← 领域外，通过 interface
  ▼
Repository (数据持久)
  │  职责: 纯 CRUD，不包含业务判断
  ▼
DB / K8s API / Redis / Jenkins ...
```

---

## 7. 对比总结

| 维度 | 当前架构 | 领域架构 |
|---|---|---|
| 组织方式 | 按技术分层（controllers/services/dao） | 按业务垂直切（k8s/cicd/monitor） |
| Service 层 | 一个大 struct，200+ 方法混在一起 | 每个域独立 struct，5-10 个方法 |
| 跨域调用 | `s.dao.xxx()` — 和域内调用无区别 | `s.otherDomain.xxx()` — 显式 interface |
| Controller 依赖 | `*services.Services`（全部方法） | 本域 interface（5-10 方法） |
| 单元测试 | Mock 整个 Services/DAO | Mock 一个 5 方法的 interface |
| 插件化 | 不支持 | `plugins.yaml` 按需启用 |
| 拆微服务 | 几乎不可能 | 抽出一个 domain 即可 |

---

## 8. 迁移路径

### Phase 1 — 修基础（1-2 天）

修复当前代码中的架构问题，为领域化铺路：

- 清理所有 `global.DB` 直接调用 → 统一走 `s.dao`
- K8s Controller 不直调 `pkg/k8s/` → 统一走 Service 层
- Model 中的 `global.DB`/`global.Logger` → 改成参数传入

### Phase 2 — 内部领域化（2-3 天）

保持单体二进制，内部按领域拆分：

1. 每个域建独立目录 `internal/domain/{name}/`
2. 定义跨域 interface（定义在使用方域内部）
3. Services 拆成独立 domain service，通过组合注入
4. Controller 改为只依赖本域 interface

产出：同一份代码，通过配置控制启用哪些域。

### Phase 3 — 拆服务（按需）

只有当某个域需要独立扩缩容/独立部署时才做：

1. 目标域抽出为独立进程
2. 跨域 interface 改为 gRPC client
3. 引入服务发现和 API 网关

---

## 9. 企业级演进

### 插件化部署

```yaml
# configs/plugins.yaml
plugins:
  auth:       enabled: true     # 认证 - 必须
  k8s:        enabled: true     # K8s 资源管理 - 必须
  cicd:       enabled: true     # 流水线 - 按需
  monitoring: enabled: true     # 监控 - 按需
  appstore:   enabled: false     # 应用商城 - 客户不需要
  image:      enabled: false     # 镜像仓库 - 客户有自己的 Harbor
```

同一套代码，不同客户不同组合，插拔式安装。

### 部署形态演进

```
阶段1 — 单体插拔             阶段2 — 独立进程           阶段3 — 微服务网格
┌──────────────────┐       ┌─────────────────┐       ┌─────┐ ┌─────┐ ┌─────┐
│ k8s-platform     │       │ k8s-platform    │       │Auth │ │Monit│ │Image│
│ ┌──────┐┌──────┐│       └─────────────────┘       └─────┘ └─────┘ └─────┘
│ │ K8s  ││CICD  ││       ┌─────────────────┐       ┌─────┐ ┌─────┐ ┌─────┐
│ ├──────┤├──────┤│       │ cicd-service    │       │ K8s │ │CICD │ │AppSt│
│ │Monit ││Image ││       │ (扩缩 x5)       │       │(x3) │ │(x5) │ │(x1) │
│ └──────┘└──────┘│       └─────────────────┘       └─────┘ └─────┘ └─────┘
└──────────────────┘
```

三个阶段共用同一套领域代码，只是 `main.go` 和部署方式不同。
