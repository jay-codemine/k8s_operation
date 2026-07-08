#!/usr/bin/env bash
# ============================================================
# K8sOperation 一键 K8s 部署脚本
# 用途: 从零部署整个平台到 K8s 集群（含中间件 + 应用）
# 使用: bash scripts/deploy-k8s.sh
#       EXPOSURE=nodeport bash scripts/deploy-k8s.sh
#       EXPOSURE=ingress DOMAIN=ops.example.com bash scripts/deploy-k8s.sh
# ============================================================
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# ---- 配置 ----
NAMESPACE="${NAMESPACE:-k8soperation}"
EXPOSURE="${EXPOSURE:-nodeport}"          # nodeport | ingress | clusterip
DOMAIN="${DOMAIN:-k8soperation.example.com}"
NODE_PORT="${NODE_PORT:-30080}"
IMAGE="${IMAGE:-registry.cn-hangzhou.aliyuncs.com/k8s-gos/k8soperation:latest}"
DEPLOY_MIDDLEWARE="${DEPLOY_MIDDLEWARE:-true}"  # 是否部署 MySQL/Redis（false=使用外部）

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOY_DIR="$ROOT_DIR/deploy"

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()    { echo -e "${RED}[FAIL]${NC} $*"; exit 1; }
step()    { echo -e "\n${BOLD}${CYAN}▶ $*${NC}"; }

banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║       K8sOperation K8s 一键部署脚本 v1.0                ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo "  命名空间:   $NAMESPACE"
    echo "  暴露方式:   $EXPOSURE"
    if [[ "$EXPOSURE" == "ingress" ]]; then
        echo "  域名:       $DOMAIN"
    elif [[ "$EXPOSURE" == "nodeport" ]]; then
        echo "  NodePort:   $NODE_PORT"
    fi
    echo "  镜像:       $IMAGE"
    echo "  部署中间件: $DEPLOY_MIDDLEWARE"
    echo ""
}

# ============================================================
# STEP 1: 前置检查
# ============================================================
preflight() {
    step "STEP 1/5: 前置检查"

    if ! command -v kubectl &>/dev/null; then
        fail "kubectl 未安装"
    fi
    success "kubectl 已安装: $(kubectl version --client --short 2>/dev/null || kubectl version --client | head -1)"

    if ! kubectl cluster-info &>/dev/null; then
        fail "无法连接 K8s 集群，请检查 kubeconfig"
    fi
    success "K8s 集群连接正常"

    # 检查部署文件
    for f in namespace.yaml secret.yaml configmap.yaml pvc.yaml service.yaml deployment.yaml; do
        if [[ ! -f "$DEPLOY_DIR/$f" ]]; then
            fail "缺少部署文件: deploy/$f"
        fi
    done
    success "部署文件完整"
}

# ============================================================
# STEP 2: 创建命名空间 + Secret
# ============================================================
create_namespace_and_secrets() {
    step "STEP 2/5: 创建命名空间和密钥"

    kubectl apply -f "$DEPLOY_DIR/namespace.yaml"
    success "命名空间 $NAMESPACE 已创建"

    kubectl apply -f "$DEPLOY_DIR/secret.yaml"
    success "Secret 已创建"

    warn "请确认 deploy/secret.yaml 中的密码已替换为实际值！"
    echo "  生成 base64: echo -n 'your-password' | base64"
}

# ============================================================
# STEP 3: 部署中间件（MySQL + Redis）
# ============================================================
deploy_middleware() {
    step "STEP 3/5: 部署中间件 (MySQL + Redis)"

    if [[ "$DEPLOY_MIDDLEWARE" != "true" ]]; then
        warn "跳过中间件部署（DEPLOY_MIDDLEWARE=false）"
        warn "请确保 configmap.yaml 中的 Database.Host 和 Cache.Address 指向外部服务"
        return 0
    fi

    if [[ ! -f "$DEPLOY_DIR/middleware.yaml" ]]; then
        fail "middleware.yaml 不存在"
    fi

    kubectl apply -f "$DEPLOY_DIR/middleware.yaml"
    success "MySQL + Redis 已部署"

    info "等待 MySQL 就绪..."
    kubectl -n "$NAMESPACE" wait --for=condition=available deployment/mysql --timeout=120s 2>/dev/null || \
        warn "MySQL 启动超时，请手动检查: kubectl -n $NAMESPACE get pods"

    info "等待 Redis 就绪..."
    kubectl -n "$NAMESPACE" wait --for=condition=available deployment/redis --timeout=60s 2>/dev/null || \
        warn "Redis 启动超时"

    # 导入 SQL 初始化脚本
    info "等待 MySQL Pod Ready..."
    sleep 10
    local mysql_pod
    mysql_pod=$(kubectl -n "$NAMESPACE" get pod -l app.kubernetes.io/component=mysql -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

    if [[ -n "$mysql_pod" ]]; then
        local sql_file="$ROOT_DIR/docs/sql/k8s_platform_full_init.sql"
        if [[ -f "$sql_file" ]]; then
            info "导入数据库初始化脚本..."
            kubectl -n "$NAMESPACE" exec -i "$mysql_pod" -- \
                mysql -u root -p"$(kubectl -n $NAMESPACE get secret k8soperation-secret -o jsonpath='{.data.DB_PASSWORD}' | base64 -d)" \
                --default-character-set=utf8mb4 < "$sql_file" && \
                success "SQL 初始化完成" || \
                warn "SQL 导入失败，请手动执行"
        fi
    fi
}

# ============================================================
# STEP 4: 部署应用（ConfigMap + PVC + Deployment）
# ============================================================
deploy_application() {
    step "STEP 4/5: 部署应用"

    # ConfigMap
    kubectl apply -f "$DEPLOY_DIR/configmap.yaml"
    success "ConfigMap 已创建"

    # PVC
    kubectl apply -f "$DEPLOY_DIR/pvc.yaml"
    success "PVC (artifacts + logs) 已创建"

    # Service (ClusterIP)
    kubectl apply -f "$DEPLOY_DIR/service.yaml"
    success "Service + ServiceAccount + RBAC 已创建"

    # Deployment
    kubectl apply -f "$DEPLOY_DIR/deployment.yaml"
    success "Deployment 已创建"

    info "等待应用就绪..."
    kubectl -n "$NAMESPACE" wait --for=condition=available deployment/k8soperation --timeout=180s 2>/dev/null || \
        warn "应用启动超时，请检查: kubectl -n $NAMESPACE logs -l app.kubernetes.io/name=k8soperation"
}

# ============================================================
# STEP 5: 暴露服务
# ============================================================
expose_service() {
    step "STEP 5/5: 暴露服务 (${EXPOSURE})"

    case "$EXPOSURE" in
        nodeport)
            kubectl apply -f "$DEPLOY_DIR/service-nodeport.yaml"
            success "NodePort Service 已创建"
            echo ""
            echo -e "${GREEN}  访问地址: http://<任意节点IP>:${NODE_PORT}${NC}"
            echo ""
            ;;
        ingress)
            if [[ ! -f "$DEPLOY_DIR/ingress.yaml" ]]; then
                fail "ingress.yaml 不存在"
            fi
            # 动态替换域名
            sed "s|k8soperation.example.com|$DOMAIN|g" "$DEPLOY_DIR/ingress.yaml" | kubectl apply -f -
            success "Ingress 已创建"
            echo ""
            echo -e "${GREEN}  访问地址: http://${DOMAIN}${NC}"
            echo "  请确保域名已解析到 Ingress Controller 的 IP"
            echo ""
            ;;
        clusterip)
            info "使用 ClusterIP 模式（仅集群内访问）"
            echo "  集群内地址: http://k8soperation.${NAMESPACE}.svc:8080"
            echo "  本地调试:  kubectl -n $NAMESPACE port-forward svc/k8soperation 8080:8080"
            ;;
        *)
            fail "不支持的暴露方式: $EXPOSURE (可选: nodeport | ingress | clusterip)"
            ;;
    esac
}

# ============================================================
# 打印最终信息
# ============================================================
print_summary() {
    echo ""
    echo -e "${BOLD}${GREEN}════════════════════════════════════════════${NC}"
    echo -e "${BOLD}${GREEN}  ✅ K8sOperation 部署完成！${NC}"
    echo -e "${BOLD}${GREEN}════════════════════════════════════════════${NC}"
    echo ""
    echo "  登录账号: admin"
    echo "  登录密码: admin123"
    echo ""
    echo "  常用命令:"
    echo "    kubectl -n $NAMESPACE get pods              # 查看 Pod 状态"
    echo "    kubectl -n $NAMESPACE logs -f deploy/k8soperation  # 查看日志"
    echo "    kubectl -n $NAMESPACE describe pod          # 排查问题"
    echo "    kubectl -n $NAMESPACE get pvc               # 查看存储"
    echo ""
    echo "  卸载:"
    echo "    kubectl delete namespace $NAMESPACE"
    echo ""
}

# ============================================================
# Main
# ============================================================
main() {
    banner
    preflight
    create_namespace_and_secrets
    deploy_middleware
    deploy_application
    expose_service
    print_summary
}

main "$@"
