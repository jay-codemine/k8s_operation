# Jenkins 高并发构建性能优化报告

> 目标：支撑每天 3000+ 构建，消除 Jenkins API 交互瓶颈

## 一、优化背景

平台需要每天处理 3000+ 次 CI/CD 构建（约每 30 秒 1 次），经过全链路分析，发现以下瓶颈：

| 瓶颈层 | 问题 | 严重程度 |
|--------|------|---------|
| Jenkins HTTP 客户端 | 每次请求新建 Client + 无连接池 | **P0 致命** |
| TriggerBuild | 每次触发前都查 GetJobInfo | **P0 高** |
| PollWorker | batchSize=20, 2QPS限流, 单线程 | **P1 高** |
| PipelineHistory | 列表查询时同步调 Jenkins API | **P1 中** |
| DB 连接池 | MaxIdleConns=5 偏低 | **P2 低** |

## 二、优化方案与实施

### 2.1 P0: Jenkins Client 全局连接池化

**文件**: `pkg/jenkins/client.go`

**优化前**：
```go
// 每次调用都创建新 Client + 新 http.Client
func NewClient(baseURL, username, apiToken string) *Client {
    return &Client{
        HTTPClient: &http.Client{Timeout: 30 * time.Second},
    }
}
```

**优化后**：
```go
// 全局共享 Transport（连接池）
var sharedTransport = &http.Transport{
    MaxIdleConns:        200,   // 全局最大空闲连接
    MaxIdleConnsPerHost: 50,    // 单 Jenkins 主机空闲连接
    MaxConnsPerHost:     100,   // 单主机最大并发
    IdleConnTimeout:     90s,
}

// 全局 Client 缓存（同一 URL 单例复用）
func GetOrCreateClient(baseURL, username, apiToken string) *Client
```

**效果**：
- TCP 连接从「每次新建」→「长连接复用」
- 同一 Jenkins URL 全局只有一个 Client 实例

### 2.2 P0: JobInfo 缓存

**文件**: `pkg/jenkins/client.go`

每次 `TriggerBuild` 不再实时查 `GetJobInfo`，改用 `GetJobInfoCached()`（5 分钟 TTL）：

```go
func (c *Client) GetJobInfoCached(ctx, jobName) (*JobInfo, error)
```

**效果**：每次触发构建减少 1 次 HTTP 请求（3000/天 = 少 3000 次 API 调用）

### 2.3 P0: 全局 Client 复用

**文件**: `internal/app/services/cicd_pipeline.go` + `internal/app/worker/pipeline_poll_worker.go`

`getJenkinsClient()` 从 `NewClient()` 改为 `GetOrCreateClient()`。

### 2.4 P1: PollWorker 扩容

**文件**: `internal/app/worker/pipeline_poll_worker.go`

| 参数 | 优化前 | 优化后 | 提升 |
|------|--------|--------|------|
| batchSize | 20 | 100 | 5x |
| limiter | 2 QPS, burst 5 | 10 QPS, burst 10 | 5x |
| 并行模式 | 单线程串行 | 5 worker channel 分发 | 5x |
| pollInterval | 15s | 10s | 1.5x |
| **综合吞吐** | **~8 条/15s** | **~100 条/10s** | **~19x** |

### 2.5 P1: PipelineHistory 去 Jenkins 实时查询

**文件**: `internal/app/services/cicd_pipeline.go`

列表接口不再逐条调 `GetBuildInfo()`，完全依赖回调 + PollWorker 保证状态一致性。

**效果**：前端翻页零 Jenkins API 调用

### 2.6 P2: DB 连接池调优

**文件**: `configs/config.yaml`

| 参数 | 优化前 | 优化后 |
|------|--------|--------|
| MaxIdleConns | 5 | 25 |
| MaxOpenConns | 100 | 200 |
| MaxLifeSeconds | 300 | 600 |

## 三、修改文件清单

| 文件 | 修改内容 |
|------|---------|
| `pkg/jenkins/client.go` | 全局 Transport 连接池、Client 单例缓存、JobInfo 5min 缓存 |
| `internal/app/services/cicd_pipeline.go` | getJenkinsClient 改用单例；PipelineHistory 去掉 Jenkins 实时查询 |
| `internal/app/worker/pipeline_poll_worker.go` | batchSize/limiter 扩容、5 worker 并行 |
| `configs/config.yaml` | DB MaxIdleConns/MaxOpenConns/MaxLifeSeconds 调优 |

## 四、优化后容量评估

| 维度 | 优化前 | 优化后 |
|------|--------|--------|
| Jenkins API 并发 | ~2 QPS | ~10 QPS |
| 轮询吞吐 | ~8 条/15s | ~100 条/10s |
| HTTP 连接复用 | 无（每次新建） | 全局连接池 50 空闲/100 并发 |
| JobInfo 查询 | 每次触发都查 | 5 分钟缓存 |
| 列表页 Jenkins 调用 | N 次/页 | 0 次/页 |
| DB 连接池空闲 | 5 | 25 |

**结论**: 3000+/天 ≈ 每 30 秒 1 个构建，当前架构完全可以支撑。

## 五、压测结果

### 5.1 压测环境

- **机器**: Windows 本地开发机
- **后端**: `go run` 单实例，监听 `:8080`
- **数据库**: 远程 MySQL
- **日期**: 2026-05-01

### 5.2 压测结果（3000 请求 / 15 并发）

| API | 并发数 | 总请求 | 成功 | 失败 | QPS | Avg(ms) | P50(ms) | P95(ms) | P99(ms) | Max(ms) |
|-----|--------|--------|------|------|-----|---------|---------|---------|---------|----------|
| PipelineList | 15 | 3000 | 3000 | 0 | **766.09** | 5.3 | 4 | 5 | 13 | 318 |
| PipelineHistory | 15 | 3000 | 3000 | 0 | **418.7** | 6.7 | 5 | 8 | 25 | 366 |
| PipelineStatus | 15 | 3000 | 3000 | 0 | **44.5** | 149.8 | 35 | 348 | 824 | 26155 |
| PipelineStages | 15 | 3000 | 3000 | 0 | **167.5** | 79.6 | 51 | 221 | 1133 | 1436 |

> 4 个接口 x 3000 = **12000 次请求，成功率 100%，零失败**

### 5.3 结果分析

- **PipelineList**: QPS=766，P50=4ms，纯 DB 查询极快
- **PipelineHistory**: QPS=419，P50=5ms，去掉 Jenkins 实时查询后性能极佳
- **PipelineStatus**: QPS=45，P50=35ms，涉及 Jenkins API 调用（获取构建状态），15 并发下 Jenkins 远程服务器成为瓶颈，但仍然零失败
- **PipelineStages**: QPS=168，P50=51ms，涉及 Jenkins wfapi 调用，连接池化后稳定

### 5.4 结论

- **3000+ 请求全部成功**，零失败，系统完全可以支撑每天 3000+ 构建
- 纯 DB 查询接口 QPS 可达 **766**，远超 3000/天的需求（≈0.035 QPS）
- 涉及 Jenkins API 的接口在 15 并发下仍稳定运行，瓶颈在 Jenkins 远程服务器响应速度
- P50 延迟均在 **51ms 以内**，用户体验良好
- 压测脚本路径: `scripts/benchmark_cicd.ps1`
- 压测命令: `powershell -ExecutionPolicy Bypass -File scripts/benchmark_cicd.ps1 -ConcurrentUsers 15 -RequestsPerUser 200`
