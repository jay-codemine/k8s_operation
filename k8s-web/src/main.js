import {createApp} from 'vue'
import App from './App.vue'
import router from './router'

//  1. 引入 Arco Vue
import ArcoVue, {Message} from '@arco-design/web-vue'
import '@arco-design/web-vue/dist/arco.css'

// 引入 UI 增强样式 (v2.0 专业版 - 可移除此行回滚到 v1.0)
import './styles/enhancement.css'

import {pinia} from '@/stores'

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

app.use(router)
app.use(pinia)

app.mount('#app')
