# =============================================================================
# Nginx CICD 闭环验证脚本
# 使用方法: 
#   1. 修改下方配置参数
#   2. 运行: .\nginx_cicd_verify.ps1
# =============================================================================

# ==================== 配置参数（需要修改） ====================
$PlatformURL = "http://localhost:8080"           # 平台地址
$PlatformToken = "your-platform-jwt-token"       # 平台登录后获取的 Token
$JenkinsURL = "http://jenkins.example.com:8080"  # Jenkins 地址
$JenkinsJob = "nginx-build-deploy"               # Jenkins Job 名称
$GitRepo = "https://github.com/your-org/nginx-demo.git"  # 你的 Git 仓库
$ImageRepo = "harbor.example.com/library/nginx-demo"     # Harbor 镜像仓库
$Namespace = "default"                           # K8s 命名空间
$DeploymentName = "nginx-demo"                   # Deployment 名称

# ==================== 以下无需修改 ====================

$BaseURL = "$PlatformURL/api/v1/k8s/cicd"
$Headers = @{
    "Authorization" = "Bearer $PlatformToken"
    "Content-Type" = "application/json"
}

function Write-Step {
    param([string]$Message)
    Write-Host ""
    Write-Host "=" * 60 -ForegroundColor Cyan
    Write-Host " $Message" -ForegroundColor Cyan
    Write-Host "=" * 60 -ForegroundColor Cyan
}

function Test-PlatformHealth {
    Write-Step "步骤 0: 检查平台健康状态"
    try {
        $health = Invoke-RestMethod -Uri "$PlatformURL/health" -Method GET -TimeoutSec 5
        Write-Host "✅ 平台服务正常" -ForegroundColor Green
        return $true
    } catch {
        Write-Host "❌ 平台服务不可用: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function New-Pipeline {
    Write-Step "步骤 1: 创建流水线"
    
    $body = @{
        name = "nginx-demo-$(Get-Date -Format 'yyyyMMddHHmmss')"
        description = "Nginx Demo CICD 验证流水线"
        git_repo = $GitRepo
        git_branch = "main"
        jenkins_url = $JenkinsURL
        jenkins_job = $JenkinsJob
        env_vars = @(
            @{ name = "IMAGE_REPO"; value = $ImageRepo }
            @{ name = "NAMESPACE"; value = $Namespace }
            @{ name = "DEPLOYMENT_NAME"; value = $DeploymentName }
            @{ name = "CONTAINER_NAME"; value = "nginx" }
        )
    } | ConvertTo-Json -Depth 3

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/create" -Method POST -Headers $Headers -Body $body
        if ($resp.code -eq 0) {
            $script:PipelineId = $resp.data.id
            Write-Host "✅ 创建成功, Pipeline ID: $($resp.data.id)" -ForegroundColor Green
            return $true
        } else {
            Write-Host "❌ 创建失败: $($resp.message)" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ 请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Start-PipelineRun {
    Write-Step "步骤 2: 触发流水线运行"
    
    $imageTag = "v1.0.$(Get-Date -Format 'yyyyMMddHHmm')"
    $callbackURL = "$PlatformURL/api/v1/k8s/cicd/pipeline/callback"
    
    $body = @{
        id = $script:PipelineId
        branch = "main"
        env_vars = @{
            IMAGE_TAG = $imageTag
            PLATFORM_CALLBACK_URL = $callbackURL
            PLATFORM_TOKEN = $PlatformToken
        }
    } | ConvertTo-Json -Depth 3

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/run" -Method POST -Headers $Headers -Body $body
        if ($resp.code -eq 0) {
            Write-Host "✅ 触发成功" -ForegroundColor Green
            Write-Host "   镜像标签: $imageTag"
            Write-Host "   回调地址: $callbackURL"
            return $true
        } else {
            Write-Host "❌ 触发失败: $($resp.message)" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ 请求失败: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Watch-BuildStatus {
    Write-Step "步骤 3: 监控构建状态"
    
    $maxWait = 300  # 最多等待 5 分钟
    $waited = 0
    $interval = 10
    
    while ($waited -lt $maxWait) {
        try {
            $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/status?id=$($script:PipelineId)" -Method GET -Headers $Headers
            
            $pipelineStatus = $resp.data.pipeline.status
            $lastRunStatus = $resp.data.pipeline.last_run_status
            $buildInfo = $resp.data.build_info
            
            $buildDisplay = if ($buildInfo) { 
                "Build #$($buildInfo.number): Building=$($buildInfo.building)" 
            } else { 
                "等待 Jenkins 响应..." 
            }
            
            Write-Host "[$waited s] Pipeline: $pipelineStatus | Run: $lastRunStatus | $buildDisplay"
            
            if ($lastRunStatus -in @("success", "failed", "aborted")) {
                if ($lastRunStatus -eq "success") {
                    Write-Host "✅ 构建成功！" -ForegroundColor Green
                } else {
                    Write-Host "❌ 构建失败: $lastRunStatus" -ForegroundColor Red
                }
                return $lastRunStatus
            }
        } catch {
            Write-Host "查询状态失败: $($_.Exception.Message)" -ForegroundColor Yellow
        }
        
        Start-Sleep -Seconds $interval
        $waited += $interval
    }
    
    Write-Host "⏰ 等待超时" -ForegroundColor Yellow
    return "timeout"
}

function Get-BuildLogs {
    Write-Step "步骤 4: 获取构建日志"
    
    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/logs?id=$($script:PipelineId)" -Method GET -Headers $Headers
        if ($resp.code -eq 0 -and $resp.data) {
            Write-Host "===== 构建日志 (最后 1000 字符) =====" -ForegroundColor Gray
            $logContent = $resp.data
            if ($logContent.Length -gt 1000) {
                $logContent = "..." + $logContent.Substring($logContent.Length - 1000)
            }
            Write-Host $logContent
            Write-Host "===== 日志结束 =====" -ForegroundColor Gray
        } else {
            Write-Host "暂无日志或获取失败" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "获取日志失败: $($_.Exception.Message)" -ForegroundColor Yellow
    }
}

function Test-K8sDeployment {
    Write-Step "步骤 5: 验证 K8s 部署状态"
    
    Write-Host "请手动执行以下命令验证:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "# 1. 查看 Deployment 当前镜像" -ForegroundColor Cyan
    Write-Host "kubectl get deployment $DeploymentName -n $Namespace -o jsonpath='{.spec.template.spec.containers[0].image}'"
    Write-Host ""
    Write-Host "# 2. 查看 Pod 状态" -ForegroundColor Cyan
    Write-Host "kubectl get pods -n $Namespace -l app=$DeploymentName"
    Write-Host ""
    Write-Host "# 3. 查看 rollout 状态" -ForegroundColor Cyan
    Write-Host "kubectl rollout status deployment/$DeploymentName -n $Namespace"
}

# ==================== 主流程 ====================

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "║         Nginx CICD 闭环验证脚本                             ║" -ForegroundColor Magenta
Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Magenta
Write-Host ""
Write-Host "配置信息:"
Write-Host "  平台地址:    $PlatformURL"
Write-Host "  Jenkins:     $JenkinsURL"
Write-Host "  Git 仓库:    $GitRepo"
Write-Host "  镜像仓库:    $ImageRepo"
Write-Host "  K8s 命名空间: $Namespace"
Write-Host "  Deployment:  $DeploymentName"
Write-Host ""

# 确认继续
$confirm = Read-Host "确认以上配置正确？(y/n)"
if ($confirm -ne "y") {
    Write-Host "已取消" -ForegroundColor Yellow
    exit 0
}

# 执行验证流程
$script:PipelineId = $null

if (-not (Test-PlatformHealth)) {
    Write-Host "请先启动平台服务" -ForegroundColor Red
    exit 1
}

if (-not (New-Pipeline)) {
    exit 1
}

if (-not (Start-PipelineRun)) {
    exit 1
}

$result = Watch-BuildStatus

Get-BuildLogs

Test-K8sDeployment

# 总结
Write-Host ""
Write-Host "═" * 60 -ForegroundColor Magenta
if ($result -eq "success") {
    Write-Host " ✅ CICD 闭环验证成功！" -ForegroundColor Green
} else {
    Write-Host " ❌ CICD 闭环验证失败，请检查日志" -ForegroundColor Red
}
Write-Host "═" * 60 -ForegroundColor Magenta
