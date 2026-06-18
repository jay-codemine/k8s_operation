# Java 项目 CI/CD 完整接入指南

> **适用范围**: Java / Spring Boot / Spring Cloud / Maven 项目  
> **模板 Job**: `k8s-builder-java`  
> **Groovy 模板**: `configs/jenkins-templates/java-spring-pipeline.groovy`  
> **完整链路**: 前端创建 → 后端推导 → Jenkins 构建 → SonarQube 扫描 → 镜像推送 → 自动部署 K8s

---

## 一、架构总览

```
┌─────────────────────────────────────────────────────────────────────────┐
│  前端 PipelineCreate.vue（5 步向导）                                      │
│  ① 基本信息 → ② 代码仓库 → ③ Jenkins 配置 → ④ 部署策略 → ⑤ 自动部署     │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │ POST /api/v1/k8s/cicd/pipeline/create
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  后端 Go (PipelineCreate Service)                                       │
│  language_type=java → 自动推导 jenkins_job=k8s-builder-java              │
│  自动注入: JAVA_VERSION=17, ENABLE_SONAR=true, MAVEN_GOALS=...          │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │ Jenkins API 触发构建
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  Jenkins k8s-builder-java (Groovy 模板)                                 │
│  Checkout → Compile → Test → SonarQube → Quality Gate → Package         │
│  → Build Image → Push Image → Callback                                  │
└──────────────────────────────┬──────────────────────────────────────────┘
                               │ HTTP 回调 (HMAC-SHA256 签名)
                               ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  后端回调处理                                                            │
│  /stage/callback (实时阶段更新) + /pipeline/callback (最终状态)           │
│  /pipeline/sonar-callback (SonarQube 指标数据)                           │
│  → auto_deploy=true 时触发 StrategicMergePatch 滚动更新 K8s              │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 二、业务项目准备（仅需 2 个文件）

业务 Java 项目**无需 Jenkinsfile**，仅需在项目根目录提供：

### 2.1 `pom.xml`（必须）

标准 Maven 项目文件。Groovy 模板会执行 `mvn clean compile`、`mvn test`、`mvn package`。

### 2.2 `Dockerfile`（必须）

推荐使用平台提供的多阶段构建模板（`docs/dockerfile/Dockerfile.java-maven`）：

```dockerfile
# ============ 构建阶段 ============
FROM maven:3.9-eclipse-temurin-17 AS builder
WORKDIR /build
COPY pom.xml .
RUN mvn dependency:go-offline -B
COPY src ./src
RUN mvn package -DskipTests -B && mv target/*.jar app.jar

# ============ 运行阶段 ============
FROM eclipse-temurin:17-jre-alpine
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /build/app.jar .
USER appuser
EXPOSE 8080
ENV JAVA_OPTS="-Xms256m -Xmx512m -XX:+UseG1GC -XX:+HeapDumpOnOutOfMemoryError"
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
    CMD wget -qO- http://localhost:8080/actuator/health || exit 1
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

---

## 三、前端创建流水线 — 字段详解

前端 5 步向导 → 提交 JSON → `POST /api/v1/k8s/cicd/pipeline/create`

### 3.1 Step 1: 基本信息

| 字段 | JSON Key | 类型 | 必填/可选 | 默认值 | 说明 |
|------|----------|------|----------|--------|------|
| 流水线名称 | `name` | string | **必填** | - | 后端校验: `required`, `between:1,100`。全局唯一，建议格式 `{项目名}-{环境}`，如 `order-service-prod` |
| 描述 | `description` | string | 可选 | `""` | 简要说明流水线用途，方便团队成员理解。不传则为空 |

**必填原因**：
- `name`: 流水线的唯一标识，用于列表展示、搜索、日志追踪。后端会检查名称唯一性，重复会返回 `流水线名称已存在` 错误

### 3.2 Step 2: 代码仓库

| 字段 | JSON Key | 类型 | 必填/可选 | 默认值 | 说明 |
|------|----------|------|----------|--------|------|
| Git 仓库地址 | `git_repo` | string | **必填** | - | 后端校验: `required`, `url`。完整的 HTTPS 仓库地址，如 `https://gitee.com/org/my-app.git` |
| Git 分支 | `git_branch` | string | 可选 | `"main"` | 默认构建分支。运行时可通过 `PipelineRunRequest.branch` 覆盖 |

**必填原因**：
- `git_repo`: Jenkins 需要此地址拉取代码，无仓库地址则无法执行任何构建。后端验证 URL 格式合法性

**可选原因**：
- `git_branch`: 后端 `NewPipelineCreateRequest()` 默认设为 `"main"`，符合主流 Git 分支命名。用户可按需修改为 `master`、`develop` 等

### 3.3 Step 3: Jenkins 配置

| 字段 | JSON Key | 类型 | 必填/可选 | 默认值 | 说明 |
|------|----------|------|----------|--------|------|
| Jenkins 服务器地址 | `jenkins_url` | string | 可选 | `""` | 留空则使用 `config.yaml` 中配置的默认 Jenkins 服务器 |
| Jenkins Job 名称 | `jenkins_job` | string | **条件必填** | `""` | **Java 项目留空即可**！后端自动推导为 `k8s-builder-java`。仅 `language_type=custom` 时必填 |
| 语言类型 | `language_type` | string | 可选 | `"custom"` | 后端校验: `in:go,java,frontend,python,custom`。**Java 项目必须选 `java`** |
| SonarQube 扫描 | — | boolean | 可选 | Java 自动启用 | 前端开关 `enable_sonar`，提交时转换为 `env_vars` 中的 `ENABLE_SONAR=true` |
| 环境变量 | `env_vars` | array | 可选 | `[]` | 格式 `[{name: "KEY", value: "VALUE"}]`。会合并到 Jenkins 构建参数中 |

**条件必填原因**：
- `jenkins_job`: 模板化架构的核心设计——4 个通用 Jenkins Job（`k8s-builder-go/java/frontend/python`）服务 100+ 项目。当 `language_type` 为非 `custom` 时，后端 `PipelineCreate` 服务自动从 `DefaultJenkinsJobMap` 映射：

  ```go
  var DefaultJenkinsJobMap = map[string]string{
      "go":       "k8s-builder-go",
      "java":     "k8s-builder-java",
      "frontend": "k8s-builder-frontend",
      "python":   "k8s-builder-python",
  }
  ```

  只有 `custom` 类型无法推导，因此必须手动填写。

**可选原因**：
- `jenkins_url`: 绝大多数团队只有一个 Jenkins 实例，在 `config.yaml` 中全局配置即可，无需每条流水线重复填写
- `language_type`: 后端默认 `custom`，但 **Java 项目务必选择 `java`**，否则无法触发自动参数注入
- `env_vars`: 后端 `injectLanguageParams()` 已为 Java 项目自动注入以下参数，通常无需手动填写：

  | 自动注入参数 | 默认值 | 用途 |
  |-------------|--------|------|
  | `JAVA_VERSION` | `17` | JDK 版本 |
  | `MAVEN_GOALS` | `clean package -DskipTests -B` | Maven 构建命令 |
  | `SKIP_TESTS` | `false` | 是否跳过测试 |
  | `ENABLE_SONAR` | `true` | SonarQube 代码扫描 |
  | `SONAR_QUALITY_GATE` | `true` | 质量门禁检查 |
  | `SONAR_SOURCES` | `src/main/java` | 扫描源码目录 |
  | `SONAR_JAVA_BINARIES` | `target/classes` | 编译输出目录 |
  | `SONAR_EXCLUSIONS` | `**/test/**,**/generated/**` | 排除扫描的文件 |
  | `DOCKERFILE_PATH` | `Dockerfile` | Dockerfile 路径 |
  | `GIT_CREDENTIAL_ID` | `gitee-id` | Git 凭证 ID |

  如需覆盖（如使用 Java 11），可在 `env_vars` 中传入 `{name: "JAVA_VERSION", value: "11"}`，`env_vars` 的优先级 > 自动注入参数

### 3.4 Step 4: 部署策略

| 字段 | JSON Key | 类型 | 必填/可选 | 默认值 | 说明 |
|------|----------|------|----------|--------|------|
| 部署环境 | `deploy_env` | string | 可选 | `"dev"` | 可选值: `dev`/`test`/`staging`/`prod`。生产环境自动要求审批 |
| 部署配置 | `deploy_config` | object | 可选 | 见下方 | 包含副本数、策略、资源配置 |

`deploy_config` 结构：

```json
{
  "replicas": 3,
  "strategy": "rollingUpdate",
  "resources": {
    "limits": { "cpu": "500m", "memory": "512Mi" },
    "requests": { "cpu": "200m", "memory": "256Mi" }
  }
}
```

| 子字段 | 默认值 | 说明 |
|--------|--------|------|
| `replicas` | `3` | Pod 副本数 |
| `strategy` | `"rollingUpdate"` | 部署策略，可选 `rollingUpdate`/`recreate` |
| `resources.limits.cpu` | `"500m"` | CPU 上限 |
| `resources.limits.memory` | `"512Mi"` | 内存上限 |
| `resources.requests.cpu` | `"200m"` | CPU 请求 |
| `resources.requests.memory` | `"256Mi"` | 内存请求 |

**可选原因**：
- `deploy_config`: 后端存储为 `map[string]any` JSON 字段，不参与后端验证。前端有默认值，不传则使用前端默认配置。此配置在自动部署时用于 K8s 滚动更新策略
- `deploy_env`: 主要影响资源模板校验和审批规则。`prod` 环境自动设置 `require_approval=true`

### 3.5 Step 5: 自动部署

| 字段 | JSON Key | 类型 | 必填/可选 | 默认值 | 说明 |
|------|----------|------|----------|--------|------|
| 启用自动部署 | `auto_deploy` | boolean | 可选 | `false` | 开启后 Jenkins 构建成功会自动更新 K8s 工作负载镜像 |
| 目标集群 | `target_cluster_id` | number | **条件必填** | `0` | `auto_deploy=true` 时必须选择目标 K8s 集群 |
| 目标命名空间 | `target_namespace` | string | **条件必填** | `""` | `auto_deploy=true` 时必须指定部署到哪个命名空间 |
| 工作负载类型 | `target_workload_kind` | string | 可选 | `"Deployment"` | 可选 `Deployment`/`StatefulSet`/`DaemonSet` |
| 工作负载名称 | `target_workload_name` | string | **条件必填** | `""` | `auto_deploy=true` 时必须指定要更新的工作负载 |
| 容器名称 | `target_container` | string | 可选 | `""` | 留空则更新工作负载中的第一个容器。多容器 Pod 时需指定 |
| 需要审批 | `require_approval` | boolean | 可选 | `true` | 生产环境建议开启。审批通过后才执行自动部署 |

**条件必填原因**：
- `target_cluster_id` / `target_namespace` / `target_workload_name`: 仅在 `auto_deploy=true` 时需要。自动部署的核心是通过 `client-go` 的 `StrategicMergePatch` 原子更新 K8s 工作负载的容器镜像，必须精确指定目标

**可选原因**：
- `auto_deploy`: 并非所有流水线都需要自动部署，有些只是 CI（构建+测试），不涉及 CD
- `target_workload_kind`: 默认 `Deployment` 是最常用的工作负载类型，90% 的场景不需要修改
- `target_container`: 大多数业务 Pod 只有 1 个业务容器，留空自动选择第一个
- `require_approval`: 开发/测试环境可关闭审批加速部署，生产环境建议保持开启

---

## 四、完整请求体示例

### 4.1 最小化（仅必填字段）

```json
{
  "name": "order-service-pipeline",
  "git_repo": "https://gitee.com/myorg/order-service.git",
  "language_type": "java"
}
```

后端自动补全为：
- `git_branch` → `"main"`
- `jenkins_job` → `"k8s-builder-java"`（从 `DefaultJenkinsJobMap["java"]` 推导）
- `ENABLE_SONAR` → `"true"`（`injectLanguageParams` 自动注入）
- `JAVA_VERSION` → `"17"`、`MAVEN_GOALS` → `"clean package -DskipTests -B"`
- 其他 SonarQube 参数全部自动注入

### 4.2 完整配置（含自动部署 + SonarQube）

```json
{
  "name": "order-service-prod",
  "description": "订单服务生产环境 Java Spring Boot 流水线",
  "git_repo": "https://gitee.com/myorg/order-service.git",
  "git_branch": "release",
  "jenkins_url": "",
  "jenkins_job": "",
  "language_type": "java",
  "env_vars": [
    { "name": "JAVA_VERSION", "value": "17" },
    { "name": "ENABLE_SONAR", "value": "true" },
    { "name": "SONAR_QUALITY_GATE", "value": "true" },
    { "name": "SPRING_PROFILES_ACTIVE", "value": "prod" }
  ],
  "deploy_config": {
    "replicas": 3,
    "strategy": "rollingUpdate",
    "resources": {
      "limits": { "cpu": "1000m", "memory": "1024Mi" },
      "requests": { "cpu": "500m", "memory": "512Mi" }
    }
  },
  "auto_deploy": true,
  "target_cluster_id": 1,
  "target_namespace": "production",
  "target_workload_kind": "Deployment",
  "target_workload_name": "order-service",
  "target_container": "",
  "deploy_env": "prod",
  "require_approval": true
}
```

### 4.3 仅 CI 不自动部署

```json
{
  "name": "order-service-ci",
  "description": "订单服务 CI 流水线（仅构建和扫描，不部署）",
  "git_repo": "https://gitee.com/myorg/order-service.git",
  "git_branch": "develop",
  "language_type": "java",
  "auto_deploy": false
}
```

---

## 五、后端处理流程详解

### 5.1 创建阶段 — `PipelineCreate` Service

```
用户提交 JSON
    │
    ▼
ValidPipelineCreateRequest 校验
    │  - name: required, between:1,100
    │  - git_repo: required, url
    │  - language_type: in:go,java,frontend,python,custom
    │
    ▼
检查名称唯一性（PipelineGetByName）
    │
    ▼
模板化推导:
    │  language_type="java" + jenkins_job=""
    │  → jenkins_job = DefaultJenkinsJobMap["java"] = "k8s-builder-java"
    │
    ▼
写入数据库 CicdPipeline 表 → 返回 pipeline_id
```

### 5.2 运行阶段 — `PipelineRun` Service

```
POST /api/v1/k8s/cicd/pipeline/run { "id": 123 }
    │
    ▼
获取流水线配置 → 检查状态（idle/running/disabled）
    │
    ▼
创建 CicdPipelineRun 记录
    │
    ▼
injectLanguageParams() 自动注入 Java 参数:
    │  JAVA_VERSION=17, MAVEN_GOALS=..., ENABLE_SONAR=true
    │  SONAR_QUALITY_GATE=true, SONAR_SOURCES=src/main/java
    │
    ▼
合并 env_vars（用户自定义覆盖自动注入）
    │
    ▼
异步触发 Jenkins 构建:
    │  client.TriggerBuildAndWait("k8s-builder-java", params, 60s)
    │
    ▼
Jenkins 使用 java-spring-pipeline.groovy 模板执行:
    ┌───────────────────────────────────────────────┐
    │ Clean Workspace → Checkout → Compile → Test   │
    │ → SonarQube Analysis → Quality Gate           │
    │ → Package → Build Image → Push Image          │
    │                                               │
    │ 每阶段完成回调: POST /stage/callback           │
    │ 最终完成回调: POST /pipeline/callback           │
    │ SonarQube 指标回传: POST /pipeline/sonar-callback│
    └───────────────────────────────────────────────┘
    │
    ▼
回调处理:
    │  - PipelineCallback: 更新运行状态、镜像地址
    │  - SaveSonarReport: 保存 bugs/vulnerabilities/coverage 等指标
    │  - auto_deploy=true → StrategicMergePatch 更新 K8s 工作负载镜像
```

---

## 六、Jenkins Groovy 模板执行阶段

`java-spring-pipeline.groovy` 的完整构建流程：

| 序号 | 阶段 | stage_type | 说明 | 可配置参数 |
|------|------|-----------|------|-----------|
| 1 | Clean Workspace | `clean` | 清理工作空间、拉取代码 | `GIT_REPO`(必填), `GIT_BRANCH`, `GIT_CREDENTIAL_ID` |
| 2 | Checkout Info | `checkout` | 获取 Git commit、生成镜像标签 | `IMAGE_REPO`(必填), `IMAGE_TAG` |
| 3 | Compile | `compile` | `mvn clean compile -DskipTests -B` | - |
| 4 | Test | `test` | `mvn test -B`（可跳过） | `SKIP_TESTS`(default:false) |
| 5 | SonarQube Analysis | `sonar` | 代码质量扫描（可关闭） | `ENABLE_SONAR`(default:true), `SONAR_PROJECT_KEY`, `SONAR_SOURCES` |
| 6 | Quality Gate | `quality_gate` | 质量门禁检查（可关闭） | `SONAR_QUALITY_GATE`(default:true) |
| 7 | Package | `dependencies` | `mvn clean package -DskipTests -B` | `MAVEN_GOALS` |
| 8 | Build Image | `build` | `nerdctl build` 构建容器镜像 | `DOCKERFILE_PATH`(default:Dockerfile), `JAVA_VERSION`(default:17) |
| 9 | Push Image | `push` | 推送镜像到 Harbor 仓库 | `IMAGE_REPO`(必填) |

### Jenkins 参数总表

| 参数 | 必填/可选 | 默认值 | 来源 | 说明 |
|------|----------|--------|------|------|
| `GIT_REPO` | **必填** | - | `pipeline.git_repo` | Git 仓库地址 |
| `GIT_BRANCH` | 可选 | `main` | `pipeline.git_branch` | 构建分支 |
| `IMAGE_REPO` | **必填** | - | 平台全局配置 | 镜像仓库地址，如 `harbor.example.com/myproject/order-service` |
| `IMAGE_TAG` | 可选 | 自动生成 | - | 留空则自动生成 `{branch}-{commit}-{timestamp}` |
| `PIPELINE_ID` | 自动 | - | `pipeline.id` | 平台流水线 ID，用于回调匹配 |
| `PLATFORM_CALLBACK_URL` | 自动 | - | `config.yaml` | 平台回调地址 |
| `JAVA_VERSION` | 可选 | `17` | `injectLanguageParams` | JDK 版本 |
| `MAVEN_GOALS` | 可选 | `clean package -DskipTests -B` | `injectLanguageParams` | Maven 构建命令 |
| `SKIP_TESTS` | 可选 | `false` | `injectLanguageParams` | 跳过单元测试 |
| `ENABLE_SONAR` | 可选 | `true` | `injectLanguageParams` | SonarQube 代码扫描 |
| `SONAR_QUALITY_GATE` | 可选 | `true` | `injectLanguageParams` | 质量门禁检查 |
| `SONAR_PROJECT_KEY` | 可选 | Job 名称 | 用户自定义 | SonarQube 项目标识 |
| `SONAR_SOURCES` | 可选 | `src/main/java` | `injectLanguageParams` | 源码扫描目录 |
| `SONAR_JAVA_BINARIES` | 可选 | `target/classes` | `injectLanguageParams` | 编译输出目录 |
| `SONAR_EXCLUSIONS` | 可选 | `**/test/**,**/generated/**` | `injectLanguageParams` | 排除扫描文件 |
| `DOCKERFILE_PATH` | 可选 | `Dockerfile` | `injectLanguageParams` | Dockerfile 路径 |
| `GIT_CREDENTIAL_ID` | 可选 | `gitee-id` | `injectLanguageParams` | Git 凭证 ID |

---

## 七、SonarQube 代码质量扫描

### 7.1 执行条件

- `ENABLE_SONAR=true`（Java 项目默认启用）
- Jenkins 已安装 SonarQube Scanner 插件
- Jenkins 系统配置中配置了名为 `SonarQube` 的服务器连接

### 7.2 扫描指标

Quality Gate 通过后，Groovy 模板调用 `sonarReportCallback()` 将以下指标回传平台：

| 指标 | 说明 | SonarQube API 字段 |
|------|------|--------------------|
| bugs | Bug 数量 | `bugs` |
| vulnerabilities | 安全漏洞数 | `vulnerabilities` |
| code_smells | 代码异味数 | `code_smells` |
| coverage | 代码覆盖率(%) | `coverage` |
| duplications | 重复率(%) | `duplicated_lines_density` |
| lines_of_code | 代码行数 | `ncloc` |
| security_hotspots | 安全热点数 | `security_hotspots` |
| reliability_rating | 可靠性评级 (A-E) | `reliability_rating` |
| security_rating | 安全性评级 (A-E) | `security_rating` |
| maintainability_rating | 可维护性评级 (A-E) | `sqale_rating` |

### 7.3 关闭 SonarQube

如需关闭代码扫描，在前端创建/编辑流水线时关闭 SonarQube 开关，或在 `env_vars` 中传入：

```json
{ "name": "ENABLE_SONAR", "value": "false" }
```

---

## 八、回调与安全机制

### 8.1 三种回调接口

| 回调类型 | 路径 | 触发时机 | 数据内容 |
|---------|------|---------|---------|
| 阶段回调 | `POST /api/v1/k8s/cicd/stage/callback` | 每个阶段完成时 | `job_name`, `build_number`, `pipeline_id`, `stage_type`, `status` |
| 最终回调 | `POST /api/v1/k8s/cicd/pipeline/callback` | 构建完成时 | `job_name`, `build_number`, `status`, `image_url`, `image_digest`, `duration_sec` |
| SonarQube 回调 | `POST /api/v1/k8s/cicd/pipeline/sonar-callback` | Quality Gate 后 | `pipeline_id`, `quality_gate`, `bugs`, `vulnerabilities`, `coverage` 等完整指标 |

### 8.2 HMAC-SHA256 签名验证

所有回调请求携带 `X-Signature` Header：

```
签名算法: HMAC-SHA256(secret, "{job_name}:{build_number}:{status_or_stage}")
```

- 平台端密钥: `config.yaml` → `jenkins.hmac_secret`
- Jenkins 端密钥: Jenkins Credentials → `hmac-secret`
- 双方密钥必须一致，否则回调被拒绝

---

## 九、自动部署到 K8s

### 9.1 触发条件

```
Jenkins 构建成功
  + auto_deploy = true
  + target_cluster_id > 0
  + target_namespace 非空
  + target_workload_name 非空
  + require_approval = false（或审批已通过）
```

### 9.2 部署实现

后端使用 `client-go` 的 `StrategicMergePatch` 原子更新容器镜像：

```go
// 构建 patch JSON
patch := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"%s","image":"%s"}]}}}}`,
    containerName, newImage)

// 原子更新
_, err = clientset.AppsV1().Deployments(namespace).Patch(
    ctx, workloadName, types.StrategicMergePatchType, []byte(patch), metav1.PatchOptions{})
```

**优势**: 原子操作、无并发冲突、自动触发 K8s 滚动更新、无需 kubectl 依赖

---

## 十、端到端操作流程

### Jenkins Job 一次性配置（推荐 Pipeline script from SCM）

```
Step 1: Jenkins 创建 k8s-builder-java Job
        │
        │  推荐方式（Pipeline script from SCM）：
        │  ┌────────────────────────────────────────────────────────────┐
        │  │ Pipeline:                                                    │
        │  │   Definition: Pipeline script from SCM                       │
        │  │   SCM: Git                                                   │
        │  │   Repository URL: 平台仓库地址                               │
        │  │     （如 https://gitee.com/your-org/k8s_operation.git）        │
        │  │   Credentials: gitee-id                                     │
        │  │   Branch: */main                                            │
        │  │   Script Path:                                               │
        │  │     configs/jenkins-templates/java-spring-pipeline.groovy    │
        │  └────────────────────────────────────────────────────────────┘
        │
        │  优势：
        │  ✔ 模板版本化管理，Git 提交即更新
        │  ✔ 无需手动粘贴 Groovy 脚本到 Jenkins
        │  ✔ 所有 Java 项目共用同一个模板
        │  ✔ 模板变更自动生效，无需重新配置 Job
        ↓

Step 2: 业务项目根目录放置 Dockerfile + pom.xml
        ↓

Step 3: 平台前端创建流水线
        ↓ (选择 language_type=java, 填写 git_repo)

Step 4: 点击运行
        ↓

Step 5: 全自动执行
        ↓ Compile → Test → SonarQube → Package → Build → Push → Callback

Step 6: 自动滚动更新 K8s（如配置了 auto_deploy）
```

### 前端界面操作步骤（5 步向导详解）

前端 `PipelineCreate.vue` 提供 5 步向导式创建界面，以下是 Java 项目的完整操作流程：

1. 进入 **CI/CD → 流水线管理** 页面，点击 **创建流水线**
2. **Step 1 基本信息**: 输入流水线名称（如 `order-service-prod`），可选填写描述
3. **Step 2 代码仓库**: 
   - 填写 Git 地址（如 `https://gitee.com/org/order-service.git`）
   - 点击「获取分支」按钮自动拉取远程分支列表，下拉选择目标分支
4. **Step 3 Jenkins**（核心步骤）: 
   - 服务类型选择 **Java**（自动联动 `language_type=java`）
   - Jenkins Job 名称**留空**（自动推导为 `k8s-builder-java`）
   - SonarQube 扫描默认已开启（Java 自动启用，显示「Java 推荐」标签）
   - 如需覆盖默认参数，展开「环境变量」添加（如 `JAVA_VERSION=11`）
5. **Step 4 部署策略**: 选择部署环境、调整副本数和资源配置
6. **Step 5 自动部署**: 
   - 开启自动部署开关
   - 下拉选择目标集群（自动加载）→ 命名空间（级联加载）→ 工作负载（级联加载）
   - 生产环境建议开启审批
7. 点击 **创建**

> **Java 特有自动行为**：前端选择 Java 语言类型后，会触发 `onServiceTypeChange('java')`，自动：
> 1. 设置 `language_type = 'java'`
> 2. 开启 `enable_sonar = true`
> 3. 提交时自动注入 `ENABLE_SONAR=true` 和 `SONAR_QUALITY_GATE=true` 到 `env_vars`
> 4. 后端自动推导 `jenkins_job = 'k8s-builder-java'`
> 5. 后端 `injectLanguageParams()` 自动注入 JAVA_VERSION、MAVEN_GOALS 等参数

---

## 十一、字段必填/可选速查表

| JSON 字段 | 类型 | 必填性 | 默认值 | 后端校验 | 可选原因 |
|-----------|------|-------|--------|---------|---------|
| `name` | string | **必填** | - | `required`, `between:1,100` | 唯一标识，不可缺省 |
| `git_repo` | string | **必填** | - | `required`, `url` | Jenkins 必须有仓库地址才能拉代码 |
| `description` | string | 可选 | `""` | 无 | 辅助说明，不影响构建流程 |
| `git_branch` | string | 可选 | `"main"` | 无 | 后端默认值覆盖，主流分支名 |
| `jenkins_url` | string | 可选 | `""` | 无 | 使用全局配置，单实例场景无需重复填写 |
| `jenkins_job` | string | 条件必填 | `""` | 无 | 模板化推导，非 custom 类型自动映射 |
| `language_type` | string | 可选 | `"custom"` | `in:go,java,frontend,python,custom` | 有默认值，但 Java 项目**务必选 `java`** |
| `env_vars` | array | 可选 | `[]` | 无 | `injectLanguageParams` 已自动注入基础参数 |
| `deploy_config` | object | 可选 | 前端默认 | 无 | 有合理默认值，按需调整 |
| `auto_deploy` | boolean | 可选 | `false` | 无 | 仅 CI 场景无需部署 |
| `target_cluster_id` | number | 条件必填 | `0` | 无 | 仅 `auto_deploy=true` 时需要 |
| `target_namespace` | string | 条件必填 | `""` | 无 | 仅 `auto_deploy=true` 时需要 |
| `target_workload_kind` | string | 可选 | `"Deployment"` | 无 | 90% 场景使用 Deployment |
| `target_workload_name` | string | 条件必填 | `""` | 无 | 仅 `auto_deploy=true` 时需要 |
| `target_container` | string | 可选 | `""` | 无 | 留空更新第一个容器 |
| `deploy_env` | string | 可选 | `"dev"` | 无 | 前端有默认值 |
| `require_approval` | boolean | 可选 | `true` | 无 | 生产环境建议开启，开发环境可关闭 |

---

## 十二、常见问题

### Q1: Java 项目必须填哪些字段？

最少只需 3 个字段：`name`、`git_repo`、`language_type: "java"`。其余全部自动推导。

### Q2: 为什么 `jenkins_job` 不是必填？

模板化架构设计：100+ 项目共用 4 个通用 Jenkins Job，后端根据 `language_type` 自动映射到 `k8s-builder-java`，用户无需关心 Job 名称。

### Q3: SonarQube 扫描会影响构建速度吗？

SonarQube 扫描通常增加 1-3 分钟。如需加速，可在 `env_vars` 中设置 `ENABLE_SONAR=false` 或在前端关闭 SonarQube 开关。

### Q4: 关闭 SonarQube 后如何重新开启？

编辑流水线，在 Step 3 打开 SonarQube 开关即可。前端会自动管理 `ENABLE_SONAR` 和 `SONAR_QUALITY_GATE` 环境变量的注入和清理。

### Q5: `language_type` 选错了会怎样？

选错（如 Java 项目选了 `go`）会导致 Jenkins 使用 `k8s-builder-go` 模板构建，找不到 `pom.xml` 导致构建失败。务必确保语言类型与项目实际技术栈一致。

### Q6: 自动部署失败会影响构建状态吗？

不会。Jenkins 构建和自动部署是解耦的：构建成功后 Jenkins 回调平台，平台再独立执行自动部署。构建状态仍为 SUCCESS，部署状态单独记录。
