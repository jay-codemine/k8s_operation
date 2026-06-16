 # ==============================================================================
# K8sOperation 前端 - 多阶段构建 Dockerfile
# ==============================================================================
# 技术栈：Vue 3 + Vite + Arco Design
# 构建产物：Nginx 托管的静态文件 + API 反向代理
#
# 构建命令（在项目根目录执行）：
#   docker build -f k8s-web/Dockerfile -t k8soperation-web:latest ./k8s-web
#
# 运行命令：
#   docker run -d --name k8soperation-web \
#     -p 80:80 \
#     k8soperation-web:latest
#
# 自定义后端地址：
#   docker run -d --name k8soperation-web \
#     -p 80:80 \
#     -e API_BACKEND_URL=http://k8soperation-api:8080 \
#     k8soperation-web:latest
# ==============================================================================


# ======================== Stage 1: 构建阶段 ========================
FROM node:22-alpine AS builder

WORKDIR /app

# 设置国内 npm 镜像加速
RUN npm config set registry https://registry.npmmirror.com

# 利用 Docker 缓存：先复制依赖声明
COPY package.json package-lock.json ./

# 安装依赖（ci = clean install，确保可复现）
RUN npm ci --prefer-offline --no-audit --no-fund

# 复制源码并构建
COPY . .
RUN npm run build


# ======================== Stage 2: 运行阶段 ========================
FROM nginx:1.27-alpine

LABEL maintainer="K8sOperation Team"
LABEL description="K8sOperation Web Frontend (Vue3 + Nginx)"

# 安装运行时工具
#RUN apk --no-cache add tzdata && \
#    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
#    echo "Asia/Shanghai" > /etc/timezone && \
#    apk del tzdata

# 从构建阶段复制产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 写入 Nginx 配置
RUN <<'NGINX_CONF' cat > /etc/nginx/conf.d/default.conf
# ============================================================
# K8sOperation Frontend Nginx Config
# ============================================================

# 后端 API 上游（通过环境变量注入或默认值）
upstream backend {
    server devops-be-svc.k8soperation.svc.cluster.local:8080;
}

server {
    listen 80;
    server_name _;

    root /usr/share/nginx/html;
    index index.html;

    # ==================== Gzip 压缩 ====================
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/javascript
        application/json
        application/javascript
        application/xml
        image/svg+xml;

    # ==================== 静态资源缓存 ====================
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    # ==================== API 反向代理 ====================
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket 支持（容器终端、实时日志等）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;

        # 大文件上传支持
        client_max_body_size 200m;
    }

    # ==================== Swagger 文档代理 ====================
    location /swagger/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
    }

    # ==================== 健康检查 ====================
    location /health {
        access_log off;
        return 200 "ok";
        add_header Content-Type text/plain;
    }

    # ==================== Vue Router History 模式 ====================
    location / {
        try_files $uri $uri/ /index.html;
    }

    # ==================== 安全头 ====================
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
NGINX_CONF

# 创建 entrypoint 脚本（支持运行时注入后端地址）
RUN <<'SCRIPT' cat > /docker-entrypoint.d/90-backend-url.sh
#!/bin/sh
# 允许通过环境变量 API_BACKEND_URL 动态配置后端地址
# 用法：docker run -e API_BACKEND_URL=http://10.0.0.1:8080 ...
if [ -n "$API_BACKEND_URL" ]; then
    echo "Configuring backend URL: $API_BACKEND_URL"
    sed -i "s|server k8soperation-api:8080;|server ${API_BACKEND_URL#http://};|g" /etc/nginx/conf.d/default.conf
fi
SCRIPT
RUN chmod +x /docker-entrypoint.d/90-backend-url.sh

# 设置文件权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    touch /var/run/nginx.pid && \
    chown nginx:nginx /var/run/nginx.pid

EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
