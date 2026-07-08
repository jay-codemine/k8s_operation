# ============================================================
# K8sOperation Environment Check Script (Windows PowerShell)
# Read-only check, no modifications
# Usage: powershell -ExecutionPolicy Bypass -File scripts\check-env.ps1
# ============================================================

$Pass = 0
$Warn = 0
$Fail = 0

function Test-Tool {
    param(
        [string]$Name,
        [bool]$Required = $false
    )
    $cmd = Get-Command $Name -ErrorAction SilentlyContinue
    if ($cmd) {
        try {
            $ver = & $Name --version 2>&1 | Select-Object -First 1
        } catch {
            $ver = "installed"
        }
        Write-Host "  [OK] $Name  ($ver)" -ForegroundColor Green
        $script:Pass++
        return $true
    } else {
        if ($Required) {
            Write-Host "  [X]  $Name  -- REQUIRED, not installed" -ForegroundColor Red
            $script:Fail++
        } else {
            Write-Host "  [!]  $Name  -- optional, not installed" -ForegroundColor Yellow
            $script:Warn++
        }
        return $false
    }
}

function Test-Port {
    param([string]$HostAddr, [int]$Port, [string]$Name)
    $addrStr = "${HostAddr}:${Port}"
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $tcp.Connect($HostAddr, $Port)
        $tcp.Close()
        Write-Host "  [OK] $Name ($addrStr) -- connected" -ForegroundColor Green
        $script:Pass++
    } catch {
        Write-Host "  [X]  $Name ($addrStr) -- connection failed" -ForegroundColor Red
        $script:Fail++
    }
}

# ---- Banner ----
Write-Host ""
Write-Host "===========================================================" -ForegroundColor Cyan
Write-Host "        K8sOperation Environment Check (Windows)            " -ForegroundColor Cyan
Write-Host "===========================================================" -ForegroundColor Cyan
Write-Host ""

# [1] Required tools
Write-Host "[1] Required Tools" -ForegroundColor White
Test-Tool "go" -Required $true
Test-Tool "git" -Required $false
Write-Host ""

# [2] Frontend tools
Write-Host "[2] Frontend Tools" -ForegroundColor White
Test-Tool "node" -Required $false
Test-Tool "npm" -Required $false
Write-Host ""

# [3] Database tools
Write-Host "[3] Database Tools" -ForegroundColor White
Test-Tool "mysql" -Required $false
Test-Tool "redis-cli" -Required $false
Write-Host ""

# [4] Container tools
Write-Host "[4] Container Tools (optional)" -ForegroundColor White
Test-Tool "docker" -Required $false
Test-Tool "kubectl" -Required $false
Write-Host ""

# [5] Service connectivity
Write-Host "[5] Service Connectivity" -ForegroundColor White
$DbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "127.0.0.1" }
$DbPort = if ($env:DB_PORT) { [int]$env:DB_PORT } else { 3306 }
$RedisHost = if ($env:REDIS_HOST) { $env:REDIS_HOST } else { "127.0.0.1" }
$RedisPort = if ($env:REDIS_PORT) { [int]$env:REDIS_PORT } else { 6379 }

Test-Port -HostAddr $DbHost -Port $DbPort -Name "MySQL"
Test-Port -HostAddr $RedisHost -Port $RedisPort -Name "Redis"
Write-Host ""

# [6] Go environment
$goCmd = Get-Command "go" -ErrorAction SilentlyContinue
if ($goCmd) {
    Write-Host "[6] Go Environment" -ForegroundColor White
    Write-Host "  GOROOT  : $(go env GOROOT)"
    Write-Host "  GOPATH  : $(go env GOPATH)"
    Write-Host "  VERSION : $(go version)"
    Write-Host "  GOPROXY : $(go env GOPROXY)"
    Write-Host ""
}

# [7] Disk space
Write-Host "[7] Disk Space" -ForegroundColor White
try {
    $drive = (Get-Location).Drive
    $driveInfo = Get-PSDrive $drive.Name
    $freeGB = [math]::Round($driveInfo.Free / 1GB, 1)
    $usedGB = [math]::Round($driveInfo.Used / 1GB, 1)
    Write-Host "  Drive $($drive.Name):  Used: ${usedGB}GB  Free: ${freeGB}GB"
} catch {
    Write-Host "  Unable to get disk info"
}
Write-Host ""

# ---- Summary ----
Write-Host "==========================================" -ForegroundColor White
Write-Host "  Pass: $Pass  Warn: $Warn  Fail: $Fail" -ForegroundColor White
Write-Host "==========================================" -ForegroundColor White

if ($Fail -gt 0) {
    Write-Host ""
    Write-Host "Required dependencies missing. Please install them first." -ForegroundColor Red
    exit 1
} else {
    Write-Host ""
    Write-Host "Environment check passed! Run quick-start.ps1 to deploy." -ForegroundColor Green
    exit 0
}
