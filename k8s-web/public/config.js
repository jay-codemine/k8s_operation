// 运行时配置 —— 由 docker-entrypoint.sh 在容器启动时动态生成
// 开发环境：Vite 代理 /api → 后端，直接用相对路径即可
// 生产环境：Nginx 反向代理 /api → 后端 Service，同样相对路径
window._CONFIG = {
  API_BASE: '/api/v1',
  WS_URL: window.location.protocol === 'https:' ? `wss://${window.location.host}` : `ws://${window.location.host}`,
  ENV: 'dev',
  APP_NAME: 'K8sOperation',
}
