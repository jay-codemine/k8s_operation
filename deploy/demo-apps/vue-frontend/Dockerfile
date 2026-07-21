# ============================================================
# 多阶段构建：Vue3 + Vite 前端 Docker 化
# 基于真实项目 kristiyan-velkov/docker-vuejs-sample
#   Vue 3.5 + Vite 7.1 + TypeScript 5.9 + Vue Router 4.5
#   Node 22（匹配 @tsconfig/node22）
# ============================================================

# ==================== Stage 1: Node 构建 ====================
FROM node:22-alpine AS build

WORKDIR /app

# --- 安装依赖（用 npm install，兼容性更好）---
COPY package.json ./
RUN npm install

# --- 复制源码 ---
COPY . .

# --- 编译（类型检查 + 打包）---
RUN npm run build
#   build = vue-tsc --build（类型检查）+ vite build（打包）→ 输出到 dist/

# ==================== Stage 2: Nginx ====================
FROM nginx:1.27-alpine AS runtime

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q --spider http://localhost:80/ || exit 1

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
