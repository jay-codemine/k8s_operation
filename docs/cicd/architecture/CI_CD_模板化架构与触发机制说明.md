# CI/CD 模板化架构与触发机制说明

> 本文档详细说明平台的模板化 CI/CD 架构设计、Jenkins 触发机制、以及业务项目零侵入的核心理念。
> 解答常见疑问：模板放在哪里？业务项目需要放 Groovy 文件吗？100 个项目怎么管理？

---

## 一、核心架构：4 个 Job 服务 100+ 项目

```
                    ┌──────────────────────────────────────────┐
                    │        平台仓库（k8s_operation）           │
                    │        只有这一个仓库存放模板               │
                    │                                          │
                    │  configs/jenkins-templates/               │
                    │    ├── java-spring-pipeline.groovy        │
                    │    ├── go-pipeline.groovy                 │
                    │    ├── frontend-pipeline.groovy           │
                    │    └── python-pipeline.groovy             │
                    └──────────────────────────────────────────┘
                                      ▲
                                      │ SCM 指向平台仓库
                    ┌─────────────────┴────────────────────────┐
                    │  Jenkins（只需创建 4 个 Job，一次性配置）    │
                    │                                          │
                    │  k8s-builder-java     ← Script Path:     │
                    │    configs/jenkins-templates/             │
                    │    java-spring-pipeline.groovy            │
                    │                                          │
                    │  k8s-builder-go       ← Script Path:     │
                    │    configs/jenkins-templates/             │
                    │    go-pipeline.groovy                     │
                    │                                          │
                    │  k8s-builder-frontend ← Script Path:     │
                    │    configs/jenkins-templates/             │
                    │    frontend-pipeline.groovy               │
                    │                                          │
                    │  k8s-builder-python   ← Script Path:     │
                    │    configs/jenkins-templates/             │
                    │    python-pipeline.groovy                 │
                    └──────────────────────────────────────────┘

   100 个业务项目：完全不需要任何 Groovy/Jenkinsfile

   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       ┌─────────────┐
   │ order-service│  │ user-service │  │ payment-svc │  ...  │ 第100个项目  │
   │             │  │             │  │             │       │             │
   │ ├── src/    │  │ ├── src/    │  │ ├── src/    │       │ ├── src/    │
   │ ├── pom.xml │  │ ├── go.mod  │  │ ├── pom.xml │       │ ├── pom.xml │
   │ └── Docker- │  │ └── Docker- │  │ └── Docker- │       │ └── Docker- │
   │     file    │  │     file    │  │     file    │       │     file    │
   │             │  │             │  │             │       │             │
   │ 没有Groovy! │  │ 没有Groovy! │  │ 没有Groovy! │       │ 没有Groovy! │
   └─────────────┘  └─────────────┘  └─────────────┘       └─────────────┘
```

---

## 二、触发机制详解：双 Git 拉取

平台触发一次 Jenkins 构建，实际涉及 **两次 Git 拉取** 和 **两次 HTTP 调用**：

### 2.1 完整触发流程

```
  用户在平台点击「运行」
       │
       ▼
  ① 平台后端（Go）构建参数
     params = {
       GIT_REPO:    "https://gitee.com/org/order-service.git"   ← 业务项目地址
       GIT_BRANCH:  "main"
       IMAGE_REPO:  "harbor.example.com/library/order-service"
       PIPELINE_ID: "15"
       PLATFORM_CALLBACK_URL: "http://platform:8080/api/v1/k8s/cicd/pipeline/callback"
       JAVA_VERSION: "17"
       ENABLE_SONAR: "true"
       ...
     }
       │
       ▼
  ② 平台调用 Jenkins REST API 触发构建
     HTTP POST → http://jenkins:8080/job/k8s-builder-java/buildWithParameters
     Body: GIT_REPO=https://gitee.com/org/order-service.git&GIT_BRANCH=main&...
     Auth: Basic Auth（Jenkins Username + API Token）
       │
       ▼
  ③ Jenkins 收到请求
     │
     ├─ 【第 1 次 Git 拉取】从平台仓库拉取 Groovy 模板
     │   Git Clone: https://gitee.com/your-org/k8s_operation.git
     │   读取文件: configs/jenkins-templates/java-spring-pipeline.groovy
     │   （这是 Jenkins Job SCM 配置决定的，与业务项目无关）
     │
     ├─ 加载模板中的 parameters{} 块，接收平台传来的参数
     │
     ├─ 【第 2 次 Git 拉取】模板 Checkout 阶段拉取业务项目代码
     │   Git Clone: ${params.GIT_REPO}  ← 即 order-service.git
     │   Branch: ${params.GIT_BRANCH}   ← 即 main
     │   （这是模板代码中 checkout 步骤决定的，参数由平台传入）
     │
     ├─ 执行构建阶段：Compile → Test → SonarQube → Package → Build Image → Push Image
     │
     ├─ 每个阶段完成 → HTTP 回调平台 /stage/callback
     │
     └─ 全部完成 → HTTP 回调平台 /pipeline/callback
           │
           ▼
  ④ 平台收到回调
     ├── 更新流水线状态
     ├── 记录镜像信息
     └── 自动部署到 K8s（如果配置了 auto_deploy）
```

### 2.2 两个仓库对比

| | 平台仓库（k8s_operation） | 业务项目仓库（如 order-service） |
|---|---|---|
| **谁拉取** | Jenkins Job 的 SCM 配置自动拉取 | Groovy 模板的 Checkout 阶段拉取 |
| **什么时候拉** | 构建开始前（加载 Pipeline 脚本） | 构建开始后（第一个 stage） |
| **拉什么** | 1 个 Groovy 模板文件 | 业务项目全部源代码 |
| **配置在哪** | Jenkins Job → Pipeline → SCM → Repository URL | 平台创建流水线时填的 `git_repo` 字段 |
| **目的** | 获取构建流程定义（怎么构建） | 获取要编译/打包/部署的代码（构建什么） |

### 2.3 触发方式源码

**平台后端触发 Jenkins：**

```go
// internal/app/services/cicd_pipeline.go → PipelineRun()

// 1. 构建参数
params := make(map[string]string)
params["GIT_BRANCH"] = branch
params["GIT_REPO"] = pipeline.GitRepo             // 业务项目仓库地址
params["PIPELINE_ID"] = fmt.Sprintf("%d", pipeline.ID)
params["PLATFORM_CALLBACK_URL"] = callbackURL + "/api/v1/k8s/cicd/pipeline/callback"

// 2. 自动注入语言特定参数（如 JAVA_VERSION=17, ENABLE_SONAR=true）
s.injectLanguageParams(pipeline, params)

// 3. 异步触发 Jenkins 构建
go s.triggerJenkinsBuild(ctx, pipeline, run, params)
```

**Jenkins 客户端调用：**

```go
// pkg/jenkins/client.go → TriggerBuild()

// 调用 Jenkins REST API
path = "/job/k8s-builder-java/buildWithParameters"
values := url.Values{}
for k, v := range params {
    values.Set(k, v)    // GIT_REPO=xxx&GIT_BRANCH=main&PIPELINE_ID=15&...
}
resp, err := c.doRequest(ctx, http.MethodPost, path, strings.NewReader(values.Encode()))
```

**Groovy 模板中拉取业务代码：**

```groovy
// configs/jenkins-templates/java-spring-pipeline.groovy

parameters {
    string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
    string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
    // ...
}

stage('Clean Workspace') {
    steps {
        deleteDir()
        script {
            checkout([
                $class: 'GitSCM',
                branches: [[name: "*/${params.GIT_BRANCH}"]],
                userRemoteConfigs: [[
                    url: params.GIT_REPO,              // ← 业务项目仓库（平台传参）
                    credentialsId: 'gitee-id'
                ]]
            ])
        }
    }
}
```

---

## 三、业务项目零侵入

### 3.1 业务项目只需要什么

```
my-java-app/                    ← 你的 Java 项目
├── Dockerfile                  ← 唯一需要添加的文件
├── pom.xml                     ← 项目本身就有
└── src/main/java/...           ← 项目本身就有

my-go-app/                      ← 你的 Go 项目
├── Dockerfile                  ← 唯一需要添加的文件
├── go.mod                      ← 项目本身就有
└── cmd/main.go                 ← 项目本身就有

my-frontend/                    ← 你的前端项目
├── Dockerfile                  ← 唯一需要添加的文件
├── package.json                ← 项目本身就有
└── src/                        ← 项目本身就有
```

**不需要的文件：**
- ❌ Jenkinsfile
- ❌ `.jenkins/` 目录
- ❌ `*.groovy` 文件
- ❌ 任何 CI/CD 配置文件

### 3.2 与传统方式对比

| 对比项 | 传统方式（每个项目放 Jenkinsfile） | 本平台模板化方式 |
|--------|----------------------------------|----------------|
| 业务项目需要 | Jenkinsfile + Dockerfile | 仅 Dockerfile |
| 100 个项目维护 | 100 份 Jenkinsfile | 0 份（统一在平台仓库） |
| 修改构建流程 | 改 100 个项目的 Jenkinsfile | 改平台仓库 1 个 Groovy 文件 |
| Jenkins Job 数量 | 100 个（每个项目 1 个 Job） | 4 个（每种语言 1 个 Job） |
| 新项目接入 | 写 Jenkinsfile + 创建 Jenkins Job | 在平台上创建流水线即可（3 分钟） |
| 构建参数管理 | 分散在各项目 Jenkinsfile 中 | 集中在平台，UI 可视化配置 |
| 版本一致性 | 各项目可能用不同版本的构建流程 | 所有项目共用同一套标准化流程 |

---

## 四、100+ 项目管理方案

### 4.1 Jenkins 侧：只需 4 个 Job（一次性配置）

| 语言 | Job 名称 | Script Path | 服务项目数 |
|------|---------|-------------|-----------|
| Java | `k8s-builder-java` | `configs/jenkins-templates/java-spring-pipeline.groovy` | 所有 Java 项目 |
| Go | `k8s-builder-go` | `configs/jenkins-templates/go-pipeline.groovy` | 所有 Go 项目 |
| Frontend | `k8s-builder-frontend` | `configs/jenkins-templates/frontend-pipeline.groovy` | 所有前端项目 |
| Python | `k8s-builder-python` | `configs/jenkins-templates/python-pipeline.groovy` | 所有 Python 项目 |

4 个 Job 的 SCM 配置**完全相同**（同一个平台仓库、同一个分支），只有 Script Path 不同。

### 4.2 平台侧：批量创建流水线

编辑 `scripts/batch-import-pipelines.json`，填入所有项目信息：

```json
{
  "skip_existing": true,
  "pipelines": [
    {"name": "order-service",   "git_repo": "https://gitee.com/org/order-service.git",   "language_type": "java"},
    {"name": "user-service",    "git_repo": "https://gitee.com/org/user-service.git",    "language_type": "go"},
    {"name": "admin-web",       "git_repo": "https://gitee.com/org/admin-web.git",       "language_type": "frontend"},
    {"name": "data-processor",  "git_repo": "https://gitee.com/org/data-processor.git",  "language_type": "python"}
  ]
}
```

运行导入脚本一次性创建：

```powershell
.\scripts\batch-import-pipelines.ps1 -ApiUrl http://platform:8080
```

详见 `docs/模板化CICD快速接入指南.md` 第 2.6 节。

### 4.3 完整管理矩阵

| 管理维度 | 100 个项目的工作量 | 说明 |
|---------|-------------------|------|
| Jenkins Job | 创建 4 个（一次性） | 按语言类型各 1 个 |
| Jenkins 凭证 | 配置 3 个（一次性） | harbor-registry、gitee-id、hmac-secret |
| 平台流水线 | 批量导入（5 分钟） | 编辑 JSON + 运行脚本 |
| 业务项目改动 | 每个项目加 Dockerfile | 无其他改动 |
| 模板更新 | 改 1 个文件（自动生效） | 所有项目下次构建自动使用新模板 |

---

## 五、常见疑问

### Q1: Jenkins Job 的 SCM 指向哪个仓库？

**指向平台仓库**（`k8s_operation.git`），不是业务项目仓库。

Jenkins Job 的 SCM 配置：
```
Repository URL: https://gitee.com/your-org/k8s_operation.git   ← 平台仓库
Script Path:    configs/jenkins-templates/java-spring-pipeline.groovy
```

### Q2: 那业务项目代码怎么拉取？

通过 Groovy 模板内的 `checkout` 步骤拉取，业务项目地址由**平台传参**（`GIT_REPO` 参数）。

### Q3: 100 个项目同时构建不会冲突吗？

不会。Jenkins 每次触发构建都是独立的 Build，参数不同（`GIT_REPO`、`IMAGE_REPO` 不同）。
同一个 Job 通过 `disableConcurrentBuilds()` 防止并发，多个项目会排队。
如需并行构建，可以创建 Job 副本（如 `k8s-builder-java-1`、`k8s-builder-java-2`）。

### Q4: 模板更新后怎么生效？

因为用的是 **Pipeline script from SCM** 模式，只需 `git push` 到平台仓库，Jenkins 下次构建时自动拉取最新模板。
不需要登录 Jenkins 手动修改任何 Job。

### Q5: 自定义语言/框架怎么办？

设置 `language_type: "custom"`，手动指定 `jenkins_job` 名称，自己在 Jenkins 创建对应的 Pipeline Job。

---

## 六、文件引用

| 文件路径 | 说明 |
|---------|------|
| `configs/jenkins-templates/java-spring-pipeline.groovy` | Java 通用构建模板 |
| `configs/jenkins-templates/go-pipeline.groovy` | Go 通用构建模板 |
| `configs/jenkins-templates/frontend-pipeline.groovy` | 前端通用构建模板 |
| `configs/jenkins-templates/python-pipeline.groovy` | Python 通用构建模板 |
| `pkg/jenkins/client.go` | Jenkins REST API 客户端（TriggerBuild） |
| `internal/app/services/cicd_pipeline.go` | 流水线运行 + 触发逻辑（PipelineRun） |
| `internal/app/models/cicd_pipeline.go` | 语言类型 → Job 名称映射表 |
| `scripts/batch-import-pipelines.json` | 批量导入 JSON 模板 |
| `scripts/batch-import-pipelines.ps1` | 批量导入 PowerShell 脚本 |
| `docs/模板化CICD快速接入指南.md` | 完整接入指南（含批量创建） |
| `docs/Java项目CICD完整接入指南.md` | Java 项目专属接入指南 |
