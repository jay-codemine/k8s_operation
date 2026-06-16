# ============================================================
# 批量导入流水线脚本
# ============================================================
# 用法：
#   1. 编辑 batch-import-pipelines.json，填入项目信息
#   2. 运行：.\scripts\batch-import-pipelines.ps1
#   3. 可选参数：
#      -JsonFile  指定 JSON 文件路径（默认 scripts/batch-import-pipelines.json）
#      -ApiUrl    平台 API 地址（默认 http://localhost:8080）
#      -Token     JWT Token（不传则自动登录获取）
#      -Username  登录用户名（默认 admin）
#      -Password  登录密码（默认 123456）
# ============================================================

param(
    [string]$JsonFile = "scripts/batch-import-pipelines.json",
    [string]$ApiUrl = "http://localhost:8080",
    [string]$Token = "",
    [string]$Username = "admin",
    [string]$Password = "123456"
)

$ErrorActionPreference = "Stop"

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  K8sOperation 批量导入流水线工具" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# 1. 检查 JSON 文件
if (-not (Test-Path $JsonFile)) {
    Write-Host "[ERROR] JSON 文件不存在: $JsonFile" -ForegroundColor Red
    Write-Host "请先编辑 scripts/batch-import-pipelines.json 再运行此脚本" -ForegroundColor Yellow
    exit 1
}

$jsonContent = Get-Content $JsonFile -Raw -Encoding UTF8
$data = $jsonContent | ConvertFrom-Json

$pipelineCount = $data.pipelines.Count
Write-Host "[INFO] 读取到 $pipelineCount 条流水线配置" -ForegroundColor Green
Write-Host "[INFO] skip_existing = $($data.skip_existing)" -ForegroundColor Green
Write-Host ""

# 2. 获取 Token（如果未提供）
if (-not $Token) {
    Write-Host "[INFO] 正在登录获取 Token..." -ForegroundColor Yellow
    try {
        $loginBody = @{
            username = $Username
            password = $Password
        } | ConvertTo-Json

        $loginResp = Invoke-RestMethod -Uri "$ApiUrl/api/v1/login" `
            -Method POST `
            -ContentType "application/json" `
            -Body $loginBody

        if ($loginResp.code -eq 0 -and $loginResp.data.token) {
            $Token = $loginResp.data.token
            Write-Host "[OK] 登录成功" -ForegroundColor Green
        } else {
            Write-Host "[ERROR] 登录失败: $($loginResp.msg)" -ForegroundColor Red
            exit 1
        }
    } catch {
        Write-Host "[ERROR] 登录请求失败: $_" -ForegroundColor Red
        Write-Host "请确保平台后端已启动: $ApiUrl" -ForegroundColor Yellow
        exit 1
    }
}

# 3. 调用批量创建 API
Write-Host ""
Write-Host "[INFO] 正在批量创建 $pipelineCount 条流水线..." -ForegroundColor Yellow
Write-Host ""

# 构建请求体（移除 _注释 和 _说明 字段）
$cleanPipelines = @()
foreach ($p in $data.pipelines) {
    $clean = @{}
    $p.PSObject.Properties | Where-Object { $_.Name -notlike "_*" } | ForEach-Object {
        $clean[$_.Name] = $_.Value
    }
    $cleanPipelines += $clean
}

$requestBody = @{
    skip_existing = [bool]$data.skip_existing
    pipelines = $cleanPipelines
} | ConvertTo-Json -Depth 10

try {
    $headers = @{
        "Authorization" = "Bearer $Token"
        "Content-Type"  = "application/json"
    }

    $response = Invoke-RestMethod -Uri "$ApiUrl/api/v1/k8s/cicd/pipeline/batch-create" `
        -Method POST `
        -Headers $headers `
        -Body ([System.Text.Encoding]::UTF8.GetBytes($requestBody))

    # 4. 显示结果
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "  批量创建结果" -ForegroundColor Cyan
    Write-Host "============================================" -ForegroundColor Cyan

    $resultData = if ($response.data) { $response.data } else { $response }

    Write-Host ""
    Write-Host "  成功: $($resultData.success_count)" -ForegroundColor Green
    Write-Host "  失败: $($resultData.fail_count)" -ForegroundColor $(if ($resultData.fail_count -gt 0) { "Red" } else { "Green" })
    Write-Host "  跳过: $($resultData.skip_count)" -ForegroundColor Yellow
    Write-Host ""

    # 详细结果
    if ($resultData.results) {
        Write-Host "  详细结果:" -ForegroundColor White
        Write-Host "  -----------------------------------------------" -ForegroundColor DarkGray
        foreach ($r in $resultData.results) {
            if ($r.skipped) {
                Write-Host "  [SKIP] $($r.name)" -ForegroundColor Yellow
            } elseif ($r.success) {
                Write-Host "  [OK]   $($r.name) -> pipeline_id=$($r.pipeline_id)" -ForegroundColor Green
            } else {
                Write-Host "  [FAIL] $($r.name) -> $($r.error)" -ForegroundColor Red
            }
        }
    }
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Cyan

} catch {
    Write-Host "[ERROR] API 调用失败: $_" -ForegroundColor Red
    exit 1
}
