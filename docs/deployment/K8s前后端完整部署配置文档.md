# K8sOperation 前后端完整 Kubernetes 部署配置文档

> 本文档涵盖后端（Go API）+ 前端（Vue3 + Nginx）+ MySQL + Redis 的完整 K8s 部署方案，包含所有 YAML 配置、PVC 持久化存储、数据库初始化、日志采集等内容。

---

## 一、整体架构

```
                    ┌─────────────────────────────────────────────────────────────────┐
                    │                    Kubernetes Cluster                            │
                    │                    namespace: k8soperation                       │
                    │                                                                  │
  ┌──────────┐     │   ┌─────────────┐     ┌────────────────────────────────────┐    │
  │  用户浏览器 │─────▶│  Ingress     │────▶│  前端 Service (port:80)             │    │
  └──────────┘     │   │  nginx-ingress│    │    ├── / → frontend Deployment     │    │
                    │   │              │    │    │     (Nginx + Vue3 dist)        │    │
                    │   │              │────▶│  后端 Service (port:8080)           │    │
                    │   │  /api/* → backend │    │    ├── /api → backend Deployment │    │
                    │   └─────────────┘     │    │     (Go Binary)               │    │
                    │                       └────────────────────────────────────┘    │
                    │                                                                  │
                    │   ┌─────────────────────────────────────────────────────────┐    │
                    │   │                    数据层                                │    │
                    │   │                                                         │    │
                    │   │   ┌──────────────┐   ┌──────────────┐                  │    │
                    │   │   │ MySQL 8.0    │   │ Redis 7.x    │                  │    │
                    │   │   │ StatefulSet  │   │ Deployment   │                  │    │
                    │   │   │ PVC: 50Gi    │   │              │                  │    │
                    │   │   └──────────────┘   └──────────────┘                  │    │
                    │   └─────────────────────────────────────────────────────────┘    │
                    │                                                                  │
                    │   ┌─────────────────────────────────────────────────────────┐    │
                    │   │                    存储层                                │    │
                    │   │                                                         │    │
                    │   │   PVC: k8soperation-artifacts  (20Gi) → 制品存储        │    │
                    │   │   PVC: k8soperation-logs       (5Gi)  → 应用日志        │    │
                    │   │   PVC: mysql-data              (50Gi) → 数据库文件      │    │
                    │   └─────────────────────────────────────────────────────────┘    │
                    └─────────────────────────────────────────────────────────────────┘
```

---

## 二、前置条件

| 条件 | 要求 |
|------|------|
| Kubernetes | v1.25+（推荐 v1.28+） |
| kubectl | 已配置集群访问权限 |
| StorageClass | 至少一个可用 SC（如 `local-path`、`nfs-client`、云厂商 SC） |
| Ingress Controller | nginx-ingress 或 traefik（提供外部访问） |
| 镜像仓库 | 可推送/拉取镜像（如阿里云 ACR、Harbor） |
| 域名 | 前端/后端域名（或统一域名按路径分流） |

---

## 三、完整部署文件清单

```
deploy/
├── namespace.yaml              # 命名空间
├── secret.yaml                 # 敏感配置（密码、Token）
├── configmap.yaml              # 后端主配置
├── configmap-nginx.yaml        # 前端 Nginx 配置        ★ 新增
├── pvc.yaml                    # 后端持久化存储（制品+日志）
├── pvc-mysql.yaml              # MySQL 数据持久化       ★ 新增
├── mysql.yaml                  # MySQL StatefulSet      ★ 新增
├── redis.yaml                  # Redis Deployment       ★ 新增
├── deployment.yaml             # 后端 Deployment
├── deployment-frontend.yaml    # 前端 Deployment        ★ 新增
├── service.yaml                # 后端 Service + RBAC
├── service-frontend.yaml       # 前端 Service           ★ 新增
├── ingress.yaml                # 统一 Ingress 入口
├── kustomization.yaml          # Kustomize 编排
└── init-job.yaml               # 数据库初始化 Job       ★ 新增
```

---

## 四、Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
    app.kubernetes.io/part-of: k8soperation-platform
```

---

## 五、Secret（敏感配置）

```yaml
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-secret
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
type: Opaque
data:
  # 生成方式：echo -n "your-value" | base64
  DB_PASSWORD: "Y2hhbmdlbWU="                          # MySQL 密码
  DB_ROOT_PASSWORD: "Y2hhbmdlbWU="                     # MySQL root 密码
  REDIS_PASSWORD: "Y2hhbmdlbWU="                       # Redis 密码
  JWT_SIGNING_KEY: "ZW9OQjAlYnY1TTc5OTVGMQ=="         # JWT 签名密钥（16+ 位）
  JENKINS_URL: "aHR0cDovL2plbmtpbnM6ODA4MC8="         # Jenkins 地址
  JENKINS_USERNAME: "YWRtaW4="                         # Jenkins 用户
  JENKINS_API_TOKEN: "Y2hhbmdlbWU="                    # Jenkins API Token
  HMAC_SECRET: "Y2hhbmdlbWU="                          # HMAC 签名密钥
  KUBECONFIG_ENCRYPT_KEY: "Y2hhbmdlbWU="               # AES-256 加密密钥（32位）
  DINGTALK_WEBHOOK: ""                                  # 钉钉 Webhook（可选）
  PLATFORM_FRONTEND_URL: "aHR0cDovL2xvY2FsaG9zdDo1MTcz" # 前端地址
```

**生成 base64 值**：
```bash
# Linux/Mac
echo -n "your-actual-password" | base64

# Windows PowerShell
[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("your-actual-password"))
```

---

## 六、MySQL 数据库部署

### 6.1 MySQL PVC

```yaml
# pvc-mysql.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-data
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: mysql
    app.kubernetes.io/component: database
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: ""           # 使用默认 StorageClass，或指定如 "nfs-client"
  resources:
    requests:
      storage: 50Gi              # 根据数据量调整
```

### 6.2 MySQL StatefulSet + Service

```yaml
# mysql.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mysql-initdb
  namespace: k8soperation
data:
  init.sql: |
    CREATE DATABASE IF NOT EXISTS `k8s-platform` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
    USE `k8s-platform`;
    -- 完整建表 SQL 请参考 docs/sql/k8s_platform_full_init.sql
    -- 此处放置核心表结构（篇幅原因省略，实际部署时需完整内容）
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mysql-config
  namespace: k8soperation
data:
  my.cnf: |
    [mysqld]
    # 基础配置
    character-set-server = utf8mb4
    collation-server = utf8mb4_unicode_ci
    default-time-zone = '+08:00'

    # 性能优化
    max_connections = 500
    innodb_buffer_pool_size = 1G
    innodb_log_file_size = 256M
    innodb_flush_log_at_trx_commit = 2
    slow_query_log = 1
    slow_query_log_file = /var/log/mysql/slow.log
    long_query_time = 2

    # 安全
    skip-name-resolve
    bind-address = 0.0.0.0

    [client]
    default-character-set = utf8mb4
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mysql
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: mysql
    app.kubernetes.io/component: database
spec:
  serviceName: mysql
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: mysql
  template:
    metadata:
      labels:
        app.kubernetes.io/name: mysql
        app.kubernetes.io/component: database
    spec:
      containers:
        - name: mysql
          image: mysql:8.0
          ports:
            - containerPort: 3306
              name: mysql
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: DB_ROOT_PASSWORD
            - name: MYSQL_DATABASE
              value: "k8s-platform"
            - name: TZ
              value: "Asia/Shanghai"
          resources:
            requests:
              cpu: 250m
              memory: 512Mi
            limits:
              cpu: "2"
              memory: 2Gi
          volumeMounts:
            - name: mysql-data
              mountPath: /var/lib/mysql
            - name: mysql-config
              mountPath: /etc/mysql/conf.d/my.cnf
              subPath: my.cnf
            - name: mysql-initdb
              mountPath: /docker-entrypoint-initdb.d
          livenessProbe:
            exec:
              command: ["mysqladmin", "ping", "-h", "localhost"]
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
          readinessProbe:
            exec:
              command: ["mysql", "-h", "localhost", "-uroot", "-p${MYSQL_ROOT_PASSWORD}", "-e", "SELECT 1"]
            initialDelaySeconds: 15
            periodSeconds: 10
      volumes:
        - name: mysql-data
          persistentVolumeClaim:
            claimName: mysql-data
        - name: mysql-config
          configMap:
            name: mysql-config
        - name: mysql-initdb
          configMap:
            name: mysql-initdb
---
apiVersion: v1
kind: Service
metadata:
  name: mysql
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: mysql
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: mysql
  ports:
    - name: mysql
      port: 3306
      targetPort: 3306
```

### 6.3 数据库初始化 Job（可选）

```yaml
# init-job.yaml - 用于导入完整 SQL 初始化数据
apiVersion: batch/v1
kind: Job
metadata:
  name: db-init
  namespace: k8soperation
spec:
  template:
    spec:
      restartPolicy: OnFailure
      containers:
        - name: db-init
          image: mysql:8.0
          command:
            - sh
            - -c
            - |
              until mysql -h mysql.k8soperation.svc -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT 1"; do
                echo "等待 MySQL 就绪..."
                sleep 5
              done
              mysql -h mysql.k8soperation.svc -uroot -p"$MYSQL_ROOT_PASSWORD" k8s-platform < /sql/k8s_platform_full_init.sql
              echo "数据库初始化完成"
          env:
            - name: MYSQL_ROOT_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: DB_ROOT_PASSWORD
          volumeMounts:
            - name: init-sql
              mountPath: /sql
      volumes:
        - name: init-sql
          configMap:
            name: mysql-initdb
```

---

## 七、Redis 部署

```yaml
# redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: redis
    app.kubernetes.io/component: cache
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: redis
  template:
    metadata:
      labels:
        app.kubernetes.io/name: redis
        app.kubernetes.io/component: cache
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          ports:
            - containerPort: 6379
              name: redis
          command:
            - redis-server
            - --requirepass
            - $(REDIS_PASSWORD)
            - --maxmemory
            - "256mb"
            - --maxmemory-policy
            - allkeys-lru
            - --appendonly
            - "yes"
            - --appendfsync
            - everysec
          env:
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: REDIS_PASSWORD
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 512Mi
          livenessProbe:
            exec:
              command: ["redis-cli", "ping"]
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            exec:
              command: ["redis-cli", "ping"]
            initialDelaySeconds: 5
            periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: redis
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: redis
  ports:
    - name: redis
      port: 6379
      targetPort: 6379
```

---

## 八、后端 PVC（制品 + 日志持久化）

```yaml
# pvc.yaml
# ============================================================
# PVC：CI/CD 制品存储（JAR/二进制/tar.gz 等构建产物）
# ============================================================
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: k8soperation-artifacts
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
    app.kubernetes.io/component: artifact-storage
spec:
  accessModes:
    - ReadWriteOnce              # 单副本 RWO；多副本需 RWX + NFS/CephFS
  storageClassName: ""           # 使用默认 StorageClass
  resources:
    requests:
      storage: 20Gi              # 每次构建约 10-50MB，按需调整
---
# ============================================================
# PVC：应用日志存储
# ============================================================
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: k8soperation-logs
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
    app.kubernetes.io/component: log-storage
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: ""
  resources:
    requests:
      storage: 5Gi               # 日志配置了轮转（50MB*5=250MB），5Gi 足够
```

### 日志存储策略说明

| 日志类型 | 文件路径 | 单文件上限 | 备份数 | 保留天数 | 压缩 |
|---------|----------|-----------|--------|---------|------|
| 系统日志 | `/app/storage/logs/app.log` | 50MB | 5 | 30天 | 是 |
| 业务日志 | `/app/storage/logs/biz.log` | 50MB | 5 | 30天 | 是 |

**替代方案：stdout + 日志采集**

如果使用 EFK/Loki 等集群级日志采集方案，可以：
1. 将 `App.LogFileName` 设为空（输出到 stdout）
2. 移除 `k8soperation-logs` PVC
3. 由 DaemonSet（如 Fluent Bit）采集容器日志

---

## 九、后端 ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8soperation-config
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
data:
  config.yaml: |
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
      Password: "${DB_PASSWORD}"
      Host: mysql.k8soperation.svc.cluster.local
      Port: "3306"
      DBName: k8s-platform
      Charset: utf8
      ParseTime: true
      MaxIdleConns: 10
      MaxOpenConns: 100
      MaxLifeSeconds: 300

    Cache:
      Type: redis
      Name: sk_sid
      Address: redis.k8soperation.svc.cluster.local:6379
      Username: ""
      Password: "${REDIS_PASSWORD}"
      MaxConnect: 10
      Network: tcp
      Secret: "k8smana"

    App:
      LogLevel: info
      TIMEZONE: "Asia/Shanghai"
      LogType: single
      LogFileName: storage/logs/app.log
      BusinessLogFileName: storage/logs/biz.log
      LogMaxSize: 50
      LogMaxBackup: 5
      LogMaxAge: 30
      LogCompress: true
      MirrorBusinessToSystem: false
      JWTMaxRefreshTime: 86400
      JWTSigningKey: "${JWT_SIGNING_KEY}"
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

    Jenkins:
      URL: "${JENKINS_URL}"
      Username: "${JENKINS_USERNAME}"
      APIToken: "${JENKINS_API_TOKEN}"
      TriggerTimeout: 60
      CallbackURL: "http://k8soperation.k8soperation.svc.cluster.local:8080"
      PlatformURL: "${PLATFORM_FRONTEND_URL}"
      HMACSecret: "${HMAC_SECRET}"
      PollInterval: 15
      MaxBuildTime: 30
      DingTalkWebhook: "${DINGTALK_WEBHOOK}"

    Security:
      KubeConfigEncryptKey: "${KUBECONFIG_ENCRYPT_KEY}"
      PasswordBcryptCost: 10
      AutoEncryptLegacyData: true

    AIAssistant:
      Enabled: false
      DefaultProvider: "qwen"
      SystemPrompt: "你是 K8s 管理平台的 AI 助手"
      ApprovalExpire: 30
      MaxHistoryRound: 20
```

---

## 十、前端 Nginx 配置 + Deployment

### 10.1 前端 Dockerfile（多阶段构建）

```dockerfile
# k8s-web/Dockerfile
# 阶段1：构建
FROM node:20-alpine AS builder
WORKDIR /app
RUN npm config set registry https://registry.npmmirror.com
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# 阶段2：运行
FROM nginx:1.25-alpine
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /app/dist /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/nginx.conf

RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

USER nginx
EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
```

### 10.2 前端 Nginx ConfigMap

```yaml
# configmap-nginx.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8soperation-nginx-config
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation-frontend
data:
  nginx.conf: |
    worker_processes auto;
    error_log /var/log/nginx/error.log warn;
    pid /var/run/nginx.pid;

    events {
        worker_connections 1024;
        multi_accept on;
    }

    http {
        include       /etc/nginx/mime.types;
        default_type  application/octet-stream;

        log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                        '$status $body_bytes_sent "$http_referer" '
                        '"$http_user_agent" "$http_x_forwarded_for"';

        access_log /var/log/nginx/access.log main;

        sendfile        on;
        tcp_nopush      on;
        tcp_nodelay     on;
        keepalive_timeout 65;

        # Gzip 压缩
        gzip on;
        gzip_vary on;
        gzip_min_length 1024;
        gzip_types text/plain text/css application/json application/javascript
                   text/xml application/xml application/xml+rss text/javascript
                   image/svg+xml;

        server {
            listen 80;
            server_name _;
            root /usr/share/nginx/html;
            index index.html;

            # 健康检查端点
            location /health {
                access_log off;
                return 200 'ok';
                add_header Content-Type text/plain;
            }

            # API 反向代理到后端 Service
            location /api/ {
                proxy_pass http://k8soperation.k8soperation.svc.cluster.local:8080;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;

                # WebSocket 支持（终端、日志流等）
                proxy_http_version 1.1;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection "upgrade";
                proxy_read_timeout 3600s;
                proxy_send_timeout 3600s;

                # 制品上传大文件限制
                client_max_body_size 200m;
            }

            # WebSocket 专用路径
            location /ws/ {
                proxy_pass http://k8soperation.k8soperation.svc.cluster.local:8080;
                proxy_http_version 1.1;
                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection "upgrade";
                proxy_set_header Host $host;
                proxy_read_timeout 3600s;
            }

            # 静态资源缓存
            location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
                expires 7d;
                add_header Cache-Control "public, immutable";
            }

            # Vue Router History 模式 - 所有路径回退到 index.html
            location / {
                try_files $uri $uri/ /index.html;
            }
        }
    }
```

### 10.3 前端 Deployment

```yaml
# deployment-frontend.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8soperation-frontend
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation-frontend
    app.kubernetes.io/component: frontend
spec:
  replicas: 2                    # 前端无状态，可多副本
  revisionHistoryLimit: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: k8soperation-frontend
  template:
    metadata:
      labels:
        app.kubernetes.io/name: k8soperation-frontend
        app.kubernetes.io/component: frontend
    spec:
      containers:
        - name: nginx
          image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-frontend:latest
          imagePullPolicy: Always
          ports:
            - name: http
              containerPort: 80
              protocol: TCP
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 500m
              memory: 256Mi
          volumeMounts:
            - name: nginx-config
              mountPath: /etc/nginx/nginx.conf
              subPath: nginx.conf
              readOnly: true
          livenessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 5
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 3
            periodSeconds: 10
      volumes:
        - name: nginx-config
          configMap:
            name: k8soperation-nginx-config
```

### 10.4 前端 Service

```yaml
# service-frontend.yaml
apiVersion: v1
kind: Service
metadata:
  name: k8soperation-frontend
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation-frontend
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: k8soperation-frontend
  ports:
    - name: http
      port: 80
      targetPort: http
      protocol: TCP
```

---

## 十一、后端 Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8soperation
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
    app.kubernetes.io/component: backend
spec:
  replicas: 1
  revisionHistoryLimit: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 0
      maxSurge: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: k8soperation
  template:
    metadata:
      labels:
        app.kubernetes.io/name: k8soperation
        app.kubernetes.io/component: backend
    spec:
      serviceAccountName: k8soperation
      terminationGracePeriodSeconds: 60
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000

      containers:
        - name: k8soperation
          image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:latest
          imagePullPolicy: Always
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP

          env:
            - name: GIN_MODE
              value: "release"
            - name: APP_CONFIG
              value: "/app/configs/config.yaml"
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: DB_PASSWORD
            - name: REDIS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: REDIS_PASSWORD
            - name: JWT_SIGNING_KEY
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: JWT_SIGNING_KEY
            - name: JENKINS_URL
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: JENKINS_URL
            - name: JENKINS_USERNAME
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: JENKINS_USERNAME
            - name: JENKINS_API_TOKEN
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: JENKINS_API_TOKEN
            - name: HMAC_SECRET
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: HMAC_SECRET
            - name: KUBECONFIG_ENCRYPT_KEY
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: KUBECONFIG_ENCRYPT_KEY
            - name: DINGTALK_WEBHOOK
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: DINGTALK_WEBHOOK
                  optional: true
            - name: PLATFORM_FRONTEND_URL
              valueFrom:
                secretKeyRef:
                  name: k8soperation-secret
                  key: PLATFORM_FRONTEND_URL
                  optional: true

          livenessProbe:
            httpGet:
              path: /healthz/live
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
            timeoutSeconds: 5
            failureThreshold: 3

          readinessProbe:
            httpGet:
              path: /healthz/ready
              port: http
            initialDelaySeconds: 15
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3

          startupProbe:
            httpGet:
              path: /healthz/live
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            failureThreshold: 30

          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: "1"
              memory: 512Mi

          volumeMounts:
            - name: config
              mountPath: /app/configs/config.yaml
              subPath: config.yaml
              readOnly: true
            - name: artifact-storage
              mountPath: /app/storage/artifacts
            - name: log-storage
              mountPath: /app/storage/logs

      volumes:
        - name: config
          configMap:
            name: k8soperation-config
        - name: artifact-storage
          persistentVolumeClaim:
            claimName: k8soperation-artifacts
        - name: log-storage
          persistentVolumeClaim:
            claimName: k8soperation-logs
```

---

## 十二、统一 Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k8soperation
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "200m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "60"
    # HTTPS (取消注释启用)
    # cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  rules:
    - host: k8sop.example.com              # ★ 替换为实际域名
      http:
        paths:
          # API 请求 → 后端
          - path: /api
            pathType: Prefix
            backend:
              service:
                name: k8soperation
                port:
                  name: http
          # WebSocket → 后端
          - path: /ws
            pathType: Prefix
            backend:
              service:
                name: k8soperation
                port:
                  name: http
          # 健康检查 → 后端
          - path: /healthz
            pathType: Prefix
            backend:
              service:
                name: k8soperation
                port:
                  name: http
          # 其他所有请求 → 前端
          - path: /
            pathType: Prefix
            backend:
              service:
                name: k8soperation-frontend
                port:
                  name: http
  # HTTPS TLS 配置（取消注释启用）
  # tls:
  #   - hosts:
  #       - k8sop.example.com
  #     secretName: k8soperation-tls
```

---

## 十三、Kustomization 编排

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: k8soperation

resources:
  - namespace.yaml
  - secret.yaml
  - configmap.yaml
  - configmap-nginx.yaml
  - pvc.yaml
  - pvc-mysql.yaml
  - mysql.yaml
  - redis.yaml
  - service.yaml
  - service-frontend.yaml
  - deployment.yaml
  - deployment-frontend.yaml
  - ingress.yaml

commonLabels:
  app.kubernetes.io/managed-by: kustomize
  app.kubernetes.io/part-of: k8soperation-platform
```

---

## 十四、一键部署流程

### 14.1 构建镜像

```bash
# ===== 后端镜像 =====
# 1. 编译 Go 二进制
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/k8s_operation ./cmd/k8soperation

# 2. 构建并推送
docker build -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v2.0.0 .
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v2.0.0

# ===== 前端镜像 =====
cd k8s-web

# 3. 构建并推送
docker build -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-frontend:v2.0.0 .
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-frontend:v2.0.0
```

### 14.2 部署到集群

```bash
# 方式一：Kustomize 一键部署（推荐）
kubectl apply -k deploy/

# 方式二：按顺序手动部署
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/secret.yaml
kubectl apply -f deploy/configmap.yaml
kubectl apply -f deploy/configmap-nginx.yaml
kubectl apply -f deploy/pvc.yaml
kubectl apply -f deploy/pvc-mysql.yaml
kubectl apply -f deploy/mysql.yaml
kubectl apply -f deploy/redis.yaml
kubectl apply -f deploy/service.yaml
kubectl apply -f deploy/service-frontend.yaml
kubectl apply -f deploy/deployment.yaml
kubectl apply -f deploy/deployment-frontend.yaml
kubectl apply -f deploy/ingress.yaml

# （可选）执行数据库初始化
kubectl apply -f deploy/init-job.yaml
```

### 14.3 验证部署

```bash
# 1. 检查所有 Pod 状态
kubectl get pods -n k8soperation -o wide
# 期望输出：
# NAME                                      READY   STATUS    RESTARTS   AGE
# k8soperation-xxx                          1/1     Running   0          2m
# k8soperation-frontend-xxx                 1/1     Running   0          2m
# mysql-0                                   1/1     Running   0          3m
# redis-xxx                                 1/1     Running   0          3m

# 2. 检查 Service
kubectl get svc -n k8soperation
# 期望输出：
# NAME                      TYPE        CLUSTER-IP      PORT(S)
# k8soperation              ClusterIP   10.96.x.x       8080/TCP
# k8soperation-frontend     ClusterIP   10.96.x.x       80/TCP
# mysql                     ClusterIP   10.96.x.x       3306/TCP
# redis                     ClusterIP   10.96.x.x       6379/TCP

# 3. 检查 PVC 绑定
kubectl get pvc -n k8soperation
# 期望：所有 PVC STATUS=Bound

# 4. 检查 Ingress
kubectl get ingress -n k8soperation

# 5. 验证后端健康
kubectl exec -it deploy/k8soperation -n k8soperation -- wget -qO- http://localhost:8080/healthz/live

# 6. 验证前端健康
kubectl exec -it deploy/k8soperation-frontend -n k8soperation -- wget -qO- http://localhost/health

# 7. 查看后端日志
kubectl logs -f deploy/k8soperation -n k8soperation

# 8. 端口转发本地测试
kubectl port-forward svc/k8soperation-frontend -n k8soperation 8000:80
# 浏览器访问 http://localhost:8000
```

---

## 十五、PVC 与存储详细说明

### 15.1 存储一览表

| PVC 名称 | 用途 | 容量 | 访问模式 | 挂载路径 | 说明 |
|----------|------|------|---------|---------|------|
| `k8soperation-artifacts` | CI/CD 制品 | 20Gi | RWO | `/app/storage/artifacts` | 构建产物（jar/bin/tar.gz） |
| `k8soperation-logs` | 应用日志 | 5Gi | RWO | `/app/storage/logs` | app.log + biz.log（轮转） |
| `mysql-data` | MySQL 数据 | 50Gi | RWO | `/var/lib/mysql` | 数据库文件 |

### 15.2 StorageClass 选型

| 环境 | 推荐 StorageClass | 说明 |
|------|-------------------|------|
| 本地开发（minikube/kind） | `standard` / `local-path` | 自带默认 SC |
| 私有集群（自建） | `nfs-client` / `ceph-rbd` | 需预装 Provisioner |
| 阿里云 ACK | `alicloud-disk-ssd` | SSD 云盘 |
| 腾讯云 TKE | `cbs-ssd` | SSD 云硬盘 |
| AWS EKS | `gp3` | EBS gp3 卷 |

### 15.3 多副本存储方案

如需后端多副本运行：

```yaml
# 方案一：改用 RWX 存储（NFS/CephFS）
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: "nfs-client"

# 方案二：制品存储改用对象存储（推荐）
# 将制品上传至 MinIO/OSS，后端代码改造对接 S3 API
# 优点：无限扩容、天然支持多副本
```

---

## 十六、日志存储方案

### 16.1 方案对比

| 方案 | 适用场景 | 优点 | 缺点 |
|------|---------|------|------|
| PVC 文件日志 | 单节点/小规模 | 简单、无需额外组件 | 不便于搜索、不支持多副本 |
| stdout + EFK | 中大规模集群 | 统一采集、可搜索 | 需部署 Elasticsearch + Fluentd |
| stdout + Loki | 轻量级集群 | 资源占用低、Grafana 集成 | 功能比 ES 少 |

### 16.2 方案一：PVC 文件日志（当前默认）

应用配置（ConfigMap 中）：
```yaml
App:
  LogFileName: storage/logs/app.log       # 系统日志
  BusinessLogFileName: storage/logs/biz.log  # 业务日志
  LogMaxSize: 50       # 单文件 50MB
  LogMaxBackup: 5      # 保留 5 个备份
  LogMaxAge: 30        # 保留 30 天
  LogCompress: true    # 压缩旧日志
```

### 16.3 方案二：stdout + Loki（推荐生产环境）

1. 修改 ConfigMap 中日志配置：
```yaml
App:
  LogFileName: ""                # 空 = 输出到 stdout
  BusinessLogFileName: ""        # 空 = 输出到 stdout
```

2. 部署 Grafana Loki Stack：
```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set grafana.enabled=true \
  --set promtail.enabled=true
```

3. 移除 `k8soperation-logs` PVC 和对应的 volumeMount。

---

## 十七、数据库备份策略

### 17.1 CronJob 自动备份

```yaml
# backup-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: mysql-backup
  namespace: k8soperation
spec:
  schedule: "0 2 * * *"          # 每天凌晨 2 点
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: backup
              image: mysql:8.0
              command:
                - sh
                - -c
                - |
                  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
                  mysqldump -h mysql.k8soperation.svc -uroot -p"$MYSQL_ROOT_PASSWORD" \
                    --single-transaction --quick --lock-tables=false \
                    k8s-platform > /backup/k8s-platform_${TIMESTAMP}.sql
                  # 清理 7 天前的备份
                  find /backup -name "*.sql" -mtime +7 -delete
                  echo "备份完成: k8s-platform_${TIMESTAMP}.sql"
              env:
                - name: MYSQL_ROOT_PASSWORD
                  valueFrom:
                    secretKeyRef:
                      name: k8soperation-secret
                      key: DB_ROOT_PASSWORD
              volumeMounts:
                - name: backup-storage
                  mountPath: /backup
          volumes:
            - name: backup-storage
              persistentVolumeClaim:
                claimName: mysql-backup        # 需创建对应 PVC
```

---

## 十八、运维命令速查

```bash
# ===== 查看状态 =====
kubectl get all -n k8soperation
kubectl top pods -n k8soperation

# ===== 扩缩容 =====
kubectl scale deployment k8soperation-frontend -n k8soperation --replicas=3

# ===== 滚动更新 =====
kubectl set image deployment/k8soperation \
  k8soperation=registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v2.1.0 \
  -n k8soperation

kubectl set image deployment/k8soperation-frontend \
  nginx=registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-frontend:v2.1.0 \
  -n k8soperation

# ===== 回滚 =====
kubectl rollout undo deployment/k8soperation -n k8soperation
kubectl rollout history deployment/k8soperation -n k8soperation

# ===== 查看日志 =====
kubectl logs -f deploy/k8soperation -n k8soperation --tail=200
kubectl logs -f deploy/k8soperation-frontend -n k8soperation

# ===== 进入容器调试 =====
kubectl exec -it deploy/k8soperation -n k8soperation -- sh
kubectl exec -it deploy/mysql-0 -n k8soperation -- mysql -uroot -p

# ===== 检查 PVC 使用量 =====
kubectl exec -it deploy/k8soperation -n k8soperation -- df -h /app/storage/artifacts
kubectl exec -it deploy/k8soperation -n k8soperation -- du -sh /app/storage/logs/*

# ===== 重启服务 =====
kubectl rollout restart deployment/k8soperation -n k8soperation
kubectl rollout restart deployment/k8soperation-frontend -n k8soperation

# ===== 数据库连接测试 =====
kubectl run mysql-client --rm -it --image=mysql:8.0 -n k8soperation -- \
  mysql -h mysql.k8soperation.svc -uroot -p
```

---

## 十九、环境变量与配置注入关系图

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Secret (k8soperation-secret)                 │
│                                                                     │
│  DB_PASSWORD / REDIS_PASSWORD / JWT_SIGNING_KEY / JENKINS_*         │
│  HMAC_SECRET / KUBECONFIG_ENCRYPT_KEY / DINGTALK_WEBHOOK            │
└─────────────────────┬───────────────────────────────────────────────┘
                      │ env.valueFrom.secretKeyRef
                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    Pod (k8soperation)                                │
│                                                                     │
│  环境变量: DB_PASSWORD=xxx, REDIS_PASSWORD=xxx, ...                 │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │  ConfigMap (k8soperation-config) → /app/configs/config.yaml  │   │
│  │  config.yaml 中使用 ${DB_PASSWORD} 占位符                    │   │
│  │  应用启动时读取环境变量替换配置值                              │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌────────────────┐  ┌────────────────┐                            │
│  │ PVC: artifacts │  │ PVC: logs      │                            │
│  │ /app/storage/  │  │ /app/storage/  │                            │
│  │   artifacts/   │  │   logs/        │                            │
│  └────────────────┘  └────────────────┘                            │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 二十、常见问题排查

| 问题 | 排查步骤 |
|------|---------|
| Pod CrashLoopBackOff | `kubectl logs <pod> -n k8soperation --previous` 查看崩溃日志 |
| PVC Pending | `kubectl describe pvc <name> -n k8soperation` 检查 StorageClass 是否存在 |
| 数据库连接失败 | 检查 MySQL Pod 是否 Ready，Service DNS 是否可达 |
| 前端 502 | 检查后端 Pod 是否健康，Nginx 反代配置是否正确 |
| Ingress 无法访问 | 检查 Ingress Controller 是否安装，域名解析是否正确 |
| 制品上传失败 | 检查 Ingress `proxy-body-size` 和 PVC 剩余空间 |
| WebSocket 断开 | 检查 Ingress `proxy-read-timeout` 是否足够长 |

---

## 附录A：完整资源清单

| 资源类型 | 名称 | 用途 |
|---------|------|------|
| Namespace | `k8soperation` | 平台专属命名空间 |
| Secret | `k8soperation-secret` | 敏感配置 |
| ConfigMap | `k8soperation-config` | 后端主配置 |
| ConfigMap | `k8soperation-nginx-config` | 前端 Nginx 配置 |
| ConfigMap | `mysql-config` | MySQL 配置 |
| ConfigMap | `mysql-initdb` | 数据库初始化 SQL |
| PVC | `k8soperation-artifacts` | 制品存储 20Gi |
| PVC | `k8soperation-logs` | 日志存储 5Gi |
| PVC | `mysql-data` | 数据库存储 50Gi |
| StatefulSet | `mysql` | MySQL 8.0 |
| Deployment | `redis` | Redis 7.x |
| Deployment | `k8soperation` | 后端 Go API |
| Deployment | `k8soperation-frontend` | 前端 Vue3+Nginx |
| Service | `mysql` | MySQL ClusterIP |
| Service | `redis` | Redis ClusterIP |
| Service | `k8soperation` | 后端 ClusterIP:8080 |
| Service | `k8soperation-frontend` | 前端 ClusterIP:80 |
| ServiceAccount | `k8soperation` | K8s API 操作权限 |
| ClusterRole | `k8soperation` | 跨NS资源管理 |
| ClusterRoleBinding | `k8soperation` | SA → Role 绑定 |
| Ingress | `k8soperation` | 统一外部入口 |
| Job | `db-init` | 数据库初始化（可选） |
| CronJob | `mysql-backup` | 数据库每日备份（可选） |

---

## 附录B：端口与服务发现

| 服务 | K8s Service DNS | 端口 | 用途 |
|------|----------------|------|------|
| 后端 API | `k8soperation.k8soperation.svc.cluster.local` | 8080 | REST API + WebSocket |
| 前端 | `k8soperation-frontend.k8soperation.svc.cluster.local` | 80 | 静态页面 + API 反代 |
| MySQL | `mysql.k8soperation.svc.cluster.local` | 3306 | 数据库 |
| Redis | `redis.k8soperation.svc.cluster.local` | 6379 | 缓存/Session |
