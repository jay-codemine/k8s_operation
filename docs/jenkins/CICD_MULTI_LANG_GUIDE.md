# CICD 多语言闭环验证指南

## 一、概述

本文档说明如何使用平台验证完整的 Jenkins CICD 闭环，支持以下语言/项目类型：

| 语言 | Jenkinsfile 模板 | Dockerfile 示例 |
|------|-----------------|-----------------|
| Java (Maven) | `Jenkinsfile.java-maven.groovy` | 见下文 |
| Go | `Jenkinsfile.golang.groovy` | 见下文 |
| Python | `Jenkinsfile.python.groovy` | 见下文 |
| Nginx/静态 | `Jenkinsfile.example` | 见下文 |

---

## 二、架构说明

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  平台参数流向                                                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  平台 POST /api/v1/k8s/cicd/pipeline/run                                        │
│    │                                                                            │
│    ├── git_repo ──────────────┐                                                 │
│    ├── git_branch ────────────┤                                                 │
│    ├── env_vars ──────────────┼────► Jenkins Job Parameters                    │
│    │   ├── IMAGE_REPO         │      (触发构建时传入)                            │
│    │   ├── IMAGE_TAG          │                                                 │
│    │   └── ...                │                                                 │
│    └── jenkins_job ───────────┘                                                 │
│                                                                                 │
│  Jenkins 执行 Jenkinsfile                                                       │
│    │                                                                            │
│    ├── Checkout(git_repo)                                                       │
│    ├── Build/Test                                                               │
│    ├── BuildImage(IMAGE_REPO:IMAGE_TAG)                                         │
│    ├── Deploy(kubectl set image)                                                │
│    └── Callback(PLATFORM_CALLBACK_URL) ──────► 平台接收更新状态                  │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 三、验证 Java 项目完整闭环

### 3.1 准备 Java 项目仓库

项目结构：
```
my-java-app/
├── Jenkinsfile              # 从 docs/jenkins/Jenkinsfile.java-maven.groovy 复制
├── Dockerfile               # 见下文
├── pom.xml
└── src/
    └── main/java/...
```

**Dockerfile（Java）：**
```dockerfile
# 多阶段构建
FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY target/*.jar app.jar
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

### 3.2 在 Jenkins 创建 Job

1. 新建 Pipeline Job，命名：`java-app-build`
2. 配置：Pipeline script from SCM
3. SCM：Git，Repository URL：`https://github.com/your-org/my-java-app.git`
4. Script Path：`Jenkinsfile`

### 3.3 在平台创建流水线

```powershell
# PowerShell 验证脚本
$BaseURL = "http://localhost:8080/api/v1/k8s/cicd"
$Token = "your-jwt-token"
$Headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# 创建 Java 流水线
$body = @{
    name = "java-app-deploy"
    description = "Java Spring Boot 应用部署"
    git_repo = "https://github.com/your-org/my-java-app.git"
    git_branch = "main"
    jenkins_url = "http://jenkins.example.com:8080"
    jenkins_job = "java-app-build"
    env_vars = @(
        @{ name = "IMAGE_REPO"; value = "harbor.example.com/library/java-app" }
        @{ name = "NAMESPACE"; value = "production" }
        @{ name = "DEPLOYMENT_NAME"; value = "java-app" }
        @{ name = "CONTAINER_NAME"; value = "java-app" }
    )
    deploy_config = @{
        namespace = "production"
        workload_kind = "Deployment"
        workload_name = "java-app"
        container_name = "java-app"
    }
} | ConvertTo-Json -Depth 3

$resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/create" -Method POST -Headers $Headers -Body $body
$pipelineId = $resp.data.id
Write-Host "创建流水线成功, ID: $pipelineId"
```

### 3.4 触发构建并验证闭环

```powershell
# 触发构建
$runBody = @{
    id = $pipelineId
    branch = "main"
    env_vars = @{
        IMAGE_TAG = "v1.0.$(Get-Date -Format 'yyyyMMddHHmm')"
        SKIP_TESTS = "false"
        PLATFORM_CALLBACK_URL = "http://platform.example.com:8080/api/v1/k8s/cicd/pipeline/callback"
        PLATFORM_TOKEN = $Token
    }
} | ConvertTo-Json -Depth 3

$runResp = Invoke-RestMethod -Uri "$BaseURL/pipeline/run" -Method POST -Headers $Headers -Body $runBody
Write-Host "触发构建: $($runResp | ConvertTo-Json)"

# 轮询状态
$maxWait = 600  # Java 构建较慢，等 10 分钟
$waited = 0
while ($waited -lt $maxWait) {
    Start-Sleep -Seconds 15
    $waited += 15
    
    $status = Invoke-RestMethod -Uri "$BaseURL/pipeline/status?id=$pipelineId" -Method GET -Headers $Headers
    $runStatus = $status.data.pipeline.last_run_status
    $building = $status.data.build_info.building
    
    Write-Host "[$waited s] Status: $runStatus, Building: $building"
    
    if ($runStatus -in @("success", "failed", "aborted")) {
        if ($runStatus -eq "success") {
            Write-Host "✅ 构建成功！" -ForegroundColor Green
        } else {
            Write-Host "❌ 构建失败！" -ForegroundColor Red
        }
        break
    }
}

# 查看日志
$logs = Invoke-RestMethod -Uri "$BaseURL/pipeline/logs?id=$pipelineId" -Method GET -Headers $Headers
Write-Host "`n===== 构建日志 ====="
Write-Host $logs.data
```

---

## 四、验证检查清单

| 验证点 | 验证方法 | 预期结果 |
|-------|---------|---------|
| 代码拉取 | Jenkins 日志 `Checking out` | 成功检出指定分支 |
| Maven 编译 | Jenkins 日志 `BUILD SUCCESS` | 编译通过 |
| 单元测试 | Jenkins 日志 `Tests run:` | 测试通过 |
| 镜像构建 | Jenkins 日志 `buildctl build` | 构建成功 |
| 镜像推送 | Harbor 仓库检查 | 镜像存在 |
| 回调平台 | 平台日志 / API 状态 | 状态更新为 success |
| K8s 部署 | `kubectl get pods` | Pod 镜像已更新 |
| Rollout 完成 | `kubectl rollout status` | 滚动更新完成 |

### 验证回调是否成功

```powershell
# 检查平台是否收到回调
$history = Invoke-RestMethod -Uri "$BaseURL/pipeline/history?id=$pipelineId&page=1&page_size=1" -Method GET -Headers $Headers
$latestRun = $history.data.list[0]

Write-Host "最近一次运行:"
Write-Host "  状态: $($latestRun.status)"
Write-Host "  构建号: $($latestRun.build_number)"
Write-Host "  耗时: $($latestRun.duration_sec) 秒"
Write-Host "  触发类型: $($latestRun.trigger_type)"
```

### 验证 K8s 部署

```powershell
# 验证 Pod 镜像是否更新
kubectl get deployment java-app -n production -o jsonpath='{.spec.template.spec.containers[0].image}'
# 预期输出: harbor.example.com/library/java-app:v1.0.202401201530

# 验证 Pod 状态
kubectl get pods -n production -l app=java-app -o wide
# 预期: Running, READY 1/1
```

---

## 五、Dockerfile 示例

### Nginx/静态站点
```dockerfile
FROM nginx:alpine
COPY dist/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
```

### Go 项目
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/server ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

### Python (FastAPI)
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple
COPY . .
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Java (Spring Boot)
```dockerfile
FROM eclipse-temurin:17-jre-alpine
WORKDIR /app
COPY target/*.jar app.jar
EXPOSE 8080
ENV JAVA_OPTS="-Xms256m -Xmx512m"
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

---

## 六、常见问题

### Q1: Jenkinsfile 放哪里？
**A:** 两种方式：
1. 放在项目仓库根目录（推荐），Jenkins Job 配置 `Pipeline script from SCM`
2. 直接在 Jenkins Job 中配置 `Pipeline script`

### Q2: git_repo 是什么？
**A:** 你的项目源码仓库地址，Jenkins 会 checkout 这个仓库的代码进行构建。

### Q3: 回调地址怎么配置？
**A:** 平台触发 Jenkins 时会自动传入 `PLATFORM_CALLBACK_URL` 参数，Jenkinsfile 中的 `post { success/failure }` 会调用这个地址。

### Q4: 如何调试？
1. 先在 Jenkins 手动触发构建，确保 Jenkinsfile 正确
2. 查看 Jenkins 控制台日志
3. 通过平台 `/pipeline/logs` 接口获取日志

---

## 七、文件清单

| 文件路径 | 说明 |
|---------|------|
| `docs/jenkins/Jenkinsfile.example` | 通用示例（Nginx） |
| `docs/jenkins/Jenkinsfile.java-maven.groovy` | Java Maven 项目 |
| `docs/jenkins/Jenkinsfile.golang.groovy` | Go 项目 |
| `docs/jenkins/Jenkinsfile.python.groovy` | Python 项目 |
| `docs/jenkins/CICD_MULTI_LANG_GUIDE.md` | 本文档 |
