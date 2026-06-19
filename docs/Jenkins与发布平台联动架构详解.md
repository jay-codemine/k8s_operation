# Jenkins + 发布平台联动架构详解

## 1. 架构总览

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          K8s 集群 (Kind v1.33)                           │
│                                                                          │
│  ┌─────────────────────────┐       ┌─────────────────────────────────┐  │
│  │   devops namespace      │       │    k8soperation namespace        │  │
│  │                         │       │                                  │  │
│  │  ┌───────────────────┐  │       │  ┌──────────┐   ┌──────────┐   │  │
│  │  │ Jenkins Master    │  │ HTTP  │  │ Backend  │   │ Frontend │   │  │
│  │  │ (StatefulSet)     │──────────│  │ (Deploy) │   │ (Deploy) │   │  │
│  │  │ - JCasC 自配置    │  │       │  │ - Go     │   │ - Vue3   │   │  │
│  │  │ - K8s Cloud      │  │       │  │ - API    │   │ - Nginx  │   │  │
│  │  └───────────────────┘  │       │  └──────────┘   └──────────┘   │  │
│  │           │              │       │       │                         │  │
│  │           ▼              │       │       │                         │  │
│  │  ┌───────────────────┐  │       │  ┌──────────┐                   │  │
│  │  │ Dynamic Pod Agent │  │       │  │  PVC     │                   │  │
│  │  │ (每次构建自动创建)  │  │       │  │ - logs   │                   │  │
│  │  │ - Golang/Node/...│  │       │  │ - artifacts│                  │  │
│  │  │ - Kaniko         │  │       │  └──────────┘                   │  │
│  │  └───────────────────┘  │       └─────────────────────────────────┘  │
│  └─────────────────────────┘                                             │
└──────────────────────────────────────────────────────────────────────────┘
```

### 工作流概览

```
用户(前端) → 平台后端 → Jenkins API(触发构建)
                                    │
                            Jenkins Pipeline 执行
                            │   各阶段 → 阶段回调 → 平台
                            │
                            构建完成 → 最终回调 → 平台
                                                    │
                                            自动部署 (Patch K8s)
                                                    │
                                            通知 (钉钉/飞书)
```

---

## 2. 核心联动机制

### 2.1 触发流程（平台 → Jenkins）

平台通过 Jenkins REST API 触发构建：

| 步骤 | 操作 | 说明 |
|------|------|------|
| 1 | 用户点击「运行流水线」 | 前端 POST `/api/v1/k8s/cicd/pipeline/run` |
| 2 | 平台组装参数 | 根据 language_type 自动注入凭证、回调地址、构建参数 |
| 3 | 调用 Jenkins API | `POST /job/{job_name}/buildWithParameters` |
| 4 | 获取 Queue ID | 从 Location Header 解析 `/queue/item/{id}/` |
| 5 | 轮询 Queue | 等待 Jenkins 分配 Build Number（超时 60s） |
| 6 | 记录 Build Number | 写入数据库，前端可查看进度 |

**平台自动注入的参数：**

```go
params := map[string]string{
    "GIT_REPO":               pipeline.GitRepo,
    "GIT_BRANCH":             run.GitBranch,
    "IMAGE_REPO":             imageRepo,
    "IMAGE_TAG":              imageTag,
    "PIPELINE_ID":            strconv.FormatInt(pipeline.ID, 10),
    "RUN_ID":                 strconv.FormatInt(run.ID, 10),
    "PLATFORM_CALLBACK_URL":  callbackURL + "/api/v1/k8s/cicd/pipeline/callback",
    "GIT_CREDENTIAL_ID":      "gitee-id",
    "REGISTRY_CREDENTIAL_ID": "harbor-registry",
    "HMAC_CREDENTIAL_ID":     "hmac-secret",
    "LANGUAGE_TYPE":          pipeline.LanguageType,
}
```

### 2.2 回调机制（Jenkins → 平台）

Jenkins 构建过程中有两类回调：

#### 阶段级回调（实时进度）

每个 Pipeline Stage 完成后调用：

```http
POST /api/v1/k8s/cicd/stage/callback
X-Signature: <HMAC-SHA256(secret, "job_name:build_number:stage_type")>
Content-Type: application/json

{
    "job_name": "k8s-builder-go",
    "build_number": 42,
    "pipeline_id": 15,
    "stage_type": "compile",
    "status": "success"
}
```

#### 最终回调（构建结果）

构建完成（成功/失败/中止）后调用：

```http
POST /api/v1/k8s/cicd/pipeline/callback
X-Signature: <HMAC-SHA256(secret, "job_name:build_number:status")>
Content-Type: application/json

{
    "job_name": "k8s-builder-go",
    "build_number": 42,
    "pipeline_id": 15,
    "status": "SUCCESS",
    "image_url": "harbor.example.com/project/app:main-abc1234-20250618",
    "image_digest": "sha256:...",
    "git_commit": "abc1234",
    "git_branch": "main",
    "duration_sec": 180,
    "build_url": "http://jenkins/job/k8s-builder-go/42/"
}
```

### 2.3 HMAC 安全签名

双端共享密钥，防止伪造回调：

```
┌─────────────────────┐          ┌─────────────────────────┐
│     Jenkins 端       │          │        平台端            │
│                     │          │                         │
│ credentials('hmac-  │          │ config.yaml:            │
│   secret')          │          │   HMACSecret: "xxx"     │
│                     │          │                         │
│ 签名: openssl dgst  │   HTTP   │ 验证: hmac.New(sha256)  │
│   -sha256 -hmac     │ ──────── │   + subtle.ConstantTime │
│                     │          │     Compare             │
└─────────────────────┘          └─────────────────────────┘
```

**签名规则：**
- 阶段回调: `HMAC-SHA256(secret, "job_name:build_number:stage_type")`
- 最终回调: `HMAC-SHA256(secret, "job_name:build_number:status")`

---

## 3. Jenkins 需要的配置

### 3.1 必装插件

| 插件 | 用途 |
|------|------|
| kubernetes | K8s 动态 Pod Agent |
| workflow-aggregator | Pipeline 核心 |
| git | Git SCM 支持 |
| configuration-as-code | JCasC 自动配置 |
| credentials-binding | 凭证绑定 |
| pipeline-stage-view | 阶段视图 |
| blueocean | Blue Ocean UI |
| sonar | SonarQube 代码扫描 |
| timestamper | 日志时间戳 |
| ws-cleanup | 工作空间清理 |

### 3.2 必要凭证（Credentials）

| Credential ID | 类型 | 用途 | 配置方式 |
|---------------|------|------|---------|
| `hmac-secret` | Secret text | 平台回调 HMAC 签名密钥 | JCasC 自动注入（从环境变量） |
| `harbor-registry` | Username/Password | 镜像仓库推送凭证 | Jenkins UI 手动创建 |
| `gitee-id` | Username/Password 或 SSH Key | Git 仓库拉取凭证 | Jenkins UI 手动创建 |

### 3.3 Kubernetes Cloud 配置

JCasC 自动配置（`deploy/jenkins/configmap.yaml`）：

```yaml
jenkins:
  numExecutors: 0          # Master 不执行构建
  mode: EXCLUSIVE
  clouds:
    - kubernetes:
        name: "kubernetes"
        serverUrl: "https://kubernetes.default.svc.cluster.local"
        namespace: "devops"
        jenkinsUrl: "http://jenkins.devops.svc.cluster.local:8080"
        jenkinsTunnel: "jenkins-agent.devops.svc.cluster.local:50000"
        containerCapStr: "20"        # 最大并发 Agent 数
        maxRequestsPerHostStr: "32"
        waitForPodSec: 600           # Agent Pod 启动超时
```

### 3.4 Jenkins Job 创建

每种语言对应一个通用 Job，通过 **Pipeline script from SCM** 方式配置：

| Job 名称 | Script Path | 服务项目 |
|----------|-------------|---------|
| `k8s-builder-go` | `configs/jenkins-templates/go-pipeline.groovy` | 所有 Go 项目 |
| `k8s-builder-java` | `configs/jenkins-templates/java-spring-pipeline.groovy` | 所有 Java/Spring 项目 |
| `k8s-builder-frontend` | `configs/jenkins-templates/frontend-pipeline.groovy` | 所有前端项目 |
| `k8s-builder-python` | `configs/jenkins-templates/python-pipeline.groovy` | 所有 Python 项目 |

**创建步骤：**
1. Jenkins → New Item → Pipeline → 命名为 `k8s-builder-go`
2. Pipeline → Definition: **Pipeline script from SCM**
3. SCM: Git → Repository URL: 平台仓库地址
4. Credentials: 选择 `gitee-id`
5. Script Path: `configs/jenkins-templates/go-pipeline.groovy`
6. 保存

> **一个 Job 服务 100+ 项目**：项目差异通过参数传入，模板不变。

---

## 4. Pipeline 模板详解（以 Go 为例）

### 4.1 Agent 定义（K8s Pod Agent + Kaniko）

```groovy
agent {
    kubernetes {
        yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: golang
    image: golang:1.24
    command: ['sleep', '99d']
    resources:
      requests: { cpu: '500m', memory: '512Mi' }
      limits:   { cpu: '2', memory: '2Gi' }
    env:
    - { name: GOPROXY, value: 'https://goproxy.cn,direct' }
    - { name: CGO_ENABLED, value: '0' }
    volumeMounts:
    - { name: go-cache, mountPath: /go/pkg/mod }
  - name: kaniko
    image: gcr.io/kaniko-project/executor:debug
    command: ['sleep', '99d']
    resources:
      requests: { cpu: '200m', memory: '256Mi' }
      limits:   { cpu: '1', memory: '1Gi' }
  volumes:
  - name: go-cache
    persistentVolumeClaim:
      claimName: jenkins-go-cache
  - name: workspace-volume
    emptyDir: {}
"""
    }
}
```

### 4.2 标准阶段流程

```
Clean Workspace → Checkout Info → Dependencies → Compile Check
       → Test → Lint → [SonarQube] → [Quality Gate]
       → [Build Binary] → [Upload Artifact]
       → Build Image (Kaniko) → Push Image → 回调平台
```

### 4.3 关键环节说明

| 阶段 | 容器 | 操作 |
|------|------|------|
| Clean + Checkout | jnlp (默认) | Git clone 业务代码 |
| Dependencies | golang | `go mod download` |
| Compile | golang | `go build -o app` |
| Test | golang | `go test ./...` |
| Build Image | kaniko | 生成/使用 Dockerfile → Kaniko 构建 |
| Push Image | kaniko | 推送到 Harbor/ACR |
| 回调 | jnlp | httpRequest 调用平台 API |

---

## 5. 平台侧处理逻辑

### 5.1 回调处理流程

```
收到回调 → HMAC 签名验证 → 查找流水线 → 查找运行记录
    → 幂等检查（防重复）→ 更新状态/镜像 → 更新阶段信息
    → 发送通知 → 自动部署（如配置）→ 同步发布记录
```

### 5.2 自动部署到 K8s

构建成功后，平台根据流水线配置自动更新 K8s 工作负载镜像：

```go
// 条件：构建成功 + 有镜像 + 配置了目标工作负载
if runStatus == "success" && image != "" {
    // 支持 Deployment / StatefulSet / DaemonSet
    patch := `{"spec":{"template":{"spec":{"containers":[{"name":"<容器>","image":"<新镜像>"}]}}}}`
    kubeClient.AppsV1().Deployments(namespace).Patch(...)
    // 等待 Rollout 完成
    waitDeploymentRollout(...)
}
```

**流水线配置字段：**

| 字段 | 说明 | 示例 |
|------|------|------|
| `auto_deploy` | 是否自动部署 | `true` |
| `target_cluster_id` | 目标集群 ID | `2` |
| `target_namespace` | 目标命名空间 | `production` |
| `target_workload_kind` | 工作负载类型 | `Deployment` |
| `target_workload_name` | 工作负载名称 | `my-app` |
| `target_container` | 容器名称 | `my-app` |
| `require_approval` | 是否需要人工审批 | `true` |

### 5.3 审批机制

如果配置了 `require_approval = true`：

```
构建成功 → 发送审批通知(钉钉/飞书) → 等待审批
    → 审批通过 → 自动部署
    → 审批拒绝 → 流水线标记失败
```

---

## 6. 平台配置说明（config.yaml）

```yaml
Jenkins:
  # === 连接配置 ===
  URL: "http://jenkins.devops.svc.cluster.local:8080/"  # Jenkins 服务地址
  Username: "ops-dev"                                    # Jenkins 用户名
  APIToken: "xxxxx"                                      # Jenkins API Token

  # === 回调与安全 ===
  CallbackURL: "http://k8soperation-backend.k8soperation.svc.cluster.local:38180"
  PlatformURL: "http://k8soperation-frontend.k8soperation.svc.cluster.local"
  HMACSecret: "your-32-bytes-secret"   # 必须与 Jenkins hmac-secret 凭证一致

  # === 凭证 ID（对应 Jenkins Credentials 中的 ID）===
  GitCredentialID: "gitee-id"
  RegistryCredentialID: "harbor-registry"
  HMACCredentialID: "hmac-secret"

  # === 运行控制 ===
  TriggerTimeout: 60       # 触发超时（秒），等待 Build Number
  PollInterval: 15         # 状态轮询间隔（秒）
  MaxBuildTime: 30         # 最大构建时间（分钟）

  # === 通知 ===
  EnableDingTalk: true
  DingTalkWebhook: "https://oapi.dingtalk.com/robot/send?access_token=xxx"
  EnableFeishu: false
  FeishuWebhook: ""
```

---

## 7. 容器存储与日志挂载（排查错误）

### 7.1 后端容器挂载一览

| 挂载路径 | 类型 | 说明 | 是否持久化 |
|----------|------|------|-----------|
| `/app/configs/config.yaml` | ConfigMap (subPath) | 平台配置文件 | 非持久化（ConfigMap） |
| `/app/configs/jenkins-templates` | ConfigMap (optional) | Pipeline 模板覆盖 | 非持久化 |
| `/app/storage/artifacts` | **PVC (20Gi)** | CI/CD 构建制品存储 | **是（hostPath → Retain）** |
| `/app/storage/logs` | **PVC (5Gi)** | 应用日志持久化 | **是（hostPath → Retain）** |

### 7.2 日志文件位置

当前后端容器内日志存储在 PVC 挂载目录：

```
/app/storage/logs/
├── app.log        ← 系统日志（全部操作、错误、堆栈）
├── biz.log        ← 业务日志（CI/CD 操作记录）
└── ai.log         ← AI 助手日志（大模型请求/响应）
```

**日志特性：**
- 系统日志同时输出到 **stdout + 文件**（`kubectl logs` 可看实时日志）
- 文件日志自动轮转（默认 1MB/文件，保留 3 份，30 天过期，gzip 压缩）
- Error 级别日志自动附带 **堆栈追踪（stacktrace）**
- 日志格式：JSON 结构化（方便 ELK/Loki 采集）

### 7.3 日志排查方式

#### 方式一：kubectl logs（实时/stdout）

```bash
# 实时查看后端日志
kubectl logs -f deploy/k8soperation -n k8soperation

# 查看之前容器的日志（崩溃后恢复场景）
kubectl logs deploy/k8soperation -n k8soperation --previous

# 查看最近 100 行
kubectl logs deploy/k8soperation -n k8soperation --tail=100
```

#### 方式二：PVC 持久化文件（历史日志）

```bash
# 进入容器查看日志文件
kubectl exec -it deploy/k8soperation -n k8soperation -- sh

# 容器内查看
ls -la /app/storage/logs/
cat /app/storage/logs/app.log          # 系统日志
cat /app/storage/logs/biz.log          # 业务日志

# 搜索错误
grep -i "error\|panic" /app/storage/logs/app.log | tail -50
```

#### 方式三：宿主机直接查看（hostPath 模式）

```bash
# 在 K8s Node 上直接查看（开发环境 hostPath）
ls /data/k8soperation/logs/
cat /data/k8soperation/logs/app.log
```

### 7.4 Jenkins 容器存储

| 挂载路径 | 类型 | 说明 |
|----------|------|------|
| `/var/jenkins_home` | PVC (jenkins-data) | Jenkins 全部数据（Job 配置、构建历史、插件） |
| `/var/jenkins_home/casc_configs` | ConfigMap | JCasC 配置文件 |

```bash
# 查看 Jenkins 日志
kubectl logs statefulset/jenkins -n devops

# 进入 Jenkins 容器
kubectl exec -it jenkins-0 -n devops -- bash
cat /var/jenkins_home/jobs/k8s-builder-go/builds/42/log  # 特定构建日志
```

### 7.5 当前存储是否适合排查错误？

**当前状态分析：**

| 能力 | 状态 | 说明 |
|------|------|------|
| 容器 stdout 日志 | ✅ 已支持 | `kubectl logs` 可实时查看 |
| 日志文件持久化 | ✅ 已支持 | PVC 挂载 `/app/storage/logs`，容器重启不丢失 |
| 崩溃前日志 | ✅ 已支持 | `--previous` 查看上一容器 + PVC 文件保留 |
| 堆栈追踪 | ✅ 已支持 | Error 级别自动附带 stacktrace |
| 日志轮转 | ✅ 已支持 | lumberjack 自动轮转 + 压缩 |
| 结构化日志 | ✅ 已支持 | JSON 格式（zap logger） |
| 制品存储 | ✅ 已支持 | PVC 20Gi，构建产物持久化 |

**结论：当前架构已经挂载了 PVC 存储日志，方便排查后端错误。**

---

## 8. 完整工作流（端到端）

```
┌─────────────────────────────────────────────────────────────────────────┐
│  1. 用户在前端创建流水线                                                  │
│     - 填写 Git 仓库、选择语言类型(go/java/frontend/python)                │
│     - 配置目标集群、命名空间、Deployment、容器名                           │
│     - 开启自动部署、审批（可选）                                          │
└──────────────────────────────────────────┬──────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  2. 用户点击"运行流水线"                                                  │
│     - 平台自动推导 Jenkins Job 名（如 k8s-builder-go）                    │
│     - 自动组装 20+ 参数（Git/镜像/回调/凭证/SonarQube...）                │
│     - 调用 Jenkins REST API 触发构建                                     │
└──────────────────────────────────────────┬──────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  3. Jenkins 创建动态 Pod Agent                                           │
│     - Kubernetes Plugin 在 devops namespace 创建构建 Pod                 │
│     - Pod 内包含业务容器(golang/node) + Kaniko 容器                      │
│     - 使用 PVC 缓存依赖（go mod cache / node_modules）                   │
└──────────────────────────────────────────┬──────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  4. Pipeline 执行（每阶段回调平台）                                       │
│     Clean → Checkout → Dependencies → Compile → Test → Lint             │
│       → [SonarQube] → [Quality Gate] → Build Image → Push Image         │
│     每个阶段完成: POST /stage/callback (X-Signature HMAC)                │
│     平台实时更新 UI 进度条                                                │
└──────────────────────────────────────────┬──────────────────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  5. 构建完成回调平台                                                      │
│     POST /pipeline/callback (X-Signature HMAC)                          │
│     携带: status, image, digest, duration, git_commit                   │
│     平台验证签名 → 更新状态 → 触发后续流程                                 │
└──────────────────────────────────────────┬──────────────────────────────┘
                                           │
                              ┌─────────────┴──────────────┐
                              │                            │
                              ▼                            ▼
              ┌──────────────────────────┐  ┌──────────────────────────┐
              │ 6a. 需要审批               │  │ 6b. 无需审批             │
              │  - 发送钉钉/飞书通知       │  │  - 直接自动部署          │
              │  - 等待审批者通过          │  │                          │
              │  - 通过后 → 自动部署       │  │                          │
              └──────────────┬─────────────┘  └──────────────┬───────────┘
                              │                              │
                              └──────────────┬───────────────┘
                                             │
                                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  7. 自动部署到 K8s                                                       │
│     - 根据 target_workload_kind 选择 Patch 方式                          │
│     - StrategicMergePatch 更新镜像                                       │
│     - 等待 Rollout 完成（健康检查通过）                                    │
│     - 发送部署结果通知                                                    │
│     - 同步创建发布记录                                                    │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 9. 网络拓扑（集群内通信）

| 源 | 目标 | Service 地址 | 端口 | 用途 |
|----|------|-------------|------|------|
| 平台后端 | Jenkins | `jenkins.devops.svc.cluster.local` | 8080 | 触发构建、获取日志 |
| Jenkins Pod Agent | Jenkins Master | `jenkins-agent.devops.svc.cluster.local` | 50000 | JNLP Agent 通信 |
| Jenkins | 平台后端 | `k8soperation-backend.k8soperation.svc.cluster.local` | 38180 | 构建回调 |
| 平台前端 | 平台后端 | `k8soperation-backend.k8soperation.svc.cluster.local` | 38180 | API 反向代理 |
| 外部用户 | Jenkins | NodePort | 30080 | Jenkins Web UI |
| 外部用户 | 平台前端 | NodePort/Ingress | - | 平台 Web UI |

---

## 10. Jenkins 从零配置清单

### 10.1 自动完成的配置（通过 JCasC）

部署 Jenkins StatefulSet 后自动配置：
- ✅ 管理员账号 (ops-dev)
- ✅ K8s Cloud（动态 Pod Agent）
- ✅ HMAC Secret 凭证
- ✅ 安全策略（禁止匿名访问）
- ✅ 必装插件（kubernetes, workflow, git, sonar 等）

### 10.2 手动配置项

| 序号 | 配置项 | 操作步骤 |
|------|--------|---------|
| 1 | 创建 Git 凭证 | Jenkins → Manage Credentials → Add → Username/Password → ID: `gitee-id` |
| 2 | 创建镜像仓库凭证 | Jenkins → Manage Credentials → Add → Username/Password → ID: `harbor-registry` |
| 3 | 创建 Go 构建 Job | New Item → Pipeline → Name: `k8s-builder-go` → SCM → Script Path: `configs/jenkins-templates/go-pipeline.groovy` |
| 4 | 创建 Java 构建 Job | 同上 → Name: `k8s-builder-java` → Script Path: `configs/jenkins-templates/java-spring-pipeline.groovy` |
| 5 | 创建前端构建 Job | 同上 → Name: `k8s-builder-frontend` → Script Path: `configs/jenkins-templates/frontend-pipeline.groovy` |
| 6 | 创建 Python 构建 Job | 同上 → Name: `k8s-builder-python` → Script Path: `configs/jenkins-templates/python-pipeline.groovy` |
| 7 | 创建 Go 缓存 PVC | `kubectl apply -f` 创建 `jenkins-go-cache` PVC（可选） |

### 10.3 生成 API Token

1. 登录 Jenkins → 点击右上角用户名 → Configure
2. API Token 区域 → Add new Token → 命名 `platform-token`
3. 生成 → 复制 Token
4. 填入平台 `Secret.yaml` 的 `JENKINS_API_TOKEN` 字段

---

## 11. 故障排查指南

### 11.1 触发构建失败

```bash
# 检查平台到 Jenkins 的网络连通性
kubectl exec -it deploy/k8soperation -n k8soperation -- \
  wget -q -O- http://jenkins.devops.svc.cluster.local:8080/api/json

# 查看平台日志中的 Jenkins 相关错误
kubectl logs deploy/k8soperation -n k8soperation | grep -i "jenkins\|trigger"
```

### 11.2 回调未收到

```bash
# 检查 Jenkins 到平台的网络连通性（在 Jenkins Pod 中执行）
kubectl exec -it jenkins-0 -n devops -- \
  curl -s http://k8soperation-backend.k8soperation.svc.cluster.local:38180/healthz/live

# 检查 HMAC 密钥是否一致
kubectl get secret k8soperation-secret -n k8soperation -o jsonpath='{.data.HMAC_SECRET}' | base64 -d
kubectl get secret jenkins-secret -n devops -o jsonpath='{.data.hmac-secret}' | base64 -d
```

### 11.3 自动部署失败

```bash
# 查看平台日志中的自动部署记录
kubectl logs deploy/k8soperation -n k8soperation | grep -i "自动部署\|auto.*deploy\|rollout"

# 查看目标 Deployment 的事件
kubectl describe deployment <name> -n <namespace>
kubectl rollout status deployment/<name> -n <namespace>
```

### 11.4 查看持久化日志

```bash
# 进入后端容器查看历史日志
kubectl exec -it deploy/k8soperation -n k8soperation -- sh -c "ls -lah /app/storage/logs/"

# 搜索 panic/错误
kubectl exec -it deploy/k8soperation -n k8soperation -- \
  grep -n "panic\|FATAL\|level.*error" /app/storage/logs/app.log | tail -30

# 查看业务日志（CI/CD 操作记录）
kubectl exec -it deploy/k8soperation -n k8soperation -- tail -50 /app/storage/logs/biz.log
```

---

## 12. 存储配置总结

### 12.1 后端 PV/PVC 配置

```yaml
# PV（开发环境 hostPath）
- k8soperation-artifacts: 20Gi, hostPath: /data/k8soperation/artifacts
- k8soperation-logs:      5Gi,  hostPath: /data/k8soperation/logs

# PVC
- k8soperation-artifacts: 20Gi, ReadWriteOnce  → 构建制品
- k8soperation-logs:      5Gi,  ReadWriteOnce  → 应用日志
```

### 12.2 生产环境建议

| 场景 | 当前方式 | 生产推荐 |
|------|---------|---------|
| 日志存储 | PVC (hostPath) | **EFK/Loki + stdout 采集**（无需 PVC） |
| 制品存储 | PVC (hostPath) | **NFS/CephFS (RWX)**（支持多副本） |
| Jenkins 数据 | PVC (StatefulSet) | 保持不变（StatefulSet + PVC 已是最佳实践） |
| 日志查看 | 手动 exec/logs | **平台内集成 Loki 查询**（已支持） |

### 12.3 是否需要额外挂载存储？

**当前已足够排查错误：**
1. `kubectl logs` → 实时 stdout 日志
2. PVC `/app/storage/logs/` → 持久化文件日志（容器重启不丢失）
3. `--previous` → 崩溃前容器日志
4. 结构化 JSON + 堆栈追踪 → 错误定位方便

**如果需要增强：**
- 可接入 Loki/ELK 进行集中日志查询
- 可开启 `MirrorBusinessToSystem: true` 将业务日志也输出到 stdout

---

## 13. API 接口总览

| 接口 | 方法 | 说明 | 调用方 |
|------|------|------|--------|
| `/api/v1/k8s/cicd/pipeline/create` | POST | 创建流水线 | 前端 |
| `/api/v1/k8s/cicd/pipeline/run` | POST | 运行流水线 | 前端 |
| `/api/v1/k8s/cicd/pipeline/callback` | POST | Jenkins 构建完成回调 | Jenkins |
| `/api/v1/k8s/cicd/stage/callback` | POST | Jenkins 阶段回调 | Jenkins |
| `/api/v1/k8s/cicd/pipeline/stages` | GET | 获取流水线阶段数据 | 前端 |
| `/api/v1/k8s/cicd/pipeline/logs` | GET | 获取构建日志 | 前端 |
| `/api/v1/k8s/cicd/pipeline/stop` | POST | 停止构建 | 前端 |
| `/api/v1/k8s/cicd/pipeline/approval` | POST | 审批操作 | 前端 |
| `/api/v1/k8s/cicd/template/verify` | GET | 验证模板配置 | 前端 |

---

## 14. 安全要点

| 安全措施 | 实现方式 |
|----------|---------|
| 回调防伪造 | HMAC-SHA256 签名验证 |
| HMAC 密钥安全 | Jenkins Credentials + K8s Secret，不通过参数传递 |
| 幂等处理 | `callback_received` 标记，防止重复回调 |
| 敏感配置 | K8s Secret + 环境变量注入，不进 Git |
| Jenkins 认证 | Basic Auth (Username + API Token) |
| CSRF 防护 | Jenkins Crumb Token 自动处理 |
| 签名时间安全 | `subtle.ConstantTimeCompare` 防时序攻击 |
