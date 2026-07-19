# Jenkins 完整配置、插件与架构说明文档

> 本文档适用于 K8s Operation Platform 的 Jenkins 全容器化部署方案。  
> Jenkins 运行在 K8s 集群 `devops` 命名空间，使用动态 Pod Agent 构建，无需在 Master 上安装任何编译工具。

---

## 快速安装顺序（生产环境参考）

```
步骤 1  修改 deploy/jenkins/secret.yaml（改密码 / HMAC 密钥 / SonarQube Token）
步骤 2  修改 deploy/jenkins/configmap.yaml（改 SonarQube 服务器地址）
步骤 3  kubectl apply -k deploy/jenkins/   ← 一键部署全部资源
步骤 4  等待 initContainer 安装插件（约 3-5 分钟）
步骤 5  浏览器访问 http://<节点IP>:30080，用 ops-dev 登录
步骤 6  手动添加 Git 凭证（k8soperation）和 Harbor 凭证（robot$test-k8soperation）
步骤 7  创建 4 个 Pipeline Job（关联 Git 仓库 + 指定 Script Path）
步骤 8  生成 API Token → 填入 configs/config.yaml 的 APIToken 字段
步骤 9  重启后端服务使配置生效
```

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

## 七、插件安装完整指南

### 7.0 两种安装方式详细对比

> **问：可以只用 initContainer 安装，或者只用 Jenkins UI 安装吗？**
> **答：两种方式都可以，全部完成安装，但适用场景和优缺点不同。**

```
插件安装方式
│
├── 方式一：initContainer 安装 ← K8s 部署时自动执行（推荐）
│                          在 statefulset.yaml 中定义插件列表
│                          Pod 启动 → initContainer 先运行 → 安装完成 → Jenkins 启动
│
└── 方式二：Jenkins UI 安装 ← Jenkins 运行中手动操作
                           Manage Jenkins → Plugins → Available plugins
```

#### 详细对比表

| 对比项 | initContainer 安装 | Jenkins UI 安装 |
|--------|------------------|----------------|
| **执行时机** | Jenkins 启动前，Pod 创建时自动运行 | Jenkins 运行中手动操作 |
| **操作方式** | `jenkins-plugin-cli` 命令行 | 网页点击安装 |
| **是否自动** | ✅ 全自动，无需人工 | ❌ 需手动操作 |
| **可重现性** | ✅ 嵌入代码，重建 Pod 自动重装 | ❌ 不在代码里，PVC 丢失后需手动重装 |
| **生效展题** | ✅ Jenkins 启动即就就绪 | ⚠️ 需重启 Jenkins 才能生效 |
| **适用场景** | ✅ 生产环境、标准化部署 | ✅ 临时补装、应急测试 |
| **插件来源** | Jenkins 插件仓库（需网络） | Jenkins 插件仓库（需网络） |
| **PVC 丢失后** | ✅ 重建不同名 Pod 自动重装 | ❌ 插件全部丢失需手动重装 |
| **网络中断** | ⚠️ 下载中断不完整会留下 `.jpi.tmp` 文件 | ⚠️ 同上 |

#### 两种方式均可完成安装圆满，建议策略

```
生产环境：  initContainer 为主 + UI 补充（应急用）
暴走工具：  两种都可以，哪个方便用哪个
全新安装：  initContainer 自动完成，无需手动操作
重装 Jenkins：  修改 statefulset.yaml + 重建 Pod ，自动安装
```

#### 已知问题：sonar 插件下载中断留下 .tmp 文件

`jenkins-plugin-cli` 在下载过程中网络中断会留下未完成的 `.jpi.tmp` 临时文件，导致下次安装时 `jenkins-plugin-cli` 误认已安装而跳过。

```bash
# 现象：jenkins-plugin-cli --plugins sonar 返回 Done 但 sonar.jpi 不存在
# 原因：存在 sonar.jpi.tmp （下载未完成）

# 修复方法：清理 .tmp 文件并重新安装
kubectl -n devops exec jenkins-0 -- sh -c \
  "rm -f /var/jenkins_home/plugins/sonar.jpi.tmp && jenkins-plugin-cli --plugins sonar"

# 验证：
kubectl -n devops exec jenkins-0 -- sh -c "ls /var/jenkins_home/plugins/ | grep sonar"
# 预期输出: sonar  sonar.jpi  sonar-quality-gates  sonar-quality-gates.jpi
```

> 根本解决：`rollout restart` 重建 Pod 时，initContainer 会清除 .tmp 文件重新下载。

---

### 7.1 必装插件清单（13 个）

以下插件通过 `initContainer` 中的 `jenkins-plugin-cli` 自动安装（**不指定版本号**，安装最新兼容版）：

> ⚠️ 注意：`blueocean` 依赖树过大（100+ 子依赖），已从列表移除，避免 init container 安装失败。

| #  | 插件 ID（安装名） | 插件全称 | 用途说明 | 被哪些模板使用 |
|----|---------|---------|----------|----------------|
| 1  | `kubernetes` | Kubernetes | K8s Cloud 集成，动态创建 Pod Agent 执行构建 | 全部 |
| 2  | `workflow-aggregator` | Pipeline (Workflow Aggregator) | Pipeline 核心语法（`pipeline{}`、`stage`、`steps` 等） | 全部 |
| 3  | `git` | Git | Git SCM 支持，拉取源代码（`checkout` 步骤） | 全部 |
| 4  | `configuration-as-code` | Configuration as Code (JCasC) | 通过 YAML 自动化配置 Jenkins（K8s Cloud、凭证等） | 系统级 |
| 5  | `credentials-binding` | Credentials Binding | Pipeline 中 `credentials()` 凭证注入 | 全部 |
| 6  | `http_request` | HTTP Request | `httpRequest()` 步骤，用于阶段回调和最终回调 | 全部 |
| 7  | `pipeline-utility-steps` | Pipeline Utility Steps | `readJSON`、`writeFile`、`fileExists` 等工具步骤 | Go / Java |
| 8  | `junit` | JUnit | `junit` 步骤解析测试报告（`**/surefire-reports/*.xml`） | Java |
| 9  | `sonar` | SonarQube Scanner | `withSonarQubeEnv()` + `waitForQualityGate()`（可选启用） | 全部（可选） |
| 10 | `pipeline-stage-view` | Pipeline: Stage View | 流水线阶段可视化图，构建历史看板 | 系统级 |
| 11 | `timestamper` | Timestamper | 构建日志每行添加时间戳 | 全部 |
| 12 | `ws-cleanup` | Workspace Cleanup | 工作空间清理（`cleanWs()` / `deleteDir()` 步骤） | 全部 |
| 13 | `ansicolor` | AnsiColor | 日志支持 ANSI 彩色输出（Maven/Go 构建日志更易读） | 全部 |
| 14 | `throttle-concurrents` | Throttle Concurrent Builds | 限制 Jenkins 构建并发数量，避免多个任务同时执行导致资源竞争 | 全部 |


### 7.2 插件分类说明

| 分类 | 插件 | 说明 |
|------|------|------|
| **构建核心** | kubernetes, workflow-aggregator, credentials-binding | Pipeline 执行必须 |
| **代码管理** | git | 源码拉取 |
| **CI 功能** | http_request, pipeline-utility-steps, junit | 回调通信、数据解析、测试报告 |
| **质量扫描** | sonar | SonarQube 集成（ENABLE_SONAR=false 可跳过） |
| **UI/体验** | pipeline-stage-view, timestamper, ansicolor | 可视化和日志优化 |
| **运维辅助** | configuration-as-code, ws-cleanup | 自动配置和清理 |

### 7.3 不需要安装的插件

> 本项目全容器化架构，以下常见插件**一律不需要安装**：

| 不需要安装 | 原因 |
|-----------|------|
| Docker Plugin / Docker Pipeline | 构建使用 Kaniko，完全不需要 Docker daemon |
| Maven Integration Plugin | Maven 运行在 K8s Pod 容器（`maven:3.9-eclipse-temurin-17`）内，不依赖 Jenkins 工具链 |
| NodeJS Plugin | Node/npm 运行在 Pod 容器（`node:18-alpine`）内，不依赖 Jenkins 工具链 |
| JDK Tool / Global Tool 配置 | JDK 通过容器镜像提供，不注册 Jenkins 全局工具 |
| Email Extension Plugin | 平台通过 Webhook 发通知，Jenkins 侧不需要 |
| Blue Ocean | 依赖树过大（100+插件），已从安装列表移除，推荐用 Stage View 替代 |
| Publish Over SSH | 部署通过 K8s API，不需要 SSH 推送 |
| Deploy to container Plugin | 部署通过 kubectl/Helm，不需要此插件 |

### 7.4 安装操作步骤（三种方式）

#### 方式 A：initContainer 自动安装（推荐，K8s 部署时自动执行）

在 `deploy/jenkins/statefulset.yaml` 的 initContainer 中已配置，Pod 启动时自动安装：

```yaml
command:
  - jenkins-plugin-cli
  - --plugins
  - >
    kubernetes
    workflow-aggregator
    git
    configuration-as-code
    credentials-binding
    http_request
    pipeline-utility-steps
    junit
    sonar
    pipeline-stage-view
    timestamper
    ws-cleanup
    ansicolor
```

> **重要**：不指定版本号，由 `jenkins-plugin-cli` 自动解析最新兼容版本。指定版本号易因依赖冲突导致 CrashLoopBackOff。

> ⚠️ **YAML 格式坑：必须用 `>-` 而非 `>`**  
> `>` 会在最后一个插件名后保留换行符 `\n`，导致 URL 变成 `ansicolor\n.hpi`，报错 `Illegal character in path`。  
> `>-` 折叠并去掉末尾换行，插件名干净正确。这是 **已知 Bug，需特别注意**。

#### 方式 B：Jenkins UI 手动安装（应急补装）

若 initContainer 未安装成功，进入 **Manage Jenkins → Plugins → Available plugins** 搜索安装：

```
git
http_request
pipeline-utility-steps
configuration-as-code
junit
sonar
pipeline-stage-view
timestamper
ws-cleanup
ansicolor
```

安装完成后勾选 **"Restart Jenkins when installation is complete"** 自动重启。

#### 方式 C：命令行补装（在运行中的 Pod 内执行）

```bash
# 利用 jenkins-plugin-cli 在运行中的 Pod 内直接补装
kubectl -n devops exec jenkins-0 -- sh -c \
  "jenkins-plugin-cli --plugins git http_request pipeline-utility-steps junit sonar pipeline-stage-view timestamper ws-cleanup ansicolor"

# 如果某个插件需要清除 .tmp 
# 先清除未完成的临时文件再重装
kubectl -n devops exec jenkins-0 -- sh -c \
  "rm -f /var/jenkins_home/plugins/sonar.jpi.tmp && jenkins-plugin-cli --plugins sonar"

# 安装完成后重启 Pod 让 Jenkins 加载新插件
kubectl -n devops rollout restart statefulset/jenkins
kubectl -n devops get pod -w
```

#### 方式 D：重建 Pod 强制重新安装（最干净）

```bash
# 应用最新 StatefulSet 配置
kubectl apply -f deploy/jenkins/statefulset.yaml

# 删除 Pod 让 StatefulSet 重建（自动触发 initContainer）
kubectl -n devops delete pod jenkins-0

# 或者用 rollout restart
kubectl -n devops rollout restart statefulset/jenkins

# 观察进度（约 3~5 分钟）
kubectl -n devops get pod -w
```

#### 验证插件安装情况

```bash
# 查看已安装的插件文件
kubectl -n devops exec jenkins-0 -- sh -c \
  "ls /var/jenkins_home/plugins/ | grep '.jpi$' | sed 's/.jpi//' | sort"
```

---

## 八、凭证（Credentials）配置清单

在 Jenkins 中需要创建以下 **3 个凭证**，供所有 Pipeline 模板使用：

### 8.1 凭证总览

> 凭证 ID 需与 `configs/config.yaml` 中的 Jenkins 配置块保持一致。当前项目实际配置如下：

| # | 凭证 ID（当前配置） | 类型 | 用途 | 作用域 | config.yaml 对应字段 |
|---|---------|------|------|--------|------|
| 1 | `k8soperation` | Username with password | Git 源码仓库认证（Gitee/GitLab/GitHub） | Global | `GitCredentialID` |
| 2 | `robot$test-k8soperation` | Username with password | Harbor 镜像仓库推送/拉取认证 | Global | `RegistryCredentialID` |
| 3 | `hmac-secret` | Secret text | Webhook HMAC 签名验证（平台回调安全校验） | Global | `HMACCredentialID` |
| 4 | `sonarqube-token` | Secret text | SonarQube 认证 Token（JCasC 自动注入） | Global | — |

> **凭证 1 & 2**：需手动添加。**凭证 3 & 4**：由 JCasC 从 K8s Secret 自动注入，无需手动创建。

### 8.2 凭证创建步骤

#### 凭证 1：`k8soperation`（Git 仓库凭证）

```
路径：Manage Jenkins → Credentials → System → Global credentials → Add Credentials

Kind:        Username with password
Scope:       Global
Username:    <你的 Gitee/GitHub 账号>
Password:    <你的 Git 密码或 Access Token>
ID:          k8soperation
Description: Git 源码仓库凭证（Gitee）
```

> **说明**：所有 Pipeline 模板中 `checkout` 阶段使用此凭证拉取代码。ID 可在 `config.yaml` 的 `GitCredentialID` 字段自定义修改。

#### 凭证 2：`robot$test-k8soperation`（Harbor 镜像仓库凭证）

```
路径：Manage Jenkins → Credentials → System → Global credentials → Add Credentials

Kind:        Username with password
Scope:       Global
Username:    robot$test-k8soperation
Password:    <Harbor 机器人账号密码>
ID:          robot$test-k8soperation
Description: Harbor 镜像仓库凭证（机器人账号）
```

> **说明**：Kaniko 构建镜像后推送到 Harbor 时使用。模板中会自动生成 `/kaniko/.docker/config.json` 进行认证。ID 可在 `config.yaml` 的 `RegistryCredentialID` 字段修改。

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
  GitCredentialID: "k8soperation"              # 当前 Git 凭证 ID
  RegistryCredentialID: "robot$test-k8soperation" # 当前镜像仓库凭证 ID
  HMACCredentialID: "hmac-secret"              # 当前 HMAC 凭证 ID
```

> 如果你在 Jenkins 中使用了不同的凭证 ID 名称，只需同步修改 `config.yaml` 即可，Pipeline 模板无需改动。

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

### 9.1.1 修改管理员用户名和密码

用户名和密码分别保存在两个文件中，**两者相互独立，可单独修改**：

| 配置项 | 配置文件 | 当前值 | 说明 |
|--------|---------|--------|------|
| 用户名 | `deploy/jenkins/configmap.yaml` 第 24 行 `id:` | `ops-dev` | JCasC 创建的 Jenkins 登录账号 |
| 密码 | `deploy/jenkins/secret.yaml` 第 17 行 `admin-password:` | `admin123`（base64编码） | K8s Secret 注入到容器环境变量 |

#### 修改用户名

编辑 `deploy/jenkins/configmap.yaml`：

```yaml
# deploy/jenkins/configmap.yaml
securityRealm:
  local:
    allowsSignup: false
    users:
      - id: "admin"          # ← 改成你想要的用户名（如 admin、jenkins 等）
        password: "${JENKINS_ADMIN_PASSWORD}"
```

#### 修改密码

密码必须先 base64 编码后填入 `deploy/jenkins/secret.yaml`：

```bash
# 第一步：生成新密码的 base64 编码
echo -n "你的新密码" | base64
# 例：echo -n "MyPass2024!" | base64  → 输出: TXlQYXNzMjAyNCE=
```

编辑 `deploy/jenkins/secret.yaml`：

```yaml
# deploy/jenkins/secret.yaml
data:
  admin-password: TXlQYXNzMjAyNCE=   # ← 替换为上面 base64 输出的值
  hmac-secret: bXlfc3VwZXJfc2VjcmV0X2htYWNfa2V5XzIwMjQ=
  sonarqube-token: eW91ci1zb25hcnF1YmUtdG9rZW4=
```

> ⚠️ **注意**：`secret.yaml` 中只存 base64 编码值，**不能写明文密码**，否则 K8s 会报错。

#### 修改后重新部署生效

```bash
# 应用 ConfigMap 和 Secret 更新
kubectl apply -f deploy/jenkins/configmap.yaml
kubectl apply -f deploy/jenkins/secret.yaml

# 重启 Jenkins Pod（让 JCasC 重新加载配置）
kubectl -n devops rollout restart statefulset/jenkins

# 查看启动状态
kubectl -n devops get pod -w
# 等待 STATUS 变为 Running（约 3~5 分钟）
```

> **重启后**：使用新用户名+新密码重新登录 Jenkins。  
> **注意**：修改用户名后，`configs/config.yaml` 中的 `Username` 字段也需要同步更新：
> ```yaml
> Jenkins:
>   Username: "admin"   # ← 与 configmap.yaml 中 id 保持一致
> ```

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

| # | PVC 名称 | 容量 | 挂载路径 | 用途 | 生产建议容量 |
|---|----------|------|---------|------|------|
| 1 | `jenkins-data` | 20Gi | `/var/jenkins_home` | Jenkins 主数据（配置、历史、插件） | 50Gi |
| 2 | `jenkins-go-cache` | 10Gi | `/go/pkg/mod` | Go 模块缓存 | 20Gi |
| 3 | `jenkins-maven-cache` | 20Gi | `/root/.m2/repository` | Maven 本地仓库 | 50Gi |
| 4 | `jenkins-npm-cache` | 10Gi | `/root/.cache/npm` | npm/前端缓存 | 20Gi |
| 5 | `jenkins-pip-cache` | 5Gi | `/root/.cache/pip` | pip 缓存 | 10Gi |

> 构建缓存 PVC 实现跨构建共享依赖，大幅加速第二次及后续构建。

### 11.1 PVC 扩容（数据满了不需要重启）

前提：StorageClass 需支持扩容（`ALLOWVOLUMEEXPANSION=true`）

```bash
# 查看 StorageClass 是否支持扩容
kubectl get storageclass

# 方式一：kubectl patch 直接扩容（最快）
kubectl patch pvc jenkins-maven-cache -n devops \
  -p '{"spec":{"resources":{"requests":{"storage":"50Gi"}}}}'

kubectl patch pvc jenkins-go-cache -n devops \
  -p '{"spec":{"resources":{"requests":{"storage":"20Gi"}}}}'

kubectl patch pvc jenkins-data -n devops \
  -p '{"spec":{"resources":{"requests":{"storage":"50Gi"}}}}'

# 方式二：修改 pvc.yaml 后 apply（推荐，配置可持久化）
# 修改 deploy/jenkins/pvc.yaml 中的 storage 值后：
kubectl apply -f deploy/jenkins/pvc.yaml
```

> **注意**：PVC 只能扩大，不能缩小。apply 后 K8s 会自动触发底层 PV 扩容，**不需要重启 Pod**。

### 11.2 缓存使用量查看

```bash
# 查看所有 Jenkins 相关 PVC 状态
kubectl get pvc -n devops

# 查看 Jenkins 主目录磁盘使用
kubectl exec -it jenkins-0 -n devops -- df -h /var/jenkins_home

# 查看 Maven 缓存大小
kubectl exec -it jenkins-0 -n devops -- du -sh /root/.m2/repository
```

### 11.3 缓存清理（满了先清再扩）

```bash
# 清理 Maven 超过 30 天未使用的依赖
kubectl exec -it jenkins-0 -n devops -- \
  find /root/.m2/repository -name "*.jar" -atime +30 -delete

# 清理 Go 模块缓存（谨慎，会影响下次构建速度）
kubectl exec -it jenkins-0 -n devops -- go clean -modcache
```

---

## 十二、Pipeline Job 创建清单

需在 Jenkins 中创建 **4 个 Pipeline Job**：

| # | Job 名称 | Script Path | 语言 | 说明 |
|---|----------|-------------|------|------|
| 1 | `go-pipeline` | `configs/jenkins-templates/go-pipeline.groovy` | Go | Go 项目通用构建 |
| 2 | `java-spring-pipeline` | `configs/jenkins-templates/java-spring-pipeline.groovy` | Java | Java/Spring Boot 通用构建 |
| 3 | `frontend-pipeline` | `configs/jenkins-templates/frontend-pipeline.groovy` | 前端 | Vue/React/Angular 通用构建 |
| 4 | `python-pipeline` | `configs/jenkins-templates/python-pipeline.groovy` | Python | Flask/FastAPI/Django 通用构建 |

### 12.1 Job 创建步骤（每个 Job 重复以下操作）

```
1. Jenkins → New Item
2. 输入 Job 名称（如 java-spring-pipeline）
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
| 6 | 试跑构建 | 手动触发 go-pipeline（填写测试参数） | Pod Agent 自动创建并执行 |
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
