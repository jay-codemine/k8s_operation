# CICD 端到端闭环验证脚本
# 用途：验证 平台 → Jenkins → K8s 完整闭环
# 
# 使用方式：
#   .\cicd_e2e_test.ps1 -BaseUrl "http://localhost:8080" -JenkinsUrl "http://jenkins:8080"
#
# 验证流程：
#   1. 登录平台获取Token
#   2. 创建测试流水线
#   3. 触发Jenkins构建
#   4. 轮询等待完成
#   5. 验证最终状态
#   6. 清理测试数据

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$JenkinsUrl = "http://jenkins.example.com:8080",
    [string]$Username = "admin",
    [string]$Password = "admin123",
    [int]$Timeout = 300,       # 超时时间（秒）
    [switch]$SkipCleanup       # 跳过清理
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message, [string]$Color = "Yellow")
    Write-Host "`n[$(Get-Date -Format 'HH:mm:ss')] $Message" -ForegroundColor $Color
}

function Write-Result {
    param([string]$Message, [bool]$Success)
    if ($Success) {
        Write-Host "  ✅ $Message" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $Message" -ForegroundColor Red
    }
}

function Invoke-API {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [object]$Body = $null
    )
    
    $uri = "$BaseUrl$Endpoint"
    $params = @{
        Method = $Method
        Uri = $uri
        Headers = $Headers
        ContentType = "application/json"
    }
    
    if ($Body) {
        $params.Body = ($Body | ConvertTo-Json -Depth 10)
    }
    
    try {
        $response = Invoke-RestMethod @params
        return $response
    } catch {
        Write-Host "  API请求失败: $uri" -ForegroundColor Red
        Write-Host "  错误: $($_.Exception.Message)" -ForegroundColor Red
        throw
    }
}

# ========================================
# 开始验证
# ========================================

Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║           CICD 完整闭环验证测试                              ║" -ForegroundColor Cyan
Write-Host "║  Platform → Jenkins → K8s → Callback → Platform             ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

Write-Host "`n配置信息:" -ForegroundColor Gray
Write-Host "  Platform URL: $BaseUrl"
Write-Host "  Jenkins URL:  $JenkinsUrl"
Write-Host "  Timeout:      $Timeout 秒"

$testResults = @{}

# ========================================
# Step 1: 登录
# ========================================
Write-Step "Step 1: 登录平台获取 Token"

try {
    $loginResp = Invoke-API -Method "POST" -Endpoint "/api/v1/auth/login" -Body @{
        username = $Username
        password = $Password
    }
    
    $token = $loginResp.data.token
    if ($token) {
        Write-Result "登录成功，Token: $($token.Substring(0, [Math]::Min(20, $token.Length)))..." $true
        $testResults["登录"] = $true
    } else {
        Write-Result "登录响应无Token" $false
        $testResults["登录"] = $false
        exit 1
    }
} catch {
    Write-Result "登录失败: $($_.Exception.Message)" $false
    $testResults["登录"] = $false
    exit 1
}

$authHeaders = @{ Authorization = "Bearer $token" }

# ========================================
# Step 2: 健康检查
# ========================================
Write-Step "Step 2: 平台健康检查"

try {
    # 检查CICD接口可用性
    $listResp = Invoke-API -Method "GET" -Endpoint "/api/v1/k8s/cicd/pipeline/list?page=1&page_size=1" -Headers $authHeaders
    
    if ($listResp.code -eq 0) {
        Write-Result "Pipeline API 可用" $true
        $testResults["健康检查"] = $true
    } else {
        Write-Result "Pipeline API 异常: $($listResp.msg)" $false
        $testResults["健康检查"] = $false
    }
} catch {
    Write-Result "健康检查失败: $($_.Exception.Message)" $false
    $testResults["健康检查"] = $false
}

# ========================================
# Step 3: 创建测试流水线
# ========================================
Write-Step "Step 3: 创建测试流水线"

$pipelineName = "e2e-test-$(Get-Date -Format 'yyyyMMddHHmmss')"
$pipelineId = $null

try {
    $createResp = Invoke-API -Method "POST" -Endpoint "/api/v1/k8s/cicd/pipeline/create" -Headers $authHeaders -Body @{
        name = $pipelineName
        description = "端到端闭环验证测试流水线"
        git_repo = "https://github.com/example/demo-app.git"
        git_branch = "main"
        jenkins_url = $JenkinsUrl
        jenkins_job = "k8s-platform-deploy"
        env_vars = @(
            @{ name = "IMAGE_REPO"; value = "harbor.io/demo/app" }
            @{ name = "NAMESPACE"; value = "default" }
            @{ name = "DEPLOYMENT_NAME"; value = "demo-app" }
            @{ name = "CONTAINER_NAME"; value = "app" }
        )
    }
    
    if ($createResp.code -eq 0 -and $createResp.data.pipeline_id) {
        $pipelineId = $createResp.data.pipeline_id
        Write-Result "流水线创建成功: ID=$pipelineId, Name=$pipelineName" $true
        $testResults["创建流水线"] = $true
    } else {
        Write-Result "创建流水线失败: $($createResp.msg)" $false
        $testResults["创建流水线"] = $false
    }
} catch {
    Write-Result "创建流水线异常: $($_.Exception.Message)" $false
    $testResults["创建流水线"] = $false
}

# ========================================
# Step 4: 验证流水线详情
# ========================================
Write-Step "Step 4: 验证流水线详情"

if ($pipelineId) {
    try {
        $detailResp = Invoke-API -Method "GET" -Endpoint "/api/v1/k8s/cicd/pipeline/detail?id=$pipelineId" -Headers $authHeaders
        
        if ($detailResp.code -eq 0 -and $detailResp.data.pipeline) {
            $pipeline = $detailResp.data.pipeline
            Write-Host "  Name: $($pipeline.name)"
            Write-Host "  Git Repo: $($pipeline.git_repo)"
            Write-Host "  Jenkins Job: $($pipeline.jenkins_job)"
            Write-Host "  Status: $($pipeline.status)"
            Write-Result "详情查询成功" $true
            $testResults["详情查询"] = $true
        } else {
            Write-Result "详情查询失败" $false
            $testResults["详情查询"] = $false
        }
    } catch {
        Write-Result "详情查询异常: $($_.Exception.Message)" $false
        $testResults["详情查询"] = $false
    }
}

# ========================================
# Step 5: 触发流水线运行
# ========================================
Write-Step "Step 5: 触发流水线运行（平台 → Jenkins）"

$runId = $null
if ($pipelineId) {
    try {
        $runResp = Invoke-API -Method "POST" -Endpoint "/api/v1/k8s/cicd/pipeline/run" -Headers $authHeaders -Body @{
            id = $pipelineId
            env_vars = @{
                IMAGE_TAG = "v1.0.0-$(Get-Date -Format 'yyyyMMddHHmmss')"
            }
        }
        
        if ($runResp.code -eq 0) {
            $runId = $runResp.data.run_id
            Write-Result "触发成功: Run ID=$runId" $true
            Write-Host "  Message: $($runResp.data.message)"
            $testResults["触发运行"] = $true
        } else {
            Write-Result "触发失败: $($runResp.msg)" $false
            $testResults["触发运行"] = $false
        }
    } catch {
        Write-Result "触发异常: $($_.Exception.Message)" $false
        $testResults["触发运行"] = $false
    }
}

# ========================================
# Step 6: 轮询等待完成
# ========================================
Write-Step "Step 6: 等待构建完成（轮询状态）"

$elapsed = 0
$interval = 10
$finalStatus = "unknown"
$lastRunStatus = "unknown"

if ($pipelineId -and $testResults["触发运行"]) {
    Write-Host "  超时设置: $Timeout 秒"
    Write-Host "  轮询间隔: $interval 秒"
    Write-Host ""
    
    while ($elapsed -lt $Timeout) {
        try {
            $statusResp = Invoke-API -Method "GET" -Endpoint "/api/v1/k8s/cicd/pipeline/status?id=$pipelineId" -Headers $authHeaders
            
            if ($statusResp.code -eq 0 -and $statusResp.data.pipeline) {
                $finalStatus = $statusResp.data.pipeline.status
                $lastRunStatus = $statusResp.data.pipeline.last_run_status
                $buildNumber = $statusResp.data.pipeline.last_build_number
                
                $minutes = [math]::Floor($elapsed / 60)
                $seconds = $elapsed % 60
                Write-Host "  [$($minutes)m $($seconds)s] Status: $finalStatus | RunStatus: $lastRunStatus | Build: #$buildNumber" -ForegroundColor Cyan
                
                # 检查是否完成
                if ($finalStatus -eq "idle" -and $lastRunStatus -in @("success", "failed", "aborted")) {
                    Write-Host ""
                    Write-Result "构建完成！最终状态: $lastRunStatus" ($lastRunStatus -eq "success")
                    $testResults["构建完成"] = ($lastRunStatus -eq "success")
                    break
                }
            }
        } catch {
            Write-Host "  [$elapsed s] 状态查询失败: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        Start-Sleep -Seconds $interval
        $elapsed += $interval
    }
    
    if ($elapsed -ge $Timeout) {
        Write-Result "等待超时！最终状态: $lastRunStatus" $false
        $testResults["构建完成"] = $false
    }
} else {
    Write-Host "  跳过（流水线未触发）" -ForegroundColor Gray
    $testResults["构建完成"] = $false
}

# ========================================
# Step 7: 验证回调闭环
# ========================================
Write-Step "Step 7: 验证回调闭环"

if ($pipelineId) {
    try {
        $finalDetailResp = Invoke-API -Method "GET" -Endpoint "/api/v1/k8s/cicd/pipeline/detail?id=$pipelineId" -Headers $authHeaders
        
        if ($finalDetailResp.code -eq 0) {
            $p = $finalDetailResp.data.pipeline
            Write-Host "  Final Status:     $($p.status)"
            Write-Host "  Last Run Status:  $($p.last_run_status)"
            Write-Host "  Last Build #:     $($p.last_build_number)"
            Write-Host "  Last Run Time:    $(if($p.last_run_time -gt 0) { [DateTimeOffset]::FromUnixTimeSeconds($p.last_run_time).LocalDateTime } else { 'N/A' })"
            
            # 回调成功的标志：build_number > 0 且 last_run_status 为终态
            $callbackSuccess = ($p.last_build_number -gt 0) -and ($p.last_run_status -in @("success", "failed", "aborted"))
            Write-Result "回调闭环验证" $callbackSuccess
            $testResults["回调闭环"] = $callbackSuccess
        }
    } catch {
        Write-Result "回调验证异常: $($_.Exception.Message)" $false
        $testResults["回调闭环"] = $false
    }
}

# ========================================
# Step 8: 获取构建日志
# ========================================
Write-Step "Step 8: 获取构建日志"

if ($pipelineId -and $testResults["构建完成"]) {
    try {
        $logsResp = Invoke-API -Method "GET" -Endpoint "/api/v1/k8s/cicd/pipeline/logs?id=$pipelineId" -Headers $authHeaders
        
        if ($logsResp.code -eq 0 -and $logsResp.data.logs) {
            $logLines = ($logsResp.data.logs -split "`n").Count
            Write-Host "  日志行数: $logLines"
            Write-Result "日志获取成功" $true
            $testResults["日志获取"] = $true
        } else {
            Write-Result "日志为空或获取失败" $false
            $testResults["日志获取"] = $false
        }
    } catch {
        Write-Result "日志获取异常: $($_.Exception.Message)" $false
        $testResults["日志获取"] = $false
    }
} else {
    Write-Host "  跳过（无成功构建）" -ForegroundColor Gray
}

# ========================================
# Step 9: 清理测试数据
# ========================================
if (-not $SkipCleanup -and $pipelineId) {
    Write-Step "Step 9: 清理测试数据"
    
    try {
        $deleteResp = Invoke-API -Method "POST" -Endpoint "/api/v1/k8s/cicd/pipeline/delete" -Headers $authHeaders -Body @{
            id = $pipelineId
        }
        
        if ($deleteResp.code -eq 0) {
            Write-Result "测试流水线已删除: $pipelineName" $true
        } else {
            Write-Result "删除失败: $($deleteResp.msg)" $false
        }
    } catch {
        Write-Result "删除异常: $($_.Exception.Message)" $false
    }
} else {
    Write-Step "Step 9: 跳过清理（-SkipCleanup）"
}

# ========================================
# 测试报告
# ========================================
Write-Host "`n╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                      测试报告                                ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan

$passed = 0
$failed = 0

foreach ($key in $testResults.Keys) {
    if ($testResults[$key]) {
        Write-Host "  ✅ $key" -ForegroundColor Green
        $passed++
    } else {
        Write-Host "  ❌ $key" -ForegroundColor Red
        $failed++
    }
}

Write-Host ""
Write-Host "  总计: $($passed + $failed) 项" -ForegroundColor White
Write-Host "  通过: $passed 项" -ForegroundColor Green
Write-Host "  失败: $failed 项" -ForegroundColor $(if ($failed -gt 0) { "Red" } else { "Green" })

if ($failed -eq 0) {
    Write-Host "`n  🎉 闭环验证全部通过！" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n  ⚠️  部分验证失败，请检查日志" -ForegroundColor Yellow
    exit 1
}

