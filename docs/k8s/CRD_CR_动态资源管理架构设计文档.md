# CRD/CR 动态资源管理平台架构设计

> **Design Philosophy**: 基于 Kubernetes DynamicClient 构建通用型自定义资源管理引擎，支持任意 CRD 的全生命周期管理，对标 Rancher Explorer / Lens / KubeSphere CRD Manager。

---

## 一、系统定位与核心价值

### 1.1 业务背景

随着云原生生态发展，Prometheus Operator、Istio、Cert-Manager、ArgoCD 等均通过 CRD 扩展 Kubernetes API。运维团队需要一个**统一管控面**来管理分散在各集群中的数百种 CRD 资源。

### 1.2 核心能力矩阵

| 能力域 | 关键特性 | 对标竞品 |
|--------|---------|---------|
| **通用 CRUD** | 基于 DynamicClient 的任意 GVR 资源增删改查 | Rancher Explorer |
| **Schema 感知** | OpenAPI v3 Schema 自动解析，表单化编辑 | Lens Extension |
| **监控类 CRD** | PrometheusRule / ServiceMonitor / PodMonitor 可视化 | Grafana Operator UI |
| **安全防护** | DryRun 校验 + RBAC + 删除保护 + 审计日志 | KubeSphere |
| **变更管理** | YAML Diff 对比 + 版本追踪 + 回滚 | ArgoCD Diff View |

### 1.3 设计原则

```
┌─────────────────────────────────────────────────────────┐
│                    Design Principles                      │
├─────────────────────────────────────────────────────────┤
│  1. Schema-Driven  — CRD Schema 驱动 UI 自动渲染        │
│  2. Zero-Config    — 零配置接入任意新 CRD               │
│  3. Safety-First   — DryRun + 删除保护 + 变更审计       │
│  4. Multi-Cluster  — 多集群 CRD 统一视图                │
│  5. Extensible     — 插件化的监控类 CRD 可视化适配器    │
└─────────────────────────────────────────────────────────┘
```

---

## 二、整体架构

### 2.1 分层架构图

```
┌──────────────────────────────────────────────────────────────────────────┐
│                            Frontend Layer                                  │
│                                                                          │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │ CRD Browser │  │ CR Instance  │  │ Schema Form  │  │ YAML Editor  │ │
│  │   列表/搜索  │  │   CRUD 面板  │  │  动态表单渲染 │  │  Diff/DryRun │ │
│  └─────────────┘  └──────────────┘  └──────────────┘  └──────────────┘ │
│  ┌─────────────────────────────────────────────────────────────────────┐ │
│  │          Monitoring CRD Adapters (PrometheusRule / ServiceMonitor)   │ │
│  └─────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────┬─────────────────────────────────────┘
                                     │ HTTP / WebSocket
┌────────────────────────────────────┼─────────────────────────────────────┐
│                            Gateway Layer                                  │
│                                     │                                    │
│  ┌──────────┐  ┌──────────┐  ┌─────┴────┐  ┌──────────┐  ┌──────────┐ │
│  │   JWT    │  │  RBAC    │  │  Router  │  │  Audit   │  │ Rate     │ │
│  │  Auth    │  │  Guard   │  │          │  │  Logger  │  │ Limiter  │ │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
└────────────────────────────────────┬─────────────────────────────────────┘
                                     │
┌────────────────────────────────────┼─────────────────────────────────────┐
│                          Service Layer                                     │
│                                     │                                    │
│  ┌──────────────────────────────────┴──────────────────────────────────┐ │
│  │                    CRD Resource Engine (核心引擎)                     │ │
│  ├─────────────┬──────────────┬──────────────┬─────────────────────────┤ │
│  │ Discovery   │  Schema      │  Lifecycle   │  Safety Guard           │ │
│  │ Service     │  Parser      │  Manager     │  (DryRun/Protection)    │ │
│  └─────────────┴──────────────┴──────────────┴─────────────────────────┘ │
│                                     │                                    │
│  ┌──────────────────────────────────┴──────────────────────────────────┐ │
│  │             Monitoring CRD Service (监控类 CRD 专用)                  │ │
│  ├──────────────┬────────────────┬─────────────────────────────────────┤ │
│  │ PrometheusRule│ ServiceMonitor │ PodMonitor                         │ │
│  │ Renderer     │ Renderer       │ Renderer                           │ │
│  └──────────────┴────────────────┴─────────────────────────────────────┘ │
└────────────────────────────────────┬─────────────────────────────────────┘
                                     │
┌────────────────────────────────────┼─────────────────────────────────────┐
│                       Infrastructure Layer                                 │
│                                     │                                    │
│  ┌──────────────┐  ┌──────────────┐│  ┌──────────────┐  ┌────────────┐ │
│  │ DynamicClient│  │ Discovery    ││  │ OpenAPI v3   │  │ Informer   │ │
│  │ (client-go)  │  │ Client       ││  │ Schema Cache │  │ (可选)     │ │
│  └──────────────┘  └──────────────┘│  └──────────────┘  └────────────┘ │
│                                     │                                    │
│                          K8s API Server                                   │
└──────────────────────────────────────────────────────────────────────────┘
```

### 2.2 核心数据流

```
用户操作 CR 资源
    │
    ▼
┌─────────────────────────────────────────────┐
│  Step 1: GVR 解析                            │
│  group/version/resource → schema.GroupVersionResource │
└────────────────────────┬────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────┐
│  Step 2: RBAC 鉴权                           │
│  SelfSubjectAccessReview → can I {verb} {resource}? │
└────────────────────────┬────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────┐
│  Step 3: DryRun 预校验（Create/Update）       │
│  DynamicClient.Create(obj, DryRunAll)        │
└────────────────────────┬────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────┐
│  Step 4: 执行操作                            │
│  DynamicClient.Resource(gvr).Namespace(ns)   │
│    .Create / .Get / .Update / .Delete        │
└────────────────────────┬────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────┐
│  Step 5: 审计记录                            │
│  AuditLog → who/what/when/diff              │
└─────────────────────────────────────────────┘
```

---

## 三、后端架构设计

### 3.1 核心包结构

```
pkg/k8s/dynamicresource/
├── client.go          // DynamicClient 封装 + GVR 解析
├── discovery.go       // CRD 发现 + Schema 缓存
├── crud.go            // 通用 CRUD 操作（List/Get/Create/Update/Delete）
├── dryrun.go          // DryRun 预校验引擎
├── diff.go            // YAML Diff 计算
├── protection.go      // 删除保护策略
├── schema_parser.go   // OpenAPI v3 Schema 解析器
└── monitoring/
    ├── prometheus_rule.go     // PrometheusRule 适配器
    ├── service_monitor.go     // ServiceMonitor 适配器
    └── pod_monitor.go         // PodMonitor 适配器
```

### 3.2 K8sClients 扩展

在现有 `K8sClients` 结构中新增 DynamicClient：

```go
// internal/app/services/types.go
type K8sClients struct {
    Config       *rest.Config
    Kube         *kubernetes.Clientset
    Dynamic      dynamic.Interface           // 新增: 通用动态客户端
    Metrics      *metricsclient.Clientset
    SupportsEvV1 bool
}
```

初始化时自动创建：
```go
// internal/app/services/k8s_cluster.go
dynClient, err := dynamic.NewForConfig(cfg)
if err != nil {
    return nil, fmt.Errorf("create dynamic client: %w", err)
}
```

### 3.3 GVR 解析引擎

```go
// pkg/k8s/dynamicresource/client.go
package dynamicresource

import (
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/discovery"
    "k8s.io/client-go/dynamic"
)

// GVRResolver 将 group/version/resource 或 CRD 名称解析为 schema.GroupVersionResource
type GVRResolver struct {
    discovery discovery.DiscoveryInterface
    cache     map[string]schema.GroupVersionResource // LRU 缓存
}

// Resolve 支持多种输入格式：
//   - "monitoring.coreos.com/v1/prometheusrules"  → 完整 GVR
//   - "prometheusrules.monitoring.coreos.com"      → CRD 名称
//   - GVR{Group:"monitoring.coreos.com", Version:"v1", Resource:"prometheusrules"}
func (r *GVRResolver) Resolve(input string) (schema.GroupVersionResource, bool, error) {
    // 1. 尝试从缓存获取
    // 2. 通过 Discovery API 查询 API Resources
    // 3. 匹配并返回 GVR + 是否 Namespaced
}
```

### 3.4 通用 CRUD 引擎

```go
// pkg/k8s/dynamicresource/crud.go
package dynamicresource

import (
    "context"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
)

type DynamicCRUD struct {
    client   dynamic.Interface
    resolver *GVRResolver
}

// List 通用列表查询（支持 labelSelector、fieldSelector、分页）
func (d *DynamicCRUD) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
    if namespace == "" {
        return d.client.Resource(gvr).List(ctx, opts)
    }
    return d.client.Resource(gvr).Namespace(namespace).List(ctx, opts)
}

// Get 获取单个资源
func (d *DynamicCRUD) Get(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
    if namespace == "" {
        return d.client.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
    }
    return d.client.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}

// Create 创建资源（支持 DryRun）
func (d *DynamicCRUD) Create(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured, dryRun bool) (*unstructured.Unstructured, error) {
    opts := metav1.CreateOptions{}
    if dryRun {
        opts.DryRun = []string{metav1.DryRunAll}
    }
    if namespace == "" {
        return d.client.Resource(gvr).Create(ctx, obj, opts)
    }
    return d.client.Resource(gvr).Namespace(namespace).Create(ctx, obj, opts)
}

// Update 更新资源（支持 DryRun + Conflict 检测）
func (d *DynamicCRUD) Update(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured, dryRun bool) (*unstructured.Unstructured, error) {
    opts := metav1.UpdateOptions{}
    if dryRun {
        opts.DryRun = []string{metav1.DryRunAll}
    }
    if namespace == "" {
        return d.client.Resource(gvr).Update(ctx, obj, opts)
    }
    return d.client.Resource(gvr).Namespace(namespace).Update(ctx, obj, opts)
}

// Delete 删除资源（支持删除保护检查）
func (d *DynamicCRUD) Delete(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, opts metav1.DeleteOptions) error {
    if namespace == "" {
        return d.client.Resource(gvr).Delete(ctx, name, opts)
    }
    return d.client.Resource(gvr).Namespace(namespace).Delete(ctx, name, opts)
}
```

### 3.5 DryRun 预校验引擎

```go
// pkg/k8s/dynamicresource/dryrun.go
package dynamicresource

// DryRunResult 预校验结果
type DryRunResult struct {
    Valid      bool              `json:"valid"`
    Errors     []ValidationError `json:"errors,omitempty"`
    Warnings   []string          `json:"warnings,omitempty"`
    ServerDiff string            `json:"server_diff,omitempty"` // 服务端处理后的 diff
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Type    string `json:"type"` // Required / Invalid / Immutable
}

// ValidateWithDryRun 通过 K8s API Server DryRun 执行服务端校验
// 优势：利用 Webhook + Schema Validation 完整链路，比客户端校验更可靠
func (d *DynamicCRUD) ValidateWithDryRun(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured, isUpdate bool) (*DryRunResult, error) {
    result := &DryRunResult{Valid: true}
    
    var err error
    if isUpdate {
        _, err = d.Update(ctx, gvr, namespace, obj, true) // dryRun=true
    } else {
        _, err = d.Create(ctx, gvr, namespace, obj, true)
    }
    
    if err != nil {
        result.Valid = false
        result.Errors = parseAPIErrors(err) // 解析 StatusError → 字段级错误
    }
    
    return result, nil
}
```

### 3.6 删除保护策略

```go
// pkg/k8s/dynamicresource/protection.go
package dynamicresource

// ProtectionPolicy 删除保护策略
type ProtectionPolicy struct {
    // 受保护的资源类型（禁止直接删除）
    ProtectedGVRs map[schema.GroupVersionResource]bool
    
    // 注解保护：带有 protection.k8soperation.io/enabled=true 的资源不可删除
    AnnotationKey   string
    AnnotationValue string
    
    // Finalizer 保护：含有 finalizer 的资源需二次确认
    RequireFinalizerConfirm bool
    
    // 实例数保护：如果 CRD 下有 CR 实例存在，禁止删除 CRD
    PreventCRDDeleteWithInstances bool
}

// CheckDeletionAllowed 检查是否允许删除
func (p *ProtectionPolicy) CheckDeletionAllowed(obj *unstructured.Unstructured) *ProtectionViolation {
    // 1. 检查注解保护
    annotations := obj.GetAnnotations()
    if annotations[p.AnnotationKey] == p.AnnotationValue {
        return &ProtectionViolation{
            Reason: "ANNOTATION_PROTECTED",
            Message: "资源受删除保护注解保护，请先移除保护标记",
        }
    }
    
    // 2. 检查 Finalizer
    if p.RequireFinalizerConfirm && len(obj.GetFinalizers()) > 0 {
        return &ProtectionViolation{
            Reason: "HAS_FINALIZERS",
            Message: fmt.Sprintf("资源含 %d 个 Finalizer，删除可能触发额外操作", len(obj.GetFinalizers())),
            RequireConfirm: true, // 需要二次确认，非完全禁止
        }
    }
    
    return nil // 允许删除
}
```

### 3.7 Schema 解析器

```go
// pkg/k8s/dynamicresource/schema_parser.go
package dynamicresource

import (
    apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// CRDSchema 解析后的 Schema 结构（供前端表单渲染）
type CRDSchema struct {
    Group       string          `json:"group"`
    Version     string          `json:"version"`
    Kind        string          `json:"kind"`
    Scope       string          `json:"scope"`        // Namespaced | Cluster
    Properties  []SchemaField   `json:"properties"`   // 顶层字段
    Required    []string        `json:"required"`
    Description string          `json:"description"`
}

// SchemaField 字段定义
type SchemaField struct {
    Name        string        `json:"name"`
    Type        string        `json:"type"`         // string/integer/boolean/array/object
    Format      string        `json:"format,omitempty"` // date-time/int64/...
    Description string        `json:"description,omitempty"`
    Required    bool          `json:"required"`
    Default     interface{}   `json:"default,omitempty"`
    Enum        []interface{} `json:"enum,omitempty"`
    Minimum     *float64      `json:"minimum,omitempty"`
    Maximum     *float64      `json:"maximum,omitempty"`
    Pattern     string        `json:"pattern,omitempty"`
    Children    []SchemaField `json:"children,omitempty"` // object 类型的子字段
    Items       *SchemaField  `json:"items,omitempty"`    // array 类型的元素 schema
}

// ParseCRDSchema 从 CRD 对象解析可渲染的 Schema
func ParseCRDSchema(crd *apiextv1.CustomResourceDefinition, version string) (*CRDSchema, error) {
    // 1. 找到目标 version 的 schema
    // 2. 递归解析 openAPIV3Schema.Properties
    // 3. 标记 required 字段
    // 4. 提取 enum、default、validation 规则
}
```

### 3.8 YAML Diff 引擎

```go
// pkg/k8s/dynamicresource/diff.go
package dynamicresource

// DiffResult YAML 差异结果
type DiffResult struct {
    HasChanges  bool         `json:"has_changes"`
    Additions   int          `json:"additions"`
    Deletions   int          `json:"deletions"`
    Hunks       []DiffHunk   `json:"hunks"`
    UnifiedDiff string       `json:"unified_diff"` // 标准 unified diff 格式
}

type DiffHunk struct {
    OldStart int      `json:"old_start"`
    NewStart int      `json:"new_start"`
    Lines    []DiffLine `json:"lines"`
}

type DiffLine struct {
    Type    string `json:"type"`    // "add" | "del" | "ctx"
    Content string `json:"content"`
    OldLine int    `json:"old_line,omitempty"`
    NewLine int    `json:"new_line,omitempty"`
}

// ComputeDiff 计算两个 Unstructured 对象的 YAML 差异
// 剥离 managedFields、resourceVersion、generation 等运行时字段
func ComputeDiff(before, after *unstructured.Unstructured) (*DiffResult, error) {
    // 1. 标准化：移除系统字段
    // 2. 序列化为 YAML
    // 3. 逐行 diff（Myers 算法）
    // 4. 构建 unified diff 输出
}
```

### 3.9 路由设计

```go
// internal/app/routers/kube_crd/crd.go
func (r *kubeCRDRouter) Inject(router *gin.RouterGroup) {
    ctrl := v1.NewCRDController()

    // ======== CRD 管理 ========
    crd := router.Group("/crd")
    {
        crd.GET("/list", ctrl.ListCRDs)              // 列出所有 CRD
        crd.GET("/detail", ctrl.GetCRD)              // CRD 详情 + Schema
        crd.GET("/schema", ctrl.GetCRDSchema)        // 获取解析后的 Schema（供表单渲染）
        crd.DELETE("/delete", ctrl.DeleteCRD)         // 删除 CRD（含保护检查）
    }

    // ======== CR (Custom Resource) 实例管理 ========
    cr := router.Group("/cr")
    {
        cr.GET("/list", ctrl.ListCRs)                // 列出 CR 实例
        cr.GET("/detail", ctrl.GetCR)                // 获取单个 CR
        cr.POST("/create", ctrl.CreateCR)            // 创建 CR（含 DryRun）
        cr.PUT("/update", ctrl.UpdateCR)             // 更新 CR（含 DryRun + Diff）
        cr.DELETE("/delete", ctrl.DeleteCR)           // 删除 CR（含保护检查）
        cr.POST("/dryrun", ctrl.DryRunCR)            // 仅 DryRun 不实际执行
        cr.POST("/diff", ctrl.DiffCR)                // 计算变更 Diff
        cr.GET("/yaml", ctrl.GetCRYaml)              // 获取 CR YAML
        cr.PUT("/apply_yaml", ctrl.ApplyCRYaml)      // 应用 YAML 更新
    }

    // ======== 监控类 CRD 专用 ========
    monitoring := router.Group("/monitoring")
    {
        monitoring.GET("/prometheus-rules", ctrl.ListPrometheusRules)
        monitoring.GET("/prometheus-rules/detail", ctrl.GetPrometheusRule)
        monitoring.POST("/prometheus-rules/create", ctrl.CreatePrometheusRule)
        monitoring.PUT("/prometheus-rules/update", ctrl.UpdatePrometheusRule)
        monitoring.DELETE("/prometheus-rules/delete", ctrl.DeletePrometheusRule)

        monitoring.GET("/service-monitors", ctrl.ListServiceMonitors)
        monitoring.GET("/service-monitors/detail", ctrl.GetServiceMonitor)
        monitoring.POST("/service-monitors/create", ctrl.CreateServiceMonitor)
        monitoring.PUT("/service-monitors/update", ctrl.UpdateServiceMonitor)
        monitoring.DELETE("/service-monitors/delete", ctrl.DeleteServiceMonitor)

        monitoring.GET("/pod-monitors", ctrl.ListPodMonitors)
        monitoring.GET("/pod-monitors/detail", ctrl.GetPodMonitor)
        monitoring.POST("/pod-monitors/create", ctrl.CreatePodMonitor)
        monitoring.PUT("/pod-monitors/update", ctrl.UpdatePodMonitor)
        monitoring.DELETE("/pod-monitors/delete", ctrl.DeletePodMonitor)
    }
}
```

### 3.10 RBAC 鉴权设计

```go
// pkg/k8s/dynamicresource/rbac.go
package dynamicresource

import (
    authv1 "k8s.io/api/authorization/v1"
)

// AccessReviewRequest RBAC 校验请求
type AccessReviewRequest struct {
    Group     string `json:"group"`
    Version   string `json:"version"`
    Resource  string `json:"resource"`
    Verb      string `json:"verb"`      // get/list/create/update/delete
    Namespace string `json:"namespace"` // 空=集群级
    Name      string `json:"name,omitempty"`
}

// CheckAccess 通过 SelfSubjectAccessReview 验证当前用户权限
func CheckAccess(ctx context.Context, kube kubernetes.Interface, req *AccessReviewRequest) (bool, string, error) {
    review := &authv1.SelfSubjectAccessReview{
        Spec: authv1.SelfSubjectAccessReviewSpec{
            ResourceAttributes: &authv1.ResourceAttributes{
                Namespace: req.Namespace,
                Verb:      req.Verb,
                Group:     req.Group,
                Version:   req.Version,
                Resource:  req.Resource,
                Name:      req.Name,
            },
        },
    }
    
    result, err := kube.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
    if err != nil {
        return false, "", err
    }
    
    return result.Status.Allowed, result.Status.Reason, nil
}

// BatchCheckAccess 批量检查权限（用于 UI 按钮启用/禁用）
func BatchCheckAccess(ctx context.Context, kube kubernetes.Interface, gvr schema.GroupVersionResource, namespace string) (*ResourcePermissions, error) {
    perms := &ResourcePermissions{}
    verbs := []string{"get", "list", "create", "update", "delete", "watch"}
    
    for _, verb := range verbs {
        allowed, _, _ := CheckAccess(ctx, kube, &AccessReviewRequest{
            Group: gvr.Group, Version: gvr.Version, Resource: gvr.Resource,
            Verb: verb, Namespace: namespace,
        })
        perms.Set(verb, allowed)
    }
    return perms, nil
}
```

---

## 四、前端架构设计

### 4.1 页面结构

```
src/views/extensions/
├── Customresourcedefinitions.vue    // CRD 列表页
├── CustomResourceInstances.vue      // CR 实例列表页（动态路由）
├── components/
│   ├── CRDDetailDrawer.vue          // CRD 详情抽屉（DDP 模式）
│   ├── CRSchemaForm.vue             // Schema 驱动的动态表单
│   ├── CRYamlEditor.vue             // YAML 编辑器 + Diff
│   ├── DryRunPanel.vue              // DryRun 结果展示
│   ├── DeleteProtectionDialog.vue   // 删除保护确认弹窗
│   └── monitoring/
│       ├── PrometheusRuleEditor.vue  // 告警规则可视化编辑
│       ├── ServiceMonitorEditor.vue  // 服务监控配置器
│       └── PodMonitorEditor.vue      // Pod 监控配置器
└── composables/
    ├── useDynamicCR.js               // CR CRUD Composable
    ├── useCRDSchema.js               // Schema 解析 Hook
    └── useYamlDiff.js                // Diff 计算 Hook
```

### 4.2 动态路由设计

```javascript
// router 配置：CR 实例页通过 CRD 名称动态路由
{
  path: '/extensions/crd/:crdName/instances',
  name: 'CRInstances',
  component: () => import('@/views/extensions/CustomResourceInstances.vue'),
  props: true,
  meta: { title: 'CR 实例管理' }
}
```

### 4.3 Schema 驱动的动态表单

```javascript
// composables/useCRDSchema.js
export function useCRDSchema(crdName) {
  const schema = ref(null)
  const loading = ref(false)
  
  const fetchSchema = async () => {
    loading.value = true
    const res = await crdApi.getSchema({ name: crdName })
    schema.value = res.data
    loading.value = false
  }
  
  // 将 Schema 转换为表单配置
  const formConfig = computed(() => {
    if (!schema.value) return []
    return schema.value.properties.map(field => ({
      key: field.name,
      label: field.name,
      type: mapSchemaTypeToInput(field), // string→input, boolean→switch, enum→select
      required: field.required,
      default: field.default,
      rules: buildValidationRules(field),
      children: field.children?.map(/* 递归 */),
    }))
  })
  
  return { schema, formConfig, loading, fetchSchema }
}
```

### 4.4 YAML Diff 组件

```vue
<!-- components/CRYamlEditor.vue -->
<template>
  <div class="yaml-editor-container">
    <!-- 模式切换 -->
    <div class="editor-toolbar">
      <button :class="{ active: mode === 'edit' }" @click="mode = 'edit'">编辑</button>
      <button :class="{ active: mode === 'diff' }" @click="computeDiff">Diff 预览</button>
      <button :class="{ active: mode === 'dryrun' }" @click="runDryRun">DryRun 校验</button>
    </div>
    
    <!-- Diff 视图 -->
    <div v-if="mode === 'diff'" class="diff-view">
      <div class="diff-stats">
        <span class="additions">+{{ diffResult.additions }}</span>
        <span class="deletions">-{{ diffResult.deletions }}</span>
      </div>
      <div class="diff-content">
        <div v-for="hunk in diffResult.hunks" class="diff-hunk">
          <div v-for="line in hunk.lines" :class="['diff-line', line.type]">
            <span class="line-num">{{ line.type === 'del' ? line.old_line : line.new_line }}</span>
            <span class="line-prefix">{{ line.type === 'add' ? '+' : line.type === 'del' ? '-' : ' ' }}</span>
            <span class="line-content">{{ line.content }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- DryRun 结果 -->
    <div v-if="mode === 'dryrun'" class="dryrun-panel">
      <div v-if="dryRunResult.valid" class="dryrun-pass">
        ✅ 服务端校验通过，可安全提交
      </div>
      <div v-else class="dryrun-errors">
        <div v-for="err in dryRunResult.errors" class="dryrun-error-item">
          <span class="field">{{ err.field }}</span>
          <span class="message">{{ err.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
```

### 4.5 监控类 CRD 可视化组件

#### PrometheusRule 可视化编辑器

```vue
<!-- components/monitoring/PrometheusRuleEditor.vue -->
<template>
  <div class="prometheus-rule-editor">
    <!-- 规则组列表 -->
    <div class="rule-groups">
      <div v-for="(group, gi) in rule.spec.groups" :key="gi" class="rule-group-card">
        <div class="group-header">
          <input v-model="group.name" placeholder="组名称" class="group-name-input"/>
          <div class="group-meta">
            <label>评估间隔</label>
            <input v-model="group.interval" placeholder="30s" class="interval-input"/>
          </div>
        </div>
        
        <!-- 告警规则列表 -->
        <div v-for="(r, ri) in group.rules" :key="ri" class="rule-item">
          <div class="rule-header">
            <span class="rule-type" :class="r.alert ? 'alert' : 'record'">
              {{ r.alert ? '🔔 Alert' : '📊 Record' }}
            </span>
            <input v-model="r.alert || r.record" class="rule-name-input"/>
          </div>
          
          <!-- PromQL 编辑器 -->
          <div class="promql-editor">
            <label>PromQL 表达式</label>
            <textarea v-model="r.expr" class="promql-textarea" rows="3"/>
          </div>
          
          <!-- 告警规则额外字段 -->
          <template v-if="r.alert">
            <div class="rule-fields">
              <div class="field">
                <label>持续时间 (for)</label>
                <input v-model="r.for" placeholder="5m"/>
              </div>
              <div class="field">
                <label>严重级别</label>
                <select v-model="r.labels.severity">
                  <option value="critical">Critical</option>
                  <option value="warning">Warning</option>
                  <option value="info">Info</option>
                </select>
              </div>
            </div>
            <div class="annotations-editor">
              <label>Annotations</label>
              <input v-model="r.annotations.summary" placeholder="summary"/>
              <textarea v-model="r.annotations.description" placeholder="description"/>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>
```

#### ServiceMonitor 配置器

```vue
<!-- components/monitoring/ServiceMonitorEditor.vue -->
<template>
  <div class="service-monitor-editor">
    <!-- 基本信息 -->
    <section class="editor-section">
      <h4>目标选择</h4>
      <div class="selector-config">
        <div class="label-selectors">
          <h5>Namespace Selector</h5>
          <div v-for="(ns, i) in monitor.spec.namespaceSelector.matchNames" :key="i">
            <input v-model="monitor.spec.namespaceSelector.matchNames[i]"/>
          </div>
        </div>
        <div class="label-selectors">
          <h5>Service Selector (matchLabels)</h5>
          <div v-for="(v, k) in monitor.spec.selector.matchLabels" :key="k" class="kv-pair">
            <input :value="k" disabled/> = <input v-model="monitor.spec.selector.matchLabels[k]"/>
          </div>
        </div>
      </div>
    </section>
    
    <!-- Endpoints 配置 -->
    <section class="editor-section">
      <h4>Endpoints</h4>
      <div v-for="(ep, i) in monitor.spec.endpoints" :key="i" class="endpoint-card">
        <div class="ep-fields">
          <div class="field"><label>Port</label><input v-model="ep.port"/></div>
          <div class="field"><label>Path</label><input v-model="ep.path" placeholder="/metrics"/></div>
          <div class="field"><label>Interval</label><input v-model="ep.interval" placeholder="30s"/></div>
          <div class="field"><label>Scheme</label>
            <select v-model="ep.scheme"><option>http</option><option>https</option></select>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
```

---

## 五、API 接口规范

### 5.1 CRD 管理接口

#### 列出所有 CRD

```
GET /api/v1/k8s/crd/list
```

**Query Params**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| group | string | ❌ | 按 Group 过滤 |
| scope | string | ❌ | Namespaced / Cluster |
| keyword | string | ❌ | 名称模糊搜索 |
| page | int | ❌ | 页码 |
| limit | int | ❌ | 每页数 |

**Response**:
```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "name": "prometheusrules.monitoring.coreos.com",
        "group": "monitoring.coreos.com",
        "version": "v1",
        "kind": "PrometheusRule",
        "scope": "Namespaced",
        "status": "Established",
        "versions": ["v1", "v1beta1"],
        "instance_count": 12,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 45
  }
}
```

#### 获取 CRD Schema

```
GET /api/v1/k8s/crd/schema?name=prometheusrules.monitoring.coreos.com&version=v1
```

**Response**:
```json
{
  "code": 0,
  "data": {
    "group": "monitoring.coreos.com",
    "version": "v1",
    "kind": "PrometheusRule",
    "scope": "Namespaced",
    "properties": [
      {
        "name": "spec",
        "type": "object",
        "required": true,
        "children": [
          {
            "name": "groups",
            "type": "array",
            "description": "告警规则组列表",
            "items": {
              "type": "object",
              "children": [
                { "name": "name", "type": "string", "required": true },
                { "name": "interval", "type": "string", "format": "duration" },
                { "name": "rules", "type": "array" }
              ]
            }
          }
        ]
      }
    ]
  }
}
```

### 5.2 CR 实例管理接口

#### 创建 CR（含 DryRun）

```
POST /api/v1/k8s/cr/create
```

**Request Body**:
```json
{
  "group": "monitoring.coreos.com",
  "version": "v1",
  "resource": "prometheusrules",
  "namespace": "monitoring",
  "yaml": "apiVersion: monitoring.coreos.com/v1\nkind: PrometheusRule\n...",
  "dry_run": true
}
```

**Response** (dry_run=true):
```json
{
  "code": 0,
  "data": {
    "dry_run_result": {
      "valid": true,
      "warnings": ["spec.groups[0].rules[2].for: 建议设置持续时间"],
      "server_diff": "..."
    }
  }
}
```

#### 计算 Diff

```
POST /api/v1/k8s/cr/diff
```

**Request Body**:
```json
{
  "group": "monitoring.coreos.com",
  "version": "v1",
  "resource": "prometheusrules",
  "namespace": "monitoring",
  "name": "my-alerts",
  "new_yaml": "apiVersion: monitoring.coreos.com/v1\n..."
}
```

**Response**:
```json
{
  "code": 0,
  "data": {
    "has_changes": true,
    "additions": 5,
    "deletions": 2,
    "unified_diff": "--- current\n+++ proposed\n@@ -10,3 +10,6 @@...",
    "hunks": [...]
  }
}
```

### 5.3 RBAC 权限查询

```
GET /api/v1/k8s/cr/permissions?group=monitoring.coreos.com&version=v1&resource=prometheusrules&namespace=monitoring
```

**Response**:
```json
{
  "code": 0,
  "data": {
    "get": true,
    "list": true,
    "create": true,
    "update": true,
    "delete": false,
    "watch": true
  }
}
```

---

## 六、安全设计

### 6.1 多层安全防护体系

```
┌─────────────────────────────────────────────────────────────┐
│                    Security Defense Layers                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Layer 1: 平台 RBAC                                         │
│  ├── JWT Token 验证                                         │
│  ├── 平台角色权限（Admin/Operator/Viewer）                    │
│  └── 集群/命名空间级权限隔离                                   │
│                                                             │
│  Layer 2: K8s RBAC                                          │
│  ├── SelfSubjectAccessReview 实时鉴权                        │
│  ├── 按 GVR + Verb 精确校验                                  │
│  └── 权限不足时 UI 按钮自动禁用                               │
│                                                             │
│  Layer 3: DryRun 预校验                                      │
│  ├── 服务端 Admission Webhook 链路完整执行                    │
│  ├── Schema Validation 字段级校验                            │
│  └── 资源配额 / LimitRange 检查                              │
│                                                             │
│  Layer 4: 删除保护                                           │
│  ├── 注解保护标记 (protection.k8soperation.io/enabled)       │
│  ├── Finalizer 检测与二次确认                                │
│  ├── CRD 删除时检查 CR 实例数                                │
│  └── 关键监控资源禁止批量删除                                 │
│                                                             │
│  Layer 5: 审计追踪                                           │
│  ├── 操作审计日志（who/what/when/where）                     │
│  ├── YAML Diff 存档（变更前后对比）                           │
│  └── 异常操作告警通知                                        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### 6.2 删除保护决策流程

```
用户点击「删除」
    │
    ▼
┌─────────────────────┐
│ 1. 平台 RBAC 校验    │──×──▶ 403 拒绝
└─────────┬───────────┘
          │ ✓
          ▼
┌─────────────────────┐
│ 2. K8s RBAC 校验     │──×──▶ "无删除权限"
└─────────┬───────────┘
          │ ✓
          ▼
┌─────────────────────┐
│ 3. 注解保护检查      │──×──▶ "资源受保护，请先移除标记"
└─────────┬───────────┘
          │ ✓
          ▼
┌─────────────────────┐
│ 4. Finalizer 检测    │──⚠──▶ 二次确认弹窗
└─────────┬───────────┘
          │ ✓ (确认)
          ▼
┌─────────────────────┐
│ 5. CR 实例数检查     │──×──▶ "CRD 下尚有 N 个实例"
│   (仅删除 CRD 时)    │
└─────────┬───────────┘
          │ ✓
          ▼
┌─────────────────────┐
│ 6. 执行删除 + 审计   │
└─────────────────────┘
```

---

## 七、监控类 CRD 可视化设计

### 7.1 PrometheusRule 管理

| 功能 | 实现方式 |
|------|---------|
| 规则组可视化 | 卡片式布局，每组一个 Card |
| PromQL 编辑 | Textarea + 语法高亮提示 |
| 严重级别 | Critical(红)/Warning(黄)/Info(蓝) 色彩编码 |
| 持续时间 | Duration 选择器 (30s / 1m / 5m / 15m) |
| 规则预览 | 实时 YAML 预览 + DryRun 校验 |
| 批量操作 | 启用/禁用单个规则组 |

### 7.2 ServiceMonitor 管理

| 功能 | 实现方式 |
|------|---------|
| 目标选择器 | Label 键值对可视化编辑 |
| Endpoint 配置 | Port/Path/Interval/Scheme 表单化 |
| TLS 配置 | 证书引用 + InsecureSkipVerify 开关 |
| relabelings | 拖拽式规则排序 |
| 关联 Service | 自动检测匹配的 Service 列表 |

### 7.3 PodMonitor 管理

| 功能 | 实现方式 |
|------|---------|
| Pod 选择器 | matchLabels + matchExpressions 编辑 |
| 端口配置 | containerPort 关联 |
| 采集路径 | path + params 配置 |
| Job Label | jobLabel 自动填充 |

### 7.4 监控 CRD GVR 常量

```go
var (
    GVRPrometheusRule = schema.GroupVersionResource{
        Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheusrules",
    }
    GVRServiceMonitor = schema.GroupVersionResource{
        Group: "monitoring.coreos.com", Version: "v1", Resource: "servicemonitors",
    }
    GVRPodMonitor = schema.GroupVersionResource{
        Group: "monitoring.coreos.com", Version: "v1", Resource: "podmonitors",
    }
    GVRAlertmanagerConfig = schema.GroupVersionResource{
        Group: "monitoring.coreos.com", Version: "v1alpha1", Resource: "alertmanagerconfigs",
    }
)
```

---

## 八、关键文件索引

| 层级 | 文件路径 | 职责 |
|------|---------|------|
| **后端核心** | | |
| 路由 | `internal/app/routers/kube_crd/crd.go` | CRD/CR 路由注册 |
| 控制器 | `internal/app/controllers/api/v1/crd/crd.go` | 请求处理 |
| 服务 | `internal/app/services/k8s_crd.go` | 业务编排 |
| 动态引擎 | `pkg/k8s/dynamicresource/client.go` | GVR 解析 |
| CRUD | `pkg/k8s/dynamicresource/crud.go` | 通用增删改查 |
| Schema | `pkg/k8s/dynamicresource/schema_parser.go` | OpenAPI Schema 解析 |
| DryRun | `pkg/k8s/dynamicresource/dryrun.go` | 预校验引擎 |
| Diff | `pkg/k8s/dynamicresource/diff.go` | YAML 差异计算 |
| 保护 | `pkg/k8s/dynamicresource/protection.go` | 删除保护策略 |
| RBAC | `pkg/k8s/dynamicresource/rbac.go` | K8s 权限校验 |
| 监控适配 | `pkg/k8s/dynamicresource/monitoring/*.go` | 监控 CRD 适配器 |
| **前端** | | |
| CRD 列表 | `k8s-web/src/views/extensions/Customresourcedefinitions.vue` | CRD 管理页 |
| CR 实例 | `k8s-web/src/views/extensions/CustomResourceInstances.vue` | CR 实例页 |
| Schema 表单 | `k8s-web/src/views/extensions/components/CRSchemaForm.vue` | 动态表单 |
| YAML 编辑器 | `k8s-web/src/views/extensions/components/CRYamlEditor.vue` | 编辑+Diff |
| 监控编辑器 | `k8s-web/src/views/extensions/components/monitoring/*.vue` | 监控 CRD UI |
| API | `k8s-web/src/api/cluster/extensions/crd.js` | API 客户端 |
| Composable | `k8s-web/src/views/extensions/composables/*.js` | 可复用逻辑 |

---

## 九、性能与可靠性设计

### 9.1 缓存策略

| 缓存层 | 对象 | TTL | 失效条件 |
|--------|------|-----|---------|
| GVR 解析缓存 | CRD → GVR 映射 | 5min | CRD 创建/删除时主动清除 |
| Schema 缓存 | OpenAPI Schema 解析结果 | 10min | CRD 版本变更时失效 |
| RBAC 缓存 | 权限校验结果 | 30s | 用户切换集群/命名空间时清除 |
| 列表分页缓存 | CR 列表 | 不缓存 | 实时查询保证一致性 |

### 9.2 错误处理

```go
// 统一错误码
var (
    ErrCRDNotFound       = errorcode.New(40401, "CRD 不存在")
    ErrCRNotFound        = errorcode.New(40402, "CR 实例不存在")
    ErrGVRResolveFailed  = errorcode.New(40001, "无法解析资源类型")
    ErrSchemaParseError  = errorcode.New(40002, "Schema 解析失败")
    ErrDryRunFailed      = errorcode.New(42201, "DryRun 校验失败")
    ErrDeleteProtected   = errorcode.New(40301, "资源受删除保护")
    ErrRBACDenied        = errorcode.New(40302, "K8s RBAC 权限不足")
    ErrConflict          = errorcode.New(40901, "资源版本冲突，请刷新后重试")
)
```

### 9.3 大列表优化

- **服务端分页**: 使用 `limit` + `continue` token 实现 K8s 原生分页
- **列表裁剪**: 列表接口只返回 `metadata` + `status` 摘要，详情按需加载
- **虚拟滚动**: 前端使用 Virtual List 组件，支持万级 CR 实例浏览

---

## 十、设计决策与 Trade-off

### 10.1 为什么选择 DynamicClient 而非代码生成？

| 维度 | DynamicClient | 代码生成 (client-gen) |
|------|--------------|---------------------|
| 接入新 CRD | 零代码，自动发现 | 需重新生成代码+编译 |
| 类型安全 | Unstructured (弱类型) | 强类型 |
| 性能 | 序列化开销略高 | 直接反序列化 |
| 适用场景 | 通用管控面 | 特定 Operator |

**决策**: 通用管控面优先选择 DynamicClient，对高频监控类 CRD 可额外封装强类型适配器。

### 10.2 为什么使用 Server-Side DryRun 而非客户端校验？

- Server-Side DryRun 会经过完整的 Admission Webhook 链路
- 包含 Validating/Mutating Webhook 的校验逻辑
- 资源配额、LimitRange 等约束只有服务端能检查
- 客户端 Schema 校验无法覆盖自定义 Webhook 规则

### 10.3 为什么监控类 CRD 需要专用适配器？

- PrometheusRule 的 `groups[].rules[].expr` 是 PromQL，需要语法感知
- ServiceMonitor 的 `selector.matchLabels` 需关联实际 Service
- 通用 Schema Form 无法提供领域特定的编辑体验
- 专用适配器 = 通用引擎 + 领域化 UI 增强

### 10.4 为什么 Diff 需要剥离系统字段？

```yaml
# 这些字段由 K8s 自动管理，不应展示在 Diff 中
metadata:
  managedFields: [...]        # 字段管理器（数百行）
  resourceVersion: "12345"    # 乐观锁版本号
  generation: 3               # 资源代数
  uid: "xxx"                  # 唯一标识
  creationTimestamp: "..."    # 创建时间
```

剥离后 Diff 只展示**用户意图变更**，避免噪音。

---

## 十一、演进路线图

```
Phase 1 (MVP)                    Phase 2                      Phase 3
━━━━━━━━━━━━━━━━━━              ━━━━━━━━━━━━━━━━━━          ━━━━━━━━━━━━━━━━━━
✅ CRD 列表/详情/Schema          🔄 PrometheusRule 编辑器     📋 CR 版本历史 + 回滚
✅ CR CRUD (DynamicClient)       🔄 ServiceMonitor 配置器     📋 Informer 实时推送
✅ YAML 编辑器                    🔄 PodMonitor 配置器        📋 跨集群 CRD 同步
✅ DryRun 校验                    🔄 Schema Form 渲染         📋 CRD 模板市场
✅ 基础删除保护                    🔄 YAML Diff 对比           📋 GitOps 集成
✅ K8s RBAC 鉴权                  🔄 审计日志 + Diff 存档     📋 Webhook 管理
```

---

## 十二、与现有架构的集成点

| 集成模块 | 集成方式 | 说明 |
|---------|---------|------|
| 多集群管理 | `K8sClients.Dynamic` | 复用现有集群工厂获取 DynamicClient |
| RBAC 权限 | `middlewares.MustGetK8sClients` | 复用集群中间件注入客户端 |
| 审计日志 | `middlewares.Audit` | 复用审计中间件记录操作 |
| 错误码体系 | `internal/errorcode/` | 扩展 CRD 相关错误码 |
| 前端路由 | `k8s-web/src/router/` | extensions 分组下注册新页面 |
| 多资源 YAML | `pkg/k8s/common/multi_yaml.go` | 复用 YAML 解析能力 |
