# K8sOperation 部署配置修改指南

> 本文档说明在 K8s 集群内部署时，前端、后端、Jenkins 各需要修改哪些配置文件。
> 适用于：**单节点 K8s 集群，MySQL/Redis/Jenkins 均在集群内部**

---

## 一、总览：需要修改的文件清单

| 组件 | 文件 | 修改内容 |
|------|------|---------|
| **后端** | `deploy/backend/secret.yaml` | 数据库、Redis、Jenkins 连接、通知配置 |
| **后端** | `deploy/backend/configmap.yaml` | CallbackURL、DefaultClusterID（按需） |
| **后端** | `deploy/backend/deployment.yaml` | 镜像版本 |
| **后端** | `deploy/backend/ingress.yaml` | 域名（可选） |
| **前端** | `deploy/frontend/deployment.yaml` | 镜像版本 |
| **前端** | `deploy/frontend/configmap.yaml` | 后端 upstream 地址（按需） |
| **前端** | `deploy/frontend/ingress.yaml` | 域名（可选） |
| **Jenkins** | `deploy/jenkins/secret.yaml` | 管理员密码、镜像仓库凭证、Git 凭证 |
| **Jenkins** | `deploy/jenkins/configmap.yaml` | 回调地址、SonarQube 地址 |

---

## 二、后端配置

### 2.1 `deploy/backend/secret.yaml`（必改）

这是**最核心**的配置文件，所有敏感连接信息都在这里：

```yaml
stringData:
  # ==================== MySQL 数据库 ====================
  DB_TYPE: "mysql"
  DB_USERNAME: "root"                    # ← 改为你的 MySQL 用户名
  DB_PASSWORD: "your-password"           # ← 改为你的 MySQL 密码
  DB_HOST: "mysql.database.svc.cluster.local"  # ← K8s 内部 Service DNS
  DB_PORT: "3306"                        # ← MySQL 端口（集群内通常 3306）
  DB_NAME: "k8s-platform"               # ← 数据库名
  DB_CHARSET: "utf8mb4"
  DB_PARSE_TIME: "true"
  DB_MAX_IDLE_CONNS: "10"
  DB_MAX_OPEN_CONNS: "100"
  DB_MAX_LIFE_SECONDS: "300"

  # ==================== Redis ====================
  CACHE_TYPE: "redis"
  CACHE_NAME: "sk_sid"
  REDIS_PASSWORD: "your-redis-password"  # ← 改为你的 Redis 密码（无密码留空）
  REDIS_ADDRESS: "redis.database.svc.cluster.local:6379"  # ← K8s 内部 Redis 地址
  CACHE_USERNAME: ""                     # Redis 6+ ACL 用户名（无则留空）
  CACHE_MAX_CONNECT: "10"
  CACHE_NETWORK: "tcp"
  CACHE_SECRET: "your-session-secret"    # ← Session 加密密钥（随机字符串）

  # ==================== Jenkins ====================
  JENKINS_URL: "http://jenkins.devops.svc.cluster.local:8080/"  # ← Jenkins Service DNS
  JENKINS_USERNAME: "ops-dev"            # ← Jenkins 管理员用户名
  JENKINS_API_TOKEN: "your-api-token"    # ← Jenkins API Token（Jenkins → 用户设置 → API Token）

  # ==================== 安全密钥 ====================
  JWT_SIGNING_KEY: "your-jwt-key-min-16-chars"   # ← JWT 签名密钥（≥16字符）
  HMAC_SECRET: "your-hmac-secret"                 # ← 回调签名密钥（与 Jenkins 一致）
  KUBECONFIG_ENCRYPT_KEY: "your-encrypt-key"      # ← KubeConfig 加密密钥

  # ==================== Jenkins 凭证 ID ====================
  GIT_CREDENTIAL_ID: "gitee-id"           # 对应 Jenkins 中的 Git 凭证 ID
  REGISTRY_CREDENTIAL_ID: "harbor-registry"  # 对应 Jenkins 中的镜像仓库凭证 ID
  HMAC_CREDENTIAL_ID: "hmac-secret"       # 对应 Jenkins 中的 HMAC 凭证 ID

  # ==================== 通知配置 ====================
  ENABLE_DINGTALK: "false"               # 钉钉通知开关
  ENABLE_FEISHU: "true"                  # 飞书通知开关
  FEISHU_WEBHOOK: "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxx"  # ← 飞书 Webhook
  FEISHU_SECRET: "your-feishu-secret"    # ← 飞书签名密钥
  DINGTALK_WEBHOOK: ""                   # 钉钉 Webhook（禁用时留空）

  # ==================== 前端地址 ====================
  PLATFORM_FRONTEND_URL: "http://k8sop.example.com"  # ← 前端公网地址（通知链接用）
```

**Service DNS 格式说明：**
```
<service-name>.<namespace>.svc.cluster.local:<port>
```

### 2.2 `deploy/backend/configmap.yaml`（按需改）

大部分通过 `${变量}` 引用 Secret，一般**不需要改**。以下情况需要修改：

| 配置项 | 位置 | 何时需要改 |
|--------|------|-----------|
| `CallbackURL` | Jenkins 段 | 后端 Service 名称/namespace 变更时 |
| `DefaultClusterID` | App 段 | 集群 ID 与默认值不同时 |
| `PlatformURL` | Jenkins 段 | 已通过 Secret 注入，通常不改 |

```yaml
Jenkins:
  CallbackURL: "http://k8soperation.k8soperation.svc:8080"  # ← 后端 Service 内部地址
```

### 2.3 `deploy/backend/deployment.yaml`（改镜像版本）

```yaml
containers:
  - name: k8soperation
    image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v13.6  # ← 改为最新版本
```

### 2.4 `deploy/backend/ingress.yaml`（可选，改域名）

```yaml
spec:
  rules:
    - host: k8soperation.example.com   # ← 替换为你的后端 API 域名
```

---

## 三、前端配置

### 3.1 `deploy/frontend/deployment.yaml`（改镜像版本）

```yaml
containers:
  - name: devops-fe
    image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:v13.6  # ← 改为最新版本
```

### 3.2 `deploy/frontend/configmap.yaml`（按需改）

前端 Nginx 通过 upstream 反向代理到后端，如果**后端 Service 名称或 namespace 变更**需要修改：

```yaml
data:
  default.conf: |
    upstream backend {
        server k8soperation.k8soperation.svc:8080;  # ← 后端 Service 地址
    }
```

> 默认值 `k8soperation.k8soperation.svc:8080` 对应后端 Service 在 `k8soperation` namespace 下，
> 如果你没改过 namespace，**不需要修改**。

### 3.3 `deploy/frontend/ingress.yaml`（可选，改域名）

```yaml
spec:
  rules:
    - host: k8sop.example.com           # ← 替换为你的前端域名
```

---

## 四、Jenkins 配置

### 4.1 `deploy/jenkins/secret.yaml`（必改）

```yaml
data:
  # 管理员密码（base64 编码）
  # echo -n "your-password" | base64
  admin-password: <base64编码的密码>

  # HMAC 签名密钥（与后端 secret.yaml 中的 HMAC_SECRET 一致）
  hmac-secret: <base64编码的HMAC密钥>

  # SonarQube Token（如不使用留默认值）
  sonarqube-token: <base64编码的token>

  # 镜像仓库凭证（阿里云 ACR）
  registry-username: <base64编码的用户名>
  registry-password: <base64编码的密码>

  # Git 凭证（Gitee/GitLab）
  gitee-username: <base64编码的用户名>
  gitee-password: <base64编码的密码或Token>
```

**生成 base64 命令：**
```bash
echo -n "your-value" | base64
```

### 4.2 `deploy/jenkins/configmap.yaml`（必改）

Jenkins CasC 配置中的**回调地址**需要改为集群内部后端地址：

```yaml
globalNodeProperties:
  - envVars:
      env:
        - key: "PLATFORM_CALLBACK_URL"
          value: "http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/pipeline/callback"
          # ↑ 改为后端 Service 的集群内部地址

        - key: "ARTIFACT_UPLOAD_URL"
          value: "http://k8soperation.k8soperation.svc:8080/api/v1/k8s/cicd/artifact/upload"
          # ↑ 同上
```

**当前值是 `host.docker.internal:38180`（本地开发地址），部署到集群必须改！**

如果使用 SonarQube，还需改：
```yaml
sonarGlobalConfiguration:
  installations:
    - name: "SonarQube"
      serverUrl: "http://sonarqube.devops.svc.cluster.local:9000"  # ← 改为实际地址
```

---

## 五、配置依赖关系图

```
┌─────────────────────────────────────────────────────────────┐
│                    deploy/backend/secret.yaml                 │
│  (MySQL/Redis/Jenkins 连接、密钥、通知 Webhook)               │
└──────────────────────────┬──────────────────────────────────┘
                           │ 环境变量注入 (secretKeyRef)
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                 deploy/backend/deployment.yaml                │
│  (容器定义、镜像版本、探针、资源限制)                          │
└──────────────────────────┬──────────────────────────────────┘
                           │ 环境变量替换 ${VAR}
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                deploy/backend/configmap.yaml                  │
│  (config.yaml 模板，引用 Secret 中的变量)                     │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│               deploy/frontend/configmap.yaml                  │
│  (Nginx 配置，upstream 指向后端 Service)                      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│               deploy/jenkins/configmap.yaml                   │
│  (JCasC 配置，回调地址指向后端 Service)                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 六、快速部署检查清单

部署前确认以下项已正确配置：

### 后端
- [ ] `secret.yaml` — DB_HOST 改为 MySQL Service DNS
- [ ] `secret.yaml` — DB_PORT 改为实际端口（通常 3306）
- [ ] `secret.yaml` — DB_PASSWORD 改为实际密码
- [ ] `secret.yaml` — REDIS_ADDRESS 改为 Redis Service DNS
- [ ] `secret.yaml` — REDIS_PASSWORD 改为实际密码
- [ ] `secret.yaml` — JENKINS_URL 确认 Jenkins Service DNS
- [ ] `secret.yaml` — JENKINS_API_TOKEN 填写实际 Token
- [ ] `secret.yaml` — HMAC_SECRET 与 Jenkins 一致
- [ ] `secret.yaml` — FEISHU_WEBHOOK 填写飞书通知地址
- [ ] `secret.yaml` — PLATFORM_FRONTEND_URL 改为前端公网地址
- [ ] `deployment.yaml` — 镜像版本正确
- [ ] `ingress.yaml` — 域名已替换（如需外部访问）

### 前端
- [ ] `deployment.yaml` — 镜像版本正确
- [ ] `configmap.yaml` — upstream 后端地址正确（默认不用改）
- [ ] `ingress.yaml` — 域名已替换（如需外部访问）

### Jenkins
- [ ] `secret.yaml` — admin-password 已修改
- [ ] `secret.yaml` — hmac-secret 与后端一致
- [ ] `secret.yaml` — registry-username/password 填写镜像仓库凭证
- [ ] `secret.yaml` — gitee-username/password 填写 Git 凭证
- [ ] `configmap.yaml` — PLATFORM_CALLBACK_URL 改为后端集群内地址
- [ ] `configmap.yaml` — ARTIFACT_UPLOAD_URL 改为后端集群内地址

---

## 七、部署命令

```bash
# 1. 创建 namespace
kubectl create namespace k8soperation
kubectl create namespace devops

# 2. 创建镜像拉取 Secret（阿里云 ACR）
kubectl create secret docker-registry aliyun-registry \
  --docker-server=registry.cn-hangzhou.aliyuncs.com \
  --docker-username=<你的用户名> \
  --docker-password=<你的密码> \
  -n k8soperation

# 3. 部署后端 + 前端
kubectl apply -k deploy/backend/
kubectl apply -k deploy/frontend/

# 4. 部署 Jenkins（独立 namespace）
kubectl apply -k deploy/jenkins/

# 5. 验证
kubectl get pods -n k8soperation
kubectl get pods -n devops
```

---

## 八、常见 Service DNS 地址参考

| 服务 | Namespace | Service DNS |
|------|-----------|-------------|
| MySQL | database | `mysql.database.svc.cluster.local:3306` |
| Redis | database | `redis.database.svc.cluster.local:6379` |
| Jenkins | devops | `jenkins.devops.svc.cluster.local:8080` |
| 后端 API | k8soperation | `k8soperation.k8soperation.svc.cluster.local:8080` |
| 前端 Web | k8soperation | `k8soperation-web.k8soperation.svc.cluster.local:80` |

> **注意：** 实际 DNS 取决于你 MySQL/Redis 部署时使用的 Service 名称和 Namespace。
> 查看命令：`kubectl get svc --all-namespaces | grep -E "mysql|redis"`
