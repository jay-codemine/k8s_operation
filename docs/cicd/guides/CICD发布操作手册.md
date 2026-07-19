# CI/CD 发布操作手册

> 本手册面向开发人员，指导如何使用 K8sOperation 平台完成项目的持续集成与持续部署。

---

## 一、平台概述

K8sOperation CI/CD 平台提供从代码提交到生产部署的全链路能力：

```
代码提交 → 创建流水线 → 触发构建 → 镜像推送 → 审批(可选) → 自动部署 → 发布单追踪
```

**核心模块：**
| 模块 | 功能 |
|------|------|
| 流水线管理 | 创建/编辑/运行/停止流水线 |
| 阶段化构建 | 14阶段全链路可视化（代码检出→部署） |
| 发布单管理 | 多集群发布、批量发布、回滚 |
| 审批流程 | 生产环境发布审批（支持飞书） |
| 制品库 | 构建产物归档与下载 |
| 环境管理 | dev/test/staging/prod 环境配置 |

---

## 二、前置条件

### 2.1 账号与权限
- 使用平台账号登录（默认 `admin/admin123`）
- 确保拥有 CI/CD 模块的操作权限

### 2.2 基础设施
- **Jenkins 服务**：已部署并配置好 Jenkins 服务
- **镜像仓库**：Harbor 或其他镜像仓库地址
- **K8s 集群**：平台已接入目标部署集群
- **Git 凭证**：Jenkins 中已配置 Git 仓库访问凭证

---

## 三、创建流水线

### 3.1 通过前端向导创建

1. 进入 **CI/CD → 流水线** 页面
2. 点击 **创建流水线** 按钮
3. 按步骤填写：

#### 步骤一：基本信息
| 字段 | 说明 | 示例 |
|------|------|------|
| 流水线名称 | 唯一标识，建议 `项目名-环境` | `user-service-prod` |
| 描述 | 流水线用途说明 | 用户服务生产环境部署 |

#### 步骤二：代码仓库
| 字段 | 说明 | 示例 |
|------|------|------|
| Git 仓库地址 | HTTPS/SSH 协议均可 | `https://gitee.com/org/user-service.git` |
| 分支 | 默认构建分支 | `main` |

> 💡 点击「获取分支」按钮可自动拉取远程分支列表

#### 步骤三：构建配置
| 字段 | 说明 | 可选值 |
|------|------|--------|
| 语言类型 | 决定使用哪个构建模板 | `go` / `java` / `frontend` / `python` / `custom` |
| Jenkins Job | 非 custom 类型可留空自动推导 | `go-pipeline` |
| Jenkins URL | Jenkins 服务地址（可选） | `http://jenkins.example.com` |

**语言类型与 Jenkins Job 映射表：**
| 语言类型 | 自动映射 Job | 构建模板 |
|---------|-------------|---------|
| go | `go-pipeline` | go-pipeline.groovy |
| java | `java-spring-pipeline` | java-spring-pipeline.groovy |
| frontend | `frontend-pipeline` | frontend-pipeline.groovy |
| python | `python-pipeline` | python-pipeline.groovy |
| custom | 需手动指定 | 自定义 |

#### 步骤四：部署配置（可选）
| 字段 | 说明 | 示例 |
|------|------|------|
| 自动部署 | 构建成功后是否自动部署 | ✅ 启用 |
| 目标集群 | 部署到哪个 K8s 集群 | 选择已接入的集群 |
| 命名空间 | 目标 Namespace | `production` |
| 工作负载类型 | Deployment/StatefulSet/DaemonSet | `Deployment` |
| 工作负载名称 | 要更新的工作负载 | `user-service` |
| 容器名称 | 要更新镜像的容器（留空则更新第一个） | `app` |
| 部署环境 | 环境标识 | `dev` / `test` / `staging` / `prod` |
| 需要审批 | 是否启用发布审批 | ✅ 生产环境建议开启 |

#### 步骤五：高级配置（可选）
| 字段 | 说明 |
|------|------|
| 启用 SonarQube | 开启代码质量扫描 |
| 启用制品上传 | 构建完成后上传制品到制品库 |
| 发布联动告警静默 | 部署期间自动静默告警 |
| 静默缓冲时间 | 部署完成后延长静默（默认10分钟） |
| 环境变量 | 自定义 Key-Value 环境变量列表 |

### 3.2 使用模板快速创建

在创建页面左侧可选择 **快速模板**，选中后自动填充预置配置。

### 3.3 批量导入流水线

适用于一次性接入多个项目：

```json
POST /api/v1/k8s/cicd/pipeline/batch-create

{
  "skip_existing": true,
  "pipelines": [
    {
      "name": "user-service",
      "git_repo": "https://gitee.com/org/user-service.git",
      "git_branch": "main",
      "language_type": "go",
      "auto_deploy": true,
      "target_namespace": "production",
      "target_workload_kind": "Deployment",
      "target_workload_name": "user-service",
      "target_container": "app"
    },
    {
      "name": "order-service",
      "git_repo": "https://gitee.com/org/order-service.git",
      "language_type": "java",
      "auto_deploy": true,
      "target_namespace": "production",
      "target_workload_kind": "Deployment",
      "target_workload_name": "order-service",
      "target_container": "app"
    }
  ]
}
```

---

## 四、运行流水线（触发构建）

### 4.1 手动触发

1. 进入 **CI/CD → 流水线** 列表
2. 找到目标流水线，点击 **运行** 按钮
3. 可选覆盖分支和环境变量：
   - **分支**：覆盖默认分支（如临时发布 hotfix 分支）
   - **环境变量**：运行时传入额外参数

### 4.2 强制运行

如果流水线显示「运行中」但实际 Jenkins 已结束（状态不同步），可使用**强制运行**：
- 自动停止旧构建
- 清理历史运行记录
- 触发新构建

### 4.3 批量运行

选中多条流水线 → 点击 **批量运行**，适用于全量发布场景。

### 4.4 停止构建

- 点击 **停止** 按钮会同时：
  1. 调用 Jenkins API 中止构建
  2. 更新平台运行记录状态为 `aborted`

---

## 五、构建阶段（14阶段全链路）

流水线运行后可实时查看各阶段执行状态：

| 阶段 | 说明 | 默认启用 |
|------|------|---------|
| 1. 清理工作空间 | 清理 Jenkins workspace | ✅ |
| 2. 代码检出 | 拉取 Git 指定分支代码 | ✅ |
| 3. 依赖下载 | go mod download / mvn install / npm install | ✅ |
| 4. 编译检查 | 编译验证 + 产出二进制 | ✅ |
| 5. 单元测试 | 执行测试用例 | ✅ |
| 6. 代码检查 | lint / format 检查 | ✅ |
| 7. SonarQube 扫描 | 代码质量扫描 | ❌ 按需 |
| 8. 质量门禁 | 检查覆盖率/Bug/异味阈值 | ❌ 按需 |
| 9. 构建制品 | 产出可部署二进制/JAR | ❌ 按需 |
| 10. 上传制品库 | 制品归档到平台 | ❌ 按需 |
| 11. 打包镜像 | Docker Build（支持 BuildKit 缓存） | ✅ |
| 12. 推送镜像 | Push 到 Harbor/Registry | ✅ |
| 13. 人工审批 | 等待审批通过后继续 | ❌ 按需 |
| 14. 部署 | 更新 K8s 工作负载镜像 | ❌ 按需 |

### 5.1 查看阶段详情

在流水线详情页：
1. 点击流水线名称进入详情
2. 查看 **运行历史** 列表
3. 点击某次运行记录可查看各阶段状态
4. 点击阶段可查看详细日志

### 5.2 审批阶段

当流水线配置了 `require_approval = true`：
1. 构建到审批阶段时暂停，状态变为 `waiting`
2. 审批人在 **CI/CD → 审批** 页面看到待审批项
3. 点击 **通过** 或 **拒绝**
4. 通过后自动继续执行后续阶段

### 5.3 部署阶段

启用 `auto_deploy` 后，构建成功会自动执行部署：
- 使用 `kubectl set image` 方式更新工作负载镜像
- 支持 Deployment / StatefulSet / DaemonSet
- 支持指定容器名称

---

## 六、查看构建日志

### 6.1 实时日志

1. 进入流水线详情页
2. 正在运行的流水线会实时刷新日志
3. 支持增量加载（从指定行号开始获取）

### 6.2 历史日志

1. 进入 **运行历史**
2. 选择某次构建记录
3. 查看完整构建日志

---

## 七、发布单管理

发布单用于**多集群部署**场景，可将一个镜像同时发布到多个集群。

### 7.1 创建发布单

进入 **CI/CD → 发布单** → **创建发布单**：

| 字段 | 说明 | 示例 |
|------|------|------|
| 应用名称 | 应用标识 | `user-service` |
| 命名空间 | 目标 Namespace | `production` |
| 工作负载类型 | Deployment/StatefulSet/DaemonSet | `Deployment` |
| 工作负载名称 | 要更新的工作负载 | `user-service` |
| 容器名称 | 目标容器 | `app` |
| 镜像仓库 | 镜像 Repo 地址 | `harbor.example.com/proj/user-service` |
| 镜像 Tag | 要部署的版本 | `v1.2.0` |
| 目标集群 | 选择要部署的集群（多选） | cluster-a, cluster-b |
| 发布策略 | rolling / blue-green | `rolling` |
| 超时时间 | 单个任务超时（秒） | `300` |
| 并发数 | 同时执行的集群数 | `3` |

### 7.2 发布单状态

| 状态 | 说明 |
|------|------|
| Pending | 已创建，等待执行 |
| Queued | 已入队，等待 Worker 处理 |
| Running | 正在执行部署 |
| Succeeded | 所有集群部署成功 |
| Failed | 部分或全部集群部署失败 |
| Canceled | 已取消 |
| Rollback | 已回滚 |

### 7.3 发布单操作

| 操作 | 适用状态 | 说明 |
|------|---------|------|
| 重试 | Failed | 重新执行失败的任务 |
| 回滚 | Succeeded/Failed | 回滚到上一个版本 |
| 取消 | Pending/Queued/Running | 取消当前发布 |
| 编辑 | Pending/Failed/Canceled | 修改发布配置 |
| 删除 | Pending/Failed/Canceled/Succeeded | 删除发布记录 |
| 批量重试 | Failed | 批量重新发布 |
| 批量回滚 | Succeeded/Failed | 批量回滚 |

---

## 八、环境管理

### 8.1 创建环境

进入 **CI/CD → 环境管理** → **创建环境**：

| 字段 | 说明 | 示例 |
|------|------|------|
| 环境名称 | 唯一标识 | `prod` |
| 显示名称 | 界面展示名 | `生产环境` |
| 关联集群 | 此环境对应的 K8s 集群 | 选择集群 |
| 命名空间 | 默认部署命名空间 | `production` |
| 需要审批 | 发布到此环境是否需审批 | ✅ 生产环境建议开启 |
| 审批人员 | 配置审批人列表 | 选择用户 |

### 8.2 推荐环境配置

```
dev（开发环境）  → 无需审批，自动部署
test（测试环境） → 无需审批，自动部署
staging（预发） → 需要审批
prod（生产）   → 需要审批
```

---

## 九、制品库

### 9.1 查看制品

进入 **CI/CD → 制品库**，可查看所有构建产物：
- 镜像类型制品（Docker Image）
- 二进制类型制品（编译产物）

### 9.2 下载制品

点击制品记录的 **下载** 按钮即可下载文件。

### 9.3 制品与构建关联

每次构建成功后如果启用了制品上传，会自动创建制品记录，关联：
- 流水线 ID
- 运行记录 ID
- Git Commit SHA
- 镜像地址

---

## 十、常用操作流程

### 10.1 新项目接入 CI/CD（推荐流程）

```
1. 确认 Git 仓库地址和默认分支
2. 确认项目语言类型（go/java/frontend/python）
3. 在平台创建流水线（选择语言类型，Job 自动推导）
4. 配置部署目标（集群、命名空间、工作负载）
5. 首次手动运行验证
6. 验证通过后启用自动部署
```

### 10.2 日常发布（开发环境）

```
1. 提交代码到 main/dev 分支
2. 进入平台，找到对应流水线
3. 点击「运行」→ 确认分支
4. 等待构建完成（约3-5分钟）
5. 自动部署完成后验证功能
```

### 10.3 生产发布

```
1. 确认要发布的分支/Tag
2. 运行流水线，选择目标分支
3. 构建到「人工审批」阶段暂停
4. 通知审批人审批
5. 审批通过后自动部署
6. 验证生产环境功能
7. 如有问题 → 使用「回滚」功能
```

### 10.4 多集群发布

```
1. 进入「发布单」页面
2. 创建发布单：
   - 填写镜像地址和 Tag
   - 选择多个目标集群
   - 设置并发数
3. 提交后系统自动逐集群执行
4. 查看每个集群的部署状态
5. 失败的集群可单独「重试」
```

### 10.5 紧急回滚

```
1. 进入流水线详情 → 部署历史
2. 找到上一个正常版本
3. 点击「回滚」到指定版本
   - 或：进入发布单 → 点击「回滚」
4. 确认服务恢复正常
```

---

## 十一、极速发布（Quick Deploy）

极速发布是平台提供的**跳过 Jenkins 构建、直接部署已有镜像**的能力，适用于紧急修复、回滚、或镜像已构建好只需更新 K8s 的场景。

### 11.1 极速发布 vs 完整发布对比

| 维度 | 极速发布（Quick Deploy） | 完整发布（Full Pipeline） |
|------|------------------------|-------------------------|
| **场景** | 镜像已构建好，只需更新 K8s 镜像 | 从代码到部署的完整流水线 |
| **是否需要 Jenkins** | ❌ 不需要 | ✅ 需要 |
| **耗时** | 几秒（直接 Patch K8s 工作负载） | 3-10 分钟（构建+推送+部署） |
| **触发方式** | 创建发布单，传入镜像地址直接部署 | 运行流水线，Jenkins 完整构建 |
| **适用场景** | 紧急热修、回滚、镜像已在仓库、多集群同步 | 日常开发、代码变更、首次部署 |
| **代码质量检查** | ❌ 无（镜像已构建） | ✅ 编译/测试/扫描/门禁 |
| **支持多集群** | ✅ 一次发布到多集群 | ✅ 自动部署到配置的集群 |

### 11.2 极速发布操作步骤

#### 方式一：通过前端界面

1. 进入 **CI/CD → 发布单** 页面
2. 点击 **创建发布单**
3. 填写以下信息：

| 字段 | 说明 | 示例 |
|------|------|------|
| 应用名称 | 应用标识 | `user-service` |
| 命名空间 | K8s Namespace | `production` |
| 工作负载类型 | Deployment/StatefulSet/DaemonSet | `Deployment` |
| 工作负载名称 | 要更新的工作负载名称 | `user-service` |
| 容器名称 | 要更新镜像的容器 | `app` |
| 镜像仓库 | Harbor/Registry 地址 | `harbor.example.com/proj/user-service` |
| 镜像 Tag | 要部署的版本 | `v1.2.0` 或 `abc123-20240101` |
| 目标集群 | 选择集群（支持多选） | cluster-prod-1, cluster-prod-2 |

4. 点击 **提交** → 系统立即执行部署
5. 在发布单详情页查看各集群部署进度

#### 方式二：通过 API 调用

```bash
# 极速发布 API —— 直接指定镜像部署到 K8s
curl -X POST http://平台地址/api/v1/k8s/cicd/release/create \
  -H "Authorization: Bearer {TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "app_name": "user-service",
    "namespace": "production",
    "workload_kind": "Deployment",
    "workload_name": "user-service",
    "container_name": "app",
    "image_repo": "harbor.example.com/proj/user-service",
    "image_tag": "v1.2.0",
    "cluster_ids": [1, 2],
    "strategy": "rolling",
    "timeout_sec": 300,
    "concurrency": 3
  }'
```

#### 方式三：基于流水线的极速发布（模板化）

如果已创建流水线，可以传入 `pipeline_id` 自动继承部署配置，极简输入：

```bash
# 只需传入 pipeline_id + image_tag 即可发布
curl -X POST http://平台地址/api/v1/k8s/cicd/release/create \
  -H "Authorization: Bearer {TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "pipeline_id": 5,
    "image_tag": "v1.2.0"
  }'
```

> 系统自动从流水线配置中继承：命名空间、工作负载、容器名、集群等所有部署参数

### 11.3 极速发布典型场景

#### 场景 1：紧急热修
```
1. 开发者本地/CI 构建好修复镜像
2. 推送镜像到 Harbor：harbor.example.com/proj/user-service:hotfix-001
3. 平台「极速发布」→ 填入镜像 Tag：hotfix-001
4. 几秒内完成部署，无需等待完整流水线
```

#### 场景 2：回滚到历史版本
```
1. 进入「发布单」列表
2. 找到上一次成功的发布记录
3. 点击「回滚」按钮
4. 系统自动创建新发布单，使用上次的旧镜像
5. 几秒内回滚完成
```

#### 场景 3：多集群同步部署
```
1. 镜像已通过 CI 验证并推送到仓库
2. 创建发布单，选择所有目标集群（如 3 个集群）
3. 设置并发数为 3（同时部署）
4. 一次操作完成全部集群更新
```

### 11.4 极速发布执行原理

```
创建发布单 → 写入 Redis Stream → Worker 消费任务 → Patch K8s Deployment → 等待 Rollout → 完成
```

执行过程：
1. **创建发布单 + 任务**：按目标集群拆分为多个 Task（原子写入数据库）
2. **入队 Redis Stream**：每个 Task 进入消费队列
3. **Worker 消费**：获取集群 K8s 客户端，执行 `Strategic Merge Patch` 更新容器镜像
4. **等待 Rollout**：轮询检查 Deployment/StatefulSet 是否全部 Ready
5. **完结发布单**：所有 Task 完成后，更新发布单状态

---

## 十二、Jenkins 完整配置指南

### 12.1 架构说明：平台如何驱动 Jenkins

```
┌─────────────────────────────────────────────────────────────────────┐
│                        K8sOperation 平台                            │
│                                                                     │
│  创建流水线（选择 language_type: go/java/frontend/python）            │
│       │                                                             │
│       ▼                                                             │
│  DefaultJenkinsJobMap 自动映射 → Jenkins Job 名称                    │
│       │                                                             │
│       ▼                                                             │
│  运行流水线 → 调用 Jenkins API 触发构建（传入 GIT_REPO、IMAGE_REPO 等）│
│       │                                                             │
│       ▼                                                             │
│  Jenkins 执行通用模板（Pipeline script from SCM）                    │
│       │                                                             │
│       ▼                                                             │
│  构建完成 → Jenkins 回调平台（携带镜像地址、状态）                    │
│       │                                                             │
│       ▼                                                             │
│  平台自动部署到 K8s（如果配置了 auto_deploy）                        │
└─────────────────────────────────────────────────────────────────────┘
```

**核心概念 `DefaultJenkinsJobMap`：**

```go
// 代码位置：internal/app/models/cicd_pipeline.go
var DefaultJenkinsJobMap = map[string]string{
    "go":       "go-pipeline",           // Go 项目 → 对应 Jenkins Job
    "java":     "java-spring-pipeline",  // Java 项目 → 对应 Jenkins Job
    "frontend": "frontend-pipeline",     // 前端项目 → 对应 Jenkins Job
    "python":   "python-pipeline",       // Python 项目 → 对应 Jenkins Job
}
```

**工作流程：**
1. 用户在平台创建流水线，选择 `language_type = go`
2. 平台自动填充 `jenkins_job = go-pipeline`（用户无需手动填写）
3. 运行流水线时，平台调用 Jenkins API 触发 `go-pipeline` 这个 Job
4. Jenkins Job 配置为 "Pipeline script from SCM"，自动拉取 `configs/jenkins-templates/go-pipeline.groovy`
5. 模板中的参数（GIT_REPO、IMAGE_REPO 等）全部由平台自动注入

**结论：用户在平台侧零 Jenkins 配置，Jenkins 只需一次性初始化 4 个通用 Job 即可。**

### 12.2 Jenkins 一次性初始化（详细步骤）

> 以下配置只需做一次，之后接入任何新项目都不需要碰 Jenkins。

#### 前置要求

| 项目 | 要求 |
|------|------|
| Jenkins 版本 | 2.387+ (LTS) |
| 必装插件 | Pipeline, Git, Credentials, HTTP Request |
| 可选插件 | SonarQube Scanner（代码扫描时需要） |
| 网络 | Jenkins 能访问 Git 仓库 + 镜像仓库 + 平台回调地址 |

#### 步骤 1：配置 3 个全局凭证

进入 Jenkins → **Manage Jenkins** → **Credentials** → **System** → **Global credentials** → **Add Credentials**

##### 凭证 ①：Git 仓库凭证
| 字段 | 值 |
|------|----|
| Kind | Username with password |
| Scope | Global |
| Username | Git 用户名（如 Gitee 账号） |
| Password | Git 密码或 Personal Access Token |
| ID | `gitee-id` |
| Description | Git 仓库拉取凭证 |

> 如果使用 Gitee/GitHub，推荐使用 Personal Access Token 代替密码

##### 凭证 ②：镜像仓库凭证
| 字段 | 值 |
|------|----|
| Kind | Username with password |
| Scope | Global |
| Username | Harbor/Registry 用户名 |
| Password | Harbor/Registry 密码 |
| ID | `harbor-registry` |
| Description | 镜像仓库推送凭证 |

> 如使用 Harbor Robot Account，Username 填 `robot$xxx`，Password 填 Token

##### 凭证 ③：HMAC 签名密钥
| 字段 | 值 |
|------|----|
| Kind | Secret text |
| Scope | Global |
| Secret | 与平台 config.yaml 中 `HMACSecret` 相同的值 |
| ID | `hmac-secret` |
| Description | 回调签名密钥（防伪造） |

> 重要：此密钥必须与平台 `configs/config.yaml` 中 `Jenkins.HMACSecret` 完全一致

#### 步骤 2：创建 4 个通用 Builder Job

每个 Job 的创建步骤完全相同，只有 **名称** 和 **Script Path** 不同。

##### Job ①：go-pipeline（Go 项目）

1. Jenkins 首页 → 左侧 **New Item**
2. 输入名称：`go-pipeline`
3. 选择类型：**Pipeline** → 点击 OK
4. 配置页面：

| 配置区域 | 字段 | 值 |
|---------|------|----|
| General | Description | K8s 平台 Go 项目通用构建模板 |
| General | ☑ This project is parameterized | 不需要手动添加参数（模板自带） |
| Pipeline | Definition | **Pipeline script from SCM** |
| Pipeline → SCM | SCM | Git |
| Pipeline → SCM | Repository URL | `https://gitee.com/your-org/k8s_operation.git`（平台代码仓库） |
| Pipeline → SCM | Credentials | 选择 `gitee-id` |
| Pipeline → SCM | Branch Specifier | `*/main` |
| Pipeline | Script Path | `configs/jenkins-templates/go-pipeline.groovy` |
| Pipeline | Lightweight checkout | ☑ 勾选（加速脚本加载） |

5. 点击 **Save** 保存

##### Job ②：java-spring-pipeline（Java/Spring Boot 项目）

完全同上步骤，仅以下不同：
| 字段 | 值 |
|------|----|
| 名称 | `java-spring-pipeline` |
| Description | K8s 平台 Java/Spring Boot 通用构建模板 |
| Script Path | `configs/jenkins-templates/java-spring-pipeline.groovy` |

> 额外要求：Jenkins 服务器需安装 JDK 和 Maven，并在 Jenkins → Manage Jenkins → Tools 中配置：
> - Maven: 名称 `Maven-3.9`，路径 `/opt/apache-maven-3.9.9`
> - JDK: 名称 `JDK-21`，路径 `/usr/lib/jvm/java-21`

##### Job ③：frontend-pipeline（前端 Vue/React 项目）

完全同上步骤，仅以下不同：
| 字段 | 值 |
|------|----|
| 名称 | `frontend-pipeline` |
| Description | K8s 平台前端项目通用构建模板（Vue/React/Angular） |
| Script Path | `configs/jenkins-templates/frontend-pipeline.groovy` |

> 额外要求：Jenkins 服务器需安装 Node.js 18+

##### Job ④：python-pipeline（Python 项目）

完全同上步骤，仅以下不同：
| 字段 | 值 |
|------|----|
| 名称 | `python-pipeline` |
| Description | K8s 平台 Python 项目通用构建模板 |
| Script Path | `configs/jenkins-templates/python-pipeline.groovy` |

> 额外要求：Jenkins 服务器需安装 Python 3.11+

#### 步骤 3：验证 Jenkins 配置

配置完成后，验证清单：

```bash
# 1. 确认 4 个 Job 都已创建
打开 Jenkins 首页，应看到：
  - go-pipeline
  - java-spring-pipeline
  - frontend-pipeline
  - python-pipeline

# 2. 确认凭证
Jenkins → Manage Jenkins → Credentials，应看到：
  - gitee-id（Username/Password）
  - harbor-registry（Username/Password）
  - hmac-secret（Secret text）

# 3. 确认网络连通（在 Jenkins 服务器上执行）
curl -I http://平台地址:38180/healthz/live    # 平台健康检查
curl -I https://gitee.com                      # Git 仓库可达
curl -I https://harbor.example.com/v2/         # 镜像仓库可达
```

### 12.3 平台侧 Jenkins 配置（config.yaml）

平台通过 `configs/config.yaml` 连接 Jenkins：

```yaml
Jenkins:
  # --- 核心连接配置 ---
  URL: "http://你的Jenkins地址:8080/"       # Jenkins 服务地址
  Username: "ops-dev"                        # Jenkins 用户名（需有 Job 触发权限）
  APIToken: "xxxxx"                          # Jenkins API Token（非登录密码！）

  # --- 回调与安全配置 ---
  TriggerTimeout: 60                         # 触发超时（秒）
  CallbackURL: "http://平台后端地址:38180"    # Jenkins 回调平台的地址
  PlatformURL: "http://前端地址:38181"        # 前端页面地址（通知链接用）
  HMACSecret: "my_super_secret_hmac_key"     # 签名密钥（需与 Jenkins hmac-secret 凭证一致）

  # --- 凭证 ID（对应 Jenkins Credentials 中的 ID） ---
  GitCredentialID: "gitee-id"                # Git 拉取凭证
  RegistryCredentialID: "harbor-registry"    # 镜像推送凭证
  HMACCredentialID: "hmac-secret"            # HMAC 签名凭证

  # --- 运行控制 ---
  PollInterval: 15                           # 状态轮询间隔（秒）
  MaxBuildTime: 30                           # 最大构建时间（分钟）
```

> **获取 Jenkins API Token 的方法：**
> 1. 登录 Jenkins → 点击右上角用户名 → Configure
> 2. API Token 区域 → Add new Token → 命名并生成
> 3. 复制 Token 填入 config.yaml 的 `APIToken` 字段

### 12.4 Jenkins 服务器环境要求

| 语言类型 | 服务器需安装 | 版本要求 |
|---------|-------------|----------|
| go | Go、nerdctl/docker | Go 1.24+, nerdctl 1.7+ |
| java | JDK、Maven、nerdctl/docker | JDK 17/21, Maven 3.9+ |
| frontend | Node.js、npm、nerdctl/docker | Node 18+, npm 9+ |
| python | Python、pip、nerdctl/docker | Python 3.11+ |
| 全部 | openssl（HMAC 签名用） | 任意版本 |

> 推荐使用 `nerdctl`（containerd 原生 CLI）替代 Docker，避免 Docker Desktop 授权问题。
> 如使用 Docker，将模板中的 `nerdctl` 替换为 `docker` 即可。

### 12.5 回调机制说明

Jenkins 构建完成后会自动回调平台接口：

| 项目 | 说明 |
|------|------|
| 回调地址 | `{CallbackURL}/api/v1/k8s/cicd/pipeline/callback` |
| 请求方式 | POST JSON |
| 安全校验 | HMAC-SHA256 签名（Header: `X-Signature`） |
| 携带信息 | job_name、build_number、status、image_url、image_digest、duration_sec |
| 幂等保证 | pipeline_id + build_number 唯一，重复回调自动忽略 |

回调触发后平台自动执行：
1. 更新流水线运行记录状态
2. 更新各阶段状态（成功/失败）
3. 如配置了 `auto_deploy`：自动部署到 K8s
4. 发送钉钉/飞书通知
5. 同步发布记录到发布管理页面

### 12.6 完成后的使用效果

配置完成后，用户在平台侧的操作：

```
1. 创建流水线 → 选择语言类型（如 go）→ 填写 Git 地址 → 保存
   ✅ Jenkins Job 自动映射为 go-pipeline（无需手动填写）

2. 点击「运行」按钮
   ✅ 平台自动调用 Jenkins API，注入所有必要参数

3. 等待构建完成
   ✅ Jenkins 执行模板，构建镜像并推送
   ✅ 构建完成后自动回调平台

4. 自动部署（如果配置了）
   ✅ 平台收到回调后自动 Patch K8s Deployment
```

**从第 2 步开始，Jenkins 的一切操作对用户透明，无需进入 Jenkins 界面。**

---

## 十三、故障排查

### 12.1 构建失败

| 现象 | 可能原因 | 解决方案 |
|------|---------|---------|
| 代码检出失败 | Git 凭证无效 | 检查 Jenkins 凭证配置 |
| 依赖下载超时 | 网络问题 | 检查 Jenkins 网络/代理配置 |
| 编译失败 | 代码错误 | 查看日志定位编译错误 |
| 镜像推送失败 | Registry 认证失败 | 检查 harbor-registry 凭证 |
| 部署失败 | K8s 集群连接失败 | 检查集群配置和网络 |

### 12.2 流水线卡住

- **状态一直「运行中」**：
  1. 检查 Jenkins 是否正常
  2. 使用「强制运行」重置状态
  3. 或手动停止流水线

- **审批阶段长时间等待**：
  1. 通知审批人操作
  2. 或取消此次运行重新触发

### 12.3 部署失败回滚

- 进入阶段详情 → 部署历史 → 选择上一版本回滚
- 或使用发布单的「回滚」功能

---

## 十四、API 快速参考

### 流水线

| 操作 | 方法 | 路径 |
|------|------|------|
| 列表 | GET | `/api/v1/k8s/cicd/pipeline/list?page=1&page_size=10` |
| 详情 | GET | `/api/v1/k8s/cicd/pipeline/detail?id=1` |
| 创建 | POST | `/api/v1/k8s/cicd/pipeline/create` |
| 更新 | POST | `/api/v1/k8s/cicd/pipeline/update` |
| 删除 | POST | `/api/v1/k8s/cicd/pipeline/delete` |
| 运行 | POST | `/api/v1/k8s/cicd/pipeline/run` |
| 停止 | POST | `/api/v1/k8s/cicd/pipeline/stop` |
| 批量运行 | POST | `/api/v1/k8s/cicd/pipeline/batch-run` |
| 批量停止 | POST | `/api/v1/k8s/cicd/pipeline/batch-stop` |
| 查看日志 | GET | `/api/v1/k8s/cicd/pipeline/logs?id=1` |
| 运行状态 | GET | `/api/v1/k8s/cicd/pipeline/status?id=1` |
| 运行历史 | GET | `/api/v1/k8s/cicd/pipeline/history?id=1` |
| 阶段数据 | GET | `/api/v1/k8s/cicd/pipeline/stages?id=1` |

### 发布单

| 操作 | 方法 | 路径 |
|------|------|------|
| 列表 | GET | `/api/v1/k8s/cicd/release/list` |
| 详情 | GET | `/api/v1/k8s/cicd/release/detail?id=1` |
| 统计 | GET | `/api/v1/k8s/cicd/release/stats` |
| 创建 | POST | `/api/v1/k8s/cicd/release/create` |
| 编辑 | POST | `/api/v1/k8s/cicd/release/update` |
| 删除 | POST | `/api/v1/k8s/cicd/release/delete` |
| 取消 | POST | `/api/v1/k8s/cicd/release/cancel` |
| 回滚 | POST | `/api/v1/k8s/cicd/release/rollback` |
| 重试 | POST | `/api/v1/k8s/cicd/release/retry` |
| 批量重试 | POST | `/api/v1/k8s/cicd/release/batch-retry` |
| 批量回滚 | POST | `/api/v1/k8s/cicd/release/batch-rollback` |

### 审批

| 操作 | 方法 | 路径 |
|------|------|------|
| 待审批列表 | GET | `/api/v1/k8s/cicd/approval/pending` |
| 审批操作 | POST | `/api/v1/k8s/cicd/approval/action` |

### 阶段

| 操作 | 方法 | 路径 |
|------|------|------|
| 阶段列表 | GET | `/api/v1/k8s/cicd/stage/list?run_id=1` |
| 阶段日志 | GET | `/api/v1/k8s/cicd/stage/logs?id=1` |
| 执行部署 | POST | `/api/v1/k8s/cicd/stage/deploy` |
| 回滚部署 | POST | `/api/v1/k8s/cicd/stage/rollback` |

---

## 十五、最佳实践

### 14.1 分支策略建议

```
main/master  →  生产环境（需审批）
release/*    →  预发环境
develop      →  测试环境（自动部署）
feature/*    →  开发环境（手动触发）
hotfix/*     →  紧急修复 → 直接发布到生产
```

### 14.2 流水线命名规范

```
格式：{项目名}-{环境}
示例：
  user-service-dev
  user-service-prod
  frontend-web-staging
```

### 14.3 环境变量安全

- 敏感信息（密码、密钥）使用 Jenkins Credentials 管理
- 不要在环境变量中明文存储密码
- 使用 `HMAC_SECRET` 保护回调接口安全

### 14.4 发布联动告警静默

生产发布时建议启用：
- 部署期间自动静默低级别告警（warning/info）
- 避免滚动更新时触发大量无效告警
- critical 级别告警默认不静默

---

## 十六、常见问题（FAQ）

**Q: 如何变更默认构建分支？**
A: 编辑流水线 → 修改「Git 分支」字段。运行时也可临时覆盖分支。

**Q: 构建成功但没有自动部署？**
A: 检查流水线是否开启了「自动部署」，以及部署配置（集群/命名空间/工作负载）是否完整填写。

**Q: Jenkins Job 不存在怎么办？**
A: 需要在 Jenkins 中创建对应名称的 Job，参考第十一节「Jenkins 配置指南」。

**Q: 如何只触发构建不部署？**
A: 关闭流水线的「自动部署」开关即可。构建成功后镜像会推送到仓库，后续可手动创建发布单部署。

**Q: 多个环境如何配置？**
A: 建议每个环境创建独立流水线，分别配置不同分支和不同目标集群/命名空间。

**Q: 发布单和流水线部署有什么区别？**
A: 流水线部署是「完整发布」模式（代码→构建→部署），需要 Jenkins；发布单是「极速发布」模式（已有镜像→直接部署），不需要 Jenkins。

**Q: 极速发布和完整发布如何选择？**
A: 
- 日常开发、代码变更 → 使用「完整发布」（运行流水线，自动构建+部署）
- 紧急修复、回滚、镜像已存在 → 使用「极速发布」（创建发布单，秒级部署）
- 多集群同步 → 使用「极速发布」（一次发布到多个集群）

**Q: 回滚是回滚到哪个版本？**
A: 回滚到 K8s 记录的上一个 ReplicaSet 版本（即上一次成功部署的版本）。
