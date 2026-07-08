# GitOps 完整操作指南：ArgoCD + Argo Workflows 与平台集成

> **版本**: v2.5.0  
> **日期**: 2026-07-08  
> **适用对象**: 平台管理员、运维工程师、开发工程师  
> **前置知识**: Kubernetes 基础、Git 基础、Docker/镜像仓库基础

---

## 目录

1. [概述：什么是 GitOps](#1-概述什么是-gitops)
2. [ArgoCD 安装与配置](#2-argocd-安装与配置)
3. [Argo Workflows 安装与配置](#3-argo-workflows-安装与配置)
4. [平台 GitOps 配置](#4-平台-gitops-配置)
5. [新增应用：完整操作流程](#5-新增应用完整操作流程)
6. [Argo Workflow 模板详解](#6-argo-workflow-模板详解)
7. [ArgoCD Application 管理](#7-argocd-application-管理)
8. [镜像更新与自动同步](#8-镜像更新与自动同步)
9. [多环境管理](#9-多环境管理)
10. [故障排查](#10-故障排查)
11. [附录：命令速查表](#11-附录命令速查表)

---

## 1. 概述：什么是 GitOps

### 1.1 概念

GitOps 是一种以 **Git 仓库作为唯一真实来源（Source of Truth）** 的运维模式：

```
开发者 Push 代码 → CI 工具构建镜像 → 更新 Git 中的 K8s 清单 → CD 工具自动同步到集群
```

| 对比维度 | 传统 Push 模式（Jenkins） | GitOps Pull 模式（ArgoCD） |
|---------|------------------------|--------------------------|
| 部署方式 | CI 工具主动调用 K8s API 部署 | CD 工具从集群内拉取 Git 状态并同步 |
| 状态存储 | CI 工具内存 / 数据库 | Git 仓库 |
| 回滚方式 | 重新部署旧版本 | Git revert + 自动同步 |
| 集群外权限 | 需要 K8s 写权限 | 不需要（ArgoCD 在集群内） |
| 安全模型 | 凭据暴露面大 | 凭据仅在集群内 |

### 1.2 平台 GitOps 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        k8s_operation 平台                        │
│                                                                 │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────────┐  │
│  │ 创建流水线 │───▶│ 触发 Workflow │───▶│ 回调更新流水线状态     │  │
│  │(GitOps模式)│    │ (Argo API)   │    │ (/pipeline/callback) │  │
│  └──────────┘    └──────┬───────┘    └──────────────────────┘  │
│                         │                                       │
└─────────────────────────┼───────────────────────────────────────┘
                          │
         ┌────────────────▼────────────────┐
         │       Argo Workflows            │
         │  ┌──────┐ ┌──────┐ ┌─────────┐  │
         │  │checkout│ │build │ │update   │  │
         │  │       │→│image │→│manifest │  │
         │  └──────┘ └──────┘ └────┬────┘  │
         └─────────────────────────┼───────┘
                                   │ git push
         ┌─────────────────────────▼───────┐
         │         Git Manifest Repo       │
         │  manifests/overlays/prod/       │
         │  └── deployment.yaml (image tag)│
         └─────────────────────────┬───────┘
                                   │ detect drift
         ┌─────────────────────────▼───────┐
         │           ArgoCD                │
         │  ┌──────────────────────────┐   │
         │  │ Application: my-app-prod │   │
         │  │ Status: Synced ✅         │   │
         │  └──────────────────────────┘   │
         └─────────────────────────┬───────┘
                                   │ sync
         ┌─────────────────────────▼───────┐
         │        Kubernetes Cluster       │
         │  ┌──────────────────────────┐   │
         │  │ Deployment: my-app       │   │
         │  │ image: registry.../myapp │   │
         │  └──────────────────────────┘   │
         └─────────────────────────────────┘
```

### 1.3 组件说明

| 组件 | 作用 | 版本要求 |
|------|------|---------|
| **ArgoCD** | GitOps CD 引擎：监听 Git 仓库 → 自动同步到 K8s | ≥ 2.10 |
| **Argo Workflows** | K8s 原生 CI 引擎：编排构建/测试/镜像打包流程 | ≥ 3.5 |
| **Git Manifest Repo** | 存放 K8s 部署清单（Deployment/Service/Ingress 等）的 Git 仓库 | 任意 Git 服务 |
| **容器镜像仓库** | 存储构建产出的容器镜像 | Harbor / ACR / Docker Hub |

---

## 2. ArgoCD 安装与配置

### 2.1 通过平台应用商城一键安装（推荐）

登录平台 → 【应用商城】→ 找到 **ArgoCD** → 点击安装：

1. 选择目标集群
2. 安装到命名空间 `argocd`（默认）
3. 等待组件就绪（约 2-3 分钟）

安装后获取初始密码：

```bash
# 获取 admin 初始密码
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# 修改密码
argocd account update-password
```

### 2.2 手动安装（如果不用平台）

```bash
# 创建命名空间
kubectl create namespace argocd

# 安装 ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 暴露服务（任选一种）
# 方式 1: NodePort
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "NodePort"}}'

# 方式 2: Ingress（生产推荐）
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: argocd-server
  namespace: argocd
spec:
  rules:
  - host: argocd.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: argocd-server
            port:
              number: 443
EOF
```

### 2.3 安装 ArgoCD CLI

```bash
# macOS
brew install argocd

# Linux
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd

# 登录
argocd login <ARGOCD_SERVER> --username admin --password <PASSWORD>
```

### 2.4 配置 Git 仓库凭据

```bash
# 添加私有 Git 仓库（SSH 方式）
argocd repo add git@github.com:your-org/manifests.git \
  --ssh-private-key-path ~/.ssh/id_rsa

# 添加私有 Git 仓库（HTTPS + 用户名密码）
argocd repo add https://git.example.com/your-org/manifests.git \
  --username your-username \
  --password your-password

# 查看已配置的仓库
argocd repo list
```

### 2.5 配置 ArgoCD Webhook（平台回调）

ArgoCD 同步状态变更时需要回调平台更新流水线状态。

在 ArgoCD 中配置 webhook：

```bash
# argocd-notifications 方式（推荐）
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
  namespace: argocd
data:
  service.webhook.platform: |
    url: http://k8soperation-backend.k8soperation.svc.cluster.local:38180/api/v1/k8s/cicd/gitops/sync-callback
    headers:
    - name: Content-Type
      value: application/json
    - name: X-Signature
      value: $hmac-secret

  template.app-sync-status: |
    webhook:
      platform:
        method: POST
        body: |
          {
            "app_name": "{{.app.metadata.name}}",
            "sync_status": "{{.app.status.sync.status}}",
            "sync_revision": "{{.app.status.sync.revision}}",
            "health_status": "{{.app.status.health.status}}"
          }

  trigger.on-sync-status-changed: |
    - when: app.status.sync.status != ''
      send: [app-sync-status]
EOF
```

---

## 3. Argo Workflows 安装与配置

### 3.1 安装 Argo Workflows

```bash
# 创建命名空间
kubectl create namespace argo

# 安装 Argo Workflows（最小化版本）
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/latest/download/quick-start-minimal.yaml

# 生产环境安装（带持久化）
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/latest/download/install.yaml
```

### 3.2 配置 Workflow 权限

```yaml
# argo-rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: workflow-builder
  namespace: argo
rules:
- apiGroups: [""]
  resources: ["pods", "pods/log"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["argoproj.io"]
  resources: ["workflows", "workflowtemplates"]
  verbs: ["create", "get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: workflow-builder-binding
  namespace: argo
subjects:
- kind: ServiceAccount
  name: argo-workflows
  namespace: argo
roleRef:
  kind: Role
  name: workflow-builder
  apiGroup: rbac.authorization.k8s.io
```

### 3.3 部署平台 WorkflowTemplate

平台提供了 4 个预置 WorkflowTemplate（对应 4 种语言），在 `configs/argo-workflow-templates/` 目录下：

```bash
# 一键部署全部模板
kubectl apply -f configs/argo-workflow-templates/

# 验证
kubectl get workflowtemplate -n argo
# 输出:
# NAME                          AGE
# frontend-build-workflow       10s
# go-build-workflow             10s
# java-spring-build-workflow    10s
# python-build-workflow         10s
```

### 3.4 安装 Argo CLI

```bash
# macOS
brew install argo

# Linux
curl -sLO https://github.com/argoproj/argo-workflows/releases/latest/download/argo-linux-amd64.gz
gunzip argo-linux-amd64.gz
chmod +x argo-linux-amd64
sudo mv argo-linux-amd64 /usr/local/bin/argo

# 验证
argo version
```

---

## 4. 平台 GitOps 配置

### 4.1 config.yaml 配置

在平台配置文件 `configs/config.yaml` 中添加 GitOps 配置段：

```yaml
GitOps:
  # ArgoCD API Server 地址（集群内访问）
  ArgoCDURL: "https://argocd-server.argocd.svc.cluster.local"
  # ArgoCD Auth Token（优先级高于用户名密码）
  ArgoCDAuthToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
  
  # Argo Workflows Server 地址
  ArgoWorkflowsURL: "https://argo-workflows-server.argo.svc.cluster.local:2746"
  # Argo Workflows Auth Token
  ArgoWorkflowsToken: "Bearer eyJhbGciOiJSUzI1NiIs..."
  
  # 平台回调地址（Argo Workflow 构建完成后回调用）
  CallbackURL: "http://k8soperation-backend.k8soperation.svc.cluster.local:38180"
  
  # HMAC 签名密钥（与 ArgoCD webhook 和 Workflow 回调共用）
  HMACSecret: "your-hmac-secret-key-change-me"
  
  # 默认 Git Manifest 仓库（新建应用时可覆盖）
  GitManifestRepo: "https://git.example.com/ops/manifests.git"
  GitManifestPath: "manifests"
  GitCredentialID: "git-ssh-key"
  
  # 通知配置
  EnableDingTalk: false
  DingTalkWebhook: ""
  EnableFeishu: false
  FeishuWebhook: ""
  FeishuSecret: ""
  
  # 同步状态轮询配置（Webhook 丢失去时的 fallback）
  SyncPollInterval: 10   # 轮询间隔（秒）
  SyncMaxWaitTime: 600   # 最大等待时间（秒）
```

### 4.2 获取 ArgoCD Auth Token

```bash
# 方式 1: 使用 admin 用户生成长期 token
argocd account generate-token --account admin

# 方式 2: 创建专用的平台 ServiceAccount
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8soperation-platform
  namespace: argocd
EOF

# 创建 Token（K8s 1.24+ 需要手动创建 Secret）
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-platform-token
  namespace: argocd
  annotations:
    kubernetes.io/service-account.name: k8soperation-platform
type: kubernetes.io/service-account-token
EOF

# 获取 Token
kubectl -n argocd get secret k8soperation-platform-token -o jsonpath="{.data.token}" | base64 -d
```

### 4.3 获取 Argo Workflows Token

```bash
# 创建 ServiceAccount + Token
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8soperation-platform
  namespace: argo
---
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-platform-token
  namespace: argo
  annotations:
    kubernetes.io/service-account.name: k8soperation-platform
type: kubernetes.io/service-account-token
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8soperation-workflow-manager
rules:
- apiGroups: ["argoproj.io"]
  resources: ["workflows", "workflowtemplates", "workflows/finalizers"]
  verbs: ["create", "get", "list", "watch", "update", "patch", "delete"]
- apiGroups: [""]
  resources: ["pods", "pods/log"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8soperation-workflow-manager
subjects:
- kind: ServiceAccount
  name: k8soperation-platform
  namespace: argo
roleRef:
  kind: ClusterRole
  name: k8soperation-workflow-manager
  apiGroup: rbac.authorization.k8s.io
EOF

# 获取 Token
kubectl -n argo get secret k8soperation-platform-token -o jsonpath="{.data.token}" | base64 -d
```

---

## 5. 新增应用：完整操作流程

以下以 **Go 语言后端服务 "user-service"** 为例，演示从零开始接入 GitOps 的完整流程。

### 5.1 前置准备清单

| 序号 | 准备项 | 说明 |
|------|--------|------|
| 1 | 源码仓库 | `https://git.example.com/team/user-service.git` |
| 2 | 镜像仓库 | `registry.example.com/team/user-service` |
| 3 | Manifest 仓库 | `https://git.example.com/ops/manifests.git` (已有) |
| 4 | K8s 集群 | 已接入平台，目标命名空间 `team-apps` |
| 5 | ArgoCD | 已安装并运行 |
| 6 | Argo Workflows | 已安装并部署平台模板 |
| 7 | Git 凭据 | ArgoCD / Argo Workflows 已配置 Git 访问凭据 |
| 8 | 镜像仓库凭据 | K8s 集群中已有 `registry-credentials` Secret |

### 5.2 第一步：准备 Manifest 仓库

在 Manifest 仓库中创建应用的部署清单目录结构：

```bash
# Clone manifest 仓库
git clone https://git.example.com/ops/manifests.git
cd manifests

# 创建应用目录
mkdir -p apps/user-service/base
mkdir -p apps/user-service/overlays/dev
mkdir -p apps/user-service/overlays/prod
```

**base/deployment.yaml**（基础模板）：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  labels:
    app: user-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: registry.example.com/team/user-service:placeholder  # Argo Workflow 会自动更新
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 15
          periodSeconds: 20
        env:
        - name: APP_ENV
          value: "production"
```

**base/service.yaml**：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
  labels:
    app: user-service
spec:
  selector:
    app: user-service
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

**base/kustomization.yaml**：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- deployment.yaml
- service.yaml
images:
- name: registry.example.com/team/user-service
  newTag: placeholder
```

**overlays/prod/kustomization.yaml**：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: team-apps
resources:
- ../../base
patches:
- patch: |-
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: user-service
    spec:
      replicas: 3
      template:
        spec:
          containers:
          - name: user-service
            resources:
              requests:
                memory: "256Mi"
                cpu: "200m"
```

提交到 Git：

```bash
git add apps/user-service/
git commit -m "feat: add user-service deployment manifests"
git push origin main
```

### 5.3 第二步：在 ArgoCD 中创建 Application

```bash
# 方式 1: ArgoCD CLI
argocd app create user-service-prod \
  --repo https://git.example.com/ops/manifests.git \
  --path apps/user-service/overlays/prod \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace team-apps \
  --sync-policy automated \
  --auto-prune \
  --self-heal

# 验证
argocd app get user-service-prod
```

**方式 2: YAML 声明式**（推荐，可版本控制）：

```yaml
# user-service-app.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: user-service-prod
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://git.example.com/ops/manifests.git
    targetRevision: main
    path: apps/user-service/overlays/prod
  destination:
    server: https://kubernetes.default.svc
    namespace: team-apps
  syncPolicy:
    automated:
      prune: true        # 自动删除 Git 中不存在的资源
      selfHeal: true     # 自动修复手动修改的资源
    syncOptions:
    - CreateNamespace=true
    - PruneLast=true
```

```bash
kubectl apply -f user-service-app.yaml
```

### 5.4 第三步：在平台创建 GitOps 流水线

登录平台 → 【CI/CD】→ 【流水线】→ 点击 **创建流水线**：

**Step 1: 基本信息**

| 字段 | 值 | 说明 |
|------|-----|------|
| 应用名称 | `user-service` | 平台内唯一标识 |
| 描述 | `用户服务 - 生产环境` | |
| 语言类型 | `go` | 对应 go-build-workflow 模板 |

**Step 2: Git 配置**

| 字段 | 值 |
|------|-----|
| Git 仓库 | `https://git.example.com/team/user-service.git` |
| Git 分支 | `main` |

**Step 3: ✨ 部署模式选择**

选择 **🔄 GitOps（ArgoCD 拉模式）**：

| 字段 | 值 | 说明 |
|------|-----|------|
| ArgoCD 应用名称 | `user-service-prod` | 与 ArgoCD Application 名称一致 |
| Manifest Git 仓库 | `https://git.example.com/ops/manifests.git` | 存放 K8s 清单的仓库 |
| Manifest 路径 | `apps/user-service/overlays/prod` | ArgoCD 同步路径 |
| ArgoCD Project | `default` | |
| 目标分支/Tag | `main` | |
| 镜像仓库地址 | `registry.example.com/team` | 不含镜像名 |
| 镜像名称 | `user-service` | |
| Dockerfile 路径 | `Dockerfile` | |
| 自动同步 | ✅ 开启 | ArgoCD 自动同步变更 |
| 资源清理 | ✅ 开启 | 自动删除多余资源 |

**Step 4: 部署配置**（可选）

| 字段 | 值 |
|------|-----|
| 自动部署 | ✅ 开启 |
| 目标集群 | 选择目标 K8s 集群 |
| 目标命名空间 | `team-apps` |
| 工作负载类型 | `Deployment` |
| 工作负载名称 | `user-service` |
| 容器名称 | `user-service` |
| 部署环境 | `prod` |

点击 **创建** → 流水线创建成功，进入详情页。

### 5.5 第四步：运行流水线

1. 在流水线详情页点击 **▶ 运行**
2. 选择分支（默认 `main`），点击确认
3. 平台自动提交 Argo Workflow，流程如下：

```
┌────────────────────────────────────────────────────────────┐
│ ① checkout        ── git clone 源代码                       │
│      状态: ✅ 完成 | 耗时: 8s  | Git SHA: a1b2c3d           │
├────────────────────────────────────────────────────────────┤
│ ② go-test         ── go vet + go test                      │
│      状态: ✅ 完成 | 耗时: 45s | 覆盖率: 87.3%              │
├────────────────────────────────────────────────────────────┤
│ ③ kaniko-build    ── 构建 Docker 镜像 + 推送到 Registry     │
│      状态: ✅ 完成 | 耗时: 2m12s                            │
│      镜像: registry.example.com/team/user-service:main-a1b2c3d │
├────────────────────────────────────────────────────────────┤
│ ④ update-manifest ── 更新 Git Manifest 中 image tag         │
│      状态: ✅ 完成 | 耗时: 5s                               │
│      更新: apps/user-service/overlays/prod/deployment.yaml  │
├────────────────────────────────────────────────────────────┤
│ ⑤ callback        ── 回调平台更新构建状态                    │
│      状态: ✅ 完成                                          │
└────────────────────────────────────────────────────────────┘
                          │
          ┌───────────────▼────────────────┐
          │   ArgoCD 检测到 Git Manifest    │
          │   中 image tag 变更             │
          │   → Auto Sync 触发              │
          │   → Deployment 滚动更新         │
          │   → Health: Healthy ✅          │
          │   → Sync: Synced ✅             │
          └───────────────┬────────────────┘
                          │
          ┌───────────────▼────────────────┐
          │   ArgoCD Webhook 回调平台       │
          │   → 平台更新部署状态: 已同步     │
          │   → 发送通知（钉钉/飞书）        │
          └────────────────────────────────┘
```

### 5.6 第五步：验证部署结果

```bash
# 查看 Deployment
kubectl -n team-apps get deployment user-service

# 查看 Pod
kubectl -n team-apps get pods -l app=user-service

# 查看 ArgoCD 同步状态
argocd app get user-service-prod

# 平台界面：流水线详情 → 同步状态显示 ✅ Synced
```

---

## 6. Argo Workflow 模板详解

### 6.1 模板与语言类型映射

| 语言类型 | WorkflowTemplate 名称 | 文件名 | 构建工具 |
|---------|----------------------|--------|---------|
| `go` | `go-build-workflow` | `go-workflow.yaml` | Go 1.24 + go vet/test |
| `java` | `java-spring-build-workflow` | `java-spring-workflow.yaml` | Maven 3.9 + SonarQube |
| `frontend` | `frontend-build-workflow` | `frontend-workflow.yaml` | Node 22 + npm |
| `python` | `python-build-workflow` | `python-workflow.yaml` | Python 3.11 + flake8/pytest |

### 6.2 平台注入参数

每个 Workflow 运行时，平台自动注入以下参数：

| 参数名 | 说明 | 示例值 |
|-------|------|--------|
| `git_repo` | 应用源码 Git 仓库 | `https://git.example.com/team/user-service.git` |
| `git_branch` | 构建分支 | `main` |
| `pipeline_id` | 平台流水线 ID | `42` |
| `run_id` | 平台运行记录 ID | `1287` |
| `platform_callback_url` | 回调地址 | `http://k8soperation.../api/v1/k8s/cicd/pipeline/callback` |
| `image_registry` | 镜像仓库地址 | `registry.example.com/team` |
| `image_repo` | 镜像名称 | `user-service` |
| `git_manifest_repo` | K8s 清单仓库 | `https://git.example.com/ops/manifests.git` |
| `manifest_path` | 清单路径 | `apps/user-service/overlays/prod` |
| `argo_app_name` | ArgoCD 应用名 | `user-service-prod` |
| `language_type` | 语言类型 | `go` |

### 6.3 自定义 Workflow 模板

如果预置模板不满足需求，可以创建自定义 WorkflowTemplate：

```yaml
# my-custom-workflow.yaml
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: my-custom-build-workflow
  namespace: argo
spec:
  entrypoint: build
  arguments:
    parameters:
      - name: git_repo
      - name: git_branch
      - name: image_registry
      - name: image_repo
      - name: pipeline_id
      - name: run_id
      - name: platform_callback_url
      - name: git_manifest_repo
      - name: manifest_path
      - name: argo_app_name

  templates:
    - name: build
      steps:
        - - name: checkout
            template: clone
        - - name: custom-build
            template: build-step
        - - name: update-manifest
            template: manifest-update
        - - name: callback
            template: notify-platform

    - name: clone
      container:
        image: alpine/git:latest
        command: [sh, -c]
        args:
          - git clone --depth 1 --branch {{workflow.parameters.git_branch}} {{workflow.parameters.git_repo}} /workspace

    - name: build-step
      container:
        image: your-custom-builder:latest
        command: [sh, -c]
        args:
          - cd /workspace && ./build.sh
        volumeMounts:
          - name: workspace
            mountPath: /workspace

    # ... 其他 step 参考预置模板
```

部署并使用：

```bash
kubectl apply -f my-custom-workflow.yaml

# 在平台创建流水线时：
# 语言类型选择 "custom"
# GitOps 配置中手动指定 WorkflowTemplate: my-custom-build-workflow
```

---

## 7. ArgoCD Application 管理

### 7.1 常用操作

```bash
# ====== 查看 ======
argocd app list                          # 列出所有 Application
argocd app get user-service-prod         # 查看详情
argocd app diff user-service-prod        # 对比 Git 与 K8s 差异
argocd app history user-service-prod     # 同步历史

# ====== 同步 ======
argocd app sync user-service-prod        # 手动同步
argocd app sync user-service-prod --prune  # 同步并删除多余资源

# ====== 回滚 ======
argocd app rollback user-service-prod 5  # 回滚到历史版本 5

# ====== 暂停/恢复 ======
argocd app set user-service-prod --sync-policy none  # 暂停自动同步（维护窗口）
argocd app set user-service-prod --sync-policy automated  # 恢复自动同步

# ====== 删除 ======
argocd app delete user-service-prod      # 删除 Application（不影响 K8s 资源）
argocd app delete user-service-prod --cascade  # 级联删除 K8s 资源
```

### 7.2 手动触发同步（通过平台）

平台也支持手动触发 ArgoCD 同步：

```bash
# API 方式
curl -X POST http://platform:38180/api/v1/k8s/cicd/gitops/sync \
  -H "Content-Type: application/json" \
  -d '{"pipeline_id": 42}'
```

或在流水线详情页点击 **🔄 手动同步** 按钮。

### 7.3 健康状态说明

| 同步状态 | 含义 |
|---------|------|
| `Synced` | Git 中的声明与集群一致 ✅ |
| `OutOfSync` | 集群中的资源与 Git 有差异 ⚠️ |
| `Unknown` | 无法判断（网络问题等）❓ |

| 健康状态 | 含义 |
|---------|------|
| `Healthy` | 所有资源正常运行 ✅ |
| `Progressing` | 正在部署/更新中 🔄 |
| `Degraded` | 资源异常 ❌ |
| `Suspended` | 已暂停（如 CronJob suspended） |
| `Missing` | 资源在 Git 中声明但集群中不存在 |

---

## 8. 镜像更新与自动同步

### 8.1 触发方式

| 方式 | 适用场景 |
|------|---------|
| **平台触发** | 开发人员在平台点击"运行" |
| **Git Push 触发** | 配置 Webhook，代码推送自动触发 |
| **定时触发** | 配置 CronWorkflow，如每日构建 |
| **手动 ArgoCD 同步** | 紧急修复时直接改 Manifest Git 仓库 |

### 8.2 镜像 Tag 策略

推荐镜像命名策略：

```
<registry>/<repo>:<branch>-<short-sha>
```

示例：
```
registry.example.com/team/user-service:main-a1b2c3d
registry.example.com/team/user-service:release-v1.2.0
registry.example.com/team/user-service:prod-20260708
```

### 8.3 ArgoCD Image Updater（可选）

如需自动检测镜像仓库中新 tag 并更新（不经过 CI Workflow）：

```bash
# 安装 ArgoCD Image Updater
kubectl apply -n argocd -f https://github.com/argoproj-labs/argocd-image-updater/stable/manifests/install.yaml

# 给 Application 添加注解
kubectl annotate app -n argocd user-service-prod \
  argocd-image-updater.argoproj.io/image-list=user-service=registry.example.com/team/user-service \
  argocd-image-updater.argoproj.io/user-service.update-strategy=semver \
  argocd-image-updater.argoproj.io/write-back-method=git:secret:argocd/git-creds
```

---

## 9. 多环境管理

### 9.1 推荐目录结构

```
manifests/
├── apps/
│   └── user-service/
│       ├── base/                    # 基础配置（所有环境共享）
│       │   ├── deployment.yaml
│       │   ├── service.yaml
│       │   └── kustomization.yaml
│       └── overlays/
│           ├── dev/                 # 开发环境
│           │   ├── kustomization.yaml
│           │   └── patch-replicas.yaml
│           ├── staging/             # 预发环境
│           │   ├── kustomization.yaml
│           │   └── patch-resources.yaml
│           └── prod/                # 生产环境
│               ├── kustomization.yaml
│               ├── patch-replicas.yaml
│               └── patch-ingress.yaml
```

### 9.2 平台中创建多环境流水线

| 环境 | 流水线名称 | Manifest Path |
|------|-----------|--------------|
| 开发 | `user-service-dev` | `apps/user-service/overlays/dev` |
| 预发 | `user-service-staging` | `apps/user-service/overlays/staging` |
| 生产 | `user-service-prod` | `apps/user-service/overlays/prod` |

每个流水线对应一个 ArgoCD Application，环境间完全隔离。

### 9.3 环境晋级流程

```
dev 构建验证通过
  → 合并到 staging 分支
    → 触发 staging 流水线
      → staging 环境验证通过
        → 创建生产环境发布单（审批）
          → 审批通过 → 合并到 main
            → 触发生产流水线 → ArgoCD 同步生产环境
```

---

## 10. 故障排查

### 10.1 流水线创建后无法运行

**现象**：点击运行后报错 "GitOps 模式未配置"

**排查**：
```bash
# 1. 检查平台 config.yaml 是否配置 GitOps 段
grep -A 5 "GitOps:" configs/config.yaml

# 2. 检查平台日志
kubectl logs -n k8soperation deployment/k8soperation-backend | grep "GitOps"

# 3. 确认 Argo Workflows 地址可达
curl -k https://argo-workflows-server.argo.svc.cluster.local:2746/api/v1/workflows
```

### 10.2 Workflow 提交失败

**现象**：流水线状态一直是 "pending"，未创建 Workflow

**排查**：
```bash
# 1. 检查 WorkflowTemplate 是否存在
kubectl get workflowtemplate -n argo | grep build-workflow

# 2. 检查 Argo Workflows Server 是否运行
kubectl -n argo get pods | grep workflow-server

# 3. 查看 Workflow 日志
argo logs -n argo @latest
argo list -n argo
```

### 10.3 构建成功但部署未更新

**现象**：Workflow 成功完成但 K8s 中镜像未更新

**排查**：
```bash
# 1. 检查 Manifest Git 仓库是否有新提交
git log --oneline -5

# 2. 检查 ArgoCD 同步状态
argocd app get user-service-prod

# 3. 查看 ArgoCD diff
argocd app diff user-service-prod

# 4. 手动触发同步
argocd app sync user-service-prod

# 5. 检查 manifest_path 是否正确
# 平台配置: apps/user-service/overlays/prod
# ArgoCD Application: 检查 spec.source.path 是否一致
```

### 10.4 ArgoCD 同步一直 OutOfSync

**常见原因**：

1. **image tag 格式不匹配**：Kustomize 中 `newTag` 与实际构建的 tag 不一致
2. **Manifest 路径错误**：平台配置的 `manifest_path` 与 ArgoCD Application 的 `path` 不匹配
3. **Git push 失败**：Workflow 中的 update-manifest 步骤没有权限推送
4. **Auto Sync 未开启**：ArgoCD Application 设置了 manual sync

### 10.5 平台同步状态不更新

**现象**：ArgoCD 已成功同步但平台仍显示 syncing

**排查**：
```bash
# 1. 检查 ArgoCD webhook 配置
kubectl -n argocd get cm argocd-notifications-cm -o yaml

# 2. 确认平台回调地址可达
curl http://k8soperation-backend.k8soperation.svc.cluster.local:38180/api/v1/k8s/cicd/gitops/sync-callback

# 3. 检查 SyncPollWorker 日志（webhook 丢失时的 fallback）
kubectl logs -n k8soperation deployment/k8soperation-backend | grep "GitOps.*Worker"
```

---

## 11. 附录：命令速查表

### ArgoCD

```bash
# 安装
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 登录
argocd login <SERVER> --username admin --password <PASSWORD>

# 应用管理
argocd app list                          # 列表
argocd app get <APP>                     # 详情
argocd app create <APP> --repo <URL> --path <PATH> --dest-server <SERVER> --dest-namespace <NS>
argocd app sync <APP>                    # 同步
argocd app diff <APP>                    # 差异对比
argocd app rollback <APP> <REVISION>     # 回滚
argocd app delete <APP>                  # 删除

# 仓库管理
argocd repo add <URL> --ssh-private-key-path ~/.ssh/id_rsa
argocd repo list
```

### Argo Workflows

```bash
# 安装
kubectl create namespace argo
kubectl apply -n argo -f https://github.com/argoproj/argo-workflows/releases/latest/download/quick-start-minimal.yaml

# 提交/查看
argo submit -n argo --from workflowtemplate/go-build-workflow -p git_repo=<URL>
argo list -n argo
argo get -n argo <WORKFLOW>
argo logs -n argo <WORKFLOW>
argo terminate -n argo <WORKFLOW>
```

### 平台 GitOps API

```bash
# 同步状态回调 (ArgoCD → 平台)
POST /api/v1/k8s/cicd/gitops/sync-callback
Body: { "app_name": "...", "sync_status": "Synced", "sync_revision": "abc123" }

# Workflow Webhook 回调
POST /api/v1/k8s/cicd/gitops/webhook
Body: { "pipeline_id": 42, "run_id": 1287, "status": "SUCCESS", "image": "registry.../app:v1" }

# 查询同步状态
GET /api/v1/k8s/cicd/gitops/app-status?app_name=user-service-prod
```

### 日常操作流程

```bash
# 开发者：修改代码 + 推送
git add . && git commit -m "feat: add user login api" && git push origin main

# 平台自动：运行流水线 → Workflow 构建 → 更新 Manifest → ArgoCD 同步
# 查看进度：平台流水线详情页 或
argo watch -n argo <WORKFLOW>

# 验证部署
kubectl -n team-apps rollout status deployment/user-service
argocd app get user-service-prod
```

---

> **提示**：本文档也适用于已有 Jenkins 流水线的存量应用。创建新流水线时选择 **Jenkins 模式** 可继续使用原有 CI/CD 流程，两套模式并存互不影响。
