#!/bin/sh
set -e

# ============================================================
# 容器启动时动态生成 config.js（同一镜像可部署到多环境）
# K8s Deployment 中设置 env 即可切换环境，无需重新构建
# ============================================================

: ${API_BASE:="/api/v1"}
: ${ENV:="prod"}
: ${APP_NAME:="K8sOperation"}
: ${BACKEND_HOST:="k8soperation-backend"}
: ${BACKEND_PORT:="8080"}

cat > /usr/share/nginx/html/config.js <<EOF
// 由 docker-entrypoint.sh 在容器启动时动态生成
window._CONFIG = {
  API_BASE: "${API_BASE}",
  ENV: "${ENV}",
  APP_NAME: "${APP_NAME}",
}
EOF

echo "[entrypoint] config.js generated"
cat /usr/share/nginx/html/config.js

# 替换 nginx.conf 中的后端地址变量
export BACKEND_HOST BACKEND_PORT
envsubst '${BACKEND_HOST} ${BACKEND_PORT}' < /etc/nginx/conf.d/default.conf > /tmp/nginx.conf
mv /tmp/nginx.conf /etc/nginx/conf.d/default.conf

echo "[entrypoint] nginx → ${BACKEND_HOST}:${BACKEND_PORT}"

exec "$@"
