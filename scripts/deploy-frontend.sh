#!/usr/bin/env bash
# ============================================================
# K8sOperation 前端一键部署脚本
# 用途: 构建前端镜像 → 推送仓库 → 部署到 K8s
# 使用:
#   bash scripts/deploy-frontend.sh
#   IMAGE_TAG=v1.2.0 bash scripts/deploy-frontend.sh
#   SKIP_BUILD=true bash scripts/deploy-frontend.sh
#   EXPOSURE=ingress DOMAIN=k8sop.example.com bash scripts/deploy-frontend.sh
# ============================================================
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# ---- 配置（可通过环境变量覆盖） ----
NAMESPACE="${NAMESPACE:-k8soperation}"
IMAGE_REGISTRY="${IMAGE_REGISTRY:-registry.cn-hangzhou.aliyuncs.com/k8s-gos}"
IMAGE_NAME="${IMAGE_NAME:-k8soperation-web}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE="${IMAGE_REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"

EXPOSURE="${EXPOSURE:-clusterip}"            # clusterip | nodeport | ingress
DOMAIN="${DOMAIN:-k8sop.example.com}"
NODE_PORT="${NODE_PORT:-30081}"
BACKEND_URL="${BACKEND_URL:-http://k8soperation.k8soperation.svc:8080}"

SKIP_BUILD="${SKIP_BUILD:-false}"            # true = 跳过构建，直接部署
REPLICAS="${REPLICAS:-2}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_DIR="$ROOT_DIR/k8s-web"
DEPLOY_DIR="$ROOT_DIR/deploy"

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()    { echo -e "${RED}[FAIL]${NC} $*"; exit 1; }
step()    { echo -e "\n${BOLD}${CYAN}▶ $*${NC}"; }

banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║       K8sOperation 前端一键部署脚本 v1.0                ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo "  命名空间:   $NAMESPACE"
    echo "  镜像:       $FULL_IMAGE"
    echo "  后端地址:   $BACKEND_URL"
    echo "  暴露方式:   $EXPOSURE"
    echo "  副本数:     $REPLICAS"
    echo "  跳过构建:   $SKIP_BUILD"
    if [[ "$EXPOSURE" == "ingress" ]]; then
        echo "  域名:       $DOMAIN"
    elif [[ "$EXPOSURE" == "nodeport" ]]; then
        echo "  NodePort:   $NODE_PORT"
    fi
    echo ""
}

# ============================================================
# STEP 1: 前置检查
# ============================================================
preflight() {
    step "STEP 1/4: 前置检查"

    # 检查 kubectl
    if ! command -v kubectl &>/dev/null; then
        fail "kubectl 未安装"
    fi
    success "kubectl 已安装"

    if ! kubectl cluster-info &>/dev/null; then
        fail "无法连接 K8s 集群，请检查 kubeconfig"
    fi
    success "K8s 集群连接正常"

    # 检查 Docker（构建时需要）
    if [[ "$SKIP_BUILD" != "true" ]]; then
        if ! command -v docker &>/dev/null; then
            fail "docker 未安装（构建镜像需要 Docker）"
        fi
        success "Docker 已安装: $(docker --version | head -1)"

        if [[ ! -f "$FRONTEND_DIR/Dockerfile" ]]; then
            fail "前端 Dockerfile 不存在: k8s-web/Dockerfile"
        fi
        success "前端 Dockerfile 已就绪"
    fi

    # 检查 K8s 部署文件
    if [[ ! -f "$DEPLOY_DIR/frontend-deployment.yaml" ]]; then
        fail "缺少部署文件: deploy/frontend-deployment.yaml"
    fi
    success "K8s 部署文件完整"
}

# ============================================================
# STEP 2: 构建并推送镜像
# ============================================================
build_and_push() {
    step "STEP 2/4: 构建并推送前端镜像"

    if [[ "$SKIP_BUILD" == "true" ]]; then
        warn "跳过构建（SKIP_BUILD=true），直接使用镜像: $FULL_IMAGE"
        return 0
    fi

    info "开始构建前端镜像..."
    docker build \
        -f "$FRONTEND_DIR/Dockerfile" \
        -t "$FULL_IMAGE" \
        "$FRONTEND_DIR"
    success "镜像构建完成: $FULL_IMAGE"

    info "推送镜像到仓库..."
    docker push "$FULL_IMAGE"
    success "镜像推送完成"
}

# ============================================================
# STEP 3: 部署到 K8s
# ============================================================
deploy_to_k8s() {
    step "STEP 3/4: 部署到 K8s"

    # 确保命名空间存在
    kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    success "命名空间 $NAMESPACE 就绪"

    # 使用 sed 动态替换镜像和后端地址后 apply
    info "应用前端 Deployment + Service + Ingress..."
    cat "$DEPLOY_DIR/frontend-deployment.yaml" | \
        sed "s|image: .*k8soperation-web:.*|image: ${FULL_IMAGE}|g" | \
        sed "s|replicas: .*|replicas: ${REPLICAS}|" | \
        sed "s|value: \"http://k8soperation.k8soperation.svc:8080\"|value: \"${BACKEND_URL}\"|g" | \
        sed "s|host: k8sop.example.com|host: ${DOMAIN}|g" | \
        kubectl apply -f -
    success "前端 Deployment 已部署"

    # 等待就绪
    info "等待前端 Pod 就绪..."
    kubectl -n "$NAMESPACE" wait --for=condition=available \
        deployment/k8soperation-web --timeout=120s 2>/dev/null || \
        warn "前端启动超时，请检查: kubectl -n $NAMESPACE get pods -l app.kubernetes.io/name=k8soperation-web"

    success "前端部署完成！"
}

# ============================================================
# STEP 4: 暴露服务
# ============================================================
expose_service() {
    step "STEP 4/4: 暴露前端服务 (${EXPOSURE})"

    case "$EXPOSURE" in
        nodeport)
            # 将 Service 类型改为 NodePort
            kubectl -n "$NAMESPACE" patch svc k8soperation-web \
                --type='json' \
                -p="[{\"op\":\"replace\",\"path\":\"/spec/type\",\"value\":\"NodePort\"},{\"op\":\"replace\",\"path\":\"/spec/ports/0/nodePort\",\"value\":${NODE_PORT}}]" \
                2>/dev/null || \
                warn "NodePort 设置可能需要手动调整"
            success "NodePort Service 已设置"
            echo ""
            echo -e "${GREEN}  访问地址: http://<任意节点IP>:${NODE_PORT}${NC}"
            ;;
        ingress)
            success "Ingress 已包含在 frontend-deployment.yaml 中"
            echo ""
            echo -e "${GREEN}  访问地址: http://${DOMAIN}${NC}"
            echo "  请确保域名已解析到 Ingress Controller 的 IP"
            ;;
        clusterip)
            info "使用 ClusterIP 模式（仅集群内访问）"
            echo "  集群内地址: http://k8soperation-web.${NAMESPACE}.svc:80"
            echo "  本地调试:   kubectl -n $NAMESPACE port-forward svc/k8soperation-web 8080:80"
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
    echo -e "${BOLD}${GREEN}  ✅ K8sOperation 前端部署完成！${NC}"
    echo -e "${BOLD}${GREEN}════════════════════════════════════════════${NC}"
    echo ""
    echo "  镜像:     $FULL_IMAGE"
    echo "  副本:     $REPLICAS"
    echo "  后端:     $BACKEND_URL"
    echo ""
    echo "  常用命令:"
    echo "    kubectl -n $NAMESPACE get pods -l app.kubernetes.io/name=k8soperation-web"
    echo "    kubectl -n $NAMESPACE logs -f deploy/k8soperation-web"
    echo "    kubectl -n $NAMESPACE describe deploy/k8soperation-web"
    echo ""
    echo "  更新镜像（滚动更新）:"
    echo "    kubectl -n $NAMESPACE set image deploy/k8soperation-web web=$FULL_IMAGE"
    echo ""
    echo "  回滚:"
    echo "    kubectl -n $NAMESPACE rollout undo deploy/k8soperation-web"
    echo ""
    echo "  卸载前端:"
    echo "    kubectl -n $NAMESPACE delete deploy,svc,ingress k8soperation-web"
    echo ""
}

# ============================================================
# Main
# ============================================================
main() {
    banner
    preflight
    build_and_push
    deploy_to_k8s
    expose_service
    print_summary
}

main "$@"
