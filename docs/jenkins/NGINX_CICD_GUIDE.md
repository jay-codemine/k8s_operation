# Nginx 项目 CICD 闭环验证 - 完整配置指南

## 目录
1. [概述](#一概述)
2. [第一部分：Jenkins 配置](#二第一部分jenkins-配置)
3. [第二部分：平台配置](#三第二部分平台配置)
4. [第三部分：Nginx 项目准备](#四第三部分nginx-项目准备)
5. [第四部分：执行验证](#五第四部分执行验证)

---

## 一、概述

### 整体流程图
```
┌─────────────────────────────────────────────────────────────────────────────────┐
│  Nginx 发布完整闭环                                                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  │
│  │ 1.平台   │    │ 2.Jenkins│    │ 3.构建   │    │ 4.回调   │    │ 5.部署   │  │
│  │ 触发构建 │───►│ 执行流水 │───►│ 推送镜像 │───►│ 平台    │───►│ K8s     │  │
│  │          │    │ 线       │    │ 到Harbor │    │          │    │          │  │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘    └──────────┘  │
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

安装以下插件：
| 插件名称 | 用途 | 必须 |
|---------|------|-----|
| **HTTP Request Plugin** | 回调平台接口 | ✅ 必须 |
| **Pipeline** | 流水线支持 | ✅ 必须 |
| **Git Plugin** | Git 代码拉取 | ✅ 必须 |
| **Kubernetes Plugin** | K8s Pod Agent | 可选 |
| **Blue Ocean** | 可视化界面 | 可选 |

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

#### 2.2 K8s kubeconfig 凭据
```
Kind:        Secret file
Scope:       Global
File:        (上传你的 kubeconfig 文件)
ID:          k8s-kubeconfig           # 重要！Jenkinsfile 中引用这个 ID
Description: K8s 集群 kubeconfig
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
名称:     nginx-build-deploy
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
│       (参数会由平台触发时传入，这里可以不手动添加)                                  │
│                                                                                 │
│  Build Triggers:                                                                 │
│    ☑ Trigger builds remotely                                                     │
│       Authentication Token: platform-trigger-token                              │
│       (可选，增加安全性)                                                          │
│                                                                                 │
│  Pipeline:                                                                       │
│    Definition: Pipeline script from SCM    ← 选择这个                            │
│    SCM:        Git                                                               │
│    Repository URL: https://github.com/your-org/nginx-demo.git                   │
│    Branch:     */main                                                            │
│    Script Path: Jenkinsfile               ← 仓库根目录的 Jenkinsfile             │
│                                                                                 │
│  或者选择:                                                                        │
│    Definition: Pipeline script            ← 直接在这里写脚本                      │
│    Script:     (粘贴 Jenkinsfile 内容)                                            │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## 三、第二部分：平台配置

### 步骤 1：修改配置文件

编辑 `configs/config.yaml`（从 config.yaml.example 复制）：

```yaml
# Jenkins 配置（平台驱动 CI/CD）
Jenkins:
  URL: "http://jenkins.example.com:8080"    # ← 改成你的 Jenkins 地址
  Username: "admin"                          # ← 改成你的 Jenkins 用户名
  APIToken: "11a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"  # ← 改成步骤3生成的 Token
  TriggerTimeout: 60
  CallbackURL: "http://platform.example.com:8080/api/v1/k8s/cicd/pipeline/callback"  # ← 改成平台地址
```

### 步骤 2：确认数据库表存在

```bash
# 执行 SQL 初始化
mysql -u root -p k8s-platform < docs/sql/k8s-platform.sql
```

### 步骤 3：启动平台服务

```bash
# 编译并启动
go build -o k8s-platform ./cmd/k8soperation/main.go
./k8s-platform
```

---

## 四、第三部分：Nginx 项目准备

### 步骤 1：创建 Git 仓库

在 GitHub/GitLab 创建仓库 `nginx-demo`，目录结构：

```
nginx-demo/
├── Jenkinsfile           # Jenkins 流水线定义
├── Dockerfile            # 镜像构建文件
├── nginx.conf            # Nginx 配置（可选）
└── html/
    └── index.html        # 静态页面
```

### 步骤 2：创建 Dockerfile

```dockerfile
# nginx-demo/Dockerfile
FROM nginx:alpine

# 复制静态文件
COPY html/ /usr/share/nginx/html/

# 可选：自定义 nginx 配置
# COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### 步骤 3：创建 index.html

```html
<!-- nginx-demo/html/index.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Nginx Demo</title>
    <style>
        body { font-family: Arial; text-align: center; padding: 50px; }
        h1 { color: #333; }
        .version { color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <h1>🎉 Nginx Demo - CICD 闭环验证</h1>
    <p>部署时间: <span id="time"></span></p>
    <p class="version">Version: v1.0.0</p>
    <script>document.getElementById('time').textContent = new Date().toLocaleString();</script>
</body>
</html>
```

### 步骤 4：创建 Jenkinsfile

```groovy
// nginx-demo/Jenkinsfile
pipeline {
    agent any
    
    parameters {
        string(name: 'GIT_REPO', defaultValue: '', description: 'Git 仓库地址')
        string(name: 'GIT_BRANCH', defaultValue: 'main', description: 'Git 分支')
        string(name: 'IMAGE_REPO', defaultValue: 'harbor.example.com/library/nginx-demo', description: '镜像仓库')
        string(name: 'IMAGE_TAG', defaultValue: 'latest', description: '镜像标签')
        string(name: 'NAMESPACE', defaultValue: 'default', description: 'K8s 命名空间')
        string(name: 'DEPLOYMENT_NAME', defaultValue: 'nginx-demo', description: 'Deployment 名称')
        string(name: 'CONTAINER_NAME', defaultValue: 'nginx', description: '容器名称')
        string(name: 'PLATFORM_CALLBACK_URL', defaultValue: '', description: '平台回调地址')
        string(name: 'PLATFORM_TOKEN', defaultValue: '', description: '平台 Token')
    }
    
    environment {
        REGISTRY_CREDS = credentials('harbor-registry')
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
            }
        }
        
        stage('Deploy to K8s') {
            when {
                expression { return params.DEPLOYMENT_NAME != '' }
            }
            steps {
                echo "🚀 部署到 K8s: ${params.NAMESPACE}/${params.DEPLOYMENT_NAME}"
                withCredentials([file(credentialsId: 'k8s-kubeconfig', variable: 'KUBECONFIG')]) {
                    sh """
                        kubectl set image deployment/${params.DEPLOYMENT_NAME} \
                            ${params.CONTAINER_NAME}=${params.IMAGE_REPO}:${params.IMAGE_TAG} \
                            -n ${params.NAMESPACE}
                        kubectl rollout status deployment/${params.DEPLOYMENT_NAME} \
                            -n ${params.NAMESPACE} --timeout=300s
                    """
                }
            }
        }
    }
    
    post {
        success {
            script {
                if (params.PLATFORM_CALLBACK_URL) {
                    httpRequest(
                        url: params.PLATFORM_CALLBACK_URL,
                        httpMode: 'POST',
                        contentType: 'APPLICATION_JSON',
                        customHeaders: [[name: 'Authorization', value: "Bearer ${params.PLATFORM_TOKEN}"]],
                        requestBody: """{
                            "job_name": "${env.JOB_NAME}",
                            "build_number": ${env.BUILD_NUMBER},
                            "status": "SUCCESS",
                            "duration": ${currentBuild.duration / 1000},
                            "message": "构建部署成功",
                            "git_commit": "${env.GIT_COMMIT}",
                            "image": "${params.IMAGE_REPO}:${params.IMAGE_TAG}"
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
                        customHeaders: [[name: 'Authorization', value: "Bearer ${params.PLATFORM_TOKEN}"]],
                        requestBody: """{
                            "job_name": "${env.JOB_NAME}",
                            "build_number": ${env.BUILD_NUMBER},
                            "status": "FAILURE",
                            "message": "构建失败"
                        }"""
                    )
                }
            }
        }
    }
}
```

### 步骤 5：提交代码

```bash
cd nginx-demo
git add .
git commit -m "Initial nginx demo project with Jenkinsfile"
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
        image: nginx:alpine    # 初始镜像，后续会被 CICD 更新
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

### 步骤 2：通过平台创建流水线

```powershell
# PowerShell 脚本
$BaseURL = "http://localhost:8080/api/v1/k8s/cicd"
$Token = "your-platform-jwt-token"    # 登录平台后获取
$Headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# 创建流水线
$body = @{
    name = "nginx-demo-pipeline"
    description = "Nginx Demo 部署流水线"
    git_repo = "https://github.com/your-org/nginx-demo.git"   # ← 改成你的仓库
    git_branch = "main"
    jenkins_url = "http://jenkins.example.com:8080"           # ← 改成你的 Jenkins
    jenkins_job = "nginx-build-deploy"                        # ← Jenkins Job 名称
    env_vars = @(
        @{ name = "IMAGE_REPO"; value = "harbor.example.com/library/nginx-demo" }  # ← 改成你的 Harbor
        @{ name = "NAMESPACE"; value = "default" }
        @{ name = "DEPLOYMENT_NAME"; value = "nginx-demo" }
        @{ name = "CONTAINER_NAME"; value = "nginx" }
    )
} | ConvertTo-Json -Depth 3

$resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/create" -Method POST -Headers $Headers -Body $body
$pipelineId = $resp.data.id
Write-Host "✅ 创建流水线成功, ID: $pipelineId"
```

### 步骤 3：触发流水线运行

```powershell
# 触发构建
$runBody = @{
    id = $pipelineId
    branch = "main"
    env_vars = @{
        IMAGE_TAG = "v1.0.$(Get-Date -Format 'yyyyMMddHHmm')"
        PLATFORM_CALLBACK_URL = "http://platform.example.com:8080/api/v1/k8s/cicd/pipeline/callback"
        PLATFORM_TOKEN = $Token
    }
} | ConvertTo-Json -Depth 3

$runResp = Invoke-RestMethod -Uri "$BaseURL/pipeline/run" -Method POST -Headers $Headers -Body $runBody
Write-Host "🚀 触发构建成功"
Write-Host $runResp | ConvertTo-Json
```

### 步骤 4：查看构建状态

```powershell
# 轮询状态
while ($true) {
    $status = Invoke-RestMethod -Uri "$BaseURL/pipeline/status?id=$pipelineId" -Method GET -Headers $Headers
    $runStatus = $status.data.pipeline.last_run_status
    $building = if ($status.data.build_info) { $status.data.build_info.building } else { $false }
    
    Write-Host "状态: $runStatus, Building: $building"
    
    if ($runStatus -in @("success", "failed", "aborted")) {
        Write-Host "构建结束: $runStatus" -ForegroundColor $(if ($runStatus -eq "success") { "Green" } else { "Red" })
        break
    }
    Start-Sleep -Seconds 10
}
```

### 步骤 5：查看构建日志

```powershell
$logs = Invoke-RestMethod -Uri "$BaseURL/pipeline/logs?id=$pipelineId" -Method GET -Headers $Headers
Write-Host "===== 构建日志 ====="
Write-Host $logs.data
```

### 步骤 6：验证 K8s 部署

```powershell
# 检查 Pod 镜像是否更新
kubectl get deployment nginx-demo -n default -o jsonpath='{.spec.template.spec.containers[0].image}'
# 预期输出: harbor.example.com/library/nginx-demo:v1.0.202401201530

# 检查 Pod 状态
kubectl get pods -n default -l app=nginx-demo
# 预期: Running, READY 2/2

# 访问服务
curl http://localhost:30080
# 预期: 看到 index.html 内容
```

---

## 六、配置检查清单

| 检查项 | 配置位置 | 状态 |
|-------|---------|------|
| Jenkins HTTP Request 插件 | Jenkins → Plugins | ☐ |
| Harbor 凭据 (harbor-registry) | Jenkins → Credentials | ☐ |
| K8s kubeconfig 凭据 (k8s-kubeconfig) | Jenkins → Credentials | ☐ |
| Jenkins API Token | Jenkins → 用户 → Configure | ☐ |
| Pipeline Job 创建 | Jenkins → New Item | ☐ |
| 平台 config.yaml Jenkins 配置 | configs/config.yaml | ☐ |
| Nginx 项目 Git 仓库 | GitHub/GitLab | ☐ |
| Jenkinsfile 提交 | Git 仓库根目录 | ☐ |
| K8s Deployment 创建 | kubectl apply | ☐ |

---

## 七、常见问题排查

### Q1: Jenkins 触发失败 "401 Unauthorized"
```
原因: API Token 配置错误
检查: 
1. Jenkins → 用户 → Configure → API Token 是否正确生成
2. 平台 config.yaml 中 Jenkins.APIToken 是否与之匹配
```

### Q2: 回调平台失败 "Connection refused"
```
原因: 平台地址不可达或未启动
检查:
1. 平台服务是否启动: curl http://platform:8080/health
2. Jenkins 能否访问平台: 从 Jenkins 机器 curl 平台地址
3. 防火墙是否开放端口
```

### Q3: 镜像推送失败 "unauthorized"
```
原因: Harbor 凭据配置错误
检查:
1. Jenkins → Credentials → harbor-registry 用户名密码是否正确
2. 在 Jenkins 机器手动测试: docker login harbor.example.com
```

### Q4: K8s 部署失败 "kubectl: command not found"
```
原因: Jenkins Agent 没有 kubectl
解决:
1. 在 Jenkins Agent 安装 kubectl
2. 或使用 Kubernetes Plugin 动态创建 Pod Agent
```

---

## 八、文件路径汇总

| 文件 | 路径 | 说明 |
|------|------|------|
| 平台配置 | `configs/config.yaml` | Jenkins URL/Token 配置 |
| Jenkinsfile 示例 | `docs/jenkins/Jenkinsfile.example` | 通用模板 |
| 本文档 | `docs/jenkins/NGINX_CICD_GUIDE.md` | 配置指南 |
| SQL 初始化 | `docs/sql/k8s-platform.sql` | 数据库表 |
