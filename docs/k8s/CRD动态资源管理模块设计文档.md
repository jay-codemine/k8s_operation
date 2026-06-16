# CRD 动态资源管理模块设计文档

## 一、概述

### 什么是 CRD？

**CRD（CustomResourceDefinition）** 是 Kubernetes 提供的扩展机制，允许用户在不修改 Kubernetes 源码的情况下，自定义新的资源类型。

- **CRD** = 资源类型的"模板定义"（类比数据库的 Table Schema）
- **CR（Custom Resource）** = 基于 CRD 创建的具体实例（类比数据库中的一行数据）

例如 Prometheus-Stack 安装后会注册以下 CRD：

| CRD 名称 | Kind | 用途 |
|-----------|------|------|
| `prometheusrules.monitoring.coreos.com` | PrometheusRule | 告警规则定义 |
| `alertmanagers.monitoring.coreos.com` | Alertmanager | AlertManager 实例配置 |
| `servicemonitors.monitoring.coreos.com` | ServiceMonitor | 服务监控采集目标 |

### 核心能力

本模块基于 Kubernetes **DynamicClient** 实现对任意 CRD/CR 的通用 CRUD 操作：

```
┌────────────────────────────────────────────────────────┐
│                  前端 CRD 管理页面                        │
│  CRD列表 → 点击进入 → CR实例列表 → 增删改查/YAML编辑     │
└──────────────────────────┬─────────────────────────────┘
                           │ HTTP API
┌──────────────────────────▼─────────────────────────────┐
│  Controller (dynamiccrd)                                │
│  路由: /api/v1/k8s/crd/* 和 /api/v1/k8s/cr/*           │
└──────────────────────────┬─────────────────────────────┘
                           │
┌──────────────────────────▼─────────────────────────────┐
│  Service (k8s_crd.go)                                   │
│  业务逻辑：GVR 组装、删除保护、DryRun 代理               │
└──────────────────────────┬─────────────────────────────┘
                           │
┌──────────────────────────▼─────────────────────────────┐
│  pkg/k8s/dynamicresource (DynamicCRUD 引擎)            │
│  底层：DynamicClient + Unstructured 对象操作            │
└──────────────────────────┬─────────────────────────────┘
                           │
                    K8s API Server
```

---

## 二、Scope（作用域）概念

CRD 有两种作用域，决定了 CR 实例的查询方式：

| 作用域 | 说明 | 查询方式 | 示例 |
|--------|------|----------|------|
| **Namespaced** | CR 归属于某个命名空间 | 可按 namespace 过滤，也可查所有 | PrometheusRule, ServiceMonitor |
| **Cluster** | CR 为集群全局资源 | 无需指定 namespace | BGPConfiguration, ClusterRole |

### 查询逻辑

```go
// pkg/k8s/dynamicresource/dynamic_crud.go

func (d *DynamicCRUD) ListCRs(ctx context.Context, gvr schema.GroupVersionResource, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
    if namespace == "" {
        // Cluster 级：查所有命名空间或集群级资源
        return d.client.Resource(gvr).List(ctx, opts)
    }
    // Namespaced 级：只查指定命名空间
    return d.client.Resource(gvr).Namespace(namespace).List(ctx, opts)
}
```

**前端行为：**
- 进入 Namespaced 类型的 CRD → 显示命名空间筛选下拉框
- 进入 Cluster 类型的 CRD → 不显示命名空间筛选（直接列出所有）

---

## 三、GVR（Group-Version-Resource）

操作 CR 实例的核心是**三元组 GVR**：

| 字段 | 含义 | 示例 |
|------|------|------|
| `group` | API Group | `monitoring.coreos.com` |
| `version` | API Version | `v1` |
| `resource` | 资源复数名 | `prometheusrules` |

GVR 从 CRD 列表响应中获取：
```json
{
  "name": "prometheusrules.monitoring.coreos.com",
  "group": "monitoring.coreos.com",
  "version": "v1",
  "kind": "PrometheusRule",
  "resource": "prometheusrules",   // ← 复数名
  "scope": "Namespaced"
}
```

---

## 四、API 接口清单

### 基础路径

```
/api/v1/k8s/crd/*    CRD 定义管理
/api/v1/k8s/cr/*     CR 实例管理
```

所有接口需要：
- `Authorization: Bearer <token>` — JWT 认证
- `X-Cluster-ID: <cluster_id>` — 目标集群标识

---

### 4.1 CRD 管理

#### 列出所有 CRD

```
GET /api/v1/k8s/crd/list
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `keyword` | string | 否 | 按名称/Kind 模糊搜索 |
| `group` | string | 否 | 按 API Group 精确过滤 |

**响应示例：**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "list": [
      {
        "name": "prometheusrules.monitoring.coreos.com",
        "group": "monitoring.coreos.com",
        "version": "v1",
        "kind": "PrometheusRule",
        "resource": "prometheusrules",
        "scope": "Namespaced",
        "versions": ["v1"],
        "status": "Established",
        "description": "",
        "created_at": "2026-05-20 23:02:41"
      }
    ],
    "total": 41
  }
}
```

#### 获取 CRD 详情

```
GET /api/v1/k8s/crd/detail?name=prometheusrules.monitoring.coreos.com
```

#### 删除 CRD

```
DELETE /api/v1/k8s/crd/delete?name=prometheusrules.monitoring.coreos.com
```

> ⚠️ 删除 CRD 会连带删除该类型下所有 CR 实例

---

### 4.2 CR 实例管理

#### 列出 CR 实例

```
GET /api/v1/k8s/cr/list
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `group` | string | 否 | API Group（如 `monitoring.coreos.com`） |
| `version` | string | **是** | API Version（如 `v1`） |
| `resource` | string | **是** | 资源复数名（如 `prometheusrules`） |
| `namespace` | string | 否 | 命名空间（为空查所有） |
| `label_selector` | string | 否 | 标签选择器（如 `app=myapp`） |

**响应示例：**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "list": [
      {
        "name": "kube-prometheus-stack-alertmanager.rules",
        "namespace": "monitoring",
        "created_at": "2026-05-20 23:02:45",
        "labels": {
          "app": "kube-prometheus-stack",
          "release": "kube-prometheus-stack"
        },
        "annotations": {},
        "uid": "abc-123-def",
        "resource_version": "12345"
      }
    ],
    "total": 5
  }
}
```

#### 获取 CR 详情

```
GET /api/v1/k8s/cr/detail?group=monitoring.coreos.com&version=v1&resource=prometheusrules&namespace=monitoring&name=my-rule
```

#### 获取 CR YAML

```
GET /api/v1/k8s/cr/yaml?group=monitoring.coreos.com&version=v1&resource=prometheusrules&namespace=monitoring&name=my-rule
```

**响应：**
```json
{
  "code": 0,
  "msg": "OK",
  "data": {
    "yaml": "apiVersion: monitoring.coreos.com/v1\nkind: PrometheusRule\n..."
  }
}
```

> 返回的 YAML 已剥离系统字段（managedFields、resourceVersion、uid 等），方便编辑后直接提交

#### 创建 CR 实例

```
POST /api/v1/k8s/cr/create
Content-Type: application/json
```

```json
{
  "group": "monitoring.coreos.com",
  "version": "v1",
  "resource": "prometheusrules",
  "namespace": "monitoring",
  "yaml": "apiVersion: monitoring.coreos.com/v1\nkind: PrometheusRule\nmetadata:\n  name: my-rule\n  namespace: monitoring\nspec:\n  groups:\n    - name: example\n      rules:\n        - alert: HighMemoryUsage\n          expr: node_memory_MemAvailable_bytes < 1073741824\n          for: 5m",
  "dry_run": false
}
```

| 字段 | 说明 |
|------|------|
| `dry_run` | 设为 `true` 时仅校验不实际创建 |

#### 更新 CR 实例

```
PUT /api/v1/k8s/cr/update
Content-Type: application/json
```

```json
{
  "group": "monitoring.coreos.com",
  "version": "v1",
  "resource": "prometheusrules",
  "namespace": "monitoring",
  "name": "my-rule",
  "yaml": "...(完整 YAML)...",
  "dry_run": false
}
```

> 后端自动处理乐观锁：先获取最新 resourceVersion 再更新

#### 删除 CR 实例

```
DELETE /api/v1/k8s/cr/delete?group=monitoring.coreos.com&version=v1&resource=prometheusrules&namespace=monitoring&name=my-rule
```

#### DryRun 校验

```
POST /api/v1/k8s/cr/dry-run
Content-Type: application/json
```

```json
{
  "group": "monitoring.coreos.com",
  "version": "v1",
  "resource": "prometheusrules",
  "namespace": "monitoring",
  "name": "my-rule",
  "yaml": "...",
  "is_update": false
}
```

**响应：**
```json
{
  "code": 0,
  "data": {
    "valid": true,
    "message": "DryRun 校验通过，可安全提交"
  }
}
```

---

## 五、错误码

| 错误码 | 说明 |
|--------|------|
| 207001 | 获取 CRD 列表失败 |
| 207002 | 获取 CRD 详情失败 |
| 207003 | 删除 CRD 失败 |
| 207004 | CRD 受删除保护，无法删除 |
| 207101 | 获取 CR 实例列表失败 |
| 207102 | 获取 CR 实例详情失败 |
| 207103 | 创建 CR 实例失败 |
| 207104 | 更新 CR 实例失败 |
| 207105 | 删除 CR 实例失败 |
| 207106 | CR 实例受删除保护，无法删除 |
| 207201 | DryRun 校验失败 |
| 207202 | YAML 解析失败 |
| 207203 | 资源类型参数无效 (GVR) |

---

## 六、核心设计特性

### 6.1 DynamicClient 通用引擎

不同于传统 Typed Client（如 `appsv1.Deployments()`），DynamicClient 使用 GVR 定位任意资源：

```go
// 操作任意 CRD 的 CR 实例，无需生成 Go 类型代码
client.Resource(gvr).Namespace(ns).List(ctx, opts)
client.Resource(gvr).Namespace(ns).Create(ctx, obj, createOpts)
```

优势：
- 无需为每种 CRD 编写 Typed Client 代码
- 一套代码支持所有 CRD（Prometheus、Calico、Istio 等）
- 运行时发现新 CRD 无需重启服务

### 6.2 删除保护

通过 Annotation 实现，防止误删关键资源：

```yaml
metadata:
  annotations:
    k8soperation.io/delete-protection: enabled
```

当检测到该标记时，平台拒绝删除并返回 `207004/207106` 错误码。

### 6.3 DryRun 预校验

创建/更新前可通过 DryRun 模式进行服务端校验：

```go
opts := metav1.CreateOptions{
    DryRun: []string{metav1.DryRunAll},
}
```

- K8s API Server 执行完整的 admission control
- 包括 webhook 校验、schema 校验等
- 不会实际写入 etcd
- 前端可在提交前一键校验

### 6.4 YAML 清洗

获取 CR YAML 时自动剥离系统管理字段：

| 剥离字段 | 原因 |
|----------|------|
| `metadata.managedFields` | 字段归属追踪，冗长无用 |
| `metadata.resourceVersion` | 乐观锁版本号，自动处理 |
| `metadata.uid` | 系统生成的唯一ID |
| `metadata.generation` | 代际号 |
| `metadata.creationTimestamp` | 创建时间（只读） |
| `status` | 状态字段由控制器管理 |
| `kubectl.kubernetes.io/last-applied-configuration` | kubectl 的 apply 记录 |

### 6.5 乐观锁更新

更新 CR 时自动处理 `resourceVersion` 冲突：

```go
// 1. 先获取最新版本
existing, err := crud.GetCR(ctx, gvr, namespace, name)
// 2. 将用户提交的 YAML 注入最新 resourceVersion
obj.SetResourceVersion(existing.GetResourceVersion())
// 3. 执行更新
crud.UpdateCR(ctx, gvr, namespace, obj, false)
```

---

## 七、前端交互流程

```
┌─────────────────────────────────────────────────────────┐
│  CRD 列表页                                              │
│  ┌──────────────┬──────────┬────────┬────────┬────────┐ │
│  │ 名称         │ Group    │ Kind   │ 版本   │ 操作   │ │
│  ├──────────────┼──────────┼────────┼────────┼────────┤ │
│  │ prometheusru │ monitor  │ Promet │ v1     │[CR实例]│ │
│  │ les.monitor  │ ing.core │ heusRu │        │[YAML]  │ │
│  │ ing.coreos   │ os.com   │ le     │        │[删除]  │ │
│  └──────────────┴──────────┴────────┴────────┴────────┘ │
│         点击 [CR 实例] 按钮                               │
└──────────────────────────┬──────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────┐
│  CR 实例列表页                                           │
│  面包屑: CRD 管理 / PrometheusRule 实例                   │
│                                                          │
│  Kind: PrometheusRule  Group: monitoring.coreos.com       │
│  Version: v1           Scope: Namespaced                  │
│                                                          │
│  [命名空间下拉: monitoring ▼]     [+ 创建实例]             │
│  ┌──────────────┬────────────┬──────────┬──────────┐    │
│  │ 名称         │ 命名空间   │ 创建时间 │ 操作     │    │
│  ├──────────────┼────────────┼──────────┼──────────┤    │
│  │ my-rule-01   │ monitoring │ 05-20    │[YAML]    │    │
│  │              │            │ 23:02    │[编辑]    │    │
│  │              │            │          │[删除]    │    │
│  └──────────────┴────────────┴──────────┴──────────┘    │
└─────────────────────────────────────────────────────────┘
```

### 操作流程

1. **查看 CRD 列表** → 统计卡片展示总数/分组/作用域分布
2. **点击 CR 实例** → 进入该 CRD 的实例列表
3. **创建实例** → 打开 YAML 编辑器（含模板）→ DryRun 校验 → 提交
4. **编辑实例** → 拉取当前 YAML（已清洗）→ 编辑 → DryRun → 更新
5. **查看 YAML** → 侧抽屉展示语法高亮 YAML
6. **删除** → 二次确认弹窗 → 执行删除

---

## 八、代码结构

```
k8s_operation/
├── pkg/k8s/dynamicresource/
│   └── dynamic_crud.go          # 核心 CRUD 引擎（DynamicClient 封装）
├── internal/app/services/
│   └── k8s_crd.go               # Service 层（业务逻辑 + 保护检查）
├── internal/app/controllers/api/v1/dynamiccrd/
│   └── dynamic_crd.go           # Controller 层（参数解析 + 响应组装）
├── internal/app/routers/kube_crd/
│   └── dynamic_crd.go           # 路由注册
├── internal/errorcode/
│   └── crd.go                   # 错误码定义（207xxx）
└── k8s-web/src/
    ├── api/cluster/extensions/
    │   └── crd.js               # 前端 API 客户端
    └── views/extensions/
        └── Customresourcedefinitions.vue  # CRD/CR 管理页面
```

---

## 九、典型使用场景

### 场景1：修改 AlertManager 配置

```
CRD列表 → alertmanagers.monitoring.coreos.com → CR实例列表
→ 点击 "main" 实例的 [编辑] → 修改 YAML 中的 replicas/storage
→ DryRun 校验 → 确认提交
```

### 场景2：创建自定义告警规则

```
CRD列表 → prometheusrules.monitoring.coreos.com → [+ 创建实例]
→ 在 YAML 编辑器中定义 alert 规则 → DryRun 校验 → 创建
```

### 场景3：查看 Calico 网络策略

```
CRD列表 → 搜索 "calico" → networkpolicies.crd.projectcalico.org
→ 查看所有网络策略实例 → 按命名空间筛选
```

### 场景4：管理 Operator 资源

```
CRD列表 → installations.operator.tigera.io → 查看/修改 Tigera 安装配置
```

---

## 十、安全与限制

| 特性 | 实现方式 |
|------|----------|
| 鉴权 | JWT Token + RBAC 角色校验 |
| 集群隔离 | `X-Cluster-ID` Header 指定目标集群 |
| 删除保护 | Annotation `k8soperation.io/delete-protection=enabled` |
| DryRun | 利用 K8s Server-side DryRun 机制 |
| 乐观锁 | 自动注入 resourceVersion 防止并发冲突 |
| YAML 清洗 | 剥离系统字段，防止提交时覆盖控制器状态 |
