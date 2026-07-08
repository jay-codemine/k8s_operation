# K8sOperation 平台完整部署手册

> 包含：后端 API、前端 Web、MySQL 8.0、Redis 7  
> 部署方式：Docker 镜像构建 + Kubernetes 部署（同时提供 docker-compose 本地开发方案）

---

## 一、config.yaml 必须配置项说明

### 1.1 启动必须（缺失会 panic）

| 配置段 | 必填字段 | 说明 |
|--------|----------|------|
| **Server** | Port, RunMode | 服务端口和运行模式(debug/release) |
| **Database** | Username, Password, Host, Port, DBName | MySQL 连接信息 |
| **Cache** | Address, Password | Redis 连接地址和密码 |
| **App** | JWTSigningKey, AppName | JWT 签名密钥(至少16位)、应用名 |
| **PodLog** | TailDefault | Pod 日志默认尾行数 |
| **Pod** | eviction.default_grace_seconds | Pod 驱逐默认等待秒数 |
| **Node** | drain.max_grace_seconds | Node 排水最大等待秒数 |
| **ErrorCode** | AllowOverride | 错误码覆盖开关 |
| **ClusterClient** | TTL | 集群客户端缓存时长 |

### 1.2 可选配置（缺失不影响启动）

| 配置段 | 说明 | 默认行为 |
|--------|------|----------|
| **Jenkins** | CI/CD 流水线配置 | 未配置则 CI/CD 功能不可用 |
| **Security** | KubeConfig 加密密钥 | 未配置则数据不加密存储 |
| **Monitoring** | Prometheus 监控地址 | 未配置则监控功能不可用 |
| **AIAssistant** | AI 助手 API 配置 | 未配置则 AI 功能不可用 |
| **LDAP** | LDAP 统一认证 | 未配置则仅本地认证 |
| **PlatformSettings** | 平台默认设置 | 使用内置默认值 |

### 1.3 最小可运行配置示例

```yaml
Server:
  RunMode: debug
  Port: 8080
  ReadTimeout: 3600
  WriteTimeout: 3600
  IdleTimeout: 300
  ShutdownTimeout: 300

Database:
  DBType: mysql
  Username: root
  Password: "your_mysql_password"    # ← 必须填写
  Host: 127.0.0.1
  Port: "3306"
  DBName: k8s-platform
  Charset: utf8mb4
  ParseTime: true
  MaxIdleConns: 5
  MaxOpenConns: 100
  MaxLifeSeconds: 300

Cache:
  Type: redis
  Name: sk_sid
  Address: 127.0.0.1:6379            # ← 必须填写
  Addresses: []
  Username: ""
  Password: "your_redis_password"    # ← 必须填写
  MaxConnect: 10
  Network: tcp
  Secret: "k8smana"

App:
  LogLevel: debug
  TIMEZONE: "Asia/Shanghai"
  LogType: single
  LogFileName: storage/logs/app.log
  LogMaxSize: 1
  LogMaxBackup: 3
  LogMaxAge: 30
  LogCompress: true
  BusinessLogFileName: storage/logs/biz.log
  MirrorBusinessToSystem: false
  JWTMaxRefreshTime: 86400
  JWTSigningKey: "your-jwt-key-at-least-16ch"   # ← 必须填写
  JWTExpireTime: 120000
  AppName: "k8soperation"
  GlobalKubeConfigPath: ""           # 本地开发可填 configs/k8s.yaml
  DefaultClusterID: 0
  AutoInitK8s: true
  AllowEmptyStart: true              # ← 允许无K8s集群空启动

PodLog:
  EnableStreaming: false
  TailDefault: 500
  TailMax: 5000
  LimitBytes: 2097152
  Timestamps: false
  Previous: false

ErrorCode:
  AllowOverride: true

ClusterClient:
  TTL: 30m
  TTLJitter: 3m

Pod:
  eviction:
    default_grace_seconds: 30

Node:
  drain:
    max_grace_seconds: 300
    ignore_daemon_sets: true
    delete_empty_dir: false
```

---

## 二、数据库初始化

### 2.1 前置条件

- MySQL 8.0+ 已安装并运行
- 已创建 root 用户或具有 CREATE DATABASE 权限的用户

### 2.2 执行初始化 SQL

SQL 文件位置：`docs/sql/k8s_platform_full_init.sql`（包含 50 张表 + 种子数据）

```bash
# 方式1：Linux / macOS
mysql -u root -p'your_password' --default-character-set=utf8mb4 < docs/sql/k8s_platform_full_init.sql

# 方式2：使用项目自带脚本（交互式）
bash scripts/init-db.sh

# 方式3：环境变量指定连接信息
DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASS=your_password bash scripts/init-db.sh

# 方式4：MySQL 客户端内执行
mysql> source /path/to/docs/sql/k8s_platform_full_init.sql
```

### 2.3 验证初始化结果

```bash
mysql -u root -p -e "USE \`k8s-platform\`; SELECT COUNT(*) AS table_count FROM information_schema.TABLES WHERE TABLE_SCHEMA='k8s-platform';"
# 预期结果：约 50 张表
```

### 2.4 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 超级管理员 |

---

## 三、Docker 镜像构建

### 3.1 后端镜像构建

#### 方式 A：平台编译 + Docker 打包（推荐，镜像最小 ~20MB）

```bash
cd /path/to/k8s_operation-main

# 1. 编译二进制（Linux amd64）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/devops-be ./cmd/k8soperation

# 2. 构建镜像
docker build -t devops-be:latest .

# 或使用 Makefile
make docker-build
```

#### 方式 B：多阶段构建（无需本地 Go 环境）

```bash
docker build -f docs/dockerfile/Dockerfile.golang.prod -t devops-be:latest .

# 或使用 Makefile
make docker-build-standalone
```

#### 后端 Dockerfile（已存在于项目根目录）

```dockerfile
FROM alpine:3.20
RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /app
RUN mkdir -p /app/storage/logs /app/configs
COPY bin/devops-be /app/devops-be
RUN chmod +x /app/devops-be
RUN chown -R app:app /app
USER app
ENV GIN_MODE=release
ENV APP_CONFIG=/app/configs/config.yaml
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz/live || exit 1
ENTRYPOINT ["/app/devops-be"]
```

### 3.2 前端镜像构建

```bash
cd /path/to/k8s_operation-main

# 在项目根目录执行（注意 -f 指向 k8s-web/Dockerfile，context 为 ./k8s-web）
docker build -f k8s-web/Dockerfile -t devops-fe:latest ./k8s-web
```

#### 前端 Dockerfile 要点

- Stage 1：`node:22-alpine` 安装依赖 + `npm run build`
- Stage 2：`nginx:1.27-alpine` 托管静态文件 + API 反向代理
- 支持环境变量 `API_BACKEND_URL` 动态配置后端地址

### 3.3 镜像推送到仓库

```bash
# 标记并推送到 Harbor（替换为你的仓库地址）
REGISTRY=harbor.example.com/your-project

docker tag devops-be:latest $REGISTRY/devops-be:v1.0.0
docker push $REGISTRY/devops-be:v1.0.0

docker tag devops-fe:latest $REGISTRY/devops-fe:v1.0.0
docker push $REGISTRY/devops-fe:v1.0.0
```

---

## 四、Kubernetes 部署（生产级）

### 4.1 部署架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Namespace: k8soperation                    │
│                                                              │
│  ┌──────────┐    ┌──────────────┐    ┌──────────────────┐   │
│  │ Ingress  │───▶│ devops-fe    │    │ devops-be        │   │
│  │ (80/443) │    │ (Nginx:80)   │───▶│ (Go API:8080)    │   │
│  └──────────┘    └──────────────┘    └────────┬─────────┘   │
│                                               │              │
│                              ┌────────────────┼──────────┐   │
│                              │                │          │   │
│                        ┌─────▼─────┐   ┌─────▼─────┐        │
│                        │  MySQL    │   │  Redis    │        │
│                        │  (3306)   │   │  (6379)   │        │
│                        └───────────┘   └───────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### 4.2 一键部署（使用已有的 deploy/ 目录）

```bash
# 完整部署（前后端 + 中间件）
kubectl apply -k deploy/

# 或分步部署
kubectl apply -f deploy/backend/namespace.yaml      # 1. 创建命名空间
kubectl apply -f deploy/backend/secret.yaml         # 2. 敏感配置
kubectl apply -f deploy/backend/middleware.yaml     # 3. MySQL + Redis
kubectl apply -f deploy/backend/configmap.yaml      # 4. 后端配置
kubectl apply -f deploy/backend/pv.yaml             # 5. 持久化存储
kubectl apply -f deploy/backend/pvc.yaml            # 6. 存储声明
kubectl apply -f deploy/backend/service.yaml        # 7. Service + RBAC
kubectl apply -f deploy/backend/deployment.yaml     # 8. 后端 Deployment
kubectl apply -k deploy/frontend/                   # 9. 前端
```

### 4.3 关键配置修改

#### Step 1：修改 Secret（`deploy/backend/secret.yaml`）

```yaml
stringData:
  DB_PASSWORD: "your_real_mysql_password"         # ← 必改
  REDIS_PASSWORD: "your_real_redis_password"      # ← 必改
  JWT_SIGNING_KEY: "your-jwt-signing-key-32ch"    # ← 必改
  KUBECONFIG_ENCRYPT_KEY: "32位AES加密密钥"       # ← 建议修改
  # Jenkins（按需配置）
  JENKINS_URL: ""
  JENKINS_USERNAME: ""
  JENKINS_API_TOKEN: ""
  HMAC_SECRET: "changeme"
```

#### Step 2：修改 ConfigMap（`deploy/backend/configmap.yaml`）

重点修改：
- `Database.Host`：指向你的 MySQL 地址（K8s 内部用 `mysql.k8soperation.svc`）
- `Cache.Address`：指向你的 Redis 地址（K8s 内部用 `redis.k8soperation.svc:6379`）

#### Step 3：初始化数据库数据

MySQL Pod 启动后执行：

```bash
# 获取 MySQL Pod 名
MYSQL_POD=$(kubectl get pod -n k8soperation -l app.kubernetes.io/component=mysql -o jsonpath='{.items[0].metadata.name}')

# 将 SQL 文件拷入 Pod 并执行
kubectl cp docs/sql/k8s_platform_full_init.sql k8soperation/$MYSQL_POD:/tmp/init.sql
kubectl exec -n k8soperation $MYSQL_POD -- mysql -u root -p'your_password' --default-character-set=utf8mb4 -e "source /tmp/init.sql"
```

### 4.4 验证部署

```bash
# 检查所有 Pod 状态
kubectl get pods -n k8soperation

# 预期输出
# NAME                              READY   STATUS    RESTARTS   AGE
# k8soperation-xxx                  1/1     Running   0          1m
# k8soperation-web-xxx              1/1     Running   0          1m
# mysql-xxx                         1/1     Running   0          2m
# redis-xxx                         1/1     Running   0          2m

# 检查后端健康
kubectl exec -n k8soperation deploy/k8soperation -- wget -qO- http://127.0.0.1:8080/healthz/live

# 查看后端日志
kubectl logs -n k8soperation deploy/k8soperation -f
```

---

## 五、Docker Compose 本地开发方案

适用于本地开发/测试环境快速启动全套服务：

### 5.1 docker-compose.yaml

```yaml
version: "3.8"

services:
  # ==================== MySQL ====================
  mysql:
    image: mysql:8.0
    container_name: k8sop-mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: admin123
      MYSQL_DATABASE: k8s-platform
      TZ: Asia/Shanghai
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-authentication-plugin=mysql_native_password
      - --max-connections=500
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./docs/sql/k8s_platform_full_init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "127.0.0.1"]
      interval: 10s
      timeout: 5s
      retries: 10
      start_period: 30s

  # ==================== Redis ====================
  redis:
    image: redis:7-alpine
    container_name: k8sop-redis
    restart: always
    command: redis-server --requirepass admin123 --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "admin123", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

  # ==================== 后端 API ====================
  backend:
    image: devops-be:latest
    container_name: k8sop-backend
    restart: always
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      GIN_MODE: release
      APP_CONFIG: /app/configs/config.yaml
    ports:
      - "8080:8080"
    volumes:
      - ./configs/config-docker.yaml:/app/configs/config.yaml:ro
      - backend_logs:/app/storage/logs
      - backend_artifacts:/app/storage/artifacts
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1:8080/healthz/live"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 15s

  # ==================== 前端 Web ====================
  frontend:
    image: devops-fe:latest
    container_name: k8sop-frontend
    restart: always
    depends_on:
      backend:
        condition: service_healthy
    environment:
      API_BACKEND_URL: "http://backend:8080"
    ports:
      - "80:80"
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://127.0.0.1/health"]
      interval: 30s
      timeout: 3s
      retries: 3

volumes:
  mysql_data:
  redis_data:
  backend_logs:
  backend_artifacts:
```

### 5.2 Docker Compose 专用配置文件

创建 `configs/config-docker.yaml`：

```yaml
Server:
  RunMode: release
  Port: 8080
  ReadTimeout: 3600
  WriteTimeout: 3600
  IdleTimeout: 300
  ShutdownTimeout: 300

Database:
  DBType: mysql
  Username: root
  Password: "admin123"
  Host: mysql                    # Docker 网络内服务名
  Port: "3306"
  DBName: k8s-platform
  Charset: utf8mb4
  ParseTime: true
  MaxIdleConns: 5
  MaxOpenConns: 100
  MaxLifeSeconds: 300

Cache:
  Type: redis
  Name: sk_sid
  Address: redis:6379            # Docker 网络内服务名
  Addresses: []
  Username: ""
  Password: "admin123"
  MaxConnect: 10
  Network: tcp
  Secret: "k8smana"

App:
  LogLevel: info
  TIMEZONE: "Asia/Shanghai"
  LogType: single
  LogFileName: storage/logs/app.log
  LogMaxSize: 50
  LogMaxBackup: 5
  LogMaxAge: 30
  LogCompress: true
  BusinessLogFileName: storage/logs/biz.log
  MirrorBusinessToSystem: false
  JWTMaxRefreshTime: 86400
  JWTSigningKey: "docker-dev-jwt-key-32chars!!"
  JWTExpireTime: 120000
  AppName: "k8soperation"
  GlobalKubeConfigPath: ""
  DefaultClusterID: 0
  AutoInitK8s: true
  AllowEmptyStart: true

PodLog:
  EnableStreaming: false
  TailDefault: 500
  TailMax: 5000
  LimitBytes: 2097152
  Timestamps: false
  Previous: false

ErrorCode:
  AllowOverride: false

ClusterClient:
  TTL: 30m
  TTLJitter: 3m

Pod:
  eviction:
    default_grace_seconds: 30

Node:
  drain:
    max_grace_seconds: 300
    ignore_daemon_sets: true
    delete_empty_dir: false

Monitoring:
  Enabled: false
  PrometheusURL: ""
  QueryTimeout: 30

AIAssistant:
  Enabled: false
```

### 5.3 一键启动

```bash
# 构建镜像（首次执行）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/devops-be ./cmd/k8soperation
docker build -t devops-be:latest .
docker build -f k8s-web/Dockerfile -t devops-fe:latest ./k8s-web

# 启动全部服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止
docker-compose down

# 停止并清除数据
docker-compose down -v
```

---

## 六、常见问题排查

### 6.1 后端启动失败：`connect db failed`

```
原因：MySQL 未启动或连接信息错误
检查：
  1. MySQL 是否运行：mysql -u root -p -e "SELECT 1"
  2. config.yaml 中 Database.Host/Port/Password 是否正确
  3. 数据库 k8s-platform 是否已创建
```

### 6.2 后端启动失败：`init redis failed`

```
原因：Redis 未启动或密码错误
检查：
  1. Redis 是否运行：redis-cli -a your_password ping
  2. config.yaml 中 Cache.Address/Password 是否正确
```

### 6.3 后端启动失败：`address already in use`

```
原因：端口 8080 已被占用
解决：
  lsof -i :8080              # 查看占用进程
  kill <PID>                  # 杀掉旧进程
  # 或修改 config.yaml 中 Server.Port
```

### 6.4 K8s 集群连接失败

```
这不影响启动（AllowEmptyStart: true）
日志会提示"K8s 集群未初始化（空启动模式）"
解决：启动后通过 Web 界面添加集群即可
```

### 6.5 数据库表不完整

```
后端启动时会自动执行 AutoMigrate 补全缺失的表和字段
如果仍有问题，重新执行 SQL 初始化：
  mysql -u root -p < docs/sql/k8s_platform_full_init.sql
```

---

## 七、生产环境清单

| 检查项 | 说明 |
|--------|------|
| ✅ MySQL 密码 | 修改为强密码，不使用默认值 |
| ✅ Redis 密码 | 修改为强密码 |
| ✅ JWT 签名密钥 | 至少 32 位随机字符串 |
| ✅ KubeConfig 加密密钥 | 32 位 AES-256 密钥 |
| ✅ HMAC Secret | Jenkins 回调签名验证密钥 |
| ✅ RunMode | 设为 `release` |
| ✅ ErrorCode.AllowOverride | 设为 `false` |
| ✅ 镜像仓库 | 使用私有 Harbor/ACR，不用 latest 标签 |
| ✅ 资源限制 | 设置合理的 CPU/Memory limits |
| ✅ 持久化存储 | MySQL/Redis 使用 PVC 或外部 RDS |
| ✅ 网络策略 | 限制 Pod 间访问（NetworkPolicy） |
| ✅ Ingress TLS | 配置 HTTPS 证书 |
