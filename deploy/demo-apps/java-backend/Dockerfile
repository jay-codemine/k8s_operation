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
# 关键：Docker 缓存优化 —— 先复制 pom.xml，再复制源码
# ============================================================

# ==================== Stage 1: Maven 构建 ====================
FROM maven:3.9-eclipse-temurin-21-alpine AS build

WORKDIR /workspace

# 第 1 层：仅复制父 POM，先安装以利用缓存
COPY pom.xml ./
# 多模块 pom 也一起复制（缓存友好——改代码不改 pom 时重用此层）
COPY domain/pom.xml    domain/pom.xml
COPY application/pom.xml application/pom.xml
COPY infrastructure/pom.xml infrastructure/pom.xml
COPY bootstrap/pom.xml   bootstrap/pom.xml

# 先下载所有依赖（不编译源码），修改源码时此层复用
RUN mvn dependency:go-offline -B -q

# 第 2 层：复制源码并构建
COPY domain/src       domain/src
COPY application/src   application/src
COPY infrastructure/src infrastructure/src
COPY bootstrap/src     bootstrap/src

RUN mvn package -DskipTests -B -q

# ==================== Stage 2: 运行镜像 ====================
FROM eclipse-temurin:21-jre-alpine AS runtime

WORKDIR /app

# 创建非 root 用户
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser

# 从构建阶段复制 JAR
COPY --from=build /workspace/bootstrap/target/*.jar app.jar

# 非 root 运行
USER appuser

# JVM 参数（生产优化）
ENV JAVA_OPTS="-Xms256m -Xmx512m -XX:+UseG1GC -XX:MaxGCPauseMillis=200"

HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/actuator/health || exit 1

EXPOSE 8080

ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
