# ============================================================
# K8sOperation Database Init Script (Windows PowerShell)
# Usage: powershell -ExecutionPolicy Bypass -File scripts\init-db.ps1
#        $env:DB_HOST="10.0.0.1"; $env:DB_PASS="mypass"; .\scripts\init-db.ps1
# ============================================================

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RootDir = Split-Path -Parent $ScriptDir

# Config (override via env vars)
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "127.0.0.1" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "3306" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "root" }
$DB_PASS = if ($env:DB_PASS) { $env:DB_PASS } else { "admin123" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "k8s-platform" }
$SqlFile = if ($env:SQL_FILE) { $env:SQL_FILE } else { Join-Path $RootDir "docs\sql\k8s_platform_full_init.sql" }

function Write-Info { param($msg) Write-Host "[INFO] $msg" -ForegroundColor Blue }
function Write-Ok   { param($msg) Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn { param($msg) Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Fail { param($msg) Write-Host "[FAIL] $msg" -ForegroundColor Red; exit 1 }

# ---- Banner ----
Write-Host ""
Write-Host "============================================" -ForegroundColor Blue
Write-Host "  K8sOperation Database Init (Windows)      " -ForegroundColor Blue
Write-Host "============================================" -ForegroundColor Blue
Write-Host ""

# Check mysql client
$mysqlCmd = Get-Command "mysql" -ErrorAction SilentlyContinue
if (-not $mysqlCmd) {
    Write-Fail "mysql client not installed. Please install MySQL Client first."
}

# Check SQL file
if (-not (Test-Path $SqlFile)) {
    Write-Fail "SQL file not found: $SqlFile"
}

# Show config
$dbAddr = "${DB_HOST}:${DB_PORT}"
Write-Info "Target: $DB_USER@$dbAddr/$DB_NAME"
Write-Info "SQL file: $SqlFile"
Write-Host ""

# Test connection
Write-Info "Testing MySQL connection..."
try {
    $testResult = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -e "SELECT 1" 2>&1
    if ($LASTEXITCODE -ne 0) { throw "Connection failed" }
    Write-Ok "MySQL connection OK"
} catch {
    Write-Fail "MySQL connection failed! Check address and password."
}

# Check if database exists
try {
    $dbExists = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -N -e "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$DB_NAME'" 2>&1
    if ($dbExists.Trim() -eq "1") {
        $tableCount = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -N -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB_NAME'" 2>&1
        Write-Warn "Database $DB_NAME exists with $($tableCount.Trim()) tables"
        Write-Host ""
        Write-Host "Choose action:"
        Write-Host "  1) Skip init (keep existing data)"
        Write-Host "  2) Drop and recreate (ALL DATA LOST!)"
        Write-Host "  3) Run CREATE IF NOT EXISTS (safe append)"
        Write-Host ""
        $choice = Read-Host "Choose [1/2/3]"

        switch ($choice) {
            "1" {
                Write-Info "Skipping database init"
                exit 0
            }
            "2" {
                Write-Warn "About to DROP database $DB_NAME!"
                $confirm = Read-Host "Type 'YES' to confirm"
                if ($confirm -ne "YES") {
                    Write-Info "Cancelled"
                    exit 0
                }
                Write-Info "Dropping database..."
                mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -e "DROP DATABASE ``$DB_NAME``"
                Write-Ok "Old database dropped"
            }
            "3" {
                Write-Info "Safe mode (CREATE IF NOT EXISTS)..."
            }
            default {
                Write-Fail "Invalid choice"
            }
        }
    }
} catch {
    # continue
}

# Execute SQL
Write-Info "Executing SQL init..."
mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" --default-character-set=utf8mb4 -e "source $SqlFile"

if ($LASTEXITCODE -eq 0) {
    Write-Ok "Database initialized!"
    Write-Host ""

    try {
        $tableCount = mysql -h $DB_HOST -P $DB_PORT -u $DB_USER "-p$DB_PASS" -N -e "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB_NAME'" 2>&1
        Write-Info "Database: $DB_NAME"
        Write-Info "Tables: $($tableCount.Trim())"
    } catch {}
    Write-Info "Default admin: admin / admin123"
    Write-Host ""
    Write-Ok "Init complete! You can now start the backend."
} else {
    Write-Fail "SQL execution failed. Check error output."
}
