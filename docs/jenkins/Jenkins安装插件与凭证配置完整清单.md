# Jenkins 完整配置、插件与架构说明文档

> 本文档适用于 K8s Operation Platform 的 Jenkins 全容器化部署方案。  
> Jenkins 运行在 K8s 集群 `devops` 命名空间，使用动态 Pod Agent 构建，无需在 Master 上安装任何编译工具。

---

## 一、平台整体架构（全 K8s Pod 容器化）

### 1.1 架构拓扑

```
K8s 集群
├── namespace: devops
│   ├── Jenkins Master (StatefulSet, 1副本)
│   │   ├── initContainer: 自动安装 13 个插件
│   │   └── 主容器: jenkins/jenkins:lts
│   └── 构建 Pod（动态创建/构建完自动销毁）
│       ├── 语言编译容器 (golang/maven/node/python)
│       └── kaniko 容器（构建+推送镜像）
│
└── namespace: k8soperation
    ├── 后端 Deployment (Go 二进制 + config.yaml)
    └── 前端 Deployment (Nginx + Vue 静态文件, 2副本)
```

### 1.2 架构特点

| 特点 | 说明 |
|------|------|
| **全 Pod 化** | Jenkins Master、构建 Agent、平台前后端全部以 K8s Pod 运行 |
| **Jenkins Master 零工具** | 不安装 JDK、Maven、Node、Go、Docker 等任何构建工具 |
| **动态 Pod Agent** | 每次构建自动创建临时 Pod，完成后自动销毁，不占用持续资源 |
| **Kaniko 构建** | 无需 Docker daemon、无需特权模式、纯用户空间构建镜像 |
| **服务间通信** | 全部通过 K8s Service DNS（如 `jenkins.devops.svc.cluster.local`） |
| **一键部署** | `kubectl apply -f deploy/jenkins/` 即可部署完整 Jenkins 环境 |

### 1.3 各组件部署方式

| 组件 | K8s 资源类型 | 命名空间 | 镜像 |
|------|------------|---------|------|
| Jenkins Master | StatefulSet | `devops` | `jenkins/jenkins:lts` |
| 构建 Agent | 临时 Pod（Kubernetes Plugin 动态创建） | `devops` | 按语言动态选择 |
| 平台后端 | Deployment | `k8soperation` | Go 编译产物 |
| 平台前端 | Deployment (×2) | `k8soperation` | Nginx + Vue |

---

## 二、容器运行时兼容性（Docker / containerd / CRI-O）

### 2.1 核心结论：与容器运行时完全无关

| K8s 节点容器运行时 | Jenkins 运行 | 构建 Pod 创建 | Kaniko 构建镜像 | 镜像推送 |
|:---:|:---:|:---:|:---:|:---:|
| **Docker** | ✅ | ✅ | ✅ | ✅ |
| **containerd** | ✅ | ✅ | ✅ | ✅ |
| **CRI-O** | ✅ | ✅ | ✅ | ✅ |

### 2.2 为什么与运行时无关

```
┌─────────────────────────────────────────────────────┐
│  kubelet (管理Pod)                                   │
│      │                                              │
│      ▼ CRI 标准接口                                  │
│  ┌──────────────────────────────────┐               │
│  │ Docker / containerd / CRI-O     │ ← 只负责拉取镜像+启动容器
│  └──────────────────────────────────┘               │
│      │                                              │
│      ▼ 启动构建 Pod                                  │
│  ┌──────────────────────────────────┐               │
│  │ maven 容器: 编译代码              │               │
│  │ kaniko 容器: 构建镜像 → HTTP推送   │ ← 不调用任何 runtime socket
│  └──────────────────────────────────┘               │
└─────────────────────────────────────────────────────┘
```

- **kubelet**：通过 CRI 接口管理 Pod，不关心底层是 Docker 还是 containerd
- **Kaniko**：纯用户空间进程，直接解析 Dockerfile → 构建文件系统层 → 用 HTTPS 推送到 Harbor
- **不需要** Docker socket、不需要 containerd socket、不需要特权模式

> K8s 1.24+ 已移除 dockershim（Docker 作为运行时），本平台**完全不受影响**。

### 2.3 Kaniko vs Docker vs nerdctl 对比

| 对比项 | docker build | nerdctl build | **Kaniko（本项目）** |
|--------|:---:|:---:|:---:|
| 需要 Docker daemon | 是 | 否 | **否** |
| 需要容器运行时 socket | Docker socket | containerd socket | **都不需要** |
| 需要特权模式 | 是 | 是 | **否** |
| 运行位置 | 宿主机 | 宿主机 | **K8s Pod 内** |
| K8s 原生兼容 | 需 DinD | 需挂载 socket | **原生兼容** |
| 安全性 | 低（特权+daemon） | 中 | **高（无特权）** |

---

## 三、镜像构建方式（Kaniko）

### 3.1 工作原理

```
构建 Pod
┌─────────────────────────────────────────────┐
│  kaniko 容器                                 │
│  ┌─────────────────────────────────────┐    │
│  │ /kaniko/executor                    │    │
│  │   --context=.                       │    │
│  │   --dockerfile=Dockerfile           │    │
│  │   --destination=harbor.com/app:tag  │    │
│  └─────────────────────────────────────┘    │
│       │                                     │
│       ├── 1. 解析 Dockerfile                 │
│       ├── 2. 在用户空间构建文件系统层          │
│       ├── 3. 生成 OCI 镜像                   │
│       └── 4. 通过 HTTPS 推送到 Harbor         │
└─────────────────────────────────────────────┘
```

### 3.2 各语言的 Kaniko 构建镜像

所有模板都在 Pod 中包含一个 `kaniko` sidecar 容器：

```yaml
- name: kaniko
  image: gcr.io/kaniko-project/executor:debug
  command: ['sleep', '99d']
```

构建产出后由 kaniko 容器执行 `/kaniko/executor` 生成并推送镜像。

### 3.3 自动生成 Dockerfile

如果项目没有自带 Dockerfile，模板会**自动生成**针对语言优化的运行时 Dockerfile：

| 语言 | 自动生成的 FROM 基础镜像 | 说明 |
|------|----------------------|------|
| Go | `alpine:3.20` | 纯静态二进制，超小镜像 |
| Java | `java:${JAVA_VERSION}-jre-alpine` | JRE 运行时（动态版本） |
| Frontend | `nginx:1.25-alpine` | Nginx 托管静态资源 |
| Python | `python:${VERSION}-slim` | Python 运行时 |

---

## 四、Java 多版本动态选择

### 4.1 支持的 Java 版本

| 版本 | 构建镜像 | 运行时镜像 |
|:---:|---------|----------|
| **8** | `maven:3.9-eclipse-temurin-8` | `java:8-jre-alpine` |
| **11** | `maven:3.9-eclipse-temurin-11` | `java:11-jre-alpine` |
| **17**（默认） | `maven:3.9-eclipse-temurin-17` | `java:17-jre-alpine` |
| **21** | `maven:3.9-eclipse-temurin-21` | `java:21-jre-alpine` |

### 4.2 版本选择方式

Pipeline 参数定义（用户在平台 UI 中选择）：

```groovy
choice(name: 'JAVA_VERSION', choices: ['17', '21', '11', '8'],
       description: 'Java 版本（同时决定构建 JDK 和运行时镜像：8/11/17/21）')
```

### 4.3 动态联动机制

```
用户选择 JAVA_VERSION=11
         │
         ▼
┌─────────────────────────────────────┐
│ ① 构建阶段 - 动态 Pod Agent 镜像      │
│   image: maven:3.9-eclipse-temurin-11│ ← 用 JDK 11 + Maven 3.9 编译
│                                     │
│ ② 打包阶段 - Kaniko 构建运行时镜像     │
│   FROM java:11-jre-alpine            │ ← 基于 JDK 11 运行时
│   COPY target/*.jar /app/app.jar     │
└─────────────────────────────────────┘
         │
         ▼
  推送到 Harbor: myapp:abc123-20240618
  （内含 JDK 11 运行环境）
```

### 4.4 版本优先级

```
运行时传入 JAVA_VERSION  >  流水线保存的 env_vars  >  平台默认值（17）
```

### 4.5 构建镜像内置环境

`maven:3.9-eclipse-temurin-{VERSION}` 镜像**已包含**：
- JDK（对应版本的 Eclipse Temurin 发行版）
- Maven 3.9.x
- 基本 Linux 工具

**不需要手动安装 JDK 或 Maven**，构建 Pod 启动后即可直接使用。

---

## 五、Maven 阿里云镜像加速（自动配置）

### 5.1 自动生成 settings.xml

Java 模板在 `Setup & Compile` 阶段**自动生成**阿里云 Maven 镜像配置：

```groovy
writeFile file: settingsFile, text: """\
<?xml version="1.0" encoding="UTF-8"?>
<settings>
  <mirrors>
    <mirror>
      <id>aliyun</id>
      <name>Aliyun Maven Mirror</name>
      <url>https://maven.aliyun.com/repository/public</url>
      <mirrorOf>central</mirrorOf>
    </mirror>
  </mirrors>
</settings>
"""
```

### 5.2 使用方式

```groovy
sh "mvn package -DskipTests -B -T 1C -s ${env.MVN_SETTINGS}"
```

### 5.3 效果

| 项目 | 说明 |
|------|------|
| 镜像源 | `https://maven.aliyun.com/repository/public` |
| 是否需要手动配置 | **不需要**，模板自动完成 |
| 加速范围 | Maven Central 仓库的所有依赖 |
| 缓存加速 | PVC `jenkins-maven-cache` 跨构建持久化 `/root/.m2/repository` |

> **首次构建**：从阿里云镜像下载（国内高速）→ 存入 PVC 缓存  
> **后续构建**：直接从 PVC 缓存读取，无需重复下载

---

## 六、部署方式概览（Jenkins Master）

| 项目 | 说明 |
|------|------|
| 镜像 | `jenkins/jenkins:lts` |
| 部署方式 | K8s StatefulSet（单副本） |
| 命名空间 | `devops` |
| 数据持久化 | PVC `jenkins-data`（20Gi） |
| 构建模式 | K8s 动态 Pod Agent + Kaniko（无需 Docker daemon） |
| 配置方式 | JCasC（Jenkins Configuration as Code） |
| 外部访问 | NodePort 30080 或 Ingress |

---

## 七、必装插件清单

以下插件通过 `initContainer` 中的 `jenkins-plugin-cli` 自动安装（不指定版本号，安装最新兼容版）：

| # | 插件名称 | 用途说明 | 被哪些模板使用 |
|---|----------|----------|----------------|
| 1 | **kubernetes** | Kubernetes Cloud 集成，动态创建 Pod Agent 执行构建任务 | 全部 |
| 2 | **workflow-aggregator** | Pipeline 核心插件集（Declarative Pipeline + Stage View） | 全部 |
| 3 | **git** | Git SCM 支持，拉取源代码 | 全部 |
| 4 | **configuration-as-code** | JCasC 插件，通过 YAML 自动化配置 Jenkins | 系统级 |
| 5 | **credentials-binding** | 凭证绑定，Pipeline 中 `credentials()` 函数 | 全部 |
| 6 | **pipeline-stage-view** | Pipeline 阶段可视化视图 | 系统级 |
| 7 | **blueocean** | Blue Ocean 现代化 Pipeline UI | 系统级 |
| 8 | **sonar** | SonarQube Scanner（`withSonarQubeEnv` / `waitForQualityGate`） | 全部（可选启用） |
| 9 | **timestamper** | 构建日志添加时间戳 | 全部 |
| 10 | **ws-cleanup** | 工作空间清理（`deleteDir()` 步骤） | Go / Python |
| 11 | **http_request** | HTTP 请求步骤（`httpRequest()`），用于阶段回调和最终回调 | 全部 |
| 12 | **pipeline-utility-steps** | Pipeline 工具步骤（`readJSON`、`writeFile`、`readFile` 等） | Go / Java |
| 13 | **junit** | JUnit 测试报告解析（`junit` 步骤） | Java |

### 插件分类说明

| 分类 | 插件 | 说明 |
|------|------|------|
| **构建核心** | kubernetes, workflow-aggregator, credentials-binding | Pipeline 执行必须 |
| **代码管理** | git | 源码拉取 |
| **CI 功能** | http_request, pipeline-utility-steps, junit | 回调通信、数据解析、测试报告 |
| **质量扫描** | sonar | SonarQube 集成（可选） |
| **UI/体验** | blueocean, pipeline-stage-view, timestamper | 可视化和日志 |
| **运维辅助** | configuration-as-code, ws-cleanup | 自动配置和清理 |

### 插件安装命令（手动安装参考）

如果不使用 K8s 部署的 initContainer 方式，可在 Jenkins CLI 手动安装：

```bash
jenkins-plugin-cli --plugins \
  kubernetes \
  workflow-aggregator \
  git \
  configuration-as-code \
  credentials-binding \
  pipeline-stage-view \
  blueocean \
  sonar \
  timestamper \
  ws-cleanup \
  http_request \
  pipeline-utility-steps \
  junit
```

> **重要**：`http_request` 插件是所有模板回调平台的核心依赖，缺少此插件会导致 `httpRequest()` 步骤报错。

---

## 八、凭证（Credentials）配置清单

在 Jenkins 中需要创建以下 **3 个凭证**，供所有 Pipeline 模板使用：

### 8.1 凭证总览

| # | 凭证 ID | 类型 | 用途 | 作用域 |
|---|---------|------|------|--------|
| 1 | `gitee-id` | Username with password | Git 源码仓库认证（Gitee/GitLab/GitHub） | Global |
| 2 | `harbor-registry` | Username with password | Docker 镜像仓库推送/拉取认证（Harbor） | Global |
| 3 | `hmac-secret` | Secret text | Webhook HMAC 签名验证（平台回调安全校验） | Global |

### 8.2 凭证创建步骤

#### 凭证 1：`gitee-id`（Git 仓库凭证）

```
路径：Manage Jenkins → Credentials → System → Global credentials → Add Credentials

Kind:        Username with password
Scope:       Global
Username:    <你的 Git 用户名>
Password:    <你的 Git 密码或 Access Token>
ID:          gitee-id
Description: Git 源码仓库凭证（Gitee）
```

> **说明**：所有 Pipeline 模板中 `checkout` 阶段使用此凭证拉取代码。如使用 GitLab/GitHub，ID 保持 `gitee-id` 不变（或在 `config.yaml` 中修改 `GitCredentialID` 字段）。

#### 凭证 2：`harbor-registry`（镜像仓库凭证）

```
路径：Manage Jenkins → Credentials → System → Global credentials → Add Credentials

Kind:        Username with password
Scope:       Global
Username:    <Harbor 用户名>
Password:    <Harbor 密码>
ID:          harbor-registry
Description: Harbor 镜像仓库凭证
```

> **说明**：Kaniko 构建镜像后推送到 Harbor 时使用。模板中会自动生成 `/kaniko/.docker/config.json` 进行认证。

#### 凭证 3：`hmac-secret`（HMAC 签名凭证）

```
路径：Manage Jenkins → Credentials → System → Global credentials → Add Credentials

Kind:        Secret text
Scope:       Global
Secret:      <与 config.yaml 中 HMACSecret 保持一致的密钥>
ID:          hmac-secret
Description: HMAC signing secret for platform callback
```

> **说明**：Pipeline 构建完成后回调平台 API 时，使用此密钥对请求体做 HMAC-SHA256 签名，平台端验证防止伪造回调。

### 8.3 凭证 ID 可配置

凭证 ID 在平台后端 `configs/config.yaml` 中可自定义覆盖：

```yaml
Jenkins:
  GitCredentialID: "gitee-id"           # 默认 Git 凭证 ID
  RegistryCredentialID: "harbor-registry" # 默认镜像仓库凭证 ID
  HMACCredentialID: "hmac-secret"        # 默认 HMAC 凭证 ID
```

如果你在 Jenkins 中使用了不同的凭证 ID 名称，只需同步修改 `config.yaml` 即可。

---

## 九、JCasC 自动化配置（Configuration as Code）

Jenkins 启动时通过 ConfigMap 挂载 JCasC 配置文件，自动完成以下配置：

### 9.1 管理员账户

```yaml
securityRealm:
  local:
    allowsSignup: false
    users:
      - id: "ops-dev"
        password: "${JENKINS_ADMIN_PASSWORD}"  # 来自 K8s Secret
```

| 配置项 | 值 |
|--------|-----|
| 用户名 | `ops-dev` |
| 密码 | 通过 K8s Secret `jenkins-secret` 注入 |
| 注册 | 禁用自助注册 |

### 9.2 Kubernetes Cloud 配置

```yaml
clouds:
  - kubernetes:
      name: "kubernetes"
      serverUrl: "https://kubernetes.default.svc.cluster.local"
      namespace: "devops"
      jenkinsUrl: "http://jenkins.devops.svc.cluster.local:8080"
      jenkinsTunnel: "jenkins-agent.devops.svc.cluster.local:50000"
      connectTimeout: 5
      readTimeout: 15
      containerCapStr: "20"         # 最大并发 Pod Agent 数
      maxRequestsPerHostStr: "32"
      retentionTimeout: 5
      waitForPodSec: 600            # Pod 启动超时（秒）
```

| 配置项 | 说明 |
|--------|------|
| K8s API | 使用集群内部 DNS |
| Jenkins URL | 集群内 Service 地址 |
| Jenkins Tunnel | Agent 通信地址（50000 端口） |
| 并发上限 | 20 个 Pod 同时构建 |
| Pod 超时 | 600 秒（镜像拉取可能较慢） |

### 9.3 全局安全配置

```yaml
authorizationStrategy:
  loggedInUsersCanDoAnything:
    allowAnonymousRead: false       # 禁止匿名访问
```

### 9.4 Jenkins URL

```yaml
unclassified:
  location:
    url: "http://jenkins.devops.svc.cluster.local:8080/"
```

---

## 十、RBAC 权限配置（K8s 侧）

Jenkins 需要在 K8s 集群中创建/管理构建 Pod，因此需要以下权限：

### 10.1 ServiceAccount

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: jenkins
  namespace: devops
```

### 10.2 ClusterRole 权限清单

| 资源 | 操作 | 用途 |
|------|------|------|
| `pods` | create, delete, get, list, watch, patch | 动态 Agent Pod 生命周期管理 |
| `pods/exec` | create, get | 在 Pod 容器中执行命令 |
| `pods/log` | get, list, watch | 读取构建日志 |
| `events` | get, list, watch | 调试排错 |
| `secrets` | get, list | 拉取镜像凭证 |
| `persistentvolumeclaims` | get, list, create | 构建缓存 PVC |

---

## 十一、持久化存储（PVC）清单

| # | PVC 名称 | 容量 | 用途 |
|---|----------|------|------|
| 1 | `jenkins-data` | 20Gi | Jenkins 主数据（配置、历史、插件） |
| 2 | `jenkins-go-cache` | 10Gi | Go 模块缓存 (`/go/pkg/mod`) |
| 3 | `jenkins-maven-cache` | 20Gi | Maven 本地仓库 (`/root/.m2/repository`) |
| 4 | `jenkins-npm-cache` | 10Gi | npm 缓存 (`/root/.npm`) |
| 5 | `jenkins-pip-cache` | 5Gi | pip 缓存 (`/root/.cache/pip`) |

> 构建缓存 PVC 实现跨构建共享依赖，大幅加速第二次及后续构建。

---

## 十二、Pipeline Job 创建清单

需在 Jenkins 中创建 **4 个 Pipeline Job**：

| # | Job 名称 | Script Path | 语言 | 说明 |
|---|----------|-------------|------|------|
| 1 | `k8s-builder-go` | `configs/jenkins-templates/go-pipeline.groovy` | Go | Go 项目通用构建 |
| 2 | `k8s-builder-java` | `configs/jenkins-templates/java-spring-pipeline.groovy` | Java | Java/Spring Boot 通用构建 |
| 3 | `k8s-builder-frontend` | `configs/jenkins-templates/frontend-pipeline.groovy` | 前端 | Vue/React/Angular 通用构建 |
| 4 | `k8s-builder-python` | `configs/jenkins-templates/python-pipeline.groovy` | Python | Flask/FastAPI/Django 通用构建 |

### 12.1 Job 创建步骤（每个 Job 重复以下操作）

```
1. Jenkins → New Item
2. 输入 Job 名称（如 k8s-builder-java）
3. 选择 "Pipeline" 类型 → OK
4. Pipeline 区域：
   - Definition: Pipeline script from SCM
   - SCM: Git
   - Repository URL: <本项目 Git 仓库地址>
   - Credentials: gitee-id
   - Branch Specifier: */main
   - Script Path: configs/jenkins-templates/java-spring-pipeline.groovy
5. 保存
```

### 12.2 各模板构建容器说明

| Job | 构建容器 | 镜像 | 用途 |
|-----|----------|------|------|
| Go | `golang` | `golang:1.24` | Go 编译 |
| Java | `maven` | `maven:3.9-eclipse-temurin-{17/21/11/8}` | Maven 构建（JDK 版本动态） |
| Frontend | `node` | `node:18-alpine` | npm 构建 |
| Python | `python` | `python:3.11-slim` | pip + pytest |
| 所有 | `kaniko` | `gcr.io/kaniko-project/executor:debug` | 镜像构建（无需 Docker） |

---

## 十三、K8s Secret 配置

Jenkins 敏感信息通过 K8s Secret 注入：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: jenkins-secret
  namespace: devops
type: Opaque
data:
  admin-password: <base64编码的管理员密码>       # echo -n "your-password" | base64
  hmac-secret: <base64编码的HMAC密钥>            # echo -n "your-hmac-key" | base64
  sonarqube-token: <base64编码的SonarQube Token>  # echo -n "squ_xxxxx" | base64（可选，启用 SonarQube 时需要）
```

---

## 十四、平台后端对接配置（config.yaml）

平台后端需配置 Jenkins 连接信息：

```yaml
Jenkins:
  URL: "http://jenkins.devops.svc.cluster.local:8080/"
  Username: "ops-dev"                            # Jenkins 管理员用户名
  APIToken: "xxxxxxxx"                           # Jenkins API Token（在 Jenkins 用户设置中生成）
  TriggerTimeout: 60                             # 触发构建超时（秒）
  CallbackURL: "http://k8soperation-backend.k8soperation.svc.cluster.local:38180"
  PlatformURL: "http://k8soperation-frontend.k8soperation.svc.cluster.local"
  GitCredentialID: "gitee-id"
  RegistryCredentialID: "harbor-registry"
  HMACCredentialID: "hmac-secret"
  HMACSecret: "xxxxx"                           # 与 Jenkins hmac-secret 凭证保持一致
  PollInterval: 15                               # 构建状态轮询间隔（秒）
  MaxBuildTime: 30                               # 最大构建时间（分钟）
  EnableDingTalk: true                           # 钉钉通知
  DingTalkWebhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
  EnableFeishu: false                            # 飞书通知
  FeishuWebhook: ""
  FeishuSecret: ""
```

### 14.1 API Token 生成方法

```
1. 登录 Jenkins → 右上角用户名 → Configure
2. API Token → Add new Token → 输入名称 → Generate
3. 复制 Token 填入 config.yaml 的 APIToken 字段
```

---

## 十五、SonarQube 集成配置（可选）

如果启用代码质量扫描（`ENABLE_SONAR=true`），需要配置 SonarQube 服务器连接。

### 15.1 SonarQube 插件 vs SonarQube 服务器（核心区分）

| 内容 | 安装/配置位置 | 作用 |
|------|-------------|------|
| `sonar` 插件 | **initContainers**（statefulset.yaml） | 为 Jenkins 提供 `withSonarQubeEnv` 和 `waitForQualityGate` Pipeline 步骤 |
| SonarQube 服务器地址 | **Jenkins 系统配置**（JCasC ConfigMap 或 UI） | 告诉 Jenkins「SonarQube 服务在哪里」 |

> **总结**：`initContainers` 只负责安装插件（让 Jenkins 会说“SonarQube 语言”），**实际地址和认证**通过 JCasC 或 Jenkins UI 配置。

### 15.2 部署拓扑选择（集群内/外部均可）

| 部署方式 | SonarQube 地址示例 | 是否支持 |
|----------|-------------------|:---:|
| **集群外部独立部署**（推荐） | `http://sonarqube.example.com:9000` | ✅ |
| 集群内部 Pod 部署 | `http://sonarqube.devops.svc.cluster.local:9000` | ✅ |
| Docker Compose 独立部署 | `http://192.168.1.100:9000` | ✅ |
| SonarCloud（云端） | `https://sonarcloud.io` | ✅ |

> **结论：外部 SonarQube 完全可以**，只要 Jenkins Pod 能通过网络访问到目标地址即可。

### 15.3 网络链路说明

```
Jenkins Pod (devops namespace)
    │
    ├── 构建 Pod 中执行 sonar-scanner / mvn sonar:sonar
    │       │
    │       ▼ HTTP/HTTPS 请求
    │   ┌──────────────────────────────────────┐
    │   │  SonarQube Server                    │
    │   │  （集群内 / 集群外 / 云端均可）          │
    │   └──────────────────────────────────────┘
    │
    └── Quality Gate Webhook 回调（可选）
            SonarQube → Jenkins Webhook URL
```

如果使用外部 SonarQube，需确保：
- Jenkins Pod 所在节点能访问外部 SonarQube 的 IP/域名 + 端口（默认 9000）
- 如有 NetworkPolicy，需开放 devops namespace 到外部的出站访问
- 如需 Quality Gate Webhook，SonarQube 需能回调 Jenkins（可通过 NodePort/Ingress 暴露）

### 15.4 JCasC 自动配置方式（推荐）

本项目已在 `deploy/jenkins/configmap.yaml` 中通过 JCasC 自动配置 SonarQube 服务器：

```yaml
# configmap.yaml 中的 JCasC 配置片段
unclassified:
  sonarGlobalConfiguration:
    buildWrapperEnabled: true
    installations:
      - name: "SonarQube"                              # Pipeline 中 withSonarQubeEnv('SonarQube') 引用此名称
        serverUrl: "http://sonarqube.example.com:9000"  # ← 修改为你实际的 SonarQube 地址
        credentialsId: "sonarqube-token"                # ← 对应下方凭证 ID
        webhookSecretId: ""                             # Quality Gate Webhook Secret（可选）
```

对应凭证在同一 JCasC 中自动注册：

```yaml
credentials:
  system:
    domainCredentials:
      - credentials:
          - string:
              scope: GLOBAL
              id: "sonarqube-token"
              secret: "${SONAR_TOKEN}"     # 来自 K8s Secret 环境变量注入
              description: "SonarQube authentication token"
```

### 15.5 K8s Secret 和环境变量配置

`deploy/jenkins/secret.yaml` 中存储 SonarQube Token：

```yaml
data:
  sonarqube-token: <base64编码的Token>   # echo -n "squ_xxxxx" | base64
```

`deploy/jenkins/statefulset.yaml` 中注入环境变量：

```yaml
env:
  - name: SONAR_TOKEN
    valueFrom:
      secretKeyRef:
        name: jenkins-secret
        key: sonarqube-token
```

### 15.6 SonarQube Token 生成步骤

```
1. 登录 SonarQube Web UI（如 http://sonarqube.example.com:9000）
2. 右上角头像 → My Account → Security
3. Generate Tokens:
   - Name: jenkins-analysis
   - Type: Global Analysis Token（推荐）或 User Token
   - Expires in: 建议 90 天或 No expiration
4. 点击 Generate → 复制 Token（格式如 squ_xxxxxxxx）
5. Base64 编码后写入 K8s Secret：
   echo -n "squ_xxxxxxxx" | base64
```

### 15.7 手动 UI 配置方式（备选）

如果不使用 JCasC，也可在 Jenkins Web UI 手动配置：

```
步骤 1：创建凭证
  Manage Jenkins → Credentials → System → Global → Add Credentials
  Kind:        Secret text
  Scope:       Global
  Secret:      squ_xxxxxxxx（你的 SonarQube Token）
  ID:          sonarqube-token
  Description: SonarQube Token

步骤 2：配置服务器
  Manage Jenkins → System → SonarQube servers → Add SonarQube
  Name:                      SonarQube
  Server URL:                http://sonarqube.example.com:9000
  Server authentication token: sonarqube-token（选择上面创建的凭证）
```

### 15.8 Quality Gate Webhook（可选）

如果启用质量门禁（`SONAR_QUALITY_GATE=true`），需要在 SonarQube 中配置 Webhook 回调 Jenkins：

```
1. SonarQube → Administration → Configuration → Webhooks → Create
2. Name: Jenkins
3. URL:
   - 集群内 SonarQube: http://jenkins.devops.svc.cluster.local:8080/sonarqube-webhook/
   - 集群外 SonarQube: http://<节点IP>:30080/sonarqube-webhook/（通过 NodePort）
4. Secret: （可选，对应 JCasC 中的 webhookSecretId）
```

### 15.9 验证清单

| # | 检查项 | 验证方式 |
|---|--------|----------|
| 1 | 插件已安装 | Jenkins → Manage Plugins → Installed → 搜索 "SonarQube Scanner" |
| 2 | 服务器已配置 | Manage Jenkins → System → SonarQube servers → 名称为 "SonarQube" |
| 3 | Token 有效 | 手动触发带 `ENABLE_SONAR=true` 的构建，检查 SonarQube Analysis 阶段日志 |
| 4 | 网络可达 | 在 Jenkins Pod 中 `curl http://sonarqube.example.com:9000/api/system/status` |
| 5 | Webhook 回调 | 检查 SonarQube → Administration → Webhooks → 最近投递记录 |

---

## 十六、网络端口清单

| 服务 | 端口 | 用途 |
|------|------|------|
| Jenkins Web UI | 8080 (ClusterIP) / 30080 (NodePort) | Web 界面和 API |
| Jenkins Agent | 50000 (ClusterIP) | JNLP Agent 通信 |

### Service 拓扑

```
jenkins (ClusterIP) ─────── 8080/50000 ──→ Jenkins Master Pod
jenkins-agent (ClusterIP) ─ 50000 ────────→ Jenkins Master Pod（Agent Tunnel）
jenkins-nodeport (NodePort) 30080 ────────→ Jenkins Master Pod（外部访问）
```

---

## 十七、一键部署命令

```bash
# 部署到 K8s 集群
kubectl apply -f deploy/jenkins/

# 验证部署状态
kubectl -n devops get pods -l app.kubernetes.io/name=jenkins
kubectl -n devops get svc -l app.kubernetes.io/name=jenkins

# 查看插件安装日志
kubectl -n devops logs jenkins-0 -c install-plugins

# 查看 Jenkins 启动日志
kubectl -n devops logs jenkins-0 -c jenkins -f
```

---

## 十八、部署后检查清单

| # | 检查项 | 命令/操作 | 预期结果 |
|---|--------|-----------|----------|
| 1 | Pod Running | `kubectl -n devops get pod jenkins-0` | STATUS = Running |
| 2 | 插件已装 | Jenkins UI → Manage Plugins → Installed | 13 个插件全部存在 |
| 3 | K8s Cloud | Manage Jenkins → Clouds → Kubernetes | 连接测试成功 |
| 4 | 凭证就绪 | Manage Jenkins → Credentials | 3 个凭证（gitee-id, harbor-registry, hmac-secret） |
| 5 | Job 创建 | Dashboard | 4 个 Pipeline Job 存在 |
| 6 | 试跑构建 | 手动触发 k8s-builder-go（填写测试参数） | Pod Agent 自动创建并执行 |
| 7 | 平台联通 | 平台 CICD 页面创建流水线并触发 | 构建状态回调正常 |

---

## 十九、故障排查

| 问题 | 排查方向 |
|------|----------|
| Pod Agent 创建失败 | 检查 RBAC、ServiceAccount 是否正确绑定 |
| 拉取代码失败 | 检查 `gitee-id` 凭证用户名密码是否正确 |
| 镜像推送失败 | 检查 `harbor-registry` 凭证、Harbor 地址是否可达 |
| 回调失败（签名错误） | 确认 `hmac-secret` 凭证与 config.yaml 中 `HMACSecret` 一致 |
| 构建超时 | 检查 PVC 是否正常绑定、镜像拉取是否可达（国内需配镜像代理） |
| JCasC 不生效 | 检查 JAVA_OPTS 中 `casc.jenkins.config` 路径与 ConfigMap 挂载一致 |
| SonarQube 扫描失败 | 检查 SonarQube 服务器地址是否可达、Token 是否有效（`curl <sonar-url>/api/system/status`） |
| Quality Gate 超时 | 检查 SonarQube Webhook 是否配置正确、是否能回调 Jenkins |

---

## 二十、JVM 参数配置

```
-Xms512m -Xmx2g                          # 内存配置
-Djenkins.install.runSetupWizard=false    # 跳过安装向导（JCasC 管理）
-Dcasc.jenkins.config=/var/jenkins_home/casc_configs  # JCasC 配置路径
```

---

## 附录：文件对照表

| 配置文件 | 路径 | 说明 |
|----------|------|------|
| StatefulSet | `deploy/jenkins/statefulset.yaml` | Jenkins 主部署 + 插件安装 |
| ConfigMap (JCasC) | `deploy/jenkins/configmap.yaml` | 自动化配置 |
| Secret | `deploy/jenkins/secret.yaml` | 管理员密码 + HMAC 密钥 + SonarQube Token |
| RBAC | `deploy/jenkins/rbac.yaml` | ServiceAccount + ClusterRole |
| PVC | `deploy/jenkins/pvc.yaml` | 持久化存储（数据 + 缓存） |
| Service | `deploy/jenkins/service.yaml` | 网络服务暴露 |
| Pipeline 模板（Go） | `configs/jenkins-templates/go-pipeline.groovy` | Go 构建模板 |
| Pipeline 模板（Java） | `configs/jenkins-templates/java-spring-pipeline.groovy` | Java 构建模板 |
| Pipeline 模板（Frontend） | `configs/jenkins-templates/frontend-pipeline.groovy` | 前端构建模板 |
| Pipeline 模板（Python） | `configs/jenkins-templates/python-pipeline.groovy` | Python 构建模板 |
| 平台配置 | `configs/config.yaml` | 后端 Jenkins 连接参数 |
