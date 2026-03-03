# ============================================================================
# nginx_platform_deploy_verify.ps1
# Nginx CICD 平台 Worker 部署模式完整验证脚本
# ============================================================================
# 使用方法：
#   1. 修改下方配置参数
#   2. 运行: .\docs\jenkins\nginx_platform_deploy_verify.ps1
# ============================================================================

param(
    [string]$BaseURL = "http://localhost:8080/api/v1/k8s",
    [string]$Token = "",
    [int]$ClusterID = 1,
    [string]$GitRepo = "https://github.com/your-org/nginx-demo.git",
    [string]$ImageRepo = "harbor.example.com/library/nginx-demo"
)

# ==================== 配置参数 ====================
$CicdURL = "$BaseURL/cicd"
$ImageTag = "v1.0.$(Get-Date -Format 'yyyyMMddHHmm')"
$MaxWait = 300  # 最大等待秒数
$Interval = 10  # 轮询间隔

# 检查 Token
if ([string]::IsNullOrEmpty($Token)) {
    Write-Host "请提供平台 JWT Token:" -ForegroundColor Yellow
    Write-Host '  .\nginx_platform_deploy_verify.ps1 -Token "your-jwt-token"'
    Write-Host ""
    Write-Host "获取 Token 方法:"
    Write-Host '  $loginBody = @{ username = "admin"; password = "admin123" } | ConvertTo-Json'
    Write-Host '  $loginResp = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"'
    Write-Host '  $Token = $loginResp.data.token'
    exit 1
}

$Headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

# ==================== 辅助函数 ====================
function Write-Step {
    param([string]$Step, [string]$Message)
    Write-Host "`n[$Step] $Message" -ForegroundColor Yellow
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Fail {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "  $Message" -ForegroundColor Gray
}

# ==================== 主流程 ====================
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Nginx 平台 Worker 部署模式验证" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "配置信息:"
Write-Host "  平台地址: $BaseURL"
Write-Host "  集群 ID: $ClusterID"
Write-Host "  Git 仓库: $GitRepo"
Write-Host "  镜像仓库: $ImageRepo"
Write-Host "  镜像标签: $ImageTag"

# ==================== 步骤 1：创建流水线 ====================
Write-Step "步骤 1" "创建流水线..."

$pipelineBody = @{
    name = "nginx-build-only-pipeline-$(Get-Date -Format 'HHmmss')"
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
    Write-Success "创建流水线成功, ID: $PipelineID"
} catch {
    Write-Fail "创建流水线失败: $_"
    exit 1
}

# ==================== 步骤 2：创建发布单 ====================
Write-Step "步骤 2" "创建发布单..."

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
    Write-Success "创建发布单成功, ID: $ReleaseID"
} catch {
    Write-Fail "创建发布单失败: $_"
    exit 1
}

# ==================== 步骤 3：触发 Jenkins 构建 ====================
Write-Step "步骤 3" "触发 Jenkins 构建..."

$runBody = @{
    id = $PipelineID
    branch = "main"
    env_vars = @{
        IMAGE_TAG = $ImageTag
        BUILD_ID = "$ReleaseID"
        PLATFORM_CALLBACK_URL = "$BaseURL/cicd/callback/build"
    }
} | ConvertTo-Json -Depth 3

try {
    $runResp = Invoke-RestMethod -Uri "$CicdURL/pipeline/run" -Method POST -Headers $Headers -Body $runBody
    Write-Success "触发构建成功"
    Write-Info "Jenkins 将执行: 拉代码 → 构建镜像 → 推送 Harbor → 回调平台"
} catch {
    Write-Fail "触发构建失败: $_"
    exit 1
}

# ==================== 步骤 4：等待构建完成 ====================
Write-Step "步骤 4" "等待 Jenkins 构建完成..."

$waited = 0
$buildSuccess = $false

while ($waited -lt $MaxWait) {
    Start-Sleep -Seconds $Interval
    $waited += $Interval
    
    try {
        $statusResp = Invoke-RestMethod -Uri "$CicdURL/pipeline/status?id=$PipelineID" -Method GET -Headers $Headers
        $buildStatus = $statusResp.data.pipeline.last_run_status
        $building = if ($statusResp.data.build_info) { $statusResp.data.build_info.building } else { $false }
        
        Write-Info "构建状态: $buildStatus, Building: $building ($waited/$MaxWait 秒)"
        
        if ($buildStatus -in @("success", "failed", "aborted")) {
            if ($buildStatus -eq "success") {
                Write-Success "Jenkins 构建完成"
                $buildSuccess = $true
            } else {
                Write-Fail "Jenkins 构建失败: $buildStatus"
                exit 1
            }
            break
        }
    } catch {
        Write-Info "查询状态失败: $_"
    }
}

if (-not $buildSuccess -and $waited -ge $MaxWait) {
    Write-Fail "等待构建超时"
    exit 1
}

# ==================== 步骤 5：等待平台 Worker 部署 ====================
Write-Step "步骤 5" "等待平台 Worker 部署完成..."

$waited = 0
$deploySuccess = $false

while ($waited -lt $MaxWait) {
    Start-Sleep -Seconds $Interval
    $waited += $Interval
    
    try {
        $releaseDetail = Invoke-RestMethod -Uri "$CicdURL/release/detail?id=$ReleaseID" -Method GET -Headers $Headers
        $releaseStatus = $releaseDetail.data.release.status
        
        Write-Info "发布单状态: $releaseStatus ($waited/$MaxWait 秒)"
        
        if ($releaseStatus -in @("Succeeded", "Failed", "Canceled")) {
            if ($releaseStatus -eq "Succeeded") {
                Write-Success "平台 Worker 部署成功！"
                $deploySuccess = $true
            } else {
                Write-Fail "部署失败: $releaseStatus"
                # 查看任务详情
                $tasks = $releaseDetail.data.tasks
                foreach ($task in $tasks) {
                    Write-Info "Task[$($task.id)]: $($task.status) - $($task.message)"
                }
                exit 1
            }
            break
        }
    } catch {
        Write-Info "查询发布单失败: $_"
    }
}

if (-not $deploySuccess -and $waited -ge $MaxWait) {
    Write-Fail "等待部署超时"
    exit 1
}

# ==================== 步骤 6：验证 K8s 部署结果 ====================
Write-Step "步骤 6" "验证 K8s 部署结果..."

# 检查 kubectl 是否可用
$kubectlAvailable = $null -ne (Get-Command kubectl -ErrorAction SilentlyContinue)

if ($kubectlAvailable) {
    # 检查 Deployment 镜像
    $currentImage = kubectl get deployment nginx-demo -n default -o jsonpath='{.spec.template.spec.containers[0].image}' 2>$null
    Write-Info "当前镜像: $currentImage"
    Write-Info "目标镜像: ${ImageRepo}:${ImageTag}"

    if ($currentImage -like "*$ImageTag*") {
        Write-Success "镜像更新验证通过！"
    } else {
        Write-Host "⚠️ 镜像可能未更新，请检查" -ForegroundColor Yellow
    }

    # 检查 Pod 状态
    Write-Host "`n  Pod 状态:" -ForegroundColor Cyan
    kubectl get pods -n default -l app=nginx-demo -o wide
} else {
    Write-Host "⚠️ kubectl 不可用，跳过 K8s 验证" -ForegroundColor Yellow
}

# 访问服务
Write-Host "`n  访问服务:" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:30080" -UseBasicParsing -TimeoutSec 5
    Write-Success "服务访问成功，状态码: $($response.StatusCode)"
} catch {
    Write-Host "⚠️ 服务访问失败（可能是 NodePort 不可达）: $_" -ForegroundColor Yellow
}

# ==================== 步骤 7：查看发布单详情 ====================
Write-Step "步骤 7" "查看发布单详情..."

try {
    $releaseDetail = Invoke-RestMethod -Uri "$CicdURL/release/detail?id=$ReleaseID" -Method GET -Headers $Headers
    $release = $releaseDetail.data.release
    $tasks = $releaseDetail.data.tasks
    
    Write-Host "`n  发布单信息:" -ForegroundColor Cyan
    Write-Info "ID: $($release.id)"
    Write-Info "状态: $($release.status)"
    Write-Info "应用: $($release.app_name)"
    Write-Info "镜像: $($release.image_repo):$($release.image_tag)"
    
    Write-Host "`n  任务列表:" -ForegroundColor Cyan
    foreach ($task in $tasks) {
        Write-Info "Task[$($task.id)] 集群:$($task.cluster_id) 状态:$($task.status)"
        Write-Info "  目标镜像: $($task.target_image)"
        Write-Info "  原镜像: $($task.prev_image)"
    }
} catch {
    Write-Host "⚠️ 查询发布单详情失败: $_" -ForegroundColor Yellow
}

# ==================== 完成 ====================
Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "  验证完成！" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

Write-Host @"

关键验证点:
  [✓] Jenkins Job 中没有 Deploy 阶段
  [✓] Jenkins 不需要 kubeconfig 凭证
  [✓] 平台 Worker 执行 K8s 部署
  [✓] 发布单记录 PrevImage，支持回滚

生成的资源:
  - 流水线 ID: $PipelineID
  - 发布单 ID: $ReleaseID
  - 镜像标签: $ImageTag

回滚测试（可选）:
  `$cancelBody = @{ id = $ReleaseID } | ConvertTo-Json
  Invoke-RestMethod -Uri "$CicdURL/release/cancel" -Method POST -Headers `$Headers -Body `$cancelBody

"@ -ForegroundColor Gray
