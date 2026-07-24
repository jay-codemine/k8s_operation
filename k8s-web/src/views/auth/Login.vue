<template>
  <div class="login-page">
    <!-- 插画背景：浅蓝天空 + 港口云原生场景 -->
    <div class="scene" aria-hidden="true">
      <div class="cloud cloud-1"></div>
      <div class="cloud cloud-2"></div>
      <div class="cloud cloud-3"></div>

      <!-- 港口插画剪影：城市天际线 + 龙门起重机 + 集装箱 + 货轮 + 海面 -->
      <svg class="harbor" viewBox="0 0 1440 460" preserveAspectRatio="xMidYMax slice">
        <!-- 远景城市天际线 -->
        <g fill="#c3daf0">
          <rect x="60" y="150" width="46" height="200" rx="3"/>
          <rect x="118" y="110" width="34" height="240" rx="3"/>
          <rect x="162" y="180" width="40" height="170" rx="3"/>
          <rect x="212" y="90" width="30" height="260" rx="3"/>
          <rect x="250" y="200" width="48" height="150" rx="3"/>
          <rect x="1180" y="170" width="42" height="180" rx="3"/>
          <rect x="1230" y="120" width="34" height="230" rx="3"/>
          <rect x="1272" y="190" width="46" height="160" rx="3"/>
          <rect x="1326" y="150" width="30" height="200" rx="3"/>
        </g>
        <!-- 中景高塔 -->
        <g fill="#a7c8e8">
          <rect x="196" y="60" width="14" height="290"/>
          <polygon points="203,30 190,72 216,72"/>
          <rect x="1256" y="70" width="12" height="280"/>
          <polygon points="1262,44 1250,82 1274,82"/>
        </g>

        <!-- 左侧龙门起重机 -->
        <g stroke="#6aa2da" stroke-width="6" fill="none" stroke-linecap="round">
          <line x1="330" y1="120" x2="330" y2="360"/>
          <line x1="400" y1="120" x2="400" y2="360"/>
          <line x1="300" y1="122" x2="470" y2="122"/>
          <line x1="330" y1="150" x2="365" y2="120"/>
          <line x1="400" y1="150" x2="365" y2="120"/>
        </g>
        <line x1="455" y1="122" x2="455" y2="205" stroke="#6aa2da" stroke-width="4"/>
        <rect x="444" y="205" width="22" height="18" rx="2" fill="#e89b5a"/>

        <!-- 右侧龙门起重机 -->
        <g stroke="#5a95d0" stroke-width="6" fill="none" stroke-linecap="round">
          <line x1="1040" y1="90" x2="1040" y2="360"/>
          <line x1="1120" y1="90" x2="1120" y2="360"/>
          <line x1="1010" y1="92" x2="1180" y2="92"/>
          <line x1="1040" y1="124" x2="1080" y2="92"/>
          <line x1="1120" y1="124" x2="1080" y2="92"/>
        </g>
        <line x1="1160" y1="92" x2="1160" y2="180" stroke="#5a95d0" stroke-width="4"/>
        <rect x="1149" y="180" width="22" height="18" rx="2" fill="#f0b36b"/>

        <!-- 集装箱堆场 -->
        <g>
          <rect x="470" y="300" width="70" height="26" rx="2" fill="#4a86c9"/>
          <rect x="470" y="272" width="70" height="26" rx="2" fill="#e89b5a"/>
          <rect x="544" y="300" width="70" height="26" rx="2" fill="#6cc0b0"/>
          <rect x="544" y="272" width="70" height="26" rx="2" fill="#4a86c9"/>
          <rect x="618" y="300" width="70" height="26" rx="2" fill="#e89b5a"/>
          <rect x="880" y="300" width="70" height="26" rx="2" fill="#6cc0b0"/>
          <rect x="880" y="272" width="70" height="26" rx="2" fill="#4a86c9"/>
          <rect x="954" y="300" width="70" height="26" rx="2" fill="#e89b5a"/>
        </g>

        <!-- 右侧货轮 -->
        <g>
          <path d="M1120 372 L1400 372 L1360 418 L1150 418 Z" fill="#3f6fa8"/>
          <rect x="1150" y="336" width="150" height="38" fill="#5a8bc4"/>
          <rect x="1156" y="308" width="40" height="28" rx="2" fill="#e89b5a"/>
          <rect x="1200" y="308" width="40" height="28" rx="2" fill="#d8e6f2"/>
          <rect x="1244" y="308" width="40" height="28" rx="2" fill="#6cc0b0"/>
          <rect x="1306" y="320" width="26" height="54" fill="#2f5885"/>
          <rect x="1312" y="326" width="14" height="12" fill="#cfe0f0"/>
        </g>

        <!-- 海面 -->
        <path d="M0 372 Q360 356 720 372 T1440 372 L1440 460 L0 460 Z" fill="#7fb6e6"/>
        <path d="M0 392 Q360 378 720 392 T1440 392 L1440 460 L0 460 Z" fill="#5f9fd8" opacity="0.85"/>
        <g stroke="#a9d0f0" stroke-width="3" stroke-linecap="round" opacity="0.7">
          <line x1="120" y1="410" x2="180" y2="410"/>
          <line x1="300" y1="428" x2="360" y2="428"/>
          <line x1="560" y1="414" x2="620" y2="414"/>
          <line x1="820" y1="430" x2="880" y2="430"/>
          <line x1="1000" y1="412" x2="1060" y2="412"/>
        </g>
      </svg>
    </div>

    <!-- 居中登录区 -->
    <main class="login-main">
      <!-- 品牌 -->
      <div class="brand">
        <div class="brand-logo">
          <svg viewBox="0 0 64 64" width="40" height="40">
            <circle cx="32" cy="32" r="14" fill="none" stroke="#2f6fd0" stroke-width="4"/>
            <circle cx="32" cy="32" r="5" fill="#2f6fd0"/>
            <g stroke="#2f6fd0" stroke-width="3" stroke-linecap="round">
              <line x1="38" y1="32" x2="54" y2="32"/>
              <line x1="26" y1="32" x2="10" y2="32"/>
              <line x1="32" y1="38" x2="32" y2="54"/>
              <line x1="32" y1="26" x2="32" y2="10"/>
              <line x1="36.2" y1="36.2" x2="47.5" y2="47.5"/>
              <line x1="27.8" y1="36.2" x2="16.5" y2="47.5"/>
              <line x1="27.8" y1="27.8" x2="16.5" y2="16.5"/>
              <line x1="36.2" y1="27.8" x2="47.5" y2="16.5"/>
            </g>
            <g fill="#2f6fd0">
              <circle cx="54" cy="32" r="3"/><circle cx="10" cy="32" r="3"/>
              <circle cx="32" cy="54" r="3"/><circle cx="32" cy="10" r="3"/>
              <circle cx="47.5" cy="47.5" r="3"/><circle cx="16.5" cy="47.5" r="3"/>
              <circle cx="16.5" cy="16.5" r="3"/><circle cx="47.5" cy="16.5" r="3"/>
            </g>
          </svg>
        </div>
        <h1 class="brand-title">K8S OPERATOR</h1>
        <p class="brand-subtitle">云原生应用交付与运维平台</p>
        <div v-if="ldapEnabled && mode === 'login'" class="ldap-badge">
          <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
          <span>LDAP 认证已启用</span>
        </div>
      </div>

      <!-- 白色登录卡片 -->
      <div class="login-card">
        <form class="login-form" @submit.prevent="handleSubmit">
          <div class="field">
            <span class="field-icon">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
              </svg>
            </span>
            <input
              v-model="form.username"
              type="text"
              placeholder="用户名"
              autocomplete="username"
              required
              aria-label="用户名"
            />
          </div>

          <div class="field">
            <span class="field-icon">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
            </span>
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="密码"
              :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
              required
              aria-label="密码"
            />
            <button type="button" class="eye" @click="showPassword = !showPassword" tabindex="-1">
              <svg v-if="!showPassword" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/>
              </svg>
            </button>
          </div>

          <div class="field" v-if="mode === 'register'">
            <span class="field-icon">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
            </span>
            <input
              v-model="form.password_confirm"
              type="password"
              placeholder="确认密码"
              autocomplete="new-password"
              required
              aria-label="确认密码"
            />
          </div>

          <div class="form-row" v-if="mode === 'login'">
            <label class="remember">
              <input type="checkbox" v-model="form.remember" />
              <span>记住我</span>
            </label>
            <a href="#" class="forgot" @click.prevent="openForgot">忘记密码?</a>
          </div>

          <button type="submit" class="submit-btn" :disabled="isLoading">
            <span v-if="!isLoading">{{ mode === 'login' ? '登 录' : '注 册' }}</span>
            <span v-else>{{ mode === 'login' ? '登录中...' : '注册中...' }}</span>
          </button>

          <div class="error-message" v-if="error">{{ error }}</div>
          <div class="success-message" v-if="success">{{ success }}</div>
          <div class="ldap-info" v-if="authMethod === 'ldap' && success">
            <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            已通过 LDAP 认证
          </div>

          <div class="switch-mode">
            <span>{{ mode === 'login' ? '还没有账号？' : '已有账号？' }}</span>
            <a href="#" @click.prevent="toggleMode">{{ mode === 'login' ? '注册新账号' : '返回登录' }}</a>
          </div>

          <div class="card-footer">© 2026 K8S Operator. All rights reserved.</div>
        </form>
      </div>
    </main>

    <!-- 忘记密码弹窗 -->
    <div v-if="forgot.visible" class="modal-mask" @click.self="closeForgot">
      <div class="modal">
        <div class="modal-header">
          <div class="modal-title">重置密码</div>
          <button class="modal-close" @click="closeForgot" aria-label="关闭">×</button>
        </div>

        <form class="modal-form" @submit.prevent="submitForgot">
          <div class="form-group">
            <label for="fp_username">用户名</label>
            <input id="fp_username" v-model="forgot.username" type="text" placeholder="请输入用户名" autocomplete="username" required />
          </div>
          <div class="form-group">
            <label for="fp_new_password">新密码</label>
            <input id="fp_new_password" v-model="forgot.newPassword" type="password" placeholder="请输入新密码（至少 6 位）" autocomplete="new-password" required />
          </div>
          <div class="form-group">
            <label for="fp_confirm">确认密码</label>
            <input id="fp_confirm" v-model="forgot.confirm" type="password" placeholder="请再次输入新密码" autocomplete="new-password" required />
          </div>

          <div class="error-message" v-if="forgot.error">{{ forgot.error }}</div>
          <div class="success-message" v-if="forgot.success">{{ forgot.success }}</div>

          <div class="modal-actions">
            <button type="button" class="btn-cancel" :disabled="forgot.loading" @click="closeForgot">取消</button>
            <button type="submit" class="btn-primary" :disabled="forgot.loading">
              <span v-if="!forgot.loading">重置密码</span>
              <span v-else>提交中...</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { login, register, forgotPassword } from '@/api/auth'
import { getLDAPStatus } from '@/api/ldap'
import { useClusterStore } from '@/stores/cluster'

const router = useRouter()
const route = useRoute()

const mode = ref('login') // login | register

const form = ref({
  username: '',
  password: '',
  password_confirm: '',
  remember: false,
})

const isLoading = ref(false)
const error = ref('')
const success = ref('')
const showPassword = ref(false)
const ldapEnabled = ref(false)
const authMethod = ref('') // 'ldap' | 'local'

// 检查 LDAP 状态
onMounted(async () => {
  try {
    const res = await getLDAPStatus()
    const data = res?.data ?? res
    ldapEnabled.value = data?.enabled || false
  } catch {
    // LDAP 状态接口不可用时静默失败
  }
})

// =====================
// 忘记密码弹窗状态
// =====================
const forgot = ref({
  visible: false,
  username: '',
  newPassword: '',
  confirm: '',
  loading: false,
  error: '',
  success: '',
})

const openForgot = () => {
  forgot.value.visible = true
  forgot.value.loading = false
  forgot.value.error = ''
  forgot.value.success = ''
  // 默认带入当前登录框用户名（可选，体验更好）
  forgot.value.username = form.value.username || ''
  forgot.value.newPassword = ''
  forgot.value.confirm = ''
}

const closeForgot = () => {
  if (forgot.value.loading) return
  forgot.value.visible = false
  forgot.value.error = ''
  forgot.value.success = ''
}

const submitForgot = async () => {
  forgot.value.error = ''
  forgot.value.success = ''

  if (!forgot.value.username || !forgot.value.newPassword || !forgot.value.confirm) {
    forgot.value.error = '请填写用户名、新密码和确认密码'
    return
  }
  if (forgot.value.newPassword.length < 6) {
    forgot.value.error = '新密码至少 6 位'
    return
  }
  if (forgot.value.newPassword !== forgot.value.confirm) {
    forgot.value.error = '两次密码不一致'
    return
  }

  forgot.value.loading = true
  try {
    await forgotPassword({
      username: forgot.value.username,
      new_password: forgot.value.newPassword, // ⚠️ 必须 snake_case（和后端 DTO 对齐）
      confirm: forgot.value.confirm,
    })

    forgot.value.success = '密码重置成功，请使用新密码登录'
    // 同步到登录框，方便直接登录
    form.value.username = forgot.value.username
    form.value.password = ''
    setTimeout(() => closeForgot(), 600)
  } catch (e) {
    // 你的 http.js 里一般已经 Message.error 了，这里兜底显示
    forgot.value.error = e?.msg || e?.message || '密码重置失败'
  } finally {
    forgot.value.loading = false
  }
}

// 切换登录/注册模式
const toggleMode = () => {
  error.value = ''
  success.value = ''
  mode.value = mode.value === 'login' ? 'register' : 'login'
  form.value.password = ''
  form.value.password_confirm = ''
}

// 表单验证
const validateForm = () => {
  if (!form.value.username || !form.value.password) {
    error.value = '请输入用户名和密码'
    return false
  }

  if (mode.value === 'register') {
    if (form.value.password.length < 6) {
      error.value = '密码至少 6 位'
      return false
    }
    if (form.value.password !== form.value.password_confirm) {
      error.value = '两次密码不一致'
      return false
    }
  }

  return true
}

// 存储认证信息
const storeAuth = (token, user, remember) => {
  const storage = remember ? localStorage : sessionStorage
  const other = remember ? sessionStorage : localStorage

  storage.setItem('token', token)
  storage.setItem('user', JSON.stringify(user || {}))
  other.removeItem('token')
  other.removeItem('user')

  // 登录时清除旧的集群选择缓存，避免携带失效的 X-Cluster-ID
  const clusterStore = useClusterStore()
  clusterStore.setCurrent(null)
}

// ✅ 兼容两种后端返回：{code,msg,data} 或直接 {user,token}
const loginRequest = async (username, password) => {
  const res = await login({ username, password })

  if (res?.code && res.code !== 0) {
    error.value = res?.msg || '登录失败'
    return null
  }

  const data = res?.data ?? res
  const token = data?.token
  const user = data?.user

  if (!token) {
    error.value = res?.msg || '登录失败'
    return null
  }

  // 记录认证方式
  authMethod.value = data?.auth_method || 'local'

  return { token, user }
}

// 处理表单提交
const handleSubmit = async () => {
  error.value = ''
  success.value = ''
  isLoading.value = true

  try {
    if (!validateForm()) return

    // 注册流程
    if (mode.value === 'register') {
      const r = await register({
        username: form.value.username,
        password: form.value.password,
        password_confirm: form.value.password_confirm,
      })

      // 兼容 {code,msg,data} 或直接成功
      if (r?.code && r.code !== 0) {
        const d0 = Array.isArray(r?.details) ? r.details[0] : ''
        error.value = r?.msg || d0 || '注册失败'
        return
      }

      success.value = '注册成功，正在登录...'

      const loginData = await loginRequest(form.value.username, form.value.password)
      if (loginData) {
        storeAuth(loginData.token, loginData.user, true)
        // 注册后跳转到原页面或集群列表页
        const redirect = route.query.redirect || '/clusters'
        router.replace(redirect)
      }
      return
    }

    // 登录流程
    const loginData = await loginRequest(form.value.username, form.value.password)
    if (loginData) {
      storeAuth(loginData.token, loginData.user, form.value.remember)
      // 跳转到原页面或集群列表页（支持从钉钉审批链接跳转后 redirect 回去）
      const redirect = route.query.redirect || '/clusters'
      // 使用 replace 而非 push，避免登录页残留在历史记录中
      router.replace(redirect)
    }
  } catch (e) {
    error.value = e?.response?.data?.msg || e?.response?.data?.message || e?.message || '请求失败'
  } finally {
    isLoading.value = false
  }
}
</script>

<style scoped>
/* ============================================================
   插画风格登录页：浅蓝天空 + 港口云原生插画 + 居中白卡片
   ============================================================ */
.login-page {
  position: relative;
  min-height: 100vh;
  min-height: 100dvh;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  background: linear-gradient(180deg, #a9d1f0 0%, #cfe8fa 45%, #eef7fd 100%);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
}

/* ==================== 插画场景 ==================== */
.scene {
  position: absolute;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

.harbor {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 100%;
  height: 46%;
  min-height: 300px;
}

/* 云朵 */
.cloud {
  position: absolute;
  background: #ffffff;
  border-radius: 50%;
  filter: blur(2px);
  opacity: 0.9;
  box-shadow: 40px 10px 0 -6px #ffffff, 82px 18px 0 -12px #ffffff, -38px 14px 0 -8px #ffffff;
}
.cloud-1 { top: 12%; left: 14%; width: 70px; height: 70px; animation: cloudDrift 26s ease-in-out infinite; }
.cloud-2 { top: 22%; right: 18%; width: 54px; height: 54px; animation: cloudDrift 32s ease-in-out infinite reverse; }
.cloud-3 { top: 8%; left: 52%; width: 46px; height: 46px; animation: cloudDrift 38s ease-in-out infinite; }

@keyframes cloudDrift {
  0%, 100% { transform: translateX(0); }
  50% { transform: translateX(30px); }
}

/* ==================== 居中登录区 ==================== */
.login-main {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 400px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* 品牌 */
.brand {
  text-align: center;
  margin-bottom: 22px;
}

.brand-logo {
  width: 74px;
  height: 74px;
  margin: 0 auto 14px;
  border-radius: 50%;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 10px 26px rgba(47, 111, 208, 0.22);
}

.brand-title {
  margin: 0;
  font-size: 26px;
  font-weight: 800;
  letter-spacing: 2px;
  color: #2b5c9e;
}

.brand-subtitle {
  margin: 8px 0 0;
  font-size: 14px;
  color: #5a7fa8;
}

.ldap-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 4px 12px;
  background: rgba(47, 111, 208, 0.1);
  border: 1px solid rgba(47, 111, 208, 0.25);
  border-radius: 20px;
  color: #2f6fd0;
  font-size: 12px;
}

/* ==================== 白色卡片 ==================== */
.login-card {
  width: 100%;
  padding: 30px 28px 22px;
  background: #ffffff;
  border-radius: 16px;
  box-shadow: 0 24px 60px rgba(43, 92, 158, 0.18);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 输入框 */
.field {
  position: relative;
  display: flex;
  align-items: center;
}

.field-icon {
  position: absolute;
  left: 14px;
  display: flex;
  align-items: center;
  color: #9bb0c9;
  pointer-events: none;
}

.field input {
  width: 100%;
  height: 46px;
  padding: 0 42px;
  border-radius: 8px;
  border: 1px solid #e3e9f0;
  background: #f7f9fc;
  color: #2b3648;
  font-size: 14px;
  box-sizing: border-box;
  transition: border-color 0.2s, box-shadow 0.2s, background 0.2s;
}

.field input::placeholder {
  color: #a7b3c4;
}

.field input:hover {
  border-color: #b9d0ea;
}

.field input:focus {
  outline: none;
  border-color: #2f6fd0;
  background: #fff;
  box-shadow: 0 0 0 3px rgba(47, 111, 208, 0.14);
}

.eye {
  position: absolute;
  right: 12px;
  background: transparent;
  border: none;
  color: #9bb0c9;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  transition: color 0.2s;
}

.eye:hover {
  color: #2f6fd0;
}

/* 记住我 / 忘记密码 */
.form-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: -2px;
}

.remember {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #6b7a90;
  font-size: 13px;
  cursor: pointer;
}

.remember input {
  width: 15px;
  height: 15px;
  accent-color: #2f6fd0;
  cursor: pointer;
}

.forgot {
  color: #2f6fd0;
  text-decoration: none;
  font-size: 13px;
  transition: color 0.2s;
}

.forgot:hover {
  color: #1e56ad;
  text-decoration: underline;
}

/* 登录按钮 */
.submit-btn {
  height: 46px;
  border-radius: 8px;
  border: none;
  background: linear-gradient(135deg, #3f83e0, #2f6fd0);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 4px;
  cursor: pointer;
  transition: all 0.25s ease;
  box-shadow: 0 8px 20px rgba(47, 111, 208, 0.3);
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 12px 26px rgba(47, 111, 208, 0.4);
}

.submit-btn:active:not(:disabled) {
  transform: translateY(0);
}

.submit-btn:disabled {
  background: #9db8dd;
  cursor: not-allowed;
  box-shadow: none;
}

/* 提示 */
.error-message {
  background: rgba(245, 108, 108, 0.1);
  color: #e15b5b;
  padding: 10px 14px;
  border-radius: 8px;
  text-align: center;
  font-size: 13px;
  border: 1px solid rgba(245, 108, 108, 0.2);
}

.success-message {
  background: rgba(82, 196, 26, 0.1);
  color: #52a41a;
  padding: 10px 14px;
  border-radius: 8px;
  text-align: center;
  font-size: 13px;
  border: 1px solid rgba(82, 196, 26, 0.2);
}

.ldap-info {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 9px 14px;
  background: rgba(47, 111, 208, 0.06);
  border: 1px solid rgba(47, 111, 208, 0.16);
  border-radius: 8px;
  color: #2f6fd0;
  font-size: 13px;
}

/* 切换登录/注册 */
.switch-mode {
  text-align: center;
  font-size: 13px;
  color: #8a97a8;
}

.switch-mode a {
  color: #2f6fd0;
  text-decoration: none;
  font-weight: 600;
  margin-left: 4px;
}

.switch-mode a:hover {
  text-decoration: underline;
}

/* 版权 */
.card-footer {
  margin-top: 4px;
  text-align: center;
  font-size: 12px;
  color: #a7b3c4;
}

/* ==================== 忘记密码 Modal ==================== */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(43, 92, 158, 0.28);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 18px;
  z-index: 999;
}

.modal {
  width: 100%;
  max-width: 420px;
  padding: 24px;
  background: #ffffff;
  border: 1px solid #eef2f7;
  border-radius: 14px;
  box-shadow: 0 20px 60px rgba(43, 92, 158, 0.25);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.modal-title {
  color: #2b3648;
  font-size: 18px;
  font-weight: 600;
}

.modal-close {
  background: transparent;
  border: none;
  color: #909bab;
  font-size: 24px;
  cursor: pointer;
  line-height: 1;
  transition: color 0.2s;
}

.modal-close:hover {
  color: #2b3648;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group label {
  color: #6b7a90;
  font-size: 13px;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  height: 44px;
  padding: 0 14px;
  border-radius: 8px;
  border: 1px solid #e3e9f0;
  background: #f7f9fc;
  color: #2b3648;
  font-size: 14px;
  box-sizing: border-box;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #2f6fd0;
  background: #fff;
  box-shadow: 0 0 0 3px rgba(47, 111, 208, 0.14);
}

.modal-actions {
  display: flex;
  gap: 12px;
  margin-top: 6px;
}

.modal-actions button {
  flex: 1;
  height: 44px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-cancel {
  background: #f2f5f9;
  border: 1px solid #e3e9f0;
  color: #6b7a90;
}

.btn-cancel:hover:not(:disabled) {
  background: #e9eef5;
}

.btn-primary {
  background: linear-gradient(135deg, #3f83e0, #2f6fd0);
  border: none;
  color: #fff;
  box-shadow: 0 6px 16px rgba(47, 111, 208, 0.28);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
}

.modal-actions button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* ==================== 无障碍：尊重系统“减少动效”设置 ==================== */
@media (prefers-reduced-motion: reduce) {
  .cloud {
    animation: none !important;
  }
}

/* ==================== 响应式 ==================== */
@media (max-width: 480px) {
  .login-card {
    padding: 24px 20px 18px;
  }

  .brand-title {
    font-size: 22px;
  }

  .harbor {
    height: 34%;
    min-height: 220px;
  }
}
</style>
