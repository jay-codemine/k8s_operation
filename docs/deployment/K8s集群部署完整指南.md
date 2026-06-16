# K8sOperation 平台 — K8s 集群部署完整指南

> 本文档覆盖从编译构建、镜像推送、数据库初始化到 K8s 资源部署的全流程。

---

## 目录

- [架构总览](#架构总览)
- [前置条件](#前置条件)
- [部署流程](#部署流程)
  - [Step 1: 编译 & 构建镜像](#step-1-编译--构建镜像)
  - [Step 2: 数据库初始化](#step-2-数据库初始化)
  - [Step 3: 配置 Secret（敏感信息）](#step-3-配置-secret敏感信息)
  - [Step 4: 配置 ConfigMap（主配置）](#step-4-配置-configmap主配置)
  - [Step 5: 一键部署](#step-5-一键部署)
- [K8s YAML 资源清单](#k8s-yaml-资源清单)
  - [1. Namespace](#1-namespace)
  - [2. Secret](#2-secret)
  - [3. ConfigMap](#3-configmap)
  - [4. PVC 持久化存储](#4-pvc-持久化存储)
  - [5. Service + ServiceAccount + RBAC](#5-service--serviceaccount--rbac)
  - [6. Deployment](#6-deployment)
  - [7. Ingress（可选）](#7-ingress可选)
  - [8. Kustomization 编排](#8-kustomization-编排)
- [Dockerfile](#dockerfile)
- [一键部署脚本](#一键部署脚本)
- [访问方式](#访问方式)
- [CI/CD 发布流程](#cicd-发布流程)
- [运维命令速查](#运维命令速查)
- [常见问题](#常见问题)

---

## 架构总览

```
┌───────────────────────────────────────────────────────────────────┐
│                        K8s Cluster                                 │
│                                                                   │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────────┐  │
│  │   Ingress    │────►│   Service    │────►│   Deployment     │  │
│  │  (nginx)     │     │  ClusterIP   │     │  k8soperation    │  │
│  │              │     │  :8080       │     │  (Go binary)     │  │
│  └──────────────┘     └──────────────┘     └────────┬─────────┘  │
│                                                      │            │
│         ┌────────────────────────────────────────────┼──────┐     │
│         │              Pod 内部                        │      │     │
│         │  ┌────────────┐  ┌──────────┐  ┌──────────┴───┐  │     │
│         │  │ ConfigMap  │  │  Secret  │  │ ServiceAccount│  │     │
│         │  │ config.yaml│  │ 密码/Token│  │  RBAC 权限   │  │     │
│         │  └────────────┘  └──────────┘  └──────────────┘  │     │
│         │  ┌────────────────────┐  ┌─────────────────────┐ │     │
│         │  │ PVC: artifacts 20Gi│  │ PVC: logs 5Gi       │ │     │
│         │  └────────────────────┘  └─────────────────────┘ │     │
│         └───────────────────────────────────────────────────┘     │
│                              │                                    │
│                    ┌─────────┴─────────┐                         │
│                    ▼                   ▼                          │
│              ┌──────────┐       ┌──────────┐                     │
│              │  MySQL   │       │  Redis   │                     │
│              │  8.x     │       │  6.x+    │                     │
│              └──────────┘       └──────────┘                     │
└───────────────────────────────────────────────────────────────────┘
```

---

## 前置条件

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| K8s 集群 | 1.22+ | 需有默认 StorageClass |
| kubectl | 1.22+ | 已配置好 kubeconfig |
| Docker/nerdctl | 20.10+ | 用于构建镜像 |
| Go | 1.22+ | 用于编译（或使用多阶段构建） |
| MySQL | 8.0+ | 可集群内/外部署 |
| Redis | 6.0+ | 可集群内/外部署 |
| 镜像仓库 | - | 集群可拉取（如阿里云 ACR、Harbor） |
| Ingress Controller | nginx-ingress（可选） | 域名暴露时需要 |

---

## 部署流程

### Step 1: 编译 & 构建镜像

```bash
# 1. 交叉编译 Linux 二进制
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" \
  -o bin/k8s_operation ./cmd/k8soperation/

# 2. 构建 Docker 镜像
docker build -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0 .

# 3. 推送到镜像仓库
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0
```

> **替代方案**：使用多阶段构建（无需本地 Go 环境）
> ```bash
> docker build -f docs/dockerfile/Dockerfile.golang.prod -t k8soperation:v1.0.0 .
> ```

### Step 2: 数据库初始化

**重要：平台不会自动建库建表，必须在部署前手动执行一次初始化 SQL。**

SQL 文件：`docs/sql/k8s_platform_full_init.sql`（约 1979 行，含 50 张表 + 种子数据）

```bash
# Linux / Mac / Git Bash
mysql -u root -p123456 --default-character-set=utf8mb4 < docs/sql/k8s_platform_full_init.sql

# PowerShell（不支持 < 重定向）
mysql -u root -p123456 --default-character-set=utf8mb4 -e "source D:/k8s-go/k8s_operation/docs/sql/k8s_platform_full_init.sql"

# 如果 MySQL 在 K8s Pod 内
kubectl cp docs/sql/k8s_platform_full_init.sql k8soperation/mysql-pod:/tmp/init.sql
kubectl exec -it -n k8soperation mysql-pod -- mysql -u root -p123456 -e "source /tmp/init.sql"
```

> SQL 为幂等脚本（`CREATE TABLE IF NOT EXISTS` + 幂等 ALTER），重复执行不会报错。

### Step 3: 配置 Secret（敏感信息）

修改 `deploy/secret.yaml`，将占位值替换为实际 base64 编码的密码：

```bash
# 生成 base64 值
echo -n "your-real-db-password" | base64
# 输出: eW91ci1yZWFsLWRiLXBhc3N3b3Jk
```

需替换的字段：
- `DB_PASSWORD` — MySQL 密码
- `REDIS_PASSWORD` — Redis 密码
- `JWT_SIGNING_KEY` — JWT 签名密钥（建议 16+ 位随机字符串）
- `JENKINS_URL` — Jenkins 地址（如 `http://jenkins:8080/`）
- `JENKINS_API_TOKEN` — Jenkins API Token
- `KUBECONFIG_ENCRYPT_KEY` — KubeConfig 加密密钥（至少 32 位）

### Step 4: 配置 ConfigMap（主配置）

修改 `deploy/configmap.yaml` 中的关键地址：

- **MySQL 地址**：`Host: mysql.k8soperation.svc` → 替换为实际地址
- **Redis 地址**：`Address: redis.k8soperation.svc:6379` → 替换为实际地址
- **Jenkins 回调**：`CallbackURL` 确保集群内可达

### Step 5: 一键部署

```bash
# 修改 deployment.yaml 中的镜像地址
# image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0

# 一键部署全部资源
kubectl apply -k deploy/

# 查看部署状态
kubectl rollout status deployment/k8soperation -n k8soperation --timeout=180s
```

---

## K8s YAML 资源清单

### 1. Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
    app.kubernetes.io/part-of: k8soperation-platform
```

---

### 2. Secret

```yaml
# 使用前请 base64 编码实际值替换下方占位符
# 生成命令：echo -n "your-value" | base64
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-secret
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
type: Opaque
data:
  # 数据库密码
  DB_PASSWORD: "Y2hhbmdlbWU="
  # Redis 密码
  REDIS_PASSWORD: "Y2hhbmdlbWU="
  # JWT 签名密钥（建议 16+ 位随机字符串）
  JWT_SIGNING_KEY: "ZW9OQjAlYnY1TTc5OTVGMQ=="
  # Jenkins API Token
  JENKINS_URL: "aHR0cDovL2plbmtpbnM6ODA4MC8="
  JENKINS_USERNAME: "YWRtaW4="
  JENKINS_API_TOKEN: "Y2hhbmdlbWU="
  # HMAC 签名密钥（Jenkins 回调验证）
  HMAC_SECRET: "Y2hhbmdlbWU="
  # KubeConfig 加密密钥（AES-256，至少 32 位）
  KUBECONFIG_ENCRYPT_KEY: "Y2hhbmdlbWU="
  # 钉钉 Webhook（可选）
  DINGTALK_WEBHOOK: ""
  # 前端公网地址（钉钉通知链接用）
  PLATFORM_FRONTEND_URL: "aHR0cDovL2xvY2FsaG9zdDo1MTcz"
```

---

### 3. ConfigMap

```yaml
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
      Host: mysql.k8soperation.svc        # K8s 内部 MySQL Service 地址
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
      Address: redis.k8soperation.svc:6379  # K8s 内部 Redis Service 地址
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
      GlobalKubeConfigPath: ""              # 留空 = InCluster 自动认证
      DefaultClusterID: 0
      AutoInitK8s: true
      AllowEmptyStart: true

    PodLog:
      EnableStreaming: false
      TailDefault: 500
      TailMax: 5000
      LimitBytes: 2097152

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
      CallbackURL: "http://k8soperation.k8soperation.svc:8080"
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

### 4. PVC 持久化存储

```yaml
# 制品存储（JAR/二进制/tar.gz 等 CI/CD 构建产物）
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
    - ReadWriteOnce           # 单副本用 RWO；多副本需改为 ReadWriteMany + NFS/CephFS
  storageClassName: ""        # 留空使用默认 StorageClass
  resources:
    requests:
      storage: 20Gi
---
# 应用日志存储（可选）
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
      storage: 5Gi
```

---

### 5. Service + ServiceAccount + RBAC

```yaml
# ClusterIP Service
apiVersion: v1
kind: Service
metadata:
  name: k8soperation
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: k8soperation
  ports:
    - name: http
      port: 8080
      targetPort: http
      protocol: TCP
---
# ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8soperation
  namespace: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
---
# ClusterRole：平台需要跨命名空间管理 K8s 资源
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
rules:
  # 核心资源
  - apiGroups: [""]
    resources:
      - pods
      - pods/log
      - pods/exec
      - pods/portforward
      - services
      - configmaps
      - secrets
      - persistentvolumes
      - persistentvolumeclaims
      - nodes
      - namespaces
      - events
      - serviceaccounts
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # 工作负载
  - apiGroups: ["apps"]
    resources:
      - deployments
      - statefulsets
      - daemonsets
      - replicasets
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # 批处理
  - apiGroups: ["batch"]
    resources:
      - jobs
      - cronjobs
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # 网络
  - apiGroups: ["networking.k8s.io"]
    resources:
      - ingresses
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # 存储
  - apiGroups: ["storage.k8s.io"]
    resources:
      - storageclasses
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # RBAC
  - apiGroups: ["rbac.authorization.k8s.io"]
    resources:
      - roles
      - clusterroles
      - rolebindings
      - clusterrolebindings
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # CRD
  - apiGroups: ["apiextensions.k8s.io"]
    resources:
      - customresourcedefinitions
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps.k8soperation.io"]
    resources: ["*"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
  # Metrics
  - apiGroups: ["metrics.k8s.io"]
    resources:
      - nodes
      - pods
    verbs: ["get", "list"]
---
# ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8soperation
  labels:
    app.kubernetes.io/name: k8soperation
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8soperation
subjects:
  - kind: ServiceAccount
    name: k8soperation
    namespace: k8soperation
```

---

### 6. Deployment

```yaml
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
      maxUnavailable: 0       # 先启新再下旧，零停机
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

          # === 环境变量（从 Secret 注入） ===
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

          # === 健康检查探针 ===
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
            failureThreshold: 30    # 最多等 150s 启动

          # === 资源限制 ===
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: "1"
              memory: 512Mi

          # === 挂载卷 ===
          volumeMounts:
            - name: config
              mountPath: /app/configs/config.yaml
              subPath: config.yaml
              readOnly: true
            - name: jenkins-templates
              mountPath: /app/configs/jenkins-templates
              readOnly: true
            - name: artifact-storage
              mountPath: /app/storage/artifacts
            - name: log-storage
              mountPath: /app/storage/logs

      # === 卷定义 ===
      volumes:
        - name: config
          configMap:
            name: k8soperation-config
        - name: jenkins-templates
          configMap:
            name: k8soperation-jenkins-templates
            optional: true        # 不存在则用镜像内置模板
        - name: artifact-storage
          persistentVolumeClaim:
            claimName: k8soperation-artifacts
        - name: log-storage
          persistentVolumeClaim:
            claimName: k8soperation-logs
```

---

### 7. Ingress（可选）

```yaml
# 需要集群已安装 Ingress Controller（如 nginx-ingress）
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
    # HTTPS（配合 cert-manager）
    # cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  rules:
    - host: k8soperation.example.com       # 替换为实际域名
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: k8soperation
                port:
                  name: http
  # HTTPS TLS 配置（可选）
  # tls:
  #   - hosts:
  #       - k8soperation.example.com
  #     secretName: k8soperation-tls
```

---

### 8. Kustomization 编排

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: k8soperation

resources:
  - namespace.yaml
  - secret.yaml
  - configmap.yaml
  - pvc.yaml
  - service.yaml       # 包含 Service + ServiceAccount + RBAC
  - deployment.yaml
  # - middleware.yaml   # 按需取消注释（K8s 内部署 MySQL + Redis）
  # - service-nodeport.yaml  # NodePort 暴露
  # - ingress.yaml      # Ingress 域名暴露

commonLabels:
  app.kubernetes.io/managed-by: kustomize
  app.kubernetes.io/part-of: k8soperation-platform
```

---

## Dockerfile

```dockerfile
FROM alpine:3.20

# 安装运行时依赖（CA证书 + 时区 + 健康检查工具）
RUN apk add --no-cache ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
RUN mkdir -p /app/storage/logs /app/configs

# 接收编译好的二进制
COPY bin/k8s_operation /app/k8s_operation
RUN chmod +x /app/k8s_operation

RUN chown -R app:app /app
USER app

ENV GIN_MODE=release
ENV APP_CONFIG=/app/configs/config.yaml
ENV K8S_CONFIG=/app/configs/k8s.yaml

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz/live || exit 1

ENTRYPOINT ["/app/k8s_operation"]
```

**镜像特点：**
- 基于 alpine:3.20，最终镜像 < 30MB
- 非 root 用户运行（app:app）
- 自带健康检查
- 仅包含运行时二进制，无编译环境

---

## 一键部署脚本

项目提供了一键部署脚本 `scripts/deploy-cluster.sh`，支持通过环境变量自定义配置：

```bash
# 设置环境变量
export REGISTRY="registry.cn-hangzhou.aliyuncs.com/k8s-gos"
export IMAGE_TAG="v1.0.0"
export DB_HOST="10.0.1.100"
export DB_PASSWORD="your-mysql-password"
export REDIS_HOST="10.0.1.100:6379"
export REDIS_PASS="your-redis-password"
export DOMAIN="k8sops.yourdomain.com"    # 可选

# 执行一键部署
chmod +x scripts/deploy-cluster.sh
./scripts/deploy-cluster.sh
```

脚本自动完成：
1. 环境检查（kubectl/docker）
2. 交叉编译 Go 二进制
3. 构建 & 推送 Docker 镜像
4. 生成 Secret（base64 编码实际密码）
5. 更新 ConfigMap（数据库/Redis 地址）
6. 更新 Deployment 镜像地址
7. 配置 Ingress（可选）
8. `kubectl apply -k deploy/` 一键部署
9. 等待 Pod 就绪

---

## 访问方式

### 方式一：Ingress（推荐生产环境）

1. 在 `kustomization.yaml` 中取消注释 `- ingress.yaml`
2. 修改 `ingress.yaml` 中的 `host` 为实际域名
3. 配置 DNS 解析指向 Ingress Controller IP
4. 访问：`http://k8soperation.yourdomain.com`

### 方式二：NodePort

创建 `deploy/service-nodeport.yaml`：
```yaml
apiVersion: v1
kind: Service
metadata:
  name: k8soperation-nodeport
  namespace: k8soperation
spec:
  type: NodePort
  selector:
    app.kubernetes.io/name: k8soperation
  ports:
    - port: 8080
      targetPort: http
      nodePort: 30080    # 30000-32767
```

访问：`http://<任意节点IP>:30080`

### 方式三：Port-Forward（临时调试）

```bash
kubectl port-forward svc/k8soperation -n k8soperation 8080:8080
# 访问: http://localhost:8080
```

---

## CI/CD 发布流程

**重要：CI/CD 流水线的部署阶段只做镜像滚动更新（Patch），不会创建新资源。**

因此首次部署必须先通过 `kubectl apply -k deploy/` 创建好 Deployment、Service 等基础资源。

后续 CI/CD 自动发布流程：

```
代码提交 → Jenkins 触发构建 → 编译 & 测试 → 构建镜像 & 推送
                                                    ↓
                              Jenkins 回调平台 → Patch Deployment 镜像
                                                    ↓
                              等待 Rollout 完成 → 钉钉/飞书通知
```

流水线配置关键字段：

| 字段 | 说明 | 示例 |
|------|------|------|
| `target_cluster_id` | 目标集群 ID | 1 |
| `target_namespace` | 命名空间 | k8soperation |
| `target_workload_kind` | 工作负载类型 | Deployment |
| `target_workload_name` | Deployment 名称 | k8soperation |
| `target_container` | 容器名称 | k8soperation |
| `auto_deploy` | 是否自动部署 | true |

---

## 运维命令速查

```bash
# 查看 Pod 状态
kubectl get pods -n k8soperation

# 查看实时日志
kubectl logs -n k8soperation -l app.kubernetes.io/name=k8soperation -f

# 进入容器 Shell
kubectl exec -it -n k8soperation deploy/k8soperation -- sh

# 查看 Deployment 详情
kubectl describe deployment k8soperation -n k8soperation

# 手动滚动更新（更新镜像）
kubectl set image deployment/k8soperation -n k8soperation \
  k8soperation=registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.1.0

# 回滚到上一版本
kubectl rollout undo deployment/k8soperation -n k8soperation

# 查看滚动更新历史
kubectl rollout history deployment/k8soperation -n k8soperation

# 扩缩容
kubectl scale deployment/k8soperation -n k8soperation --replicas=3

# 删除全部资源
kubectl delete -k deploy/

# 查看 Secret（解码）
kubectl get secret k8soperation-secret -n k8soperation -o jsonpath='{.data.DB_PASSWORD}' | base64 -d

# 查看 ConfigMap
kubectl get configmap k8soperation-config -n k8soperation -o yaml
```

---

## 常见问题

### Q1: Pod 启动失败 CrashLoopBackOff

```bash
# 查看日志
kubectl logs -n k8soperation -l app.kubernetes.io/name=k8soperation --previous

# 常见原因：
# 1. MySQL/Redis 地址不可达 → 检查 ConfigMap 中的 Host 地址
# 2. 数据库未初始化 → 执行 SQL 初始化脚本
# 3. Secret 值编码错误 → 重新 base64 编码
```

### Q2: PVC Pending

```bash
kubectl describe pvc k8soperation-artifacts -n k8soperation

# 常见原因：
# 1. 无默认 StorageClass → kubectl get sc，确认有 default
# 2. StorageClass 不支持 → 在 pvc.yaml 中指定具体 storageClassName
```

### Q3: 如何使用外部 MySQL/Redis

修改 `deploy/configmap.yaml` 中：
```yaml
Database:
  Host: 10.0.1.100              # 外部 MySQL IP
Cache:
  Address: 10.0.1.100:6379      # 外部 Redis 地址
```

如果外部数据库需要特殊网络策略，确保 Pod 网络可达。

### Q4: 多副本部署

1. PVC 需改为 `ReadWriteMany`（NFS/CephFS）
2. `deployment.yaml` 中 `replicas` 改为目标值
3. 或使用外部对象存储替代本地 PVC

### Q5: 默认登录账号

- 用户名：`admin`
- 密码：`admin123`

首次登录后建议立即修改密码。

---

## 文件对照表

| 文件路径 | 作用 |
|---------|------|
| `deploy/namespace.yaml` | 命名空间隔离 |
| `deploy/secret.yaml` | 敏感信息（密码/Token） |
| `deploy/configmap.yaml` | 主配置文件 config.yaml |
| `deploy/pvc.yaml` | 持久化存储（制品 20Gi + 日志 5Gi） |
| `deploy/service.yaml` | Service + ServiceAccount + ClusterRole + Binding |
| `deploy/deployment.yaml` | 核心工作负载（健康检查/资源限制/滚动更新） |
| `deploy/ingress.yaml` | 可选，域名暴露 |
| `deploy/kustomization.yaml` | Kustomize 编排入口 |
| `Dockerfile` | 生产级纯运行时镜像（alpine < 30MB） |
| `scripts/deploy-cluster.sh` | 一键部署脚本 |
| `docs/sql/k8s_platform_full_init.sql` | 数据库初始化（50 张表 + 种子数据） |
