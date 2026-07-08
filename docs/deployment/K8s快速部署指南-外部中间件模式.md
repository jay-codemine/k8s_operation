# K8sOperation 平台 K8s 快速部署指南（外部 MySQL + Redis）

> **适用场景**：MySQL 和 Redis 由外部集群/云服务提供，K8s 仅部署平台前后端服务。

---

## 一、架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Cluster                            │
│                                                                   │
│  ┌─────────────────┐       ┌─────────────────────┐              │
│  │  k8soperation   │──────▶│  k8soperation-web   │              │
│  │   (后端 Go)     │       │   (前端 Nginx)       │              │
│  │   Port: 8080    │       │   Port: 80           │              │
│  └────────┬────────┘       └─────────────────────┘              │
│           │                                                       │
└───────────┼───────────────────────────────────────────────────────┘
            │
   ┌────────┴─────────────────────────────┐
   │          外部中间件集群                │
   │                                       │
   │  ┌──────────────┐  ┌──────────────┐  │
   │  │  MySQL       │  │  Redis       │  │
   │  │  (外部)      │  │  (Cluster)   │  │
   │  └──────────────┘  └──────────────┘  │
   └───────────────────────────────────────┘
```

---

## 二、前置条件

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| Kubernetes | v1.24+ | 支持 kubectl + kustomize |
| kubectl | v1.24+ | 配置好集群 kubeconfig |
| Docker | 20.10+ | 用于构建镜像 |
| 外部 MySQL | 5.7+ / 8.0 | 已创建数据库 `k8s-platform` |
| 外部 Redis | 6.0+ | Cluster 模式或单节点均可 |
| Harbor（可选） | 2.x | 私有镜像仓库 |

---

## 三、镜像构建

### 3.1 后端镜像

```bash
# 在项目根目录执行
# Step 1: 编译 Go 二进制
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath \
  -ldflags="-s -w" -o bin/devops-be ./cmd/k8soperation

# Step 2: 构建 Docker 镜像
docker build -f docker/backend/Dockerfile --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest .

# Step 3: 推送到镜像仓库
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest
```

### 3.2 前端镜像

```bash
# 在项目根目录执行
# 构建前端镜像（多阶段构建，自动 npm install + build）
docker build -f docker/frontend/Dockerfile --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest ./k8s-web

# 推送
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest
```

> **提示**：如使用其他镜像仓库，请替换 `registry.cn-hangzhou.aliyuncs.com/k8s-gos/` 为实际地址，并同步修改 `deploy/backend/deployment.yaml` 和 `deploy/frontend/deployment.yaml` 中的 `image` 字段。

---

## 四、配置修改（重要）

### 4.1 修改外部 MySQL 连接

编辑 `deploy/backend/configmap.yaml`，修改 `Database` 部分：

```yaml
Database:
  DBType: mysql
  Username: your_mysql_user          # ← 替换
  Password: "${DB_PASSWORD}"         # 通过 Secret 注入
  Host: your-mysql-host.example.com  # ← 替换为外部 MySQL 地址
  Port: "3306"
  DBName: k8s-platform               # ← 确保数据库已创建
  Charset: utf8
  ParseTime: true
  MaxIdleConns: 10
  MaxOpenConns: 100
  MaxLifeSeconds: 300
```

### 4.2 修改外部 Redis 连接

编辑 `deploy/backend/configmap.yaml`，修改 `Cache` 部分：

**Redis Cluster 模式**（推荐）：
```yaml
Cache:
  Type: redis
  Name: sk_sid
  Address: ""                                  # Cluster 模式留空
  Addresses:                                   # ← 填写所有 Cluster 节点
    - redis-node1.example.com:6379
    - redis-node2.example.com:6379
    - redis-node3.example.com:6379
  Username: ""
  Password: "${REDIS_PASSWORD}"
  MaxConnect: 10
  Network: tcp
```

**Redis 单节点模式**：
```yaml
Cache:
  Type: redis
  Name: sk_sid
  Address: redis.example.com:6379              # ← 单节点地址
  Addresses: []                                # 置空则使用单节点模式
  Username: ""
  Password: "${REDIS_PASSWORD}"
  MaxConnect: 10
  Network: tcp
```

### 4.3 修改 Secret（敏感信息）

编辑 `deploy/backend/secret.yaml`：

```yaml
stringData:
  DB_PASSWORD: "your-mysql-password"           # ← 替换为实际密码
  REDIS_PASSWORD: "your-redis-password"        # ← 替换为实际密码
  JWT_SIGNING_KEY: "your-jwt-secret-key"       # ← 替换（建议随机 16+ 位）
  JENKINS_URL: "http://your-jenkins:8080/"     # ← Jenkins 地址
  JENKINS_USERNAME: "admin"                    # ← Jenkins 用户
  JENKINS_API_TOKEN: "your-jenkins-token"      # ← Jenkins API Token
  HMAC_SECRET: "your-hmac-secret"              # ← 回调签名密钥
  KUBECONFIG_ENCRYPT_KEY: "32-char-aes-key!!"  # ← AES-256 密钥（>=32位）
  CACHE_SECRET: "your-cache-secret"
  DINGTALK_WEBHOOK: ""                         # 可选：钉钉通知
  PLATFORM_FRONTEND_URL: "http://your-domain"  # 前端访问地址
```

### 4.4 修改镜像拉取凭证（如需私有仓库）

```bash
# 在集群中创建 Harbor 拉取凭证
kubectl create namespace k8soperation

kubectl create secret docker-registry harbor-secret \
  --namespace=k8soperation \
  --docker-server=registry.cn-hangzhou.aliyuncs.com \
  --docker-username=your-user \
  --docker-password=your-password
```

> 如使用公开镜像，可移除 deployment.yaml 中的 `imagePullSecrets` 部分。

---

## 五、一键部署

### 5.1 Kustomize 部署（推荐）

```bash
# 预览将要部署的资源
kubectl kustomize deploy/

# 一键部署前后端
kubectl apply -k deploy/
```

部署内容包含：
| 组件 | 资源类型 |
|------|----------|
| Namespace | `k8soperation` |
| Secret | `k8soperation-secret`（敏感配置） |
| ConfigMap | `k8soperation-config`（后端配置）、`k8soperation-web-nginx`（Nginx 配置） |
| PV/PVC | 制品存储(20Gi) + 日志存储(5Gi) |
| ServiceAccount | `k8soperation`（含 ClusterRole 权限） |
| Deployment | 后端(1 副本) + 前端(2 副本) |
| Service | 后端(ClusterIP:8080) + 前端(ClusterIP:80) |

### 5.2 分模块部署

```bash
# 仅后端
kubectl apply -k deploy/backend/

# 仅前端
kubectl apply -k deploy/frontend/
```

---

## 六、暴露服务（选择一种方式）

### 方式 A：NodePort（简单快速，适合测试）

编辑 `deploy/backend/kustomization.yaml`，取消注释 `service-nodeport.yaml`：

```yaml
resources:
  - namespace.yaml
  - secret.yaml
  - configmap.yaml
  - pv.yaml
  - pvc.yaml
  - service.yaml
  - deployment.yaml
  - service-nodeport.yaml         # ← 取消注释
```

编辑 `deploy/frontend/kustomization.yaml`，取消注释 `service-nodeport.yaml`：

```yaml
resources:
  - configmap.yaml
  - deployment.yaml
  - service.yaml
  - service-nodeport.yaml         # ← 取消注释
```

重新部署：
```bash
kubectl apply -k deploy/
```

访问地址：
- 前端：`http://<节点IP>:30081`
- 后端 API：`http://<节点IP>:30080`

### 方式 B：Ingress（生产推荐）

**前提**：集群已安装 Ingress Controller（如 nginx-ingress）。

1. 编辑 `deploy/frontend/ingress.yaml`，替换域名：
```yaml
rules:
  - host: k8sop.your-domain.com    # ← 替换为实际域名
```

2. 编辑 `deploy/backend/kustomization.yaml` 和 `deploy/frontend/kustomization.yaml`，取消注释 `ingress.yaml`。

3. 配置 DNS 解析，将域名指向集群 Ingress Controller 的外部 IP。

4. 重新部署：
```bash
kubectl apply -k deploy/
```

访问：`http://k8sop.your-domain.com`

---

## 七、验证部署

### 7.1 检查 Pod 状态

```bash
kubectl get pods -n k8soperation -w
```

预期输出：
```
NAME                               READY   STATUS    RESTARTS   AGE
k8soperation-xxxx-yyyy             1/1     Running   0          60s
k8soperation-web-xxxx-yyyy         1/1     Running   0          60s
k8soperation-web-xxxx-zzzz         1/1     Running   0          60s
```

### 7.2 检查服务状态

```bash
kubectl get svc -n k8soperation
```

### 7.3 健康检查

```bash
# 后端健康检查
kubectl exec -n k8soperation deploy/k8soperation -- \
  wget -qO- http://127.0.0.1:8080/healthz/live

# 前端健康检查
kubectl exec -n k8soperation deploy/k8soperation-web -- \
  wget -qO- http://127.0.0.1/health
```

### 7.4 查看日志

```bash
# 后端日志
kubectl logs -n k8soperation -l app.kubernetes.io/name=k8soperation -f

# 前端日志
kubectl logs -n k8soperation -l app.kubernetes.io/name=k8soperation-web -f
```

---

## 八、完整部署命令清单（TL;DR）

```bash
# ==================== 0. 构建镜像 ====================
# 后端
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/devops-be ./cmd/k8soperation
docker build -f docker/backend/Dockerfile.runtime --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest .
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:latest

# 前端
docker build -f docker/frontend/Dockerfile --platform linux/amd64 \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest ./k8s-web
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:latest

# ==================== 1. 创建命名空间 & 镜像凭证 ====================
kubectl create namespace k8soperation
kubectl create secret docker-registry harbor-secret \
  --namespace=k8soperation \
  --docker-server=registry.cn-hangzhou.aliyuncs.com \
  --docker-username=YOUR_USER \
  --docker-password=YOUR_PASS

# ==================== 2. 修改配置（见第四章） ====================
# vim deploy/backend/configmap.yaml   → 修改 MySQL/Redis 地址
# vim deploy/backend/secret.yaml      → 修改所有密码/密钥

# ==================== 3. 一键部署 ====================
kubectl apply -k deploy/

# ==================== 4. 验证 ====================
kubectl get pods -n k8soperation -w
kubectl get svc -n k8soperation
```

---

## 九、配置参考：目录结构

```
deploy/
├── kustomization.yaml          # 根 Kustomize（引用 backend/ + frontend/）
├── backend/
│   ├── kustomization.yaml      # 后端资源清单
│   ├── namespace.yaml          # Namespace 定义
│   ├── secret.yaml             # 敏感配置（密码/Token）
│   ├── configmap.yaml          # 后端主配置（config.yaml）
│   ├── pv.yaml                 # PersistentVolume（hostPath）
│   ├── pvc.yaml                # PersistentVolumeClaim
│   ├── service.yaml            # ClusterIP Service + RBAC
│   ├── deployment.yaml         # 后端 Deployment
│   ├── service-nodeport.yaml   # [可选] NodePort 暴露
│   ├── ingress.yaml            # [可选] Ingress 暴露
│   └── middleware.yaml         # [不需要] K8s 内部 MySQL/Redis
└── frontend/
    ├── kustomization.yaml      # 前端资源清单
    ├── configmap.yaml          # Nginx 配置
    ├── deployment.yaml         # 前端 Deployment
    ├── service.yaml            # ClusterIP Service
    ├── service-nodeport.yaml   # [可选] NodePort 暴露
    └── ingress.yaml            # [可选] Ingress 暴露
```

---

## 十、常见问题排查

### Q1: Pod 一直 CrashLoopBackOff

```bash
kubectl logs -n k8soperation deploy/k8soperation --previous
kubectl describe pod -n k8soperation -l app.kubernetes.io/name=k8soperation
```

**常见原因**：
- MySQL/Redis 地址不可达 → 检查 ConfigMap 配置是否正确
- 密码错误 → 检查 Secret 中的密码
- 数据库未创建 → 确保 `k8s-platform` 数据库已存在

### Q2: 前端无法访问后端 API

确认 Nginx ConfigMap 中的 upstream 地址正确：
```yaml
upstream backend {
    server k8soperation.k8soperation.svc:8080;  # K8s 内部 Service DNS
}
```

### Q3: 镜像拉取失败 (ImagePullBackOff)

```bash
kubectl describe pod -n k8soperation <pod-name>
```

- 检查 `harbor-secret` 是否创建正确
- 检查镜像名称和 tag 是否一致

### Q4: PV 绑定失败

如果集群不支持 hostPath，需要修改 `pv.yaml` 为实际的存储方案（如 NFS、云盘 CSI）：
```yaml
spec:
  # 替换 hostPath 为 NFS 示例
  nfs:
    server: nfs-server.example.com
    path: /data/k8soperation/artifacts
```

### Q5: 数据库初始化

首次部署时后端会自动通过 GORM AutoMigrate 创建表结构，确保：
- 数据库 `k8s-platform` 已创建
- 数据库用户有 DDL 权限（CREATE TABLE、ALTER TABLE）

---

## 十一、升级与回滚

### 滚动升级

```bash
# 更新后端镜像
kubectl set image deployment/k8soperation \
  k8soperation=registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-be:v2.0.0 \
  -n k8soperation

# 更新前端镜像
kubectl set image deployment/k8soperation-web \
  devops-fe=registry.cn-hangzhou.aliyuncs.com/k8s-gos/devops-fe:v2.0.0 \
  -n k8soperation
```

### 回滚

```bash
# 查看历史版本
kubectl rollout history deployment/k8soperation -n k8soperation

# 回滚到上一版本
kubectl rollout undo deployment/k8soperation -n k8soperation
kubectl rollout undo deployment/k8soperation-web -n k8soperation
```

---

## 十二、卸载

```bash
# 删除所有部署资源
kubectl delete -k deploy/

# 删除镜像凭证
kubectl delete secret harbor-secret -n k8soperation

# 删除 PV 上的数据（谨慎操作）
# kubectl delete pv k8soperation-artifacts k8soperation-logs
```
