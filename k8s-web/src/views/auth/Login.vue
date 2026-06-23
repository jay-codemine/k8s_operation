<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <div class="logo">
          <!-- Kubernetes Logo -->
          <svg xmlns="http://www.w3.org/2000/svg" width="52" height="52" viewBox="0 0 100 100">
            <g fill="#326ce5">
              <path
                d="M39.971 5.05c-3.607-.418-7.152.532-9.957 2.694l-1.315 1.035 1.703 2.17 1.227-1.006c1.946-1.606 4.387-2.537 6.984-2.537 3.97 0 7.513 2.053 9.616 5.298l1.127 1.629 2.012-1.388-1.217-1.772c-2.563-3.722-6.534-6.13-10.984-6.23z"
              />
              <path
                d="M42.026 94.796c3.608.418 7.153-.53 9.958-2.693l1.315-1.035-1.702-2.17-1.228 1.006c-1.947 1.606-4.388 2.537-6.985 2.537-3.969 0-7.512-2.053-9.615-5.298l-1.127-1.63-2.012 1.388 1.217 1.772c2.564 3.723 6.535 6.13 10.984 6.23z"
              />
            </g>
          </svg>
        </div>
        <h2>Kubernetes Admin Dashboard</h2>
        <p>{{ mode === 'login' ? '请登录以继续' : '创建新账号' }}</p>
        <div v-if="ldapEnabled && mode === 'login'" class="ldap-badge">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
          <span>LDAP 认证已启用</span>
        </div>
      </div>

      <form class="login-form" @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="username">用户名</label>
          <input
            id="username"
            v-model="form.username"
            type="text"
            placeholder="请输入用户名"
            autocomplete="username"
            required
            aria-label="用户名"
          />
        </div>

        <div class="form-group">
          <label for="password">密码</label>
          <div class="password-wrapper">
            <input
              id="password"
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="请输入密码"
              :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
              required
              aria-label="密码"
            />
            <button type="button" class="toggle-password" @click="showPassword = !showPassword" tabindex="-1">
              <svg v-if="!showPassword" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                <circle cx="12" cy="12" r="3"/>
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/>
                <line x1="1" y1="1" x2="23" y2="23"/>
              </svg>
            </button>
          </div>
        </div>

        <div class="form-group" v-if="mode === 'register'">
          <label for="password_confirm">确认密码</label>
          <input
            id="password_confirm"
            v-model="form.password_confirm"
            type="password"
            placeholder="请再次输入密码"
            autocomplete="new-password"
            required
            aria-label="确认密码"
          />
        </div>

        <div class="form-options" v-if="mode === 'login'">
          <div class="remember-me">
            <input type="checkbox" id="remember" v-model="form.remember" />
            <label for="remember">记住我</label>
          </div>

          <!-- ✅ 修改点：弹出重置密码弹窗 -->
          <a href="#" class="forgot-password" @click.prevent="openForgot">
            忘记密码?
          </a>
        </div>

        <div class="button-group">
          <button type="submit" class="login-btn" :disabled="isLoading">
            <span v-if="!isLoading">{{ mode === 'login' ? '登录' : '注册' }}</span>
            <span v-else>{{ mode === 'login' ? '登录中...' : '注册中...' }}</span>
          </button>

          <button type="button" class="register-btn" :disabled="isLoading" @click="toggleMode">
            {{ mode === 'login' ? '注册' : '返回登录' }}
          </button>
        </div>

        <div class="error-message" v-if="error">{{ error }}</div>
        <div class="success-message" v-if="success">{{ success }}</div>
        <div class="ldap-info" v-if="authMethod === 'ldap' && success">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
          已通过 LDAP 认证
        </div>
      </form>
    </div>

    <!-- =========================
         ✅ 忘记密码弹窗（遮罩 + 表单）
         ========================= -->
    <div v-if="forgot.visible" class="modal-mask" @click.self="closeForgot">
      <div class="modal">
        <div class="modal-header">
          <div class="modal-title">重置密码</div>
          <button class="modal-close" @click="closeForgot" aria-label="关闭">×</button>
        </div>

        <form class="modal-form" @submit.prevent="submitForgot">
          <div class="form-group">
            <label for="fp_username">用户名</label>
            <input
              id="fp_username"
              v-model="forgot.username"
              type="text"
              placeholder="请输入用户名"
              autocomplete="username"
              required
            />
          </div>

          <div class="form-group">
            <label for="fp_new_password">新密码</label>
            <input
              id="fp_new_password"
              v-model="forgot.newPassword"
              type="password"
              placeholder="请输入新密码（至少 6 位）"
              autocomplete="new-password"
              required
            />
          </div>

          <div class="form-group">
            <label for="fp_confirm">确认密码</label>
            <input
              id="fp_confirm"
              v-model="forgot.confirm"
              type="password"
              placeholder="请再次输入新密码"
              autocomplete="new-password"
              required
            />
          </div>

          <div class="error-message" v-if="forgot.error">{{ forgot.error }}</div>
          <div class="success-message" v-if="forgot.success">{{ forgot.success }}</div>

          <div class="modal-actions">
            <button type="button" class="register-btn" :disabled="forgot.loading" @click="closeForgot">
              取消
            </button>
            <button type="submit" class="login-btn" :disabled="forgot.loading">
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
  } catch (e) {
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
:root {
  --primary-color: #165DFF;
  --secondary-color: #4080FF;
}

.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  min-height: 100dvh;
  padding: 24px;
  overflow-y: auto;
  background: linear-gradient(135deg, #1d2129 0%, #2a3042 50%, #1d2129 100%);
}

.login-box {
  width: 100%;
  max-width: 420px;
  padding: 40px 36px;
  background: rgba(255, 255, 255, .06);
  backdrop-filter: blur(16px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, .1);
  box-shadow: 0 20px 60px rgba(0, 0, 0, .3);
}

.login-header {
  text-align: center;
  margin-bottom: 28px;
}

.login-header h2 {
  color: #f8fafc;
  font-size: 26px;
  margin-bottom: 8px;
}

.login-header p {
  color: #94a3b8;
  font-size: 14px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.form-group label {
  color: #e2e8f0;
  font-size: 13px;
}

.form-group input {
  width: 100%;
  padding: 12px 16px;
  border-radius: 10px;
  background: rgba(255, 255, 255, .08);
  border: 1px solid rgba(255, 255, 255, .18);
  color: #f8fafc;
  box-sizing: border-box;
}

.password-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-wrapper input {
  padding-right: 44px;
}

.toggle-password {
  position: absolute;
  right: 12px;
  background: transparent;
  border: none;
  color: #94a3b8;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.toggle-password:hover {
  color: #e2e8f0;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
}

.remember-me {
  display: flex;
  align-items: center;
  gap: 6px;
}

.remember-me label {
  color: #94a3b8;
  font-size: 14px;
}

.forgot-password {
  color: #4080FF;
  text-decoration: none;
  font-size: 14px;
  transition: color 0.3s ease;
}

.forgot-password:hover {
  color: #6aa1ff;
  text-decoration: underline;
}

.button-group {
  display: flex;
  gap: 12px;
  margin-top: 20px;
}

.login-btn {
  padding: 14px;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--primary-color), var(--secondary-color));
  color: #fff;
  font-weight: 600;
  border: none;
  cursor: pointer;
  transition: all 0.3s ease;
  flex: 1;
  font-size: 15px;
  box-shadow: 0 4px 14px rgba(22, 93, 255, 0.3);
}

.login-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, var(--secondary-color), #165DFF);
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(22, 93, 255, 0.4);
}

.login-btn:disabled {
  background: linear-gradient(135deg, #4080FF, #6aa1ff);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.register-btn {
  padding: 14px;
  border-radius: 10px;
  background: rgba(255, 255, 255, .06);
  color: #c9cdd4;
  border: 1px solid rgba(255, 255, 255, .2);
  cursor: pointer;
  transition: all 0.3s ease;
  flex: 1;
  font-size: 15px;
}

.register-btn:hover:not(:disabled) {
  background: rgba(22, 93, 255, 0.1);
  color: #4080FF;
  border-color: rgba(64, 128, 255, .4);
  transform: translateY(-1px);
  box-shadow: 0 8px 20px rgba(22, 93, 255, 0.15);
}

.register-btn:disabled {
  background: rgba(255, 255, 255, .04);
  color: #86909c;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.error-message {
  background: rgba(239, 68, 68, .15);
  color: #fecaca;
  padding: 12px;
  border-radius: 10px;
  text-align: center;
}

.success-message {
  background: rgba(34, 197, 94, .15);
  color: #bbf7d0;
  padding: 12px;
  border-radius: 10px;
  text-align: center;
}

/* =====================
   Modal 样式
   ===================== */
.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(2, 6, 23, 0.65);
  backdrop-filter: blur(6px);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 18px;
  z-index: 999;
}

.modal {
  width: 100%;
  max-width: 440px;
  padding: 18px 18px 16px;
  background: rgba(29, 33, 41, .95);
  border: 1px solid rgba(255, 255, 255, .1);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, .4);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.modal-title {
  color: #f8fafc;
  font-size: 16px;
  font-weight: 600;
}

.modal-close {
  background: transparent;
  border: none;
  color: #cbd5e1;
  font-size: 22px;
  cursor: pointer;
  line-height: 1;
}

.modal-close:hover {
  color: #ffffff;
}

.modal-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  margin-top: 6px;
}

@media (max-width: 480px) {
  .login-container {
    align-items: flex-start;
    padding: 16px;
  }

  .login-box {
    padding: 28px 20px;
    border-radius: 16px;
  }

  .login-header h2 {
    font-size: 22px;
  }

  .button-group,
  .modal-actions {
    flex-direction: column;
  }
}

/* LDAP 标识样式 */
.ldap-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
  padding: 6px 14px;
  background: rgba(34, 197, 94, 0.12);
  border: 1px solid rgba(34, 197, 94, 0.3);
  border-radius: 20px;
  color: #86efac;
  font-size: 12px;
}

.ldap-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 14px;
  background: rgba(22, 93, 255, 0.1);
  border: 1px solid rgba(22, 93, 255, 0.25);
  border-radius: 10px;
  color: #4080FF;
  font-size: 13px;
}
</style>
