# 模板化 CI/CD 快速接入指南

> 本文档详细说明如何通过 K8sOperation 平台，以**模板化方式**快速接入一个新项目的 CI/CD 流水线。
> 以 Java Spring Boot 项目为例，全流程 **10 分钟**完成接入，无需编写 Jenkinsfile。

---

## 一、架构总览

### 1.1 核心设计理念

```
4 个通用 Jenkins Job → 服务 100+ 项目
```

| 语言类型 | Jenkins Job 名称 | 模板文件（平台侧） | 项目侧需要什么 |
|---------|-----------------|-------------------|---------------|
| Java | `k8s-builder-java` | `configs/jenkins-templates/java-spring-pipeline.groovy` | Dockerfile + pom.xml |
| Go | `k8s-builder-go` | `configs/jenkins-templates/go-pipeline.groovy` | Dockerfile + go.mod |
| Frontend | `k8s-builder-frontend` | `configs/jenkins-templates/frontend-pipeline.groovy` | Dockerfile + package.json |
| Python | `k8s-builder-python` | `configs/jenkins-templates/python-pipeline.groovy` | Dockerfile + requirements.txt |

### 1.2 模板存放位置

```
K8sOperation 平台仓库（本仓库）
├── configs/jenkins-templates/          ← Groovy 模板文件（Jenkins Job 使用）
│   ├── java-spring-pipeline.groovy     ← Java 通用模板
│   ├── go-pipeline.groovy              ← Go 通用模板
│   ├── frontend-pipeline.groovy        ← 前端通用模板
│   └── python-pipeline.groovy          ← Python 通用模板
├── docs/dockerfile/                    ← 参考 Dockerfile
│   ├── Dockerfile.java-maven           ← Java 多阶段构建
│   ├── Dockerfile.golang               ← Go 多阶段构建
│   ├── Dockerfile.nginx                ← Nginx 前端
│   └── Dockerfile.python               ← Python
```

**关键点：模板文件不需要放在业务项目中！**

- ✅ Groovy 模板放在 **平台仓库** 的 `configs/jenkins-templates/` 下
- ✅ Jenkins Job 使用 **Pipeline script from SCM** 指向平台仓库（推荐）
- ✅ 业务项目只需要 **Dockerfile** 和 **构建配置文件**（pom.xml/go.mod 等）
- ❌ 业务项目 **不需要** Jenkinsfile

### 1.3 参数传递流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                       模板化发布参数流向                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  开发者在平台创建流水线                                                   │
│    ├── language_type: "java"              ← 自动映射 k8s-builder-java    │
│    ├── git_repo: 业务项目仓库地址                                        │
│    ├── git_branch: "main"                                               │
│    └── 部署配置: namespace/deployment/container                          │
│                                                                         │
│  平台触发 Jenkins 构建                                                   │
│    ├── GIT_REPO ──────────────┐                                         │
│    ├── GIT_BRANCH ────────────┤                                         │
│    ├── IMAGE_REPO ────────────┼────► Jenkins Job Parameters             │
│    ├── IMAGE_TAG  ────────────┤     （由通用模板接收）                     │
│    ├── PIPELINE_ID ───────────┤                                         │
│    └── PLATFORM_CALLBACK_URL ─┘                                         │
│                                                                         │
│  Jenkins 执行通用模板                                                    │
│    ├── 1. Checkout（拉取业务项目代码）                                    │
│    ├── 2. Compile（mvn compile）                                        │
│    ├── 3. Test（mvn test，可跳过）                                       │
│    ├── 4. SonarQube（代码质量扫描，可选）                                 │
│    ├── 5. Quality Gate（质量门禁，可选）                                  │
│    ├── 6. Package（mvn package）                                        │
│    ├── 7. Build Image（nerdctl build）                                  │
│    ├── 8. Push Image（推送到 Harbor）                                    │
│    └── 9. Callback（回调平台）                                           │
│                                                                         │
│  平台收到回调后                                                          │
│    ├── 更新流水线状态                                                    │
│    ├── 记录镜像信息                                                      │
│    └── 自动部署到 K8s（如果配置了 auto_deploy）                           │
│         └── client-go StrategicMergePatch → 触发滚动更新                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 二、一键接入 - Java Spring Boot 项目

### 2.1 前提条件

| 条件 | 说明 |
|------|------|
| Jenkins | 已部署，已安装 Pipeline 插件 |
| Harbor | 已部署，已创建镜像仓库 |
| K8sOperation 平台 | 已部署运行 |
| Jenkins 凭证 | 已配置 `harbor-registry`（Harbor 账号）、`gitee-id`（Git 账号）、`hmac-secret`（回调签名） |
| 目标集群 | 已在平台添加 K8s 集群 |

### 2.2 Step 1：Jenkins 创建通用 Job（只需做一次）

> **这一步整个平台只需要做一次**，创建后所有同类型项目共用这个 Job，无需重复配置。

#### 2.2.1 详细操作步骤（以 Java 为例）

**① 新建 Pipeline Job**

1. 登录 Jenkins 管理界面（如 `http://your-jenkins:8080`）
2. 点击左侧菜单 **新建任务（New Item）**
3. 输入任务名称：**`k8s-builder-java`**（必须与平台映射名一致）
4. 选择类型：**Pipeline**
5. 点击 **确定（OK）**

**② 配置 Pipeline script from SCM（推荐）**

进入 Job 配置页面，拉到底部 **Pipeline** 区域：

| 序号 | 配置项 | 操作 | 说明 |
|------|--------|------|------|
| 1 | Definition | 下拉选择 **Pipeline script from SCM** | 模板从 Git 仓库自动拉取，而非手动粘贴 |
| 2 | SCM | 选择 **Git** | 版本控制类型 |
| 3 | Repository URL | 填写 **平台仓库地址** | 如 `https://gitee.com/your-org/k8s_operation.git` |
| 4 | Credentials | 选择 **gitee-id** | 提前在 Jenkins 凭证管理中添加的 Git 账号凭证 |
| 5 | Branch Specifier | 填写 `*/main` | 模板所在分支（根据实际情况修改） |
| 6 | Script Path | 填写模板路径（见下表） | **关键配置**：指向仓库中的 Groovy 模板文件 |
| 7 | 轻量级checkout | 可选勾选 | 加速拉取，只获取 Groovy 文件 |

**③ 点击保存**

> 💡 **重要**：保存后 Jenkins 会自动从 Git 仓库拉取模板文件。模板中已定义 `parameters` 块，**参数化构建会自动生效**，无需手动添加任何参数。

#### 2.2.2 四种语言 Job 完整配置表

每种语言创建一个 Job，**整个平台共 4 个 Job 即可服务 100+ 项目**：

| 语言 | Job 名称（必须一致） | Script Path | 附加说明 |
|------|---------------------|-------------|----------|
| **Java** | `k8s-builder-java` | `configs/jenkins-templates/java-spring-pipeline.groovy` | 含 SonarQube 代码扫描 + Quality Gate 质量门禁 |
| **Go** | `k8s-builder-go` | `configs/jenkins-templates/go-pipeline.groovy` | 含 golangci-lint 代码检查 |
| **Frontend** | `k8s-builder-frontend` | `configs/jenkins-templates/frontend-pipeline.groovy` | 支持 Vue/React/Angular/Next.js |
| **Python** | `k8s-builder-python` | `configs/jenkins-templates/python-pipeline.groovy` | 含 flake8 + pytest 检查 |

> 以上 4 个 Job 的 Repository URL、Credentials、Branch Specifier **完全相同**，只有 **Script Path** 不同。

#### 2.2.3 Jenkins 必须预配置的凭证（Credentials）

在 Jenkins → **Manage Jenkins → Credentials → System → Global credentials** 中添加：

| 凭证 ID | 类型 | 用途 | 示例 |
|---------|------|------|------|
| `harbor-registry` | Username with password | Harbor 镜像仓库登录 | 用户名: `admin`，密码: `Harbor12345` |
| `gitee-id` | Username with password | Git 仓库拉取代码 | 用户名: `git-user`，密码/Token: `xxx` |
| `hmac-secret` | Secret text | 回调签名密钥 | 与 `config.yaml` 中 `jenkins.hmac_secret` 一致 |

#### 2.2.4 Jenkins 必须安装的插件

| 插件名 | 用途 | Java 必需 |
|--------|------|-----------|
| Pipeline | Pipeline 基础支持 | ✅ |
| Git | Git SCM 支持 | ✅ |
| HTTP Request | 阶段/最终回调平台 | ✅ |
| SonarQube Scanner | 代码质量扫描 | ✅（Java 必需） |

> 如果是 Java 项目，还需在 Jenkins → **Manage Jenkins → System → SonarQube servers** 中配置名为 `SonarQube` 的服务器连接。

#### 2.2.5 SCM 模式 vs 手动粘贴对比

| 对比项 | Pipeline script from SCM（推荐） | Pipeline script（手动粘贴） |
|--------|--------------------------------|---------------------------|
| 模板更新 | Git 提交即生效，Jenkins 自动拉取最新 | 需手动登录 Jenkins 逐个 Job 粘贴 |
| 版本管理 | 有完整 Git 历史，可回滚 | 无版本记录 |
| 多人协作 | Code Review + PR 流程 | 无法追踪谁改了什么 |
| 维护成本 | 极低（改一处生效所有 Job） | 高（N 个 Job 需改 N 次） |
| 初始配置 | 稍复杂（需配置 SCM） | 简单（直接粘贴） |

> **结论**：除调试阶段外，生产环境**必须使用 SCM 模式**。

### 2.3 Step 2：业务项目准备（每个新项目做一次）

业务项目只需要在**根目录**放一个 `Dockerfile`：

```
my-java-app/                ← 你的 Java 项目
├── Dockerfile              ← 唯一需要添加的文件
├── pom.xml
└── src/
    └── main/java/...
```

**推荐 Dockerfile（多阶段构建）：**

```dockerfile
# ============ 构建阶段 ============
FROM maven:3.9-eclipse-temurin-17 AS builder
WORKDIR /build

# 先复制 pom.xml 利用 Docker 缓存
COPY pom.xml .
RUN mvn dependency:go-offline -B

# 复制源码并构建
COPY src ./src
RUN mvn package -DskipTests -B && mv target/*.jar app.jar

# ============ 运行阶段 ============
FROM eclipse-temurin:17-jre-alpine

# 时区设置
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /build/app.jar .
USER appuser

EXPOSE 8080

# JVM 参数（可通过 K8s 环境变量覆盖）
ENV JAVA_OPTS="-Xms256m -Xmx512m -XX:+UseG1GC -XX:+HeapDumpOnOutOfMemoryError"

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
    CMD wget -qO- http://localhost:8080/actuator/health || exit 1

ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

**简化版 Dockerfile（如果 Jenkins 模板已经编译）：**

```dockerfile
FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY target/*.jar app.jar
EXPOSE 8080
ENV JAVA_OPTS="-Xms256m -Xmx512m"
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

> 💡 **模板的 Package 阶段已经执行了 `mvn package`**，所以简化版可以直接 COPY target/*.jar。
> 如果用多阶段构建的完整版 Dockerfile，构建会更独立、更可靠。

### 2.4 Step 3：平台创建流水线

> ℹ️ **快速提示**：创建流水线只有 **2 个必填字段**（name + git_repo），其余全部可选。详见下方「字段说明」。

#### 方式一：前端界面操作（推荐）

前端提供了 **5 步向导式** 创建界面（`PipelineCreate.vue`），每步只关注一类配置：

**① Step 1：基本信息**

| 字段 | 必填 | 示例 | 说明 |
|------|------|------|------|
| 流水线名称 | ✅ 必填 | `order-service-prod` | 全局唯一，建议格式：`{项目名}-{环境}` |
| 描述 | 可选 | `订单服务生产环境流水线` | 纯展示用 |

**② Step 2：代码仓库**

| 字段 | 必填 | 示例 | 说明 |
|------|------|------|------|
| Git 仓库地址 | ✅ 必填 | `https://gitee.com/org/order-service.git` | 业务项目仓库（非平台仓库） |
| Git 分支 | ✅ 必填 | `main` | 可点击「获取分支」按钮自动拉取分支列表 |

> 💡 点击「获取分支」按钮后，会自动调用后端接口获取远程分支列表，可直接下拉选择。

**③ Step 3：Jenkins 配置**（核心步骤）

| 字段 | 必填 | 示例 | 说明 |
|------|------|------|------|
| Jenkins 服务器地址 | 可选 | 留空 | 留空则使用 `config.yaml` 中的全局配置 |
| Jenkins Job 名称 | **留空即可** | 留空 | 后端根据语言类型自动推导为 `k8s-builder-java` |
| 服务类型（语言） | 推荐选择 | `Java` | **必须选择正确的语言类型**，关系到 Jenkins Job 自动映射 |
| SonarQube 扫描 | 可选 | ✅ 开启 | Java 项目选择后自动开启，标记「Java 推荐」 |
| 环境变量 | 可选 | `JAVA_VERSION=17` | 可展开添加自定义变量，覆盖自动注入的默认值 |

> **自动映射规则**：选择语言类型后，后端自动从 `DefaultJenkinsJobMap` 映射：
> | 选择的语言类型 | 自动映射的 Jenkins Job | Jenkins 使用的模板（SCM 拉取） |
> |--------------|----------------------|-----------------------------|
> | Java | `k8s-builder-java` | `java-spring-pipeline.groovy` |
> | Go | `k8s-builder-go` | `go-pipeline.groovy` |
> | Frontend | `k8s-builder-frontend` | `frontend-pipeline.groovy` |
> | Python | `k8s-builder-python` | `python-pipeline.groovy` |
> | Custom | **必须手动填写** | 自定义 |

> 💡 **Java 项目特有**：选择 Java 后，SonarQube 开关自动开启，并自动注入以下参数：
> `JAVA_VERSION=17`、`MAVEN_GOALS`、`ENABLE_SONAR=true`、`SONAR_QUALITY_GATE=true`、`SONAR_SOURCES`、`SONAR_JAVA_BINARIES` 等

**④ Step 4：部署策略**

| 字段 | 必填 | 示例 | 说明 |
|------|------|------|------|
| 部署环境 | 可选 | `dev` / `test` / `staging` / `prod` | 点击环境芯片选择，prod 自动开启审批 |
| 副本数 | 可选 | `3` | 通过 +/- 按钮调整 |
| 部署策略 | 可选 | `滚动更新` / `重建` | 默认滚动更新 |
| 资源配置 | 可选 | CPU: 500m / Memory: 512Mi | 可展开调整资源限制和请求 |

**⑤ Step 5：自动部署配置**

| 字段 | 必填 | 示例 | 说明 |
|------|------|------|------|
| 启用自动部署 | 可选 | ✅ 开启 | 开启后构建成功自动更新 K8s 镜像 |
| 目标集群 | 条件必填 | 下拉选择 | 自动加载已添加的集群列表（权限过滤） |
| 目标命名空间 | 条件必填 | 下拉选择 | 选择集群后自动加载命名空间列表 |
| 工作负载类型 | 可选 | `Deployment` | 可选 Deployment/StatefulSet/DaemonSet |
| 工作负载名称 | 条件必填 | 下拉选择 | 选择命名空间后自动加载工作负载列表 |
| 容器名称 | 可选 | 留空 | 留空则更新第一个容器 |
| 需要审批 | 可选 | ✅ 开启 | 生产环境建议开启 |

> 💡 **级联加载**：前端实现了智能级联 —— 选择集群后自动加载命名空间，选择命名空间后自动加载工作负载，无需手动输入。

**⑥ 点击“创建流水线”提交**

提交时前端会：
- 自动将 SonarQube 开关状态同步到 `env_vars`（`ENABLE_SONAR=true/false`）
- 自动清理内部字段（`enable_sonar`），只提交后端需要的数据
- 后端收到后自动推导 `jenkins_job`、自动注入语言特定参数

#### 前端 5 步向导与后端处理流程

```
前端 PipelineCreate.vue（5 步向导）
    │
    │  Step 1: 基本信息（name + description）
    │  Step 2: 代码仓库（git_repo + git_branch）
    │  Step 3: Jenkins（language_type + enable_sonar + env_vars）
    │  Step 4: 部署策略（deploy_env + replicas + resources）
    │  Step 5: 自动部署（auto_deploy + cluster + namespace + workload）
    │
    │  点击「创建流水线」
    │
    ▼  POST /api/v1/k8s/cicd/pipeline/create

后端 PipelineCreate Service
    │
    ├── ValidPipelineCreateRequest 校验
    │      name: required, between:1,100
    │      git_repo: required, url
    │      language_type: in:go,java,frontend,python,custom
    │
    ├── 检查名称唯一性
    │
    ├── 模板化推导：
    │      language_type="java" + jenkins_job=""
    │      → jenkins_job = DefaultJenkinsJobMap["java"] = "k8s-builder-java"
    │
    ├── injectLanguageParams() 自动注入语言参数
    │      JAVA_VERSION=17, MAVEN_GOALS=..., ENABLE_SONAR=true ...
    │
    └── 写入数据库 → 返回 pipeline_id
```

#### 方式二：API 调用

```bash
curl -X POST http://platform:8080/api/v1/k8s/cicd/pipeline/create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-java-app-deploy",
    "language_type": "java",
    "git_repo": "https://gitee.com/org/my-java-app.git",
    "git_branch": "main",
    "auto_deploy": true,
    "target_cluster_id": 1,
    "target_namespace": "production",
    "target_workload_kind": "Deployment",
    "target_workload_name": "my-java-app",
    "target_container": "my-java-app",
    "deploy_env": "production",
    "require_approval": true,
    "env_vars": [
      {"name": "IMAGE_REPO", "value": "harbor.example.com/library/my-java-app"}
    ]
  }'
```

**注意**：`language_type` 设为 `"java"` 后，**不需要填 `jenkins_job`**，平台自动映射为 `k8s-builder-java`。

### 2.5 创建流水线字段说明（必填 vs 选填）

#### 必填字段

| 字段 | JSON 键 | 校验规则 | 说明 |
|------|---------|---------|------|
| 流水线名称 | `name` | `required`, `between:1,100` | 唯一标识，不可重复 |
| Git 仓库地址 | `git_repo` | `required`, `url` | 业务项目仓库 HTTPS 地址 |

#### 选填字段（有默认值 — 模板/代码自动提供）

| 字段 | JSON 键 | 默认值 | 选填原因 |
|------|---------|--------|----------|
| Git 分支 | `git_branch` | `"main"` | `NewPipelineCreateRequest()` 构造函数设置默认值 |
| 语言类型 | `language_type` | `"custom"` | 构造函数设置默认值；选择具体语言后自动推导 Jenkins Job |
| Jenkins Job | `jenkins_job` | 由 `language_type` 自动映射 | **模板化核心**：选择 `java` → 自动映射 `k8s-builder-java`，无需手动填写 |
| Jenkins URL | `jenkins_url` | 使用全局 `config.yaml` 中的配置 | 仅在需要覆盖全局 Jenkins 地址时才填 |

> **为什么 `jenkins_job` 是选填？**
>
> 这是模板化架构的核心设计。后端 `PipelineCreate` 服务中有自动推导逻辑：
> ```go
> // internal/app/services/cicd_pipeline.go
> if jenkinsJob == "" && languageType != models.LanguageTypeCustom {
>     if job, ok := models.DefaultJenkinsJobMap[languageType]; ok {
>         jenkinsJob = job  // java → k8s-builder-java
>     }
> }
> ```
> 映射表定义在 `internal/app/models/cicd_pipeline.go`：
> ```go
> var DefaultJenkinsJobMap = map[string]string{
>     "go":       "k8s-builder-go",
>     "java":     "k8s-builder-java",
>     "frontend": "k8s-builder-frontend",
>     "python":   "k8s-builder-python",
> }
> ```
> 只有 `language_type: "custom"` 时才需要手动指定 `jenkins_job`。

#### 选填字段（部署配置 — 不影响构建）

| 字段 | JSON 键 | 默认行为 | 选填原因 |
|------|---------|---------|----------|
| 描述 | `description` | 空 | 纯展示信息，不影响功能 |
| 自动部署 | `auto_deploy` | `false` | 默认只构建不部署，按需开启 |
| 目标集群 | `target_cluster_id` | 无 | 仅 `auto_deploy=true` 时需要 |
| 目标命名空间 | `target_namespace` | 无 | 仅 `auto_deploy=true` 时需要 |
| 工作负载类型 | `target_workload_kind` | `"Deployment"` | 代码中默认值：`if workloadKind == "" { workloadKind = "Deployment" }` |
| 工作负载名称 | `target_workload_name` | 无 | 仅 `auto_deploy=true` 时需要 |
| 容器名称 | `target_container` | 无 | 仅 `auto_deploy=true` 时需要 |
| 部署环境 | `deploy_env` | 无 | 用于审批判断和通知展示 |
| 需要审批 | `require_approval` | `false` | 生产环境建议开启 |
| 环境变量 | `env_vars` | `[]` | Jenkins 构建参数，按需覆盖 |
| 部署配置(旧) | `deploy_config` | `{}` | 兼容旧版 JSON 格式，新项目无需使用 |

#### 创建流水线最简示例

**纯构建（不自动部署）— 只需 2 个必填字段 + 1 个推荐字段：**
```json
{
  "name": "my-java-app-build",
  "git_repo": "https://gitee.com/org/my-java-app.git",
  "language_type": "java"
}
```

**构建 + 自动部署（完整配置）：**
```json
{
  "name": "my-java-app-deploy",
  "git_repo": "https://gitee.com/org/my-java-app.git",
  "language_type": "java",
  "auto_deploy": true,
  "target_cluster_id": 1,
  "target_namespace": "production",
  "target_workload_name": "my-java-app",
  "target_container": "my-java-app",
  "deploy_env": "production",
  "require_approval": true
}
```

### 2.6 批量创建流水线（100+ 项目场景）

> 当需要接入大量微服务（50~200 个）时，逐个在前端创建流水线效率低下。平台提供了 **批量创建 API** + **PowerShell 导入脚本**，支持一次性导入所有项目。

#### 方式一：PowerShell 脚本批量导入（推荐）

**① 编辑导入配置文件**

编辑 `scripts/batch-import-pipelines.json`，按模板格式填写所有项目信息：

```json
{
  "skip_existing": true,
  "pipelines": [
    {
      "name": "order-service",
      "description": "订单服务",
      "git_repo": "https://gitee.com/your-org/order-service.git",
      "git_branch": "main",
      "language_type": "java",
      "auto_deploy": true,
      "target_cluster_id": 1,
      "target_namespace": "production",
      "target_workload_name": "order-service",
      "deploy_env": "prod",
      "require_approval": true
    },
    {
      "name": "payment-service",
      "git_repo": "https://gitee.com/your-org/payment-service.git",
      "language_type": "java"
    }
  ]
}
```

> 💡 `skip_existing: true` 表示已存在同名流水线时自动跳过，不会报错。每条记录最少只需 `name` + `git_repo` + `language_type` 三个字段。

**② 运行导入脚本**

```powershell
# 默认参数（localhost:8080，admin/123456）
.\scripts\batch-import-pipelines.ps1

# 指定平台地址和账号
.\scripts\batch-import-pipelines.ps1 -ApiUrl http://192.168.1.100:8080 -Username admin -Password your_password

# 使用已有 Token
.\scripts\batch-import-pipelines.ps1 -Token "eyJhbGciOiJIUz..."
```

**③ 查看执行结果**

```
============================================
  批量创建结果
============================================
  成功: 3
  失败: 0
  跳过: 2

  详细结果:
  -----------------------------------------------
  [OK]   order-service -> pipeline_id=15
  [OK]   user-service -> pipeline_id=16
  [OK]   admin-web -> pipeline_id=17
  [SKIP] data-processor
  [SKIP] payment-service
============================================
```

#### 方式二：直接调用批量创建 API

```bash
curl -X POST http://platform:8080/api/v1/k8s/cicd/pipeline/batch-create \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "skip_existing": true,
    "pipelines": [
      {"name": "svc-a", "git_repo": "https://gitee.com/org/svc-a.git", "language_type": "java"},
      {"name": "svc-b", "git_repo": "https://gitee.com/org/svc-b.git", "language_type": "go"},
      {"name": "svc-c", "git_repo": "https://gitee.com/org/svc-c.git", "language_type": "frontend"}
    ]
  }'
```

#### 批量创建 API 说明

| 参数 | 类型 | 说明 |
|------|------|------|
| `skip_existing` | bool | `true` 时跳过已存在同名流水线，`false` 时报错 |
| `pipelines` | array | 流水线数组，单次最多 200 条 |
| `pipelines[].name` | string | **必填** — 流水线名称（全局唯一） |
| `pipelines[].git_repo` | string | **必填** — Git 仓库地址 |
| `pipelines[].language_type` | string | 推荐 — `java`/`go`/`frontend`/`python`/`custom` |
| 其他字段 | — | 与单条创建 API 相同（git_branch、auto_deploy、deploy_env 等） |

> **返回字段**：`success_count`（成功数）、`fail_count`（失败数）、`skip_count`（跳过数）、`results`（每条详细状态）

#### 批量导入最佳实践

| 场景 | 建议 |
|------|------|
| 首次迁移 100+ 项目 | 从 Git 平台导出项目列表 → 用脚本/Excel 生成 JSON → 运行导入脚本 |
| 新增环境（dev → staging） | 复制 JSON 文件，修改 `deploy_env` 和 `target_namespace` → 重新导入 |
| 增量添加 | 设置 `skip_existing: true`，新项目追加到 JSON 末尾即可 |
| 不同团队/命名空间 | 按团队拆分多个 JSON 文件，分别执行 |

### 2.7 Step 4：触发构建

#### 前端操作
1. 进入流水线详情页 → 点击 **运行**
2. 可选覆盖参数（分支、镜像标签等）
3. 点击确认

#### API 触发
```bash
curl -X POST http://platform:8080/api/v1/k8s/cicd/pipeline/run \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "branch": "main",
    "env_vars": {
      "IMAGE_TAG": "v1.0.0",
      "SKIP_TESTS": "false",
      "ENABLE_SONAR": "true"
    }
  }'
```

### 2.8 Step 5：构建过程（全自动）

流水线执行过程完全自动化，每个阶段完成后实时回调平台：

```
[1] Clean Workspace  → 清理 + 拉取业务项目代码
[2] Checkout Info    → 提取 Git commit、生成镜像标签     → 回调 stage/callback
[3] Compile          → mvn clean compile                → 回调 stage/callback
[4] Test             → mvn test（可跳过）                → 回调 stage/callback
[5] SonarQube        → 代码质量扫描（可选）               → 回调 stage/callback
[6] Quality Gate     → 质量门禁检查（可选）               → 回调 stage/callback
[7] Package          → mvn package，生成 JAR             → 回调 stage/callback
[8] Build Image      → nerdctl build -t <image>         → 回调 stage/callback
[9] Push Image       → nerdctl push 到 Harbor           → 回调 stage/callback
[最终] Callback       → 最终状态回调 pipeline/callback
```

### 2.9 Step 6：自动部署 — 基于 client-go StrategicMergePatch（如果配置了 auto_deploy）

#### 滚动更新实现方式

> **推荐方式：client-go StrategicMergePatch**（本平台已采用）
>
> | 方式 | 原理 | 优缺点 |
> |------|------|--------|
> | ❌ kubectl set image | Shell 执行 kubectl 命令行 | 依赖 kubectl 二进制、难以获取返回值、安全风险（命令注入） |
> | ❌ client-go Update | 获取完整对象→修改→PUT 回去 | 有并发冲突风险（ResourceVersion 过期）、传输数据量大 |
> | ✅ **client-go StrategicMergePatch** | 只发送需要修改的字段差异 | **原子操作、无并发冲突、网络开销最小、生产级最佳实践** |

#### 实际代码实现

平台使用 **3 层架构** 实现滚动更新：

**第 1 层：Patch Builder（构造 JSON Patch）**
```go
// pkg/k8s/deployment/patchbuilder/patch_builder.go
func BuildImagePatch(containerName, image string) ([]byte, error) {
    patch := map[string]interface{}{
        "spec": map[string]interface{}{
            "template": map[string]interface{}{
                "spec": map[string]interface{}{
                    "containers": []map[string]string{{
                        "name":  containerName,
                        "image": image,
                    }},
                },
            },
        },
    }
    return json.Marshal(patch)
}
```

**第 2 层：Patch Deployment（调用 K8s API）**
```go
// pkg/k8s/deployment/patch_deploy.go
func PatchDeployment(ctx context.Context, Kube kubernetes.Interface,
    namespace, name string, patch []byte) (*appv1.Deployment, error) {
    return Kube.AppsV1().Deployments(namespace).Patch(
        ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
}

func PatchDeploymentImage(ctx context.Context, Kube kubernetes.Interface,
    namespace, name, containerName, image string) (*appv1.Deployment, error) {
    patchImage, _ := patchbuilder.BuildImagePatch(containerName, image)
    return PatchDeployment(ctx, Kube, namespace, name, patchImage)
}
```

**第 3 层：CI/CD 自动部署（异步执行 + 等待 Rollout）**
```go
// internal/app/services/cicd_pipeline.go → executeAutoDeployAsync()
switch workloadKind {
case "Deployment":
    patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`,
        pipeline.TargetContainer, image)
    _, err = kubeClient.AppsV1().Deployments(pipeline.TargetNamespace).Patch(
        ctx, pipeline.TargetWorkloadName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
    if err == nil {
        err = s.waitAutoDeployRollout(ctx, kubeClient, ...) // 轮询等待 Rollout 完成
    }
case "StatefulSet":
    // 同样使用 StrategicMergePatch
case "DaemonSet":
    // 同样使用 StrategicMergePatch
}
```

#### 滚动更新完整流程

构建成功后，平台自动执行：

1. **Patch 更新镜像**：`client-go StrategicMergePatch` 只更新 `spec.template.spec.containers[].image`
2. **K8s 控制器触发滚动更新**：Deployment Controller 检测到 PodTemplate 变化，按策略逐步替换旧 Pod
3. **平台轮询等待 Rollout 完成**：每 5 秒检查 `UpdatedReplicas/AvailableReplicas/ObservedGeneration`
4. **超时检测**：默认 5 分钟超时，检测 `ProgressDeadlineExceeded` 条件
5. **钉钉通知**：Rollout 完成/失败后自动发送通知
6. **审批流程**：生产环境如开启 `require_approval`，部署前需人工审批通过

#### 其他滚动更新操作（前端 Deployment 管理页面）

| 操作 | API | 实现方式 |
|------|-----|----------|
| 修改滚动策略 | `POST /deployment/update-strategy` | `client.AppsV1().Deployments().Update()` 修改 Strategy 字段 |
| 暂停滚动更新 | `POST /deployment/pause` | `StrategicMergePatch` 设置 `spec.paused=true` |
| 恢复滚动更新 | `POST /deployment/resume` | `StrategicMergePatch` 设置 `spec.paused=false` |
| 查看 Rollout 状态 | `GET /deployment/rollout-status` | 获取 Deployment + ReplicaSets 计算进度 |
| 滚动重启 | `POST /deployment/restart` | Patch 注解 `kubectl.kubernetes.io/restartedAt` |
| 修改副本数 | `POST /deployment/change_replicas` | `StrategicMergePatch` 修改 `spec.replicas` |
| 修改镜像 | `POST /deployment/change_image` | `PatchDeploymentImage()` → StrategicMergePatch |

### 2.10 Step 7：模板架构与滚动更新参数

滚动更新策略可通过平台 Deployment 管理页面调整：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `maxSurge` | `25%` | 滚动更新时最多超出的 Pod 数（数字或百分比） |
| `maxUnavailable` | `25%` | 滚动更新时最多不可用的 Pod 数 |
| `minReadySeconds` | `0` | Pod 就绪后最少等待秒数（防止假就绪） |
| `progressDeadlineSeconds` | `600` | Rollout 进度截止时间，超时则标记失败 |
| `revisionHistoryLimit` | `10` | 保留的历史 ReplicaSet 数量（用于回滚） |

---

## 三、其他语言快速接入

### 3.1 Go 项目

```
my-go-app/
├── Dockerfile        ← 唯一需要添加的
├── go.mod
├── go.sum
└── cmd/main.go
```

**Dockerfile：**
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

**平台创建**：`language_type: "go"` → 自动映射 `k8s-builder-go`

### 3.2 前端项目（Vue/React）

```
my-frontend/
├── Dockerfile        ← 唯一需要添加的
├── nginx.conf        ← Nginx 配置（可选，推荐）
├── package.json
└── src/
```

**Dockerfile：**
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
```

**平台创建**：`language_type: "frontend"` → 自动映射 `k8s-builder-frontend`

### 3.3 Python 项目

```
my-python-app/
├── Dockerfile        ← 唯一需要添加的
├── requirements.txt
└── main.py
```

**Dockerfile：**
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple
COPY . .
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

**平台创建**：`language_type: "python"` → 自动映射 `k8s-builder-python`

---

## 四、Jenkins 凭证配置（一次性）

在 Jenkins → 凭据 → 系统 → 全局凭据 中添加：

| 凭证 ID | 类型 | 用途 |
|---------|------|------|
| `harbor-registry` | Username with password | Harbor 镜像仓库登录（用户名 + 密码） |
| `gitee-id` | Username with password | Git 仓库拉取代码（用户名 + token） |
| `hmac-secret` | Secret text | 回调签名密钥（与平台 config.yaml 中的 `HMACSecret` 一致） |

**平台 config.yaml Jenkins 配置：**

```yaml
Jenkins:
  URL: "http://jenkins.example.com:8080/"
  Username: "admin"
  APIToken: "your-jenkins-api-token"
  TriggerTimeout: 60
  CallbackURL: "http://platform-backend:8080"    # 后端 API 地址（Jenkins 回调用）
  PlatformURL: "http://platform-frontend:30851"  # 前端页面地址（钉钉通知链接用）
  HMACSecret: "your-hmac-secret-key"
  PollInterval: 15
  MaxBuildTime: 30
  DingTalkWebhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
```

---

## 五、模板架构 Q&A

### Q1: Jenkinsfile/Groovy 模板放在哪里？

**放在平台仓库中**，不放在业务项目中：

```
K8sOperation/configs/jenkins-templates/
├── java-spring-pipeline.groovy     ← Jenkins Job 直接使用这个内容
├── go-pipeline.groovy
├── frontend-pipeline.groovy
└── python-pipeline.groovy
```

**为什么不放在业务项目根目录？**

| 方式 | 说明 | 优缺点 |
|------|------|--------|
| ❌ 每个项目放 Jenkinsfile | 传统方式，每个项目仓库都要维护一份 | 100 个项目 = 100 份维护 |
| ✅ 平台统一模板 | 4 个通用 Job，所有差异通过参数传入 | 修改模板一处生效，零维护 |

### Q2: 业务项目需要什么？

**只需要两样东西：**

1. **Dockerfile**（放在项目根目录）
2. **构建配置文件**（pom.xml / go.mod / package.json / requirements.txt）

不需要 Jenkinsfile、不需要 .jenkins 配置、不需要任何 CI 相关文件。

### Q3: 如何修改构建流程？

- **修改某个阶段的行为**：编辑 `configs/jenkins-templates/java-spring-pipeline.groovy`，所有 Java 项目立即生效
- **跳过单元测试**：触发时传参 `SKIP_TESTS=true`
- **跳过 SonarQube**：触发时传参 `ENABLE_SONAR=false`
- **自定义 Maven 命令**：触发时传参 `MAVEN_GOALS="clean package -Pprod -DskipTests -B"`

### Q4: 如何验证模板配置是否正确？

```bash
# 调用模板验证 API
curl http://platform:8080/api/v1/k8s/cicd/pipeline/template-verify \
  -H "Authorization: Bearer <token>"
```

返回所有语言模板的完整配置信息，包括 Jenkins Job 名称、阶段列表、默认参数、凭证检查等。

### Q5: 自定义语言/框架怎么办？

设置 `language_type: "custom"`，手动指定 `jenkins_job` 名称，自己维护 Jenkins Pipeline 脚本。

### Q6: 一个 Jenkins Job 同时构建 100 个项目不会冲突吗？

不会。每次触发时传入不同的 `GIT_REPO`、`IMAGE_REPO` 等参数，Jenkins Pipeline 是无状态的，每次构建独立。
Jenkins 通过 `disableConcurrentBuilds()` 防止同一 Job 并发，多个项目会排队执行。
如果需要并发，可以创建多个同名 Job 的副本（如 `k8s-builder-java-1`、`k8s-builder-java-2`）。

---

## 六、完整接入 Checklist

### 新项目接入（5 分钟）

- [ ] 业务项目根目录添加 `Dockerfile`
- [ ] 在 Harbor 创建镜像仓库（如 `harbor.example.com/library/my-java-app`）
- [ ] 在平台创建流水线（选择 `language_type=java`）
- [ ] 配置部署目标（集群、命名空间、Deployment、容器名）
- [ ] 触发一次构建验证闭环

### 首次平台搭建（30 分钟，一次性）

- [ ] 部署 K8sOperation 平台（后端 + 前端）
- [ ] 配置 config.yaml（数据库、Redis、Jenkins）
- [ ] 导入数据库（`docs/sql/k8s_platform_full_init.sql`）
- [ ] 在 Jenkins 创建 4 个通用 Job（Pipeline script from SCM 模式）
  - [ ] `k8s-builder-java` ← Script Path: `configs/jenkins-templates/java-spring-pipeline.groovy`
  - [ ] `k8s-builder-go` ← Script Path: `configs/jenkins-templates/go-pipeline.groovy`
  - [ ] `k8s-builder-frontend` ← Script Path: `configs/jenkins-templates/frontend-pipeline.groovy`
  - [ ] `k8s-builder-python` ← Script Path: `configs/jenkins-templates/python-pipeline.groovy`
- [ ] 在 Jenkins 配置 3 个凭证（harbor-registry、gitee-id、hmac-secret）
- [ ] 在平台添加 K8s 集群
- [ ] 验证模板配置：`GET /api/v1/k8s/cicd/pipeline/template-verify`

---

## 七、文件引用清单

| 文件路径 | 说明 |
|---------|------|
| `configs/jenkins-templates/java-spring-pipeline.groovy` | Java 通用构建模板（284 行） |
| `configs/jenkins-templates/go-pipeline.groovy` | Go 通用构建模板 |
| `configs/jenkins-templates/frontend-pipeline.groovy` | 前端通用构建模板 |
| `configs/jenkins-templates/python-pipeline.groovy` | Python 通用构建模板 |
| `docs/dockerfile/Dockerfile.java-maven` | Java Dockerfile 参考（多阶段构建） |
| `docs/dockerfile/Dockerfile.golang` | Go Dockerfile 参考 |
| `docs/dockerfile/Dockerfile.nginx` | Nginx Dockerfile 参考 |
| `docs/dockerfile/Dockerfile.python` | Python Dockerfile 参考 |
| `internal/app/models/cicd_pipeline.go` | 语言类型常量 + Job 映射表 |
| `internal/app/services/cicd_pipeline.go` | 流水线创建 + 模板自动推导 + 自动部署逻辑 |
| `internal/app/services/cicd_executor.go` | 发布单部署执行器（PatchDeploymentImage + waitRollout） |
| `internal/app/services/cicd_stage.go` | 阶段化部署执行器（多工作负载类型） |
| `pkg/k8s/deployment/patch_deploy.go` | Deployment Patch 底层封装（StrategicMergePatch） |
| `pkg/k8s/deployment/patchbuilder/patch_builder.go` | Patch JSON 构造器（镜像/副本数） |
| `pkg/k8s/deployment/rolling_update.go` | 滚动更新策略管理 + 暂停/恢复 + Rollout 状态查询 |
| `pkg/k8s/deployment/restart.go` | 滚动重启（Patch restartedAt 注解） |
| `internal/app/requests/cicd_pipeline.go` | 流水线创建请求模型 + 校验规则 |
| `configs/config.yaml.example` | Jenkins 配置示例 |
| `scripts/batch-import-pipelines.json` | 批量导入 JSON 模板（5 种语言示例） |
| `scripts/batch-import-pipelines.ps1` | PowerShell 批量导入脚本（自动登录 + 调用 API） |
| `docs/sql/k8s_platform_full_init.sql` | 数据库初始化脚本 |
