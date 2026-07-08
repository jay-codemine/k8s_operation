#!/usr/bin/env bash
# ============================================================
# CICD 发布管理完整链路验证脚本
# 用途: 验证从创建发布单 → Redis入队 → Worker消费 → K8s滚动更新 → 飞书通知
# 使用: chmod +x scripts/verify-cicd-release.sh && ./scripts/verify-cicd-release.sh
# ============================================================

set -euo pipefail
IFS=$'\n\t'

# ---- 颜色定义 ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

# ---- 配置 ----
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

API_BASE="${API_BASE:-http://127.0.0.1:38180}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6380}"
REDIS_PASS="${REDIS_PASS:-admin123}"

# 测试配置（根据实际情况修改）
TEST_PIPELINE_ID="${TEST_PIPELINE_ID:-}"  # 留空则自动查找
TEST_IMAGE_TAG="${TEST_IMAGE_TAG:-v$(date +%Y%m%d%H%M%S)}"
TEST_NAMESPACE="${TEST_NAMESPACE:-default}"

# ---- 辅助函数 ----
info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
fail()    { echo -e "${RED}[FAIL]${NC} $*"; }
step()    { echo -e "\n${BOLD}${CYAN}▶ $*${NC}"; }

# ---- 清理函数 ----
cleanup() {
    if [[ -n "${TEST_RELEASE_ID:-}" ]]; then
        info "清理测试数据: release_id=$TEST_RELEASE_ID"
        # 可选：删除测试发布单
        # curl -s -X POST "$API_BASE/api/v1/k8s/cicd/release/delete" \
        #   -H "Content-Type: application/json" \
        #   -H "Authorization: Bearer $TOKEN" \
        #   -d "{\"id\": $TEST_RELEASE_ID}" > /dev/null 2>&1
    fi
}
trap cleanup EXIT

# ---- JWT Token 获取 ----
get_token() {
    step "STEP 0: 获取认证 Token"
    
    local login_resp
    login_resp=$(curl -s -X POST "$API_BASE/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}')
    
    TOKEN=$(echo "$login_resp" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [[ -z "$TOKEN" ]]; then
        fail "登录失败，请检查后端是否启动"
        echo "响应: $login_resp"
        exit 1
    fi
    
    success "Token 获取成功"
}

# ---- 检查后端服务 ----
check_backend() {
    step "STEP 1: 检查后端服务"
    
    if curl -s "$API_BASE/api/v1/health" > /dev/null 2>&1; then
        success "后端服务运行正常 ($API_BASE)"
    else
        fail "后端服务不可达 ($API_BASE)"
        info "请先启动后端: go run cmd/k8soperation/main.go"
        exit 1
    fi
}

# ---- 检查 Redis ----
check_redis() {
    step "STEP 2: 检查 Redis 连接"
    
    if command -v redis-cli &>/dev/null; then
        if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" --no-auth-warning ping 2>/dev/null | grep -q PONG; then
            success "Redis 连接正常 ($REDIS_HOST:$REDIS_PORT)"
        else
            fail "Redis 连接失败"
            exit 1
        fi
    else
        warn "redis-cli 未安装，跳过 Redis 检查"
    fi
}

# ---- 查找测试流水线 ----
find_test_pipeline() {
    step "STEP 3: 查找测试流水线"
    
    if [[ -n "$TEST_PIPELINE_ID" ]]; then
        success "使用指定流水线 ID: $TEST_PIPELINE_ID"
        return 0
    fi
    
    info "自动查找可用的流水线..."
    
    local list_resp
    list_resp=$(curl -s -X GET "$API_BASE/api/v1/k8s/cicd/pipeline/list?page=1&page_size=10" \
        -H "Authorization: Bearer $TOKEN")
    
    # 查找第一个有 TargetClusterID 和 LastDeployImage 的流水线
    TEST_PIPELINE_ID=$(echo "$list_resp" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    pipelines = data.get('data', {}).get('list', [])
    for p in pipelines:
        if p.get('target_cluster_id', 0) > 0 and p.get('last_deploy_image'):
            print(p['id'])
            sys.exit(0)
    print('')
except:
    print('')
" 2>/dev/null || echo "")
    
    if [[ -z "$TEST_PIPELINE_ID" ]]; then
        warn "未找到可用的流水线（需要有 target_cluster_id 和 last_deploy_image）"
        info "请手动指定: TEST_PIPELINE_ID=xxx $0"
        exit 1
    fi
    
    success "找到测试流水线 ID: $TEST_PIPELINE_ID"
    
    # 获取流水线详情
    local detail_resp
    detail_resp=$(curl -s -X GET "$API_BASE/api/v1/k8s/cicd/pipeline/detail?id=$TEST_PIPELINE_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    PIPELINE_NAME=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['name'])" 2>/dev/null || echo "unknown")
    TARGET_CLUSTER=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['target_cluster_id'])" 2>/dev/null || echo "0")
    WORKLOAD_NAME=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['target_workload_name'])" 2>/dev/null || echo "unknown")
    CONTAINER_NAME=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['target_container'])" 2>/dev/null || echo "unknown")
    
    info "流水线名称: $PIPELINE_NAME"
    info "目标集群: $TARGET_CLUSTER"
    info "工作负载: $WORKLOAD_NAME"
    info "容器名称: $CONTAINER_NAME"
}

# ---- 创建发布单 ----
create_release() {
    step "STEP 4: 创建发布单"
    
    info "镜像标签: $TEST_IMAGE_TAG"
    
    local create_resp
    create_resp=$(curl -s -X POST "$API_BASE/api/v1/k8s/cicd/release/create" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "{
            \"pipeline_id\": $TEST_PIPELINE_ID,
            \"image_tag\": \"$TEST_IMAGE_TAG\",
            \"namespace\": \"$TEST_NAMESPACE\",
            \"message\": \"自动化测试 - $(date '+%Y-%m-%d %H:%M:%S')\"
        }")
    
    info "创建响应: $create_resp"
    
    TEST_RELEASE_ID=$(echo "$create_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['release_id'])" 2>/dev/null || echo "")
    
    if [[ -z "$TEST_RELEASE_ID" ]]; then
        fail "创建发布单失败"
        exit 1
    fi
    
    success "发布单创建成功: release_id=$TEST_RELEASE_ID"
}

# ---- 检查发布单状态 ----
check_release_status() {
    step "STEP 5: 检查发布单状态（轮询）"
    
    local max_wait=120  # 最多等待 120 秒
    local interval=3    # 每 3 秒检查一次
    local elapsed=0
    
    while [[ $elapsed -lt $max_wait ]]; do
        local detail_resp
        detail_resp=$(curl -s -X GET "$API_BASE/api/v1/k8s/cicd/release/detail?id=$TEST_RELEASE_ID" \
            -H "Authorization: Bearer $TOKEN")
        
        local status
        status=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['release']['status'])" 2>/dev/null || echo "unknown")
        
        local message
        message=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['release']['message'])" 2>/dev/null || echo "")
        
        info "[$elapsed s] 状态: $status | $message"
        
        if [[ "$status" == "Succeeded" ]]; then
            success "✅ 发布成功！"
            return 0
        elif [[ "$status" == "Failed" ]]; then
            fail "❌ 发布失败: $message"
            return 1
        elif [[ "$status" == "Canceled" ]]; then
            warn "⚠️  发布已取消"
            return 1
        fi
        
        sleep $interval
        elapsed=$((elapsed + interval))
    done
    
    fail "⏱️  超时 ($max_wait s)，当前状态: $status"
    return 1
}

# ---- 检查任务状态 ----
check_task_status() {
    step "STEP 6: 检查任务详情"
    
    local tasks_resp
    tasks_resp=$(curl -s -X GET "$API_BASE/api/v1/k8s/cicd/release/tasks?release_id=$TEST_RELEASE_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "$tasks_resp" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    tasks = data.get('data', {}).get('tasks', [])
    print(f'任务数量: {len(tasks)}')
    for t in tasks:
        print(f\"  任务 #{t['id']}: cluster={t['cluster_id']}, status={t['status']}, target_image={t.get('target_image', 'N/A')}\")
        if t.get('message'):
            print(f\"    消息: {t['message']}\")
        if t.get('prev_image'):
            print(f\"    原镜像: {t['prev_image']}\")
except Exception as e:
    print(f'解析失败: {e}')
" 2>/dev/null || echo "无法解析任务数据"
}

# ---- 检查 Redis Stream ----
check_redis_stream() {
    step "STEP 7: 检查 Redis Stream"
    
    if ! command -v redis-cli &>/dev/null; then
        warn "redis-cli 未安装，跳过"
        return 0
    fi
    
    local stream_len
    stream_len=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" --no-auth-warning XLEN cicd:deploy:stream 2>/dev/null || echo "0")
    info "Stream 长度: $stream_len"
    
    local pending_count
    pending_count=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" --no-auth-warning XPENDING cicd:deploy:stream cicd:deploy:group - + 10 2>/dev/null | grep -c "^[0-9]" || echo "0")
    info "待处理消息: $pending_count"
    
    if [[ "$stream_len" == "0" && "$pending_count" == "0" ]]; then
        success "✅ Stream 已清空（所有消息已消费）"
    else
        warn "Stream 中仍有未消费消息（可能 Worker 未启动）"
    fi
}

# ---- 检查后端日志 ----
check_backend_logs() {
    step "STEP 8: 检查后端日志"
    
    local log_file="$ROOT_DIR/storage/logs/app.log"
    
    if [[ ! -f "$log_file" ]]; then
        warn "日志文件不存在: $log_file"
        return 0
    fi
    
    info "最近 20 行相关日志:"
    grep -E "processing task|executeTask|patched deployment|rollout complete|task succeeded|task failed|release_id=$TEST_RELEASE_ID" "$log_file" | tail -20 || echo "未找到相关日志"
}

# ---- 验证 K8s 部署 ----
verify_k8s_deployment() {
    step "STEP 9: 验证 K8s 滚动更新"
    
    if ! command -v kubectl &>/dev/null; then
        warn "kubectl 未安装，跳过 K8s 验证"
        return 0
    fi
    
    # 从发布单详情获取 namespace 和 workload
    local detail_resp
    detail_resp=$(curl -s -X GET "$API_BASE/api/v1/k8s/cicd/release/detail?id=$TEST_RELEASE_ID" \
        -H "Authorization: Bearer $TOKEN")
    
    local namespace workload_name
    namespace=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['release']['namespace'])" 2>/dev/null || echo "")
    workload_name=$(echo "$detail_resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['release']['workload_name'])" 2>/dev/null || echo "")
    
    if [[ -z "$namespace" || -z "$workload_name" ]]; then
        warn "无法获取 namespace 或 workload_name"
        return 0
    fi
    
    info "检查 Deployment: $namespace/$workload_name"
    
    # 获取 Deployment 的镜像
    local current_image
    current_image=$(kubectl get deployment "$workload_name" -n "$namespace" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "error")
    
    if [[ "$current_image" == "error" ]]; then
        fail "Deployment 不存在或无法访问"
        return 1
    fi
    
    success "当前镜像: $current_image"
    
    # 检查是否包含测试的镜像标签
    if echo "$current_image" | grep -q "$TEST_IMAGE_TAG"; then
        success "✅ 镜像已更新为测试版本: $TEST_IMAGE_TAG"
    else
        warn "⚠️  镜像未更新为测试版本（可能是多集群或其他原因）"
    fi
    
    # 检查 Pod 状态
    info "Pod 状态:"
    kubectl get pods -n "$namespace" -l "app=$workload_name" --no-headers 2>/dev/null | head -5 || echo "无法获取 Pod 列表"
}

# ---- 主流程 ----
main() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════╗"
    echo "║        CICD 发布管理完整链路验证脚本 v1.0               ║"
    echo "╚══════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    info "API 地址: $API_BASE"
    info "测试镜像标签: $TEST_IMAGE_TAG"
    info "测试命名空间: $TEST_NAMESPACE"
    echo ""
    
    # 执行验证步骤
    check_backend
    check_redis
    get_token
    find_test_pipeline
    create_release
    check_release_status
    
    # 详细检查
    check_task_status
    check_redis_stream
    check_backend_logs
    verify_k8s_deployment
    
    # 最终总结
    step "验证完成"
    success "发布单 ID: $TEST_RELEASE_ID"
    success "测试镜像: $TEST_IMAGE_TAG"
    info "请检查飞书/钉钉是否收到通知"
    info "请检查 K8s Pod 是否滚动更新"
}

main "$@"
