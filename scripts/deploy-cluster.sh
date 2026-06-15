#!/bin/bash
# ============================================================
# K8sOperation - K8s Cluster One-Click Deploy Script
# ============================================================
# Usage:
#   chmod +x scripts/deploy-cluster.sh
#   ./scripts/deploy-cluster.sh
#
# Prerequisites:
#   - kubectl configured with target cluster access
#   - docker (or nerdctl) for building/pushing images
#   - A container registry accessible from the cluster
#   - MySQL 8.x + Redis accessible from the cluster
#
# Environment Variables (override defaults):
#   REGISTRY     - Image registry (e.g. registry.cn-hangzhou.aliyuncs.com/k8s-gos)
#   IMAGE_TAG    - Image tag (default: latest)
#   DB_HOST      - MySQL host (default: mysql.k8soperation.svc)
#   DB_PASSWORD  - MySQL password
#   REDIS_HOST   - Redis address (default: redis.k8soperation.svc:6379)
#   REDIS_PASS   - Redis password
#   DOMAIN       - Ingress domain (optional, e.g. k8sops.example.com)
# ============================================================

set -e

# ---- Color output ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $*"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[FAIL]${NC} $*"; exit 1; }
step()  { echo -e "\n${CYAN}========== $* ==========${NC}"; }

# ---- Configuration ----
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

REGISTRY="${REGISTRY:-registry.cn-hangzhou.aliyuncs.com/k8s-gos}"
IMAGE_NAME="k8soperation"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
NAMESPACE="k8soperation"
DEPLOY_DIR="deploy"

# Database defaults
DB_HOST="${DB_HOST:-mysql.k8soperation.svc}"
DB_PORT="${DB_PORT:-3306}"
DB_NAME="${DB_NAME:-k8s-platform}"
DB_USER="${DB_USER:-root}"
DB_PASSWORD="${DB_PASSWORD:-changeme}"
REDIS_HOST="${REDIS_HOST:-redis.k8soperation.svc:6379}"
REDIS_PASS="${REDIS_PASS:-changeme}"

# Security defaults
JWT_KEY="${JWT_KEY:-$(openssl rand -base64 16 2>/dev/null || echo 'eoNB0%bv5M7995F1')}"
HMAC_SECRET="${HMAC_SECRET:-$(openssl rand -base64 24 2>/dev/null || echo 'f8Kx9mQa2LpR7tYs3VbN6dHe1Zx4JuWq')}"
KUBE_ENCRYPT_KEY="${KUBE_ENCRYPT_KEY:-K8sOp@2024!SecureKey#AES256Encrypt}"

# Jenkins (optional)
JENKINS_URL="${JENKINS_URL:-}"
JENKINS_USER="${JENKINS_USER:-admin}"
JENKINS_TOKEN="${JENKINS_TOKEN:-}"
DINGTALK_WEBHOOK="${DINGTALK_WEBHOOK:-}"
PLATFORM_URL="${PLATFORM_URL:-}"

# ============================================================
step "Step 0: Environment Check"
# ============================================================

command -v kubectl >/dev/null 2>&1 || error "kubectl not found"
command -v docker >/dev/null 2>&1 || error "docker not found (or use: export DOCKER=nerdctl)"

kubectl cluster-info >/dev/null 2>&1 || error "Cannot connect to K8s cluster. Check kubeconfig."
info "kubectl: $(kubectl version --client --short 2>/dev/null)"
info "cluster: connected"
info "target image: $FULL_IMAGE"

# ============================================================
step "Step 1: Build & Push Image"
# ============================================================

info "Cross-compiling for linux/amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o bin/k8s_operation ./cmd/k8soperation/
info "Binary built: bin/k8s_operation ($(du -h bin/k8s_operation | cut -f1))"

info "Building Docker image: $FULL_IMAGE"
docker build -t "$FULL_IMAGE" -f Dockerfile .

info "Pushing image to registry..."
docker push "$FULL_IMAGE"
info "Image pushed successfully"

# ============================================================
step "Step 2: Generate Secret (base64 encoded)"
# ============================================================

b64() { echo -n "$1" | base64 | tr -d '\n'; }

cat > "${DEPLOY_DIR}/secret.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: k8soperation-secret
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/name: k8soperation
type: Opaque
data:
  DB_PASSWORD: "$(b64 "$DB_PASSWORD")"
  REDIS_PASSWORD: "$(b64 "$REDIS_PASS")"
  JWT_SIGNING_KEY: "$(b64 "$JWT_KEY")"
  JENKINS_URL: "$(b64 "$JENKINS_URL")"
  JENKINS_USERNAME: "$(b64 "$JENKINS_USER")"
  JENKINS_API_TOKEN: "$(b64 "$JENKINS_TOKEN")"
  HMAC_SECRET: "$(b64 "$HMAC_SECRET")"
  KUBECONFIG_ENCRYPT_KEY: "$(b64 "$KUBE_ENCRYPT_KEY")"
  DINGTALK_WEBHOOK: "$(b64 "$DINGTALK_WEBHOOK")"
  PLATFORM_FRONTEND_URL: "$(b64 "$PLATFORM_URL")"
EOF

info "Secret generated with actual values"

# ============================================================
step "Step 3: Update ConfigMap (DB/Redis address)"
# ============================================================

# Update database host in configmap if non-default
if [ "$DB_HOST" != "mysql.k8soperation.svc" ]; then
    sed -i "s|Host: mysql.k8soperation.svc|Host: ${DB_HOST}|g" "${DEPLOY_DIR}/configmap.yaml"
    info "ConfigMap DB host updated: $DB_HOST"
fi
if [ "$REDIS_HOST" != "redis.k8soperation.svc:6379" ]; then
    sed -i "s|Address: redis.k8soperation.svc:6379|Address: ${REDIS_HOST}|g" "${DEPLOY_DIR}/configmap.yaml"
    info "ConfigMap Redis address updated: $REDIS_HOST"
fi

# ============================================================
step "Step 4: Update Deployment Image"
# ============================================================

sed -i "s|image: .*k8soperation.*|image: ${FULL_IMAGE}|g" "${DEPLOY_DIR}/deployment.yaml"
info "Deployment image updated: $FULL_IMAGE"

# ============================================================
step "Step 5: Configure Ingress (optional)"
# ============================================================

DOMAIN="${DOMAIN:-}"
if [ -n "$DOMAIN" ]; then
    sed -i "s|host: k8soperation.example.com|host: ${DOMAIN}|g" "${DEPLOY_DIR}/ingress.yaml"
    # Enable ingress in kustomization
    sed -i 's|# - ingress.yaml|- ingress.yaml|g' "${DEPLOY_DIR}/kustomization.yaml"
    info "Ingress enabled: $DOMAIN"
else
    warn "No DOMAIN set, skipping Ingress. Use NodePort or kubectl port-forward."
fi

# ============================================================
step "Step 6: Deploy to K8s Cluster"
# ============================================================

info "Applying all resources..."
kubectl apply -k "${DEPLOY_DIR}/"

if [ $? -ne 0 ]; then
    error "Deployment failed!"
fi
info "All resources applied"

# ============================================================
step "Step 7: Wait for Pod Ready (max 180s)"
# ============================================================

info "Waiting for deployment rollout..."
kubectl rollout status deployment/k8soperation -n "$NAMESPACE" --timeout=180s

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}============================================================${NC}"
    echo -e "${GREEN}  Deploy SUCCESS!${NC}"
    echo -e "${GREEN}============================================================${NC}"
    echo ""
    echo -e "  Image:     $FULL_IMAGE"
    echo -e "  Namespace: $NAMESPACE"
    echo -e "  Account:   admin / admin123"
    echo ""

    if [ -n "$DOMAIN" ]; then
        echo -e "  Access URL: http://${DOMAIN}"
    else
        NODE_PORT=$(kubectl get svc k8soperation -n "$NAMESPACE" -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null)
        if [ -n "$NODE_PORT" ]; then
            NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null)
            echo -e "  NodePort:  http://${NODE_IP}:${NODE_PORT}"
        fi
        echo -e "  Port-forward: kubectl port-forward svc/k8soperation -n $NAMESPACE 8080:8080"
    fi
    echo ""
    echo -e "  ${CYAN}Useful commands:${NC}"
    echo -e "    kubectl get pods -n $NAMESPACE"
    echo -e "    kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=k8soperation -f"
    echo -e "    kubectl exec -it -n $NAMESPACE deploy/k8soperation -- sh"
    echo -e "    kubectl delete -k ${DEPLOY_DIR}/"
    echo ""
else
    warn "Deployment not ready within 180s. Check:"
    kubectl get pods -n "$NAMESPACE"
    echo ""
    echo "  kubectl describe pod -n $NAMESPACE -l app.kubernetes.io/name=k8soperation"
    echo "  kubectl logs -n $NAMESPACE -l app.kubernetes.io/name=k8soperation"
fi
