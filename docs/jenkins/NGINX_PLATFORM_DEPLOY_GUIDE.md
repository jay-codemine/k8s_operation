# Nginx 项目 CICD 闭环验证 - 平台 Worker 部署模式

## 目录
1. [概述](#一概述)
2. [第一部分：Jenkins 配置](#二第一部分jenkins-配置)
3. [第二部分：平台配置](#三第二部分平台配置)
4. [第三部分：Nginx 项目准备](#四第三部分nginx-项目准备)
5. [第四部分：执行验证](#五第四部分执行验证)
6. [架构对比](#六架构对比)

---

## 一、概述

### 模式说明

本文档采用 **平台 Worker 部署模式**：
- **Jenkins 职责**：CI（代码拉取 → 镜像构建 → 推送 Harbor → 回调平台）
- **平台 Worker 职责**：CD（接收回调 → 创建发布单 → 部署到 K8s）
- **Jenkins 不需要 kubeconfig 凭证**

### 整体流程图
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  Nginx 发布完整闭环（平台 Worker 部署模式）                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │ 1.平台   │    │ 2.Jenkins│    │ 3.构建   │    │ 4.Jenkins│    │ 5.平台   │  │
│  │ 触发构建 │───►│ 拉代码   │───►│ 推送镜像 │───►│ 回调平台 │───►│ Worker   │  │
│  │          │    │ 构建镜像 │    │ 到Harbor │    │ (不部署) │    │ 部署K8s  │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
│                                                                                 │
│  关键区别：                                                                      │
│  - Jenkins 不执行 kubectl，不需要 kubeconfig 凭证                                │
│  - 平台 Worker 通过 K8s API 直接 Patch Deployment                               │
│  - 平台记录 PrevImage，支持精确回滚                                              │
│                                                                                 │
│  涉及组件：                                                                      │
│  - Jenkins (http://jenkins.example.com:8080)                                    │
│  - Harbor  (http://harbor.example.com)                                          │
│  - K8s 平台 (http://platform.example.com:8080)                                  │
│  - Git 仓库 (https://github.com/your-org/nginx-demo.git)                        │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 二、第一部分：Jenkins 配置

### 步骤 1：安装必要插件

登录 Jenkins → **Manage Jenkins** → **Plugins** → **Available plugins**

| 插件名称 | 用途 | 必须 |
|---------|------|-----|
| **HTTP Request Plugin** | 回调平台接口 | ✅ 必须 |
| **Pipeline** | 流水线支持 | ✅ 必须 |
| **Git Plugin** | Git 代码拉取 | ✅ 必须 |
| ~~Kubernetes Plugin~~ | ~~K8s Pod Agent~~ | ❌ 不需要 |

```
安装路径: Jenkins → Manage Jenkins → Plugins → Available plugins
搜索: "HTTP Request" → 勾选 → Install
```

### 步骤 2：创建 Jenkins 凭据

路径：**Manage Jenkins** → **Credentials** → **System** → **Global credentials** → **Add Credentials**

#### 2.1 Harbor 镜像仓库凭据
```
Kind:        Username with password
Scope:       Global
Username:    admin                    # Harbor 用户名
Password:    Harbor12345              # Harbor 密码
ID:          harbor-registry          # 重要！Jenkinsfile 中引用这个 ID
Description: Harbor 镜像仓库认证
```

#### 2.2 K8s kubeconfig 凭据 - ⚠️ 本模式不需要
```
本模式下 Jenkins 不执行 kubectl 部署，无需配置 k8s-kubeconfig 凭据。
部署由平台 Worker 完成，平台已配置好 K8s 集群连接。
```

### 步骤 3：生成 API Token

路径：点击右上角用户名 → **Configure** → **API Token** → **Add new Token**

```
Token名称: platform-trigger
点击 Generate → 复制生成的 Token（只显示一次！）

示例 Token: 11a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
```

**重要：把这个 Token 记下来，后面配置平台要用！**

### 步骤 4：创建 Pipeline Job

路径：**Dashboard** → **New Item**

```
名称:     nginx-build-only
类型:     Pipeline（选择这个）
点击 OK
```

配置页面：
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  Pipeline Job 配置                                                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  General:                                                                       │
│    ☑ This project is parameterized                                              │
│       (参数会由平台触发时传入)                                                    │
│                                                                                 │
│  Pipeline:                                                                       │
│    Definition: Pipeline script from SCM                                          │
│    SCM:        Git                                                               │
│    Repository URL: https://github.com/your-org/nginx-demo.git                   │
│    Branch:     */main                                                            │
│    Script Path: Jenkinsfile.build-only     ← 注意：使用纯构建版本               │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 三、第二部分：平台配置

### 步骤 1：修改配置文件

编辑 `configs/config.yaml`：

```yaml
# Jenkins 配置（仅用于触发 CI 构建）
Jenkins:
  URL: "http://jenkins.example.com:8080"    # ← 改成你的 Jenkins 地址
  Username: "admin"                          # ← 改成你的 Jenkins 用户名
  APIToken: "11a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"  # ← 改成步骤3生成的 Token
  TriggerTimeout: 60
  CallbackURL: "http://platform.example.com:8080/api/v1/k8s/cicd/callback/build"

# K8s 集群配置（Worker 部署使用）
# 平台已通过 k8s_cluster 表管理多集群凭证，无需在 Jenkins 配置 kubeconfig
```

### 步骤 2：确认 K8s 集群已注册

平台需要先注册 K8s 集群，Worker 才能部署：

```sql
-- 查看已注册集群
SELECT id, name, api_server, status FROM k8s_cluster WHERE is_del = 0;
```

如果没有集群，通过平台 API 注册：
```powershell
$body = @{
    name = "prod-cluster"
    api_server = "https://kubernetes.default.svc:6443"
    kubeconfig = (Get-Content -Path "~/.kube/config" -Raw)
} | ConvertTo-Json

Invoke-RestMethod -Uri "$BaseURL/cluster/create" -Method POST -Headers $Headers -Body $body
```

### 步骤 3：启动平台服务

```bash
# 确保 Redis 已启动（Worker 使用 Redis Stream 消费任务）
redis-server

# 编译并启动平台
go build -o k8s-platform ./cmd/k8soperation/main.go
./k8s-platform
```

**验证 Worker 启动日志：**
```
INFO    cicd worker started    {"consumer": "hostname-1234567890", "concurrency": 3}
```

---

## 四、第三部分：Nginx 项目准备

### 步骤 1：创建 Git 仓库

在 GitHub/GitLab 创建仓库 `nginx-demo`，目录结构：

```
nginx-demo/
├── Jenkinsfile.build-only    # Jenkins 流水线（仅构建，不部署）
├── Dockerfile                # 镜像构建文件
└── html/
    └── index.html            # 静态页面
```

### 步骤 2：创建 Dockerfile

```dockerfile
# nginx-demo/Dockerfile
FROM nginx:alpine

# 复制静态文件
COPY html/ /usr/share/nginx/html/

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### 步骤 3：创建 index.html

```html
<!-- nginx-demo/html/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Nginx Demo - Platform Worker Mode</title>
    <style>
        body { font-family: Arial; text-align: center; padding: 50px; }
        h1 { color: #333; }
        .mode { color: #4CAF50; font-weight: bold; }
        .version { color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <h1>🎉 Nginx Demo</h1>
    <p class="mode">部署模式: 平台 Worker 模式</p>
    <p>部署时间: <span id="time"></span></p>
    <p class="version">Version: v1.0.0</p>
    <script>document.getElementById('time').textContent = new Date().toLocaleString();</script>
</body>
</html>
```

### 步骤 4：创建 Jenkinsfile（仅构建版本）

```groovy
// nginx-demo/Jenkinsfile.build-only
// ============================================================================
// 平台 Worker 部署模式 - Jenkins 仅负责 CI（构建+推送镜像）
// 部署由平台 Worker 完成，Jenkins 不需要 kubeconfig
// ============================================================================

pipeline {
    agent any
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: 'harbor.example.com/library/nginx-demo', description: '镜像仓库')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'BUILD_ID', defaultValue: '', description: '发布单关联的构建ID')
        // 注意：没有 NAMESPACE、DEPLOYMENT_NAME 等部署参数
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
        // 注意：没有 KUBECONFIG_CREDS，不需要 K8s 凭证
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo "📦 拉取代码: ${params.GIT_REPO}"
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
        
        stage('Build Image') {
            steps {
                echo "🐳 构建镜像: ${params.IMAGE_REPO}:${params.IMAGE_TAG}"
                sh """
                    docker login -u ${REGISTRY_CREDS_USR} -p ${REGISTRY_CREDS_PSW} harbor.example.com
                    docker build -t ${params.IMAGE_REPO}:${params.IMAGE_TAG} .
                    docker push ${params.IMAGE_REPO}:${params.IMAGE_TAG}
                """
                script {
                    // 获取镜像 digest（用于精确匹配）
                    env.IMAGE_DIGEST = sh(
                        script: "docker inspect --format='{{index .RepoDigests 0}}' ${params.IMAGE_REPO}:${params.IMAGE_TAG} | cut -d'@' -f2",
                        returnStdout: true
                    ).trim()
                }
            }
        }
        
        // ⚠️ 注意：没有 Deploy 阶段！部署由平台 Worker 完成
    }
    
    post {
        success {
            script {
                echo "✅ 构建成功，回调平台触发部署..."
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
                            "message": "镜像构建推送成功，等待平台部署",
                            "git_commit": "${env.GIT_COMMIT}"
                        }"""
                    )
                }
            }
        }
        failure {
            script {
                echo "❌ 构建失败，通知平台..."
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        requestBody: """{
                            "build_id": ${params.BUILD_ID ?: env.BUILD_NUMBER},
                            "status": "FAILURE",
                            "message": "镜像构建失败"
                        }"""
                    )
                }
            }
        }
        always {
            cleanWs()
        }
    }
}
```

### 步骤 5：提交代码

```bash
cd nginx-demo
git add .
git commit -m "Add build-only Jenkinsfile for platform worker mode"
git push origin main
```

---

## 五、第四部分：执行验证

### 步骤 1：在 K8s 创建 Nginx Deployment（首次部署）

```yaml
# nginx-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-demo
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx-demo
  template:
    metadata:
      labels:
        app: nginx-demo
    spec:
      containers:
      - name: nginx
        image: nginx:alpine    # 初始镜像，后续会被平台 Worker 更新
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-demo
  namespace: default
spec:
  type: NodePort
  selector:
    app: nginx-demo
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30080
```

```bash
kubectl apply -f nginx-deployment.yaml
```

### 步骤 2：完整验证脚本（PowerShell）

```powershell
# ============================================================================
# nginx_platform_deploy_verify.ps1
# 平台 Worker 部署模式完整验证脚本
# ============================================================================

$BaseURL = "http://localhost:8080/api/v1/k8s"
$CicdURL = "$BaseURL/cicd"
$Token = "your-platform-jwt-token"    # 登录平台后获取
$Headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# 配置参数
$ClusterID = 1                                              # 已注册的集群 ID
$GitRepo = "https://github.com/your-org/nginx-demo.git"    # Git 仓库
$ImageRepo = "harbor.example.com/library/nginx-demo"        # Harbor 镜像仓库
$ImageTag = "v1.0.$(Get-Date -Format 'yyyyMMddHHmm')"       # 镜像标签

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Nginx 平台 Worker 部署模式验证" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

# ==================== 步骤 1：创建流水线 ====================
Write-Host "`n[步骤 1] 创建流水线..." -ForegroundColor Yellow

$pipelineBody = @{
    name = "nginx-build-only-pipeline"
    description = "Nginx 构建流水线（平台 Worker 部署）"
    git_repo = $GitRepo
    git_branch = "main"
    jenkins_job = "nginx-build-only"
    env_vars = @(
        @{ name = "IMAGE_REPO"; value = $ImageRepo }
    )
} | ConvertTo-Json -Depth 3

try {
    $pipelineResp = Invoke-RestMethod -Uri "$CicdURL/pipeline/create" -Method POST -Headers $Headers -Body $pipelineBody
    $PipelineID = $pipelineResp.data.id
    Write-Host "✅ 创建流水线成功, ID: $PipelineID" -ForegroundColor Green
} catch {
    Write-Host "❌ 创建流水线失败: $_" -ForegroundColor Red
    exit 1
}

# ==================== 步骤 2：创建发布单 ====================
Write-Host "`n[步骤 2] 创建发布单..." -ForegroundColor Yellow

$releaseBody = @{
    app_name = "nginx-demo"
    namespace = "default"
    workload_kind = "Deployment"
    workload_name = "nginx-demo"
    container_name = "nginx"
    strategy = "rolling"
    timeout_sec = 300
    concurrency = 1
    image_repo = $ImageRepo
    image_tag = $ImageTag
    cluster_ids = @($ClusterID)
} | ConvertTo-Json -Depth 3

try {
    $releaseResp = Invoke-RestMethod -Uri "$CicdURL/release/create" -Method POST -Headers $Headers -Body $releaseBody
    $ReleaseID = $releaseResp.data.id
    Write-Host "✅ 创建发布单成功, ID: $ReleaseID" -ForegroundColor Green
} catch {
    Write-Host "❌ 创建发布单失败: $_" -ForegroundColor Red
    exit 1
}

# ==================== 步骤 3：触发 Jenkins 构建 ====================
Write-Host "`n[步骤 3] 触发 Jenkins 构建..." -ForegroundColor Yellow

$runBody = @{
    id = $PipelineID
    branch = "main"
    env_vars = @{
        IMAGE_TAG = $ImageTag
        BUILD_ID = $ReleaseID
        PLATFORM_CALLBACK_URL = "http://platform.example.com:8080/api/v1/k8s/cicd/callback/build"
    }
} | ConvertTo-Json -Depth 3

try {
    $runResp = Invoke-RestMethod -Uri "$CicdURL/pipeline/run" -Method POST -Headers $Headers -Body $runBody
    Write-Host "✅ 触发构建成功" -ForegroundColor Green
} catch {
    Write-Host "❌ 触发构建失败: $_" -ForegroundColor Red
    exit 1
}

# ==================== 步骤 4：等待构建完成 ====================
Write-Host "`n[步骤 4] 等待 Jenkins 构建完成..." -ForegroundColor Yellow

$maxWait = 300  # 最大等待 5 分钟
$waited = 0
$interval = 10

while ($waited -lt $maxWait) {
    Start-Sleep -Seconds $interval
    $waited += $interval
    
    try {
        $statusResp = Invoke-RestMethod -Uri "$CicdURL/pipeline/status?id=$PipelineID" -Method GET -Headers $Headers
        $buildStatus = $statusResp.data.pipeline.last_run_status
        $building = if ($statusResp.data.build_info) { $statusResp.data.build_info.building } else { $false }
        
        Write-Host "  构建状态: $buildStatus, Building: $building ($waited/$maxWait 秒)"
        
        if ($buildStatus -in @("success", "failed", "aborted")) {
            if ($buildStatus -eq "success") {
                Write-Host "✅ Jenkins 构建完成" -ForegroundColor Green
            } else {
                Write-Host "❌ Jenkins 构建失败: $buildStatus" -ForegroundColor Red
                exit 1
            }
            break
        }
    } catch {
        Write-Host "  查询状态失败: $_" -ForegroundColor Gray
    }
}

# ==================== 步骤 5：等待平台 Worker 部署 ====================
Write-Host "`n[步骤 5] 等待平台 Worker 部署完成..." -ForegroundColor Yellow

$waited = 0
while ($waited -lt $maxWait) {
    Start-Sleep -Seconds $interval
    $waited += $interval
    
    try {
        $releaseDetail = Invoke-RestMethod -Uri "$CicdURL/release/detail?id=$ReleaseID" -Method GET -Headers $Headers
        $releaseStatus = $releaseDetail.data.release.status
        
        Write-Host "  发布单状态: $releaseStatus ($waited/$maxWait 秒)"
        
        if ($releaseStatus -in @("Succeeded", "Failed", "Canceled")) {
            if ($releaseStatus -eq "Succeeded") {
                Write-Host "✅ 平台 Worker 部署成功！" -ForegroundColor Green
            } else {
                Write-Host "❌ 部署失败: $releaseStatus" -ForegroundColor Red
                # 查看任务详情
                $tasks = $releaseDetail.data.tasks
                foreach ($task in $tasks) {
                    Write-Host "  Task[$($task.id)]: $($task.status) - $($task.message)" -ForegroundColor Gray
                }
                exit 1
            }
            break
        }
    } catch {
        Write-Host "  查询发布单失败: $_" -ForegroundColor Gray
    }
}

# ==================== 步骤 6：验证 K8s 部署结果 ====================
Write-Host "`n[步骤 6] 验证 K8s 部署结果..." -ForegroundColor Yellow

# 检查 Deployment 镜像
$currentImage = kubectl get deployment nginx-demo -n default -o jsonpath='{.spec.template.spec.containers[0].image}'
Write-Host "  当前镜像: $currentImage"
Write-Host "  目标镜像: ${ImageRepo}:${ImageTag}"

if ($currentImage -like "*$ImageTag*") {
    Write-Host "✅ 镜像更新验证通过！" -ForegroundColor Green
} else {
    Write-Host "⚠️ 镜像可能未更新，请检查" -ForegroundColor Yellow
}

# 检查 Pod 状态
Write-Host "`n  Pod 状态:"
kubectl get pods -n default -l app=nginx-demo -o wide

# 访问服务
Write-Host "`n  访问服务:"
try {
    $response = Invoke-WebRequest -Uri "http://localhost:30080" -UseBasicParsing -TimeoutSec 5
    Write-Host "✅ 服务访问成功，状态码: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "⚠️ 服务访问失败: $_" -ForegroundColor Yellow
}

# ==================== 完成 ====================
Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "  验证完成！" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host @"

关键验证点：
1. Jenkins Job 中没有 Deploy 阶段 ✓
2. Jenkins 不需要 kubeconfig 凭证 ✓
3. 平台 Worker 执行 K8s 部署 ✓
4. 发布单记录 PrevImage，支持回滚 ✓

回滚测试（可选）：
`$cancelBody = @{ id = $ReleaseID } | ConvertTo-Json
Invoke-RestMethod -Uri "$CicdURL/release/cancel" -Method POST -Headers `$Headers -Body `$cancelBody
"@
```

### 步骤 3：运行验证

```powershell
# 执行验证脚本
.\docs\jenkins\nginx_platform_deploy_verify.ps1
```

---

## 六、架构对比

### 模式对比表

| 对比项 | Jenkins 全流程模式 | 平台 Worker 模式（本文档） |
|--------|-------------------|-------------------------|
| **Jenkins 需要 kubeconfig** | ✅ 需要 | ❌ **不需要** |
| **谁执行 K8s 部署** | Jenkins (kubectl) | **平台 Worker (K8s API)** |
| **多集群支持** | 困难 | **容易（平台统一管理）** |
| **部署审计** | Jenkins 日志 | **平台数据库完整记录** |
| **回滚能力** | K8s 原生 | **平台记录 PrevImage，精确回滚** |
| **安全性** | kubeconfig 分散存储 | **凭证集中管理，最小权限** |

### 回调接口说明

Jenkins 构建完成后回调平台：

**接口：** `POST /api/v1/k8s/cicd/callback/build`

**请求体：**
```json
{
    "build_id": 123,
    "status": "SUCCESS",
    "image_repo": "harbor.example.com/library/nginx-demo",
    "image_tag": "v1.0.202401201530",
    "image_digest": "sha256:abc123...",
    "message": "镜像构建推送成功"
}
```

**平台处理流程：**
```
1. 根据 build_id 查找关联的发布单
2. 更新发布单的镜像信息
3. 将部署任务入队到 Redis Stream
4. Worker 消费任务，执行 K8s Patch Deployment
5. 记录 PrevImage（用于回滚）
6. 更新任务和发布单状态
```

---

## 七、配置检查清单

| 检查项 | 配置位置 | 状态 |
|-------|---------|------|
| Jenkins HTTP Request 插件 | Jenkins → Plugins | ☐ |
| Harbor 凭据 (harbor-registry) | Jenkins → Credentials | ☐ |
| ~~K8s kubeconfig 凭据~~ | ~~Jenkins → Credentials~~ | **不需要** |
| Jenkins API Token | Jenkins → 用户 → Configure | ☐ |
| Pipeline Job 创建 | Jenkins → New Item | ☐ |
| 平台 config.yaml Jenkins 配置 | configs/config.yaml | ☐ |
| 平台 K8s 集群注册 | k8s_cluster 表 | ☐ |
| 平台 Worker 启动 | 启动日志确认 | ☐ |
| Nginx 项目 Git 仓库 | GitHub/GitLab | ☐ |
| Jenkinsfile.build-only 提交 | Git 仓库根目录 | ☐ |
| K8s Deployment 创建 | kubectl apply | ☐ |

---

## 八、常见问题排查

### Q1: 构建成功但发布单状态一直是 Queued
```
原因: Worker 未启动或 Redis 连接失败
检查:
1. 平台启动日志是否有 "cicd worker started"
2. Redis 服务是否正常: redis-cli ping
3. 查看 Worker 日志: grep "cicd" logs/app.log
```

### Q2: 发布单状态变成 Failed，提示"获取K8s客户端失败"
```
原因: 平台未注册 K8s 集群或集群凭证失效
检查:
1. 查询已注册集群: SELECT * FROM k8s_cluster WHERE is_del = 0;
2. 测试集群连接: kubectl --kubeconfig=/path/to/config get nodes
3. 确认 cluster_ids 参数正确
```

### Q3: Jenkins 回调平台失败
```
原因: 平台地址不可达或回调 URL 配置错误
检查:
1. 从 Jenkins 机器 curl 平台地址
2. 确认 PLATFORM_CALLBACK_URL 参数正确
3. 检查平台日志是否收到回调请求
```

---

## 九、文件路径汇总

| 文件 | 路径 | 说明 |
|------|------|------|
| 本文档 | `docs/jenkins/NGINX_PLATFORM_DEPLOY_GUIDE.md` | 平台 Worker 部署模式指南 |
| 验证脚本 | `docs/jenkins/nginx_platform_deploy_verify.ps1` | PowerShell 验证脚本 |
| 平台配置 | `configs/config.yaml` | Jenkins URL/Token 配置 |
| Worker 代码 | `internal/app/worker/cicd_worker.go` | 部署任务消费者 |
| Executor 代码 | `internal/app/services/cicd_executor.go` | K8s 部署执行器 |
| 发布单服务 | `internal/app/services/cicd_release.go` | 发布单业务逻辑 |
