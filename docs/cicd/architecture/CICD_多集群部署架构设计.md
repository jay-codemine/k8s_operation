# CICD 多集群部署架构设计文档

> 本文档详细说明平台 CICD 多集群部署的架构设计，包括技术选型、核心组件、部署流程。

---

## 一、架构选型

### 1.1 设计目标

| 目标 | 要求 |
|------|------|
| 多集群支持 | 一次发布可部署到多个 K8s 集群 |
| 高并发 | 支持大量部署任务并行执行 |
| 水平扩展 | Worker 可横向扩容 |
| 消息可靠 | 任务不丢失、支持重试 |
| 优雅停机 | 不中断正在执行的任务 |

### 1.2 技术选型：Worker + Goroutine 组合模式

采用 **「多 Worker 进程 + 单 Worker 内多 Goroutine」** 组合模式：

| 层级 | 实现方式 | 数量 | 作用 |
|------|----------|------|------|
| **Worker 进程** | 独立进程（可多实例部署） | N 个 | 水平扩展，高可用 |
| **Goroutine** | 单 Worker 内并发协程 | 默认 3 个 | 并发处理任务 |

**选型对比**：

| 方案 | 优点 | 缺点 | 结论 |
|------|------|------|------|
| 纯多 Worker | 隔离性好 | 资源开销大 | ❌ |
| 纯多 Goroutine | 轻量高效 | 单点故障 | ❌ |
| **Worker + Goroutine** | 兼顾扩展性和高效 | - | ✅ 采用 |

---

## 二、架构总览

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       CICD 多集群部署架构                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   ┌──────────────┐                                                          │
│   │   API 层     │  触发发布 → 创建 Release + Tasks → 写入 Redis Stream     │
│   └──────┬───────┘                                                          │
│          │                                                                   │
│          ▼                                                                   │
│   ┌──────────────────────────────────────────────────────────────┐          │
│   │                  Redis Stream (消息队列)                      │          │
│   │                  cicd:deploy:stream                           │          │
│   │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐     │          │
│   │  │ Task 1 │ │ Task 2 │ │ Task 3 │ │ Task 4 │ │ Task 5 │ ... │          │
│   │  │集群A   │ │集群B   │ │集群A   │ │集群C   │ │集群B   │     │          │
│   │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘     │          │
│   └──────────────────────────┬───────────────────────────────────┘          │
│                              │                                               │
│          ┌───────────────────┼───────────────────┐                          │
│          │                   │                   │                          │
│          ▼                   ▼                   ▼                          │
│   ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                   │
│   │  Worker 1   │     │  Worker 2   │     │  Worker N   │ (可水平扩展)      │
│   │ ┌─────────┐ │     │ ┌─────────┐ │     │ ┌─────────┐ │                   │
│   │ │Goroutine│ │     │ │Goroutine│ │     │ │Goroutine│ │                   │
│   │ │  #1~#3  │ │     │ │  #1~#3  │ │     │ │  #1~#3  │ │                   │
│   │ └─────────┘ │     │ └─────────┘ │     │ └─────────┘ │                   │
│   └──────┬──────┘     └──────┬──────┘     └──────┬──────┘                   │
│          └───────────────────┼───────────────────┘                          │
│                              ▼                                               │
│   ┌──────────────────────────────────────────────────────────────┐          │
│   │              ClusterClientFactory (集群客户端工厂)            │          │
│   │  ┌────────────┐  ┌────────────┐  ┌────────────┐             │          │
│   │  │ K8s Client │  │ K8s Client │  │ K8s Client │             │          │
│   │  │  集群 A    │  │  集群 B    │  │  集群 C    │             │          │
│   │  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘             │          │
│   └────────┼───────────────┼───────────────┼────────────────────┘          │
│            ▼               ▼               ▼                                │
│   ┌────────────────┐ ┌────────────────┐ ┌────────────────┐                 │
│   │  K8s 集群 A    │ │  K8s 集群 B    │ │  K8s 集群 C    │                 │
│   └────────────────┘ └────────────────┘ └────────────────┘                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 三、核心组件

### 3.1 CicdWorker - 部署任务消费者

**文件**：`internal/app/worker/cicd_worker.go`

```go
type CicdWorker struct {
    stream       *infra.RedisStream         // Redis Stream 客户端
    dao          *dao.Dao                   // 数据访问层
    executor     *services.CicdTaskExecutor // 部署执行器
    consumerName string                     // 消费者名称
    concurrency  int                        // 并发数（默认 3）
    stopCh       chan struct{}              // 停止信号
    wg           sync.WaitGroup             // 优雅停机
}
```

**启动流程**：
```go
func (w *CicdWorker) Start(ctx context.Context) error {
    // 1. 创建消费者组
    w.stream.CreateGroup(ctx, CicdDeployStream, CicdDeployGroup, "0")
    
    // 2. 处理 Pending 消息（Worker 重启后恢复）
    go w.processPendingMessages(ctx)
    
    // 3. 启动 N 个消费协程
    for i := 0; i < w.concurrency; i++ {
        w.wg.Add(1)
        go w.consumeLoop(ctx, i)
    }
}
```

### 3.2 Redis Stream - 消息队列

**文件**：`internal/app/infra/redis_stream.go`

```go
const (
    CicdDeployStream = "cicd:deploy:stream"   // 队列名
    CicdDeployGroup  = "cicd-deploy-workers"  // 消费者组
)
```

**核心方法**：

| 方法 | 功能 |
|------|------|
| `XAdd` | 写入任务消息 |
| `XReadGroup` | 消费者组读取 |
| `XReadGroupPending` | 读取 Pending（重试） |
| `XAck` | 确认消息已处理 |
| `XClaim` | 认领超时消息 |

### 3.3 ClusterClientFactory - 集群客户端工厂

**文件**：`internal/app/services/k8s_cluster_factory.go`

```go
type ClusterClientFactory struct {
    mu          sync.RWMutex
    m           map[uint32]*cachedClients  // 缓存各集群客户端
    g           singleflight.Group         // 防并发创建
    baseTTL     time.Duration              // 缓存 TTL
    jitterRange time.Duration              // 随机抖动（防雪崩）
}
```

### 3.4 CicdTaskExecutor - 部署执行器

**文件**：`internal/app/services/cicd_executor.go`

```go
func (e *CicdTaskExecutor) Execute(ctx context.Context, task *CicdReleaseTask, release *CicdRelease) *CicdExecuteResult {
    // 1. 根据 ClusterID 获取 K8s 客户端
    cli, _ := e.factory.GetClient(ctx, task.ClusterID)
    
    // 2. 根据工作负载类型执行部署
    switch release.WorkloadKind {
    case "Deployment":
        return e.executeDeployment(ctx, cli.Kube, task, release)
    case "StatefulSet":
        return e.executeStatefulSet(ctx, cli.Kube, task, release)
    case "DaemonSet":
        return e.executeDaemonSet(ctx, cli.Kube, task, release)
    }
}
```

---

## 四、两层并发架构详解

### 4.1 架构图解

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Redis Stream                                        │
│                         cicd:deploy:stream                                       │
│   ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐ ┌─────┐             │
│   │Msg1 │ │Msg2 │ │Msg3 │ │Msg4 │ │Msg5 │ │Msg6 │ │Msg7 │ │Msg8 │  ...        │
│   └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘ └─────┘             │
└───────────────────────────────────┬─────────────────────────────────────────────┘
                                    │
                   Consumer Group: cicd-deploy-workers
                   (自动负载均衡，每条消息只被一个消费者处理)
                                    │
         ┌──────────────────────────┼──────────────────────┐
         │                          │                      │
         ▼                          ▼                      ▼
┌─────────────────┐        ┌─────────────────┐    ┌─────────────────┐
│   Worker 1      │        │   Worker 2      │    │   Worker 3      │
│   (进程 A)      │        │   (进程 B)      │    │   (进程 C)      │
│   K8s Pod 1     │        │   K8s Pod 2     │    │   K8s Pod 3     │
│                 │        │                 │    │                 │
│ consumerName:   │        │ consumerName:   │    │ consumerName:   │
│ pod1-1709xxxx   │        │ pod2-1709xxxx   │    │ pod3-1709xxxx   │
│                 │        │                 │    │                 │
│ ┌─────────────┐ │        │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Goroutine 0 │ │        │ │ Goroutine 0 │ │    │ │ Goroutine 0 │ │
│ │ (消费循环)  │ │        │ │ (消费循环)  │ │    │ │ (消费循环)  │ │
│ ├─────────────┤ │        │ ├─────────────┤ │    │ ├─────────────┤ │
│ │ Goroutine 1 │ │        │ │ Goroutine 1 │ │    │ │ Goroutine 1 │ │
│ │ (消费循环)  │ │        │ │ (消费循环)  │ │    │ │ (消费循环)  │ │
│ ├─────────────┤ │        │ ├─────────────┤ │    │ ├─────────────┤ │
│ │ Goroutine 2 │ │        │ │ Goroutine 2 │ │    │ │ Goroutine 2 │ │
│ │ (消费循环)  │ │        │ │ (消费循环)  │ │    │ │ (消费循环)  │ │
│ └─────────────┘ │        │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘        └─────────────────┘    └─────────────────┘

总并发能力 = 3 Worker × 3 Goroutine = 9 个任务同时执行
```

### 4.2 第一层：Worker 进程层（水平扩展）

**核心原理**：Redis Stream 消费者组

```
                    Redis Stream
                    ┌─────────────────────────────────┐
                    │ Msg1  Msg2  Msg3  Msg4  Msg5    │
                    └──────────────┬──────────────────┘
                                   │
                    Consumer Group: cicd-deploy-workers
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    ▼                    ▼
         Consumer A           Consumer B           Consumer C
         (Worker 1)           (Worker 2)           (Worker 3)
              │                    │                    │
              ▼                    ▼                    ▼
           Msg1, Msg4          Msg2, Msg5          Msg3
```

**关键点**：
- 同一消费者组内，每条消息只会被一个消费者处理（自动负载均衡）
- 多个 Worker 进程属于同一个消费者组
- Redis 自动分配消息，无需手动实现分片

**Worker 创建代码**：
```go
// cicd_worker.go
func NewCicdWorker(rdb *redis.Client, factory *services.ClusterClientFactory) *CicdWorker {
    hostname, _ := os.Hostname()  // 获取主机名（K8s Pod 名）
    
    return &CicdWorker{
        stream:       infra.NewRedisStream(rdb),
        // 消费者名称 = 主机名 + 时间戳（保证唯一）
        consumerName: hostname + "-" + strconv.FormatInt(time.Now().Unix(), 10),
        concurrency:  3,  // 每个 Worker 启动 3 个 Goroutine
        stopCh:       make(chan struct{}),
    }
}
```

### 4.3 第二层：Goroutine 层（单机并发）

**启动多个消费协程**：
```go
// cicd_worker.go
func (w *CicdWorker) Start(ctx context.Context) error {
    // 1. 创建消费者组
    w.stream.CreateGroup(ctx, infra.CicdDeployStream, infra.CicdDeployGroup, "0")
    
    // 2. 处理 Pending 消息（Worker 重启后恢复）
    go w.processPendingMessages(ctx)
    
    // 3. 启动 N 个消费协程（关键！）
    for i := 0; i < w.concurrency; i++ {  // concurrency = 3
        w.wg.Add(1)
        go w.consumeLoop(ctx, i)  // 每个协程独立消费
    }
    
    return nil
}
```

**图解**：
```
                    Worker 进程
                    ┌─────────────────────────────────┐
                    │                                 │
       Start()      │    for i := 0; i < 3; i++ {    │
          │         │        go consumeLoop(i)       │
          ▼         │    }                           │
                    │                                 │
     ┌──────────────┼──────────────┬─────────────────┤
     │              │              │                 │
     ▼              ▼              ▼                 │
 Goroutine 0   Goroutine 1   Goroutine 2            │
 consumeLoop   consumeLoop   consumeLoop            │
     │              │              │                 │
     │    独立消费循环，互不干扰    │                 │
     │              │              │                 │
     ▼              ▼              ▼                 │
   处理 Msg1     处理 Msg2     处理 Msg3             │
                    │                                 │
                    └─────────────────────────────────┘
```

### 4.4 消费循环与消息处理

**消费循环**：
```go
// cicd_worker.go
func (w *CicdWorker) consumeLoop(ctx context.Context, workerID int) {
    defer w.wg.Done()  // 退出时通知 WaitGroup
    
    for {
        select {
        case <-w.stopCh:   // 收到停止信号
            return
        case <-ctx.Done(): // 上下文取消
            return
        default:
            w.consumeOnce(ctx, workerID)  // 消费一条消息
        }
    }
}
```

**消费一条消息**：
```go
// cicd_worker.go
func (w *CicdWorker) consumeOnce(ctx context.Context, workerID int) {
    // XREADGROUP GROUP cicd-deploy-workers consumer-name 
    //     COUNT 1 BLOCK 5000 STREAMS cicd:deploy:stream >
    streams, err := w.stream.XReadGroup(
        ctx,
        infra.CicdDeployStream,  // 队列名
        infra.CicdDeployGroup,   // 消费者组
        w.consumerName,          // 消费者名（唯一）
        1,                       // 一次读取 1 条
        5*time.Second,           // 阻塞等待 5 秒
    )
    
    if err != nil {
        if err == redis.Nil {
            return  // 没有新消息，继续循环
        }
        return
    }
    
    // 处理消息
    for _, stream := range streams {
        for _, msg := range stream.Messages {
            w.processMessage(ctx, msg)
        }
    }
}
```

**关键参数**：
- `COUNT 1`：每次只读取 1 条，确保负载均衡
- `BLOCK 5000`：阻塞等待 5 秒，避免空转消耗 CPU
- `">"`：只读取新消息

### 4.5 消息 ACK 机制

```
                消息生命周期
                
    ┌──────────────────────────────────────┐
    │              Redis Stream             │
    │                                       │
    │   [Msg1]  [Msg2]  [Msg3]  [Msg4]     │
    │      │                                │
    │      │  XREADGROUP (读取)            │
    │      ▼                                │
    │   Pending List (待处理列表)           │
    │   ┌─────────────────────────────┐    │
    │   │ Msg1 → Consumer: worker1    │    │
    │   │        ReadTime: 10:00:00   │    │
    │   └─────────────────────────────┘    │
    │                                       │
    │      │  处理中...                     │
    │      │                                │
    │      │  XACK (确认)                   │
    │      ▼                                │
    │   从 Pending List 移除                │
    │                                       │
    └──────────────────────────────────────┘
```

**ACK 代码**：
```go
// cicd_worker.go
func (w *CicdWorker) processMessage(ctx context.Context, msg redis.XMessage) {
    // 1. 解析消息
    taskID, _ := strconv.ParseInt(msg.Values["task_id"].(string), 10, 64)
    releaseID, _ := strconv.ParseInt(msg.Values["release_id"].(string), 10, 64)
    
    // 2. 执行部署任务
    w.executeTask(ctx, taskID, releaseID)
    
    // 3. ACK 确认消息已处理（重要！）
    w.ackMessage(ctx, msg.ID)
}

func (w *CicdWorker) ackMessage(ctx context.Context, msgID string) {
    // XACK cicd:deploy:stream cicd-deploy-workers <msg_id>
    w.stream.XAck(ctx, infra.CicdDeployStream, infra.CicdDeployGroup, msgID)
}
```

### 4.6 优雅停机

```go
// cicd_worker.go
func (w *CicdWorker) Stop() {
    close(w.stopCh)  // 发送停止信号
    w.wg.Wait()      // 等待所有 Goroutine 退出
}
```

**流程图**：
```
    Stop() 调用
         │
         ▼
    close(stopCh)  ─────────┬─────────────┬─────────────┐
                            │             │             │
                            ▼             ▼             ▼
                      Goroutine 0   Goroutine 1   Goroutine 2
                            │             │             │
                      case <-stopCh:      │             │
                            │             │             │
                            ▼             ▼             ▼
                         return        return        return
                            │             │             │
                            └──────┬──────┴──────┬──────┘
                                   │             │
                                   ▼             ▼
                            wg.Done()      wg.Done()
                                   │
                                   ▼
                            wg.Wait() 返回
                                   │
                                   ▼
                            Worker 完全停止
```

### 4.7 两层并发对比

| 特性 | Worker 层（进程） | Goroutine 层（协程） |
|------|------------------|---------------------|
| 扩展方式 | K8s 扩副本 | 配置 concurrency 参数 |
| 资源隔离 | 完全隔离 | 共享内存 |
| 故障影响 | 单 Worker 故障不影响其他 | 单协程 panic 可能影响整个 Worker |
| 通信方式 | Redis Stream | 共享变量 |
| 适用场景 | 跨机器扩展 | 单机并发 |
| 开销 | 较大（进程级） | 很小（KB 级） |

---

## 五、数据模型

### 4.1 发布单与任务关系

```
┌─────────────────┐
│  CicdRelease    │  发布单（一次发布）
│  - ID           │
│  - WorkloadKind │
│  - Namespace    │
│  - Concurrency  │  ← 并发数
└────────┬────────┘
         │ 1:N
         ▼
┌─────────────────┐
│ CicdReleaseTask │  部署任务（每集群一个）
│  - ID           │
│  - ReleaseID    │
│  - ClusterID    │  ← 目标集群
│  - TargetImage  │
│  - PrevImage    │  ← 原镜像（回滚用）
│  - Status       │
└─────────────────┘
```

### 4.2 相关数据库表

| 表名 | 用途 |
|------|------|
| `cicd_release` | 发布单 |
| `cicd_release_task` | 部署任务（每集群一条） |
| `cicd_release_stage` | 发布阶段 |
| `kube_cluster` | K8s 集群配置 |

---

## 六、部署流程

```
1. 用户发起发布
      │
      ▼
2. 创建 Release + 多个 Task（每个集群一个）
      │
      ▼
3. 每个 Task 写入 Redis Stream（携带 cluster_id）
      │
      ▼
4. Worker Goroutine 竞争消费
      │
      ▼
5. 根据 task.ClusterID 获取对应 K8s 客户端
      │
      ▼
6. Patch 更新镜像 → 等待 Rollout 完成
      │
      ▼
7. 更新 Task 状态 → ACK 消息 → 完结 Release
```

---

## 七、关键代码

### 6.1 任务执行

```go
// cicd_worker.go
func (w *CicdWorker) executeTask(ctx context.Context, taskID, releaseID int64) {
    // 1. 获取任务
    task, _ := w.dao.CicdTaskGetByID(ctx, taskID)
    
    // 2. 获取发布单
    release, _ := w.dao.CicdReleaseGetByID(ctx, releaseID)
    
    // 3. 标记开始执行
    w.dao.CicdTaskMarkStarted(ctx, taskID)
    
    // 4. 执行部署
    result := w.executor.Execute(ctx, task, release)
    
    // 5. 保存原镜像（用于回滚）
    if result.PrevImage != "" {
        w.dao.CicdTaskUpdatePrevImage(ctx, taskID, result.PrevImage)
    }
    
    // 6. 更新状态
    if result.Success {
        w.markTaskSucceeded(ctx, taskID, releaseID, result.Message)
    } else {
        w.markTaskFailed(ctx, taskID, releaseID, result.Message)
    }
}
```

### 6.2 等待 Rollout 完成

```go
// cicd_executor.go
func (e *CicdTaskExecutor) waitDeploymentRollout(ctx context.Context, kube kubernetes.Interface, namespace, name string, timeout time.Duration) error {
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()
    
    timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    for {
        select {
        case <-timeoutCtx.Done():
            return fmt.Errorf("rollout timeout")
        case <-ticker.C:
            dp, _ := kube.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
            
            // 检查是否完成
            if isDeploymentRolloutComplete(dp) {
                return nil
            }
            
            // 检查是否失败
            for _, cond := range dp.Status.Conditions {
                if cond.Reason == "ProgressDeadlineExceeded" {
                    return fmt.Errorf("rollout failed: %s", cond.Message)
                }
            }
        }
    }
}
```

---

## 八、架构优势

| 特性 | 实现 | 价值 |
|------|------|------|
| **水平扩展** | 多 Worker + Redis 消费者组 | 高吞吐量 |
| **高并发** | 单 Worker 多 Goroutine | 充分利用 CPU |
| **多集群** | Task 携带 ClusterID | 一次发布多集群 |
| **消息可靠** | ACK + Pending 重试 | 不丢消息 |
| **优雅停机** | WaitGroup + stopCh | 不中断任务 |
| **回滚支持** | 记录 PrevImage | 一键回滚 |
| **客户端复用** | Factory + 缓存 | 减少连接开销 |
| **防雪崩** | TTL + Jitter | 缓存平滑过期 |

---

## 九、扩展部署

### 8.1 单机部署

```yaml
# 单 Worker，3 个 Goroutine
concurrency: 3
```

### 8.2 多实例部署（K8s）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cicd-worker
spec:
  replicas: 3  # 3 个 Worker 实例
  template:
    spec:
      containers:
      - name: worker
        env:
        - name: WORKER_CONCURRENCY
          value: "3"  # 每个实例 3 个 Goroutine
```

**总并发能力**：3 Worker × 3 Goroutine = **9 个并发任务**

---

## 十、面试要点

**Q: 多集群部署是用多个 Worker 还是 Goroutine？**

**A**: 两者结合：
1. **Worker 层面**：多实例部署，通过 Redis Stream 消费者组负载均衡
2. **Goroutine 层面**：单 Worker 内启动多个协程并发消费
3. **多集群实现**：Task 携带 `cluster_id`，执行时通过 Factory 获取对应集群客户端

这种设计既能**水平扩展**（加 Worker），又能**充分利用单机 CPU**（多 Goroutine），同时通过 Redis Stream 保证消息可靠投递。

---

## 十一、相关文件

| 文件 | 说明 |
|------|------|
| `internal/app/worker/cicd_worker.go` | Worker 主逻辑 |
| `internal/app/infra/redis_stream.go` | Redis Stream 封装 |
| `internal/app/services/k8s_cluster_factory.go` | 集群客户端工厂 |
| `internal/app/services/cicd_executor.go` | 部署执行器 |
| `internal/app/models/cicd_release.go` | 发布单模型 |
| `internal/app/models/cicd_release_task.go` | 任务模型 |
| `internal/bootstrap/bootstrap.go` | Worker 启动入口 |
