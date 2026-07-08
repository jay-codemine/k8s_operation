# K8sOperation 平台 + Jenkins 容器化完整部署文档

> **技术栈**：Go + Gin（后端） | Vue3 + Vite + Nginx（前端） | Jenkins LTS + JCasC + K8s 动态 Agent  
> **集群方案**：Kind（3 节点） | Kustomize 声明式部署  
> **版本**：Kubernetes v1.33.x | Go 1.24 | Node 22 | Jenkins LTS

---

## 目录

1. [架构总览](#1-架构总览)
2. [环境前置条件](#2-环境前置条件)
3. [Kind 集群创建](#3-kind-集群创建)
4. [镜像构建](#4-镜像构建)
5. [K8s 资源部署](#5-k8s-资源部署)
6. [Jenkins 配置与集成](#6-jenkins-配置与集成)
7. [中间件依赖](#7-中间件依赖)
8. [访问方式与验证](#8-访问方式与验证)
9. [配置详解](#9-配置详解)
10. [常见问题与排查](#10-常见问题与排查)

---

## 1. 架构总览

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Kind 集群 (3 节点)                            │
│  ┌──────────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ control-plane    │  │   worker     │  │   worker2    │         │
│  │ (port:30080/81)  │  │              │  │              │         │
│  └──────────────────┘  └──────────────┘  └──────────────┘         │
│                                                                     │
│  Namespace: k8soperation              Namespace: devops             │
│  ┌───────────────────────┐            ┌────────────────────┐       │
│  │ Deployment:           │            │ StatefulSet:       │       │
│  │   k8soperation (BE)   │◄──────────►│   jenkins (Master) │       │
│  │   k8soperation-web(FE)│            │                    │       │
│  └───────────────────────┘            │ 动态 Pod Agent     │       │
│           │                            └────────────────────┘       │
│           ▼                                                         │
│  ┌───────────────────────┐                                         │
│  │ 宿主机 (host.docker.internal)                                   │
│  │   MySQL :3307                                                   │
│  │   Redis :6380                                                   │
│  └───────────────────────┘                                         │
└─────────────────────────────────────────────────────────────────────┘
```

### 组件关系

| 组件 | 类型 | Namespace | 说明 |
|------|------|-----------|------|
| k8soperation | Deployment (1副本) | k8soperation | 平台后端 API |
| k8soperation-web | Deployment (2副本) | k8soperation | 平台前端 Nginx |
| jenkins | StatefulSet (1副本) | devops | CI/CD 引擎 |
| MySQL | 外部服务 | - | 宿主机 3307 端口 |
| Redis | 外部服务 | - | 宿主机 6380 端口 |

### 网络通信

- **前端 → 后端**：Nginx 反向代理到 `k8soperation.k8soperation.svc.cluster.local:8080`
- **后端 → Jenkins**：通过 `http://jenkins.devops.svc.cluster.local:8080/` 调用 API
- **后端 → K8s API**：InCluster 模式，ServiceAccount 自动认证
- **后端 → MySQL/Redis**：通过 `host.docker.internal` 访问宿主机
- **Jenkins → K8s API**：ServiceAccount `jenkins` + ClusterRole 动态创建 Agent Pod

---

## 2. 环境前置条件

### 2.1 必需软件

| 软件 | 版本要求 | 用途 |
|------|----------|------|
| Docker Desktop | 最新版 | 容器运行时 + Kind 底座 |
| Kind | v0.20+ | 本地 K8s 集群 |
| kubectl | v1.28+ | 集群管理 CLI |
| Go | 1.24+ | 后端编译（可选，Docker 内编译） |
| Node.js | 22+ | 前端构建（可选，Docker 内构建） |

### 2.2 宿主机服务

```bash
# MySQL（端口 3307）
# 数据库名：k8s-platform，用户：root，密码：admin123

# Redis（端口 6380）
# 密码：admin123
```

### 2.3 Docker Desktop 配置

确保 Docker Desktop 开启以下设置：
- Kubernetes → Enable Kubernetes（可选，Kind 独立运行不依赖）
- Resources → Memory ≥ 8GB（推荐）
- Settings → General → 确认 `host.docker.internal` DNS 可用

---

## 3. Kind 集群创建

### 3.1 集群配置文件

创建 `kind-config.yaml`：

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    extraPortMappings:
      # Jenkins NodePort
      - containerPort: 30080
        hostPort: 30080
        protocol: TCP
      # 前端 NodePort
      - containerPort: 30081
        hostPort: 30081
        protocol: TCP
  - role: worker
  - role: worker
```

### 3.2 创建集群

```bash
# 创建集群
kind create cluster --name desktop --config kind-config.yaml

# 验证
kubectl cluster-info
kubectl get nodes
# 预期输出：
# desktop-control-plane   Ready   control-plane   ...   v1.33.x
# desktop-worker          Ready   <none>          ...   v1.33.x
# desktop-worker2         Ready   <none>          ...   v1.33.x
```

### 3.3 加载镜像到 Kind

```bash
# 构建后需要将镜像加载到 Kind 集群中
kind load docker-image devops-be:latest --name desktop
kind load docker-image devops-fe:latest --name desktop
# Jenkins 使用公共镜像，Kind 会自动拉取（或预加载）
kind load docker-image jenkins/jenkins:lts --name desktop
```

---

## 4. 镜像构建

### 4.1 后端镜像

**Dockerfile 位置**：`docker/backend/Dockerfile`

**技术特点**：
- 多阶段构建：`golang:1.24-alpine` → `alpine:3.20`
- 静态编译（CGO_ENABLED=0），最终镜像 < 25MB
- 国内镜像加速（Go Proxy: goproxy.cn，Alpine: aliyun）
- 非 root 用户运行（UID=1000）
- 内置健康检查

```bash
# 在项目根目录执行
docker build -f docker/backend/Dockerfile -t devops-be:latest .

# Apple Silicon（M1/M2）构建 Linux amd64 镜像
docker build -f docker/backend/Dockerfile --platform linux/amd64 -t devops-be:latest .

# 推送到阿里云 Registry
docker tag devops-be:latest registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest
```

**构建产物**：
- 二进制：`/app/devops-be`
- 内置模板：`/app/configs/jenkins-templates/`、`/app/configs/dockerfile-templates/`

### 4.2 前端镜像

**Dockerfile 位置**：`docker/frontend/Dockerfile`

**技术特点**：
- 多阶段构建：`node:22-alpine` → `nginx:1.27-alpine`
- 内置 Nginx 配置：API 反向代理 + Gzip + 静态资源缓存 + WebSocket
- 支持运行时通过 `API_BACKEND_URL` 环境变量动态配置后端地址
- Vue Router History 模式（`try_files`）
- 安全头：X-Frame-Options、X-Content-Type-Options、X-XSS-Protection

```bash
# 在项目根目录执行（注意 context 是 ./k8s-web）
docker build -f docker/frontend/Dockerfile -t devops-fe:latest ./k8s-web

# Apple Silicon 构建
docker build -f docker/frontend/Dockerfile --platform linux/amd64 -t devops-fe:latest ./k8s-web

# 推送到阿里云 Registry
docker tag devops-fe:latest registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest
```

### 4.3 加载镜像到 Kind 集群

```bash
# 将本地构建的镜像加载到 Kind（避免从 Registry 拉取）
kind load docker-image devops-be:latest --name desktop
kind load docker-image devops-fe:latest --name desktop
```

---

## 5. K8s 资源部署

### 5.1 部署目录结构

```
deploy/
├── kustomization.yaml          # 顶层：前端 + 后端一键部署
├── backend/
│   ├── kustomization.yaml      # 后端 Kustomization
│   ├── namespace.yaml          # Namespace: k8soperation
│   ├── secret.yaml             # 敏感配置（DB/Redis/JWT/Jenkins）
│   ├── configmap.yaml          # 应用配置模板（config.yaml）
│   ├── pv.yaml                 # PersistentVolume（hostPath）
│   ├── pvc.yaml                # PersistentVolumeClaim
│   ├── service.yaml            # Service + ServiceAccount + RBAC
│   └── deployment.yaml         # 后端 Deployment
├── frontend/
│   ├── kustomization.yaml      # 前端 Kustomization
│   ├── configmap.yaml          # Nginx 配置（可选覆盖）
│   ├── deployment.yaml         # 前端 Deployment（2副本）
│   ├── service.yaml            # ClusterIP Service
│   └── service-nodeport.yaml   # NodePort Service（端口 30081）
└── jenkins/
    ├── kustomization.yaml      # Jenkins Kustomization
    ├── namespace.yaml          # Namespace: devops
    ├── rbac.yaml               # ServiceAccount + ClusterRole
    ├── secret.yaml             # 管理员密码 + HMAC 密钥
    ├── pvc.yaml                # Jenkins 数据 + 构建缓存 PVC
    ├── configmap.yaml          # JCasC 配置
    ├── statefulset.yaml        # Jenkins Master StatefulSet
    └── service.yaml            # ClusterIP + Agent + NodePort
```

### 5.2 一键部署命令

```bash
# ===== 方式一：分模块部署（推荐） =====

# 1. 部署后端（含 namespace、RBAC、PV/PVC、ConfigMap、Secret、Service、Deployment）
kubectl apply -k deploy/backend/

# 2. 部署前端（含 Deployment、Service、NodePort）
kubectl apply -k deploy/frontend/

# 3. 部署 Jenkins（独立 namespace: devops）
kubectl apply -k deploy/jenkins/

# ===== 方式二：前后端一键部署 =====
kubectl apply -k deploy/
# 注意：Jenkins 需单独部署（不同 namespace）
kubectl apply -k deploy/jenkins/
```

### 5.3 验证部署状态

```bash
# 检查后端
kubectl -n k8soperation get pods,svc,pvc
# 预期：k8soperation-xxx  1/1  Running

# 检查前端
kubectl -n k8soperation get pods -l app.kubernetes.io/component=frontend
# 预期：k8soperation-web-xxx  1/1  Running（2个副本）

# 检查 Jenkins
kubectl -n devops get pods,svc,pvc
# 预期：jenkins-0  1/1  Running

# 查看所有资源
kubectl -n k8soperation get all
kubectl -n devops get all
```

---

## 6. Jenkins 配置与集成

### 6.1 Jenkins 架构

- **部署方式**：StatefulSet（保证稳定网络标识 + 持久存储）
- **Master 模式**：`numExecutors: 0`，Master 不执行构建
- **Agent 模式**：Kubernetes Cloud 动态 Pod Agent
- **配置管理**：JCasC（Configuration as Code）自动配置
- **插件安装**：initContainer 通过 `jenkins-plugin-cli` 预装

### 6.2 预装插件列表

| 插件 | 用途 |
|------|------|
| kubernetes | K8s Cloud 动态 Agent |
| workflow-aggregator | Pipeline 核心 |
| git | Git SCM 集成 |
| configuration-as-code | JCasC 支持 |
| credentials-binding | 凭证绑定 |
| pipeline-stage-view | Pipeline 可视化 |
| blueocean | 新版 UI |
| sonar | SonarQube 代码质量 |
| timestamper | 构建时间戳 |
| ws-cleanup | 工作空间清理 |

### 6.3 Kubernetes Cloud 配置

```yaml
# JCasC 自动配置（deploy/jenkins/configmap.yaml）
clouds:
  - kubernetes:
      name: "kubernetes"
      serverUrl: "https://kubernetes.default.svc.cluster.local"
      namespace: "devops"
      jenkinsUrl: "http://jenkins.devops.svc.cluster.local:8080"
      jenkinsTunnel: "jenkins-agent.devops.svc.cluster.local:50000"
      containerCapStr: "20"        # 最大并发 Agent 数
      waitForPodSec: 600           # Agent Pod 启动超时
```

### 6.4 Jenkins RBAC 权限

Jenkins ServiceAccount 拥有以下 ClusterRole 权限：
- Pod 管理（create/delete/get/list/watch/patch）- 动态 Agent
- Pod exec（创建构建容器内命令）
- Pod log（查看构建日志）
- Secret 读取（拉取私有镜像）
- PVC 管理（构建缓存）

### 6.5 Jenkins 构建缓存

| PVC 名称 | 容量 | 用途 |
|----------|------|------|
| jenkins-data | 20Gi | Jenkins 主数据（jobs/plugins/config） |
| jenkins-go-cache | 10Gi | Go module 缓存 |
| jenkins-maven-cache | 20Gi | Maven 本地仓库缓存 |
| jenkins-npm-cache | 10Gi | npm 缓存 |
| jenkins-pip-cache | 5Gi | pip 缓存 |

### 6.6 平台与 Jenkins 集成

后端通过以下配置与 Jenkins 通信：

```yaml
Jenkins:
  URL: "http://jenkins.devops.svc.cluster.local:8080/"
  Username: "ops-dev"
  APIToken: "<Jenkins API Token>"
  CallbackURL: "http://k8soperation.k8soperation.svc:8080"  # Jenkins 回调平台
  HMACSecret: "<HMAC签名密钥>"                                # 回调验签
```

---

## 7. 中间件依赖

### 7.1 MySQL

| 配置项 | 值 |
|--------|-----|
| Host | host.docker.internal（Kind Pod 内访问宿主机） |
| Port | 3307 |
| Database | k8s-platform |
| Username | root |
| Password | admin123 |
| Charset | utf8mb4 |
| MaxOpenConns | 100 |
| MaxIdleConns | 10 |

### 7.2 Redis

| 配置项 | 值 |
|--------|-----|
| Address | host.docker.internal:6380 |
| Password | admin123 |
| Network | tcp |

### 7.3 连接说明

在 Kind 集群中，Pod 通过 `host.docker.internal` DNS 名称访问宿主机运行的 MySQL 和 Redis。这是 Docker Desktop 内置的 DNS 解析，将该域名指向宿主机 IP。

**注意**：如果使用 Linux 原生 Docker（非 Docker Desktop），需要使用 `--add-host=host.docker.internal:host-gateway` 或直接配置宿主机 IP。

---

## 8. 访问方式与验证

### 8.1 服务暴露端口

| 服务 | 类型 | 端口 | 访问地址 |
|------|------|------|----------|
| Jenkins Web UI | NodePort | 30080 | http://localhost:30080 |
| 平台前端 | NodePort | 30081 | http://localhost:30081 |
| 平台后端 API | ClusterIP | 8080 | 集群内部 k8soperation.k8soperation.svc:8080 |
| Jenkins Agent | ClusterIP | 50000 | 集群内部 jenkins-agent.devops.svc:50000 |

### 8.2 验证步骤

```bash
# 1. 验证后端 API 健康
kubectl -n k8soperation exec -it deploy/k8soperation -- wget -qO- http://localhost:8080/healthz/live
# 或通过前端 Nginx 代理：
curl http://localhost:30081/api/v1/health

# 2. 验证前端页面
curl -s http://localhost:30081/ | head -5
# 应返回 HTML 内容

# 3. 验证 Jenkins
curl -s http://localhost:30080/login
# 应返回 Jenkins 登录页面

# 4. 验证后端日志
kubectl -n k8soperation logs -f deploy/k8soperation --tail=50

# 5. 验证 Jenkins 日志
kubectl -n devops logs -f jenkins-0 --tail=50
```

### 8.3 Jenkins 登录凭证

| 用户名 | 密码 | 说明 |
|--------|------|------|
| ops-dev | admin123 | JCasC 自动配置的管理员账号 |

---

## 9. 配置详解

### 9.1 配置注入机制

平台采用 **环境变量模板替换** 机制：

1. `ConfigMap` 中的 `config.yaml` 使用 `${ENV_VAR}` 占位符
2. `Secret` 中存储实际敏感值（明文 stringData）
3. `Deployment` 将 Secret 的 key 注入为环境变量
4. 应用启动时通过 `os.ExpandEnv()` 展开配置文件中的占位符

```
Secret (敏感值) → ENV 环境变量 → ConfigMap 模板展开 → Viper 读取最终配置
```

### 9.2 后端 ConfigMap 关键配置

```yaml
Server:
  Port: 8080
  ReadTimeout: 3600       # 支持长连接（WebSocket 终端）
  WriteTimeout: 3600

App:
  JWTExpireTime: 120000   # JWT 有效期（秒）
  AutoInitK8s: true       # 自动初始化集群连接
  GlobalKubeConfigPath: "" # 留空 = InCluster 模式

Jenkins:
  CallbackURL: "http://k8soperation.k8soperation.svc:8080"
  PollInterval: 15        # 构建状态轮询间隔（秒）
  MaxBuildTime: 30        # 最大构建时间（分钟）
```

### 9.3 RBAC 权限体系

#### 后端 ServiceAccount (`k8soperation`)

拥有 ClusterRole 级别权限，可跨 namespace 管理：
- 核心资源：Pods、Services、ConfigMaps、Secrets、PV/PVC、Nodes、Namespaces
- 工作负载：Deployments、StatefulSets、DaemonSets、Jobs、CronJobs
- 网络：Ingresses
- 存储：StorageClasses
- RBAC：Roles、ClusterRoles、RoleBindings
- CRD：AppConfig 等自定义资源
- Metrics：节点和 Pod 监控指标

#### Jenkins ServiceAccount (`jenkins`)

拥有 ClusterRole 级别权限，用于动态 Agent：
- Pod 创建/删除/执行（构建 Agent）
- Pod 日志查看
- Secret 读取（镜像拉取凭证）
- PVC 管理（构建缓存）

### 9.4 持久化存储

#### 后端 PV/PVC

| PV/PVC | 容量 | hostPath | 用途 |
|--------|------|----------|------|
| k8soperation-artifacts | 20Gi | /data/k8soperation/artifacts | CI/CD 构建制品 |
| k8soperation-logs | 5Gi | /data/k8soperation/logs | 应用日志 |

#### Jenkins PVC

| PVC | 容量 | 用途 |
|-----|------|------|
| jenkins-data | 20Gi | Jenkins 主数据 |
| jenkins-go-cache | 10Gi | Go mod 缓存 |
| jenkins-maven-cache | 20Gi | Maven 仓库缓存 |
| jenkins-npm-cache | 10Gi | npm 缓存 |
| jenkins-pip-cache | 5Gi | pip 缓存 |

---

## 10. 常见问题与排查

### 10.1 集群不可用 (Cluster Unavailable)

**症状**：平台管理页面显示集群状态为"不可用"

**原因**：数据库中存储的 kubeconfig server 地址从 Pod 内不可达（如 `127.0.0.1:xxxx`）

**解决**：
```bash
# 1. 进入后端 Pod
kubectl -n k8soperation exec -it deploy/k8soperation -- sh

# 2. 获取 Pod 内 ServiceAccount Token
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
CA=$(cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt | base64 | tr -d '\n')

# 3. 生成正确的 kubeconfig（使用集群内部地址）
# server 地址必须是：https://kubernetes.default.svc

# 4. 通过平台 API 更新集群配置
curl -X PUT http://localhost:8080/api/v1/k8s/cluster/update \
  -H "Authorization: Bearer <token>" \
  -d '{"id": <cluster_id>, "kube_config": "<base64_kubeconfig>"}'

# 5. 重新初始化
curl -X POST http://localhost:8080/api/v1/k8s/cluster/init
```

### 10.2 Jenkins 插件安装超时

**症状**：initContainer `install-plugins` 长时间未完成

**解决**：
```bash
# 查看日志
kubectl -n devops logs jenkins-0 -c install-plugins

# 如果是网络问题，可预先下载插件 hpi 文件挂载
# 或配置 Jenkins 国内更新源
```

### 10.3 后端无法连接 MySQL/Redis

**症状**：后端 Pod CrashLoopBackOff，日志显示连接超时

**排查**：
```bash
# 1. 确认宿主机服务运行
# MySQL 监听 3307，Redis 监听 6380

# 2. 从 Pod 内测试连通性
kubectl -n k8soperation exec -it deploy/k8soperation -- \
  wget -qO- --timeout=3 http://host.docker.internal:3307 || echo "port open"

# 3. 确认 Docker Desktop 已启用 host.docker.internal
```

### 10.4 前端页面空白 / API 404

**症状**：前端页面加载但 API 请求返回 404 或 502

**排查**：
```bash
# 检查 Nginx upstream 配置
kubectl -n k8soperation exec -it deploy/k8soperation-web-xxx -- \
  cat /etc/nginx/conf.d/default.conf

# 确认后端 Service 存在
kubectl -n k8soperation get svc k8soperation

# 确认 DNS 解析
kubectl -n k8soperation exec -it deploy/k8soperation-web-xxx -- \
  nslookup k8soperation.k8soperation.svc.cluster.local
```

### 10.5 JWT Token 问题

**症状**：登录后立即跳回登录页

**排查**：
```bash
# 确认 JWT 配置
kubectl -n k8soperation get secret k8soperation-secret -o jsonpath='{.data.JWT_SIGNING_KEY}' | base64 -d

# 确认 ConfigMap 中 JWTExpireTime 值合理（单位：秒）
kubectl -n k8soperation get cm k8soperation-config -o yaml | grep JWTExpire
# 预期：JWTExpireTime: 120000
```

### 10.6 镜像拉取失败

**症状**：Pod 处于 ImagePullBackOff

**解决**：
```bash
# 方式一：确认 imagePullSecrets 存在
kubectl -n k8soperation get secret aliyun-registry

# 方式二：使用 Kind 本地加载
kind load docker-image devops-be:latest --name desktop
kind load docker-image devops-fe:latest --name desktop

# 方式三：修改 imagePullPolicy 为 IfNotPresent（Kind 本地加载后）
```

---

## 附录 A：完整部署快速命令

```bash
# ===== 一键部署全流程 =====

# 1. 创建 Kind 集群
kind create cluster --name desktop --config kind-config.yaml

# 2. 构建镜像
docker build -f docker/backend/Dockerfile -t devops-be:latest .
docker build -f docker/frontend/Dockerfile -t devops-fe:latest ./k8s-web

# 3. 加载镜像到 Kind
kind load docker-image devops-be:latest --name desktop
kind load docker-image devops-fe:latest --name desktop

# 4. 部署后端
kubectl apply -k deploy/backend/

# 5. 部署前端
kubectl apply -k deploy/frontend/

# 6. 部署 Jenkins
kubectl apply -k deploy/jenkins/

# 7. 等待所有 Pod 就绪
kubectl -n k8soperation wait --for=condition=Ready pod --all --timeout=300s
kubectl -n devops wait --for=condition=Ready pod --all --timeout=600s

# 8. 验证
echo "平台前端: http://localhost:30081"
echo "Jenkins:  http://localhost:30080"
```

## 附录 B：更新部署

```bash
# ===== 更新后端 =====
docker build -f docker/backend/Dockerfile -t devops-be:latest .
kind load docker-image devops-be:latest --name desktop
kubectl -n k8soperation rollout restart deployment/k8soperation

# ===== 更新前端 =====
docker build -f docker/frontend/Dockerfile -t devops-fe:latest ./k8s-web
kind load docker-image devops-fe:latest --name desktop
kubectl -n k8soperation rollout restart deployment/k8soperation-web

# ===== 查看滚动更新状态 =====
kubectl -n k8soperation rollout status deployment/k8soperation
kubectl -n k8soperation rollout status deployment/k8soperation-web
```

## 附录 C：清理环境

```bash
# 删除所有 K8s 资源
kubectl delete -k deploy/backend/
kubectl delete -k deploy/frontend/
kubectl delete -k deploy/jenkins/

# 删除 Kind 集群
kind delete cluster --name desktop
```
