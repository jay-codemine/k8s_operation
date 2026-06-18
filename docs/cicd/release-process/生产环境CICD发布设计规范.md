# 生产环境 CI/CD 发布设计规范

> K8sOperation 平台 · 企业级生产发布最佳实践 · 多环境分级管控方案

---

## 一、设计总纲

### 1.1 核心原则

生产环境发布遵循一条铁律：

> **代码必须扫、测试不能跳、门禁必须过、审批不能省、发布可回滚、全程有通知。**

### 1.2 设计理念

| 理念 | 说明 | 落地方式 |
|------|------|---------|
| **分级管控** | 环境越接近生产，管控越严格 | dev 宽松 → staging 中等 → prod 最严 |
| **质量左移** | 问题发现越早成本越低 | 编译 → 测试 → 扫描 → 门禁，逐层过滤 |
| **快速失败** | 任何阶段失败立即中止 | Jenkins Pipeline 串行执行，失败即停 |
| **镜像不变** | 同一镜像跨环境晋升 | `image@sha256:digest` 保证一致性 |
| **发布可逆** | 每次发布都可回滚 | 记录历史版本 + K8s 原生 Rollout Undo |
| **全程可追溯** | 从代码到上线全链路审计 | 阶段回调 + 钉钉通知 + 数据库记录 |

---

## 二、多环境分级管控

### 2.1 四环境模型

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         K8sOperation 多环境发布链                         │
│                                                                          │
│  ┌─────────┐    ┌─────────┐    ┌──────────┐    ┌──────────────┐         │
│  │   DEV    │───▶│  TEST   │───▶│ STAGING  │───▶│    PROD      │         │
│  │  开发环境 │    │  测试环境 │    │  预发环境  │    │   生产环境    │         │
│  └─────────┘    └─────────┘    └──────────┘    └──────────────┘         │
│                                                                          │
│  自动化: 高 ──────────────────────────────────────────────▶ 低            │
│  严格度: 低 ──────────────────────────────────────────────▶ 高            │
│  人工干预: 无 ─────────────────────────────────────────────▶ 多            │
└──────────────────────────────────────────────────────────────────────────┘
```

### 2.2 各环境管控对比

| 维度 | DEV 开发 | TEST 测试 | STAGING 预发 | **PROD 生产** |
|------|---------|----------|-------------|--------------|
| **触发方式** | Git Push 自动 | MR 合并自动 | 手动触发 | **手动触发** |
| **代码分支** | `feature/*` | `develop` | `release` | **`release` / `main`** |
| **单元测试** | 可跳过 | 必须执行 | 必须执行 | **必须执行** |
| **代码扫描** | 关闭 | 开启（不阻断） | 开启（质量门禁） | **开启 + 严格质量门禁** |
| **质量门禁** | `SONAR_QUALITY_GATE=false` | 仅警告 | 必须通过 | **必须通过** |
| **人工审批** | 不需要 | 不需要 | 可选 | **强制审批** |
| **审批人** | - | - | 技术负责人 | **技术负责人 + 运维负责人** |
| **部署方式** | 自动部署 | 自动部署 | 自动部署 | **审批通过后自动部署** |
| **回滚支持** | 不需要 | 手动回滚 | 一键回滚 | **支持一键回滚 + 自动告警** |
| **通知范围** | 无 | 构建结果 | 全流程通知 | **全流程 + 值班群通知** |
| **发布窗口** | 随时 | 工作时间 | 工作时间 | **工作日 10:00-17:00** |
| **操作审计** | 基础日志 | 基础日志 | 完整审计 | **完整审计 + 留档** |

### 2.3 环境对应的平台配置

平台中通过 `deploy_env` 字段区分环境，钉钉通知中会自动显示环境标识：

```go
// 平台已实现的环境标识
DeployEnvDev     = "dev"      // 🔧 开发环境
DeployEnvTest    = "test"     // 🧪 测试环境
DeployEnvStaging = "staging"  // 📦 预发环境
DeployEnvProd    = "prod"     // 🚀 生产环境
```

---

## 三、生产环境完整发布流程

### 3.1 发布全流程图

```
                    ┌──────────────────────────────┐
                    │  第一步：开发者提交 MR 到 release  │
                    └──────────────┬───────────────┘
                                   ▼
                    ┌──────────────────────────────┐
                    │  第二步：Code Review + MR 合并    │
                    │  - 至少 2 位审核人通过            │
                    │  - CI 自动运行（test 环境通过）    │
                    └──────────────┬───────────────┘
                                   ▼
                    ┌──────────────────────────────┐
                    │  第三步：运维/负责人在平台触发构建  │
                    │  - 选择 prod 流水线              │
                    │  - 确认分支为 release/main        │
                    └──────────────┬───────────────┘
                                   ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    CI 阶段（Jenkins Pipeline 自动执行）                │
│                                                                      │
│  ① Clean      清理工作空间 + 拉取代码                                  │
│      │                                                               │
│      ▼                                                               │
│  ② Checkout   采集 Git 信息，生成镜像标签                               │
│      │         tag = {branch}-{commit}-{timestamp}                   │
│      ▼                                                               │
│  ③ Compile    编译检查（语法错误、依赖缺失）                             │
│      │         Java: mvn clean compile -DskipTests -B                │
│      ▼                                                               │
│  ④ Test       单元测试（生产环境不允许跳过）                             │
│      │         Java: mvn test -B                                     │
│      │         报告: **/target/surefire-reports/*.xml                 │
│      ▼                                                               │
│  ⑤ SonarQube  代码质量扫描                                             │
│      │         Bug/漏洞/异味/覆盖率/重复率                              │
│      │         mvn sonar:sonar → SonarQube Server                    │
│      ▼                                                               │
│  ⑥ Quality    质量门禁检查（生产环境必须开启）                           │
│    Gate       waitForQualityGate()                                   │
│      │         status = OK → 继续                                    │
│      │         status ≠ OK → 构建失败，中止发布                        │
│      ▼                                                               │
│  ⑦ Package    打包（mvn clean package）                               │
│      │         归档产物: **/target/*.jar                              │
│      ▼                                                               │
│  ⑧ Build      构建容器镜像                                             │
│    Image      nerdctl build -t {image}:{tag}                        │
│      │         注入 git.commit / git.branch / build.number 标签      │
│      ▼                                                               │
│  ⑨ Push       推送到 Harbor 镜像仓库                                   │
│    Image      nerdctl push {image}:{tag}                            │
│      │         获取 image digest (sha256)                            │
│      ▼                                                               │
│  Jenkins 回调平台                                                     │
│  callbackPlatform('SUCCESS', image, digest, ...)                     │
└──────────────────────────────┬───────────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    平台回调处理                                        │
│                                                                      │
│  1. HMAC-SHA256 签名验证（防伪造）                                     │
│  2. 幂等检查（callback_received 标记，防重复处理）                       │
│  3. 更新运行记录状态 → success                                        │
│  4. 更新流水线状态 → idle                                             │
│  5. 更新阶段状态（所有构建阶段 → success）                              │
│  6. 发送钉钉通知「⏳ 待审批」                                          │
└──────────────────────────────┬───────────────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    人工审批阶段（生产环境必须）                          │
│                                                                      │
│  审批人收到钉钉通知，点击链接进入平台审批页面                             │
│                                                                      │
│  审批人看到：                                                         │
│  ┌────────────────────────────────────────────────┐                  │
│  │  流水线: user-service-prod                      │                  │
│  │  环境: 🚀 生产环境                               │                  │
│  │  分支: release                                  │                  │
│  │  构建号: #42                                    │                  │
│  │  镜像: harbor.example.com/app/user:release-abc  │                  │
│  │  SonarQube: Quality Gate ✅ OK                  │                  │
│  │                                                 │                  │
│  │  [✅ 批准]  [❌ 拒绝]                            │                  │
│  └────────────────────────────────────────────────┘                  │
│                                                                      │
│  ┌────────┐                   ┌────────┐                             │
│  │  批准   │                   │  拒绝   │                             │
│  └───┬────┘                   └───┬────┘                             │
│      │                            │                                   │
│      ▼                            ▼                                   │
│  继续自动部署                  记录拒绝原因                              │
│                               发送拒绝通知                              │
│                               流水线结束                                │
└──────────────────────────────┬───────────────────────────────────────┘
                               │ (审批通过)
                               ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    自动部署阶段                                        │
│                                                                      │
│  1. 获取目标集群 K8s 客户端                                            │
│     - 多集群模式: target_cluster_id → 指定集群                         │
│     - 单集群模式: 默认管理集群                                          │
│                                                                      │
│  2. 构造镜像地址（优先 image@digest 确保不可变）                         │
│                                                                      │
│  3. Patch 更新工作负载                                                 │
│     Deployment / StatefulSet / DaemonSet                              │
│     StrategicMergePatch → 更新容器镜像                                 │
│                                                                      │
│  4. 等待 Rollout 完成（5 分钟超时）                                     │
│     ┌─────────────────────────────────────────┐                      │
│     │  每 5 秒轮询 Deployment 状态:            │                      │
│     │  - ObservedGeneration >= Generation ？   │                      │
│     │  - UpdatedReplicas == 期望副本数 ？       │                      │
│     │  - AvailableReplicas == 期望副本数 ？     │                      │
│     │  - ProgressDeadlineExceeded ？ → 失败    │                      │
│     └─────────────────────────────────────────┘                      │
│                                                                      │
│  5. 更新流水线部署信息                                                  │
│     last_deploy_image / last_deploy_digest / last_deploy_time         │
│                                                                      │
│  6. 发送钉钉通知                                                       │
│     - 成功: ✅ 自动部署成功                                             │
│     - 失败: ❌ 自动部署失败 + 错误详情                                   │
└──────────────────────────────────────────────────────────────────────┘
```

### 3.2 状态机流转

```
Pipeline 状态:

    idle ──(用户触发)──▶ running ──(回调完成)──▶ idle
                                    │
                            last_run_status = success / failed / aborted

Run 状态:

    pending ──(Jenkins响应)──▶ running ──(回调)──▶ success ──(审批)──▶ deployed
                                        │                     │
                                        └──▶ failed           └──▶ rejected
                                        │
                                        └──▶ aborted

Stage 状态:

    pending ──▶ running ──▶ success
                    │
                    └──▶ failed
                    │
                    └──▶ skipped (when 条件不满足)

    pending ──▶ waiting (审批阶段，构建成功后)
                    │
                    ├──▶ success (批准)
                    └──▶ failed  (拒绝)
```

---

## 四、生产环境推荐配置

### 4.1 流水线创建配置

```json
{
  "name": "user-service-prod",
  "description": "用户服务 - 生产环境发布",
  "language_type": "java",
  "git_repo": "https://gitee.com/company/user-service.git",
  "git_branch": "release",
  "deploy_env": "prod",

  "auto_deploy": true,
  "require_approval": true,
  "target_cluster_id": 1,
  "target_namespace": "production",
  "target_workload_kind": "Deployment",
  "target_workload_name": "user-service",
  "target_container": "user-service",

  "env_vars": [
    { "name": "SKIP_TESTS", "value": "false" },
    { "name": "ENABLE_SONAR", "value": "true" },
    { "name": "SONAR_QUALITY_GATE", "value": "true" },
    { "name": "JAVA_VERSION", "value": "17" }
  ]
}
```

### 4.2 平台全局配置 (config.yaml)

```yaml
Jenkins:
  URL: "http://jenkins.internal:8080/"
  Username: "admin"
  APIToken: "xxxxxxxxxx"
  TriggerTimeout: 60                              # 触发超时(秒)
  CallbackURL: "http://k8s-platform:10537"        # 后端 API（Jenkins 回调用）
  PlatformURL: "http://k8s-web.internal:30851"    # 前端页面（钉钉链接用）
  HMACSecret: "your-hmac-secret-key"              # HMAC 签名密钥（Jenkins 和平台必须一致）
  PollInterval: 15                                # 轮询间隔(秒)
  MaxBuildTime: 30                                # 最大构建时间(分钟)
  DingTalkWebhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
```

### 4.3 SonarQube 质量门禁配置

生产环境推荐使用 **"Sonar Way"** 或更严格的自定义 Quality Profile：

| 指标 | 推荐阈值 | 说明 |
|------|---------|------|
| 新代码 Bug | **= 0** | 新增代码不允许有 Bug |
| 新代码漏洞 | **= 0** | 新增代码不允许有安全漏洞 |
| 新代码覆盖率 | **≥ 80%** | 业界公认合理标准，核心项目可提到 85% |
| 新代码重复率 | **≤ 3%** | 杜绝复制粘贴 |
| 安全热点审核 | **100% 已审核** | 所有安全敏感代码需人工确认 |
| 可维护性评级 | **A** | 新代码不允许降级 |
| 可靠性评级 | **A** | 新代码不允许引入 Bug |
| 安全性评级 | **A** | 新代码不允许引入漏洞 |

> **关键原则 "Clean as You Code"**：只看「新代码」指标，不要求一次性修所有历史问题，存量代码随迭代逐步改善。

---

## 五、安全机制

### 5.1 回调安全

```
Jenkins Pipeline ──────────────────────────▶ 平台后端
                                             │
    1. HMAC-SHA256 签名                      │ 验证流程:
       签名数据: "job_name:build_number:status"  │
       密钥: hmac-secret (Jenkins Credentials)   │ ① 提取 X-Signature Header
                                             │ ② 用相同密钥计算期望签名
    2. X-Signature Header 传递               │ ③ crypto/subtle.ConstantTimeCompare
                                             │    (防时序攻击)
    3. 签名不匹配 → 返回 401 Unauthorized    │ ④ 不匹配 → 拒绝回调
```

### 5.2 幂等控制

```
第一次回调:
    callback_received = 0 → 正常处理 → 更新 callback_received = 1

重复回调（网络重试等）:
    callback_received = 1 → 直接返回 "回调已处理（重复请求）" → 不重复执行部署
```

### 5.3 镜像不可变原则

```
构建阶段:
    nerdctl build → 生成 image:tag
    nerdctl push  → 获取 image@sha256:digest

部署阶段:
    优先使用 image@sha256:digest 部署（非 tag）
    确保部署的镜像与构建产物完全一致，防止 tag 被覆盖导致不一致
```

平台代码实现：
```go
// 构造最终镜像地址（优先使用 image@digest 确保一致性）
finalImage := image
if imageDigest != "" {
    finalImage = image[:idx] + "@" + imageDigest
}
```

---

## 六、分支策略

### 6.1 推荐分支模型

```
feature/xxx ──┐
feature/yyy ──┤──(MR)──▶ develop ──(MR)──▶ release ──(Tag)──▶ main
feature/zzz ──┘              │                 │                  │
                             │                 │                  │
                         DEV 部署          STAGING 部署       PROD 部署
                         TEST 部署
```

| 分支 | 用途 | 部署环境 | 保护规则 |
|------|------|---------|---------|
| `feature/*` | 功能开发 | 本地 / dev | 开发者自由推送 |
| `develop` | 开发主线 | dev / test | MR 合并，CI 必须通过 |
| `release` | 预发/生产 | staging / prod | MR 合并，≥2 人审核 |
| `main` | 生产稳定版 | prod | 仅从 release 合并，禁止直推 |
| `hotfix/*` | 紧急修复 | staging → prod | 可从 main 拉取，修复后合并回 develop |

### 6.2 生产发布流程

```
1. 开发完成 → feature/* 分支提 MR 到 develop
2. develop CI 通过 → 合并
3. release 准备 → 从 develop 拉 MR 到 release
4. staging 验证通过 → 运维在平台触发 prod 流水线
5. CI + SonarQube 通过 → 等待审批
6. 审批通过 → 自动部署到生产
7. 发布成功 → release 合并到 main（打 Tag）
```

---

## 七、回滚策略

### 7.1 回滚方式

| 策略 | 实现方式 | 触发条件 | 恢复时间 |
|------|---------|---------|---------|
| **K8s 原生 Rollout Undo** | `kubectl rollout undo deployment/xxx` | Rollout 超时自动触发 | < 1 分钟 |
| **平台一键回滚** | 选择历史部署版本回滚 | 用户手动触发 | < 2 分钟 |
| **镜像回滚** | 用上一个成功的 `last_deploy_image` 重新部署 | 用户手动触发 | < 3 分钟 |
| **Git Revert + 重新构建** | 代码回退后重新走完整 CI/CD | 代码层面回滚 | 10-20 分钟 |

### 7.2 平台已实现的回滚能力

```
回滚通知（已实现）:
    ↩️ 回滚成功 / ❌ 回滚失败
    包含：流水线名称、环境、工作负载、目标版本、回滚前后镜像、操作人

取消部署（已实现）:
    ⏹️ 部署已取消（未执行时取消）
    ↩️ 部署已回滚（已执行时取消并回滚）
```

### 7.3 建议的回滚 SOP

```
发现线上异常
    │
    ├── 1. 确认异常是否与本次发布有关
    │      - 查看发布时间 vs 异常开始时间
    │      - 对比 Pod 镜像版本
    │
    ├── 2. 决定回滚方式
    │      - 小问题 → Hotfix 修复后重新发布
    │      - 严重问题 → 立即平台一键回滚
    │
    ├── 3. 执行回滚
    │      - 平台操作 → 选择上一个成功的版本
    │      - 等待 Rollout 完成
    │
    ├── 4. 验证回滚结果
    │      - 检查 Pod 状态
    │      - 验证核心接口正常
    │
    └── 5. 事后复盘
           - 记录问题原因
           - 完善测试用例
           - 更新质量门禁规则
```

---

## 八、通知闭环

### 8.1 生产发布通知链路

```
① 构建开始    ──▶ 钉钉通知 「🚀 构建已触发」
                    - 流水线名称、环境、分支、构建号
                    - [查看流水线详情] [查看 Jenkins 日志]

② 构建成功    ──▶ 钉钉通知 「✅ 构建成功」
                    - 镜像地址、耗时
                    - ⏳ 等待审批提醒（如需审批）

③ 待审批      ──▶ 钉钉通知 「⏳ 待审批」@审批人
                    - 流水线、环境、分支、镜像
                    - [点击进行审批] → 跳转平台审批页面

④ 审批结果    ──▶ 钉钉通知
                    - 批准: 继续自动部署
                    - 拒绝: 记录原因，通知开发者

⑤ 部署结果    ──▶ 钉钉通知
                    - ✅ 自动部署成功: 命名空间/工作负载/镜像
                    - ❌ 自动部署失败: 错误信息 + 链接

⑥ 回滚结果    ──▶ 钉钉通知
                    - ↩️ 回滚成功: 目标版本、回滚前后镜像
                    - ❌ 回滚失败: 错误信息

⑦ 构建失败    ──▶ 钉钉通知 「❌ 构建失败」
                    - 错误信息
                    - [查看流水线详情] [查看 Jenkins 日志]
```

### 8.2 钉钉消息示例（生产部署成功）

```markdown
### ✅ 自动部署成功

**流水线**: user-service-prod

**环境**: 🚀 生产环境

**命名空间**: production

**工作负载**: Deployment/user-service

**镜像**: harbor.example.com/app/user-service:release-abc1234-20260413

**时间**: 2026-04-13 14:30:25

---
🔗 [查看流水线详情](http://k8s-web:30851/cicd/pipelines/1?tab=stages)

🛠 [查看 Jenkins 构建](http://jenkins:8080/job/k8s-builder-java/42/console)
```

---

## 九、故障场景与应急预案

### 9.1 常见故障场景

| 故障 | 影响 | 应急预案 |
|------|------|---------|
| **Jenkins 宕机** | 无法触发构建 | 切换到备用 Jenkins / 手动 nerdctl 构建 |
| **SonarQube 不可达** | 代码扫描失败 | 临时设置 `ENABLE_SONAR=false` 跳过扫描 |
| **Harbor 不可达** | 镜像推送/拉取失败 | 检查 Harbor 状态 / 切换备用 Registry |
| **K8s API Server 不可达** | 部署失败 | 检查集群网络 / 切换集群 |
| **Rollout 超时** | 部署卡住 | 检查镜像拉取 / 资源配额 / Pod 事件 |
| **质量门禁失败** | 构建中止 | 修复代码问题 / 紧急时临时关闭门禁 |
| **审批超时** | 部署延迟 | 管理员强制审批（admin_override） |
| **回调丢失** | 状态不同步 | 平台轮询 Jenkins 构建状态自动同步 |

### 9.2 紧急发布流程（Hotfix）

当生产出现紧急 Bug 需要快速修复时：

```
1. 从 main 拉取 hotfix/* 分支
2. 修复代码 + 单元测试
3. 直接触发 prod 流水线（hotfix 分支）
4. CI 通过后快速审批（可简化审批流程）
5. 部署到生产
6. 验证修复
7. hotfix 合并回 develop + main
```

紧急场景下可调整的参数：

| 参数 | 正常值 | 紧急值 | 说明 |
|------|--------|--------|------|
| `SONAR_QUALITY_GATE` | `true` | `false` | 关闭质量门禁（仅扫描不阻断） |
| `SKIP_TESTS` | `false` | 视情况 | 如修复不影响测试可跳过 |
| 审批流程 | 双人审批 | 单人快速审批 | 紧急放行 |

> **重要**：紧急发布后必须补齐测试 + 代码扫描，防止技术债务积累。

---

## 十、发布检查清单（Checklist）

### 10.1 发布前

- [ ] 代码已通过 Code Review（≥2 人）
- [ ] 所有单元测试通过
- [ ] SonarQube 质量门禁通过（Quality Gate = OK）
- [ ] staging 环境已验证
- [ ] 数据库变更已提前执行（如有）
- [ ] 配置变更已确认（ConfigMap/Secret）
- [ ] 回滚方案已准备
- [ ] 相关人员已通知（开发/测试/运维/产品）
- [ ] 在发布窗口内（工作日 10:00-17:00）

### 10.2 发布中

- [ ] 构建日志无异常
- [ ] 代码扫描报告无严重问题
- [ ] 镜像推送成功
- [ ] 审批已通过
- [ ] Rollout 正常完成
- [ ] Pod 全部 Ready
- [ ] 健康检查通过

### 10.3 发布后（观测 15-30 分钟）

- [ ] 核心接口响应正常
- [ ] 错误率无明显上升（5xx < 0.1%）
- [ ] 响应时间无明显劣化（P99 无突增）
- [ ] 资源使用正常（CPU/Memory 无异常飙升）
- [ ] 业务指标正常（订单/支付/登录等核心链路）
- [ ] 日志无异常堆栈

---

## 十一、进阶演进方向

### 11.1 灰度发布（Canary）

```
当前: 全量滚动更新（K8s 原生 RollingUpdate）
未来: 灰度发布

          全量 Pod
         ┌───────┐
阶段1:   │▓░░░░░░│  10% 新版本（金丝雀）
阶段2:   │▓▓▓░░░░│  30% 新版本
阶段3:   │▓▓▓▓░░░│  50% 新版本
阶段4:   │▓▓▓▓▓▓▓│  100% 新版本
         └───────┘
  任何阶段监控异常 → 自动回滚
```

实现方式：Istio VirtualService + DestinationRule 流量切分

### 11.2 蓝绿部署（Blue-Green）

```
当前版本（Blue）:  Deployment/user-service-blue  ← 接收 100% 流量
新版本（Green）:   Deployment/user-service-green ← 0% 流量

验证 Green 正常后:
  Service 切换指向 Green → Green 接收 100% 流量
  异常时秒级切回 Blue
```

### 11.3 发布后自动观测

```
部署完成后自动触发:
  ├── Prometheus 查询 5xx 错误率
  ├── Prometheus 查询 P99 延迟
  ├── K8s 检查 Pod 重启次数
  └── 15 分钟内异常 → 自动回滚 + 告警
```

---

## 十二、涉及文件清单

| 文件 | 说明 |
|------|------|
| `configs/config.yaml` | 平台全局配置（Jenkins/钉钉/HMAC） |
| `configs/jenkins-templates/go-pipeline.groovy` | Go 通用构建模板 |
| `configs/jenkins-templates/java-spring-pipeline.groovy` | Java 通用构建模板（含 SonarQube） |
| `configs/jenkins-templates/frontend-pipeline.groovy` | 前端通用构建模板 |
| `internal/app/models/cicd_pipeline.go` | 数据模型（状态常量/结构体） |
| `internal/app/services/cicd_pipeline.go` | 核心业务逻辑（触发/回调/部署/回滚） |
| `internal/app/services/cicd_notify.go` | 钉钉通知服务（7 种通知类型） |
| `internal/app/services/cicd_stage.go` | 阶段管理（审批/部署/状态同步） |
| `internal/app/controllers/api/v1/cicd/pipeline_controller.go` | API 控制器 |
| `internal/app/routers/kube_cicd/cicd_router.go` | 路由注册 |
| `k8s-web/src/views/cicd/PipelineDetail.vue` | 流水线详情页 |
| `k8s-web/src/components/cicd/CodeQualityPanel.vue` | 代码质量可视化面板 |

---

## 十三、总结

K8sOperation 平台的生产发布设计，围绕 **安全、质量、效率** 三个核心：

| 维度 | 实现 |
|------|------|
| **安全** | HMAC 签名回调 + 幂等控制 + 镜像 Digest + 人工审批 + 三层 RBAC |
| **质量** | 编译检查 + 单元测试 + SonarQube 扫描 + Quality Gate 门禁 |
| **效率** | 模板化发布（4 Job 服务 100+ 项目） + 自动部署 + 钉钉通知闭环 |
| **可靠** | 快速失败 + 失败自动重置 + 一键回滚 + Rollout 超时检测 |
| **可观测** | 阶段回调实时同步 + 钉钉全流程通知 + 完整审计日志 |

> **一句话总结：生产环境的发布，不是追求快，而是追求稳。每一步都有检查，每一步都可回退。**
