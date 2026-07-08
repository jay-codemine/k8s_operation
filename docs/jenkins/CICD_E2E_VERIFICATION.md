# CICD 完整闭环验证方案

## 一、架构概述

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         K8sOperation 管理平台                                │
│  ┌────────────────┐                                                         │
│  │  Pipelines.vue │  ← 用户点击"运行"                                       │
│  └───────┬────────┘                                                         │
│          │ POST /api/v1/k8s/cicd/pipeline/run                               │
│          ▼                                                                  │
│  ┌────────────────┐   ┌──────────────────┐                                  │
│  │PipelineService │──▶│  JenkinsClient   │                                  │
│  │ • 参数组装      │   │ • TriggerBuild() │                                  │
│  │ • 状态更新      │   │ • WaitForBuild() │                                  │
│  └────────────────┘   └────────┬─────────┘                                  │
└────────────────────────────────┼────────────────────────────────────────────┘
                                 │ POST /job/{name}/buildWithParameters
                                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         Jenkins Server                                       │
│  ┌─────────────────────────────────────────────────────────────────┐        │
│  │                    Pipeline Job                                  │        │
│  │  ┌──────────┐   ┌─────────────┐   ┌──────────────┐              │        │
│  │  │ Checkout │──▶│ Build Image │──▶│ Deploy to K8s │             │        │
│  │  │ (Git)    │   │ (BuildKit)  │   │ (kubectl)     │             │        │
│  │  └──────────┘   └─────────────┘   └──────┬───────┘              │        │
│  │                                          │                       │        │
│  │  post { success/failure/aborted }        │                       │        │
│  │       │                                  │                       │        │
│  │       ▼ httpRequest 回调                 │                       │        │
│  └───────┼──────────────────────────────────┼───────────────────────┘        │
└──────────┼──────────────────────────────────┼────────────────────────────────┘
           │                                  │
           │ POST /pipeline/callback          │ kubectl set image / rollout
           ▼                                  ▼
┌──────────────────────┐            ┌─────────────────────────────────────────┐
│ 平台更新流水线状态    │            │              K8s Cluster                 │
│ • status: success    │            │  ┌────────────────────────────────────┐  │
│ • last_run_time      │            │  │ Deployment: nginx                  │  │
│ • build_number       │            │  │ Image: harbor.io/app:v1.2.3        │  │
└──────────────────────┘            │  │ Replicas: 3/3 Ready                │  │
                                    │  └────────────────────────────────────┘  │
                                    └─────────────────────────────────────────┘
```

---

## 二、前置条件

### 2.1 Jenkins 配置

**1. 安装必要插件**
```
- HTTP Request Plugin（用于回调平台）
- Kubernetes Plugin（用于动态 Pod Agent）
- Pipeline Plugin
- Git Plugin
```

**2. 创建 Jenkins Credentials**
| ID | 类型 | 用途 |
|----|------|------|
| `harbor-registry` | Username/Password | 镜像仓库认证 |
| `k8s-kubeconfig` | Secret file | K8s 集群 kubeconfig |
| `github-token` | Secret text | Git 仓库访问 |

**3. 创建 Pipeline Job**
```
Job 名称: k8s-platform-deploy
类型: Pipeline
定义: Pipeline script from SCM
SCM: Git
Repository URL: https://github.com/your-org/your-repo.git
Script Path: Jenkinsfile
```

### 2.2 平台配置

修改 `configs/config.yaml`：
```yaml
Jenkins:
  URL: "http://jenkins.example.com:8080"
  Username: "admin"
  APIToken: "your-api-token"  # Jenkins → 用户设置 → API Token
  TriggerTimeout: 60
  CallbackURL: "http://platform.example.com:8080/api/v1/k8s/cicd/pipeline/callback"
```

### 2.3 数据库初始化

```bash
mysql -u root -p k8s-platform < docs/sql/cicd_pipeline.sql
```

---

## 三、闭环验证步骤

### Step 1: 创建流水线（平台侧）

```bash
# 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')

# 创建流水线
curl -X POST http://localhost:8080/api/v1/k8s/cicd/pipeline/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo-app-pipeline",
    "description": "演示应用部署流水线",
    "git_repo": "https://github.com/your-org/demo-app.git",
    "git_branch": "main",
    "jenkins_url": "http://jenkins.example.com:8080",
    "jenkins_job": "k8s-platform-deploy",
    "env_vars": [
      {"name": "IMAGE_REPO", "value": "harbor.io/demo/app"},
      {"name": "NAMESPACE", "value": "default"},
      {"name": "DEPLOYMENT_NAME", "value": "demo-app"},
      {"name": "CONTAINER_NAME", "value": "app"}
    ]
  }'

# 预期响应
# {"code":0,"data":{"pipeline_id":1}}
```

### Step 2: 触发构建（平台→Jenkins）

```bash
# 运行流水线
curl -X POST http://localhost:8080/api/v1/k8s/cicd/pipeline/run \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "env_vars": {
      "IMAGE_TAG": "v1.0.0-'$(date +%Y%m%d%H%M%S)'"
    }
  }'

# 预期响应
# {"code":0,"data":{"message":"流水线启动成功","run_id":1}}
```

**验证点**：
- [ ] Jenkins Job 被触发（查看 Jenkins Blue Ocean）
- [ ] 构建参数正确传入（GIT_REPO, GIT_BRANCH, IMAGE_TAG 等）

### Step 3: 监控构建状态

```bash
# 轮询状态（每5秒）
while true; do
  STATUS=$(curl -s http://localhost:8080/api/v1/k8s/cicd/pipeline/status?id=1 \
    -H "Authorization: Bearer $TOKEN" | jq -r '.data.pipeline.status')
  
  echo "[$(date +%H:%M:%S)] Pipeline Status: $STATUS"
  
  if [[ "$STATUS" == "idle" ]]; then
    echo "构建完成！"
    break
  fi
  
  sleep 5
done
```

### Step 4: 验证 K8s 部署

```bash
# 检查 Deployment 镜像是否更新
kubectl get deployment demo-app -n default -o jsonpath='{.spec.template.spec.containers[0].image}'
# 预期输出: harbor.io/demo/app:v1.0.0-20260225120000

# 检查 Pod 状态
kubectl get pods -n default -l app=demo-app
# 预期输出: 3/3 Running

# 检查 Rollout 状态
kubectl rollout status deployment/demo-app -n default
# 预期输出: deployment "demo-app" successfully rolled out
```

### Step 5: 验证回调闭环

```bash
# 检查流水线最终状态
curl -s http://localhost:8080/api/v1/k8s/cicd/pipeline/detail?id=1 \
  -H "Authorization: Bearer $TOKEN" | jq '.data.pipeline | {status, last_run_status, last_build_number}'

# 预期输出
# {
#   "status": "idle",
#   "last_run_status": "success",
#   "last_build_number": 1
# }
```

---

## 四、自动化验证脚本

### 4.1 完整闭环测试脚本

```powershell
# docs/postman/cicd_e2e_test.ps1
# 端到端闭环测试脚本

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$JenkinsUrl = "http://jenkins.example.com:8080",
    [string]$Username = "admin",
    [string]$Password = "admin123"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " CICD 完整闭环验证测试" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# Step 1: 登录
Write-Host "`n[Step 1] 登录获取 Token..." -ForegroundColor Yellow
$loginBody = @{ username = $Username; password = $Password } | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
if (-not $token) { Write-Host "登录失败！" -ForegroundColor Red; exit 1 }
Write-Host "  Token: $($token.Substring(0,20))..." -ForegroundColor Green

# Step 2: 创建流水线
Write-Host "`n[Step 2] 创建测试流水线..." -ForegroundColor Yellow
$pipelineName = "e2e-test-$(Get-Date -Format 'yyyyMMddHHmmss')"
$createBody = @{
    name = $pipelineName
    description = "E2E测试流水线"
    git_repo = "https://github.com/example/test-app.git"
    git_branch = "main"
    jenkins_url = $JenkinsUrl
    jenkins_job = "k8s-platform-deploy"
} | ConvertTo-Json

$headers = @{ Authorization = "Bearer $token" }
$createResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/k8s/cicd/pipeline/create" -Method POST -Body $createBody -Headers $headers -ContentType "application/json"
$pipelineId = $createResp.data.pipeline_id
if (-not $pipelineId) { Write-Host "创建流水线失败！" -ForegroundColor Red; exit 1 }
Write-Host "  Pipeline ID: $pipelineId" -ForegroundColor Green

# Step 3: 运行流水线
Write-Host "`n[Step 3] 触发流水线运行..." -ForegroundColor Yellow
$runBody = @{ id = $pipelineId } | ConvertTo-Json
$runResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/k8s/cicd/pipeline/run" -Method POST -Body $runBody -Headers $headers -ContentType "application/json"
$runId = $runResp.data.run_id
Write-Host "  Run ID: $runId" -ForegroundColor Green
Write-Host "  Message: $($runResp.data.message)" -ForegroundColor Green

# Step 4: 等待完成
Write-Host "`n[Step 4] 等待构建完成..." -ForegroundColor Yellow
$maxWait = 300  # 5分钟超时
$elapsed = 0
$interval = 10

while ($elapsed -lt $maxWait) {
    $statusResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/k8s/cicd/pipeline/status?id=$pipelineId" -Headers $headers
    $status = $statusResp.data.pipeline.status
    $runStatus = $statusResp.data.pipeline.last_run_status
    
    Write-Host "  [$([math]::Floor($elapsed/60))m $($elapsed%60)s] Status: $status | RunStatus: $runStatus" -ForegroundColor Cyan
    
    if ($status -eq "idle" -and $runStatus -in @("success", "failed", "aborted")) {
        break
    }
    
    Start-Sleep -Seconds $interval
    $elapsed += $interval
}

# Step 5: 验证结果
Write-Host "`n[Step 5] 验证最终结果..." -ForegroundColor Yellow
$detailResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/k8s/cicd/pipeline/detail?id=$pipelineId" -Headers $headers
$finalStatus = $detailResp.data.pipeline.last_run_status
$buildNumber = $detailResp.data.pipeline.last_build_number

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host " 测试结果" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Pipeline Name: $pipelineName"
Write-Host "  Final Status: $finalStatus"
Write-Host "  Build Number: $buildNumber"

if ($finalStatus -eq "success") {
    Write-Host "`n  ✅ 闭环验证成功！" -ForegroundColor Green
} else {
    Write-Host "`n  ❌ 闭环验证失败！" -ForegroundColor Red
}

# Step 6: 清理（可选）
Write-Host "`n[Step 6] 清理测试数据..." -ForegroundColor Yellow
$deleteBody = @{ id = $pipelineId } | ConvertTo-Json
$deleteResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/k8s/cicd/pipeline/delete" -Method POST -Body $deleteBody -Headers $headers -ContentType "application/json"
Write-Host "  清理完成" -ForegroundColor Green
```

---

## 五、关键验证点检查清单

### 5.1 平台侧验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| 流水线创建 | `GET /pipeline/detail?id=1` | `code=0`, `pipeline.name` 存在 |
| Jenkins触发 | 查看Jenkins Blue Ocean | Job被触发，参数正确 |
| 状态更新 | `GET /pipeline/status?id=1` | `status=running` |
| 回调接收 | 查看后端日志 `grep "JenkinsCallback"` | 回调成功记录 |
| 最终状态 | `GET /pipeline/detail?id=1` | `last_run_status=success` |

### 5.2 Jenkins侧验证

| 检查项 | 验证方法 | 预期结果 |
|--------|---------|---------|
| Job触发 | Jenkins Console | `Started by remote host` |
| 参数传递 | Console日志 | 显示正确的 GIT_REPO, IMAGE_TAG |
| 镜像构建 | Console日志 | `buildctl build` 成功 |
| K8s部署 | Console日志 | `kubectl set image` 成功 |
| 回调发送 | Console日志 | `httpRequest` 返回200 |

### 5.3 K8s侧验证

| 检查项 | 验证命令 | 预期结果 |
|--------|---------|---------|
| 镜像更新 | `kubectl get deploy -o jsonpath='{.spec...image}'` | 新镜像标签 |
| Pod状态 | `kubectl get pods` | All Running |
| Rollout | `kubectl rollout status` | successfully rolled out |
| 服务可用 | `curl http://service-endpoint/health` | HTTP 200 |

---

## 六、故障排查

### 6.1 常见问题

**问题1: Jenkins触发失败**
```bash
# 检查Jenkins连通性
curl -u admin:$API_TOKEN http://jenkins:8080/api/json

# 日志关键字
grep "TriggerBuild" /app/logs/app.log
```

**问题2: 回调未收到**
```bash
# 检查Jenkins网络
curl -X POST http://platform:8080/api/v1/k8s/cicd/pipeline/callback \
  -H "Content-Type: application/json" \
  -d '{"job_name":"test","build_number":1,"status":"SUCCESS"}'

# Jenkins日志
cat /var/jenkins_home/jobs/xxx/builds/xxx/log | grep httpRequest
```

**问题3: K8s部署失败**
```bash
# 检查kubeconfig权限
kubectl auth can-i update deployments -n default

# 查看Events
kubectl describe deployment demo-app -n default
```

---

## 七、监控指标

| 指标名 | 说明 | 告警阈值 |
|--------|------|---------|
| `cicd_pipeline_run_total` | 流水线运行总数 | - |
| `cicd_pipeline_run_duration_seconds` | 运行时长 | >600s |
| `cicd_pipeline_run_status{status="failed"}` | 失败次数 | >3/hour |
| `cicd_jenkins_callback_latency_ms` | 回调延迟 | >5000ms |
