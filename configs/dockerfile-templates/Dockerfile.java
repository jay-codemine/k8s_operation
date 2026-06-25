# ============================================
# Java 项目 - 纯运行时 Dockerfile（平台编译模式）
# ============================================
# 设计理念：
#   - 不包含 Maven/Gradle 编译环境
#   - 仅接收平台 mvn package 产出的 JAR
#   - 使用 JRE 而非 JDK，镜像更小
#   - 生产级 JVM 参数调优
#   - 探针注入由流水线 Prepare Build Agents 阶段自动完成
#
# 配合流水线使用：
#   Jenkins Package 阶段产出 target/*.jar
#   Prepare Build Agents 阶段从平台拉取已启用的探针（如 OTEL/SkyWalking）
#   Build & Push Image 阶段自动生成包含探针的 Dockerfile
#
# 探针管理说明：
#   本模板为「无探针基础版」，构建时不依赖任何外部 Agent JAR
#   流水线会根据平台「构建探针管理」配置自动注入探针 COPY/ENV
#   如平台 API 不可用，则跳过探针注入，镜像可正常运行
#
# 基础镜像说明：
#   使用 eclipse-temurin jammy (Ubuntu 22.04) 基底，自带 glibc
#   兼容 snappy-java / Netty / RocksDB 等所有原生库
#   支持 JAVA_VERSION: 11, 17, 21
# ============================================

ARG JAVA_VERSION=17
FROM eclipse-temurin:${JAVA_VERSION}-jre-jammy

ENV TZ=Asia/Shanghai
WORKDIR /app

# 安装时区配置 + 基础诊断工具
RUN apt-get update && apt-get install -y --no-install-recommends \
    tzdata curl && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    rm -rf /var/lib/apt/lists/*

# 创建非 root 用户（Debian/Ubuntu 写法）
RUN groupadd -r appgroup && useradd -r -g appgroup -d /app appuser

# 创建日志目录并授权
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app

# ==== 探针注入占位 ====
# 以下由流水线 Build & Push Image 阶段动态注入（平台可用时）：
#   COPY .agents/opentelemetry-javaagent/opentelemetry-javaagent.jar /app/opentelemetry-javaagent.jar
#   ENV OTEL_OPTS="-javaagent:/app/opentelemetry-javaagent.jar ..."
# 如需配置探针，请在平台「构建探针管理」中上传并启用，流水线会自动注入
# ========================

# 复制 JAR（由流水线 Package 阶段产出）
COPY target/*.jar /app/app.jar

# 切换用户
USER appuser

EXPOSE 8080

# 探针环境变量（默认为空，流水线动态生成 Dockerfile 时会注入实际值）
# 部署时也可通过 K8s env 覆盖，例如：
#   OTEL_OPTS: "-javaagent:/app/opentelemetry-javaagent.jar -Dotel.service.name=my-service"
#   JAVA_TOOL_OPTIONS: "-javaagent:/app/skywalking-agent.jar"
ENV OTEL_OPTS=""

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
