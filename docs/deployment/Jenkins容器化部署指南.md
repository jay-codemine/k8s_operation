# Jenkins 容器化部署指南（K8s 集群内）

> Jenkins 作为 K8sOperation 平台的 CI/CD 构建引擎，采用 StatefulSet 方式部署在 K8s 集群内，
> 构建任务通过 Kubernetes Plugin 动态创建 Pod Agent 执行。

---

## 目录

- [一、架构设计](#一架构设计)
- [二、环境要求](#二环境要求)
- [三、部署文件说明](#三部署文件说明)
- [四、修改 Secret 配置](#四修改-secret-配置)
- [五、修改 ConfigMap（JCasC）](#五修改-configmapjcasc)
- [六、一键部署](#六一键部署)
- [七、验证部署](#七验证部署)
- [八、访问 Jenkins Web UI](#八访问-jenkins-web-ui)
- [九、与平台对接配置](#九与平台对接配置)
- [十、构建缓存说明](#十构建缓存说明)
- [十一、运维操作](#十一运维操作)
- [十二、常见问题排查](#十二常见问题排查)

---

## 一、架构设计

```
┌─────────────────────────────────────────────────────────────────────────┐
│  K8s Cluster                                                             │
│                                                                           │
│  Namespace: devops                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐     │
│  │  StatefulSet: jenkins                                            │     │
│  │  ┌──────────────────────────────────────────────────────────┐   │     │
│  │  │ Pod: jenkins-0                                            │   │     │
│  │  │  initContainer: install-plugins (自动安装插件)             │   │     │
│  │  │  container: jenkins (LTS, port 8080/50000)                │   │     │
│  │  │    ├── PVC jenkins-data → /var/jenkins_home               │   │     │
│  │  │    └── ConfigMap casc → /var/jenkins_home/casc_configs    │   │     │
│  │  └──────────────────────────────────────────────────────────┘   │     │
│  └─────────────────────────────────────────────────────────────────┘     │
│                                                                           │
│  Services:                                                                │
│    jenkins (ClusterIP:8080) ← 平台 API 调用                              │
│    jenkins-agent (ClusterIP:50000) ← Agent 连接                          │
│    jenkins-nodeport (NodePort:30080) ← 外部 Web UI 访问                  │
│                                                                           │
│  动态 Agent Pods（构建时创建，完成后销毁）:                                │
│    ┌─────────────────────┐  ┌──────────────────────┐                     │
│    │ Pod: java-build-xxx  │  │ Pod: go-build-xxx    │                     │
│    │ ├── maven container  │  │ ├── golang container │                     │
│    │ └── kaniko container │  │ └── kaniko container │                     │
│    └─────────────────────┘  └──────────────────────┘                     │
│                                                                           │
│  PVCs（构建缓存，加速后续构建）:                                           │
│    jenkins-maven-cache (20Gi)                                             │
│    jenkins-go-cache (10Gi)                                                │
│    jenkins-npm-cache (10Gi)                                               │
│    jenkins-pip-cache (5Gi)                                                │
└─────────────────────────────────────────────────────────────────────────┘
```

### 核心设计理念

| 特性 | 说明 |
|------|------|
| **StatefulSet** | 稳定的网络标识和持久化存储 |
| **Master 不构建** | numExecutors=0，全部交给动态 Pod Agent |
| **JCasC** | Configuration as Code，配置全部声明式管理 |
| **动态 Agent** | 每次构建创建独立 Pod，完成后自动销毁 |
| **构建缓存** | Maven/Go/npm/pip 缓存持久化，加速重复构建 |
| **initContainer 安装插件** | 启动前自动安装必要插件 |

---

## 二、环境要求

| 组件 | 要求 | 说明 |
|------|------|------|
| K8s 集群 | 1.28+ | Jenkins 和 Agent Pod 运行环境 |
| 节点内存 | ≥ 4Gi 可用 | Jenkins Master 需要 2-8Gi |
| 存储 | hostPath 或 StorageClass | PVC 持久化 |
| 网络 | Pod 间可通信 | Jenkins ↔ Agent、Jenkins ↔ 平台 |
| 镜像拉取 | 可访问 Docker Hub | `jenkins/jenkins:lts` |

---

## 三、部署文件说明

所有文件位于 `deploy/jenkins/` 目录：

```
deploy/jenkins/
├── kustomization.yaml     # Kustomize 编排入口
├── namespace.yaml         # devops 命名空间
├── rbac.yaml              # ServiceAccount + ClusterRole（Agent 管理权限）
├── secret.yaml            # 敏感信息（密码、Token、凭证）
├── pv.yaml                # 静态 PV（构建缓存）
├── pvc.yaml               # PVC 定义（主数据 + 4 种语言缓存）
├── configmap.yaml         # JCasC 配置（Kubernetes Cloud、凭证、安全策略）
├── statefulset.yaml       # Jenkins Master StatefulSet
└── service.yaml           # Service 定义（ClusterIP + Agent + NodePort）
```

---

## 四、修改 Secret 配置

文件路径：`deploy/jenkins/secret.yaml`

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: jenkins-secret
  namespace: devops
type: Opaque
data:
  # ===== Jenkins 管理员密码 =====
  # 生成命令：echo -n "your-password" | base64
  admin-password: YWRtaW4xMjM=                           # ← 修改！默认 admin123

  # ===== HMAC 签名密钥（回调验签，需与平台 Secret 中的 HMAC_SECRET 一致） =====
  # echo -n "a9XkP7mQ2vNc8Lr4TzY6HwFd3BjUs5Ge" | base64
  hmac-secret: YTlYa1A3bVEydk5jOExyNFR6WTZId0ZkM0JqVXM1R2U=

  # ===== SonarQube Token（可选，代码质量扫描） =====
  # echo -n "your-sonarqube-token" | base64
  sonarqube-token: eW91ci1zb25hcnF1YmUtdG9rZW4=

  # ===== 镜像仓库凭证（阿里云 ACR / Harbor） =====
  # echo -n "your-registry-username" | base64
  registry-username: MTU4NjIzMjY0OTA=
  # echo -n "your-registry-password" | base64
  registry-password: ZGM1MjE1MjEuLjA=                    # ← 修改为实际密码

  # ===== Git 凭证（Gitee / GitHub / GitLab） =====
  # echo -n "your-git-username" | base64
  gitee-username: amF5LWtpbQ==
  # echo -n "your-personal-access-token" | base64
  gitee-password: NDdhZjc0NjgwM2Q2NmFjZDFjMjExYTEzZjkzM2U3MjA=  # ← Personal Access Token
```

### 必须修改的配置

| 字段 | 说明 | 生成方式 |
|------|------|----------|
| `admin-password` | Jenkins 管理员密码 | `echo -n "StrongPass123!" \| base64` |
| `hmac-secret` | HMAC 签名密钥 | 与平台 `HMAC_SECRET` 保持一致 |
| `registry-username` | 镜像仓库用户名 | 阿里云 ACR 用户名 |
| `registry-password` | 镜像仓库密码 | 阿里云 ACR 密码 |
| `gitee-username` | Git 用户名 | Gitee/GitHub 用户名 |
| `gitee-password` | Git Token | Personal Access Token（非登录密码） |

---

## 五、修改 ConfigMap（JCasC）

文件路径：`deploy/jenkins/configmap.yaml`

这是 Jenkins Configuration as Code (JCasC) 配置，Jenkins 启动时自动加载。

### 5.1 需要修改的关键配置

```yaml
data:
  casc.yaml: |
    jenkins:
      # ===== 全局环境变量 =====
      globalNodeProperties:
        - envVars:
            env:
              - key: "PLATFORM_CALLBACK_URL"
                # ★ 修改为平台后端的回调地址
                # K8s 集群内部署：
                value: "http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/pipeline/callback"
                # Docker Desktop 本地开发：
                # value: "http://host.docker.internal:38180/api/v1/k8s/cicd/pipeline/callback"
              - key: "ARTIFACT_UPLOAD_URL"
                value: "http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/artifact/upload"

      # ===== 管理员账号（密码从 Secret 注入） =====
      securityRealm:
        local:
          allowsSignup: false
          users:
            - id: "ops-dev"               # ← 用户名，与平台 JENKINS_USERNAME 对应
              password: "${JENKINS_ADMIN_PASSWORD}"

      # ===== Kubernetes Cloud 配置 =====
      clouds:
        - kubernetes:
            name: "kubernetes"
            serverUrl: "https://kubernetes.default.svc.cluster.local"
            namespace: "devops"           # Agent Pod 创建的命名空间
            jenkinsUrl: "http://jenkins.devops.svc.cluster.local:8080"
            jenkinsTunnel: "jenkins-agent.devops.svc.cluster.local:50000"
            containerCapStr: "20"         # ← 最大并发 Agent Pod 数
            waitForPodSec: 600            # Agent Pod 最大等待时间

    # ===== 凭证配置（通过 K8s Secret 环境变量注入） =====
    credentials:
      system:
        domainCredentials:
          - credentials:
              - string:
                  id: "hmac-secret"
                  secret: "${HMAC_SECRET}"
              - string:
                  id: "sonarqube-token"
                  secret: "${SONAR_TOKEN}"
              - usernamePassword:
                  id: "harbor-registry"       # ← 平台流水线引用的镜像仓库凭证 ID
                  username: "${REGISTRY_USERNAME}"
                  password: "${REGISTRY_PASSWORD}"
              - usernamePassword:
                  id: "gitee-id"              # ← 平台流水线引用的 Git 凭证 ID
                  username: "${GITEE_USERNAME}"
                  password: "${GITEE_PASSWORD}"

    # ===== SonarQube 配置（可选） =====
    unclassified:
      sonarGlobalConfiguration:
        installations:
          - name: "SonarQube"
            serverUrl: "http://sonarqube.example.com:9000"  # ← 修改为实际 SonarQube 地址
            credentialsId: "sonarqube-token"
```

### 5.2 PLATFORM_CALLBACK_URL 配置规则

| 部署场景 | CallbackURL 值 |
|----------|----------------|
| Jenkins 与平台同集群 | `http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/pipeline/callback` |
| Docker Desktop 本地开发 | `http://host.docker.internal:38180/api/v1/k8s/cicd/pipeline/callback` |
| 平台部署在集群外 | `http://<平台IP>:<端口>/api/v1/k8s/cicd/pipeline/callback` |

---

## 六、一键部署

```bash
# 1. 修改 Secret（密码、Token、凭证）
vim deploy/jenkins/secret.yaml

# 2. 修改 ConfigMap（回调地址、SonarQube 地址等）
vim deploy/jenkins/configmap.yaml

# 3. 一键部署
kubectl apply -k deploy/jenkins/

# 4. 查看部署状态（Jenkins 启动较慢，约 2-3 分钟）
kubectl -n devops get pods -w
```

### 部署顺序

```
namespace → rbac → secret → pv → pvc → configmap → statefulset → service
```

---

## 七、验证部署

```bash
# 1. 检查 Pod 状态
kubectl -n devops get pods
# 期望输出：
# NAME        READY   STATUS    RESTARTS   AGE
# jenkins-0   1/1     Running   0          3m

# 2. 检查 initContainer 日志（插件安装）
kubectl -n devops logs jenkins-0 -c install-plugins

# 3. 检查 Jenkins 主容器日志
kubectl -n devops logs jenkins-0 -c jenkins | tail -20
# 看到 "Jenkins is fully up and running" 即成功

# 4. 检查 Service
kubectl -n devops get svc
# NAME              TYPE        CLUSTER-IP      PORT(S)
# jenkins           ClusterIP   10.x.x.x       8080/TCP,50000/TCP
# jenkins-agent     ClusterIP   10.x.x.x       50000/TCP
# jenkins-nodeport  NodePort    10.x.x.x       8080:30080/TCP

# 5. 检查 PVC 绑定状态
kubectl -n devops get pvc
# 所有 PVC 应为 Bound 状态
```

---

## 八、访问 Jenkins Web UI

### 方式 A：NodePort（默认开启）

```bash
# 访问地址
http://<任意节点IP>:30080

# Docker Desktop
http://localhost:30080

# 登录凭证
# 用户名：ops-dev
# 密码：secret.yaml 中 admin-password 的值（默认 admin123）
```

### 方式 B：kubectl port-forward（临时访问）

```bash
kubectl -n devops port-forward svc/jenkins 8080:8080
# 访问 http://localhost:8080
```

### 方式 C：Ingress（生产环境）

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: jenkins
  namespace: devops
spec:
  ingressClassName: nginx
  rules:
    - host: jenkins.your-domain.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: jenkins
                port:
                  name: http
```

---

## 九、与平台对接配置

Jenkins 部署完成后，需要在平台侧（`deploy/backend/secret.yaml`）配置以下对应信息：

| 平台配置项 | 值 | 说明 |
|-----------|-----|------|
| `JENKINS_URL` | `http://jenkins.devops.svc.cluster.local:8080/` | Jenkins 集群内 Service 地址 |
| `JENKINS_USERNAME` | `ops-dev` | 与 JCasC 中配置的用户名一致 |
| `JENKINS_API_TOKEN` | 从 Jenkins UI 获取 | 用户设置 → Security → API Token |
| `HMAC_SECRET` | 与 Jenkins Secret 中的 `hmac-secret` 一致 | 回调签名验证 |
| `GIT_CREDENTIAL_ID` | `gitee-id` | Jenkins 中的 Git 凭证 ID |
| `REGISTRY_CREDENTIAL_ID` | `harbor-registry` | Jenkins 中的镜像仓库凭证 ID |
| `HMAC_CREDENTIAL_ID` | `hmac-secret` | Jenkins 中的 HMAC 凭证 ID |

### 获取 Jenkins API Token

```bash
# 1. 登录 Jenkins Web UI
# 2. 右上角 → 用户名 → Configure → API Token → Add new Token
# 3. 复制生成的 Token 填入平台 Secret 的 JENKINS_API_TOKEN
```

---

## 十、构建缓存说明

### 10.1 缓存 PVC 列表

| PVC | 容量 | 挂载路径（Agent Pod 内） | 用途 |
|-----|------|--------------------------|------|
| `jenkins-maven-cache` | 20Gi | `/root/.m2/repository` | Maven 依赖缓存 |
| `jenkins-go-cache` | 10Gi | `/go/pkg/mod` | Go Module 缓存 |
| `jenkins-npm-cache` | 10Gi | `/root/.npm` | npm 依赖缓存 |
| `jenkins-pip-cache` | 5Gi | `/root/.cache/pip` | Python pip 缓存 |

### 10.2 缓存在 Groovy 模板中的引用

```groovy
// java-spring-pipeline.groovy 中的 Pod 模板
spec:
  containers:
  - name: maven
    volumeMounts:
    - name: maven-cache
      mountPath: /root/.m2/repository
  volumes:
  - name: maven-cache
    persistentVolumeClaim:
      claimName: jenkins-maven-cache
```

### 10.3 节点数据目录

PV 使用 hostPath，数据存储在节点上：

```
/data/jenkins/
├── maven-cache/    # Maven 仓库缓存
├── go-cache/       # Go 模块缓存
├── npm-cache/      # npm 缓存
└── pip-cache/      # pip 缓存
```

---

## 十一、运维操作

### 11.1 重启 Jenkins

```bash
kubectl -n devops rollout restart statefulset jenkins
# 或
kubectl -n devops delete pod jenkins-0
```

### 11.2 查看/修改 JCasC 配置

```bash
# 修改配置
kubectl -n devops edit configmap jenkins-casc-config

# 重启使配置生效
kubectl -n devops delete pod jenkins-0
```

### 11.3 手动安装插件

```bash
# 进入 Jenkins Pod
kubectl -n devops exec -it jenkins-0 -- bash

# 使用 CLI 安装插件
jenkins-plugin-cli --plugins <plugin-name>

# 或通过 Web UI：Manage Jenkins → Plugins → Available
```

### 11.4 备份 Jenkins 数据

```bash
# 备份整个 jenkins-data PVC
kubectl -n devops exec jenkins-0 -- tar czf /tmp/jenkins-backup.tar.gz /var/jenkins_home
kubectl -n devops cp jenkins-0:/tmp/jenkins-backup.tar.gz ./jenkins-backup.tar.gz
```

### 11.5 扩展 Agent 并发数

修改 ConfigMap 中的 `containerCapStr`：

```yaml
clouds:
  - kubernetes:
      containerCapStr: "50"    # 从 20 扩展到 50
```

### 11.6 资源调整

修改 `statefulset.yaml` 中的 resources：

```yaml
resources:
  requests:
    cpu: "1"       # 最小 CPU
    memory: 2Gi    # 最小内存
  limits:
    cpu: "4"       # 最大 CPU
    memory: 8Gi    # 最大内存（大项目构建时 Jenkins 内存消耗较高）
```

---

## 十二、常见问题排查

### Q1: initContainer 插件安装超时/失败

```bash
# 查看日志
kubectl -n devops logs jenkins-0 -c install-plugins

# 原因：网络问题导致无法下载插件
# 解决：
# 1. 确认节点能访问 updates.jenkins.io
# 2. 或提前下载 .hpi 文件放入 PVC 的 plugins/ 目录
# 3. Jenkins 启动后可从 UI 手动安装
```

### Q2: Agent Pod 创建失败

```bash
# 查看 Jenkins 日志
kubectl -n devops logs jenkins-0 -c jenkins | grep -i "agent\|pod\|error"

# 常见原因：
# 1. RBAC 权限不足 → 检查 ClusterRoleBinding
kubectl get clusterrolebinding jenkins-agent-manager-binding -o yaml

# 2. PVC 未绑定 → 检查缓存 PVC 状态
kubectl -n devops get pvc

# 3. 节点资源不足 → 检查节点可用资源
kubectl top nodes
```

### Q3: 构建回调失败（平台收不到）

```bash
# 1. 确认回调地址正确
kubectl -n devops exec jenkins-0 -- env | grep PLATFORM_CALLBACK_URL

# 2. 从 Jenkins Pod 测试连通性
kubectl -n devops exec jenkins-0 -- curl -s http://k8soperation.k8soperation.svc:8080/healthz/live

# 3. 如果平台在集群外（本地开发）
# 确保使用 host.docker.internal（Docker Desktop）
# 或节点 IP（裸机 K8s）
```

### Q4: Jenkins 重启后凭证丢失

```
✅ 本方案不会出现此问题！
凭证通过 JCasC + K8s Secret 环境变量注入，每次启动自动创建。
无需手动在 Jenkins UI 中添加凭证。
```

### Q5: 镜像仓库凭证 "harbor-registry" 不存在

```bash
# 检查 Secret 中的 registry-username / registry-password 是否正确
kubectl -n devops get secret jenkins-secret -o jsonpath='{.data.registry-username}' | base64 -d

# JCasC 中的凭证 ID 必须与 Groovy 模板中引用的一致：
# configmap.yaml:  id: "harbor-registry"
# pipeline.groovy: credentials("harbor-registry")
```

### Q6: Agent Pod 镜像拉取失败

```bash
# 国内环境 gcr.io 不可达
# 解决：修改 Groovy 模板中的 Kaniko 镜像为国内源
# configs/jenkins-templates/java-spring-pipeline.groovy:
#   原：image: gcr.io/kaniko-project/executor:debug
#   改：image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/kaniko:latest
```

---

## 附录：完整部署命令速查

```bash
# ========== Jenkins 一键部署 ==========

# Step 1: 修改敏感配置
vim deploy/jenkins/secret.yaml

# Step 2: 修改 JCasC（回调地址、SonarQube）
vim deploy/jenkins/configmap.yaml

# Step 3: 部署
kubectl apply -k deploy/jenkins/

# Step 4: 等待就绪（约 2-3 分钟）
kubectl -n devops wait --for=condition=ready pod/jenkins-0 --timeout=300s

# Step 5: 获取访问地址
echo "Jenkins URL: http://$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[0].address}'):30080"
echo "Username: ops-dev"
echo "Password: $(kubectl -n devops get secret jenkins-secret -o jsonpath='{.data.admin-password}' | base64 -d)"

# Step 6: 验证与平台连通
kubectl -n devops exec jenkins-0 -- curl -sf http://k8soperation.k8soperation.svc:8080/healthz/live && echo "OK"
```
