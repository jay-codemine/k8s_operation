# CICD Pipeline 接口完整验证脚本
# 用于验证流水线全闭环功能

$BaseURL = "http://localhost:8080/api/v1/k8s/cicd"
$Token = "YOUR_TOKEN_HERE"  # 替换为实际 Token

$Headers = @{
    "Content-Type" = "application/json"
    "Authorization" = "Bearer $Token"
}

# 存储创建的资源 ID
$script:PipelineId = 0

function Write-Step {
    param([string]$Message)
    Write-Host "`n========================================" -ForegroundColor Cyan
    Write-Host $Message -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
}

function Test-Response {
    param($Response, [string]$ExpectedMsg = "")
    if ($Response.code -eq 0) {
        Write-Host "✅ 成功" -ForegroundColor Green
        if ($ExpectedMsg) { Write-Host "   $ExpectedMsg" -ForegroundColor Gray }
        return $true
    } else {
        Write-Host "❌ 失败: $($Response.msg)" -ForegroundColor Red
        if ($Response.data) {
            Write-Host "   详情: $($Response.data | ConvertTo-Json -Compress)" -ForegroundColor Yellow
        }
        return $false
    }
}

# ==================== 1. 流水线列表 ====================
function Test-PipelineList {
    Write-Step "1. 测试获取流水线列表 GET /pipeline/list"
    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/list?page=1&page_size=10" -Method GET -Headers $Headers
        if (Test-Response $resp "获取流水线列表成功") {
            Write-Host "   当前流水线数量: $($resp.data.total)" -ForegroundColor Gray
            return $true
        }
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
    return $false
}

# ==================== 2. 创建流水线 ====================
function Test-PipelineCreate {
    Write-Step "2. 测试创建流水线 POST /pipeline/create"
    
    $timestamp = Get-Date -Format "yyyyMMddHHmmss"
    $body = @{
        name = "test-pipeline-$timestamp"
        description = "自动化测试流水线"
        git_repo = "https://github.com/jay-codemine/k8s_operation.git"
        git_branch = "main"
        jenkins_url = ""
        jenkins_job = "test-job"
        env_vars = @(
            @{ name = "ENV"; value = "test" }
        )
        deploy_config = @{
            replicas = 2
            strategy = "rollingUpdate"
        }
    } | ConvertTo-Json -Depth 4

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/create" -Method POST -Headers $Headers -Body $body
        if (Test-Response $resp) {
            $script:PipelineId = $resp.data.pipeline_id
            Write-Host "   创建的流水线 ID: $script:PipelineId" -ForegroundColor Gray
            return $true
        }
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}

# ==================== 3. 获取流水线详情 ====================
function Test-PipelineDetail {
    Write-Step "3. 测试获取流水线详情 GET /pipeline/detail"
    
    if ($script:PipelineId -eq 0) {
        Write-Host "⚠️ 跳过: 没有可用的流水线 ID" -ForegroundColor Yellow
        return $false
    }

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/detail?id=$script:PipelineId" -Method GET -Headers $Headers
        if (Test-Response $resp) {
            Write-Host "   流水线名称: $($resp.data.pipeline.name)" -ForegroundColor Gray
            Write-Host "   Git 仓库: $($resp.data.pipeline.git_repo)" -ForegroundColor Gray
            return $true
        }
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}

# ==================== 4. 更新流水线 ====================
function Test-PipelineUpdate {
    Write-Step "4. 测试更新流水线 POST /pipeline/update"
    
    if ($script:PipelineId -eq 0) {
        Write-Host "⚠️ 跳过: 没有可用的流水线 ID" -ForegroundColor Yellow
        return $false
    }

    $body = @{
        id = $script:PipelineId
        description = "更新后的描述 - $(Get-Date -Format 'HH:mm:ss')"
    } | ConvertTo-Json

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/update" -Method POST -Headers $Headers -Body $body
        return Test-Response $resp "描述已更新"
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}

# ==================== 5. 获取流水线状态 ====================
function Test-PipelineStatus {
    Write-Step "5. 测试获取流水线状态 GET /pipeline/status"
    
    if ($script:PipelineId -eq 0) {
        Write-Host "⚠️ 跳过: 没有可用的流水线 ID" -ForegroundColor Yellow
        return $false
    }

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/status?id=$script:PipelineId" -Method GET -Headers $Headers
        if (Test-Response $resp) {
            Write-Host "   当前状态: $($resp.data.pipeline.status)" -ForegroundColor Gray
            Write-Host "   上次运行状态: $($resp.data.pipeline.last_run_status)" -ForegroundColor Gray
            return $true
        }
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}

# ==================== 6. 获取运行历史 ====================
function Test-PipelineHistory {
    Write-Step "6. 测试获取运行历史 GET /pipeline/history"
    
    if ($script:PipelineId -eq 0) {
        Write-Host "⚠️ 跳过: 没有可用的流水线 ID" -ForegroundColor Yellow
        return $false
    }

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/history?id=$script:PipelineId&page=1&page_size=10" -Method GET -Headers $Headers
        if (Test-Response $resp) {
            Write-Host "   历史记录数: $($resp.data.total)" -ForegroundColor Gray
            return $true
        }
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}

# ==================== 7. 测试运行流水线（需要 Jenkins）====================
function Test-PipelineRun {
    Write-Step "7. 测试运行流水线 POST /pipeline/run (需要 Jenkins 配置)"
    
    if ($script:PipelineId -eq 0) {
        Write-Host "⚠️ 跳过: 没有可用的流水线 ID" -ForegroundColor Yellow
        return $false
    }

    $body = @{
        id = $script:PipelineId
        branch = "main"
        env_vars = @{
            BUILD_ENV = "test"
        }
    } | ConvertTo-Json -Depth 3

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/run" -Method POST -Headers $Headers -Body $body
        if ($resp.code -eq 0) {
            Write-Host "✅ 流水线启动成功" -ForegroundColor Green
            Write-Host "   运行 ID: $($resp.data.run_id)" -ForegroundColor Gray
            return $true
        } else {
            # Jenkins 未配置时预期失败
            Write-Host "⚠️ 运行失败 (可能是 Jenkins 未配置): $($resp.msg)" -ForegroundColor Yellow
            return $false
        }
    } catch {
        Write-Host "⚠️ 请求异常 (可能是 Jenkins 未配置): $($_.Exception.Message)" -ForegroundColor Yellow
    }
    return $false
}

# ==================== 8. 测试获取日志（需要 Jenkins）====================
function Test-PipelineLogs {
    Write-Step "8. 测试获取构建日志 GET /pipeline/logs (需要 Jenkins 配置)"
    
    if ($script:PipelineId -eq 0) {
        Write-Host "⚠️ 跳过: 没有可用的流水线 ID" -ForegroundColor Yellow
        return $false
    }

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/logs?id=$script:PipelineId" -Method GET -Headers $Headers
        if ($resp.code -eq 0) {
            Write-Host "✅ 获取日志成功" -ForegroundColor Green
            $logPreview = if ($resp.data.logs.Length -gt 100) { $resp.data.logs.Substring(0, 100) + "..." } else { $resp.data.logs }
            Write-Host "   日志预览: $logPreview" -ForegroundColor Gray
            return $true
        } else {
            Write-Host "⚠️ 获取日志失败 (可能无构建记录): $($resp.msg)" -ForegroundColor Yellow
            return $false
        }
    } catch {
        Write-Host "⚠️ 请求异常: $($_.Exception.Message)" -ForegroundColor Yellow
    }
    return $false
}

# ==================== 9. 测试 Git 分支获取 ====================
function Test-GitBranches {
    Write-Step "9. 测试获取 Git 分支 POST /git/branches"
    
    $body = @{
        repo_url = "https://github.com/jay-codemine/k8s_operation.git"
    } | ConvertTo-Json

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/git/branches" -Method POST -Headers $Headers -Body $body
        if (Test-Response $resp) {
            $branchNames = ($resp.data.branches | ForEach-Object { $_.name }) -join ", "
            Write-Host "   分支列表: $branchNames" -ForegroundColor Gray
            return $true
        }
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}

# ==================== 10. 删除流水线 ====================
function Test-PipelineDelete {
    Write-Step "10. 测试删除流水线 POST /pipeline/delete"
    
    if ($script:PipelineId -eq 0) {
        Write-Host "⚠️ 跳过: 没有可用的流水线 ID" -ForegroundColor Yellow
        return $false
    }

    $body = @{
        id = $script:PipelineId
    } | ConvertTo-Json

    try {
        $resp = Invoke-RestMethod -Uri "$BaseURL/pipeline/delete" -Method POST -Headers $Headers -Body $body
        return Test-Response $resp "流水线已删除"
    } catch {
        Write-Host "❌ 请求异常: $($_.Exception.Message)" -ForegroundColor Red
    }
    return $false
}

# ==================== 主测试流程 ====================
function Start-AllTests {
    Write-Host "`n" -NoNewline
    Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
    Write-Host "║          CICD Pipeline 接口完整验证                          ║" -ForegroundColor Magenta
    Write-Host "║          $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')                                    ║" -ForegroundColor Magenta
    Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Magenta

    $results = @{
        "流水线列表" = Test-PipelineList
        "创建流水线" = Test-PipelineCreate
        "流水线详情" = Test-PipelineDetail
        "更新流水线" = Test-PipelineUpdate
        "流水线状态" = Test-PipelineStatus
        "运行历史" = Test-PipelineHistory
        "运行流水线" = Test-PipelineRun
        "获取日志" = Test-PipelineLogs
        "Git分支" = Test-GitBranches
        "删除流水线" = Test-PipelineDelete
    }

    # 汇总报告
    Write-Host "`n"
    Write-Host "╔══════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
    Write-Host "║                     测试结果汇总                             ║" -ForegroundColor Magenta
    Write-Host "╚══════════════════════════════════════════════════════════════╝" -ForegroundColor Magenta
    
    $passed = 0
    $failed = 0
    $skipped = 0

    foreach ($test in $results.GetEnumerator()) {
        $status = if ($test.Value -eq $true) { 
            $passed++
            "✅ 通过"
        } elseif ($test.Value -eq $false) { 
            $failed++
            "❌ 失败/跳过"
        } else {
            $skipped++
            "⚠️ 未知"
        }
        Write-Host "  $($test.Key.PadRight(15)) : $status"
    }

    Write-Host "`n  ────────────────────────────────────────"
    Write-Host "  通过: $passed | 失败/跳过: $failed" -ForegroundColor $(if ($failed -eq 0) { "Green" } else { "Yellow" })
    Write-Host ""
}

# 运行测试
Start-AllTests
