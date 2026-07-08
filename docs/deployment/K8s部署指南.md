# K8sOperation 平台 Kubernetes 部署指南

> 本文档详细说明如何将 K8sOperation 平台部署到 Kubernetes Deployment 中运行，包括前置条件、配置说明、部署步骤、验证方法和运维建议。

---

## 一、架构概览

```
                         ┌─────────────────────────────────────────────┐
                         │            Kubernetes Cluster                │
                         │                                             │
  ┌──────────┐           │   ┌──────────┐    ┌──────────────────────┐  │
  │ 浏览器    │───Ingress──▶ │  Service  │──▶│     Deployment       │  │
  │ (前端)   │           │   │  :8080   │    │  ┌────────────────┐  │  │
  └──────────┘           │   └──────────┘    │  │ k8soperation   │  │  │
                         │                    │  │                │  │  │
  ┌──────────┐           │                    │  │ /healthz/live  │  │  │
  │ Jenkins  │───callback──▶ (Service)  ──▶   │  │ /healthz/ready │  │  │
  └──────────┘           │                    │  └───────┬────────┘  │  │
                         │                    │          │           │  │
                         │   ┌───────────┐    │  ┌───────▼────────┐  │  │
                         │   │  ConfigMap │    │  │  PVC: artifacts│  │  │
                         │   │  + Secret  │    │  │  PVC: logs     │  │  │
                         │   └───────────┘    │  └────────────────┘  │  │
                         │                    └──────────────────────┘  │
                         │                                             │
                         │   ┌───────────┐    ┌───────────┐            │
                         │   │   MySQL    │    │   Redis    │           │
                         │   │  Service   │    │  Service   │           │
                         │   └───────────┘    └───────────┘            │
                         └─────────────────────────────────────────────┘
```

---

## 二、前置条件

| 条件 | 说明 |
|------|------|
| Kubernetes | v1.25+（推荐 v1.28+） |
| kubectl | 已配置集群访问权限 |
| StorageClass | 至少一个可用的 StorageClass（用于 PVC） |
| MySQL | 8.0+，已创建数据库 `k8s-platform` |
| Redis | 6.0+，可被集群内 Pod 访问 |
| 镜像仓库 | 已推送 k8soperation 镜像 |
| Ingress Controller | （可选）如需外部访问，需要 nginx-ingress 或 traefik |

---

## 三、部署文件清单

所有部署文件位于项目根目录 `deploy/` 下：

```
deploy/
├── kustomization.yaml      # Kustomize 入口（一键部署）
├── namespace.yaml           # 命名空间
├── secret.yaml              # 敏感配置（密码、Token、密钥）
├── configmap.yaml           # 主配置文件 config.yaml
├── pvc.yaml                 # 持久化存储（制品 + 日志）
├── service.yaml             # Service + ServiceAccount + RBAC
├── deployment.yaml          # 核心 Deployment
└── ingress.yaml             # Ingress 外部访问（可选）
```

---

## 四、配置说明

### 4.1 敏感配置（Secret）

编辑 `deploy/secret.yaml`，将占位值替换为实际的 base64 编码值：

```bash
# 生成 base64 编码值
echo -n "your-db-password" | base64
echo -n "your-redis-password" | base64
echo -n "your-jwt-signing-key" | base64
echo -n "http://jenkins:8080/" | base64
echo -n "admin" | base64
echo -n "your-jenkins-api-token" | base64
echo -n "your-hmac-secret-32-chars" | base64
echo -n "your-aes256-encrypt-key-32-chars" | base64
```

需要配置的 Secret Key：

| Key | 说明 | 示例原始值 |
|-----|------|-----------|
| `DB_PASSWORD` | MySQL 数据库密码 | `admin123` |
| `REDIS_PASSWORD` | Redis 密码 | `admin123` |
| `JWT_SIGNING_KEY` | JWT 签名密钥（16+ 位随机） | `eoNB0%bv5M7995F1` |
| `JENKINS_URL` | Jenkins 地址 | `http://jenkins:8080/` |
| `JENKINS_USERNAME` | Jenkins 用户名 | `admin` |
| `JENKINS_API_TOKEN` | Jenkins API Token | 从 Jenkins 获取 |
| `HMAC_SECRET` | HMAC 签名密钥（Jenkins 回调验证） | 32 位随机字符串 |
| `KUBECONFIG_ENCRYPT_KEY` | KubeConfig AES 加密密钥 | 32 位随机字符串 |
| `DINGTALK_WEBHOOK` | 钉钉通知 Webhook（可选） | 完整 URL |
| `PLATFORM_FRONTEND_URL` | 前端公网地址（钉钉通知链接用） | `https://k8sop.example.com` |

### 4.2 主配置（ConfigMap）

编辑 `deploy/configmap.yaml` 中的 `config.yaml`，主要调整：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `Database.Host` | `mysql.k8soperation.svc` | MySQL K8s Service 地址 |
| `Cache.Address` | `redis.k8soperation.svc:6379` | Redis K8s Service 地址 |
| `App.GlobalKubeConfigPath` | `""` (空) | 空字符串 = 自动使用 InCluster 配置 |
| `App.LogLevel` | `info` | 生产环境建议 info |
| `App.LogMaxSize` | `50` | 单文件 50MB |
| `Jenkins.CallbackURL` | `http://k8soperation.k8soperation.svc:8080` | Jenkins 回调地址（集群内部） |
| `ErrorCode.AllowOverride` | `false` | 生产环境禁止错误码覆盖 |

> **重要**：ConfigMap 中的 `${DB_PASSWORD}` 等占位符需要配合应用启动时的环境变量替换。如果应用不支持运行时变量替换，请直接在 ConfigMap 中填写明文值（不推荐），或者通过 init-container 渲染配置文件。

### 4.3 持久化存储（PVC）

| PVC 名称 | 用途 | 默认大小 | 访问模式 |
|----------|------|---------|---------|
| `k8soperation-artifacts` | CI/CD 制品文件 | 20Gi | ReadWriteOnce |
| `k8soperation-logs` | 应用日志 | 5Gi | ReadWriteOnce |

**多副本注意事项**：
- `ReadWriteOnce` (RWO) 仅支持单节点挂载，适合单副本部署
- 如需多副本，需将 PVC 改为 `ReadWriteMany` (RWX)，底层需要 NFS/CephFS/云厂商共享存储
- 或者将制品存储改为对象存储（MinIO/阿里云 OSS），需改造后端代码

### 4.4 RBAC 权限

`deploy/service.yaml` 包含 ServiceAccount + ClusterRole + ClusterRoleBinding：

- **ClusterRole**：平台需要跨命名空间管理 K8s 资源（Pod/Deployment/Service/ConfigMap/Secret/Node/PV/PVC/Ingress/StorageClass/RBAC/CRD/Metrics 等）
- 根据实际需求收敛权限，最小权限原则

### 4.5 Ingress（可选）

编辑 `deploy/ingress.yaml`：
- 替换 `host: k8soperation.example.com` 为实际域名
- 如需 HTTPS，取消 `tls` 和 `cert-manager.io` 注解的注释
- `proxy-body-size: 200m` 已设置，确保大制品文件上传不被 Nginx 拦截

---

## 五、部署步骤

### 5.1 构建并推送镜像

```bash
# 1. 编译二进制
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/k8s_operation ./cmd/k8soperation

# 2. 构建 Docker 镜像
docker build -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v2.0.0 .

# 3. 推送到镜像仓库
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v2.0.0
```

### 5.2 修改配置

```bash
# 1. 编辑 Secret（替换 base64 编码的敏感值）
vim deploy/secret.yaml

# 2. 编辑 ConfigMap（调整数据库/Redis/Jenkins 地址）
vim deploy/configmap.yaml

# 3. 编辑 Deployment（修改镜像地址和 tag）
vim deploy/deployment.yaml
# 修改 image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v2.0.0

# 4. 编辑 PVC（根据需求调整存储大小和 StorageClass）
vim deploy/pvc.yaml
```

### 5.3 一键部署

```bash
# 方式一：使用 Kustomize（推荐）
kubectl apply -k deploy/

# 方式二：按顺序手动部署
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/secret.yaml
kubectl apply -f deploy/configmap.yaml
kubectl apply -f deploy/pvc.yaml
kubectl apply -f deploy/service.yaml
kubectl apply -f deploy/deployment.yaml
# kubectl apply -f deploy/ingress.yaml    # 按需
```

### 5.4 验证部署

```bash
# 1. 检查 Pod 状态
kubectl get pods -n k8soperation
# 期望：STATUS=Running, READY=1/1

# 2. 查看启动日志
kubectl logs -n k8soperation deployment/k8soperation -f

# 3. 健康检查
kubectl exec -n k8soperation deployment/k8soperation -- wget -qO- http://127.0.0.1:8080/healthz/live
# 期望输出：ok

kubectl exec -n k8soperation deployment/k8soperation -- wget -qO- http://127.0.0.1:8080/healthz/ready
# 期望输出：ok

# 4. 检查 PVC 绑定状态
kubectl get pvc -n k8soperation
# 期望：STATUS=Bound

# 5. 端口转发测试（本地浏览器访问）
kubectl port-forward -n k8soperation svc/k8soperation 8080:8080
# 浏览器打开 http://localhost:8080
```

---

## 六、项目已具备的 K8s 就绪特性

| 特性 | 状态 | 代码位置 | 说明 |
|------|------|---------|------|
| 存活探针 | 已实现 | `internal/health/health.go` → `/healthz/live` | 仅检查进程存活 |
| 就绪探针 | 已实现 | `internal/health/health.go` → `/healthz/ready` | 检查 DB 连通性（300ms 超时） |
| 优雅退出 | 已实现 | `pkg/shutdown/shutdown.go` | 30s 超时 + 禁用 KeepAlive |
| 非 root 运行 | 已实现 | `Dockerfile` → `USER app` | 安全加固 |
| 环境变量配置 | 已支持 | `GIN_MODE` / `APP_CONFIG` / `K8S_CONFIG` | 12-Factor App |
| 日志轮转 | 已实现 | lumberjack | 自动压缩归档 |
| K8s 失败不阻塞 | 已实现 | `bootstrap.go` L61 | K8s 初始化失败仅 Warn |
| CICD Worker | 已实现 | `bootstrap.go` L74 | Redis Worker 失败不阻塞 |
| DB 自动迁移 | 已实现 | `initialize/db.go` | 启动时自动 AutoMigrate |

---

## 七、外部依赖说明

### 7.1 MySQL

| 属性 | 说明 |
|------|------|
| 版本 | 8.0+ |
| 数据库名 | `k8s-platform`（需提前创建） |
| 字符集 | UTF-8 |
| 连接方式 | ConfigMap 中 `Database.Host` 指定 Service 地址 |

**K8s 内部署 MySQL 示例**（也可使用云 RDS）：
```bash
# 使用 Helm 快速部署
helm install mysql bitnami/mysql \
  --namespace k8soperation \
  --set auth.rootPassword=your-password \
  --set auth.database=k8s-platform \
  --set primary.persistence.size=10Gi
```

### 7.2 Redis

| 属性 | 说明 |
|------|------|
| 版本 | 6.0+ |
| 用途 | Session 存储 + CICD Worker 队列 |
| 连接方式 | ConfigMap 中 `Cache.Address` 指定 Service 地址 |

```bash
# 使用 Helm 快速部署
helm install redis bitnami/redis \
  --namespace k8soperation \
  --set auth.password=your-password \
  --set architecture=standalone \
  --set master.persistence.size=2Gi
```

### 7.3 Jenkins

| 属性 | 说明 |
|------|------|
| 版本 | 2.400+ |
| 网络要求 | Jenkins 需要能访问到 k8soperation Service（回调地址） |
| 配置 | Secret 中 `JENKINS_URL` / `JENKINS_API_TOKEN` |
| 回调 | `Jenkins.CallbackURL` 设为 `http://k8soperation.k8soperation.svc:8080` |

---

## 八、运维指南

### 8.1 滚动更新

```bash
# 更新镜像版本
kubectl set image deployment/k8soperation k8soperation=registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v2.1.0 -n k8soperation

# 或修改 deployment.yaml 后 apply
kubectl apply -f deploy/deployment.yaml
```

更新策略：`maxUnavailable: 0, maxSurge: 1`（先启新 Pod，就绪后再下旧 Pod，零停机）。

### 8.2 查看日志

```bash
# 实时日志
kubectl logs -n k8soperation deployment/k8soperation -f

# 如果挂载了日志 PVC，也可以进容器查看
kubectl exec -n k8soperation deployment/k8soperation -- ls -la /app/storage/logs/
kubectl exec -n k8soperation deployment/k8soperation -- cat /app/storage/logs/biz.log
```

### 8.3 制品文件管理

```bash
# 查看制品存储使用量
kubectl exec -n k8soperation deployment/k8soperation -- du -sh /app/storage/artifacts/

# 清理旧制品（通过 API 删除，会同步删除文件）
curl -X DELETE http://k8soperation:8080/api/v1/k8s/cicd/artifact/batch-delete \
  -H "Authorization: Bearer <token>" \
  -d '{"ids": [1, 2, 3]}'
```

### 8.4 配置热更新

```bash
# 更新 ConfigMap 后需要重启 Pod（ConfigMap 不支持自动热加载）
kubectl rollout restart deployment/k8soperation -n k8soperation
```

### 8.5 扩缩容

```bash
# 注意：多副本需要 PVC 改为 ReadWriteMany
kubectl scale deployment/k8soperation --replicas=2 -n k8soperation
```

---

## 九、常见问题

### Q1: Pod 启动失败，日志显示 `db ping failed`
- 检查 MySQL Service 是否可达：`kubectl exec -n k8soperation deployment/k8soperation -- wget -qO- http://mysql:3306`
- 检查 Secret 中 `DB_PASSWORD` 是否正确（base64 编码）
- 检查 ConfigMap 中 `Database.Host` 是否正确

### Q2: Jenkins 回调失败
- 确认 Jenkins 能访问到 `http://k8soperation.k8soperation.svc:8080`
- 如果 Jenkins 在集群外，需要通过 Ingress/NodePort 暴露 Service
- 检查 `HMAC_SECRET` 是否与 Jenkins Credentials 中一致

### Q3: 制品上传超时
- 检查 Ingress `proxy-body-size` 注解是否够大（默认 200m）
- 检查 PVC 是否已满：`kubectl exec ... -- df -h /app/storage/artifacts/`
- 如果通过 Ingress 上传，确认 `proxy-read-timeout` 足够长

### Q4: K8s 集群管理功能不可用
- 检查 ServiceAccount 权限：`kubectl auth can-i get pods --as=system:serviceaccount:k8soperation:k8soperation`
- 如果管理的是外部集群，需要通过 Web 界面添加 kubeconfig

### Q5: 多副本部署制品存储冲突
- 将 PVC `accessModes` 改为 `ReadWriteMany`
- 底层需要支持 RWX 的存储（NFS/CephFS/云厂商共享存储）
- 或者将制品存储改为对象存储（MinIO/OSS），需改造后端 `ArtifactUpload` 代码
