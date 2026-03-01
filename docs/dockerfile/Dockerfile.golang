# Dockerfile - Go 项目多阶段构建
# ============================================
# 特点：
#   - 多阶段构建，最终镜像仅包含二进制
#   - 使用 distroless 或 alpine 基础镜像
#   - 支持健康检查
# ============================================

# ============ 构建阶段 ============
FROM golang:1.22-alpine AS builder

# 设置国内代理
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /build

# 先复制 go.mod 利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建
COPY . .
RUN go build -ldflags="-s -w" -o /app/server ./cmd/main.go

# ============ 运行阶段 ============
FROM alpine:3.19

# 安装时区和 CA 证书
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /app/server .

# 复制配置文件（如需要）
# COPY configs/ ./configs/

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

# 启动命令
ENTRYPOINT ["./server"]
