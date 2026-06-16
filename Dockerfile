# ==============================================================================
# K8s Operation Platform - 纯运行时 Dockerfile（生产级）
# ==============================================================================
# 架构：平台编译 + Docker 纯打包
#
# 职责分工：
#   - Jenkins 平台：go mod download → go build → 产出二进制
#   - Dockerfile：仅将二进制打包为最小镜像（无任何编译环境）
#
# 构建方式：
#   方式1（默认）：先在 Jenkins/本地编译，再 docker build
#     $ go build -trimpath -ldflags="-s -w" -o bin/k8s_operation ./cmd/k8soperation
#     $ docker build -t k8soperation:latest .
#
#   方式2（全栈一体）：前后端多阶段构建，含 Nginx + supervisord
#     $ docker build -f docs/dockerfile/Dockerfile.k8soperation -t k8soperation:latest .
#
#   方式3（纯后端）：仅 API 服务，适合前后端分离部署
#     $ docker build -f docs/dockerfile/Dockerfile.k8soperation.api -t k8soperation-api:latest .
# ==============================================================================

FROM alpine:3.20

# 安装运行时依赖（CA证书 + 时区 + 健康检查工具）
RUN apk add --no-cache ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
RUN mkdir -p /app/storage/logs /app/configs

# 接收平台编译好的二进制（由 Jenkins Build Binary 阶段 或 本地 go build 产出）
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
