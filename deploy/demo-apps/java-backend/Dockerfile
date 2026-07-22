# ============================================================
# 多阶段构建：多模块 Maven Spring Boot 项目 Docker 化
# 来源：github.com/educostadev/Designing-Hexagonal-Architecture-with-Java-SpringBoot
#
# 模块结构：
#   parent (pom)
#     ├── domain       — 领域模型（无框架依赖）
#     ├── application  — 用例/端口接口
#     ├── infrastructure — 适配器实现（DB、REST、消息等）
#     └── bootstrap    — Spring Boot 启动类 + 配置
#
# 配置：bootstrap/src/main/resources/application.yml
# Docker Compose：infrastructure/datastore-mysql/docker-compose.yml
# ============================================================

# ==================== Stage 1: Maven 构建 ====================
FROM maven:3.9-eclipse-temurin-21-alpine AS build

WORKDIR /workspace

# 第 1 层：先复制 pom.xml（改源码不改 pom 时此层全量缓存）
COPY pom.xml ./
COPY domain/pom.xml           domain/pom.xml
COPY application/pom.xml      application/pom.xml
COPY infrastructure/pom.xml   infrastructure/pom.xml
COPY bootstrap/pom.xml        bootstrap/pom.xml

RUN mvn dependency:go-offline -B -q

# 第 2 层：复制源码并构建（跳过测试）
COPY domain/src           domain/src
COPY application/src      application/src
COPY infrastructure/src   infrastructure/src
COPY bootstrap/src        bootstrap/src

RUN mvn package -DskipTests -B -q

# ==================== Stage 2: 运行镜像 ====================
FROM eclipse-temurin:21-jre-alpine AS runtime

WORKDIR /app

# wget 用于健康检查（Alpine busybox wget）
# tzdata 用于时区
RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser

# 从构建阶段复制 JAR（bootstrap 是可运行模块）
COPY --from=build /workspace/bootstrap/target/*.jar app.jar

USER appuser

ENV JAVA_OPTS="-Xms256m -Xmx512m -XX:+UseG1GC -XX:MaxGCPauseMillis=200"

# 健康检查：如果项目含 Actuator 依赖用 /actuator/health
# 否则删掉这一行，K8s 用 TCP probe 代替
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:8080/actuator/health || exit 1

EXPOSE 8080

ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
