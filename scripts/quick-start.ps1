# ============================================================
# K8sOperation Quick Start Script (Windows PowerShell)
# Steps: Check env -> Init DB -> Generate config -> Build backend -> Build frontend -> Start
# Usage: powershell -ExecutionPolicy Bypass -File scripts\quick-start.ps1
# ============================================================

$ErrorActionPreference = "Stop"

# ---- Root directory ----
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $ScriptDir

# ---- Default config (override via env vars) ----
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "127.0.0.1" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "3306" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "root" }
$DB_PASS = if ($env:DB_PASS) { $env:DB_PASS } else { "admin123" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "k8s-platform" }

$REDIS_HOST = if ($env:REDIS_HOST) { $env:REDIS_HOST } else { "127.0.0.1" }
$REDIS_PORT = if ($env:REDIS_PORT) { $env:REDIS_PORT } else { "6379" }
$REDIS_PASS = if ($env:REDIS_PASS) { $env:REDIS_PASS } else { "admin123" }

$BACKEND_PORT = if ($env:BACKEND_PORT) { $env:BACKEND_PORT } else { "8080" }
$FRONTEND_PORT = if ($env:FRONTEND_PORT) { $env:FRONTEND_PORT } else { "5173" }

# ---- Helper functions ----
function Write-Info    { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Blue }
function Write-Ok      { param($msg) Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn    { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Fail    { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red }
function Write-Step    { param($msg) Write-Host "" ; Write-Host ">>> $msg" -ForegroundColor Cyan }

function Test-Command {
    param([string]$Name)
    $cmd = Get-Command $Name -ErrorAction SilentlyContinue
    if ($cmd) {
        try {
            $ver = & $Name --version 2>&1 | Select-Object -First 1
            Write-Ok "$Name  ($ver)"
        } catch {
            Write-Ok "$Name  (installed)"
        }
        return $true
    } else {
        Write-Fail "$Name  not installed"
        return $false
    }
}

function Show-Banner {
    Write-Host ""
    Write-Host "===========================================================" -ForegroundColor Cyan
    Write-Host "        K8sOperation Quick Start v2.0 (Windows)             " -ForegroundColor Cyan
    Write-Host "  GitHub: https://gitee.com/jay-kim/k8s_operation           " -ForegroundColor Cyan
    Write-Host "===========================================================" -ForegroundColor Cyan
    Write-Host ""
}

# ============================================================
# STEP 1: Environment Check
# ============================================================
function Step-CheckEnvironment {
    Write-Step "STEP 1/6: Environment Check"

    $hasError = $false

    if (-not (Test-Command "go")) {
        Write-Fail "Go not installed. Please install Go 1.21+ (https://go.dev/dl/)"
        $hasError = $true
    }

    if (-not (Test-Command "node")) {
        Write-Warn "Node.js not installed, frontend build will be skipped"
    }

    Test-Command "npm" | Out-Null
    if (-not (Test-Command "mysql")) {
        Write-Warn "mysql client not installed, manual DB init required"
    }
    Test-Command "git" | Out-Null

    if ($hasError) {
        Write-Fail "Required dependencies missing. Please install first."
        exit 1
    }
}

# ============================================================
# STEP 2: Check Services
# ============================================================
function Step-CheckServices {
    Write-Step "STEP 2/6: Check Services (MySQL / Redis)"

    # Check MySQL
    $dbAddr = "${DB_HOST}:${DB_PORT}"
    Write-Info "Checking MySQL ($dbAddr) ..."
    $mysqlCmd = Get-Command "mysql" -ErrorAction SilentlyContinue
    if ($mysqlCmd) {
        try {
            $result = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -e "SELECT 1" 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Ok "MySQL connection OK"
            } else {
                throw "Connection failed"
            }
        } catch {
            Write-Fail "MySQL connection failed!"
            Write-Host "  - Ensure MySQL is running"
            Write-Host "  - Address: $dbAddr"
            Write-Host "  - User: $DB_USER / Password: $DB_PASS"
            Write-Host ""
            Write-Host '  Override via env: $env:DB_HOST="xxx"; $env:DB_PASS="xxx"; .\scripts\quick-start.ps1'
            exit 1
        }
    } else {
        Write-Warn "mysql client not available, skipping MySQL check"
    }

    # Check Redis
    $redisAddr = "${REDIS_HOST}:${REDIS_PORT}"
    Write-Info "Checking Redis ($redisAddr) ..."
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $tcp.Connect($REDIS_HOST, [int]$REDIS_PORT)
        $tcp.Close()
        Write-Ok "Redis port reachable"
    } catch {
        Write-Fail "Redis connection failed! Ensure Redis is running on $redisAddr"
        exit 1
    }
}

# ============================================================
# STEP 3: Init Database
# ============================================================
function Step-InitDatabase {
    Write-Step "STEP 3/6: Initialize Database"

    $SqlFile = Join-Path $RootDir "docs\sql\k8s_platform_full_init.sql"

    if (-not (Test-Path $SqlFile)) {
        Write-Fail "SQL file not found: $SqlFile"
        exit 1
    }

    $mysqlCmd = Get-Command "mysql" -ErrorAction SilentlyContinue
    if ($mysqlCmd) {
        try {
            $dbExists = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -N -e "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$DB_NAME'" 2>&1
            if ($dbExists.Trim() -eq "1") {
                $tableCount = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -N -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB_NAME'" 2>&1
                if ([int]$tableCount.Trim() -gt 10) {
                    Write-Warn "Database $DB_NAME exists with $($tableCount.Trim()) tables"
                    $answer = Read-Host "Re-initialize? (y/N)"
                    if ($answer -ne "y") {
                        Write-Info "Skipping database init"
                        return
                    }
                }
            }
        } catch {
            # continue
        }

        Write-Info "Executing SQL init..."
        $initResult = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" --default-character-set=utf8mb4 -e "source $SqlFile" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Ok "Database initialized!"
            Write-Info "Default admin: admin / admin123"
        } else {
            Write-Fail "Database init failed: $initResult"
            exit 1
        }
    } else {
        Write-Warn "mysql client not installed, please run manually:"
        Write-Host "  mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASS < $SqlFile"
    }
}

# ============================================================
# STEP 4: Generate Config Files
# ============================================================
function Step-SetupConfigs {
    Write-Step "STEP 4/6: Generate Config Files"

    $configFile = Join-Path $RootDir "configs\config.yaml"
    $configExample = Join-Path $RootDir "configs\config.yaml.example"
    $k8sFile = Join-Path $RootDir "configs\k8s.yaml"
    $k8sExample = Join-Path $RootDir "configs\k8s.yaml.example"

    if (Test-Path $configFile) {
        Write-Warn "configs\config.yaml already exists, skipping"
    } else {
        if (Test-Path $configExample) {
            $content = Get-Content $configExample -Raw -Encoding UTF8
            $content = $content -replace "Host: localhost", "Host: $DB_HOST"
            $content = $content -replace "Port: 3306", "Port: $DB_PORT"
            $content = $content -replace "Username: root", "Username: $DB_USER"
            $content = $content -replace "Password: admin123", "Password: $DB_PASS"
            $redisAddr = "${REDIS_HOST}:${REDIS_PORT}"
            $content = $content -replace "Address: 127.0.0.1:6379", "Address: $redisAddr"
            $content | Set-Content $configFile -Encoding UTF8
            Write-Ok "Generated configs\config.yaml"
        } else {
            Write-Fail "config.yaml.example not found"
            exit 1
        }
    }

    if (Test-Path $k8sFile) {
        Write-Warn "configs\k8s.yaml already exists, skipping"
    } else {
        if (Test-Path $k8sExample) {
            Copy-Item $k8sExample $k8sFile
            Write-Ok "Generated configs\k8s.yaml (replace with real KubeConfig)"
        } else {
            Write-Warn "k8s.yaml.example not found, K8s features need manual config"
        }
    }

    Write-Info "Config files at: $RootDir\configs\"
}

# ============================================================
# STEP 5: Build Backend
# ============================================================
function Step-BuildBackend {
    Write-Step "STEP 5/6: Build Backend (Go)"

    Set-Location $RootDir

    Write-Info "Downloading Go dependencies..."
    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Fail "Go dependency download failed"
        exit 1
    }

    $logDir = Join-Path $RootDir "storage\logs"
    if (-not (Test-Path $logDir)) {
        New-Item -ItemType Directory -Path $logDir -Force | Out-Null
    }

    $binDir = Join-Path $RootDir "bin"
    if (-not (Test-Path $binDir)) {
        New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    }

    Write-Info "Building backend..."
    go build -trimpath -ldflags="-s -w" -o "bin\k8soperation.exe" ./cmd/k8soperation
    if ($LASTEXITCODE -ne 0) {
        Write-Fail "Backend build failed"
        exit 1
    }
    Write-Ok "Backend built: bin\k8soperation.exe"
}

# ============================================================
# STEP 6: Build Frontend
# ============================================================
function Step-BuildFrontend {
    Write-Step "STEP 6/6: Build Frontend (Vue3)"

    $nodeCmd = Get-Command "node" -ErrorAction SilentlyContinue
    if (-not $nodeCmd) {
        Write-Warn "Node.js not installed, skipping frontend build"
        return
    }

    Set-Location (Join-Path $RootDir "k8s-web")

    $nodeModules = Join-Path $RootDir "k8s-web\node_modules"
    if (-not (Test-Path $nodeModules)) {
        Write-Info "Installing frontend dependencies..."
        npm install
        if ($LASTEXITCODE -ne 0) {
            Write-Fail "Frontend dependency install failed"
            exit 1
        }
    } else {
        Write-Info "Frontend dependencies already installed"
    }

    Write-Info "Building frontend..."
    npx vite build
    if ($LASTEXITCODE -ne 0) {
        Write-Fail "Frontend build failed"
        exit 1
    }
    Write-Ok "Frontend built: k8s-web\dist\"
}

# ============================================================
# Show Result
# ============================================================
function Step-ShowResult {
    Write-Host ""
    Write-Host "==========================================================" -ForegroundColor Green
    Write-Host "  All done!" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "-- Start Backend --" -ForegroundColor White
    Write-Host "  cd $RootDir"
    Write-Host "  .\bin\k8soperation.exe"
    Write-Host "  # Backend: http://localhost:$BACKEND_PORT"
    Write-Host ""
    Write-Host "-- Start Frontend (dev mode) --" -ForegroundColor White
    Write-Host "  cd $RootDir\k8s-web"
    Write-Host "  npm run dev"
    Write-Host "  # Frontend: http://localhost:$FRONTEND_PORT"
    Write-Host ""
    Write-Host "-- Docker Deploy --" -ForegroundColor White
    Write-Host "  make docker-build  # Build image"
    Write-Host "  make docker-run    # Start container"
    Write-Host ""
    Write-Host "-- Login Info --" -ForegroundColor White
    Write-Host "  URL:      http://localhost:$FRONTEND_PORT"
    Write-Host "  Username: admin"
    Write-Host "  Password: admin123"
    Write-Host ""

    $answer = Read-Host "Start backend now? (Y/n)"
    if ($answer -ne "n") {
        Write-Info "Starting backend..."
        Set-Location $RootDir
        $backendProcess = Start-Process -FilePath ".\bin\k8soperation.exe" -PassThru -NoNewWindow
        Start-Sleep -Seconds 3

        if (-not $backendProcess.HasExited) {
            Write-Ok "Backend started (PID: $($backendProcess.Id))"

            $nodeCmd = Get-Command "node" -ErrorAction SilentlyContinue
            if ($nodeCmd) {
                $feAnswer = Read-Host "Also start frontend dev server? (Y/n)"
                if ($feAnswer -ne "n") {
                    Write-Info "Starting frontend dev server..."
                    Set-Location (Join-Path $RootDir "k8s-web")
                    Start-Process -FilePath "npm" -ArgumentList "run","dev" -NoNewWindow
                    Start-Sleep -Seconds 3
                    Write-Ok "Frontend started: http://localhost:$FRONTEND_PORT"
                }
            }

            Write-Host ""
            Write-Ok "All services running! Press Ctrl+C to stop"
            Write-Host "Tip: Stop backend with Stop-Process -Id $($backendProcess.Id)"
            $backendProcess.WaitForExit()
        } else {
            Write-Fail "Backend failed to start. Check storage\logs\app.log"
        }
    }
}

# ============================================================
# Main
# ============================================================
Show-Banner
Step-CheckEnvironment
Step-CheckServices
Step-InitDatabase
Step-SetupConfigs
Step-BuildBackend
Step-BuildFrontend
Step-ShowResult
