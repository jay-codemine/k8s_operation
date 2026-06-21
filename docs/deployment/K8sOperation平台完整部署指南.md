# K8sOperation 平台完整部署指南

> 从交叉编译 → 镜像构建 → K8s 集群部署全流程

---

## 目录

- [一、环境要求](#一环境要求)
- [二、项目结构概览](#二项目结构概览)
- [三、交叉编译](#三交叉编译)
- [四、Docker 镜像构建](#四docker-镜像构建)
- [五、推送镜像到仓库](#五推送镜像到仓库)
- [六、K8s 部署配置详解](#六k8s-部署配置详解)
- [七、修改 config.yaml](#七修改-configyaml)
- [八、修改 secret.yaml](#八修改-secretyaml)
- [九、一键部署](#九一键部署)
- [十、验证部署](#十验证部署)
- [十一、更新升级](#十一更新升级)
- [十二、常见问题排查](#十二常见问题排查)

---

## 一、环境要求

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| Go | 1.24+ | 交叉编译需要（多阶段构建可免） |
| Docker / nerdctl | 20.10+ | 镜像构建 |
| kubectl | 1.28+ | K8s 集群操作 |
| K8s 集群 | 1.28+ | 目标部署环境 |
| MySQL | 8.0+ | 数据库 |
| Redis | 6.0+ | 缓存 |

### 外部依赖说明

平台运行依赖：
- **MySQL**：存储流水线、集群、用户等业务数据
- **Redis**：Session 管理、构建状态缓存
- **Jenkins**：CI/CD 构建引擎（独立部署，参见 Jenkins 部署文档）

---

## 二、项目结构概览

```
k8s_operation/
├── cmd/k8soperation/main.go       # 程序入口
├── docker/
│   ├── backend/
│   │   ├── Dockerfile             # 多阶段构建（推荐，Docker 内编译）
│   │   └── Dockerfile.runtime     # 纯运行时（需先本地编译）
│   └── frontend/
│       └── Dockerfile             # 前端多阶段构建
├── deploy/
│   └── backend/                   # K8s 部署清单
│       ├── kustomization.yaml     # Kustomize 编排
│       ├── namespace.yaml
│       ├── secret.yaml            # 敏感配置
│       ├── configmap.yaml         # 应用配置
│       ├── pv.yaml / pvc.yaml     # 持久化存储
│       ├── service.yaml           # Service + RBAC
│       ├── deployment.yaml        # 主 Deployment
│       ├── ingress.yaml           # Ingress（可选）
│       └── service-nodeport.yaml  # NodePort（可选）
├── configs/
│   ├── config.yaml                # 本地开发配置
│   └── jenkins-templates/         # Jenkins 流水线模板
└── Makefile                       # 构建自动化
```

---

## 三、交叉编译

### 3.1 方式一：本地交叉编译（推荐开发环境）

在 macOS/Windows 上编译 Linux 二进制：

```bash
# Linux AMD64（绝大多数云服务器）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags="-s -w" -o bin/devops-be ./cmd/k8soperation

# Linux ARM64（Apple Silicon 对应的 Linux 架构）
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -trimpath -ldflags="-s -w" -o bin/devops-be ./cmd/k8soperation
```

**参数说明：**
| 参数 | 作用 |
|------|------|
| `CGO_ENABLED=0` | 禁用 CGO，生成静态链接二进制 |
| `GOOS=linux` | 目标操作系统 |
| `GOARCH=amd64/arm64` | 目标 CPU 架构 |
| `-trimpath` | 移除编译路径信息（安全） |
| `-ldflags="-s -w"` | 去除符号表和调试信息（减小体积） |

### 3.2 方式二：使用 Makefile

```bash
# 默认编译当前平台
make build

# 指定目标平台
GOOS=linux GOARCH=amd64 make build
```

### 3.3 方式三：Docker 多阶段构建（无需本地 Go 环境）

完全在 Docker 内完成编译，适合 CI/CD 或没有 Go 环境的机器：

```bash
docker build -f docker/backend/Dockerfile -t devops-be:latest .
```

---

## 四、Docker 镜像构建

### 4.1 多阶段构建（推荐，一步到位）

```bash
# 标准构建
docker build -f docker/backend/Dockerfile -t devops-be:latest .

# 指定平台（Apple Silicon 构建 amd64 镜像）
docker build -f docker/backend/Dockerfile \
  --platform linux/amd64 \
  -t devops-be:latest .

# 多架构构建（需 docker buildx）
docker buildx build -f docker/backend/Dockerfile \
  --platform linux/amd64,linux/arm64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v1.0.0 \
  --push .
```

### 4.2 纯运行时构建（Jenkins/CI 已编译好二进制）

```bash
# 前提：bin/devops-be 已通过交叉编译生成
docker build -f docker/backend/Dockerfile.runtime -t devops-be:latest .
```

### 4.3 前端镜像构建

```bash
docker build -f docker/frontend/Dockerfile -t devops-fe:latest ./k8s-web
```

### 4.4 镜像标签规范

```bash
# 格式：<仓库地址>/<镜像名>:<版本号>
registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.4
registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:v13.4
```

---

## 五、推送镜像到仓库

```bash
# 登录阿里云 ACR
docker login registry.cn-hangzhou.aliyuncs.com

# 打标签
docker tag devops-be:latest registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.4

# 推送
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.4
```

---

## 六、K8s 部署配置详解

### 6.1 部署架构

```
┌────────────────────────────────────────────────────────────────┐
│  K8s Cluster                                                    │
│                                                                  │
│  Namespace: k8soperation                                         │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Deployment: k8soperation                                │    │
│  │  ┌───────────────────────────────────────────────────┐   │    │
│  │  │ Pod                                                │   │    │
│  │  │  initContainer: fix-permissions (root, 修复权限)   │   │    │
│  │  │  container: k8soperation (UID=1000, port 8080)     │   │    │
│  │  │    ├── ConfigMap → /app/configs/config.yaml        │   │    │
│  │  │    ├── PVC artifacts → /app/storage/artifacts      │   │    │
│  │  │    └── PVC logs → /app/storage/logs                │   │    │
│  │  └───────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  Service: k8soperation (ClusterIP:8080)                          │
│  ServiceAccount + ClusterRole (操作 K8s API)                      │
│  Ingress / NodePort (外部访问)                                    │
└────────────────────────────────────────────────────────────────┘
```

### 6.2 各文件作用

| 文件 | 作用 |
|------|------|
| `namespace.yaml` | 创建 `k8soperation` 命名空间 |
| `secret.yaml` | 数据库密码、Redis 密码、JWT Key 等敏感配置 |
| `configmap.yaml` | 应用主配置 config.yaml（引用 Secret 中的环境变量） |
| `pv.yaml` | PersistentVolume（hostPath，开发环境用） |
| `pvc.yaml` | PersistentVolumeClaim（制品 20Gi + 日志 5Gi） |
| `service.yaml` | ClusterIP Service + ServiceAccount + RBAC |
| `deployment.yaml` | 主 Deployment 定义 |
| `ingress.yaml` | Ingress（可选，需 Ingress Controller） |
| `service-nodeport.yaml` | NodePort 暴露（可选，端口 30080） |

---

## 七、修改 config.yaml

配置文件以 ConfigMap 形式挂载，文件路径：`deploy/backend/configmap.yaml`

### 7.1 完整配置说明

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8soperation-config
  namespace: k8soperation
data:
  config.yaml: |
    # ============ 服务配置 ============
    Server:
      RunMode: release          # release/debug，生产环境必须 release
      Port: 8080                # 服务端口（与 Deployment containerPort 一致）
      ReadTimeout: 3600         # 读超时（秒），支持 WebSocket 长连接
      WriteTimeout: 3600
      IdleTimeout: 300
      ShutdownTimeout: 300      # 优雅退出超时

    # ============ 数据库（值从 Secret 环境变量注入） ============
    Database:
      DBType: "${DB_TYPE}"            # mysql / postgres
      Username: "${DB_USERNAME}"
      Password: "${DB_PASSWORD}"
      Host: "${DB_HOST}"              # K8s 内部：mysql.default.svc.cluster.local
      Port: "${DB_PORT}"              # MySQL: 3306
      DBName: "${DB_NAME}"            # 数据库名
      Charset: "${DB_CHARSET}"        # utf8mb4
      ParseTime: "${DB_PARSE_TIME}"   # true
      MaxIdleConns: "${DB_MAX_IDLE_CONNS}"    # 10
      MaxOpenConns: "${DB_MAX_OPEN_CONNS}"    # 100
      MaxLifeSeconds: "${DB_MAX_LIFE_SECONDS}" # 300s

    # ============ Redis 缓存（值从 Secret 环境变量注入） ============
    Cache:
      Type: "${CACHE_TYPE}"           # redis
      Name: "${CACHE_NAME}"           # 前缀名
      Address: "${REDIS_ADDRESS}"     # K8s 内部：redis.default.svc:6379
      Addresses: []                   # Cluster 模式用
      Username: "${CACHE_USERNAME}"
      Password: "${REDIS_PASSWORD}"
      MaxConnect: "${CACHE_MAX_CONNECT}"  # 10
      Network: "${CACHE_NETWORK}"         # tcp
      Secret: "${CACHE_SECRET}"

    # ============ 应用配置 ============
    App:
      LogLevel: info              # debug/info/warn/error
      TIMEZONE: "Asia/Shanghai"
      LogType: single             # single/daily
      LogFileName: storage/logs/app.log
      BusinessLogFileName: storage/logs/biz.log
      LogMaxSize: 50              # 单文件最大 MB
      LogMaxBackup: 5
      LogMaxAge: 30               # 保留天数
      LogCompress: true
      JWTMaxRefreshTime: 86400
      JWTSigningKey: "${JWT_SIGNING_KEY}"
      JWTExpireTime: 120000       # Token 过期时间（秒）
      AppName: "k8soperation"
      GlobalKubeConfigPath: ""    # K8s 内部署留空（自动 InCluster）
      DefaultClusterID: 3
      AutoInitK8s: true
      AllowEmptyStart: true       # 允许无集群连接时启动

    # ============ Jenkins 配置 ============
    Jenkins:
      URL: "${JENKINS_URL}"                   # http://jenkins.devops.svc.cluster.local:8080/
      Username: "${JENKINS_USERNAME}"         # Jenkins 用户名
      APIToken: "${JENKINS_API_TOKEN}"        # Jenkins API Token
      TriggerTimeout: 60
      CallbackURL: "http://k8soperation.k8soperation.svc:8080"  # ★ 重要：K8s 内部 Service 地址
      PlatformURL: "${PLATFORM_FRONTEND_URL}"                    # 前端地址（通知链接用）
      HMACSecret: "${HMAC_SECRET}"
      GitCredentialID: "${GIT_CREDENTIAL_ID}"
      RegistryCredentialID: "${REGISTRY_CREDENTIAL_ID}"
      HMACCredentialID: "${HMAC_CREDENTIAL_ID}"
      PollInterval: 15
      MaxBuildTime: 30
      EnableDingTalk: ${ENABLE_DINGTALK}
      DingTalkWebhook: "${DINGTALK_WEBHOOK}"
      EnableFeishu: ${ENABLE_FEISHU}
      FeishuWebhook: "${FEISHU_WEBHOOK}"
      FeishuSecret: "${FEISHU_SECRET}"

    # ============ 安全配置 ============
    Security:
      KubeConfigEncryptKey: "${KUBECONFIG_ENCRYPT_KEY}"   # 16/24/32 字节 AES 密钥
      PasswordBcryptCost: 10
      AutoEncryptLegacyData: true
```

### 7.2 关键配置项说明

| 配置项 | 生产环境建议 | 说明 |
|--------|-------------|------|
| `Server.RunMode` | `release` | 禁用 debug 日志和 Swagger |
| `Jenkins.CallbackURL` | `http://k8soperation.k8soperation.svc:8080` | Jenkins 构建完成回调地址，必须是 K8s 集群内可达的地址 |
| `App.GlobalKubeConfigPath` | 留空 `""` | K8s 内部署自动使用 InCluster Config |
| `App.LogLevel` | `info` | 生产环境不要用 debug |

---

## 八、修改 secret.yaml

文件路径：`deploy/backend/secret.yaml`

### 8.1 完整 Secret 配置

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-secret
  namespace: k8soperation
type: Opaque
stringData:
  # ==================== 数据库配置 ====================
  DB_TYPE: "mysql"
  DB_USERNAME: "root"                        # ← 修改为实际用户名
  DB_PASSWORD: "your-db-password"            # ← 修改为实际密码
  DB_HOST: "mysql.default.svc.cluster.local" # ← K8s 内部 MySQL Service 地址
  DB_PORT: "3306"
  DB_NAME: "k8s-platform"                    # ← 数据库名
  DB_CHARSET: "utf8mb4"
  DB_PARSE_TIME: "true"
  DB_MAX_IDLE_CONNS: "10"
  DB_MAX_OPEN_CONNS: "100"
  DB_MAX_LIFE_SECONDS: "300s"

  # ==================== Redis/Cache 配置 ====================
  CACHE_TYPE: "redis"
  CACHE_NAME: "sk_sid"
  REDIS_PASSWORD: "your-redis-password"      # ← 修改为实际密码
  REDIS_ADDRESS: "redis.default.svc:6379"    # ← K8s 内部 Redis Service 地址
  CACHE_USERNAME: ""
  CACHE_MAX_CONNECT: "10"
  CACHE_NETWORK: "tcp"
  CACHE_SECRET: "k8smana"                    # ← Session 加密密钥

  # ==================== JWT 签名密钥 ====================
  JWT_SIGNING_KEY: "your-32-chars-jwt-key!!" # ← 至少 16 字符

  # ==================== Jenkins 配置 ====================
  JENKINS_URL: "http://jenkins.devops.svc.cluster.local:8080/"  # ← Jenkins Service 地址
  JENKINS_USERNAME: "ops-dev"                # ← Jenkins 用户名
  JENKINS_API_TOKEN: "your-jenkins-api-token" # ← Jenkins → 用户 → 配置 → API Token

  # ==================== HMAC 签名密钥 ====================
  HMAC_SECRET: "your-hmac-secret-key"        # ← 与 Jenkins Secret 中的 hmac-secret 保持一致

  # ==================== KubeConfig 加密密钥 ====================
  KUBECONFIG_ENCRYPT_KEY: "your-16or32-char-key!"  # ← 16/24/32 字节 AES 密钥

  # ==================== Jenkins 凭证 ID ====================
  GIT_CREDENTIAL_ID: "gitee-id"              # ← 对应 Jenkins 中 Git 凭证的 ID
  REGISTRY_CREDENTIAL_ID: "harbor-registry"  # ← 对应 Jenkins 中镜像仓库凭证的 ID
  HMAC_CREDENTIAL_ID: "hmac-secret"          # ← 对应 Jenkins 中 HMAC 密钥凭证的 ID

  # ==================== 通知配置 ====================
  ENABLE_DINGTALK: "true"                    # ← 启用钉钉通知
  DINGTALK_WEBHOOK: "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
  ENABLE_FEISHU: "false"                     # ← 启用飞书通知
  FEISHU_WEBHOOK: ""
  FEISHU_SECRET: ""

  # ==================== 前端地址 ====================
  PLATFORM_FRONTEND_URL: "http://your-domain.com"  # ← 前端公网地址（通知链接用）
```

### 8.2 必须修改的配置项

| 配置项 | 说明 | 生成方式 |
|--------|------|----------|
| `DB_PASSWORD` | MySQL 密码 | 你的数据库密码 |
| `DB_HOST` | MySQL 地址 | K8s Service DNS 或外部 IP |
| `REDIS_PASSWORD` | Redis 密码 | 你的 Redis 密码 |
| `REDIS_ADDRESS` | Redis 地址 | K8s Service DNS 或外部 IP |
| `JWT_SIGNING_KEY` | JWT 签名密钥 | `openssl rand -base64 32` |
| `JENKINS_API_TOKEN` | Jenkins API Token | Jenkins UI → 用户设置 → API Token |
| `HMAC_SECRET` | 回调签名密钥 | `openssl rand -hex 16` |
| `KUBECONFIG_ENCRYPT_KEY` | AES 加密密钥 | `openssl rand -hex 16`（32字符） |

---

## 九、一键部署

### 9.1 使用 Kustomize 部署（推荐）

```bash
# 1. 修改 secret.yaml 中的敏感配置
vim deploy/backend/secret.yaml

# 2. 修改 configmap.yaml 中的应用配置（一般不需要改）
vim deploy/backend/configmap.yaml

# 3. 修改 deployment.yaml 中的镜像地址
#    将 image: k8soperation:local 替换为实际镜像
sed -i 's|image: k8soperation:local|image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.4|' deploy/backend/deployment.yaml

# 4. 一键部署
kubectl apply -k deploy/backend/

# 5. 查看部署状态
kubectl -n k8soperation get pods -w
```

### 9.2 部署顺序说明

Kustomize 按 `kustomization.yaml` 中的 resources 顺序部署：

```
namespace.yaml → secret.yaml → configmap.yaml → pv.yaml → pvc.yaml → service.yaml → deployment.yaml
```

### 9.3 暴露服务（选择一种）

**方式 A：NodePort（简单，适合开发/测试）**
```bash
kubectl apply -f deploy/backend/service-nodeport.yaml
# 访问：http://<任意节点IP>:30080
```

**方式 B：Ingress（生产推荐）**
```bash
# 修改域名
sed -i 's|k8soperation.example.com|your-domain.com|' deploy/backend/ingress.yaml
kubectl apply -f deploy/backend/ingress.yaml
```

---

## 十、验证部署

```bash
# 1. 检查 Pod 状态
kubectl -n k8soperation get pods
# 期望输出：k8soperation-xxx   1/1   Running

# 2. 检查日志
kubectl -n k8soperation logs -f deployment/k8soperation

# 3. 健康检查
kubectl -n k8soperation exec deploy/k8soperation -- wget -qO- http://127.0.0.1:8080/healthz/live
# 期望输出：{"status":"ok"}

kubectl -n k8soperation exec deploy/k8soperation -- wget -qO- http://127.0.0.1:8080/healthz/ready
# 期望输出：{"status":"ok","checks":{"database":"ok","redis":"ok"}}

# 4. 外部访问测试（NodePort 模式）
curl http://<节点IP>:30080/healthz/live
```

---

## 十一、更新升级

### 11.1 滚动更新镜像

```bash
# 方式 1：直接修改镜像
kubectl -n k8soperation set image deployment/k8soperation \
  k8soperation=registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.5

# 方式 2：修改 deployment.yaml 后 apply
kubectl apply -k deploy/backend/

# 查看滚动更新状态
kubectl -n k8soperation rollout status deployment/k8soperation
```

### 11.2 回滚

```bash
# 查看历史版本
kubectl -n k8soperation rollout history deployment/k8soperation

# 回滚到上一版本
kubectl -n k8soperation rollout undo deployment/k8soperation

# 回滚到指定版本
kubectl -n k8soperation rollout undo deployment/k8soperation --to-revision=2
```

### 11.3 修改配置（无需重建镜像）

```bash
# 修改 ConfigMap
kubectl -n k8soperation edit configmap k8soperation-config

# 修改 Secret
kubectl -n k8soperation edit secret k8soperation-secret

# 重启 Pod 使配置生效
kubectl -n k8soperation rollout restart deployment/k8soperation
```

---

## 十二、常见问题排查

### Q1: Pod 一直 Pending
```bash
kubectl -n k8soperation describe pod <pod-name>
# 检查 Events 中的错误原因（常见：PVC 未绑定、节点资源不足）
```

### Q2: Pod CrashLoopBackOff
```bash
kubectl -n k8soperation logs <pod-name> --previous
# 常见原因：数据库连接失败、Redis 连接失败、配置错误
```

### Q3: 镜像拉取失败
```bash
# 创建镜像拉取 Secret
kubectl -n k8soperation create secret docker-registry aliyun-registry \
  --docker-server=registry.cn-hangzhou.aliyuncs.com \
  --docker-username=<用户名> \
  --docker-password=<密码>
```

### Q4: Jenkins 回调不通
```bash
# 确认 CallbackURL 配置正确
# K8s 内部署：http://k8soperation.k8soperation.svc:8080
# 验证连通性
kubectl -n devops exec deploy/jenkins -- curl -s http://k8soperation.k8soperation.svc:8080/healthz/live
```

### Q5: 权限不足（RBAC）
```bash
# 检查 ServiceAccount 是否绑定 ClusterRole
kubectl get clusterrolebinding k8soperation -o yaml
```

---

## 附录 A：完整部署命令速查

```bash
# === 一站式部署（从源码到 K8s） ===

# Step 1: 构建镜像
docker build -f docker/backend/Dockerfile \
  --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.4 .

# Step 2: 推送镜像
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.4

# Step 3: 修改部署配置
vim deploy/backend/secret.yaml      # 填写真实密码/Token
vim deploy/backend/deployment.yaml  # 修改 image 字段

# Step 4: 部署到 K8s
kubectl apply -k deploy/backend/

# Step 5: 验证
kubectl -n k8soperation get pods -w
kubectl -n k8soperation logs -f deploy/k8soperation
```

## 附录 B：前端部署

```bash
# 构建前端镜像
docker build -f docker/frontend/Dockerfile \
  --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:v13.4 ./k8s-web

# 推送
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:v13.4

# 部署到 K8s
kubectl apply -k deploy/frontend/
```
