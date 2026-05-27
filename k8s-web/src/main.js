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

const app = createApp(App)

// 2. 注册 Arco 组件（核心）
app.use(ArcoVue)

// Message 你可以留着（可选）
app.config.globalProperties.$message = Message

// 3. 注册权限插件（v-permission 指令 + $hasPermission 方法）
setupPermission(app)

app.use(router)
app.use(pinia)

app.mount('#app')
