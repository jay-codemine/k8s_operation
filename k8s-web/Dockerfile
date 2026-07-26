# ============================================================
# Stage 1: Build
# ============================================================
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:22-alpine3.22 AS build

WORKDIR /app

# 利用 Docker 层缓存：先只拷贝依赖描述文件
COPY package*.json ./
RUN npm install

# 拷贝源码并构建
COPY . .
RUN npm run build

# ============================================================
# ★ 自动注入运行时配置脚本（不修改项目源码）
# 在构建产物 index.html 的 <head> 中插入 <script src="/config.js">
# config.js 由容器启动时 docker-entrypoint.sh 动态生成
# ============================================================
RUN sed -i 's|<head>|<head>\n    <script src="/config.js"><\/script>|' /app/dist/index.html

# ============================================================
# Stage 2: Runtime (Nginx)
# ============================================================
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/nginx:1.27-alpine

# 安装 gettext（提供 envsubst）用于启动时动态替换 nginx 配置变量
RUN apk add --no-cache gettext

# 替换默认 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 拷贝构建产物
COPY --from=build /app/dist /usr/share/nginx/html

# 拷贝启动脚本
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:80/ || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
