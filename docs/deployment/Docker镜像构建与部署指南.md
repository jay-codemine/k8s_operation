# K8sOperation Docker 镜像构建与 K8s 部署完整指南

> 本文档从 **Dockerfile 编写 → 镜像构建 → 本地验证 → K8s 部署** 全流程详细说明，涵盖前后端独立构建与部署。

---

## 一、整体流程总览

```
第1步                    第2步                    第3步                    第4步
编写 Dockerfile          构建镜像                 本地 Docker 运行验证      部署到 Kubernetes
─────────────────── → ─────────────────── → ─────────────────── → ───────────────────
定义构建规则              docker build            docker run              kubectl apply
                          docker push             curl / 浏览器访问        验证 Pod/Service/Ingress
```

---

## 二、Dockerfile 文件清单

| 文件路径 | 用途 | 构建阶段 | 镜像大小 |
|---------|------|---------|---------|
| `k8s-web/Dockerfile` | 前端（Vue3 + Vite → Nginx） | 2 阶段 | ~30MB |
| `docs/dockerfile/Dockerfile.k8soperation.api` | 后端 Go（源码编译 → Alpine） | 2 阶段 | ~25MB |
| `docs/dockerfile/Dockerfile.k8soperation` | 全栈一体（前+后+Nginx+supervisord） | 3 阶段 | ~55MB |
| `Dockerfile` (根目录) | Jenkins 预编译模式（仅打包） | 1 阶段 | ~15MB |

---

## 三、前端 Dockerfile 详解

### 3.1 文件位置

[`k8s-web/Dockerfile`](/Users/dai/k8s_operation-main/k8s-web/Dockerfile)

### 3.2 Dockerfile 完整代码

```dockerfile
# ==============================================================================
# K8sOperation 前端 - 多阶段构建 Dockerfile
# 技术栈：Vue 3 + Vite + Arco Design
# 运行环境：nginx:1.27-alpine
# ==============================================================================

# ======================== Stage 1: 构建阶段 ========================
FROM node:22-alpine AS builder

WORKDIR /app

# 设置国内 npm 镜像加速
RUN npm config set registry https://registry.npmmirror.com

# 利用 Docker 缓存：先复制依赖声明
COPY package.json package-lock.json ./

# 安装依赖
RUN npm ci --prefer-offline --no-audit --no-fund

# 复制源码并构建
COPY . .
RUN npm run build


# ======================== Stage 2: 运行阶段 ========================
FROM nginx:1.27-alpine

LABEL maintainer="K8sOperation Team"
LABEL description="K8sOperation Web Frontend (Vue3 + Nginx)"

# 安装时区
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 从构建阶段复制产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 写入 Nginx 配置
RUN cat > /etc/nginx/conf.d/default.conf << 'NGINX_CONF'
upstream backend {
    server k8soperation-api:8080;
}

server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_comp_level 6;
    gzip_types text/plain text/css application/json application/javascript image/svg+xml;

    # 静态资源缓存 30 天
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # API 反向代理
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        client_max_body_size 200m;
    }

    # 健康检查
    location /health {
        access_log off;
        return 200 "ok";
        add_header Content-Type text/plain;
    }

    # Vue Router History 模式回退
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
NGINX_CONF

# 运行时后端地址注入脚本
RUN cat > /docker-entrypoint.d/90-backend-url.sh << 'SCRIPT'
#!/bin/sh
if [ -n "$API_BACKEND_URL" ]; then
    echo "Configuring backend URL: $API_BACKEND_URL"
    sed -i "s|server k8soperation-api:8080;|server ${API_BACKEND_URL#http://};|g" /etc/nginx/conf.d/default.conf
fi
SCRIPT
RUN chmod +x /docker-entrypoint.d/90-backend-url.sh

# 权限设置
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    touch /var/run/nginx.pid && \
    chown nginx:nginx /var/run/nginx.pid

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
```

### 3.3 构建流程图

```
┌───────────────────────────────┐
│  Stage 1: node:22-alpine       │
│                               │
│  1. npm ci (安装依赖)          │
│  2. npm run build (Vite 编译)  │
│  3. 产出 dist/ 目录            │
├───────────────────────────────┤
│  Stage 2: nginx:1.27-alpine    │
│                               │
│  1. 复制 dist/ → /usr/share/   │
│  2. 写入 Nginx 反代配置         │
│  3. 注入后端地址环境变量脚本     │
│  4. 最终镜像 ≈ 30MB            │
└───────────────────────────────┘
```

### 3.4 配套文件

```
k8s-web/
├── Dockerfile          ← 前端 Dockerfile
├── .dockerignore       ← 排除 node_modules、.git 等
├── package.json
├── vite.config.js
└── src/
```

`.dockerignore` 内容：

```
node_modules
dist
.git
.vscode
*.md
.env.development
.cache
```

---

## 四、后端 Dockerfile 详解

### 4.1 文件位置

[`docs/dockerfile/Dockerfile.k8soperation.api`](/Users/dai/k8s_operation-main/docs/dockerfile/Dockerfile.k8soperation.api)

### 4.2 Dockerfile 完整代码

```dockerfile
# ==============================================================================
# K8sOperation 后端 - 多阶段构建 Dockerfile（纯 API 服务）
# 技术栈：Go 1.24 + Gin + client-go
# 运行环境：alpine:3.20
# ==============================================================================

# ======================== 构建阶段 ========================
FROM golang:1.24-alpine AS builder

ARG APP_VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk --no-cache add git ca-certificates tzdata

WORKDIR /build

# 依赖缓存层
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 源码编译
COPY . .
RUN GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")} && \
    BUILD_TIME=${BUILD_TIME:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")} && \
    go build -trimpath \
        -ldflags="-s -w \
            -X main.Version=${APP_VERSION} \
            -X main.GitCommit=${GIT_COMMIT} \
            -X main.BuildTime=${BUILD_TIME}" \
        -o /app/k8soperation ./cmd/k8soperation

# ======================== 运行阶段 ========================
FROM alpine:3.20

LABEL maintainer="K8sOperation Team"
LABEL description="K8sOperation API Server"

# 运行时依赖
RUN apk --no-cache add ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
RUN mkdir -p /app/configs /app/storage/logs /app/storage/artifacts && \
    chown -R app:app /app

# 复制二进制
COPY --from=builder /app/k8soperation /app/k8soperation
RUN chmod +x /app/k8soperation

# 复制配置模板
COPY configs/config.yaml.example /app/configs/config.yaml.example
COPY configs/k8s.yaml.example /app/configs/k8s.yaml.example

USER app

ENV GIN_MODE=release
ENV APP_CONFIG=/app/configs/config.yaml
ENV K8S_CONFIG=/app/configs/k8s.yaml

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz/live || exit 1

ENTRYPOINT ["/app/k8soperation"]
```

### 4.3 构建流程图

```
┌───────────────────────────────────────────┐
│  Stage 1: golang:1.24-alpine (构建)        │
│                                           │
│  1. go mod download (下载依赖)             │
│  2. go build (静态编译)                    │
│  3. 产出 /app/k8soperation 二进制           │
├───────────────────────────────────────────┤
│  Stage 2: alpine:3.20 (运行)               │
│                                           │
│  1. 安装 ca-certificates + wget            │
│  2. 创建 app 非 root 用户                  │
│  3. COPY --from=builder 二进制             │
│  4. 最终镜像 ≈ 25MB                        │
└───────────────────────────────────────────┘
```

### 4.4 关键设计说明

| 设计点 | 说明 |
|--------|------|
| `CGO_ENABLED=0` | 静态链接，不依赖系统 libc |
| `-ldflags="-s -w"` | 去掉调试信息，减小体积 |
| `-trimpath` | 去除二进制中的本地路径 |
| `go mod verify` | 验证依赖完整性 |
| 非 root 用户 | 安全加固，防止提权 |

---

## 五、镜像构建命令

### 5.1 构建前端镜像

```bash
# 在项目根目录执行
docker build -f k8s-web/Dockerfile \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-web:v1.0.0 \
  ./k8s-web
```

### 5.2 构建后端镜像

```bash
# 在项目根目录执行
docker build -f docs/dockerfile/Dockerfile.k8soperation.api \
  --build-arg APP_VERSION=v1.0.0 \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -t registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0 \
  .
```

### 5.3 推送到镜像仓库

```bash
# 登录仓库（首次执行）
docker login registry.cn-hangzhou.aliyuncs.com

# 推送前端
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-web:v1.0.0

# 推送后端
docker push registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0
```

### 5.4 验证镜像

```bash
# 查看镜像
docker images | grep k8soperation

# 查看镜像分层
docker history registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0

# 查看镜像详情
docker inspect registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0
```

---

## 六、本地 Docker 运行验证

### 6.1 启动后端

```bash
docker run -d --name k8soperation-api \
  -p 8080:8080 \
  -v $(pwd)/configs/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/configs/k8s.yaml:/app/configs/k8s.yaml:ro \
  registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0
```

### 6.2 启动前端

```bash
docker run -d --name k8soperation-web \
  -p 80:80 \
  -e API_BACKEND_URL=http://host.docker.internal:8080 \
  registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-web:v1.0.0
```

> `host.docker.internal` 是 Docker 访问 Mac 宿主机的特殊 DNS 地址。

### 6.3 验证服务

```bash
# 查看容器状态
docker ps

# 后端健康检查
curl http://localhost:8080/healthz/live    # → ok

# 前端健康检查
curl http://localhost:80/health           # → ok

# 浏览器访问
open http://localhost
```

### 6.4 查看日志

```bash
# 后端日志
docker logs -f k8soperation-api

# 前端日志（Nginx）
docker logs -f k8soperation-web
```

### 6.5 清理容器

```bash
docker stop k8soperation-web k8soperation-api
docker rm k8soperation-web k8soperation-api
```

---

## 七、部署到 Kubernetes

### 7.1 部署架构

```
                    ┌──────────────────────────────────────────────┐
                    │              Kubernetes Cluster               │
                    │                                              │
  浏览器 ──Ingress──▶  ┌─────────────────────┐                     │
                    │  │  k8soperation-web   │  (Nginx, port 80)   │
                    │  │  /api/* ──proxy──▶  │──────────────────┐  │
                    │  │  /* → index.html    │                  │  │
                    │  └─────────────────────┘                  ▼  │
                    │                              ┌─────────────────┐
                    │                              │ k8soperation-api │
                    │                              │ (Go, port 8080)  │
                    │                              │                  │
                    │                              │  ├── MySQL       │
                    │                              │  ├── Redis       │
                    │                              │  └── PVC(日志/制品)
                    │                              └─────────────────┘
                    └──────────────────────────────────────────────┘
```

### 7.2 修改部署文件

**后端** [`deploy/deployment.yaml`](/Users/dai/k8s_operation-main/deploy/deployment.yaml)：

```yaml
# 修改第 39 行，替换为你的镜像地址和版本
image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.0.0
```

**前端** [`deploy/frontend-deployment.yaml`](/Users/dai/k8s_operation-main/deploy/frontend-deployment.yaml)：

```yaml
# 修改后端 Service 地址（默认已正确配置）
env:
  - name: API_BACKEND_URL
    value: "http://k8soperation.k8soperation.svc:8080"

# 修改第 39 行，替换为你的镜像地址和版本
image: registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-web:v1.0.0

# 修改 Ingress 域名（第 117 行）
- host: k8sop.example.com                  # ← 替换为你的域名
```

### 7.3 部署后端

```bash
# 一键部署（后端 + Secret + ConfigMap + PVC + Service）
kubectl apply -k deploy/

# 或者按顺序手动部署
kubectl apply -f deploy/namespace.yaml
kubectl apply -f deploy/secret.yaml
kubectl apply -f deploy/configmap.yaml
kubectl apply -f deploy/pvc.yaml
kubectl apply -f deploy/service.yaml
kubectl apply -f deploy/deployment.yaml
```

### 7.4 部署前端

```bash
kubectl apply -f deploy/frontend-deployment.yaml
```

### 7.5 验证部署

```bash
# 查看 Pod 状态
kubectl get pods -n k8soperation

# 期望输出：
# NAME                                  READY   STATUS    RESTARTS   AGE
# k8soperation-xxxx-xxxx                1/1     Running   0          2m
# k8soperation-web-xxxx-xxxx            1/1     Running   0          2m
# k8soperation-web-xxxx-yyyy            1/1     Running   0          2m

# 查看后端日志
kubectl logs -n k8soperation -l app.kubernetes.io/name=k8soperation -f

# 查看前端日志
kubectl logs -n k8soperation -l app.kubernetes.io/name=k8soperation-web -f

# 查看 Service
kubectl get svc -n k8soperation

# 查看 Ingress
kubectl get ingress -n k8soperation
```

### 7.6 访问服务

```bash
# 方式1：通过 Ingress（推荐，需配置域名）
open https://k8sop.example.com

# 方式2：通过 NodePort（测试用）
# 取消 deploy/service-nodeport.yaml 的注释后执行：
kubectl apply -f deploy/service-nodeport.yaml
kubectl get svc k8soperation-web -n k8soperation

# 方式3：端口转发（快速测试）
kubectl port-forward -n k8soperation svc/k8soperation-web 8080:80
open http://localhost:8080
```

### 7.7 更新镜像（发布新版本）

```bash
# 方式1：修改 YAML 后重新 apply
# 编辑 deploy/deployment.yaml，修改 image tag
kubectl apply -f deploy/deployment.yaml

# 方式2：直接修改（无需编辑文件）
kubectl set image deployment/k8soperation \
  k8soperation=registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:v1.1.0 \
  -n k8soperation

kubectl set image deployment/k8soperation-web \
  web=registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation-web:v1.1.0 \
  -n k8soperation

# 查看滚动更新进度
kubectl rollout status deployment/k8soperation -n k8soperation
kubectl rollout status deployment/k8soperation-web -n k8soperation
```

### 7.8 回滚

```bash
# 查看历史版本
kubectl rollout history deployment/k8soperation -n k8soperation

# 回滚到上一个版本
kubectl rollout undo deployment/k8soperation -n k8soperation

# 回滚到指定版本
kubectl rollout undo deployment/k8soperation -n k8soperation --to-revision=2
```

---

## 八、本地完整测试（Docker Compose）

适合本地一键启动全部组件（MySQL + Redis + 后端 + 前端）：

```yaml
# docker-compose.yaml（放在项目根目录）
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: admin123
      MYSQL_DATABASE: k8s-platform
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass admin123
    ports:
      - "6379:6379"

  api:
    build:
      context: .
      dockerfile: docs/dockerfile/Dockerfile.k8soperation.api
      args:
        APP_VERSION: dev
    ports:
      - "8080:8080"
    volumes:
      - ./configs/config.yaml:/app/configs/config.yaml:ro
    depends_on:
      - mysql
      - redis

  web:
    build:
      context: ./k8s-web
      dockerfile: Dockerfile
    ports:
      - "80:80"
    environment:
      - API_BACKEND_URL=http://api:8080
    depends_on:
      - api

volumes:
  mysql-data:
```

```bash
# 一键启动
docker compose up -d

# 访问
# 前端：http://localhost
# API：http://localhost:8080
# 默认账号：admin / admin123
```

---

## 九、Dockerfile 文件索引

```
项目根目录/
├── Dockerfile                                            # Jenkins 预编译模式
├── .dockerignore                                         # 全局忽略文件
├── k8s-web/
│   ├── Dockerfile                                       # ★ 前端多阶段构建
│   └── .dockerignore                                    # 前端忽略文件
├── docs/dockerfile/
│   ├── Dockerfile.k8soperation                          # ★ 全栈一体
│   ├── Dockerfile.k8soperation.api                      # ★ 纯后端 API
│   ├── Dockerfile.golang                                # 通用 Go 模板
│   ├── Dockerfile.golang.prod                           # 通用 Go 生产模板
│   ├── Dockerfile.nginx                                 # Nginx 模板
│   └── Dockerfile.python                                # Python 模板
└── deploy/
    ├── deployment.yaml                                  # 后端 K8s 部署
    └── frontend-deployment.yaml                         # ★ 前端 K8s 部署
```

---

## 十、生产最佳实践

| 实践项 | 说明 |
|--------|------|
| 非 root 运行 | 后端用 `app` 用户，Nginx 用 `nginx` 用户 |
| 最小镜像 | Alpine base，无编译工具链 |
| 分层缓存 | go.mod / package.json 先复制，命中缓存加速构建 |
| 健康检查 | HEALTHCHECK + K8s probes 双重保障 |
| 配置外挂 | 通过 Volume/Secret 注入，不写死进镜像 |
| 滚动更新 | K8s Deployment 策略：maxUnavailable=0, maxSurge=1 |
| 资源限制 | 后端 512Mi / 前端 256Mi |
