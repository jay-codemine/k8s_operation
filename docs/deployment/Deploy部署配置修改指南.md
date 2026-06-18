# K8sOperation Deploy 部署配置修改指南

> 本文档详细说明 `deploy/` 目录下所有 YAML 文件，哪些需要修改、修改什么内容。
>
> **前提条件**：镜像已推送至阿里云仓库
> - `registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest`
> - `registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest`

---

## 一、目录总览

```
deploy/
├── kustomization.yaml              # 总入口（不需要改）
├── backend/
│   ├── kustomization.yaml          # 后端资源清单（不需要改）
│   ├── namespace.yaml              # 命名空间（不需要改）
│   ├── secret.yaml                 # ⚠️ 【必须修改】敏感信息
│   ├── configmap.yaml              # ⚠️ 【必须修改】数据库/Redis 地址
│   ├── deployment.yaml             # ⚠️ 【需要修改】imagePullSecrets
│   ├── service.yaml                # ✅ 不需要改
│   ├── pv.yaml                     # ⚡ 【按需修改】存储路径
│   ├── pvc.yaml                    # ✅ 不需要改
│   ├── service-nodeport.yaml       # 未启用（按需）
│   └── ingress.yaml                # 未启用（按需）
└── frontend/
    ├── kustomization.yaml          # 前端资源清单（不需要改）
    ├── configmap.yaml              # ✅ 不需要改（Nginx 配置已正确）
    ├── deployment.yaml             # ⚠️ 【需要修改】imagePullSecrets
    ├── service.yaml                # ✅ 不需要改
    ├── service-nodeport.yaml       # ⚡ 【按需修改】NodePort 端口号
    └── ingress.yaml                # 未启用（按需）
```

---

## 二、后端 - 必须修改的文件

### 2.1 `deploy/backend/secret.yaml` 【必须修改】

存放所有敏感信息（密码、Token、密钥），**部署前必须替换为真实值**。

```yaml
stringData:
  # ======================== 数据库 ========================
  DB_PASSWORD: "your-mysql-password"          # ← 改为你的 MySQL 密码

  # ======================== Redis ========================
  REDIS_PASSWORD: "your-redis-password"       # ← 改为你的 Redis 密码

  # ======================== JWT ========================
  JWT_SIGNING_KEY: "your-jwt-secret-key"      # ← 改为你的 JWT 签名密钥（建议 16+ 位随机字符串）

  # ======================== Jenkins ========================
  JENKINS_URL: "http://your-jenkins:8080/"    # ← 改为你的 Jenkins 地址
  JENKINS_USERNAME: "admin"                   # ← 改为你的 Jenkins 用户名
  JENKINS_API_TOKEN: "your-jenkins-token"     # ← 改为你的 Jenkins API Token

  # ======================== 安全 ========================
  HMAC_SECRET: "random-32-chars-string"       # ← 改为随机字符串（Jenkins 回调签名）
  KUBECONFIG_ENCRYPT_KEY: "must-be-32-chars-long-for-aes256"  # ← 改为 32 位字符串（AES-256）
  CACHE_SECRET: "your-cache-secret"           # ← 改为 Session 加密密钥

  # ======================== 可选 ========================
  DINGTALK_WEBHOOK: ""                        # ← 有钉钉通知需求则填入 Webhook URL
  PLATFORM_FRONTEND_URL: "http://NODE_IP:30081"  # ← 改为前端访问地址（钉钉消息跳转链接）
```

---

### 2.2 `deploy/backend/configmap.yaml` 【必须修改】

配置文件中需修改的具体字段：

#### Database 部分

```yaml
Database:
  DBType: mysql
  Username: dev_super                                           # ← 改为你的 MySQL 用户名
  Password: "${DB_PASSWORD}"                                    # 不需要改，运行时从 Secret 注入
  Host: mysql-for-test.global.svc.cluster.local                 # ← 【必改】改为你的外部 MySQL 地址
  Port: "3306"                                                  # ← 如果端口不同则修改
  DBName: k8s-platform                                          # ← 改为你的数据库名
```

**修改示例**（外部 MySQL）：
```yaml
  Host: 10.0.0.100           # 外部 MySQL IP
  Port: "3306"
  DBName: k8s_operation      # 你的数据库名
```

#### Cache (Redis) 部分

```yaml
Cache:
  Type: redis
  Name: sk_sid
  Address: redis.k8soperation.svc:6379                                        # 单节点地址（二选一）
  Addresses:                                                                   # Cluster 模式节点列表（二选一）
    - redis6-cluster-auth-redis-cluster.global.svc.cluster.local:6379         # ← 【必改】改为你的 Redis 地址
  Username: ""                                                                 # ← 如有用户名则填写
  Password: "${REDIS_PASSWORD}"                                                # 不需要改，运行时从 Secret 注入
```

**修改示例**（外部 Redis 单节点）：
```yaml
Cache:
  Address: 10.0.0.101:6379       # 外部 Redis 地址
  Addresses: []                   # 单节点模式：清空 Addresses 列表
```

**修改示例**（外部 Redis Cluster）：
```yaml
Cache:
  Address: ""                     # Cluster 模式：清空 Address
  Addresses:
    - 10.0.0.101:6379
    - 10.0.0.102:6379
    - 10.0.0.103:6379
```

#### Jenkins 部分（可选）

```yaml
Jenkins:
  URL: "${JENKINS_URL}"                                        # 从 Secret 注入
  Username: "${JENKINS_USERNAME}"                               # 从 Secret 注入
  APIToken: "${JENKINS_API_TOKEN}"                              # 从 Secret 注入
  CallbackURL: "http://k8soperation.k8soperation.svc:8080"     # ✅ 不需要改（K8s 内部回调）
  PlatformURL: "${PLATFORM_FRONTEND_URL}"                      # 从 Secret 注入
```

> Jenkins 相关值都在 `secret.yaml` 中配置，configmap 中不需要改。

---

### 2.3 `deploy/backend/deployment.yaml` 【需要修改】

需要修改 `imagePullSecrets` 名称：

**当前值**：
```yaml
imagePullSecrets:
  - name: harbor-secret          # ← 旧的 Harbor 仓库 Secret
```

**改为**（如果阿里云仓库是私有仓库）：
```yaml
imagePullSecrets:
  - name: aliyun-registry        # ← 改为阿里云仓库 Secret 名称
```

**如果阿里云仓库是公开仓库**，可以直接删除 `imagePullSecrets` 这两行。

> 创建 Secret 命令：
> ```bash
> kubectl create secret docker-registry aliyun-registry \
>   --docker-server=registry.cn-hangzhou.aliyuncs.com \
>   --docker-username=你的用户名 \
>   --docker-password="你的密码" \
>   -n k8soperation
> ```

---

### 2.4 `deploy/backend/pv.yaml` 【按需修改】

PV 使用 `hostPath` 存储（适合单节点），需确认路径在 Linux 节点上可写：

```yaml
# 制品存储
hostPath:
  path: /data/k8soperation/artifacts    # ← 确认此路径在节点上存在或可自动创建
  type: DirectoryOrCreate

# 日志存储
hostPath:
  path: /data/k8soperation/logs         # ← 确认此路径在节点上存在或可自动创建
  type: DirectoryOrCreate
```

> **一般不用改**，`DirectoryOrCreate` 会自动创建目录。如果你的节点数据盘挂载在其他路径（如 `/opt/data`），则需要修改。

---

## 三、前端 - 需要修改的文件

### 3.1 `deploy/frontend/deployment.yaml` 【需要修改】

同后端，修改 `imagePullSecrets`：

**当前值**：
```yaml
imagePullSecrets:
  - name: harbor-secret
```

**改为**：
```yaml
imagePullSecrets:
  - name: aliyun-registry        # 与后端一致
```

或者如果是公开仓库则删除此行。

---

### 3.2 `deploy/frontend/service-nodeport.yaml` 【按需修改】

```yaml
spec:
  type: NodePort
  ports:
    - name: http
      port: 80
      targetPort: http
      nodePort: 30081              # ← 可自定义端口号（30000-32767 范围内）
```

> 默认 30081 端口，部署后通过 `http://NODE_IP:30081` 访问前端。如端口冲突可改为其他值。

---

## 四、不需要修改的文件

| 文件 | 说明 |
|------|------|
| `deploy/kustomization.yaml` | 顶层入口，引用 backend/ 和 frontend/ |
| `deploy/backend/kustomization.yaml` | 后端资源清单 |
| `deploy/backend/namespace.yaml` | 创建 `k8soperation` 命名空间 |
| `deploy/backend/service.yaml` | ClusterIP Service + RBAC（已正确） |
| `deploy/backend/pvc.yaml` | PVC 配置（已正确） |
| `deploy/frontend/kustomization.yaml` | 前端资源清单 |
| `deploy/frontend/configmap.yaml` | Nginx 反向代理配置（upstream 已正确指向后端） |
| `deploy/frontend/service.yaml` | ClusterIP Service |

---

## 五、快速修改清单（Checklist）

```
部署前必做：
□ 1. 修改 deploy/backend/secret.yaml
     - DB_PASSWORD（MySQL 密码）
     - REDIS_PASSWORD（Redis 密码）
     - JWT_SIGNING_KEY（JWT 密钥）
     - JENKINS_URL / JENKINS_USERNAME / JENKINS_API_TOKEN
     - KUBECONFIG_ENCRYPT_KEY（改为 32 位随机字符串）

□ 2. 修改 deploy/backend/configmap.yaml
     - Database.Host（MySQL 地址）
     - Database.Username（MySQL 用户名）
     - Database.DBName（数据库名）
     - Cache.Address 或 Cache.Addresses（Redis 地址）

□ 3. 修改 deploy/backend/deployment.yaml
     - imagePullSecrets 名称改为 aliyun-registry（或删除）

□ 4. 修改 deploy/frontend/deployment.yaml
     - imagePullSecrets 名称改为 aliyun-registry（或删除）

可选：
□ 5. 修改 deploy/frontend/service-nodeport.yaml
     - nodePort 端口号（默认 30081）

□ 6. 修改 deploy/backend/pv.yaml
     - hostPath 路径（默认 /data/k8soperation/）
```

---

## 六、部署命令

```bash
# 0. 创建镜像拉取 Secret（私有仓库才需要）
kubectl create namespace k8soperation
kubectl create secret docker-registry aliyun-registry \
  --docker-server=registry.cn-hangzhou.aliyuncs.com \
  --docker-username=你的用户名 \
  --docker-password="你的密码" \
  -n k8soperation

# 1. 一键部署（前后端全部）
kubectl apply -k deploy/

# 2. 查看部署状态
kubectl get pods -n k8soperation -w

# 3. 查看服务暴露端口
kubectl get svc -n k8soperation

# 4. 访问前端
# http://NODE_IP:30081
```

---

## 七、常见问题

### Q: ImagePullBackOff 拉取镜像失败
```bash
# 检查 Secret 是否存在
kubectl get secret -n k8soperation
# 如果阿里云仓库是公开的，删掉 deployment 里的 imagePullSecrets 即可
```

### Q: Pod CrashLoopBackOff 启动失败
```bash
# 查看日志
kubectl logs -n k8soperation deploy/k8soperation --tail=50
# 常见原因：数据库连接失败、Redis 连接失败、配置文件错误
```

### Q: PVC Pending 状态
```bash
# 查看 PV 和 PVC 状态
kubectl get pv,pvc -n k8soperation
# hostPath 需要节点上目录可写，确认 /data/k8soperation/ 权限
```
