# ==============================================================================
# K8sOperation 后端 - 多阶段构建 Dockerfile（生产级，推荐）
# ==============================================================================
# 特点：
#   - Docker 内完成交叉编译，macOS/Windows/Linux 均可构建
#   - 无需本地 Go 环境
#   - 最终镜像 < 25MB（Alpine + 静态二进制）
#   - 非 root 用户运行
#   - 国内镜像加速（Go Proxy + Alpine 镜像源）
#
# 构建命令（在项目根目录执行）：
#   docker build -f docker/backend/Dockerfile -t devops-be:latest .
#
# 指定平台构建（Apple Silicon → Linux amd64）：
#   docker build -f docker/backend/Dockerfile --platform linux/amd64 -t devops-be:latest .
#
# 多架构构建：
#   docker buildx build -f docker/backend/Dockerfile --platform linux/amd64,linux/arm64 -t devops-be:latest --push .
# ==============================================================================

# ======================== Stage 1: 编译阶段 ========================
FROM docker.m.daocloud.io/library/golang:1.24-alpine AS builder

# 国内镜像加速
RUN sed -i 's#https://dl-cdn.alpinelinux.org#https://mirrors.aliyun.com#g' /etc/apk/repositories

# 安装编译依赖
RUN apk --no-cache add git ca-certificates tzdata

# Go 环境变量
ENV GOPROXY=https://goproxy.cn,direct
ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /build

# 利用 Docker 缓存：先下载依赖
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# 复制全部源码
COPY . .

# 编译（静态链接 + 压缩符号表）
RUN GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") && \
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    go build -trimpath \
        -ldflags="-s -w -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${BUILD_TIME}" \
        -o /app/devops-be ./cmd/k8soperation


# ======================== Stage 2: 运行阶段 ========================
FROM docker.m.daocloud.io/library/alpine:3.20

# 国内镜像加速
RUN sed -i 's#https://dl-cdn.alpinelinux.org#https://mirrors.aliyun.com#g' /etc/apk/repositories

# 安装运行时依赖（CA 证书 + 时区）
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建非 root 用户（UID=1000 与 K8s Deployment securityContext 一致）
RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app

WORKDIR /app

# 创建必要目录
RUN mkdir -p /app/storage/logs /app/storage/artifacts /app/configs && \
    chown -R app:app /app

# 从编译阶段复制二进制
COPY --from=builder /app/devops-be /app/devops-be

# 注：Jenkins/模板文件由 Git 仓库直接提供，不需要打进镜像

RUN chmod +x /app/devops-be && chown app:app /app/devops-be

USER app

# 环境变量
ENV GIN_MODE=release
ENV APP_CONFIG=/app/configs/config.yaml

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz/live || exit 1

ENTRYPOINT ["/app/devops-be"]
