#!/usr/bin/env bash
# ============================================================
# K8sOperation 一键快速启动脚本 (Linux / macOS)
# 用途: 检查环境 → 初始化数据库 → 生成配置 → 编译后端 → 安装前端 → 启动服务
# 使用: chmod +x scripts/quick-start.sh && ./scripts/quick-start.sh
# ============================================================

set -euo pipefail
IFS=$'\n\t'

# ---- 颜色定义 ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# ---- 根目录（脚本所在目录的上一级）----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# ---- 默认配置（可通过环境变量覆盖）----
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASS="${DB_PASS:-admin123}"
DB_NAME="${DB_NAME:-k8s-platform}"

REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASS="${REDIS_PASS:-admin123}"

BACKEND_PORT="${BACKEND_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-5173}"

# ---- 辅助函数 ----
info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()    { echo -e "${RED}[FAIL]${NC} $*"; }
step()    { echo -e "\n${BOLD}${CYAN}▶ $*${NC}"; }

check_cmd() {
    if command -v "$1" &>/dev/null; then
        local ver
        ver=$("$1" --version 2>/dev/null | head -1 || echo "installed")
        success "$1 ✓  ($ver)"
        return 0
    else
        fail "$1 ✗  未安装"
        return 1
    fi
}

banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║           K8sOperation 快速启动脚本 v2.0                ║"
    echo "║  GitHub: https://gitee.com/jay-kim/k8s_operation        ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# ============================================================
# STEP 1: 环境检查
# ============================================================
check_environment() {
    step "STEP 1/6: 环境依赖检查"

    local has_error=0

    # Go (必须)
    if ! check_cmd go; then
        fail "Go 未安装，请先安装 Go 1.21+ (https://go.dev/dl/)"
        has_error=1
    else
        local go_ver
        go_ver=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
        if [[ $(echo "$go_ver < 1.21" | bc -l 2>/dev/null || echo 0) == 1 ]]; then
            warn "Go 版本 $go_ver 可能过低，推荐 1.21+"
        fi
    fi

    # Node.js (前端需要)
    if ! check_cmd node; then
        warn "Node.js 未安装，前端将无法编译 (https://nodejs.org/)"
        warn "如果只需要后端，可以忽略"
    else
        local node_ver
        node_ver=$(node -v | tr -d 'v' | cut -d. -f1)
        if [[ "$node_ver" -lt 20 ]]; then
            warn "Node.js 版本较低，推荐 v20+"
        fi
    fi

    # npm
    check_cmd npm || warn "npm 未安装"

    # MySQL 客户端 (初始化数据库需要)
    check_cmd mysql || warn "mysql 客户端未安装，需手动初始化数据库"

    # Redis 客户端
    check_cmd redis-cli || warn "redis-cli 未安装，无法自动检测 Redis"

    # Git (可选)
    check_cmd git || warn "git 未安装"

    if [[ $has_error -eq 1 ]]; then
        fail "存在必要依赖缺失，请先安装后重试"
        exit 1
    fi
}

# ============================================================
# STEP 2: 检查基础服务
# ============================================================
check_services() {
    step "STEP 2/6: 检查基础服务 (MySQL / Redis)"

    # 检查 MySQL
    info "检查 MySQL ($DB_HOST:$DB_PORT) ..."
    if command -v mysql &>/dev/null; then
        if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -e "SELECT 1" &>/dev/null; then
            success "MySQL 连接正常"
        else
            fail "MySQL 连接失败！请确认:"
            echo "  - MySQL 服务已启动"
            echo "  - 地址: $DB_HOST:$DB_PORT"
            echo "  - 用户: $DB_USER / 密码: $DB_PASS"
            echo ""
            echo "  可通过环境变量覆盖: DB_HOST=xxx DB_PORT=xxx DB_USER=xxx DB_PASS=xxx $0"
            exit 1
        fi
    else
        warn "mysql 客户端不可用，跳过 MySQL 检查（请确保 MySQL 已启动）"
    fi

    # 检查 Redis
    info "检查 Redis ($REDIS_HOST:$REDIS_PORT) ..."
    if command -v redis-cli &>/dev/null; then
        if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" --no-auth-warning ping 2>/dev/null | grep -q PONG; then
            success "Redis 连接正常"
        else
            # 尝试无密码连接
            if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping 2>/dev/null | grep -q PONG; then
                success "Redis 连接正常 (无密码)"
            else
                fail "Redis 连接失败！请确认 Redis 已启动"
                exit 1
            fi
        fi
    else
        warn "redis-cli 不可用，跳过 Redis 检查（请确保 Redis 已启动）"
    fi
}

# ============================================================
# STEP 3: 初始化数据库
# ============================================================
init_database() {
    step "STEP 3/6: 初始化数据库"

    local SQL_FILE="$ROOT_DIR/docs/sql/k8s_platform_full_init.sql"

    if [[ ! -f "$SQL_FILE" ]]; then
        fail "SQL 文件不存在: $SQL_FILE"
        exit 1
    fi

    # 检查数据库是否已存在
    if command -v mysql &>/dev/null; then
        local db_exists
        db_exists=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -N -e \
            "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$DB_NAME'" 2>/dev/null || echo "0")

        if [[ "$db_exists" == "1" ]]; then
            local table_count
            table_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -N -e \
                "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB_NAME'" 2>/dev/null || echo "0")

            if [[ "$table_count" -gt 10 ]]; then
                warn "数据库 $DB_NAME 已存在且包含 $table_count 张表"
                read -rp "是否要重新初始化？(y/N): " answer
                if [[ "${answer,,}" != "y" ]]; then
                    info "跳过数据库初始化"
                    return 0
                fi
            fi
        fi

        info "正在执行 SQL 初始化..."
        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" \
            --default-character-set=utf8mb4 < "$SQL_FILE"
        success "数据库初始化完成！"
        info "默认管理员账户: admin / admin123"
    else
        warn "mysql 客户端未安装，请手动执行:"
        echo "  mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASS < $SQL_FILE"
    fi
}

# ============================================================
# STEP 4: 生成配置文件
# ============================================================
setup_configs() {
    step "STEP 4/6: 生成配置文件"

    local config_file="$ROOT_DIR/configs/config.yaml"
    local config_example="$ROOT_DIR/configs/config.yaml.example"
    local k8s_file="$ROOT_DIR/configs/k8s.yaml"
    local k8s_example="$ROOT_DIR/configs/k8s.yaml.example"

    # config.yaml
    if [[ -f "$config_file" ]]; then
        warn "configs/config.yaml 已存在，跳过"
    else
        if [[ -f "$config_example" ]]; then
            cp "$config_example" "$config_file"
            # 替换数据库配置
            sed -i "s|Host: localhost|Host: $DB_HOST|g" "$config_file"
            sed -i "s|Port: 3306|Port: $DB_PORT|g" "$config_file"
            sed -i "s|Username: root|Username: $DB_USER|g" "$config_file"
            sed -i "s|Password: admin123|Password: $DB_PASS|g" "$config_file"
            # 替换 Redis 配置
            sed -i "s|Address: 127.0.0.1:6379|Address: $REDIS_HOST:$REDIS_PORT|g" "$config_file"
            sed -i "s|Password: \"admin123\"|Password: \"$REDIS_PASS\"|g" "$config_file"
            success "已生成 configs/config.yaml (基于 example 模板)"
        else
            fail "config.yaml.example 不存在，请手动创建 configs/config.yaml"
            exit 1
        fi
    fi

    # k8s.yaml
    if [[ -f "$k8s_file" ]]; then
        warn "configs/k8s.yaml 已存在，跳过"
    else
        if [[ -f "$k8s_example" ]]; then
            cp "$k8s_example" "$k8s_file"
            success "已生成 configs/k8s.yaml (示例配置，请替换为真实 KubeConfig)"
        else
            warn "k8s.yaml.example 不存在，K8s 集群功能需手动配置"
        fi
    fi

    info "配置文件位于: $ROOT_DIR/configs/"
    info "可按需修改 config.yaml 中的 Database / Cache / Jenkins / AIAssistant 等配置"
}

# ============================================================
# STEP 5: 编译后端
# ============================================================
build_backend() {
    step "STEP 5/6: 编译后端 (Go)"

    cd "$ROOT_DIR"

    info "下载 Go 依赖..."
    go mod download

    # 创建日志目录
    mkdir -p storage/logs

    info "编译后端..."
    go build -trimpath -ldflags="-s -w" -o bin/k8soperation ./cmd/k8soperation
    success "后端编译完成: bin/k8soperation"
}

# ============================================================
# STEP 6: 安装前端依赖 + 编译
# ============================================================
build_frontend() {
    step "STEP 6/6: 编译前端 (Vue3)"

    if ! command -v node &>/dev/null; then
        warn "Node.js 未安装，跳过前端编译"
        return 0
    fi

    cd "$ROOT_DIR/k8s-web"

    if [[ ! -d "node_modules" ]]; then
        info "安装前端依赖..."
        npm install
    else
        info "前端依赖已安装，跳过 npm install"
    fi

    info "编译前端..."
    npx vite build
    success "前端编译完成: k8s-web/dist/"
}

# ============================================================
# 启动服务
# ============================================================
start_services() {
    step "启动服务"

    echo ""
    echo -e "${BOLD}${GREEN}✅ 初始化全部完成！${NC}"
    echo ""
    echo -e "${BOLD}── 启动后端 ──${NC}"
    echo "  cd $ROOT_DIR"
    echo "  ./bin/k8soperation"
    echo "  # 后端运行在 http://localhost:$BACKEND_PORT"
    echo ""
    echo -e "${BOLD}── 启动前端（开发模式）──${NC}"
    echo "  cd $ROOT_DIR/k8s-web"
    echo "  npm run dev"
    echo "  # 前端运行在 http://localhost:$FRONTEND_PORT"
    echo ""
    echo -e "${BOLD}── Docker 部署 ──${NC}"
    echo "  make docker-build  # 构建镜像"
    echo "  make docker-run    # 启动容器"
    echo ""
    echo -e "${BOLD}── 登录信息 ──${NC}"
    echo "  管理后台: http://localhost:$FRONTEND_PORT"
    echo "  默认账号: admin"
    echo "  默认密码: admin123"
    echo ""

    # 询问是否立即启动
    read -rp "是否立即启动后端服务？(Y/n): " answer
    if [[ "${answer,,}" != "n" ]]; then
        info "启动后端..."
        cd "$ROOT_DIR"
        ./bin/k8soperation &
        BACKEND_PID=$!
        sleep 2

        if kill -0 $BACKEND_PID 2>/dev/null; then
            success "后端已启动 (PID: $BACKEND_PID)"

            # 启动前端
            if command -v node &>/dev/null; then
                read -rp "是否同时启动前端开发服务器？(Y/n): " fe_answer
                if [[ "${fe_answer,,}" != "n" ]]; then
                    info "启动前端开发服务器..."
                    cd "$ROOT_DIR/k8s-web"
                    npm run dev &
                    sleep 3
                    success "前端已启动: http://localhost:$FRONTEND_PORT"
                fi
            fi

            echo ""
            success "所有服务已启动！按 Ctrl+C 停止"
            wait
        else
            fail "后端启动失败，请检查日志: storage/logs/app.log"
        fi
    fi
}

# ============================================================
# Main
# ============================================================
main() {
    banner

    check_environment
    check_services
    init_database
    setup_configs
    build_backend
    build_frontend
    start_services
}

main "$@"
