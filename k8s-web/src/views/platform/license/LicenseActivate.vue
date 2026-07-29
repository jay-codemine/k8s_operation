<template>
  <div class="license-page">
    <div class="license-card">
      <!-- 头部 -->
      <div class="card-head">
        <div class="logo-badge">🔐</div>
        <h1 class="title">平台授权激活</h1>
        <p class="subtitle">K8s Operation Platform · License Activation</p>
      </div>

      <!-- 当前授权状态 -->
      <div class="status-banner" :class="status.licensed ? 'ok' : 'bad'">
        <span class="status-dot"></span>
        <template v-if="status.licensed">
          已授权 · {{ status.licensee || '-' }} · {{ editionText }} · 有效期至 {{ expireText }}
        </template>
        <template v-else>
          {{ status.reason || '平台未授权，请激活 License' }}
        </template>
      </div>

      <!-- 机器码 -->
      <div class="field-block">
        <div class="field-label">
          <span>本机机器码</span>
          <span class="field-hint">将机器码发送给软件供应商以申请 License</span>
        </div>
        <div class="machine-row">
          <code class="machine-id">{{ status.machine_id || '加载中...' }}</code>
          <button class="btn-copy" @click="copyMachineId">{{ copied ? '✓ 已复制' : '复制' }}</button>
        </div>
      </div>

      <!-- License 输入 -->
      <div class="field-block">
        <div class="field-label">
          <span>License 授权码</span>
          <span class="field-hint">粘贴完整的 K8SOP-LICENSE 文本</span>
        </div>
        <textarea
          v-model="licenseText"
          class="license-input"
          rows="5"
          placeholder="K8SOP-LICENSE.xxxxx.xxxxx"
          spellcheck="false"
        ></textarea>
      </div>

      <!-- 激活结果 -->
      <div v-if="errorMsg" class="result-msg error">{{ errorMsg }}</div>
      <div v-if="successMsg" class="result-msg success">{{ successMsg }}</div>

      <!-- 操作 -->
      <button class="btn-activate" :disabled="activating || !licenseText.trim()" @click="activate">
        {{ activating ? '激活中...' : '激活 License' }}
      </button>

      <div class="card-footer">
        <a class="link" @click="goLogin">← 返回登录</a>
        <span class="footer-tip">激活成功后即可正常使用平台全部功能</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import http from '@/api/http'

const status = ref({})
const licenseText = ref('')
const activating = ref(false)
const copied = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

const editionText = computed(() => {
  const map = { enterprise: '企业版', professional: '专业版', standard: '标准版', trial: '试用版' }
  return map[status.value.edition] || status.value.edition || '-'
})

const expireText = computed(() => {
  if (!status.value.expire_at) return '永久'
  return new Date(status.value.expire_at * 1000).toLocaleDateString('zh-CN')
})

const loadStatus = async () => {
  try {
    const res = await http.get('/api/v1/platform/license/status', { _silent: true })
    status.value = res?.data || {}
  } catch (e) {
    console.error('获取授权状态失败', e)
  }
}

const copyMachineId = async () => {
  const id = status.value.machine_id
  if (!id) return
  try {
    await navigator.clipboard.writeText(id)
  } catch {
    // 非 https 环境降级
    const ta = document.createElement('textarea')
    ta.value = id
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
  }
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}

const activate = async () => {
  errorMsg.value = ''
  successMsg.value = ''
  activating.value = true
  try {
    const res = await http.post(
      '/api/v1/platform/license/activate',
      { license: licenseText.value.trim() },
      { _silent: true }
    )
    status.value = res?.data || {}
    successMsg.value = '激活成功！2 秒后跳转...'
    setTimeout(() => {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token')
      window.location.assign(token ? '/dashboard' : '/login')
    }, 2000)
  } catch (e) {
    const msg = (Array.isArray(e?.details) && e.details[0]) || e?.msg || e?.message || '激活失败，请检查 License 是否正确'
    errorMsg.value = msg
  } finally {
    activating.value = false
  }
}

const goLogin = () => window.location.assign('/login')

onMounted(loadStatus)
</script>

<style scoped>
.license-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(ellipse at 20% 0%, #1c2b4a 0%, #12141f 55%, #0d0f17 100%);
  padding: 24px;
}

.license-card {
  width: 100%;
  max-width: 560px;
  background: #1e2030;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 40px 44px 32px;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.45);
}

.card-head {
  text-align: center;
  margin-bottom: 24px;
}

.logo-badge {
  width: 56px;
  height: 56px;
  margin: 0 auto 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 26px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(78, 143, 247, 0.25), rgba(78, 143, 247, 0.08));
  border: 1px solid rgba(78, 143, 247, 0.35);
}

.title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #eef1f8;
  letter-spacing: 1px;
}

.subtitle {
  margin: 6px 0 0;
  font-size: 12px;
  color: #6b7390;
  letter-spacing: 0.5px;
}

.status-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  margin-bottom: 22px;
}

.status-banner.ok {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  color: #4ade80;
}

.status-banner.bad {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  color: #fbbf24;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  flex-shrink: 0;
}

.field-block {
  margin-bottom: 20px;
}

.field-label {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #b9c0d4;
}

.field-hint {
  font-size: 11px;
  font-weight: 400;
  color: #6b7390;
}

.machine-row {
  display: flex;
  gap: 10px;
}

.machine-id {
  flex: 1;
  padding: 10px 14px;
  background: #151722;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 15px;
  letter-spacing: 1.5px;
  color: #4e8ff7;
  user-select: all;
}

.btn-copy {
  padding: 0 18px;
  background: rgba(78, 143, 247, 0.12);
  border: 1px solid rgba(78, 143, 247, 0.4);
  border-radius: 8px;
  color: #4e8ff7;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-copy:hover {
  background: rgba(78, 143, 247, 0.22);
}

.license-input {
  width: 100%;
  box-sizing: border-box;
  padding: 12px 14px;
  background: #151722;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #dde2ee;
  font-family: 'JetBrains Mono', Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  resize: vertical;
  outline: none;
  transition: border-color 0.2s;
  word-break: break-all;
}

.license-input:focus {
  border-color: #4e8ff7;
  box-shadow: 0 0 0 3px rgba(78, 143, 247, 0.15);
}

.license-input::placeholder {
  color: #4a5068;
}

.result-msg {
  padding: 10px 14px;
  border-radius: 8px;
  font-size: 13px;
  margin-bottom: 16px;
  word-break: break-all;
}

.result-msg.error {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #f87171;
}

.result-msg.success {
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.35);
  color: #4ade80;
}

.btn-activate {
  width: 100%;
  padding: 12px 0;
  background: linear-gradient(135deg, #4e8ff7, #3b6fd9);
  border: none;
  border-radius: 8px;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 2px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-activate:hover:not(:disabled) {
  filter: brightness(1.1);
  box-shadow: 0 6px 20px rgba(78, 143, 247, 0.35);
}

.btn-activate:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.card-footer {
  margin-top: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.link {
  font-size: 13px;
  color: #4e8ff7;
  cursor: pointer;
}

.link:hover {
  text-decoration: underline;
}

.footer-tip {
  font-size: 11px;
  color: #6b7390;
}
</style>
