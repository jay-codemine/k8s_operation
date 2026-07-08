#!/usr/bin/env bash
# ============================================================
# CICD 发布问题快速诊断脚本
# 用途: 快速定位发布失败的根本原因
# ============================================================

set -euo pipefail

API_BASE="${API_BASE:-http://127.0.0.1:38180}"
REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6380}"
REDIS_PASS="${REDIS_PASS:-admin123}"

echo "╔══════════════════════════════════════════════════╗"
echo "║     CICD 发布问题快速诊断                        ║"
echo "╚══════════════════════════════════════════════════╝"
echo ""

# 1. 获取 Token
echo "▶ 获取认证 Token..."
TOKEN=$(curl -s -X POST "$API_BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | \
  python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null || echo "")

if [[ -z "$TOKEN" ]]; then
  echo "❌ 登录失败，后端可能未启动"
  exit 1
fi
echo "✅ Token 获取成功"
echo ""

# 2. 查看流水线列表
echo "▶ 查看流水线列表..."
PIPELINES=$(curl -s -X GET "$API_BASE/api/v1/k8s/cicd/pipeline/list?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN")

echo "$PIPELINES" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    pipelines = data.get('data', {}).get('list', [])
    print(f'找到 {len(pipelines)} 条流水线:')
    for p in pipelines:
        print(f\"  ID={p['id']}, 名称={p['name']}, 集群={p.get('target_cluster_id','N/A')}, 镜像={p.get('last_deploy_image','N/A')}\")
except Exception as e:
    print(f'解析失败: {e}')
" 2>/dev/null || echo "无法解析流水线数据"
echo ""

# 3. 检查 Redis Stream
echo "▶ 检查 Redis Stream..."
if command -v redis-cli &>/dev/null; then
  STREAM_LEN=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" --no-auth-warning XLEN cicd:deploy:stream 2>/dev/null || echo "0")
  echo "Stream 长度: $STREAM_LEN"
  
  if [[ "$STREAM_LEN" -gt 0 ]]; then
    echo "最近 3 条消息:"
    redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" --no-auth-warning XRANGE cicd:deploy:stream - + COUNT 3 2>/dev/null | head -20
  fi
else
  echo "redis-cli 未安装"
fi
echo ""

# 4. 查看最近发布单
echo "▶ 查看最近 5 条发布单..."
RELEASES=$(curl -s -X GET "$API_BASE/api/v1/k8s/cicd/release/list?page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN")

echo "$RELEASES" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    releases = data.get('data', {}).get('list', [])
    print(f'找到 {len(releases)} 条发布单:')
    for r in releases:
        print(f\"  ID={r['id']}, 应用={r['app_name']}, 状态={r['status']}, 镜像={r.get('image_repo','')}:{r.get('image_tag','')}\")
        if r.get('message'):
            print(f\"    消息: {r['message']}\")
except Exception as e:
    print(f'解析失败: {e}')
" 2>/dev/null || echo "无法解析发布单数据"
echo ""

# 5. 查看后端日志错误
echo "▶ 最近后端错误日志..."
LOG_FILE="storage/logs/app.log"
if [[ -f "$LOG_FILE" ]]; then
  grep -E "CicdReleaseCreate|panic|error" "$LOG_FILE" | tail -5 | python3 -c "
import sys, json
for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    try:
        data = json.loads(line)
        level = data.get('level', 'unknown')
        msg = data.get('msg', '')
        err = data.get('error', '')
        if err:
            print(f\"  [{level}] {msg}: {err[:100]}\")
        else:
            print(f\"  [{level}] {msg}\")
    except:
        print(f\"  {line[:150]}\")
" 2>/dev/null || echo "日志解析失败"
else
  echo "日志文件不存在: $LOG_FILE"
fi
echo ""

echo "══════════════════════════════════════════════════"
echo "诊断完成。请检查以上输出，常见问题:"
echo "1. 流水线没有 target_cluster_id 或 last_deploy_image"
echo "2. Redis Stream 中有消息但 Worker 未消费"
echo "3. 后端日志中有 panic 或 error"
echo "4. 后端未重启，代码未生效"
echo "══════════════════════════════════════════════════"
