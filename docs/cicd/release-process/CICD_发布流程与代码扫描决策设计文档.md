# CI/CD 发布流程与代码扫描决策设计文档

> K8sOperation 平台 · 完整发布生命周期 · 阶段决策树 · 失败处理策略

---

## 一、发布流程全景图

### 1.1 完整发布生命周期

一次完整的 CI/CD 发布经历以下阶段，每个阶段都有明确的**通过/失败决策点**：

```
用户触发构建
    │
    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                        CI 阶段（Jenkins Pipeline 执行）                   │
│                                                                          │
│  ① Clean ──▶ ② Checkout ──▶ ③ Compile ──▶ ④ Test ──▶ ⑤ SonarQube     │
│     │            │              │             │           │               │
│   失败→中止    失败→中止      失败→中止     失败→中止   失败→中止         │
│                                                          │               │
│                                                    ⑥ Quality Gate        │
│                                                     │       │            │
│                                                  通过     未通过→中止     │
│                                                     │                    │
│                                            ⑦ Package ──▶ ⑧ Build Image  │
│                                                │             │           │
│                                              失败→中止     失败→中止     │
│                                                              │           │
│                                                      ⑨ Push Image        │
│                                                        │       │         │
│                                                      成功    失败→中止   │
└────────────────────────────────────────────────────────┼─────────────────┘
                                                         │
                                                    Jenkins 回调
                                                         │
┌────────────────────────────────────────────────────────┼─────────────────┐
│                   CD 阶段（平台处理）                    │                 │
│                                                         ▼                 │
│                                              ┌───── 构建成功？─────┐      │
│                                              │                    │      │
│                                              是                   否     │
│                                              │                    │      │
│                                    ┌── 需要审批？──┐          记录失败    │
│                                    │              │          发送通知    │
│                                    是             否          ▼ 结束     │
│                                    │              │                      │
│                               等待人工审批    ⑩ 自动部署                 │
│                                 │     │         │     │                  │
│                              批准   拒绝     Rollout  失败               │
│                                │      │      成功     │                  │
│                           ⑩ 部署   记录拒绝   │    记录失败              │
│                              │       通知    通知成功  通知失败           │
│                              ▼                                           │
│                          发布完成                                        │
└──────────────────────────────────────────────────────────────────────────┘
```

### 1.2 阶段清单（按语言类型）

| 阶段 | Go | Java | Frontend | Python | 说明 |
|------|:--:|:----:|:--------:|:------:|------|
| Clean Workspace | ✅ | ✅ | ✅ | ✅ | 清理 + 拉取代码 |
| Checkout Info | ✅ | ✅ | ✅ | ✅ | 采集 Git 信息、生成镜像标签 |
| Dependencies | ✅ | - | ✅ | ✅ | 下载依赖 (go mod / npm ci / pip) |
| Compile | ✅ | ✅ | ✅ | - | 编译检查 |
| Test | ✅ | ✅ | ✅ | ✅ | 单元测试（可跳过） |
| Lint | ✅ | - | - | ✅ | 代码风格检查 |
| **SonarQube Analysis** | - | ✅ | - | - | 代码质量扫描（可选） |
| **Quality Gate** | - | ✅ | - | - | 质量门禁（可选） |
| Package | - | ✅ | - | - | Maven 打包 |
| Build Image | ✅ | ✅ | ✅ | ✅ | nerdctl 构建镜像 |
| Push Image | ✅ | ✅ | ✅ | ✅ | 推送到 Harbor |
| **人工审批** | 可选 | 可选 | 可选 | 可选 | 平台阶段（生产环境） |
| **自动部署** | 可选 | 可选 | 可选 | 可选 | 平台阶段（K8s Patch） |

---

## 二、每个阶段的决策逻辑

### 2.1 阶段决策树总览

```
每个 Jenkins 阶段
    │
    ├── 执行成功 → stageCallback(stage, 'success') → 继续下一阶段
    │
    └── 执行失败 → stageCallback(stage, 'failed') → 整个 Pipeline 中止
                                                      │
                                                      ▼
                                              callbackPlatform('FAILURE', msg)
                                                      │
                                                      ▼
                                              平台记录失败 + 发送通知
```

**核心原则：Jenkins Pipeline 中任何阶段失败，后续阶段全部跳过，整条流水线标记为失败。**

### 2.2 各阶段详细决策

#### ① Clean Workspace + 代码拉取

```
检查 GIT_REPO 是否为空？
  └── 空 → error("GIT_REPO 不能为空") → Pipeline 立即终止
  └── 非空 → Git Clone
                └── 成功 → 继续
                └── 失败（网络/权限/分支不存在）→ Pipeline 中止
```

**失败原因**：Git 仓库地址错误、凭证无效、分支不存在、网络不通

#### ② Checkout Info

```
采集 Git 信息：commit hash / branch name / timestamp
生成镜像标签：{branch}-{commit}-{timestamp}
  └── 成功 → stageCallback('checkout', 'success') → 继续
  └── 失败 → stageCallback('checkout', 'failed') → Pipeline 中止
```

**失败原因**：极少失败，仅当 Git 仓库为空或磁盘空间不足时

#### ③ Compile（编译）

| 语言 | 编译命令 | 失败后果 |
|------|---------|---------|
| Go | `go test -run "^$" ./...` | Pipeline 中止 |
| Java | `mvn clean compile -DskipTests -B` | Pipeline 中止 |
| Frontend | `npm run build` | Pipeline 中止 |

**失败原因**：语法错误、缺少依赖、不兼容的版本

#### ④ Test（单元测试）

```
SKIP_TESTS 参数 = true？
  └── 是 → 跳过此阶段（when 条件不满足） → 直接进入下一阶段
  └── 否 → 执行测试
              └── 通过 → stageCallback('test', 'success') → 继续
              └── 失败 → stageCallback('test', 'failed') → Pipeline 中止
```

**特殊处理**：
- Go 模板：无测试文件时自动跳过（`find . -name '*_test.go'`）
- Frontend 模板：lint 和 test 使用 `|| true`，不会因测试失败中止
- Java 模板：使用 `junit` 插件归档测试报告

#### ⑤ SonarQube Analysis（代码扫描）

```
ENABLE_SONAR = true？
  ├── 否 → 跳过扫描 → 直接进入打包阶段
  └── 是 → 执行 mvn sonar:sonar
              └── 成功 → stageCallback('sonar', 'success') → 进入质量门禁
              └── 失败 → stageCallback('sonar', 'failed') → Pipeline 中止
```

**失败原因**：SonarQube Server 不可达、Token 过期、磁盘满、分析超时

#### ⑥ Quality Gate（质量门禁） ← 关键决策点

```
ENABLE_SONAR = true 且 SONAR_QUALITY_GATE = true？
  ├── 否 → 跳过质量门禁 → 直接进入打包阶段
  └── 是 → waitForQualityGate()
              │
              ├── status = "OK"    → ✅ 通过 → 继续构建
              ├── status = "WARN"  → ⚠️ 警告 → Pipeline 中止（当前设计）
              └── status = "ERROR" → ❌ 未通过 → Pipeline 中止
```

**这是整个流水线最核心的质量卡控点。** 当前设计：

| Quality Gate 状态 | 含义 | Pipeline 行为 |
|-------------------|------|--------------|
| `OK` | 所有指标达标 | **继续构建** |
| `WARN` | 有警告但可接受 | **构建失败**（严格模式） |
| `ERROR` | 代码质量不达标 | **构建失败** |

**失败后处理**：
- Jenkins 控制台输出 "SonarQube Quality Gate 未通过: {status}"
- 回调平台标记 quality_gate 阶段为 failed
- 开发者需登录 SonarQube Dashboard 查看具体问题并修复

#### ⑦-⑨ Package → Build Image → Push Image

```
每个阶段：
  └── 成功 → 继续下一阶段
  └── 失败 → Pipeline 中止
```

**Build Image 使用 nerdctl**（containerd 环境，非 Docker）

---

## 三、触发构建的前置条件

### 3.1 什么情况下可以发布？

平台在触发构建前会进行一系列检查：

```go
// 服务层 PipelineRun 方法的完整检查链
func PipelineRun(req, userID) {
    ① 流水线存在？              → 不存在则拒绝
    ② 流水线未被禁用？          → disabled 则拒绝
    ③ 流水线未在运行中？        → running 则拒绝（除非 force=true）
    ④ Jenkins 配置完整？        → 配置缺失则失败
    ⑤ 以上全部通过 → 允许触发
}
```

**详细决策表**：

| 检查项 | 条件 | 结果 |
|--------|------|------|
| 流水线不存在 | `PipelineGetByID` 返回 404 | 拒绝，返回 "流水线不存在" |
| 流水线已禁用 | `status == "disabled"` | 拒绝，返回 "流水线已禁用" |
| 正在运行 + 非强制 | `status == "running" && !force` | 拒绝，返回 "请等待完成或使用强制运行" |
| 正在运行 + 强制 | `status == "running" && force` | 停止旧构建 → 重新触发 |
| 上次失败/中止 | `last_run_status == failed/aborted` | **允许**，自动重置状态 |
| 状态正常 | `status == "idle"` | **允许** |

### 3.2 触发方式

| 触发方式 | trigger_type | 说明 |
|---------|-------------|------|
| **手动触发** | `manual` | 用户在 Web 界面点击"运行" |
| **Webhook 触发** | `webhook` | Git Push / PR Merge 自动触发 |
| **定时触发** | `scheduled` | Cron 定时任务触发 |

### 3.3 构建参数传递链

```
用户/Webhook 请求
       │
       ▼
   平台服务层
       │
       ├── 1. 基础参数：GIT_REPO, GIT_BRANCH, IMAGE_REPO
       ├── 2. 回调参数：PIPELINE_ID, PLATFORM_CALLBACK_URL
       ├── 3. 语言参数：injectLanguageParams() 自动注入
       │      └── Go:    GO_VERSION=1.24
       │      └── Java:  JAVA_VERSION=17, ENABLE_SONAR=true, ...
       │      └── Front: NODE_VERSION=18, BUILD_COMMAND=npm run build
       │      └── Python: PYTHON_VERSION=3.11
       ├── 4. 流水线环境变量：pipeline.EnvVars（优先级低）
       └── 5. 请求环境变量：req.EnvVars（优先级最高）
              │
              ▼
        Jenkins Job（参数化构建）
```

---

## 四、构建失败场景与处理策略

### 4.1 失败场景分类

| 失败类型 | 阶段 | 严重程度 | 自动恢复？ | 处理方式 |
|---------|------|---------|----------|---------|
| **Git 拉取失败** | Checkout | 🔴 严重 | 否 | 检查仓库地址/凭证/网络 |
| **编译失败** | Compile | 🔴 严重 | 否 | 修复代码语法错误 |
| **测试失败** | Test | 🟡 中等 | 否 | 修复测试 或 `SKIP_TESTS=true` |
| **代码扫描失败** | SonarQube | 🟡 中等 | 否 | 检查 SonarQube Server 状态 |
| **质量门禁未通过** | Quality Gate | 🟡 中等 | 否 | 修复代码质量问题 |
| **镜像构建失败** | Build | 🔴 严重 | 否 | 检查 Dockerfile 和依赖 |
| **镜像推送失败** | Push | 🟡 中等 | 可重试 | 检查 Harbor 凭证/磁盘空间 |
| **Jenkins 触发失败** | 触发阶段 | 🔴 严重 | 否 | 检查 Jenkins 配置 |
| **Jenkins 超时** | 触发阶段 | 🟡 中等 | 可重试 | 增大超时时间 |
| **部署失败** | Deploy | 🔴 严重 | 否 | 检查集群/命名空间/权限 |
| **Rollout 超时** | Deploy | 🟡 中等 | 否 | 检查镜像是否可拉取/资源是否足够 |

### 4.2 失败后会发生什么？

```
Jenkins 阶段失败
       │
       ├── 1. stageCallback(stage, 'failed')  → 平台记录该阶段失败
       │
       ├── 2. Pipeline 后续阶段全部跳过
       │
       ├── 3. callbackPlatform('FAILURE', msg)  → 最终回调
       │                │
       │                ▼
       │        平台收到回调
       │                │
       │                ├── 更新运行记录状态 → "failed"
       │                ├── 更新流水线状态   → "idle"（允许重新触发）
       │                ├── 发送钉钉通知     → 通知开发者
       │                └── 跳过自动部署     → 不会部署失败的构建
       │
       └── 4. 用户可以：
               ├── 查看日志 → 定位失败原因
               ├── 修复代码 → 重新提交
               └── 重新运行 → 无需 force，失败状态自动重置
```

### 4.3 失败后能否继续/跳过？

**当前设计采用"严格模式"：任何阶段失败都中止整条流水线。**

但提供了以下"弹性"机制：

| 场景 | 弹性策略 | 如何使用 |
|------|---------|---------|
| 不想跑测试 | `SKIP_TESTS=true` | 创建流水线时设置 |
| 不想做代码扫描 | `ENABLE_SONAR=false` | 创建流水线时设置 |
| 扫描了但不想卡构建 | `SONAR_QUALITY_GATE=false` | 只扫描不阻断 |
| 前端 Lint 失败不阻断 | 模板内 `\|\| true` | 默认行为 |
| 上次失败想重跑 | 直接重新运行 | 平台自动重置 failed 状态 |
| 正在运行想强制重跑 | `force=true` | 停止旧构建，启动新构建 |

### 4.4 灵活度矩阵

```
                 严格 ◄─────────────────────────► 宽松
                  │                                 │
质量门禁         SONAR_QUALITY_GATE=true          SONAR_QUALITY_GATE=false
代码扫描         ENABLE_SONAR=true                 ENABLE_SONAR=false
单元测试         SKIP_TESTS=false                  SKIP_TESTS=true
前端Lint         修改模板去掉 || true               默认不阻断
人工审批         require_approval=true              require_approval=false
```

---

## 五、回调与状态同步机制

### 5.1 双层回调协议

```
┌─────────────────────────────────────────────────────┐
│                    Jenkins Pipeline                   │
│                                                       │
│  每个阶段完成后 ──▶ stageCallback(type, status)      │
│                          │                            │
│                     POST /stage/callback              │
│                     X-Signature: HMAC-SHA256           │
│                          │                            │
│  Pipeline 结束后 ──▶ callbackPlatform(status, msg)   │
│                          │                            │
│                     POST /pipeline/callback            │
│                     X-Signature: HMAC-SHA256           │
│                     Body: image_url, digest, ...      │
└──────────────────────┼───────────────────────────────┘
                       │
                       ▼
               ┌───────────────┐
               │  平台后端     │
               │               │
               │  1. 验证签名  │
               │  2. 幂等检查  │
               │  3. 更新状态  │
               │  4. 自动部署  │
               │  5. 发送通知  │
               └───────────────┘
```

### 5.2 安全机制

| 机制 | 实现方式 | 说明 |
|------|---------|------|
| **HMAC-SHA256 签名** | `X-Signature` Header | 防止伪造回调 |
| **签名数据格式** | `job_name:build_number:status` | 冒号分隔 |
| **时序攻击防护** | `crypto/subtle.ConstantTimeCompare` | 常量时间比较 |
| **幂等控制** | `callback_received` 标记 | 防止重复处理 |
| **未配置密钥** | 跳过验证（开发模式） | 生产环境必须配置 |

### 5.3 状态机

```
Pipeline 状态:
  idle ──(触发)──▶ running ──(完成)──▶ idle
                              │
                              ├── last_run_status = success
                              ├── last_run_status = failed
                              └── last_run_status = aborted

Run 状态:
  pending ──(Jenkins响应)──▶ running ──(回调)──▶ success / failed / aborted
```

---

## 六、构建成功后的部署决策

### 6.1 部署决策树

```
Jenkins 构建成功 + 镜像推送成功
         │
         ▼
    auto_deploy = true？
         │
    ┌────┴────┐
    否        是
    │         │
  不部署   部署配置完整？（namespace + workload + container）
    │         │
    │    ┌────┴────┐
    │    否        是
    │    │         │
    │  跳过部署  require_approval = true？
    │              │
    │         ┌────┴────┐
    │         是        否
    │         │         │
    │   创建审批记录  获取 K8s 客户端
    │   发送审批通知    │
    │   等待人工审批    ├── target_cluster_id > 0？
    │                   │       ├── 是 → 多集群模式（指定集群）
    │                   │       └── 否 → 单集群模式（默认管理集群）
    │                   │
    │                   ▼
    │              Patch 更新工作负载镜像
    │                   │
    │              等待 Rollout 完成（5分钟超时）
    │                   │
    │              ┌────┴────┐
    │              成功      失败
    │              │         │
    │         更新部署信息  更新部署状态=failed
    │         发送成功通知  发送失败通知
    │
    ▼
  发布结束
```

### 6.2 部署环境与审批策略

| 环境 | deploy_env | 建议审批策略 | 代码扫描策略 |
|------|-----------|------------|------------|
| **开发环境** | `dev` | 无需审批 | 扫描但不阻断 |
| **测试环境** | `test` | 无需审批 | 扫描 + 警告 |
| **预发环境** | `staging` | 可选审批 | 扫描 + 质量门禁 |
| **生产环境** | `prod` | **强制审批** | **扫描 + 严格质量门禁** |

### 6.3 Rollout 完成判定

```go
// 成功条件（全部满足）：
dp.Status.ObservedGeneration >= dp.Generation     // 控制器已处理
dp.Status.UpdatedReplicas == replicas              // 所有副本已更新
dp.Status.AvailableReplicas == replicas            // 所有副本可用

// 失败条件：
cond.Type == "Progressing" && cond.Reason == "ProgressDeadlineExceeded"  // 超时
```

---

## 七、通知闭环

### 7.1 通知时机

| 事件 | 通知类型 | 接收方 |
|------|---------|--------|
| 构建开始 | 钉钉消息 | 触发人 |
| 构建成功 | 钉钉消息 | 触发人 |
| 构建失败 | 钉钉消息 | 触发人 |
| 需要审批 | 钉钉消息 | 审批人 |
| 部署成功 | 钉钉消息 | 触发人 + 运维 |
| 部署失败 | 钉钉消息 | 触发人 + 运维 |
| Rollout 完成 | 钉钉消息 | 触发人 |

### 7.2 Jenkins 最终回调包含的信息

```json
{
  "job_name": "k8s-builder-java",
  "build_number": 42,
  "status": "SUCCESS",
  "pipeline_id": 1,
  "image_url": "harbor.example.com/app/user-service:main-abc1234-20260413",
  "image_digest": "sha256:abc123...",
  "image_with_digest": "harbor.example.com/app/user-service@sha256:abc123...",
  "git_commit": "abc1234",
  "git_branch": "main",
  "duration_sec": 180,
  "message": "Java 项目构建成功 | SonarQube: OK",
  "build_url": "http://jenkins.example.com/job/k8s-builder-java/42/"
}
```

---

## 八、配置参数速查表

### 8.1 控制发布行为的关键参数

| 参数 | 默认值 | 作用 | 影响范围 |
|------|--------|------|---------|
| `SKIP_TESTS` | `false` | 跳过单元测试 | Test 阶段 |
| `ENABLE_SONAR` | `true`(Java) | 启用代码扫描 | SonarQube 阶段 |
| `SONAR_QUALITY_GATE` | `true` | 启用质量门禁 | Quality Gate 阶段 |
| `auto_deploy` | `false` | 构建后自动部署 | 部署阶段 |
| `require_approval` | `false` | 部署前需人工审批 | 审批阶段 |
| `force` | `false` | 强制重跑（停止旧构建） | 触发阶段 |

### 8.2 超时配置

| 超时项 | 默认值 | 配置位置 |
|--------|--------|---------|
| Jenkins Pipeline 总超时 | Go: 30min / Java: 45min / Frontend: 30min | Groovy 模板 `options.timeout` |
| Jenkins 触发等待 | 60s | `config.yaml` → `TriggerTimeout` |
| Quality Gate 等待 | 5min | SonarQube Server 配置 |
| Rollout 完成等待 | 5min | 服务层硬编码 |
| Rollout 轮询间隔 | 5s | 服务层硬编码 |

---

## 九、完整状态流转图

```
                    ┌──────────────┐
                    │   用户触发    │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
  前置检查失败 ◄─── │ 前置检查     │
  （返回错误）      │ ①存在 ②启用  │
                    │ ③非运行/强制 │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │ 创建运行记录  │  status = pending
                    │ 创建阶段记录  │
                    │ 更新→running │
                    └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │ 异步触发     │
  Jenkins配置错误 ◄─│ Jenkins 构建 │  status = running
  status=failed     └──────┬───────┘
                           ▼
                    ┌──────────────┐
                    │ Jenkins 执行  │
                    │ Pipeline     │  阶段回调持续推送
                    └──────┬───────┘
                      ┌────┴────┐
                      ▼         ▼
               ┌──────────┐ ┌──────────┐
               │ 构建成功  │ │ 构建失败  │
               │ status=   │ │ status=   │
               │ success   │ │ failed    │
               └────┬─────┘ └────┬─────┘
                    │             │
            ┌───────┤         记录失败
            ▼       ▼         发送通知
       需要审批  无需审批      ▼ 结束
            │       │
       等待审批  自动部署
            │    ┌──┴──┐
         ┌──┴──┐ 成功  失败
       批准  拒绝 │     │
         │     │ 通知  通知
       部署   结束 ▼    ▼
         │      结束  结束
       成功/失败
         │
         ▼
       结束
```

---

## 十、总结：设计原则

| 原则 | 实现方式 |
|------|---------|
| **快速失败** | 任何阶段失败立即中止，不浪费计算资源 |
| **弹性可配** | 测试/扫描/门禁/审批均可按项目灵活开关 |
| **幂等安全** | 回调幂等 + HMAC 签名 + CAS 状态机 |
| **状态可追溯** | 每个阶段独立记录状态、时长、日志 |
| **失败可恢复** | 失败后自动重置状态，无需人工干预即可重跑 |
| **渐进式质量** | 代码扫描默认开启但可关闭，适应不同成熟度的项目 |
| **通知闭环** | 构建/审批/部署每个关键节点都有钉钉通知 |
| **多环境适配** | dev 宽松 → prod 严格，质量要求逐级提升 |
