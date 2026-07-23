<#
================================================================
 K8sOperation 一键部署脚本 (Windows PowerShell)
================================================================
 基于 docker-compose 一键拉起：MySQL + Redis + 后端 + 前端

 用法：
   .\deploy.ps1            # 一键部署（构建镜像 + 启动全部服务）
   .\deploy.ps1 up         # 同上
   .\deploy.ps1 down       # 停止并删除容器（保留数据卷）
   .\deploy.ps1 clean      # 停止并删除容器 + 数据卷（彻底清空，慎用）
   .\deploy.ps1 restart    # 重启全部服务
   .\deploy.ps1 status     # 查看服务状态
   .\deploy.ps1 logs       # 跟踪查看全部日志
================================================================
#>

$ErrorActionPreference = "Stop"

# ---------- 定位项目根目录（脚本在 deploy/quick-deploy/ 下，根目录为上两级）----------
$ScriptDir   = Split-Path -Parent $MyInvocation.MyCommand.Definition
$ProjectRoot = (Resolve-Path (Join-Path $ScriptDir "..\..")).Path
$ComposeFile = Join-Path $ProjectRoot "docker-compose.yaml"

function Write-Step($msg)  { Write-Host "`n==> $msg" -ForegroundColor Cyan }
function Write-Ok($msg)    { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn($msg)  { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg)   { Write-Host "[ERR]  $msg" -ForegroundColor Red }

# ---------- 前置检查 ----------
function Check-Prereq {
    Write-Step "检查运行环境"

    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        Write-Err "未检测到 docker，请先安装 Docker Desktop：https://www.docker.com/products/docker-desktop/"
        exit 1
    }
    Write-Ok "docker 已安装"

    # 检查 docker 守护进程是否运行
    docker info *> $null
    if ($LASTEXITCODE -ne 0) {
        Write-Err "Docker 守护进程未运行，请先启动 Docker Desktop 再重试。"
        exit 1
    }
    Write-Ok "Docker 守护进程运行中"

    # 检查 docker compose (v2)
    docker compose version *> $null
    if ($LASTEXITCODE -ne 0) {
        Write-Err "未检测到 docker compose (v2)，请升级 Docker Desktop。"
        exit 1
    }
    Write-Ok "docker compose 可用"

    if (-not (Test-Path $ComposeFile)) {
        Write-Err "未找到 docker-compose.yaml：$ComposeFile"
        exit 1
    }
    Write-Ok "找到 compose 文件：$ComposeFile"
}

# ---------- 部署 ----------
function Do-Up {
    Check-Prereq
    Write-Step "构建镜像并启动服务（首次会拉取/构建，请耐心等待）"
    Push-Location $ProjectRoot
    try {
        docker compose up -d --build
        if ($LASTEXITCODE -ne 0) { Write-Err "docker compose 启动失败，请查看上方输出。"; exit 1 }
    } finally {
        Pop-Location
    }

    Write-Step "等待服务健康检查通过（最多 120 秒）"
    $deadline = (Get-Date).AddSeconds(120)
    while ((Get-Date) -lt $deadline) {
        Push-Location $ProjectRoot
        $unhealthy = docker compose ps --format "{{.Service}} {{.Health}}" 2>$null | Where-Object { $_ -match "(starting|unhealthy)" }
        Pop-Location
        if (-not $unhealthy) { break }
        Start-Sleep -Seconds 5
        Write-Host "." -NoNewline
    }
    Write-Host ""

    Show-Status
    Show-Access
}

function Do-Down {
    Write-Step "停止并删除容器（保留数据卷）"
    Push-Location $ProjectRoot
    try { docker compose down } finally { Pop-Location }
    Write-Ok "已停止"
}

function Do-Clean {
    Write-Warn "即将删除容器 + 数据卷（MySQL/Redis 数据将全部清空）！"
    $confirm = Read-Host "确认请输入 yes"
    if ($confirm -ne "yes") { Write-Host "已取消"; return }
    Push-Location $ProjectRoot
    try { docker compose down -v } finally { Pop-Location }
    Write-Ok "已彻底清空"
}

function Do-Restart {
    Write-Step "重启全部服务"
    Push-Location $ProjectRoot
    try { docker compose restart } finally { Pop-Location }
    Show-Status
}

function Show-Status {
    Write-Step "服务状态"
    Push-Location $ProjectRoot
    try { docker compose ps } finally { Pop-Location }
}

function Do-Logs {
    Push-Location $ProjectRoot
    try { docker compose logs -f --tail=100 } finally { Pop-Location }
}

function Show-Access {
    Write-Host ""
    Write-Host "============================================================" -ForegroundColor Green
    Write-Host " K8sOperation 部署完成！访问信息如下：" -ForegroundColor Green
    Write-Host "============================================================" -ForegroundColor Green
    Write-Host " 前端控制台 : http://localhost"
    Write-Host " 后端 API   : http://localhost:8080"
    Write-Host " Swagger    : http://localhost:8080/swagger/index.html"
    Write-Host ""
    Write-Host " 默认管理员 : admin / 123456   (首次启动自动创建)" -ForegroundColor Yellow
    Write-Host ""
    Write-Host " MySQL      : localhost:3306  (root / admin123, db=k8s-platform)"
    Write-Host " Redis      : localhost:6379  (密码 admin123)"
    Write-Host "------------------------------------------------------------"
    Write-Host " 查看状态   : .\deploy\quick-deploy\deploy.ps1 status"
    Write-Host " 查看日志   : .\deploy\quick-deploy\deploy.ps1 logs"
    Write-Host " 停止服务   : .\deploy\quick-deploy\deploy.ps1 down"
    Write-Host "============================================================" -ForegroundColor Green
}

# ---------- 入口 ----------
$action = if ($args.Count -ge 1) { $args[0].ToLower() } else { "up" }

switch ($action) {
    "up"      { Do-Up }
    "deploy"  { Do-Up }
    "down"    { Do-Down }
    "clean"   { Do-Clean }
    "restart" { Do-Restart }
    "status"  { Show-Status }
    "logs"    { Do-Logs }
    default   {
        Write-Err "未知命令：$action"
        Write-Host "可用命令：up | down | clean | restart | status | logs"
        exit 1
    }
}
