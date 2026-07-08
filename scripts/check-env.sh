#!/usr/bin/env bash
# ============================================================
# K8sOperation 环境检查脚本 (Linux / macOS)
# 仅检查环境，不做任何修改操作
# 使用: bash scripts/check-env.sh
# ============================================================

set -uo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

PASS=0
WARN=0
FAIL=0

check() {
    local name="$1"
    local required="$2"  # yes / no
    local min_ver="${3:-}"

    if command -v "$name" &>/dev/null; then
        local ver
        ver=$("$name" --version 2>/dev/null | head -1 || echo "installed")
        echo -e "  ${GREEN}✓${NC} $name  ($ver)"
        ((PASS++))
        return 0
    else
        if [[ "$required" == "yes" ]]; then
            echo -e "  ${RED}✗${NC} $name  — ${RED}必需${NC}，未安装"
            ((FAIL++))
        else
            echo -e "  ${YELLOW}△${NC} $name  — 可选，未安装"
            ((WARN++))
        fi
        return 1
    fi
}

check_port() {
    local host="$1" port="$2" name="$3"
    if (echo >/dev/tcp/"$host"/"$port") 2>/dev/null; then
        echo -e "  ${GREEN}✓${NC} $name ($host:$port) — 可连接"
        ((PASS++))
    else
        echo -e "  ${RED}✗${NC} $name ($host:$port) — 无法连接"
        ((FAIL++))
    fi
}

echo -e "${BOLD}${CYAN}"
echo "╔══════════════════════════════════════════════╗"
echo "║    K8sOperation 环境检查报告                 ║"
echo "╚══════════════════════════════════════════════╝"
echo -e "${NC}"

# ---- 必须工具 ----
echo -e "${BOLD}[1] 必要工具${NC}"
check "go"     yes
check "git"    no
echo ""

# ---- 前端工具 ----
echo -e "${BOLD}[2] 前端工具${NC}"
check "node"   no
check "npm"    no
echo ""

# ---- 数据库工具 ----
echo -e "${BOLD}[3] 数据库工具${NC}"
check "mysql"     no
check "redis-cli" no
echo ""

# ---- Docker (可选) ----
echo -e "${BOLD}[4] 容器工具 (可选)${NC}"
check "docker"  no
check "nerdctl" no
check "kubectl" no
echo ""

# ---- 服务连通性 ----
echo -e "${BOLD}[5] 服务连通性${NC}"
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"

check_port "$DB_HOST" "$DB_PORT" "MySQL"
check_port "$REDIS_HOST" "$REDIS_PORT" "Redis"
echo ""

# ---- Go 环境 ----
if command -v go &>/dev/null; then
    echo -e "${BOLD}[6] Go 环境详情${NC}"
    echo "  GOROOT : $(go env GOROOT)"
    echo "  GOPATH : $(go env GOPATH)"
    echo "  GOVERSION: $(go version)"
    echo "  GOPROXY: $(go env GOPROXY)"
    echo ""
fi

# ---- 磁盘空间 ----
echo -e "${BOLD}[7] 磁盘空间${NC}"
df -h . 2>/dev/null | head -2 || echo "  无法获取磁盘信息"
echo ""

# ---- 汇总 ----
echo -e "${BOLD}════════════════════════════════════${NC}"
echo -e "  通过: ${GREEN}$PASS${NC}  警告: ${YELLOW}$WARN${NC}  失败: ${RED}$FAIL${NC}"
echo -e "${BOLD}════════════════════════════════════${NC}"

if [[ $FAIL -gt 0 ]]; then
    echo -e "\n${RED}存在必要依赖缺失，请先安装后再运行 quick-start.sh${NC}"
    exit 1
else
    echo -e "\n${GREEN}环境检查通过！可以运行 quick-start.sh 开始部署${NC}"
    exit 0
fi
