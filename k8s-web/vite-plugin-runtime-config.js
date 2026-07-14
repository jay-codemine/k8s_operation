// vite-plugin-runtime-config.js
// 自动将运行时配置注入 index.html，无需手动修改 index.html 或 http.js
//
// 新项目接入只需：
//   1. 将本文件复制到项目根目录
//   2. 在 vite.config.js 中 import 并 plugins 加上一行
//   3. 创建 public/config.js（可选，有默认值）
//
// 原理：
//   - 开发时：Vite 自动注入 config.js 到 HTML
//   - 构建时：生成独立的 config.js 文件，容器启动时 docker-entrypoint.sh 可覆盖

import fs from 'node:fs'
import path from 'node:path'

const PLUGIN_NAME = 'vite-plugin-runtime-config'

export function runtimeConfig(options = {}) {
  const {
    configFile = 'public/config.js',       // 运行时配置文件路径
    injectTo = 'head-prepend',             // 注入位置：head-prepend | head | body-prepend | body
  } = options

  let configJsContent = ''

  return {
    name: PLUGIN_NAME,

    // buildStart 时读取 config.js 内容，后续注入
    buildStart() {
      const fullPath = path.resolve(configFile)
      if (fs.existsSync(fullPath)) {
        configJsContent = fs.readFileSync(fullPath, 'utf-8')
        console.log(`[${PLUGIN_NAME}] 已读取运行时配置: ${configFile}`)
      } else {
        // 默认生成
        configJsContent = `window._CONFIG = {
  API_BASE: '/api/v1',
  ENV: 'dev',
  APP_NAME: 'K8sOperation',
}`
        console.log(`[${PLUGIN_NAME}] 使用默认运行时配置（${configFile} 不存在）`)
      }
    },

    // 自动注入到 HTML
    transformIndexHtml: {
      order: 'pre',
      handler(html) {
        const scriptTag = `<script>${configJsContent}</script>`
        switch (injectTo) {
          case 'head-prepend':
            return html.replace('<head>', `<head>\n    ${scriptTag}`)
          case 'body-prepend':
            return html.replace('<body>', `<body>\n    ${scriptTag}`)
          case 'body':
            return html.replace('</body>', `  ${scriptTag}\n  </body>`)
          case 'head':
          default:
            return html.replace('</head>', `  ${scriptTag}\n  </head>`)
        }
      },
    },

    // build 完成后，复制 config.js 到 dist/（供 docker-entrypoint.sh 覆盖）
    closeBundle() {
      const fullPath = path.resolve(configFile)
      if (fs.existsSync(fullPath)) {
        // config.js 在 public/ 下，Vite 会自动复制到 dist/
        // 这里不需要额外处理，只是打印一条提示
        const outDir = path.resolve('dist', path.basename(configFile))
        console.log(`[${PLUGIN_NAME}] 运行时配置文件已输出到 ${outDir}`)
        console.log(`[${PLUGIN_NAME}] 容器部署时由 docker-entrypoint.sh 动态覆盖此文件`)
      }
    },
  }
}
