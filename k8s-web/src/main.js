import {createApp} from 'vue'
import App from './App.vue'
import router from './router'

//  1. 引入 Arco Vue
import ArcoVue, {Message} from '@arco-design/web-vue'
import '@arco-design/web-vue/dist/arco.css'

// 引入全局主题变量 (v3.0 - 现代化配色)
import './styles/theme-variables.css'
import {pinia} from '@/stores'

// 引入权限插件
import { setupPermission } from '@/directives/permission'

// chart.js 注册保持不变
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from 'chart.js'

ChartJS.register(
  LineElement,
  PointElement,
  LinearScale,
  CategoryScale,
  Tooltip,
  Legend,
  Filler
)

import AppIcon from '@/components/AppIcon.vue'

const app = createApp(App)
app.component('AppIcon', AppIcon)

// 2. 注册 Arco 组件（核心）
app.use(ArcoVue)

// Message 你可以留着（可选）
app.config.globalProperties.$message = Message

// 3. 注册权限插件（v-permission 指令 + $hasPermission 方法）
setupPermission(app)

app.use(router)
app.use(pinia)

// ================================================================
// 全局 Vue 错误处理：防止组件渲染错误导致白屏
// ================================================================
app.config.errorHandler = (err, instance, info) => {
  console.error('[Vue Error]', err, 'component:', instance?.$options?.name || instance?.type?.name || 'unknown', 'info:', info)

  // chunk 加载失败：强制刷新（新版部署后旧 chunk 404）
  if (err?.message?.includes('Failed to fetch dynamically imported module')) {
    window.location.reload()
    return
  }

  // 其他渲染错误：记录但不崩溃（Vue 3 默认会替换掉出错的组件为空）
  // 保留错误提示，便于排查
}

// 等待路由初始化完成后再挂载（确保权限等异步数据加载完毕）
router.isReady().then(() => {
  app.mount('#app')
})
