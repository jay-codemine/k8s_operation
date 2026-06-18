# CI/CD 发布流程与镜像仓库架构说明

> 版本：1.0  
> 更新时间：2026-04-22  
> 适用范围：K8s Operation 平台 CI/CD 全链路

---

## 一、发布流程全景

### 1.1 完整发布链路

```
开发者推代码 → 平台触发 Jenkins → Jenkins 执行构建模板
                                      │
                                      ├── 1. Clean + Checkout (清理 + 拉取代码)
                                      ├── 2. Dependencies     (下载依赖)
                                      ├── 3. Compile Check    (编译检查)
                                      ├── 4. Test             (单元测试)
                                      ├── 5. Lint             (代码检查)
                                      ├── 6. SonarQube        (质量扫描) [可选]
                                      ├── 7. Build Binary     (编译二进制) [Go 项目]
                                      ├── 8. Upload Artifact  (上传制品到制品库) [可选]
                                      ├── 9. Build Image      (nerdctl build)
                                      └── 10. Push Image      (nerdctl push → 镜像仓库)
                                              │
                                              ↓ 回调平台
                                    平台收到 image_url + image_digest
                                              │
                                    ├── 自动部署 → PatchDeploymentImage
                                    ├── 需审批   → 创建审批记录，等待人工确认
                                    └── 制品记录 → 关联 Git Commit + Digest
```

### 1.2 流水线阶段类型

| 阶段 | stage_type | 说明 |
|------|-----------|------|
| 清理工作空间 | `clean` | 删除旧文件、拉取代码 |
| 代码检出 | `checkout` | Git clone + 提取 commit 信息 |
| 依赖下载 | `dependencies` | go mod download / npm install / mvn dependency:resolve |
| 编译检查 | `compile` | go build / mvn compile |
| 单元测试 | `test` | go test / pytest / mvn test |
| 代码检查 | `lint` | golangci-lint / eslint |
| SonarQube 扫描 | `sonar` | 代码质量与安全扫描 |
| 质量门禁 | `quality_gate` | SonarQube 结果判定（通过/未通过） |
| 编译产物 | `build_binary` | 编译出二进制文件（Go 项目） |
| 上传制品 | `upload_artifact` | 上传 JAR/Binary/Dist 到平台制品库 |
| 构建镜像 | `build` | nerdctl build |
| 推送镜像 | `push` | nerdctl push 到镜像仓库 |
| 人工审批 | `approval` | 生产环境需人工审批 |
| 部署 | `deploy` | 更新 K8s Deployment 镜像 |

---

## 二、三大模块及职责

平台中有三个相互协作的模块，各司其职：

### 2.1 模块关系图

```
┌─────────────────────────────────────────────────────────────────┐
│                     CI/CD 流水线 (Pipeline)                       │
│  编排构建流程 → 触发 Jenkins → 阶段追踪 → 回调 → 自动部署          │
│  表: cicd_pipeline / cicd_pipeline_run / cicd_pipeline_stage      │
└──────────────┬───────────────────────────────┬──────────────────┘
               │                               │
               ↓                               ↓
┌──────────────────────────┐   ┌──────────────────────────────────┐
│  镜像仓库管理              │   │  制品库 (Artifact)                 │
│  (Image Registry)         │   │                                   │
│                           │   │  记录每次构建的产出物               │
│  管理仓库连接配置          │   │  JAR / Binary / Dist → 可下载      │
│  浏览仓库中的镜像/Tag     │   │  Image 类型 → 记录引用(不存文件)    │
│  Tag 清理策略              │   │  关联 Git Commit + Pipeline        │
│                           │   │                                   │
│  表: image_registry       │   │  表: cicd_artifact                │
│  页面: ImageRepositories  │   │  页面: Artifacts                  │
│        Images             │   │                                   │
│        CleanupPolicies    │   │                                   │
└──────────────────────────┘   └──────────────────────────────────┘
```

### 2.2 职责对比

| 维度 | 镜像仓库管理 | 制品库 | CI/CD 流水线 |
|------|------------|--------|-------------|
| **核心职责** | 管理"仓库在哪里" | 管理"每次构建产出了什么" | 编排"怎么构建和部署" |
| **类比** | Harbor Admin UI | Maven Nexus / GitHub Releases | Jenkins + ArgoCD |
| **数据库表** | `image_registry` | `cicd_artifact` | `cicd_pipeline` / `_run` / `_stage` |
| **存储内容** | 仓库连接信息（地址/凭据/状态） | 构建产物文件 + 镜像引用记录 | 流水线配置 + 运行记录 + 阶段日志 |
| **前端页面** | 仓库列表 / 镜像浏览 / 清理策略 | 制品列表 / 上传下载 | 流水线管理 / 详情 / 发布 |

### 2.3 为什么三个模块都需要保留？

- **镜像仓库管理** — 解决"镜像存在哪里"的问题
  - 管理多个仓库（Harbor、阿里云 ACR、Docker Hub）的连接配置
  - 浏览已有镜像和 Tag，方便运维查看
  - 配置 Tag 清理策略，防止镜像无限膨胀
  - 连接状态检测，发现仓库异常

- **制品库** — 解决"构建了什么"的问题
  - 记录每次构建的产物（JAR、二进制、前端 dist 包），支持下载回溯
  - Image 类型只记录镜像引用（image_repo + tag + digest），不重复存储文件
  - 关联 Git Commit + 流水线，实现完整的构建追溯链
  - 支持手动上传制品（不通过流水线的场景）

- **CI/CD 流水线** — 解决"怎么构建和发布"的问题
  - 编排多阶段构建流程（从代码到镜像）
  - 触发 Jenkins 执行、接收回调、追踪阶段状态
  - 自动部署到 K8s（更新 Deployment 镜像）
  - 审批流程、钉钉通知、Rollout 监控

---

## 三、镜像仓库支持情况

### 3.1 支持的仓库类型

| 仓库类型 | type 值 | 认证方式 | 典型地址 |
|---------|---------|---------|---------|
| **Harbor** (私有) | `harbor` | `username` + `password` | `harbor.example.com/project/app` |
| **阿里云 ACR** | `acr` | `access_key_id` + `access_key_secret` + `region` | `registry.cn-hangzhou.aliyuncs.com/ns/app` |
| **Docker Hub** | `docker` | `username` + `password` | `docker.io/library/nginx` |
| **Google GCR** | `gcr` | 服务账号 JSON | `gcr.io/project/app` |
| **AWS ECR** | `ecr` | IAM 凭证 | `123456.dkr.ecr.us-east-1.amazonaws.com/app` |
| **Quay.io** | `quay` | `username` + `password` | `quay.io/org/app` |

### 3.2 数据库模型

```go
type ImageRegistry struct {
    ID              int64  // 主键
    Name            string // 仓库名称（唯一）
    Type            string // 类型: docker, harbor, acr, gcr, ecr, quay
    URL             string // 仓库地址
    Username        string // 用户名
    Password        string // 密码（加密存储，不返回前端）
    AccessKeyID     string // 阿里云 AccessKey ID
    AccessKeySecret string // 阿里云 AccessKey Secret（加密存储）
    Region          string // 区域（如 cn-hangzhou）
    Insecure        bool   // 是否跳过 TLS 验证
    IsDefault       bool   // 是否默认仓库
    Status          string // 连接状态: connected / disconnected / unknown
}
```

### 3.3 Jenkins 中如何使用

Jenkins 模板通过 `IMAGE_REPO` 参数接收镜像仓库地址，自动提取 Host 进行登录：

```groovy
// 参数传入
parameters {
    string(name: 'IMAGE_REPO', description: '镜像仓库地址')
    // 例如: registry.cn-hangzhou.aliyuncs.com/k8s-gos/springboot-hello
    // 例如: harbor.mycompany.com/prod/order-service
}

// 自动登录
environment {
    REGISTRY_CREDS = credentials('harbor-registry')  // Jenkins 凭据
}

stage('Push Image') {
    steps {
        script {
            def registryHost = params.IMAGE_REPO.split('/')[0]
            // registryHost = "registry.cn-hangzhou.aliyuncs.com" 或 "harbor.mycompany.com"
            sh """
                echo \${REGISTRY_CREDS_PSW} | nerdctl login -u \${REGISTRY_CREDS_USR} \
                    --password-stdin ${registryHost}
                nerdctl push ${env.FULL_IMAGE}
            """
        }
    }
}
```

---

## 四、制品库架构

### 4.1 制品类型

| 类型 | artifact_type | 存储方式 | 典型场景 |
|------|-------------|---------|---------|
| Java JAR | `jar` | 文件上传（本地/S3/OSS） | Spring Boot 可执行 JAR |
| Java WAR | `war` | 文件上传 | Tomcat 部署包 |
| Go 二进制 | `binary` | 文件上传 | 编译后的可执行文件 |
| 前端 Dist | `dist` | 文件上传 | Vue/React 构建产物 (dist.tar.gz) |
| Python Wheel | `wheel` | 文件上传 | pip 安装包 |
| **Docker 镜像** | `image` | **仅记录引用**（不存储文件） | 镜像 repo + tag + digest |
| 通用压缩包 | `archive` | 文件上传 | 其他格式 |

### 4.2 数据库模型

```go
type CicdArtifact struct {
    ID           int64  // 主键
    PipelineID   int64  // 关联流水线
    RunID        int64  // 关联运行记录
    BuildNumber  int    // Jenkins 构建号

    // 制品信息
    Name         string // 制品名称（如 order-service-1.0.0.jar）
    ArtifactType string // 类型: jar/binary/dist/image/...
    Version      string // 版本号
    LanguageType string // 语言: go/java/frontend/python

    // 存储信息（文件类型）
    FilePath    string // 文件存储路径
    FileSize    int64  // 文件大小（字节）
    Sha256      string // SHA256 校验和
    StorageType string // 存储类型: local/s3/oss

    // 镜像信息（image 类型）
    ImageRepo   string // 镜像仓库地址
    ImageTag    string // 镜像 Tag
    ImageDigest string // 镜像 Digest (sha256:...)

    // Git 追溯
    GitRepo     string // Git 仓库地址
    GitBranch   string // 分支
    GitCommit   string // Commit SHA
}
```

### 4.3 Jenkins 中上传制品

Go 模板中的制品上传阶段（`configs/jenkins-templates/go-pipeline.groovy`）：

```groovy
stage('Upload Artifact') {
    when { expression { return params.ARTIFACT_UPLOAD_URL?.trim() && env.BINARY_PATH } }
    steps {
        script {
            sh """
                curl -s -X POST '${params.ARTIFACT_UPLOAD_URL}' \
                    -F 'file=@${env.BINARY_PATH}' \
                    -F 'pipeline_id=${params.PIPELINE_ID}' \
                    -F 'run_id=${params.RUN_ID}' \
                    -F 'build_number=${env.BUILD_NUMBER}' \
                    -F 'version=${env.FINAL_TAG}' \
                    -F 'language_type=go' \
                    -F 'artifact_type=binary' \
                    -F 'git_repo=${params.GIT_REPO}' \
                    -F 'git_branch=${env.GIT_BRANCH_NAME}' \
                    -F 'git_commit=${env.GIT_COMMIT_FULL}'
            """
        }
    }
}
```

---

## 五、自动部署机制

### 5.1 回调触发

Jenkins 构建完成后，通过 HTTP 回调通知平台：

```groovy
// Jenkins 回调 payload
def payload = [
    job_name          : env.JOB_NAME,
    build_number      : env.BUILD_NUMBER,
    status            : 'SUCCESS',
    pipeline_id       : params.PIPELINE_ID,
    image_url         : env.FULL_IMAGE,          // 完整镜像地址 + tag
    image_digest      : env.IMAGE_DIGEST,         // sha256:...
    image_with_digest : env.IMAGE_WITH_DIGEST,    // 镜像地址 + @sha256:...
    git_commit        : env.GIT_COMMIT_SHORT,
    git_branch        : env.GIT_BRANCH_NAME,
    duration_sec      : (currentBuild.duration / 1000),
    message           : '构建成功'
]
```

### 5.2 平台处理回调

```
收到回调
  │
  ├── 1. 查找关联流水线（pipeline_id 或 job_name）
  ├── 2. 更新运行记录状态（success/failed）
  ├── 3. 检查是否启用自动部署（auto_deploy = true）
  │       │
  │       ├── 未启用 → 结束
  │       ├── 配置不完整（namespace/workload/container 缺失）→ 跳过
  │       └── 启用 ↓
  │
  ├── 4. 检查是否需要审批（require_approval = true）
  │       │
  │       ├── 需要 → 创建审批记录，等待人工确认
  │       └── 不需要 ↓
  │
  ├── 5. 获取目标集群 K8s 客户端
  │       ├── 多集群 → 根据 target_cluster_id 初始化
  │       └── 单集群 → 使用默认管理集群
  │
  ├── 6. 构造最终镜像地址
  │       └── 优先使用 image@sha256:digest（确保一致性）
  │
  └── 7. 执行部署
          └── PatchDeploymentImage(namespace, name, container, finalImage)
              → 等待 Rollout 完成
              → 发送钉钉通知
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

实现代码：
```go
// 构造最终镜像地址（优先使用 image@digest 确保一致性）
finalImage := image
if imageDigest != "" {
    if idx := strings.LastIndex(image, ":"); idx > 0 && !strings.Contains(image[idx:], "/") {
        finalImage = image[:idx] + "@" + imageDigest
    } else {
        finalImage = image + "@" + imageDigest
    }
}
```

---

## 六、Jenkins 模板一览

平台内置 4 套语言模板（`configs/jenkins-templates/`）：

| 模板 | 文件 | 特有阶段 |
|------|------|---------|
| **Go 项目** | `go-pipeline.groovy` | 编译产物(build_binary) + 制品上传(upload_artifact) + 自动生成 Dockerfile |
| **Java Spring** | `java-spring-pipeline.groovy` | Maven 编译 + Spring Boot JAR |
| **前端 (Vue/React)** | `frontend-pipeline.groovy` | npm build + Nginx 镜像 |
| **Python** | `python-pipeline.groovy` | pip install + pytest |

所有模板共享相同的：
- **参数协议**：`GIT_REPO`, `IMAGE_REPO`, `IMAGE_TAG`, `PIPELINE_ID`, `PLATFORM_CALLBACK_URL`
- **回调协议**：阶段级 `stageCallback()` + 最终 `callbackPlatform()`
- **安全机制**：HMAC-SHA256 签名验证
- **构建工具**：nerdctl (containerd 生态，替代 docker CLI)

---

## 七、典型发布示例

以 `springboot-hello` 项目为例：

```
1. 平台创建流水线
   ├── Git 仓库:    gitee.com/xxx/springboot-hello
   ├── 镜像仓库:    registry.cn-hangzhou.aliyuncs.com/k8s-gos/springboot-hello
   ├── 构建模板:    java-spring-pipeline
   ├── 部署目标:    Deployment default/hello-springboot
   └── 目标容器:    springboot-hello

2. 触发构建 → Jenkins 执行
   ├── Checkout:    git clone + branch=main
   ├── Compile:     mvn compile
   ├── Test:        mvn test
   ├── Build:       nerdctl build → registry...aliyuncs.com/.../springboot-hello:main-abc123-20260422
   └── Push:        nerdctl push → 推到阿里云 ACR

3. Jenkins 回调平台
   ├── image_url:    registry.cn-hangzhou.aliyuncs.com/k8s-gos/springboot-hello:main-abc123-20260422
   ├── image_digest: sha256:a1b2c3d4...
   └── status:       SUCCESS

4. 平台自动部署
   ├── 构造镜像:    registry...aliyuncs.com/.../springboot-hello@sha256:a1b2c3d4...
   ├── Patch:       PatchDeploymentImage(default, hello-springboot, springboot-hello, image)
   ├── Rollout:     等待滚动更新完成
   └── 通知:        钉钉推送部署结果
```

---

## 八、私有仓库 vs 阿里云仓库

### 8.1 选型对比

| 维度 | Harbor (私有部署) | 阿里云 ACR |
|------|-----------------|-----------|
| **部署方式** | 自建（需服务器） | 云服务（开箱即用） |
| **成本** | 服务器 + 运维成本 | 按量计费，个人版免费 |
| **带宽** | 内网传输速度快 | 公网/VPC 看网络配置 |
| **安全** | 完全可控 | 阿里云安全体系 |
| **镜像扫描** | 需额外配置 Trivy/Clair | 内置安全扫描 |
| **适用场景** | 大规模企业、数据合规要求 | 中小团队、快速上手 |
| **集成方式** | Jenkins 凭据 `harbor-registry` | Jenkins 凭据（AccessKey 或 RAM） |

### 8.2 当前系统如何切换

Jenkins 模板中镜像仓库地址完全通过参数传入，**切换仓库只需修改流水线配置中的 `IMAGE_REPO`**：

```
# 使用阿里云 ACR
IMAGE_REPO = registry.cn-hangzhou.aliyuncs.com/k8s-gos/springboot-hello

# 切换到 Harbor
IMAGE_REPO = harbor.mycompany.com/prod/springboot-hello
```

Jenkins 凭据 `harbor-registry` 需要在 Jenkins 中配置对应仓库的用户名/密码。

---

## 九、关键文件索引

| 文件 | 说明 |
|------|------|
| `configs/jenkins-templates/go-pipeline.groovy` | Go 项目构建模板（524 行） |
| `configs/jenkins-templates/java-spring-pipeline.groovy` | Java Spring 构建模板 |
| `configs/jenkins-templates/frontend-pipeline.groovy` | 前端构建模板 |
| `configs/jenkins-templates/python-pipeline.groovy` | Python 构建模板 |
| `internal/app/services/cicd_pipeline.go` | 流水线核心逻辑（回调处理 + 自动部署） |
| `internal/app/services/cicd_stage.go` | 阶段执行逻辑（部署阶段 + Rollout 监控） |
| `internal/app/services/cicd_artifact.go` | 制品库服务（上传/下载/记录） |
| `internal/app/models/cicd_artifact.go` | 制品数据模型 |
| `internal/app/models/image_registry.go` | 镜像仓库数据模型 |
| `internal/app/controllers/api/v1/image/image_registry.go` | 镜像仓库 API |
| `internal/app/controllers/api/v1/cicd/artifact_controller.go` | 制品库 API |
| `pkg/k8s/deployment/patch_deploy.go` | Deployment 镜像 Patch（部署执行层） |
| `docs/sql/cicd_artifact.sql` | 制品库建表 SQL |
| `docs/sql/k8s_platform_full_init.sql` | 全量初始化 SQL（含 image_registry） |
