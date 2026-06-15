<template>
  <div class="ldap-settings-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
        </div>
        <div class="header-text">
          <h1>LDAP 认证管理</h1>
          <p>管理 LDAP/Active Directory 集成配置、测试连接和用户同步</p>
        </div>
      </div>
      <div class="header-actions">
        <button class="btn-secondary" @click="refreshStatus">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="1 4 1 10 7 10"/>
            <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
          </svg>
          刷新状态
        </button>
      </div>
    </div>

    <!-- 状态概览卡片 -->
    <div class="status-cards">
      <div class="status-card" :class="config.enabled ? 'active' : 'inactive'">
        <div class="card-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline v-if="config.enabled" points="22 4 12 14.01 9 11.01"/>
            <line v-else x1="18" y1="6" x2="6" y2="18"/>
          </svg>
        </div>
        <div class="card-content">
          <span class="card-label">LDAP 状态</span>
          <span class="card-value">{{ config.enabled ? '已启用' : '未启用' }}</span>
        </div>
      </div>

      <div class="status-card" :class="connectionStatus">
        <div class="card-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
            <polyline points="22,6 12,13 2,6"/>
          </svg>
        </div>
        <div class="card-content">
          <span class="card-label">连接状态</span>
          <span class="card-value">{{ connectionText }}</span>
        </div>
      </div>

      <div class="status-card info">
        <div class="card-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
        </div>
        <div class="card-content">
          <span class="card-label">同步用户数</span>
          <span class="card-value">{{ syncedUsers }}</span>
        </div>
      </div>

      <div class="status-card info">
        <div class="card-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>
        <div class="card-content">
          <span class="card-label">最后同步</span>
          <span class="card-value">{{ lastSync || '从未' }}</span>
        </div>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="content-body">
      <!-- 连接配置区 -->
      <div class="section-card">
        <div class="section-header">
          <h2>连接配置</h2>
          <p>LDAP/AD 服务器连接信息（修改需重启后端生效）</p>
        </div>

        <div class="config-grid">
          <div class="config-item">
            <label>服务器地址</label>
            <div class="config-value">{{ config.host || '-' }}</div>
          </div>
          <div class="config-item">
            <label>端口</label>
            <div class="config-value">{{ config.port || '-' }}</div>
          </div>
          <div class="config-item">
            <label>TLS 加密</label>
            <div class="config-value">
              <span class="badge" :class="config.use_tls ? 'badge-green' : 'badge-gray'">
                {{ config.use_tls ? '已启用' : '未启用' }}
              </span>
            </div>
          </div>
          <div class="config-item">
            <label>Bind DN</label>
            <div class="config-value mono">{{ config.bind_dn || '-' }}</div>
          </div>
          <div class="config-item">
            <label>Base DN</label>
            <div class="config-value mono">{{ config.base_dn || '-' }}</div>
          </div>
          <div class="config-item">
            <label>用户过滤器</label>
            <div class="config-value mono">{{ config.user_filter || '-' }}</div>
          </div>
          <div class="config-item">
            <label>本地回退</label>
            <div class="config-value">
              <span class="badge" :class="config.local_fallback ? 'badge-green' : 'badge-gray'">
                {{ config.local_fallback ? '已启用' : '未启用' }}
              </span>
            </div>
          </div>
          <div class="config-item">
            <label>自动创建用户</label>
            <div class="config-value">
              <span class="badge" :class="config.auto_create ? 'badge-green' : 'badge-gray'">
                {{ config.auto_create ? '已启用' : '未启用' }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 操作区 -->
      <div class="section-card">
        <div class="section-header">
          <h2>操作</h2>
          <p>测试连接和同步 LDAP 用户到平台</p>
        </div>

        <div class="action-group">
          <div class="action-item">
            <div class="action-info">
              <div class="action-icon blue">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                  <polyline points="22 4 12 14.01 9 11.01"/>
                </svg>
              </div>
              <div class="action-text">
                <span class="action-title">测试连接</span>
                <span class="action-desc">验证 LDAP 服务器连接和 Bind 凭据是否正确</span>
              </div>
            </div>
            <button class="btn-primary" @click="handleTestConnection" :disabled="testing">
              {{ testing ? '测试中...' : '测试连接' }}
            </button>
          </div>

          <div class="action-item">
            <div class="action-info">
              <div class="action-icon purple">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polyline points="1 4 1 10 7 10"/>
                  <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"/>
                </svg>
              </div>
              <div class="action-text">
                <span class="action-title">全量同步</span>
                <span class="action-desc">从 LDAP 拉取所有用户并同步到平台（自动匹配组-角色映射）</span>
              </div>
            </div>
            <button class="btn-primary" @click="handleSyncUsers" :disabled="syncing">
              {{ syncing ? '同步中...' : '立即同步' }}
            </button>
          </div>
        </div>

        <!-- 操作结果 -->
        <div v-if="actionResult" class="action-result" :class="actionResult.type">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
            <path v-if="actionResult.type === 'success'" d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline v-if="actionResult.type === 'success'" points="22 4 12 14.01 9 11.01"/>
            <circle v-if="actionResult.type === 'error'" cx="12" cy="12" r="10"/>
            <line v-if="actionResult.type === 'error'" x1="15" y1="9" x2="9" y2="15"/>
            <line v-if="actionResult.type === 'error'" x1="9" y1="9" x2="15" y2="15"/>
          </svg>
          <span>{{ actionResult.message }}</span>
        </div>
      </div>

      <!-- 组-角色映射区 -->
      <div class="section-card">
        <div class="section-header">
          <h2>LDAP 组 → 平台角色映射</h2>
          <p>定义 LDAP 组与平台 RBAC 角色的映射关系（通过 config.yaml 配置）</p>
        </div>

        <div class="mapping-table" v-if="config.group_role_mapping && config.group_role_mapping.length > 0">
          <div class="table-header">
            <div class="col-ldap">LDAP 组</div>
            <div class="col-role">平台角色</div>
            <div class="col-cluster">集群域</div>
            <div class="col-cicd">发布域</div>
          </div>
          <div class="table-row" v-for="(mapping, idx) in config.group_role_mapping" :key="idx">
            <div class="col-ldap">
              <span class="group-tag">{{ mapping.ldap_group }}</span>
            </div>
            <div class="col-role">
              <span class="role-tag">{{ mapping.platform_role }}</span>
            </div>
            <div class="col-cluster">
              <span class="access-tag" :class="'access-' + mapping.cluster_access">
                {{ mapping.cluster_access || 'none' }}
              </span>
            </div>
            <div class="col-cicd">
              <span class="access-tag" :class="'access-' + mapping.cicd_access">
                {{ mapping.cicd_access || 'none' }}
              </span>
            </div>
          </div>
        </div>

        <div class="empty-state" v-else>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="48" height="48">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
            <circle cx="9" cy="7" r="4"/>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
          </svg>
          <p>暂无映射配置</p>
          <span>请在 config.yaml 中配置 LDAP.GroupRoleMapping 字段</span>
        </div>
      </div>

      <!-- 配置提示 -->
      <div class="section-card hint-card">
        <div class="section-header">
          <h2>配置说明</h2>
        </div>
        <div class="hint-content">
          <p>LDAP 配置位于后端 <code>configs/config.yaml</code> 的 <code>LDAP</code> 段：</p>
          <pre class="config-example">LDAP:
  Enabled: true
  Host: "ldap.example.com"
  Port: 389
  UseTLS: false
  BindDN: "cn=admin,dc=example,dc=com"
  BindPassword: "password"
  BaseDN: "ou=people,dc=example,dc=com"
  UserFilter: "(uid={0})"
  AutoCreate: true
  LocalFallback: true
  GroupRoleMapping:
    - LDAPGroup: "devops-team"
      PlatformRole: "devops"
      ClusterAccess: "admin"
      CICDAccess: "admin"
    - LDAPGroup: "backend-dev"
      PlatformRole: "developer"
      ClusterAccess: "write"
      CICDAccess: "write"</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getLDAPConfig, testLDAPConnection, syncLDAPUsers, getLDAPStatus } from '@/api/ldap'

const config = ref({})
const testing = ref(false)
const syncing = ref(false)
const actionResult = ref(null)
const syncedUsers = ref(0)
const lastSync = ref('')
const connectionStatus = ref('unknown') // 'active' | 'inactive' | 'unknown'

const connectionText = computed(() => {
  if (connectionStatus.value === 'active') return '已连接'
  if (connectionStatus.value === 'inactive') return '未连接'
  return '未知'
})

const loadConfig = async () => {
  try {
    const res = await getLDAPConfig()
    const data = res?.data ?? res
    config.value = data || {}
  } catch (e) {
    config.value = {}
  }
}

const loadStatus = async () => {
  try {
    const res = await getLDAPStatus()
    const data = res?.data ?? res
    connectionStatus.value = data?.connected ? 'active' : 'inactive'
    syncedUsers.value = data?.synced_users || 0
    lastSync.value = data?.last_sync || ''
  } catch (e) {
    connectionStatus.value = 'unknown'
  }
}

const refreshStatus = () => {
  loadConfig()
  loadStatus()
}

const handleTestConnection = async () => {
  testing.value = true
  actionResult.value = null
  try {
    const res = await testLDAPConnection()
    const data = res?.data ?? res
    if (data?.success || res?.code === 0) {
      actionResult.value = { type: 'success', message: 'LDAP 连接成功！服务器响应正常。' }
      connectionStatus.value = 'active'
    } else {
      actionResult.value = { type: 'error', message: data?.message || res?.msg || '连接失败' }
      connectionStatus.value = 'inactive'
    }
  } catch (e) {
    actionResult.value = { type: 'error', message: e?.msg || e?.message || '连接测试失败' }
    connectionStatus.value = 'inactive'
  } finally {
    testing.value = false
  }
}

const handleSyncUsers = async () => {
  syncing.value = true
  actionResult.value = null
  try {
    const res = await syncLDAPUsers()
    const data = res?.data ?? res
    const count = data?.synced_count || data?.count || 0
    actionResult.value = {
      type: 'success',
      message: `同步完成！共同步 ${count} 名用户。`
    }
    syncedUsers.value = count
    lastSync.value = new Date().toLocaleString()
  } catch (e) {
    actionResult.value = { type: 'error', message: e?.msg || e?.message || '同步失败' }
  } finally {
    syncing.value = false
  }
}

onMounted(() => {
  loadConfig()
  loadStatus()
})
</script>

<style scoped>
.ldap-settings-page {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(139, 92, 246, 0.15));
  display: flex;
  align-items: center;
  justify-content: center;
  color: #818cf8;
}

.header-icon svg {
  width: 24px;
  height: 24px;
}

.header-text h1 {
  font-size: 22px;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0;
}

.header-text p {
  font-size: 13px;
  color: #94a3b8;
  margin: 4px 0 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.btn-primary, .btn-secondary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(99, 102, 241, 0.4);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.06);
  color: #cbd5e1;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #f1f5f9;
}

.btn-secondary svg {
  width: 16px;
  height: 16px;
}

/* 状态卡片 */
.status-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
  margin-bottom: 24px;
}

.status-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 18px 20px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: all 0.2s;
}

.status-card.active { border-color: rgba(34, 197, 94, 0.3); background: rgba(34, 197, 94, 0.06); }
.status-card.inactive { border-color: rgba(239, 68, 68, 0.3); background: rgba(239, 68, 68, 0.06); }
.status-card.info { border-color: rgba(99, 102, 241, 0.2); }
.status-card.unknown { border-color: rgba(251, 191, 36, 0.3); background: rgba(251, 191, 36, 0.06); }

.card-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.status-card.active .card-icon { background: rgba(34, 197, 94, 0.15); color: #4ade80; }
.status-card.inactive .card-icon { background: rgba(239, 68, 68, 0.15); color: #f87171; }
.status-card.info .card-icon { background: rgba(99, 102, 241, 0.15); color: #818cf8; }
.status-card.unknown .card-icon { background: rgba(251, 191, 36, 0.15); color: #fbbf24; }

.card-icon svg { width: 20px; height: 20px; }

.card-content { display: flex; flex-direction: column; }
.card-label { font-size: 12px; color: #94a3b8; }
.card-value { font-size: 16px; font-weight: 600; color: #f1f5f9; margin-top: 2px; }

/* 内容区 */
.content-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-card {
  padding: 24px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.section-header {
  margin-bottom: 20px;
}

.section-header h2 {
  font-size: 16px;
  font-weight: 600;
  color: #f1f5f9;
  margin: 0;
}

.section-header p {
  font-size: 13px;
  color: #94a3b8;
  margin: 4px 0 0;
}

/* 配置网格 */
.config-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.config-item {
  padding: 14px 16px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.config-item label {
  font-size: 12px;
  color: #94a3b8;
  display: block;
  margin-bottom: 6px;
}

.config-value {
  font-size: 14px;
  color: #e2e8f0;
  word-break: break-all;
}

.config-value.mono {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: #a5b4fc;
}

.badge {
  display: inline-flex;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.badge-green { background: rgba(34, 197, 94, 0.12); color: #86efac; }
.badge-gray { background: rgba(148, 163, 184, 0.12); color: #94a3b8; }

/* 操作区 */
.action-group {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.action-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.action-info {
  display: flex;
  align-items: center;
  gap: 14px;
}

.action-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.action-icon.blue { background: rgba(59, 130, 246, 0.15); color: #60a5fa; }
.action-icon.purple { background: rgba(139, 92, 246, 0.15); color: #a78bfa; }

.action-icon svg { width: 20px; height: 20px; }

.action-text { display: flex; flex-direction: column; }
.action-title { font-size: 14px; font-weight: 500; color: #f1f5f9; }
.action-desc { font-size: 12px; color: #94a3b8; margin-top: 2px; }

/* 操作结果 */
.action-result {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 16px;
  padding: 14px 18px;
  border-radius: 10px;
  font-size: 13px;
}

.action-result.success {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.25);
  color: #86efac;
}

.action-result.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.25);
  color: #fca5a5;
}

/* 映射表格 */
.mapping-table {
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.table-header, .table-row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 1fr 1fr;
  padding: 12px 18px;
}

.table-header {
  background: rgba(255, 255, 255, 0.04);
  font-size: 12px;
  color: #94a3b8;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.table-row {
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  font-size: 13px;
  color: #e2e8f0;
  align-items: center;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.group-tag {
  display: inline-flex;
  padding: 3px 10px;
  border-radius: 6px;
  background: rgba(251, 191, 36, 0.12);
  color: #fcd34d;
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
}

.role-tag {
  display: inline-flex;
  padding: 3px 10px;
  border-radius: 6px;
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
  font-size: 12px;
  font-weight: 500;
}

.access-tag {
  display: inline-flex;
  padding: 3px 8px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
}

.access-none { background: rgba(148, 163, 184, 0.1); color: #94a3b8; }
.access-read { background: rgba(34, 197, 94, 0.1); color: #86efac; }
.access-write { background: rgba(59, 130, 246, 0.1); color: #93c5fd; }
.access-admin { background: rgba(239, 68, 68, 0.1); color: #fca5a5; }

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 40px;
  color: #64748b;
}

.empty-state p {
  font-size: 15px;
  color: #94a3b8;
  margin: 0;
}

.empty-state span {
  font-size: 12px;
}

/* 配置提示 */
.hint-card {
  border-color: rgba(251, 191, 36, 0.15);
  background: rgba(251, 191, 36, 0.03);
}

.hint-content {
  font-size: 13px;
  color: #cbd5e1;
  line-height: 1.8;
}

.hint-content code {
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(99, 102, 241, 0.12);
  color: #a5b4fc;
  font-size: 12px;
}

.config-example {
  margin-top: 12px;
  padding: 16px 20px;
  border-radius: 10px;
  background: rgba(0, 0, 0, 0.3);
  color: #a5b4fc;
  font-size: 12px;
  font-family: 'JetBrains Mono', monospace;
  line-height: 1.8;
  overflow-x: auto;
  white-space: pre;
}

@media (max-width: 768px) {
  .ldap-settings-page { padding: 16px; }
  .page-header { flex-direction: column; align-items: flex-start; gap: 12px; }
  .status-cards { grid-template-columns: 1fr 1fr; }
  .config-grid { grid-template-columns: 1fr; }
  .action-item { flex-direction: column; gap: 12px; align-items: flex-start; }
  .table-header, .table-row { grid-template-columns: 1fr 1fr; gap: 8px; }
}
</style>
