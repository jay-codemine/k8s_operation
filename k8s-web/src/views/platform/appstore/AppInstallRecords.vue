<template>
  <div class="records-page">
    <!-- ====== 面包屑 ====== -->
    <div class="page-breadcrumb">
      <a-breadcrumb>
        <a-breadcrumb-item>
          <router-link to="/platform/appstore" class="bc-link">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/>
              <rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>
            </svg>
            应用商城
          </router-link>
        </a-breadcrumb-item>
        <a-breadcrumb-item>已部署应用</a-breadcrumb-item>
      </a-breadcrumb>
    </div>

    <!-- ====== 页头 ====== -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">
          <span class="title-icon">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#fff" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="3"/><polyline points="9,11 12,14 22,4"/></svg>
          </span>
          已部署应用
        </h2>
        <span class="record-count" v-if="total > 0">共 {{ total }} 条记录</span>
      </div>
      <div class="header-right">
        <a-select v-model="filterStatus" placeholder="状态筛选" allow-clear style="width:140px" @change="fetchRecords">
          <a-option :value="0">全部状态</a-option>
          <a-option :value="1">安装中</a-option>
          <a-option :value="2">运行中</a-option>
          <a-option :value="3">安装失败</a-option>
          <a-option :value="4">卸载中</a-option>
          <a-option :value="5">已卸载</a-option>
        </a-select>
        <a-button @click="fetchRecords" :loading="loading">
          <template #icon><icon-refresh /></template>
          刷新
        </a-button>
        <a-button type="primary" @click="$router.push('/platform/appstore')">
          <template #icon><icon-arrow-left /></template>
          返回商城
        </a-button>
      </div>
    </div>

    <!-- ====== 统计概览 ====== -->
    <div class="stats-bar">
      <div class="stat-card" :class="{ active: filterStatus === 0 || !filterStatus }" @click="setFilter(0)">
        <div class="stat-num">{{ stats.total }}</div>
        <div class="stat-label">全部</div>
      </div>
      <div class="stat-card running" :class="{ active: filterStatus === 2 }" @click="setFilter(2)">
        <span class="stat-dot running"></span>
        <div class="stat-num">{{ stats.running }}</div>
        <div class="stat-label">运行中</div>
      </div>
      <div class="stat-card installing" :class="{ active: filterStatus === 1 }" @click="setFilter(1)">
        <span class="stat-dot installing"></span>
        <div class="stat-num">{{ stats.installing }}</div>
        <div class="stat-label">安装中</div>
      </div>
      <div class="stat-card failed" :class="{ active: filterStatus === 3 }" @click="setFilter(3)">
        <span class="stat-dot failed"></span>
        <div class="stat-num">{{ stats.failed }}</div>
        <div class="stat-label">失败</div>
      </div>
      <div class="stat-card uninstalled" :class="{ active: filterStatus === 5 }" @click="setFilter(5)">
        <span class="stat-dot uninstalled"></span>
        <div class="stat-num">{{ stats.uninstalled }}</div>
        <div class="stat-label">已卸载</div>
      </div>
    </div>

    <!-- ====== 加载 ====== -->
    <div v-if="loading" class="page-loading">
      <a-spin size="28" /><span>加载部署记录...</span>
    </div>

    <!-- ====== 空状态 ====== -->
    <div v-else-if="records.length === 0" class="page-empty">
      <a-empty description="暂无部署记录">
        <a-button type="primary" size="small" @click="$router.push('/platform/appstore')">去应用商城安装</a-button>
      </a-empty>
    </div>

    <!-- ====== 记录列表 ====== -->
    <div v-else class="records-grid">
      <div
        v-for="record in records"
        :key="record.id"
        class="record-card"
        :class="'s-' + record.status"
        @click="goDetail(record)"
      >
        <!-- 状态角标 -->
        <div class="card-status-bar" :class="'bar-s' + record.status">
          <span class="status-indicator" :class="'ind-s' + record.status"></span>
          {{ statusLabel(record.status) }}
        </div>

        <!-- 应用信息 -->
        <div class="card-app-row">
          <div class="card-icon" :class="getIconClass(record.app_name)">
            {{ getIconLetter(record.app_name) }}
          </div>
          <div class="card-app-info">
            <div class="card-app-name">{{ record.app_name }}</div>
            <div class="card-app-version">
              <a-tag size="small" color="arcoblue">v{{ record.version }}</a-tag>
            </div>
          </div>
        </div>

        <!-- 部署信息 -->
        <div class="card-deploy-info">
          <div class="deploy-row">
            <svg width="13" height="13" viewBox="0 0 14 14" fill="none" stroke="#86909c" stroke-width="1.3"><circle cx="7" cy="7" r="5"/><circle cx="7" cy="7" r="2"/></svg>
            <span class="deploy-label">集群</span>
            <span class="deploy-value">{{ record.cluster_name }}</span>
          </div>
          <div class="deploy-row">
            <svg width="13" height="13" viewBox="0 0 14 14" fill="none" stroke="#86909c" stroke-width="1.3"><rect x="1.5" y="3.5" width="11" height="7" rx="1.5"/><line x1="1.5" y1="6.5" x2="12.5" y2="6.5"/></svg>
            <span class="deploy-label">命名空间</span>
            <span class="deploy-value"><code>{{ record.namespace }}</code></span>
          </div>
          <div class="deploy-row">
            <svg width="13" height="13" viewBox="0 0 14 14" fill="none" stroke="#86909c" stroke-width="1.3"><path d="M2.5,11 L7,2.5 L11.5,11 Z"/></svg>
            <span class="deploy-label">Release</span>
            <span class="deploy-value"><code>{{ record.release_name }}</code></span>
          </div>
        </div>

        <!-- 消息 -->
        <div v-if="record.message" class="card-message" :class="{ error: record.status === 3 }">
          {{ record.message }}
        </div>

        <!-- 底部操作 -->
        <div class="card-footer">
          <span class="card-time">{{ formatTimestamp(record.created_at) }}</span>
          <div class="card-actions">
            <a-button size="mini" type="primary" @click.stop="goDetail(record)">
              <template #icon><icon-eye /></template>
              详情
            </a-button>
            <a-button
              v-if="record.status === 2 || record.status === 3"
              size="mini" type="outline" status="danger"
              @click.stop="handleUninstall(record)"
            >
              <template #icon><icon-delete /></template>
              卸载
            </a-button>
          </div>
        </div>
      </div>
    </div>

    <!-- ====== 分页 ====== -->
    <div v-if="total > pageSize" class="pagination-wrap">
      <a-pagination
        v-model:current="currentPage"
        :total="total"
        :page-size="pageSize"
        show-total
        show-jumper
        @change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Message, Modal } from '@arco-design/web-vue'
import {
  IconRefresh, IconArrowLeft, IconEye, IconDelete
} from '@arco-design/web-vue/es/icon'
import { getInstallList, uninstallApp } from '@/api/platform/appstore.js'

const router = useRouter()

// ====== 状态 ======
const loading = ref(false)
const records = ref([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = 20
const filterStatus = ref(0)

// 统计
const allRecords = ref([])
const stats = computed(() => {
  const list = allRecords.value
  return {
    total: list.length,
    running: list.filter(r => r.status === 2).length,
    installing: list.filter(r => r.status === 1).length,
    failed: list.filter(r => r.status === 3).length,
    uninstalled: list.filter(r => r.status === 5).length,
  }
})

// ====== 数据加载 ======
async function fetchRecords() {
  loading.value = true
  try {
    const params = { page: currentPage.value, page_size: pageSize }
    if (filterStatus.value > 0) {
      params.status = filterStatus.value
    }
    const res = await getInstallList(params)
    records.value = res?.data?.list || []
    total.value = res?.data?.total || 0
  } catch (e) {
    console.error('获取安装记录失败:', e)
    records.value = []
  } finally {
    loading.value = false
  }
}

async function fetchAllStats() {
  try {
    const res = await getInstallList({ page: 1, page_size: 200 })
    allRecords.value = res?.data?.list || []
  } catch { /* ignore */ }
}

function setFilter(s) {
  filterStatus.value = s
  currentPage.value = 1
  fetchRecords()
}

function handlePageChange(page) {
  currentPage.value = page
  fetchRecords()
}

// ====== 操作 ======
function goDetail(record) {
  router.push(`/platform/appstore/install/${record.id}`)
}

function handleUninstall(record) {
  Modal.warning({
    title: '确认卸载',
    content: `确定要卸载「${record.app_name}」(${record.release_name} @ ${record.cluster_name}/${record.namespace})？`,
    okText: '确认卸载',
    cancelText: '取消',
    hideCancel: false,
    onOk: async () => {
      try {
        await uninstallApp(record.id)
        Message.success('卸载成功')
        fetchRecords()
        fetchAllStats()
      } catch (e) {
        Message.error('卸载失败: ' + (e?.msg || e?.message || ''))
      }
    }
  })
}

// ====== 工具 ======
function statusLabel(s) {
  const m = { 1: '安装中', 2: '运行中', 3: '安装失败', 4: '卸载中', 5: '已卸载' }
  return m[s] || '未知'
}

function formatTimestamp(ts) {
  if (!ts) return ''
  try { return new Date(ts * 1000).toLocaleString('zh-CN') } catch { return '' }
}

const iconColorMap = {
  'ingress': 'ic-blue', 'nginx': 'ic-green', 'prometheus': 'ic-orange',
  'grafana': 'ic-purple', 'argocd': 'ic-teal', 'elasticsearch': 'ic-yellow',
  'cert-manager': 'ic-red', 'metallb': 'ic-indigo', 'redis': 'ic-red',
  'mysql': 'ic-blue', 'kafka': 'ic-dark', 'harbor': 'ic-teal',
  'loki': 'ic-orange', 'longhorn': 'ic-green'
}
function getIconClass(name) {
  if (!name) return 'ic-blue'
  const n = name.toLowerCase()
  for (const [key, cls] of Object.entries(iconColorMap)) {
    if (n.includes(key)) return cls
  }
  const colors = ['ic-blue', 'ic-green', 'ic-orange', 'ic-purple', 'ic-teal', 'ic-red']
  let hash = 0
  for (let i = 0; i < n.length; i++) hash = n.charCodeAt(i) + ((hash << 5) - hash)
  return colors[Math.abs(hash) % colors.length]
}
function getIconLetter(name) {
  if (!name) return '?'
  return name.charAt(0).toUpperCase()
}

// ====== 生命周期 ======
onMounted(() => {
  fetchRecords()
  fetchAllStats()
})
</script>

<style scoped>
.records-page { padding: 0; }

/* 面包屑 */
.page-breadcrumb { margin-bottom: 16px; }
.bc-link { display: inline-flex; align-items: center; gap: 6px; color: #4e5969; text-decoration: none; }
.bc-link:hover { color: #326ce5; }

/* 页头 */
.page-header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px;
}
.header-left { display: flex; align-items: baseline; gap: 12px; }
.page-title {
  margin: 0; font-size: 20px; font-weight: 700; color: #1d2129;
  display: flex; align-items: center; gap: 10px;
}
.title-icon {
  width: 34px; height: 34px; border-radius: 10px; display: flex; align-items: center; justify-content: center;
  background: linear-gradient(135deg, #326ce5, #5b8ff9);
}
.record-count { font-size: 13px; color: #86909c; }
.header-right { display: flex; gap: 10px; align-items: center; }

/* 统计栏 */
.stats-bar {
  display: flex; gap: 12px; margin-bottom: 20px;
  padding: 16px 20px; background: #f7f8fa; border-radius: 12px;
}
.stat-card {
  flex: 1; display: flex; flex-direction: column; align-items: center; gap: 4px;
  padding: 12px 8px; border-radius: 10px; cursor: pointer;
  background: #fff; border: 2px solid transparent; transition: all 0.2s;
  position: relative;
}
.stat-card:hover { border-color: #c9cdd4; }
.stat-card.active { border-color: #326ce5; background: #f0f5ff; }
.stat-card.running.active { border-color: #00b42a; background: #f0faf2; }
.stat-card.failed.active { border-color: #f53f3f; background: #fff5f3; }
.stat-card.installing.active { border-color: #326ce5; background: #f0f5ff; }
.stat-card.uninstalled.active { border-color: #86909c; background: #f7f8fa; }
.stat-num { font-size: 24px; font-weight: 700; color: #1d2129; line-height: 1; }
.stat-label { font-size: 12px; color: #86909c; }
.stat-dot {
  width: 8px; height: 8px; border-radius: 50%; position: absolute; top: 10px; right: 10px;
}
.stat-dot.running { background: #00b42a; }
.stat-dot.installing { background: #326ce5; animation: pulse-fade 1.5s infinite; }
.stat-dot.failed { background: #f53f3f; }
.stat-dot.uninstalled { background: #c0c4cc; }
@keyframes pulse-fade { 0%,100% { opacity:1; } 50% { opacity:0.3; } }

/* 加载/空 */
.page-loading, .page-empty {
  display: flex; flex-direction: column; align-items: center; gap: 12px;
  padding: 80px 0; color: #86909c;
}

/* 记录网格 */
.records-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(380px, 1fr)); gap: 16px;
}

/* 记录卡片 */
.record-card {
  background: #fff; border: 1px solid #e5e6eb; border-radius: 12px;
  padding: 0; cursor: pointer; transition: all 0.25s; overflow: hidden;
  display: flex; flex-direction: column;
}
.record-card:hover { border-color: #c9cdd4; box-shadow: 0 4px 20px rgba(0,0,0,0.08); transform: translateY(-2px); }
.record-card.s-2 { border-top: 3px solid #00b42a; }
.record-card.s-3 { border-top: 3px solid #f53f3f; }
.record-card.s-1 { border-top: 3px solid #326ce5; }
.record-card.s-4 { border-top: 3px solid #ff7d00; }
.record-card.s-5 { border-top: 3px solid #c0c4cc; }

/* 状态栏 */
.card-status-bar {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; font-size: 12px; font-weight: 600;
  background: #f7f8fa; color: #4e5969;
}
.card-status-bar.bar-s2 { background: #f0faf2; color: #00b42a; }
.card-status-bar.bar-s3 { background: #fff5f3; color: #f53f3f; }
.card-status-bar.bar-s1 { background: #f0f5ff; color: #326ce5; }
.card-status-bar.bar-s4 { background: #fff8f0; color: #ff7d00; }
.card-status-bar.bar-s5 { background: #f7f8fa; color: #86909c; }
.status-indicator {
  width: 7px; height: 7px; border-radius: 50%;
}
.ind-s2 { background: #00b42a; }
.ind-s3 { background: #f53f3f; }
.ind-s1 { background: #326ce5; animation: pulse-fade 1.5s infinite; }
.ind-s4 { background: #ff7d00; animation: pulse-fade 1.5s infinite; }
.ind-s5 { background: #c0c4cc; }

/* 应用行 */
.card-app-row {
  display: flex; align-items: center; gap: 14px; padding: 16px 16px 12px;
}
.card-icon {
  width: 44px; height: 44px; border-radius: 12px; display: flex; align-items: center;
  justify-content: center; font-size: 20px; font-weight: 700; color: #fff; flex-shrink: 0;
}
.ic-blue { background: linear-gradient(135deg, #326ce5, #5b8ff9); }
.ic-green { background: linear-gradient(135deg, #00b42a, #23c343); }
.ic-orange { background: linear-gradient(135deg, #e8740c, #f59e0b); }
.ic-purple { background: linear-gradient(135deg, #722ed1, #9254de); }
.ic-teal { background: linear-gradient(135deg, #0fc6c2, #14c9c9); }
.ic-red { background: linear-gradient(135deg, #f53f3f, #f76560); }
.ic-yellow { background: linear-gradient(135deg, #e8b900, #fadb14); }
.ic-indigo { background: linear-gradient(135deg, #3730a3, #6366f1); }
.ic-dark { background: linear-gradient(135deg, #374151, #6b7280); }
.card-app-name { font-size: 16px; font-weight: 600; color: #1d2129; }
.card-app-version { margin-top: 4px; }

/* 部署信息 */
.card-deploy-info { padding: 0 16px 12px; }
.deploy-row {
  display: flex; align-items: center; gap: 6px; padding: 4px 0;
  font-size: 13px; color: #4e5969;
}
.deploy-label { color: #86909c; min-width: 52px; }
.deploy-value { font-weight: 500; }
.deploy-value code {
  font-size: 12px; background: #f2f3f5; padding: 1px 6px; border-radius: 4px;
  font-family: 'JetBrains Mono', monospace;
}

/* 消息 */
.card-message {
  margin: 0 16px 12px; padding: 8px 12px; font-size: 12px;
  color: #86909c; background: #f7f8fa; border-radius: 6px;
  line-height: 1.5; overflow: hidden; text-overflow: ellipsis;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;
}
.card-message.error { background: #fff2f0; color: #f53f3f; }

/* 底部 */
.card-footer {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 16px; border-top: 1px solid #f2f3f5; margin-top: auto;
}
.card-time { font-size: 12px; color: #c0c4cc; }
.card-actions { display: flex; gap: 6px; }

/* 分页 */
.pagination-wrap { display: flex; justify-content: center; margin-top: 24px; }

/* 响应式 */
@media (max-width: 768px) {
  .records-grid { grid-template-columns: 1fr; }
  .stats-bar { flex-wrap: wrap; }
  .stat-card { min-width: 80px; }
  .page-header { flex-direction: column; gap: 12px; align-items: flex-start; }
}
</style>
