# ============================================
# Java 项目 - 纯运行时 Dockerfile（平台编译模式）
# ============================================
# 设计理念：
#   - 不包含 Maven/Gradle 编译环境
#   - 仅接收平台 mvn package 产出的 JAR
#   - 使用 JRE 而非 JDK，镜像更小
#   - 生产级 JVM 参数调优
#   - 内置 OpenTelemetry Java Agent（可观测性）
#
# 配合流水线使用：
#   Jenkins Package 阶段产出 target/*.jar
#   流水线 Prepare OTEL Agent 阶段准备 opentelemetry-javaagent.jar
#   Build Image 阶段执行 nerdctl build
#
# OTEL Agent JAR 管理策略（由流水线统一管理，Dockerfile 仅做 COPY）：
#   全自动模式：流水线从平台「构建探针管理」自动拉取所有已启用 Agent
#   降级模式：项目自带 opentelemetry-javaagent.jar 或从 Maven 下载
#   流水线会将 agent jar 统一放到 .agents/{name}/ 目录
#
# 基础镜像说明：
#   使用阿里云镜像源，国内拉取更快
#   registry.cn-hangzhou.aliyuncs.com/k8s-gos/java:17-jre-alpine
# ============================================

ARG JAVA_VERSION=17
FROM registry.cn-hangzhou.aliyuncs.com/k8s-gos/java:${JAVA_VERSION}-jre-alpine

ENV TZ=Asia/Shanghai
WORKDIR /app

# 创建非 root 用户（Alpine 写法）
RUN addgroup -S appgroup && adduser -S -G appgroup appuser

# 创建日志目录并授权
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

# 复制 OpenTelemetry Java Agent（由流水线 Prepare Build Agents 阶段准备到 .agents/ 目录）
# 注意：若使用平台自动生成模式，此 COPY 行会被动态替换为所有已启用探针
COPY .agents/opentelemetry-javaagent/opentelemetry-javaagent.jar /app/opentelemetry-javaagent.jar

# 复制 JAR（由流水线 Package 阶段产出）
COPY target/*.jar /app/app.jar

# 切换用户
USER appuser

EXPOSE 8080

# OpenTelemetry Agent 配置（通过环境变量控制，部署时可覆盖）
# - service.name: 服务名称（部署时通过 K8s env 覆盖为实际服务名）
# - traces.exporter: 链路导出协议，默认 otlp
# - metrics/logs.exporter: 默认关闭，按需开启
# - endpoint: OTel Collector 地址（部署时覆盖为集群内实际地址）
ENV OTEL_OPTS="\
-javaagent:/app/opentelemetry-javaagent.jar \
-Dotel.service.name=java-app \
-Dotel.traces.exporter=otlp \
-Dotel.metrics.exporter=none \
-Dotel.logs.exporter=none \
-Dotel.exporter.otlp.endpoint=http://otel-collector-monitoring.svc.cluster.local:4318"

# 生产级 JVM 参数（可通过环境变量 JAVA_OPTS 覆盖）
# - MaxRAMPercentage: 容器内存自适应，比固定 -Xmx 更适合 K8s
# - G1GC: 低延迟垃圾回收
# - HeapDump: OOM 时自动生成堆转储
# - GC Log: 便于生产排查
ENV JAVA_OPTS="\
-XX:MaxRAMPercentage=75.0 \
-XX:+UseG1GC \
-XX:+HeapDumpOnOutOfMemoryError \
-XX:HeapDumpPath=/app/logs \
-Xlog:gc*:file=/app/logs/gc.log:time,uptime,level \
-Djava.security.egd=file:/dev/./urandom"

ENTRYPOINT ["sh", "-c", "exec java $OTEL_OPTS $JAVA_OPTS -jar /app/app.jar"]
