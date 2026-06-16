#!/bin/bash
# ============================================================
# K8sOperation 本地 K8s 部署验证脚本
# 前置条件：Docker Desktop + Kubernetes 已启用
# ============================================================
set -e

echo "=============================================="
echo " K8sOperation 本地 K8s 部署验证"
echo "=============================================="
echo ""

# 检查 kubectl
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl 未安装"
    exit 1
fi

# 检查集群连接
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ K8s 集群未连接！请先启用 Docker Desktop Kubernetes："
    echo "   Docker Desktop → Settings → Kubernetes → Enable Kubernetes"
    exit 1
fi

echo "✅ K8s 集群已连接"
kubectl cluster-info | head -1

# 检查本地镜像
echo ""
echo "🔍 检查本地 Docker 镜像..."
if ! docker images | grep -q "k8soperation.*latest"; then
    echo "❌ 后端镜像 k8soperation:latest 不存在，请先构建："
    echo "   make build && docker build -t k8soperation:latest ."
    exit 1
fi
if ! docker images | grep -q "k8soperation-web.*latest"; then
    echo "❌ 前端镜像 k8soperation-web:latest 不存在，请先构建："
    echo "   docker build -f k8s-web/Dockerfile -t k8soperation-web:latest ./k8s-web"
    exit 1
fi
echo "✅ 镜像已就绪"

# 部署
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
echo ""
echo "🚀 部署到 K8s..."
kubectl apply -f "$SCRIPT_DIR/all-in-one.yaml"

# 等待 Pod 就绪
echo ""
echo "⏳ 等待 Pod 就绪（最多 120 秒）..."
kubectl -n k8soperation wait --for=condition=ready pod -l app.kubernetes.io/name=k8soperation-web --timeout=120s 2>/dev/null || true

echo ""
echo "📊 Pod 状态："
kubectl -n k8soperation get pods -o wide

echo ""
echo "📊 Service 状态："
kubectl -n k8soperation get svc

echo ""
echo "=============================================="
echo " 🎉 部署完成！访问地址："
echo "=============================================="
echo ""
echo "  🌐 前端：http://localhost:30080"
echo "  🔧 后端：http://localhost:30880/healthz/live"
echo ""
echo "  验证命令："
echo "    curl -s http://localhost:30080/health"
echo "    curl -s http://localhost:30880/healthz/live"
echo ""
echo "  清理命令："
echo "    kubectl delete -f $SCRIPT_DIR/all-in-one.yaml"
echo "=============================================="
