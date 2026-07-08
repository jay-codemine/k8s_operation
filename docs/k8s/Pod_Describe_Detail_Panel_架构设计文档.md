# Pod Describe Detail Panel (DDP) 架构设计文档

> 功能定位：可视化实现 `kubectl describe pod` 的 Web 端等效能力，以右侧抽屉面板形式展示 Pod 完整详情及实时事件。

---

## 一、功能概述

DDP（Describe Detail Panel）是平台工作负载管理模块的核心交互组件，为用户提供类似 `kubectl describe pod <name>` 的可视化体验。点击 Pod 行上的「详情」按钮后，从右侧滑出 880px 宽度的抽屉面板，展示 Pod 的完整运行状态。

### 核心能力

| 能力 | 说明 |
|------|------|
| 基本信息 | Name、Namespace、UID、创建时间、Node、Pod IP、Host IP、Phase、QoS Class、Restart Policy、Service Account |
| Conditions | PodScheduled、Initialized、ContainersReady、Ready 等 |
| Containers | 镜像、命令、参数、端口、资源请求/限制、就绪状态、重启次数、环境变量 |
| Init Containers | 初始化容器详情 |
| Node Selector | 节点选择器标签 |
| Tolerations | 污点容忍规则 |
| Labels | Pod 标签键值对 |
| Annotations | Pod 注解（可折叠） |
| Volumes | 挂载卷列表（可折叠） |
| Events | 最近 50 条相关事件（含 Warning 高亮）|

---

## 二、整体架构

```
┌──────────────────────────────────────────────────────────────────┐
│                         Frontend (Vue 3)                          │
│                                                                  │
│  ┌────────────┐    ┌────────────┐    ┌─────────────────────┐    │
│  │  Pods.vue  │───▶│  pods.js   │───▶│  HTTP Client (axios) │    │
│  │  (DDP组件) │    │  (API层)   │    │                     │    │
│  └────────────┘    └────────────┘    └─────────────────────┘    │
│                                              │                   │
└──────────────────────────────────────────────│───────────────────┘
                                               │ HTTP
┌──────────────────────────────────────────────│───────────────────┐
│                         Backend (Go/Gin)     ▼                   │
│                                                                  │
│  ┌──────────┐    ┌──────────────┐    ┌───────────────────┐      │
│  │  Router  │───▶│  Controller  │───▶│     Service       │      │
│  │ pod.go   │    │   pod.go     │    │   k8s_pod.go      │      │
│  └──────────┘    └──────────────┘    └───────────────────┘      │
│                                              │                   │
│                                              ▼                   │
│                                   ┌───────────────────┐          │
│                                   │    pkg/k8s/pod    │          │
│                                   │   detail.go       │          │
│                                   └───────────────────┘          │
│                                              │                   │
│                                              ▼                   │
│                                   ┌───────────────────┐          │
│                                   │ K8s API Server    │          │
│                                   │ (client-go)       │          │
│                                   └───────────────────┘          │
└──────────────────────────────────────────────────────────────────┘
```

---

## 三、后端架构设计

### 3.1 分层结构

```
Router层 → Controller层 → Service层 → pkg层 → K8s API
```

#### 路由注册

**文件**: `internal/app/routers/kube_pod/pod.go`

```go
router.GET("/detail", pod.Detail)   // Pod 详情
```

事件查询复用 Deployment Events 路由：
```
POST /api/v1/k8s/deployment/events  (body: kind=Pod)
```

#### Controller 层

**文件**: `internal/app/controllers/api/v1/pod/pod.go`

```go
func (c *PodController) Detail(ctx *gin.Context) {
    param := requests.NewKubePodDetailRequest()
    // 1. 参数校验 (namespace + name)
    // 2. 获取多集群 K8s 客户端
    cli := middlewares.MustGetK8sClients(ctx)
    // 3. 调用 Service 层
    pod, err := svc.KubePodDetail(ctx.Request.Context(), cli, param)
    // 4. 返回完整 Pod 对象 (corev1.Pod)
    resp.Success(pod)
}
```

#### Service 层

**文件**: `internal/app/services/k8s_pod.go`

```go
func (s *Services) KubePodDetail(ctx context.Context, cli *K8sClients, param *requests.KubePodDetailRequest) (*corev1.Pod, error) {
    p, err := pod.GetPodDetail(ctx, cli.Kube, param.Namespace, param.Name)
    return p, nil
}
```

#### pkg 层（K8s 底层调用）

**文件**: `pkg/k8s/pod/detail.go`

```go
func GetPodDetail(ctx context.Context, kube kubernetes.Interface, namespace, name string) (*corev1.Pod, error) {
    pod, err := kube.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
    return pod, nil
}
```

### 3.2 事件查询架构

事件使用独立的通用查询接口，通过 `kind` + `name` 字段选择器过滤。

**文件**: `pkg/k8s/event/`

```
ListEvents()
  ├── 优先 events.k8s.io/v1 (新版API)
  │     └── fieldSelector: regarding.kind=Pod, regarding.name=<name>
  └── 回退 core/v1 Events (旧版API)
        └── fieldSelector: involvedObject.kind=Pod, involvedObject.name=<name>
```

**字段选择器构建** (`builders.go`):
```go
// CoreV1 版本
func BuildFieldSelectorCoreV1(kind, name, typ, reason string) string {
    // involvedObject.kind=Pod
    // involvedObject.name=<pod-name>
    // type=Warning (可选)
}
```

**时间过滤与排序**: 使用 `eventutil.ApplySinceAndSort()` 按 `since_seconds` 过滤最近事件，并按时间倒序排列。

---

## 四、前端架构设计

### 4.1 组件结构

DDP 采用 Vue 3 Composition API + `<Teleport>` 实现全屏抽屉，避免被父级 `overflow:hidden` 截断。

```
<Teleport to="body">
  └── div.ddp-overlay (遮罩层, z-index:9000)
       └── <Transition name="ddp-slide"> (动画)
            └── div.ddp-panel (880px 宽度面板)
                 ├── div.ddp-hd (Header: 深色渐变背景)
                 │    ├── Pod名称 + 命名空间徽章
                 │    ├── 状态指示器 (LED + 文字)
                 │    └── 关闭按钮
                 └── div.ddp-body (Body: 可滚动)
                      ├── section: 基本信息 (2列Grid)
                      ├── section: Conditions (Chip列表)
                      ├── section: Containers (卡片 × N)
                      ├── section: Init Containers
                      ├── section: Node Selector
                      ├── section: Tolerations
                      ├── section: Labels
                      ├── section: Annotations (可折叠)
                      ├── section: Volumes (可折叠)
                      └── section: Events (列表, Warning高亮)
```

### 4.2 数据加载策略

**文件**: `k8s-web/src/views/workloads/Pods.vue`

```javascript
const openDetail = async (pod) => {
  // 1. 重置状态
  showDetailModal.value = true;
  loadingDetail.value = true;
  detailData.value = null;
  ddpEventsData.value = [];

  // 2. 并行加载 detail + events（Promise.all）
  const [detailRes, eventsRes] = await Promise.all([
    podsApi.detail({ namespace: pod.namespace, name: pod.name }),
    podsApi.events({ namespace: pod.namespace, name: pod.name, limit: 50 })
  ]);

  // 3. 数据解包（兼容不同响应格式）
  detailData.value = detailRes?.data || pod.raw;
  ddpEventsData.value = eventsRes?.data?.events || eventsRes?.data?.items || [];
};
```

**设计要点**:
- `Promise.all` 并行请求，减少用户等待时间
- 失败时 fallback 到列表页已有的 `pod.raw` 数据
- Events 支持 `events` 和 `items` 两种后端返回格式

### 4.3 API 客户端层

**文件**: `k8s-web/src/api/cluster/workloads/pods.js`

```javascript
// Pod 详情
detail(params) {
  return http.get(`${K8S_BASE}/pod/detail`, { params })
  // GET /api/v1/k8s/pod/detail?namespace=xxx&name=yyy
}

// Pod 事件（复用通用事件接口）
events(params) {
  return http.post(`${K8S_BASE}/deployment/events`, {
    namespace: params.namespace,
    kind: 'Pod',           // 通过 kind 区分资源类型
    name: params.name,
    limit: params.limit || 50,
    since_seconds: params.since_seconds || 3600,
  })
}
```

### 4.4 状态判定逻辑

Pod 状态不仅依赖 `status.phase`，而是**严格按容器就绪状态判定**：

```javascript
const ddpStatusText = (d) => {
  if (!d?.status) return 'UNKNOWN';
  const phase = d.status.phase;
  const containers = d.status.containerStatuses || [];
  
  // 核心逻辑：所有容器 Ready 才算 RUNNING
  if (containers.length > 0 && containers.every(c => c.ready)) return 'RUNNING';
  if (phase === 'Succeeded') return 'SUCCEEDED';
  if (phase === 'Failed') return 'FAILED';
  if (phase === 'Pending') return 'PENDING';
  return phase?.toUpperCase() || 'UNKNOWN';
};
```

**状态与样式映射**:

| 状态 | CSS Class | LED颜色 | 含义 |
|------|-----------|---------|------|
| RUNNING / SUCCEEDED | `status-ok` | 绿色 (#4ade80) | 所有容器就绪 |
| PENDING | `status-warn` | 黄色 (#facc15) | 等待调度/拉取镜像 |
| FAILED | `status-err` | 红色 (#f87171) | 容器异常退出 |
| UNKNOWN | `status-warn` | 黄色 | 无法获取状态 |

### 4.5 事件时间格式化

```javascript
const ddpEvAge = (ts) => {
  if (!ts) return '';
  const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
  if (diff < 60)    return diff + 's ago';
  if (diff < 3600)  return Math.floor(diff / 60) + 'm ago';
  if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
  return Math.floor(diff / 86400) + 'd ago';
};
```

---

## 五、UI/UX 设计规范

### 5.1 布局参数

| 参数 | 值 | 说明 |
|------|-----|------|
| 面板宽度 | 880px (max 92vw) | 响应式适配 |
| 遮罩层 | rgba(15,23,42,.45) + blur(2px) | 深色半透明 |
| Header 背景 | linear-gradient(135deg, #0f172a, #1e3a5f) | 深蓝渐变 |
| Header 内边距 | 18px 28px | |
| Body 内边距 | 24px 28px | |
| z-index | 9000 | 高于其他模态框 |

### 5.2 动画配置

```css
/* 进场: 从右侧滑入 */
.ddp-slide-enter-active {
  transition: transform .32s cubic-bezier(.4, 0, .2, 1);
}
/* 出场: 向右滑出 */
.ddp-slide-leave-active {
  transition: transform .22s cubic-bezier(.4, 0, 1, 1);
}
/* 初始/终态: 完全在右侧屏幕外 */
.ddp-slide-enter-from, .ddp-slide-leave-to {
  transform: translateX(100%);
}
```

### 5.3 信息展示组件

| 组件 | 用途 | 样式 |
|------|------|------|
| `ddp-kv-grid` | 键值对网格 | 2列 grid, 6px/18px 间距 |
| `ddp-chip` | 标签/条件展示 | 圆角12px, 蓝底 (#e0e7ff) |
| `ddp-chip.anno` | 注解标签 | 黄底 (#fef3c7) |
| `ddp-container-card` | 容器详情卡片 | 白色背景, 1px边框, 8px圆角 |
| `ddp-events-list` | 事件列表 | 4列grid, max-height 280px |

---

## 六、API 接口规范

### 6.1 获取 Pod 详情

```
GET /api/v1/k8s/pod/detail
```

**Request Query**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| namespace | string | ✅ | 命名空间 |
| name | string | ✅ | Pod 名称 |

**Response** (code=0):
```json
{
  "code": 0,
  "data": {
    "metadata": {
      "name": "my-pod",
      "namespace": "default",
      "uid": "xxx",
      "creationTimestamp": "2024-01-01T00:00:00Z",
      "labels": { "app": "demo" },
      "annotations": { ... }
    },
    "spec": {
      "nodeName": "node-1",
      "containers": [...],
      "initContainers": [...],
      "volumes": [...],
      "nodeSelector": { ... },
      "tolerations": [...],
      "restartPolicy": "Always",
      "serviceAccountName": "default"
    },
    "status": {
      "phase": "Running",
      "podIP": "10.0.0.1",
      "hostIP": "192.168.1.1",
      "qosClass": "Burstable",
      "conditions": [...],
      "containerStatuses": [...]
    }
  }
}
```

### 6.2 获取 Pod 事件

```
POST /api/v1/k8s/deployment/events
```

**Request Body**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| namespace | string | ✅ | 命名空间 |
| kind | string | ✅ | 资源类型，固定为 `"Pod"` |
| name | string | ✅ | Pod 名称 |
| type | string | ❌ | 事件类型过滤 (Normal/Warning) |
| limit | int | ❌ | 返回条数 (默认50) |
| since_seconds | int | ❌ | 最近N秒 (默认3600) |

**Response** (code=0):
```json
{
  "code": 0,
  "data": {
    "events": [
      {
        "namespace": "default",
        "kind": "Pod",
        "name": "my-pod",
        "type": "Normal",
        "reason": "Scheduled",
        "message": "Successfully assigned...",
        "count": 1,
        "event_time": "2024-01-01T00:00:00Z",
        "source_component": "default-scheduler"
      }
    ],
    "next": ""
  }
}
```

---

## 七、关键文件索引

| 层级 | 文件路径 | 职责 |
|------|---------|------|
| 前端页面 | `k8s-web/src/views/workloads/Pods.vue` | DDP 模板 + 逻辑 + 样式 |
| API 客户端 | `k8s-web/src/api/cluster/workloads/pods.js` | HTTP 请求封装 |
| 后端路由 | `internal/app/routers/kube_pod/pod.go` | 路由注册 |
| 后端控制器 | `internal/app/controllers/api/v1/pod/pod.go` | 请求处理 |
| 后端服务 | `internal/app/services/k8s_pod.go` | 业务编排 |
| K8s 底层 | `pkg/k8s/pod/detail.go` | client-go 调用 |
| 事件查询 | `pkg/k8s/event/ListEvents.go` | 事件统一查询入口 |
| 事件构建 | `pkg/k8s/event/builders.go` | FieldSelector + EventItem 转换 |
| 事件过滤 | `pkg/eventutil/event_util.go` | 时间过滤 + 排序 |

---

## 八、数据流时序图

```
用户点击「详情」
    │
    ▼
openDetail(pod)
    │
    ├── 重置状态 (loadingDetail=true, detailData=null)
    │
    ▼
Promise.all([
    │
    ├── GET /api/v1/k8s/pod/detail ───────────────────────────────┐
    │     │                                                        │
    │     ▼                                                        │
    │   Controller.Detail()                                        │
    │     → Service.KubePodDetail()                                │
    │       → pod.GetPodDetail()                                   │
    │         → kube.CoreV1().Pods(ns).Get(name)                   │
    │           → K8s API Server                                   │
    │                                                              │
    ├── POST /api/v1/k8s/deployment/events (kind=Pod) ────────────┤
    │     │                                                        │
    │     ▼                                                        │
    │   Controller.ListEvents()                                    │
    │     → Service.KubeDeploymentEvents()                         │
    │       → event.ListEvents()                                   │
    │         → BuildFieldSelectorCoreV1(kind=Pod, name=xxx)       │
    │           → kube.CoreV1().Events(ns).List(opts)              │
    │             → K8s API Server                                 │
])                                                                 │
    │                                                              │
    ▼                                                              │
数据解包 & 渲染 ◀──────────────────────────────────────────────────┘
    │
    ├── detailData = response.data (完整 corev1.Pod)
    ├── ddpEventsData = response.data.events
    │
    ▼
ddpStatusText(detailData) → 计算状态
ddpStatusClass() → 映射样式
DDP 面板渲染完成
```

---

## 九、设计决策说明

### 9.1 为什么用 Teleport 而非内联抽屉？

- **避免层叠上下文问题**: 父级若有 `overflow:hidden` 或 `transform`，内联抽屉会被截断
- **z-index 管理简单**: 直接挂载到 `body`，不受父级 z-index 影响
- **遮罩层全屏覆盖**: 点击遮罩可关闭，用户体验一致

### 9.2 为什么 Events 复用 Deployment 路由？

- 后端事件查询通过 `kind` 字段选择器统一过滤
- 一套 API 支持 Pod / Deployment / StatefulSet / DaemonSet 等所有资源类型
- 减少路由和 Controller 代码重复

### 9.3 为什么状态判定不仅看 Phase？

```
Phase=Running 但容器未就绪 → 实际服务不可用
```

- Kubernetes 的 `status.phase` 只反映 Pod 生命周期阶段
- 容器可能处于 `CrashLoopBackOff` 或 `ImagePullBackOff` 但 Phase 仍为 Running
- **严格按 containerStatuses.every(c => c.ready) 判定**确保显示真实可用状态

### 9.4 为什么使用 Promise.all 并行加载？

- Detail 和 Events 是独立的 API 调用，无数据依赖
- 并行请求将总耗时从 T1+T2 降为 max(T1,T2)
- 失败隔离：单个接口失败不影响另一个的数据展示

---

## 十、扩展性设计

该 DDP 模式已复用到平台其他工作负载页面：

| 资源类型 | 文件 | 状态判定差异 |
|---------|------|-------------|
| Pod | `Pods.vue` | containerStatuses.every(c => c.ready) |
| Deployment | `Deployments.vue` | readyReplicas === replicas |
| StatefulSet | `Statefulsets.vue` | readyReplicas === replicas |
| DaemonSet | `Daemonsets.vue` | numberReady === desiredNumberScheduled |
| Job | `Jobs.vue` | succeeded >= completions |
| CronJob | `Cronjobs.vue` | lastScheduleTime 是否正常 |
| Service | `Services.vue` | type + clusterIP 存在性 |
| Ingress | `Ingress.vue` | rules 配置完整性 |

每个资源页面的 DDP 组件遵循相同的 UI 结构，但 `ddpStatusText()` 根据资源类型有不同的判定逻辑。
