# K8s 工作负载实时状态监听与容器终端 — 设计文档

> 版本：v1.0 | 更新日期：2026-04-22

---

## 一、功能概述

本次为 K8sOperation 平台新增两大核心能力，覆盖全部 5 类 K8s 工作负载（Deployment / StatefulSet / DaemonSet / Job / CronJob）：

| 功能 | 说明 |
|------|------|
| **实时状态监听（Resource Watcher）** | 参考 KubeSphere 的工作负载状态追踪机制，当用户执行镜像更新/重启等操作后，自动开启快速轮询，实时展示资源状态从 `Updating → Progressing → Running` 的完整变化过程，并拉取关联 Events 显示在右下角浮窗中 |
| **容器终端（Container Terminal）** | 类似 `kubectl exec -it` 的 Web Terminal 体验，基于 WebSocket + SPDY 双协议桥接，支持交互式 Shell、自动 Shell 检测、窗口自适应、心跳保活、全屏模式 |

---

## 二、整体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                          前端 (Vue 3 SPA)                           │
│                                                                     │
│  ┌───────────────────┐    ┌─────────────────────────────────────┐  │
│  │  KubeTerminal.vue │    │  useResourceWatcher.js (composable) │  │
│  │  ┌─────────────┐  │    │  ┌──────────┐  ┌──────────────────┐│  │
│  │  │  xterm.js   │  │    │  │ 状态轮询  │  │ 事件轮询         ││  │
│  │  │  FitAddon   │  │    │  │ getStatus │  │ getEvents        ││  │
│  │  └──────┬──────┘  │    │  └────┬─────┘  └────┬─────────────┘│  │
│  │         │ stdin    │    │       │ HTTP         │ HTTP         │  │
│  │         │ stdout   │    │       ↓              ↓              │  │
│  │  ┌──────┴──────┐  │    │  /detail API    /events API         │  │
│  │  │  WebSocket  │  │    └─────────────────────────────────────┘  │
│  │  │  Client     │  │                                              │
│  │  └──────┬──────┘  │    集成到 6 个工作负载页面：                   │
│  └─────────┼─────────┘    Deployments / StatefulSets / DaemonSets  │
│            │               Jobs / CronJobs / Pods                   │
└────────────┼───────────────────────────────────────────────────────┘
             │ ws://host/api/v1/k8s/pod/terminal?...
             ↓
┌─────────────────────────────────────────────────────────────────────┐
│                        后端 (Go + Gin)                              │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │  terminal.go (Controller)                                    │   │
│  │  - HTTP → WebSocket 协议升级 (gorilla/websocket)             │   │
│  │  - 自动获取第一个容器名                                       │   │
│  │  - 调用 DetectShell 自动检测可用 Shell                        │   │
│  │  - 调用 ExecInPod 启动交互式 exec 流                          │   │
│  └──────────────────────────┬──────────────────────────────────┘   │
│                              │                                      │
│  ┌──────────────────────────┴──────────────────────────────────┐   │
│  │  exec.go (pkg/k8s/pod)                                      │   │
│  │  - WebSocketTerminal：实现 io.Reader/Writer + SizeQueue      │   │
│  │  - 桥接 WebSocket ↔ K8s SPDY remotecommand                  │   │
│  │  - 30s 心跳保活                                               │   │
│  │  - TerminalMessage JSON 协议通信                              │   │
│  └──────────────────────────┬──────────────────────────────────┘   │
│                              │ SPDY                                 │
│                              ↓                                      │
│                     K8s API Server → kubelet → Container            │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 三、核心功能一：容器终端（Container Terminal）

### 3.1 通信协议设计

前后端通过 WebSocket 传输 JSON 格式的 `TerminalMessage`：

```go
type TerminalMessage struct {
    Op   string `json:"op"`   // 操作类型
    Data string `json:"data"` // 消息体
    Rows uint16 `json:"rows"` // 终端行数 (resize 时使用)
    Cols uint16 `json:"cols"` // 终端列数 (resize 时使用)
}
```

**Op 操作类型说明：**

| Op | 方向 | 说明 |
|---|---|---|
| `stdin` | 前端 → 后端 | 用户键盘输入 |
| `stdout` | 后端 → 前端 | 容器标准输出 |
| `stderr` | 后端 → 前端 | 容器标准错误 |
| `resize` | 前端 → 后端 | 终端窗口大小变化，携带 rows/cols |
| `ping` | 双向 | 心跳探测 |
| `pong` | 双向 | 心跳回应 |

### 3.2 后端实现

#### 3.2.1 文件结构

```
pkg/k8s/pod/
  └── exec.go                    # WebSocket ↔ SPDY 桥接核心

internal/app/controllers/api/v1/pod/
  └── terminal.go                # Terminal WebSocket 控制器

internal/app/routers/kube_pod/
  └── pod.go                     # 路由注册: GET /terminal
```

#### 3.2.2 API 接口

```
GET /api/v1/k8s/pod/terminal (WebSocket)
```

**请求参数（Query）：**

| 参数 | 必填 | 说明 |
|------|------|------|
| `namespace` | ✅ | Pod 所在命名空间 |
| `name` | ✅ | Pod 名称 |
| `container` | ❌ | 容器名称，不传则自动取第一个容器 |
| `shell` | ❌ | Shell 类型（bash/sh/zsh），不传则自动检测 |
| `token` | ✅ | JWT 认证 Token |
| `cluster_id` | ❌ | 集群 ID（多集群场景） |

#### 3.2.3 WebSocketTerminal 核心类

`WebSocketTerminal` 是核心桥接器，同时实现了三个接口：

```
io.Reader     ← 从 WebSocket 读取用户输入 → 转发给 K8s exec stdin
io.Writer     ← 接收 K8s exec stdout/stderr → 写入 WebSocket 推送给前端
SizeQueue     ← 接收前端 resize 事件 → 同步给 K8s exec TTY
```

**数据流转过程：**

```
用户键盘
   │
   ↓ onData
浏览器 xterm.js
   │
   ↓ WebSocket send (op: stdin)
后端 WebSocketTerminal.Read()
   │
   ↓ SPDY stream stdin
K8s API Server → kubelet → Container Shell
   │
   ↓ SPDY stream stdout/stderr
后端 WebSocketTerminal.Write()
   │
   ↓ WebSocket send (op: stdout)
浏览器 xterm.js → term.write()
   │
   ↓ 渲染
终端界面显示
```

#### 3.2.4 Shell 自动检测

`DetectShell` 函数按优先级依次尝试 `bash → sh → zsh`：

```go
func DetectShell(ctx, config, kube, namespace, pod, container) string {
    shells := []string{"bash", "sh", "zsh"}
    for _, sh := range shells {
        // 通过 exec "which <shell>" 探测是否可用
        // 成功返回该 shell，失败继续尝试下一个
    }
    return "sh" // fallback
}
```

#### 3.2.5 心跳保活机制

```
后端: 每 30 秒发送 ping
前端: 每 25 秒发送 ping
双方: 收到 ping 回复 pong
```

任意一方检测到连接断开（WebSocket close/error）时自动清理资源。

### 3.3 前端实现

#### 3.3.1 KubeTerminal.vue 组件

**文件路径：** `k8s-web/src/components/KubeTerminal.vue`（605 行）

**Props：**

| Prop | 类型 | 必填 | 说明 |
|------|------|------|------|
| `visible` | Boolean | ✅ | 控制终端显示/隐藏 |
| `namespace` | String | ✅ | Pod 命名空间 |
| `podName` | String | ✅ | Pod 名称 |
| `containerName` | String | ❌ | 容器名称 |
| `shell` | String | ❌ | 指定 Shell 类型 |

**Emits：** `close` - 用户关闭终端时触发

**功能特性：**

| 特性 | 说明 |
|------|------|
| Tokyo Night 主题 | 暗色终端配色，参考 VS Code Tokyo Night 主题 |
| xterm.js 渲染 | 基于 `@xterm/xterm` v5 + `@xterm/addon-fit` 自适应 |
| WebSocket 双向通信 | JSON 协议 TerminalMessage 格式 |
| 自动 Shell 检测 | 后端检测 + 前端输出文本分析 |
| 窗口自适应 | ResizeObserver 监听 + FitAddon 自动适配 |
| 全屏模式 | 一键切换全屏/窗口模式 |
| 心跳保活 | 25 秒间隔 ping/pong |
| 重连机制 | 手动一键重连，清屏并重新建立连接 |
| 拖拽支持 | 标题栏可拖拽移动窗口 |
| 连接状态指示 | 实时显示 connected/connecting/disconnected/error |
| 快捷键提示 | 底部状态栏显示 Ctrl+C / Ctrl+D / exit |
| 终端尺寸显示 | 实时显示 cols×rows（如 120×35） |
| Teleport 弹层 | 使用 Vue Teleport 渲染到 body，避免 z-index 问题 |
| 进出场动画 | CSS Transition 平滑 slide + scale 动画 |

#### 3.3.2 终端使用入口

| 工作负载 | 入口位置 | 说明 |
|----------|----------|------|
| **Pods** | 表格操作列 + 卡片视图 | 直接打开当前 Pod 终端 |
| **Deployments** | Pod 列表弹窗 | 选择关联 Pod 打开终端 |
| **StatefulSets** | Pod 列表弹窗 | 选择关联 Pod 打开终端 |
| **DaemonSets** | Pod 列表弹窗 | 选择关联 Pod 打开终端 |
| **Jobs** | 表格操作列 | 自动查找 Job 关联的运行中 Pod |
| **CronJobs** | Job Pods 弹窗 | 通过 CronJob → Job → Pod 打开终端 |

---

## 四、核心功能二：实时状态监听（Resource Watcher）

### 4.1 设计思路

参考 KubeSphere 的资源状态追踪机制：

1. 用户执行变更操作（如镜像更新、Job 重启、CronJob 手动触发）
2. 系统自动开启快速轮询，追踪资源状态变化
3. 右下角浮窗实时展示：阶段、进度条、Events 列表
4. 状态到达终态（Running/Complete/Failed）后自动停止并回调

### 4.2 前端 Composable 设计

**文件路径：** `k8s-web/src/composables/useResourceWatcher.js`（212 行）

#### 4.2.1 API 接口

```javascript
const {
  // ===== 响应式状态 =====
  watching,      // ref<boolean>  是否在监听中
  watchTarget,   // ref<object>   监听目标 { namespace, name, kind }
  watchPhase,    // ref<string>   当前阶段 Updating/Progressing/Running/Failed/Timeout
  watchProgress, // ref<number>   进度 0~100
  watchEvents,   // ref<array>    事件列表 [{ type, reason, message, time }]
  watchElapsed,  // ref<number>   已用时间（秒）
  
  // ===== 方法 =====
  startWatching, // (target, options) => void  开始监听
  stopWatching,  // () => void                停止监听
  formatElapsed, // (seconds) => string       格式化耗时 "1m 23s"
  phaseColor,    // (phase) => string         获取阶段颜色
  phaseIcon,     // (phase) => string         获取阶段图标
} = useResourceWatcher()
```

#### 4.2.2 startWatching 参数说明

```javascript
startWatching(
  // 监听目标
  { namespace: 'default', name: 'my-deploy', kind: 'Deployment' },
  
  // 配置选项
  {
    // [必须] 获取资源最新状态
    getStatus: async () => {
      return {
        status: 'Running',         // 资源状态字符串
        desiredReplicas: 3,        // 期望副本数
        readyReplicas: 2,          // 就绪副本数
        updatedReplicas: 3,        // 已更新副本数
      }
    },
    
    // [必须] 获取资源关联事件
    getEvents: async () => {
      return [
        { type: 'Normal', reason: 'Scheduled', message: 'Pod assigned', time: '...' },
        { type: 'Warning', reason: 'BackOff', message: 'Restarting', time: '...' },
      ]
    },
    
    // [可选] 状态到达终态时的回调
    onComplete: ({ success, elapsed }) => {
      if (success) Message.success(`已就绪，耗时 ${elapsed}s`)
      refreshList()  // 刷新列表
    },
    
    pollInterval: 2000,    // [可选] 状态轮询间隔，默认 3000ms
    eventInterval: 4000,   // [可选] 事件轮询间隔，默认 5000ms
    timeout: 300000,       // [可选] 超时自动停止，默认 300000ms (5分钟)
  }
)
```

#### 4.2.3 状态阶段判定逻辑

```
                  ┌─────────────┐
                  │   开始监听   │
                  └──────┬──────┘
                         │
                         ↓
                  ┌─────────────┐
                  │  Updating   │  初始阶段
                  │  🔄 黄色    │
                  └──────┬──────┘
                         │
              ┌──────────┼──────────┐
              │          │          │
              ↓          ↓          ↓
     ┌──────────────┐  ┌──────┐  ┌──────────┐
     │ Progressing  │  │Failed│  │ Timeout   │
     │ ⏳ 蓝色      │  │❌ 红 │  │ ⏰ 橙色   │
     │ updated>0    │  │      │  │ >5min     │
     └──────┬───────┘  └──────┘  └───────────┘
            │
            ↓
     ┌─────────────┐
     │   Running   │  ready >= desired && updated >= desired
     │   ✅ 绿色   │
     └─────────────┘
```

**具体判定规则：**

| 条件 | 阶段 | 颜色 |
|------|------|------|
| `status` 为 `running/available/complete` 且 `ready >= desired` | **Running** ✅ | `#9ece6a` 绿色 |
| `status` 为 `failed/crashloopbackoff/error` | **Failed** ❌ | `#f7768e` 红色 |
| `updatedReplicas > 0` 且 `< desired` | **Progressing** ⏳ | `#7aa2f7` 蓝色 |
| `readyReplicas < desired` | **Updating** 🔄 | `#e0af68` 黄色 |
| 超时（默认 5 分钟） | **Timeout** ⏰ | `#ff9e64` 橙色 |

**进度计算公式：**

```
progress = min(100, round(readyReplicas / max(desiredReplicas, 1) * 100))
```

### 4.3 各工作负载的 Watcher 集成

#### 4.3.1 Deployment

```javascript
// 触发时机：镜像更新后
const startDeploymentWatcher = (deploy) => {
  startWatching(target, {
    getStatus: () => deploymentsApi.rolloutStatus({ namespace, name }),
    getEvents: () => deploymentsApi.events({ namespace, name }),
    onComplete: ({ success }) => { /* 刷新列表 */ },
  })
}
```

- **状态 API：** `deploymentsApi.rolloutStatus()` — 专用 rollout 状态接口
- **事件 API：** `deploymentsApi.events()` — 获取 Deployment 关联 Events
- **触发入口：** 内联镜像编辑保存 / 弹窗镜像更新

#### 4.3.2 StatefulSet

```javascript
// 触发时机：镜像更新后
const startStatefulSetWatcher = (sts) => {
  startWatching(target, {
    getStatus: async () => {
      const res = await statefulsetsApi.detail({ namespace, name })
      return {
        status: d.status,
        desiredReplicas: d.replicas,
        readyReplicas: d.ready_replicas,
        updatedReplicas: d.updated_replicas,
      }
    },
    getEvents: () => statefulsetsApi.events({ namespace, name }),
  })
}
```

- **特殊处理：** 无 rolloutStatus API，使用 `detail` 接口获取状态字段
- **字段映射：** `replicas` / `ready_replicas` / `updated_replicas`

#### 4.3.3 DaemonSet

```javascript
// 触发时机：镜像更新后
const startDaemonSetWatcher = (ds) => {
  startWatching(target, {
    getStatus: async () => {
      const res = await daemonsetsApi.detail({ namespace, name })
      return {
        desiredReplicas: d.desired_number_scheduled,
        readyReplicas: d.number_ready,
        updatedReplicas: d.updated_number_scheduled,
      }
    },
    getEvents: () => daemonsetsApi.events({ namespace, name }),
  })
}
```

- **特殊处理：** DaemonSet 的副本字段与 Deployment/StatefulSet 不同
- **字段映射：** `desired_number_scheduled` / `number_ready` / `updated_number_scheduled`

#### 4.3.4 Job

```javascript
// 触发时机：重启 Job 后
const startJobWatcher = (job) => {
  startWatching(target, {
    getStatus: async () => {
      const res = await jobsApi.detail({ namespace, name })
      return {
        status: d.status,
        desiredReplicas: d.completions || 1,
        readyReplicas: d.succeeded || 0,
        updatedReplicas: d.succeeded || 0,
      }
    },
    getEvents: () => jobsApi.events({ namespace, name }),
    timeout: 600000, // Job 可能运行很久，超时延长到 10 分钟
  })
}
```

- **特殊处理：** Job 的 template 不可变（immutable），不支持镜像更新
- **监听目标：** 追踪 `succeeded/completions` 完成进度，而非副本就绪
- **触发入口：** "重新运行"操作后自动开启监听新 Job

#### 4.3.5 CronJob

```javascript
// 触发时机：手动触发 CronJob 后
const startCronJobWatcher = (cj, jobName) => {
  startWatching(target, {
    getStatus: async () => {
      if (jobName) {
        // 监听触发的具体 Job 完成状态
        return { status, desiredReplicas: completions, readyReplicas: succeeded }
      }
      // 监听 CronJob 本身状态
      return { status: 'Active/Suspended', ... }
    },
    getEvents: () => cronjobsApi.events({ namespace, name }),
  })
}
```

- **特殊处理：** 手动触发 CronJob 会创建一个新 Job，监听该 Job 的完成状态
- **触发入口：** "手动触发"操作 → 后端返回 `job_name` → 监听该 Job

### 4.4 Watcher 浮窗 UI

右下角固定定位的暗色浮窗，Tokyo Night 主题风格：

```
┌─────────────────────────────────────┐
│  🔄 Deployment 状态监听          ✕  │  ← header: 图标 + 标题 + 关闭
├─────────────────────────────────────┤
│  Progressing              1m 23s    │  ← 阶段名(带颜色) + 耗时
│  ████████████░░░░░  67%             │  ← 进度条(颜色跟随阶段)
│                                     │
│  Normal  Successfully pulled image  │  ← Events 列表
│  Normal  Created container nginx    │     type 颜色区分
│  Normal  Started container nginx    │     最多显示 4 条最新
│  Warning BackOff: restarting...     │
└─────────────────────────────────────┘
```

**样式规格：**

| 属性 | 值 |
|------|---|
| 位置 | `fixed; right: 24px; bottom: 24px` |
| 宽度 | 370px |
| 背景 | `#1a1b2e`（深海蓝） |
| Header 背景 | `linear-gradient(135deg, #1e1f35, #252845)` |
| 边框 | `1px solid rgba(99, 102, 241, 0.25)`（靛蓝光边） |
| 圆角 | 14px |
| 阴影 | `0 8px 32px rgba(0,0,0,0.45)` |
| 字体色 | `#e2e8f0` |
| z-index | 9000 |
| 动画 | `slide-up` 进出场（translateY + opacity，0.35s） |

---

## 五、涉及文件清单

### 5.1 后端文件

| 文件 | 行数 | 说明 |
|------|------|------|
| `pkg/k8s/pod/exec.go` | 231 | WebSocket ↔ SPDY 桥接核心，含 TerminalMessage 协议、WebSocketTerminal 类、ExecInPod 函数、DetectShell 检测 |
| `internal/app/controllers/api/v1/pod/terminal.go` | 114 | Terminal WebSocket 控制器，处理协议升级、容器自动选择、Shell 检测、日志记录 |
| `internal/app/routers/kube_pod/pod.go` | 41 | 路由注册 `GET /terminal` |

### 5.2 前端公共组件/工具

| 文件 | 行数 | 说明 |
|------|------|------|
| `k8s-web/src/components/KubeTerminal.vue` | 605 | 终端 UI 组件（xterm.js + WebSocket + Tokyo Night 主题） |
| `k8s-web/src/composables/useResourceWatcher.js` | 212 | 状态监听 composable（轮询 + 阶段判定 + Events 收集） |

### 5.3 前端 API 层

| 文件 | 新增方法 | 说明 |
|------|----------|------|
| `k8s-web/src/api/cluster/workloads/pods.js` | `terminalUrl()` | WebSocket URL 构造 |
| `k8s-web/src/api/cluster/workloads/jobs.js` | `events()` | Job Events 接口 |
| `k8s-web/src/api/cluster/workloads/cronjobs.js` | `events()`, `getYaml()`, `applyYaml()` | CronJob 增强接口 |

### 5.4 前端工作负载页面

| 文件 | 新增功能 |
|------|----------|
| `k8s-web/src/views/workloads/Deployments.vue` | KubeTerminal + useResourceWatcher（startDeploymentWatcher） |
| `k8s-web/src/views/workloads/Pods.vue` | KubeTerminal（表格+卡片两个入口） |
| `k8s-web/src/views/workloads/StatefulSets.vue` | KubeTerminal + useResourceWatcher（startStatefulSetWatcher） |
| `k8s-web/src/views/workloads/Daemonsets.vue` | KubeTerminal + useResourceWatcher（startDaemonSetWatcher） |
| `k8s-web/src/views/workloads/Jobs.vue` | KubeTerminal + useResourceWatcher（startJobWatcher） |
| `k8s-web/src/views/workloads/CronJobs.vue` | KubeTerminal + useResourceWatcher（startCronJobWatcher） |

### 5.5 依赖项

| 依赖 | 版本 | 说明 |
|------|------|------|
| `github.com/gorilla/websocket` | v1.5.4-pre | Go WebSocket 库 |
| `k8s.io/client-go/tools/remotecommand` | v0.34.x | K8s SPDY exec |
| `@xterm/xterm` | ^5.x | 前端终端模拟器 |
| `@xterm/addon-fit` | ^0.10.x | xterm 窗口自适应插件 |

---

## 六、安全设计

| 安全项 | 实现方式 |
|--------|----------|
| **认证** | WebSocket URL 携带 JWT Token，后端中间件统一校验 |
| **权限** | 终端入口仅对有操作权限的角色显示（viewer 角色不可见） |
| **跨域** | WebSocket Upgrader 配置 `CheckOrigin` |
| **超时** | Watcher 默认 5 分钟超时自动停止，避免资源泄漏 |
| **资源清理** | 组件 `onBeforeUnmount` 时自动断开 WebSocket、停止轮询 |
| **心跳** | 前后端双向 ping/pong（25s/30s），检测僵尸连接 |

---

## 七、使用场景示例

### 场景 1：Deployment 滚动更新追踪

1. 用户在 Deployments 页面点击镜像列的编辑图标
2. 输入新镜像地址 `nginx:1.25` → 点击确认
3. 系统自动调用 `updateImage` API，成功后触发 `startDeploymentWatcher`
4. 右下角浮窗弹出，显示状态：`Updating 🔄`
5. 随着 Pod 逐步替换：`Progressing ⏳ 33%` → `67%` → `100%`
6. 全部就绪后显示：`Running ✅`，弹出 Toast 提示 "已就绪（耗时 45s）"
7. 2 秒后浮窗自动消失，列表自动刷新

### 场景 2：进入容器终端排查问题

1. 用户在 Pods 页面找到问题 Pod，点击 `>_` 终端按钮
2. KubeTerminal 弹出，状态灯显示 `connecting...`
3. 后端自动检测到容器有 `bash`，建立 SPDY 连接
4. 终端显示绿色连接成功信息：`✓ Connected to default/my-pod [nginx]`
5. 用户输入 `tail -f /var/log/nginx/error.log` 查看实时日志
6. 完成排查后输入 `exit` 或点击关闭按钮断开连接

### 场景 3：CronJob 手动触发并追踪

1. 用户在 CronJobs 页面，通过"更多"菜单选择 "手动触发"
2. 确认后系统调用 `trigger` API，后端创建新 Job
3. 自动开启 Watcher 追踪该 Job 的运行状态
4. 浮窗显示：`Updating 🔄` → `Progressing ⏳` → `Running ✅ Job 已完成（耗时 12s）`
5. 用户可通过 Job Pods 弹窗进入 Pod 终端查看执行详情
