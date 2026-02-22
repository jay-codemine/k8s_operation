import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import vueDevTools from 'vite-plugin-vue-devtools'
import {fileURLToPath, URL} from 'node:url'

export default defineConfig({
  plugins: [vue(), vueJsx(), vueDevTools()],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },

  server: {
    port: 5173,
    strictPort: true,
    host: 'localhost',

    // ✅ 关键：开发代理
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080', // 你的 Gin 后端
        changeOrigin: true,

        // 如果你后端路径本身就带 /api，可以不 rewrite
        // rewrite: (path) => path.replace(/^\/api/, ''),
      },
    },
  },
})

