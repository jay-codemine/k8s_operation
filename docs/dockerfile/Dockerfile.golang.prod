# Dockerfile - Go 项目多阶段构建（生产级优化）
# ============================================
# 特点：
#   - 多阶段构建，最终镜像 < 20MB
#   - CGO 禁用，静态链接
#   - 非 root 用户运行
#   - 健康检查支持
#   - 支持配置文件挂载
# ============================================

# ============ 构建参数 ============
ARG GO_VERSION=1.22

# ============ 构建阶段 ============
FROM golang:${GO_VERSION}-alpine AS builder

# 构建参数
ARG APP_NAME=server
ARG MAIN_PATH=./cmd/main.go

# 设置国内代理
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 安装构建依赖（git 用于获取版本信息）
RUN apk --no-cache add git ca-certificates tzdata

WORKDIR /build

# 先复制 go.mod/go.sum 利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制源码
COPY . .

# 获取 Git 版本信息（如果可用）
RUN GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") && \
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown") && \
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    go build -ldflags="-s -w \
        -X main.GitCommit=${GIT_COMMIT} \
        -X main.GitBranch=${GIT_BRANCH} \
        -X main.BuildTime=${BUILD_TIME}" \
        -o /app/${APP_NAME} ${MAIN_PATH}

# ============ 运行阶段 ============
FROM alpine:3.19

ARG APP_NAME=server

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata

# 创建非 root 用户
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# 创建工作目录和配置目录
WORKDIR /app
RUN mkdir -p /app/configs /app/logs && \
    chown -R appuser:appgroup /app

# 从构建阶段复制二进制
COPY --from=builder /app/${APP_NAME} .

# 复制配置文件模板（如需要）
# COPY --chown=appuser:appgroup configs/ ./configs/

# 切换到非 root 用户
USER appuser

# 暴露端口（根据实际应用调整）
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# 启动命令
ENTRYPOINT ["./server"]

# 默认参数（可被覆盖）
# CMD ["-c", "./configs/config.yaml"]
