#!/usr/bin/env bash
# ================================================================
#  K8sOperation 一键部署脚本 (Linux / macOS)
# ================================================================
#  基于 docker-compose 一键拉起：MySQL + Redis + 后端 + 前端
#
#  用法：
#    ./deploy.sh            # 一键部署（构建镜像 + 启动全部服务）
#    ./deploy.sh up         # 同上
#    ./deploy.sh down       # 停止并删除容器（保留数据卷）
#    ./deploy.sh clean      # 停止并删除容器 + 数据卷（彻底清空，慎用）
#    ./deploy.sh restart    # 重启全部服务
#    ./deploy.sh status     # 查看服务状态
#    ./deploy.sh logs       # 跟踪查看全部日志
# ================================================================

set -euo pipefail

# ---------- 颜色输出 ----------
C_CYAN='\033[36m'; C_GREEN='\033[32m'; C_YELLOW='\033[33m'; C_RED='\033[31m'; C_RESET='\033[0m'
step() { echo -e "\n${C_CYAN}==> $1${C_RESET}"; }
ok()   { echo -e "${C_GREEN}[OK]   $1${C_RESET}"; }
warn() { echo -e "${C_YELLOW}[WARN] $1${C_RESET}"; }
err()  { echo -e "${C_RED}[ERR]  $1${C_RESET}"; }

# ---------- 定位项目根目录（脚本在 deploy/quick-deploy/ 下，根目录为上两级）----------
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.yaml"

# ---------- 前置检查 ----------
check_prereq() {
    step "检查运行环境"

    if ! command -v docker >/dev/null 2>&1; then
        err "未检测到 docker，请先安装 Docker：https://docs.docker.com/engine/install/"
        exit 1
    fi
    ok "docker 已安装"

    if ! docker info >/dev/null 2>&1; then
        err "Docker 守护进程未运行，请先启动 Docker 再重试（Linux: systemctl start docker）。"
        exit 1
    fi
    ok "Docker 守护进程运行中"

    if ! docker compose version >/dev/null 2>&1; then
        err "未检测到 docker compose (v2)，请升级 Docker 或安装 compose 插件。"
        exit 1
    fi
    ok "docker compose 可用"

    if [ ! -f "$COMPOSE_FILE" ]; then
        err "未找到 docker-compose.yaml：$COMPOSE_FILE"
        exit 1
    fi
    ok "找到 compose 文件：$COMPOSE_FILE"
}

# ---------- 部署 ----------
do_up() {
    check_prereq
    step "构建镜像并启动服务（首次会拉取/构建，请耐心等待）"
    ( cd "$PROJECT_ROOT" && docker compose up -d --build )

    step "等待服务健康检查通过（最多 120 秒）"
    local deadline=$(( $(date +%s) + 120 ))
    while [ "$(date +%s)" -lt "$deadline" ]; do
        local unhealthy
        unhealthy="$(cd "$PROJECT_ROOT" && docker compose ps --format '{{.Service}} {{.Health}}' 2>/dev/null | grep -E '(starting|unhealthy)' || true)"
        [ -z "$unhealthy" ] && break
        sleep 5
        printf "."
    done
    echo ""

    show_status
    show_access
}

do_down() {
    step "停止并删除容器（保留数据卷）"
    ( cd "$PROJECT_ROOT" && docker compose down )
    ok "已停止"
}

do_clean() {
    warn "即将删除容器 + 数据卷（MySQL/Redis 数据将全部清空）！"
    read -r -p "确认请输入 yes: " confirm
    if [ "$confirm" != "yes" ]; then echo "已取消"; return; fi
    ( cd "$PROJECT_ROOT" && docker compose down -v )
    ok "已彻底清空"
}

do_restart() {
    step "重启全部服务"
    ( cd "$PROJECT_ROOT" && docker compose restart )
    show_status
}

show_status() {
    step "服务状态"
    ( cd "$PROJECT_ROOT" && docker compose ps )
}

do_logs() {
    ( cd "$PROJECT_ROOT" && docker compose logs -f --tail=100 )
}

show_access() {
    echo ""
    echo -e "${C_GREEN}============================================================${C_RESET}"
    echo -e "${C_GREEN} K8sOperation 部署完成！访问信息如下：${C_RESET}"
    echo -e "${C_GREEN}============================================================${C_RESET}"
    echo " 前端控制台 : http://localhost"
    echo " 后端 API   : http://localhost:8080"
    echo " Swagger    : http://localhost:8080/swagger/index.html"
    echo ""
    echo -e "${C_YELLOW} 默认管理员 : admin / 123456   (首次启动自动创建)${C_RESET}"
    echo ""
    echo " MySQL      : localhost:3306  (root / admin123, db=k8s-platform)"
    echo " Redis      : localhost:6379  (密码 admin123)"
    echo "------------------------------------------------------------"
    echo " 查看状态   : ./deploy/quick-deploy/deploy.sh status"
    echo " 查看日志   : ./deploy/quick-deploy/deploy.sh logs"
    echo " 停止服务   : ./deploy/quick-deploy/deploy.sh down"
    echo -e "${C_GREEN}============================================================${C_RESET}"
}

# ---------- 入口 ----------
action="${1:-up}"
case "$action" in
    up|deploy) do_up ;;
    down)      do_down ;;
    clean)     do_clean ;;
    restart)   do_restart ;;
    status)    show_status ;;
    logs)      do_logs ;;
    *)
        err "未知命令：$action"
        echo "可用命令：up | down | clean | restart | status | logs"
        exit 1
        ;;
esac
