# 多语言 Dockerfile 与 Jenkinsfile 模板指南

## 目录
1. [概述](#一概述)
2. [Go 项目](#二go-项目)
3. [Java Maven 项目](#三java-maven-项目)
4. [Python 项目](#四python-项目)
5. [Nginx 静态项目](#五nginx-静态项目)
6. [通用 Jenkinsfile（平台 Worker 模式）](#六通用-jenkinsfile平台-worker-模式)

---

## 一、概述

### 模式说明

本文档提供 **平台 Worker 部署模式** 下的多语言模板：
- **Jenkins 职责**：CI（代码拉取 → 编译/打包 → 镜像构建 → 推送 Harbor → 回调平台）
- **平台 Worker 职责**：CD（接收回调 → 部署到 K8s）
- **Jenkins 不需要 kubeconfig 凭证**

### 项目结构标准

```
your-project/
├── Dockerfile                    # 镜像构建文件
├── Jenkinsfile.build-only        # Jenkins 流水线（仅构建）
├── src/                          # 源代码目录
└── ...
```

---

## 二、Go 项目

### 2.1 Dockerfile（多阶段构建）

```dockerfile
# ============================================================================
# Dockerfile - Go 项目多阶段构建
# 适用：Gin / Echo / Fiber 等 Go Web 框架
# ============================================================================

# ==================== 阶段1：构建 ====================
FROM golang:1.21-alpine AS builder

# 设置 Go 环境
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn,direct

WORKDIR /build

# 先复制依赖文件，利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN go build -ldflags="-s -w" -o app ./cmd/main.go

# ==================== 阶段2：运行 ====================
FROM alpine:3.18

# 安装基础工具（可选）
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/app .

# 复制配置文件（如果有）
# COPY --from=builder /build/configs ./configs

# 非 root 用户运行（安全最佳实践）
RUN adduser -D -u 1000 appuser
USER appuser

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./app"]
```

### 2.2 项目结构示例

```
go-project/
├── Dockerfile
├── Jenkinsfile.build-only
├── go.mod
├── go.sum
├── cmd/
│   └── main.go                   # 入口文件
├── internal/
│   ├── handler/                  # HTTP 处理器
│   ├── service/                  # 业务逻辑
│   └── repository/               # 数据访问
├── configs/
│   └── config.yaml               # 配置文件
└── Makefile
```

### 2.3 示例 main.go

```go
// cmd/main.go
package main

import (
    "log"
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 健康检查接口
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })
    
    // 业务接口
    r.GET("/api/hello", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Hello from Go!"})
    })
    
    log.Println("Server starting on :8080")
    r.Run(":8080")
}
```

---

## 三、Java Maven 项目

### 3.1 Dockerfile（多阶段构建）

```dockerfile
# ============================================================================
# Dockerfile - Java Maven 项目多阶段构建
# 适用：Spring Boot / Spring Cloud 项目
# ============================================================================

# ==================== 阶段1：构建 ====================
FROM maven:3.9-eclipse-temurin-17 AS builder

WORKDIR /build

# 先复制 pom.xml，利用依赖缓存
COPY pom.xml .
RUN mvn dependency:go-offline -B

# 复制源码并打包
COPY src ./src
RUN mvn package -DskipTests -B

# ==================== 阶段2：运行 ====================
FROM eclipse-temurin:17-jre-alpine

# 设置时区
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从构建阶段复制 JAR 文件
COPY --from=builder /build/target/*.jar app.jar

# 非 root 用户运行
RUN adduser -D -u 1000 appuser
USER appuser

EXPOSE 8080

# JVM 参数优化（可通过环境变量覆盖）
ENV JAVA_OPTS="-Xms256m -Xmx512m -XX:+UseG1GC -XX:MaxGCPauseMillis=200"

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/actuator/health || exit 1

ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

### 3.2 简化版 Dockerfile（无多阶段，JAR 预构建）

```dockerfile
# ============================================================================
# Dockerfile - Java 简化版（JAR 已在 Jenkins 构建）
# 使用场景：Jenkins 已执行 mvn package，直接复制 JAR
# ============================================================================

FROM eclipse-temurin:17-jre-alpine

RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

# 复制 Jenkins 构建的 JAR 文件
COPY target/*.jar app.jar

RUN adduser -D -u 1000 appuser
USER appuser

EXPOSE 8080

ENV JAVA_OPTS="-Xms256m -Xmx512m"

HEALTHCHECK --interval=30s --timeout=3s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/actuator/health || exit 1

ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

### 3.3 项目结构示例

```
java-project/
├── Dockerfile
├── Jenkinsfile.build-only
├── pom.xml
├── src/
│   └── main/
│       ├── java/
│       │   └── com/example/demo/
│       │       ├── DemoApplication.java
│       │       └── controller/
│       │           └── HealthController.java
│       └── resources/
│           └── application.yml
└── Makefile
```

### 3.4 示例 application.yml

```yaml
# src/main/resources/application.yml
server:
  port: 8080

spring:
  application:
    name: java-demo

management:
  endpoints:
    web:
      exposure:
        include: health,info
  endpoint:
    health:
      show-details: always
```

---

## 四、Python 项目

### 4.1 Dockerfile（多阶段构建）

```dockerfile
# ============================================================================
# Dockerfile - Python 项目多阶段构建
# 适用：Flask / Django / FastAPI 项目
# ============================================================================

# ==================== 阶段1：依赖安装 ====================
FROM python:3.11-slim AS builder

WORKDIR /build

# 安装构建工具
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# 创建虚拟环境
RUN python -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# 安装依赖
COPY requirements.txt .
RUN pip install --no-cache-dir -i https://pypi.tuna.tsinghua.edu.cn/simple \
    --upgrade pip && \
    pip install --no-cache-dir -i https://pypi.tuna.tsinghua.edu.cn/simple \
    -r requirements.txt

# ==================== 阶段2：运行 ====================
FROM python:3.11-slim

# 设置时区
RUN apt-get update && apt-get install -y --no-install-recommends \
    tzdata wget \
    && rm -rf /var/lib/apt/lists/* \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

# 从构建阶段复制虚拟环境
COPY --from=builder /opt/venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# 复制源码
COPY . .

# 非 root 用户运行
RUN useradd -m -u 1000 appuser
USER appuser

EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# FastAPI 示例（使用 uvicorn）
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]

# Flask 示例
# CMD ["gunicorn", "--bind", "0.0.0.0:8000", "app:app"]

# Django 示例
# CMD ["gunicorn", "--bind", "0.0.0.0:8000", "myproject.wsgi:application"]
```

### 4.2 简化版 Dockerfile

```dockerfile
# ============================================================================
# Dockerfile - Python 简化版
# ============================================================================

FROM python:3.11-slim

WORKDIR /app

# 设置 pip 镜像源
ENV PIP_INDEX_URL=https://pypi.tuna.tsinghua.edu.cn/simple

# 安装依赖
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# 复制源码
COPY . .

EXPOSE 8000

# FastAPI
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### 4.3 项目结构示例

```
python-project/
├── Dockerfile
├── Jenkinsfile.build-only
├── requirements.txt
├── main.py                       # FastAPI 入口
├── app/
│   ├── __init__.py
│   ├── api/
│   │   └── routes.py
│   ├── models/
│   │   └── schemas.py
│   └── services/
│       └── business.py
└── tests/
    └── test_api.py
```

### 4.4 示例 main.py（FastAPI）

```python
# main.py
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="Python Demo API")

class HealthResponse(BaseModel):
    status: str

@app.get("/health", response_model=HealthResponse)
def health_check():
    return {"status": "healthy"}

@app.get("/api/hello")
def hello():
    return {"message": "Hello from Python!"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

### 4.5 示例 requirements.txt

```
fastapi==0.104.1
uvicorn[standard]==0.24.0
pydantic==2.5.2
```

---

## 五、Nginx 静态项目

### 5.1 Dockerfile

```dockerfile
# ============================================================================
# Dockerfile - Nginx 静态项目
# 适用：Vue / React / Angular 前端项目
# ============================================================================

# ==================== 阶段1：构建（如需 Node.js 构建）====================
FROM node:18-alpine AS builder

WORKDIR /build

# 安装依赖
COPY package*.json ./
RUN npm ci --registry=https://registry.npmmirror.com

# 构建
COPY . .
RUN npm run build

# ==================== 阶段2：运行 ====================
FROM nginx:alpine

# 设置时区
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 删除默认配置
RUN rm -rf /usr/share/nginx/html/*

# 从构建阶段复制静态文件
COPY --from=builder /build/dist /usr/share/nginx/html

# 自定义 Nginx 配置（可选）
# COPY nginx.conf /etc/nginx/nginx.conf
COPY nginx.default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
```

### 5.2 简化版 Dockerfile（纯静态文件）

```dockerfile
# ============================================================================
# Dockerfile - Nginx 简化版（无 Node.js 构建）
# 使用场景：静态 HTML 文件或已构建的 dist 目录
# ============================================================================

FROM nginx:alpine

RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# 复制静态文件
COPY html/ /usr/share/nginx/html/

# 自定义配置（可选）
# COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
```

### 5.3 Nginx 配置文件示例

```nginx
# nginx.default.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # 健康检查接口
    location /health {
        access_log off;
        return 200 '{"status":"healthy"}';
        add_header Content-Type application/json;
    }

    # SPA 路由支持
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Gzip 压缩
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
    gzip_min_length 1024;
}
```

### 5.4 项目结构示例

```
nginx-project/
├── Dockerfile
├── Jenkinsfile.build-only
├── nginx.default.conf
├── package.json                  # 如果是 Vue/React 项目
├── src/                          # 源码（Vue/React）
│   └── ...
└── html/                         # 纯静态文件（简化版）
    └── index.html
```

---

## 六、通用 Jenkinsfile（平台 Worker 模式）

### 6.1 Go 项目 Jenkinsfile

```groovy
// Jenkinsfile.build-only - Go 项目（平台 Worker 部署模式）
pipeline {
    agent any
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'BUILD_ID', defaultValue: '', description: '发布单 ID')
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        GOPROXY = 'https://goproxy.cn,direct'
        CGO_ENABLED = '0'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]]
                ])
                script {
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
            }
        }
        
        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                sh 'go test -v -cover ./...'
            }
        }
        
        stage('Build & Push Image') {
            steps {
                sh """
                    docker login -u ${REGISTRY_CREDS_USR} -p ${REGISTRY_CREDS_PSW} \$(echo ${params.IMAGE_REPO} | cut -d'/' -f1)
                    docker build -t ${params.IMAGE_REPO}:${params.IMAGE_TAG} .
                    docker push ${params.IMAGE_REPO}:${params.IMAGE_TAG}
                """
                script {
                    env.IMAGE_DIGEST = sh(
                        script: "docker inspect --format='{{index .RepoDigests 0}}' ${params.IMAGE_REPO}:${params.IMAGE_TAG} | cut -d'@' -f2 || echo ''",
                        returnStdout: true
                    ).trim()
                }
            }
        }
        
        // ⚠️ 没有 Deploy 阶段
    }
    
    post {
        success {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{
                            "build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER},
                            "status": "SUCCESS",
                            "image_repo": "${params.IMAGE_REPO}",
                            "image_tag": "${params.IMAGE_TAG}",
                            "image_digest": "${env.IMAGE_DIGEST ?: ''}",
                            "message": "Go 项目构建成功",
                            "git_commit": "${env.GIT_COMMIT}"
                        }"""
                    )
                }
            }
        }
        failure {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{"build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER}, "status": "FAILURE", "message": "Go 项目构建失败"}"""
                    )
                }
            }
        }
        always { cleanWs() }
    }
}
```

### 6.2 Java Maven 项目 Jenkinsfile

```groovy
// Jenkinsfile.build-only - Java Maven 项目（平台 Worker 部署模式）
pipeline {
    agent any
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'BUILD_ID', defaultValue: '', description: '发布单 ID')
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
    }
    
    tools {
        maven 'Maven-3.9'
        jdk 'JDK-17'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]]
                ])
                script {
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
            }
        }
        
        stage('Build') {
            steps {
                sh 'mvn clean compile -DskipTests -B'
            }
        }
        
        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                sh 'mvn test -B'
            }
            post {
                always {
                    junit allowEmptyResults: true, testResults: '**/target/surefire-reports/*.xml'
                }
            }
        }
        
        stage('Package') {
            steps {
                sh 'mvn package -DskipTests -B'
            }
        }
        
        stage('Build & Push Image') {
            steps {
                sh """
                    docker login -u ${REGISTRY_CREDS_USR} -p ${REGISTRY_CREDS_PSW} \$(echo ${params.IMAGE_REPO} | cut -d'/' -f1)
                    docker build -t ${params.IMAGE_REPO}:${params.IMAGE_TAG} .
                    docker push ${params.IMAGE_REPO}:${params.IMAGE_TAG}
                """
                script {
                    env.IMAGE_DIGEST = sh(
                        script: "docker inspect --format='{{index .RepoDigests 0}}' ${params.IMAGE_REPO}:${params.IMAGE_TAG} | cut -d'@' -f2 || echo ''",
                        returnStdout: true
                    ).trim()
                }
            }
        }
        
        // ⚠️ 没有 Deploy 阶段
    }
    
    post {
        success {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{
                            "build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER},
                            "status": "SUCCESS",
                            "image_repo": "${params.IMAGE_REPO}",
                            "image_tag": "${params.IMAGE_TAG}",
                            "image_digest": "${env.IMAGE_DIGEST ?: ''}",
                            "message": "Java 项目构建成功",
                            "git_commit": "${env.GIT_COMMIT}"
                        }"""
                    )
                }
            }
        }
        failure {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{"build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER}, "status": "FAILURE", "message": "Java 项目构建失败"}"""
                    )
                }
            }
        }
        always { cleanWs() }
    }
}
```

### 6.3 Python 项目 Jenkinsfile

```groovy
// Jenkinsfile.build-only - Python 项目（平台 Worker 部署模式）
pipeline {
    agent any
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'BUILD_ID', defaultValue: '', description: '发布单 ID')
        booleanParam(name: 'SKIP_TESTS', defaultValue: false, description: '跳过测试')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        PIP_INDEX_URL = 'https://pypi.tuna.tsinghua.edu.cn/simple'
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]]
                ])
                script {
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
            }
        }
        
        stage('Lint') {
            steps {
                sh '''
                    pip install flake8 -q
                    flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics || true
                '''
            }
        }
        
        stage('Test') {
            when { expression { return !params.SKIP_TESTS } }
            steps {
                sh '''
                    pip install -r requirements.txt -q
                    pip install pytest pytest-cov -q
                    pytest --cov=. tests/ || true
                '''
            }
        }
        
        stage('Build & Push Image') {
            steps {
                sh """
                    docker login -u ${REGISTRY_CREDS_USR} -p ${REGISTRY_CREDS_PSW} \$(echo ${params.IMAGE_REPO} | cut -d'/' -f1)
                    docker build -t ${params.IMAGE_REPO}:${params.IMAGE_TAG} .
                    docker push ${params.IMAGE_REPO}:${params.IMAGE_TAG}
                """
                script {
                    env.IMAGE_DIGEST = sh(
                        script: "docker inspect --format='{{index .RepoDigests 0}}' ${params.IMAGE_REPO}:${params.IMAGE_TAG} | cut -d'@' -f2 || echo ''",
                        returnStdout: true
                    ).trim()
                }
            }
        }
        
        // ⚠️ 没有 Deploy 阶段
    }
    
    post {
        success {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{
                            "build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER},
                            "status": "SUCCESS",
                            "image_repo": "${params.IMAGE_REPO}",
                            "image_tag": "${params.IMAGE_TAG}",
                            "image_digest": "${env.IMAGE_DIGEST ?: ''}",
                            "message": "Python 项目构建成功",
                            "git_commit": "${env.GIT_COMMIT}"
                        }"""
                    )
                }
            }
        }
        failure {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{"build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER}, "status": "FAILURE", "message": "Python 项目构建失败"}"""
                    )
                }
            }
        }
        always { cleanWs() }
    }
}
```

### 6.4 Nginx 项目 Jenkinsfile

```groovy
// Jenkinsfile.build-only - Nginx 项目（平台 Worker 部署模式）
pipeline {
    agent any
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: '', description: '镜像仓库')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'BUILD_ID', defaultValue: '', description: '发布单 ID')
        booleanParam(name: 'BUILD_FRONTEND', defaultValue: false, description: '是否需要 npm build')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout([
                    $class: 'GitSCM',
                    branches: [[name: "*/${params.GIT_BRANCH}"]],
                    userRemoteConfigs: [[url: params.GIT_REPO]]
                ])
                script {
                    env.GIT_COMMIT = sh(script: 'git rev-parse --short HEAD', returnStdout: true).trim()
                }
            }
        }
        
        stage('Build Frontend') {
            when { expression { return params.BUILD_FRONTEND } }
            steps {
                sh '''
                    npm ci --registry=https://registry.npmmirror.com
                    npm run build
                '''
            }
        }
        
        stage('Build & Push Image') {
            steps {
                sh """
                    docker login -u ${REGISTRY_CREDS_USR} -p ${REGISTRY_CREDS_PSW} \$(echo ${params.IMAGE_REPO} | cut -d'/' -f1)
                    docker build -t ${params.IMAGE_REPO}:${params.IMAGE_TAG} .
                    docker push ${params.IMAGE_REPO}:${params.IMAGE_TAG}
                """
                script {
                    env.IMAGE_DIGEST = sh(
                        script: "docker inspect --format='{{index .RepoDigests 0}}' ${params.IMAGE_REPO}:${params.IMAGE_TAG} | cut -d'@' -f2 || echo ''",
                        returnStdout: true
                    ).trim()
                }
            }
        }
        
        // ⚠️ 没有 Deploy 阶段
    }
    
    post {
        success {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{
                            "build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER},
                            "status": "SUCCESS",
                            "image_repo": "${params.IMAGE_REPO}",
                            "image_tag": "${params.IMAGE_TAG}",
                            "image_digest": "${env.IMAGE_DIGEST ?: ''}",
                            "message": "Nginx 项目构建成功",
                            "git_commit": "${env.GIT_COMMIT}"
                        }"""
                    )
                }
            }
        }
        failure {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{"build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER}, "status": "FAILURE", "message": "Nginx 项目构建失败"}"""
                    )
                }
            }
        }
        always { cleanWs() }
    }
}
```

---

## 七、快速对照表

| 语言 | 基础镜像 | 构建产物 | 运行端口 | 健康检查路径 |
|------|---------|---------|---------|-------------|
| **Go** | `golang:1.21-alpine` / `alpine:3.18` | 二进制文件 | 8080 | `/health` |
| **Java** | `maven:3.9-eclipse-temurin-17` / `eclipse-temurin:17-jre-alpine` | JAR 文件 | 8080 | `/actuator/health` |
| **Python** | `python:3.11-slim` | 源码 + 依赖 | 8000 | `/health` |
| **Nginx** | `node:18-alpine` / `nginx:alpine` | 静态文件 | 80 | `/health` |

---

## 八、验证命令

```bash
# Go 项目本地测试
docker build -t go-demo:test .
docker run -p 8080:8080 go-demo:test
curl http://localhost:8080/health

# Java 项目本地测试
docker build -t java-demo:test .
docker run -p 8080:8080 java-demo:test
curl http://localhost:8080/actuator/health

# Python 项目本地测试
docker build -t python-demo:test .
docker run -p 8000:8000 python-demo:test
curl http://localhost:8000/health

# Nginx 项目本地测试
docker build -t nginx-demo:test .
docker run -p 80:80 nginx-demo:test
curl http://localhost/health
```
