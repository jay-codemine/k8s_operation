 # ==============================================================================
# K8sOperation 前端 - Nginx 运行时 Dockerfile
# ==============================================================================
# 技术栈：Vue 3 + Vite + Arco Design
# 构建产物：Nginx 托管的静态文件 + API 反向代理
#
# 说明：CI 已通过 fe-compile-job 完成 npm build，产物为 dist/ 目录。
#       本 Dockerfile 仅负责将 dist 打包进 Nginx 镜像，不做二次编译。
#
# 构建命令（在项目根目录执行）：
#   docker build -f k8s-web/Dockerfile -t k8soperation-web:latest ./k8s-web
# ==============================================================================

FROM nginx:1.27-alpine

LABEL maintainer="K8sOperation Team"
LABEL description="K8sOperation Web Frontend (Vue3 + Nginx)"

# 复制 CI 构建好的静态产物
COPY dist/ /usr/share/nginx/html

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
    BACKEND_HOST="${API_BACKEND_URL#http://}"
    BACKEND_HOST="${BACKEND_HOST#https://}"
    # 不用 sed -i：ConfigMap subPath 挂载下 rename 会 Resource busy
    sed "s|server devops-be-svc.k8soperation.svc.cluster.local:8080;|server ${BACKEND_HOST};|g" \
        /etc/nginx/conf.d/default.conf > /tmp/default.conf.tmp
    if [ -s /tmp/default.conf.tmp ]; then
        cat /tmp/default.conf.tmp > /etc/nginx/conf.d/default.conf 2>/dev/null || \
            echo "WARNING: Cannot update backend URL (read-only mount?), using built-in config"
        rm -f /tmp/default.conf.tmp
    fi
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
