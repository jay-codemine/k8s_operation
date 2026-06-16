# K8sOperation K8s 部署完整指南

> 版本: 2.3.0 | 更新日期: 2026-06-15

---

## 一、架构总览

```
┌─────────────────────────────────────────────────────────────┐
│                    K8s 集群                                   │
│                                                              │
│  ┌───────────────┐     ┌─────────────────────────────────┐  │
│  │ Ingress/NodePort │──▶│  Service (k8soperation:8080)   │  │
│  └───────────────┘     └──────────────┬──────────────────┘  │
│                                       │                      │
│                          ┌────────────▼─────────────┐       │
│                          │  Deployment (k8soperation)│       │
│                          │  ├─ ConfigMap (config)    │       │
│                          │  ├─ Secret (credentials)  │       │
│                          │  ├─ PVC (artifacts)       │       │
│                          │  └─ PVC (logs)            │       │
│                          └────────────┬─────────────┘       │
│                                       │                      │
│                    ┌──────────────────┼──────────────────┐  │
│                    │                  │                   │  │
│          ┌─────────▼──────┐  ┌───────▼───────┐          │  │
│          │ MySQL (PVC 10G)│  │ Redis (PVC 2G)│          │  │
│          └────────────────┘  └───────────────┘          │  │
│                                                          │  │
└─────────────────────────────────────────────────────────────┘
```

## 二、文件清单

| 文件 | 用途 | 必须 |
|------|------|------|
| `deploy/namespace.yaml` | 命名空间 `k8soperation` | ✅ |
| `deploy/secret.yaml` | 数据库密码、JWT密钥等敏感信息 | ✅ |
| `deploy/configmap.yaml` | 应用主配置 (config.yaml) | ✅ |
| `deploy/pvc.yaml` | 制品存储 (20Gi) + 日志存储 (5Gi) | ✅ |
| `deploy/service.yaml` | ClusterIP Service + ServiceAccount + RBAC | ✅ |
| `deploy/deployment.yaml` | 后端 Deployment（含健康检查、资源限制） | ✅ |
| `deploy/service-nodeport.yaml` | NodePort 暴露（端口 30080） | 按需 |
| `deploy/ingress.yaml` | Ingress 暴露（需 Ingress Controller） | 按需 |
| `deploy/middleware.yaml` | MySQL 8.0 + Redis 7 部署（K8s 内） | 按需 |
| `deploy/kustomization.yaml` | Kustomize 一键部署入口 | ✅ |
| `scripts/deploy-k8s.sh` | 一键部署脚本（自动化全流程） | 推荐 |
| `docs/sql/k8s_platform_full_init.sql` | 数据库初始化 SQL（50张表+种子数据） | ✅ |

## 三、需要持久化挂载的数据

### 3.1 必须挂载

| 组件 | 挂载路径 | PVC 名称 | 建议容量 | 说明 |
|------|----------|----------|----------|------|
| **MySQL 数据** | `/var/lib/mysql` | `mysql-data` | 10-50Gi | 全部业务数据、用户信息、集群配置 |
| **Redis 数据** | `/data` | `redis-data` | 2Gi | Session、缓存、AOF 持久化 |
| **CI/CD 制品** | `/app/storage/artifacts` | `k8soperation-artifacts` | 20Gi | 构建产物（JAR/二进制/tar.gz） |

### 3.2 建议挂载

| 组件 | 挂载路径 | PVC 名称 | 建议容量 | 说明 |
|------|----------|----------|----------|------|
| **应用日志** | `/app/storage/logs` | `k8soperation-logs` | 5Gi | 运行日志（也可用 stdout + 日志采集替代） |

### 3.3 ConfigMap 挂载（只读）

| 挂载内容 | 容器路径 | 来源 |
|----------|----------|------|
| 应用配置 | `/app/configs/config.yaml` | ConfigMap: `k8soperation-config` |
| Jenkins 模板 | `/app/configs/jenkins-templates/` | ConfigMap: `k8soperation-jenkins-templates`（可选） |

## 四、快速部署（一键脚本）

### 4.1 使用部署脚本

```bash
# 方式1: NodePort 暴露（默认）
bash scripts/deploy-k8s.sh

# 方式2: NodePort 自定义端口
EXPOSURE=nodeport NODE_PORT=31080 bash scripts/deploy-k8s.sh

# 方式3: Ingress 暴露（需集群已安装 Ingress Controller）
EXPOSURE=ingress DOMAIN=ops.mycompany.com bash scripts/deploy-k8s.sh

# 方式4: 使用外部数据库（不部署 MySQL/Redis）
DEPLOY_MIDDLEWARE=false bash scripts/deploy-k8s.sh
```

### 4.2 使用 Kustomize 一键部署

```bash
# 部署所有资源（不含中间件）
kubectl apply -k deploy/

# 查看生成的 YAML 预览
kubectl kustomize deploy/
```

## 五、手动分步部署

### 5.1 准备工作

```bash
# 确认集群连通
kubectl cluster-info

# 确认 StorageClass 可用（用于动态 PV 分配）
kubectl get storageclass
```

### 5.2 Step 1: 创建命名空间

```bash
kubectl apply -f deploy/namespace.yaml
```

### 5.3 Step 2: 配置 Secret（⚠️ 必须修改！）

编辑 `deploy/secret.yaml`，将占位符替换为实际的 Base64 编码值：

```bash
# 生成 base64 编码
echo -n "your-mysql-password" | base64
echo -n "your-redis-password" | base64
echo -n "your-jwt-signing-key-at-least-16-chars" | base64
echo -n "your-kubeconfig-encrypt-key-32-chars!" | base64
```

然后应用：

```bash
kubectl apply -f deploy/secret.yaml
```

### 5.4 Step 3: 部署中间件（二选一）

#### 方案 A: K8s 内部署 MySQL + Redis（开发/测试环境）

```bash
kubectl apply -f deploy/middleware.yaml

# 等待就绪
kubectl -n k8soperation wait --for=condition=available deployment/mysql --timeout=120s
kubectl -n k8soperation wait --for=condition=available deployment/redis --timeout=60s

# 导入初始化 SQL
MYSQL_POD=$(kubectl -n k8soperation get pod -l app.kubernetes.io/component=mysql -o jsonpath='{.items[0].metadata.name}')
kubectl -n k8soperation exec -i $MYSQL_POD -- mysql -u root -p"$(kubectl -n k8soperation get secret k8soperation-secret -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)" --default-character-set=utf8mb4 < docs/sql/k8s_platform_full_init.sql
```

#### 方案 B: 使用外部数据库（生产环境推荐）

修改 `deploy/configmap.yaml` 中的地址：

```yaml
Database:
  Host: your-rds-endpoint.mysql.rds.aliyuncs.com   # 外部 MySQL 地址
  Port: "3306"
  
Cache:
  Address: your-redis.redis.rds.aliyuncs.com:6379   # 外部 Redis 地址
```

### 5.5 Step 4: 部署应用

```bash
# ConfigMap + PVC + Service + Deployment
kubectl apply -f deploy/configmap.yaml
kubectl apply -f deploy/pvc.yaml
kubectl apply -f deploy/service.yaml
kubectl apply -f deploy/deployment.yaml

# 等待应用就绪
kubectl -n k8soperation wait --for=condition=available deployment/k8soperation --timeout=180s

# 验证
kubectl -n k8soperation get pods
kubectl -n k8soperation logs -f deployment/k8soperation --tail=50
```

### 5.6 Step 5: 暴露服务（三选一）

#### 方式 A: NodePort（无需 Ingress Controller）

```bash
kubectl apply -f deploy/service-nodeport.yaml

# 访问: http://<任意节点IP>:30080
# 查看节点 IP
kubectl get nodes -o wide
```

#### 方式 B: Ingress（推荐生产使用）

```bash
# 前提: 集群已安装 Ingress Controller (如 ingress-nginx)
# 修改 deploy/ingress.yaml 中的域名后:
kubectl apply -f deploy/ingress.yaml

# 访问: http://k8soperation.example.com
# 确保域名 DNS 解析到 Ingress Controller 的 External IP
```

#### 方式 C: Port-Forward（临时调试）

```bash
kubectl -n k8soperation port-forward svc/k8soperation 8080:8080

# 访问: http://localhost:8080
```

## 六、初始化脚本说明

### 6.1 数据库初始化

| 脚本 | 平台 | 用途 |
|------|------|------|
| `scripts/init-db.sh` | Linux/macOS | 独立初始化数据库（支持交互选择） |
| `scripts/init-db.ps1` | Windows | 同上（PowerShell 版） |
| `docs/sql/k8s_platform_full_init.sql` | 通用 | 完整 SQL（50张表 + 种子数据） |

```bash
# Linux 使用
DB_HOST=127.0.0.1 DB_PASS=123456 bash scripts/init-db.sh

# Windows PowerShell
$env:DB_PASS="123456"; powershell -File scripts/init-db.ps1
```

### 6.2 一键启动（本地开发）

| 脚本 | 平台 | 说明 |
|------|------|------|
| `scripts/quick-start.sh` | Linux/macOS | 环境检查→初始化DB→生成配置→编译→启动 |
| `scripts/quick-start.ps1` | Windows | 同上 |

### 6.3 K8s 部署

| 脚本 | 说明 |
|------|------|
| `scripts/deploy-k8s.sh` | 一键部署到 K8s（含中间件+应用+暴露） |

### 6.4 环境变量覆盖

所有脚本支持通过环境变量自定义配置：

```bash
# 数据库
DB_HOST=10.0.0.1 DB_PORT=3306 DB_USER=root DB_PASS=mypassword DB_NAME=k8s-platform

# Redis
REDIS_HOST=10.0.0.2 REDIS_PORT=6379 REDIS_PASS=myredispass

# K8s 部署
NAMESPACE=k8soperation EXPOSURE=nodeport NODE_PORT=30080
IMAGE=your-registry/k8soperation:v2.3.0
DEPLOY_MIDDLEWARE=true   # true=部署MySQL/Redis, false=使用外部
```

## 七、镜像构建

### 7.1 本地构建

```bash
# 编译二进制
go build -trimpath -ldflags="-s -w" -o bin/k8s_operation ./cmd/k8soperation

# 构建镜像
docker build -t k8soperation:latest .

# 推送到仓库
docker tag k8soperation:latest registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:latest
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:latest
```

### 7.2 使用 Makefile

```bash
make docker-build                    # 编译 + 打镜像
make docker-push REGISTRY=registry.cn-hangzhou.aliyuncs.com/k8s-gos  # 推送
```

### 7.3 CI/CD 自动构建（Jenkins）

项目已配置 Jenkinsfile，提交代码后自动触发：
1. `go build` 编译二进制
2. `docker build` 构建镜像
3. 推送到阿里云容器镜像仓库
4. 自动更新 K8s Deployment

## 八、验证部署

```bash
# 1. 检查 Pod 状态
kubectl -n k8soperation get pods -o wide

# 期望输出:
# NAME                            READY   STATUS    RESTARTS   AGE
# k8soperation-xxx                1/1     Running   0          2m
# mysql-xxx                       1/1     Running   0          3m
# redis-xxx                       1/1     Running   0          3m

# 2. 健康检查
curl http://<访问地址>/healthz/ready
# 期望: ok

# 3. 登录测试
curl -X POST http://<访问地址>/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
# 期望: {"code":0,"msg":"OK","data":{"token":"eyJ..."}}

# 4. 查看 PVC 状态
kubectl -n k8soperation get pvc

# 5. 查看日志
kubectl -n k8soperation logs -f deployment/k8soperation --tail=100
```

## 九、常见问题排查

### Q1: Pod CrashLoopBackOff

```bash
kubectl -n k8soperation describe pod <pod-name>
kubectl -n k8soperation logs <pod-name> --previous
```

常见原因：
- Secret 中密码未正确 Base64 编码
- MySQL 未就绪时应用已启动（startupProbe 会重试）
- PVC 未绑定（检查 StorageClass）

### Q2: MySQL 连接失败

```bash
# 检查 MySQL Pod 日志
kubectl -n k8soperation logs deployment/mysql

# 进入 MySQL Pod 手动测试
kubectl -n k8soperation exec -it deployment/mysql -- mysql -u root -p
```

### Q3: PVC Pending

```bash
kubectl -n k8soperation describe pvc
kubectl get storageclass

# 如果没有默认 StorageClass，需手动创建 PV 或指定 storageClassName
```

### Q4: Ingress 无法访问

```bash
# 检查 Ingress Controller 是否运行
kubectl -n ingress-nginx get pods

# 查看 Ingress 状态
kubectl -n k8soperation describe ingress k8soperation

# 确认域名解析
nslookup k8soperation.example.com
```

## 十、生产环境建议

| 项目 | 开发/测试 | 生产 |
|------|-----------|------|
| 数据库 | K8s 内 MySQL Pod | 外部 RDS (阿里云/腾讯云) |
| Redis | K8s 内 Redis Pod | 外部 Redis 集群 |
| 存储 | 默认 StorageClass | 高性能 SSD StorageClass |
| 副本数 | 1 | 2+ (需 RWX 存储) |
| Ingress | NodePort | Ingress + HTTPS (cert-manager) |
| 镜像 | :latest | 固定版本标签 (如 :v2.3.0) |
| Secret | 明文 YAML | External Secrets / Vault |
| 监控 | 可选 | Prometheus + Grafana |
| 日志 | PVC 存储 | EFK/Loki 日志采集 |
| 备份 | 无 | 定时 mysqldump / xtrabackup |

## 十一、卸载

```bash
# 完全卸载（包括所有数据！）
kubectl delete namespace k8soperation

# 仅卸载应用（保留数据库数据）
kubectl -n k8soperation delete deployment k8soperation
kubectl -n k8soperation delete svc k8soperation k8soperation-nodeport
kubectl -n k8soperation delete ingress k8soperation
```

---

## 附录: 默认账号信息

| 服务 | 用户名 | 默认密码 | 说明 |
|------|--------|----------|------|
| 平台管理后台 | admin | admin123 | 首次登录后建议修改 |
| MySQL | root | (Secret 中配置) | 对应 `DB_PASSWORD` |
| Redis | - | (Secret 中配置) | 对应 `REDIS_PASSWORD` |
