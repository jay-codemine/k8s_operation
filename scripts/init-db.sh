#!/usr/bin/env bash
# ============================================================
# K8sOperation 数据库初始化脚本 (Linux / macOS)
# 用途: 独立执行数据库初始化（创建库 + 建表 + 初始数据）
# 使用: bash scripts/init-db.sh
#       DB_HOST=10.0.0.1 DB_PASS=mypass bash scripts/init-db.sh
# ============================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# 配置（可通过环境变量覆盖）
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-root}"
DB_PASS="${DB_PASS:-admin123}"
DB_NAME="${DB_NAME:-k8s-platform}"
SQL_FILE="${SQL_FILE:-$ROOT_DIR/docs/sql/k8s_platform_full_init.sql}"

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()    { echo -e "${RED}[FAIL]${NC} $*"; exit 1; }

echo -e "${BLUE}"
echo "============================================"
echo "  K8sOperation 数据库初始化"
echo "============================================"
echo -e "${NC}"

# 检查 mysql 客户端
if ! command -v mysql &>/dev/null; then
    fail "mysql 客户端未安装，请先安装 mysql-client"
fi

# 检查 SQL 文件
if [[ ! -f "$SQL_FILE" ]]; then
    fail "SQL 文件不存在: $SQL_FILE"
fi

# 显示配置
info "目标数据库: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
info "SQL 文件: $SQL_FILE"
echo ""

# 测试连接
info "测试 MySQL 连接..."
if ! mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -e "SELECT 1" &>/dev/null; then
    fail "MySQL 连接失败！请检查地址和密码"
fi
success "MySQL 连接正常"

# 检查数据库是否已存在
db_exists=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -N -e \
    "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME='$DB_NAME'" 2>/dev/null || echo "0")

if [[ "$db_exists" == "1" ]]; then
    table_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -N -e \
        "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB_NAME'" 2>/dev/null || echo "0")

    warn "数据库 $DB_NAME 已存在，包含 $table_count 张表"
    echo ""
    echo "请选择操作:"
    echo "  1) 跳过初始化（保留现有数据）"
    echo "  2) 删除重建（会丢失所有数据！）"
    echo "  3) 仅执行 CREATE IF NOT EXISTS（安全追加）"
    echo ""
    read -rp "请选择 [1/2/3]: " choice

    case "$choice" in
        1)
            info "跳过数据库初始化"
            exit 0
            ;;
        2)
            warn "即将删除并重建数据库 $DB_NAME ！"
            read -rp "确认删除？输入 'YES' 继续: " confirm
            if [[ "$confirm" != "YES" ]]; then
                info "已取消"
                exit 0
            fi
            info "删除数据库..."
            mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -e "DROP DATABASE \`$DB_NAME\`"
            success "旧数据库已删除"
            ;;
        3)
            info "安全模式执行（CREATE IF NOT EXISTS）..."
            ;;
        *)
            fail "无效选择"
            ;;
    esac
fi

# 执行 SQL
info "正在执行 SQL 初始化..."
mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" \
    --default-character-set=utf8mb4 < "$SQL_FILE"

if [[ $? -eq 0 ]]; then
    success "数据库初始化完成！"
    echo ""

    # 统计
    table_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASS" -N -e \
        "SELECT COUNT(*) FROM information_schema.TABLES WHERE TABLE_SCHEMA='$DB_NAME'" 2>/dev/null)
    info "数据库: $DB_NAME"
    info "表数量: $table_count"
    info "默认管理员: admin / admin123"
    echo ""
    success "初始化完成！现在可以启动后端了"
else
    fail "SQL 执行失败，请检查错误输出"
fi
