<!-- src/App.vue -->
<template>
  <RouterView />
</template>

<script setup>
import { RouterView } from 'vue-router'
</script>

<style>
/* ===== 全局重置 ===== */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* ===== 响应式根字体大小 ===== */
html {
  /* 
   * 响应式字体：基于视口宽度自动缩放
   * - 最小 13px（小屏幕/高分屏 100% 缩放）
   * - 最大 16px（大屏幕）
   * - 中间值：0.8vw（视口宽度的 0.8%）
   * 
   * 1920px 屏幕：0.8% = 15.36px
   * 1440px 屏幕：0.8% = 11.52px → 取 min 13px
   * 2560px 屏幕：0.8% = 20.48px → 取 max 16px
   */
  font-size: clamp(13px, 0.8vw, 16px);
  height: 100%;
  width: 100%;
  overflow: hidden;
}

body {
  font-size: 1rem; /* 继承 html 的响应式字体 */
  height: 100%;
  width: 100%;
  overflow: hidden;
  line-height: 1.5;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  height: 100%;
  width: 100%;
  overflow: auto;
}

/* ===== 响应式工具类 ===== */
/* 可在任何组件中使用这些类 */
.text-xs { font-size: 0.75rem; }   /* 12px @ 16px base */
.text-sm { font-size: 0.875rem; }  /* 14px @ 16px base */
.text-base { font-size: 1rem; }    /* 16px @ 16px base */
.text-lg { font-size: 1.125rem; }  /* 18px @ 16px base */
.text-xl { font-size: 1.25rem; }   /* 20px @ 16px base */

/* ===== 针对不同屏幕尺寸的细调 ===== */
/* 超大屏幕（4K / 2560px+） */
@media (min-width: 2560px) {
  html {
    font-size: 16px;
  }
}

/* 大屏幕（1920px - 2559px）*/
@media (min-width: 1920px) and (max-width: 2559px) {
  html {
    font-size: 15px;
  }
}

/* 中等屏幕（1440px - 1919px）*/
@media (min-width: 1440px) and (max-width: 1919px) {
  html {
    font-size: 14px;
  }
}

/* 小屏幕（1280px - 1439px）*/
@media (min-width: 1280px) and (max-width: 1439px) {
  html {
    font-size: 13px;
  }
}

/* 更小屏幕（1024px - 1279px）*/
@media (min-width: 1024px) and (max-width: 1279px) {
  html {
    font-size: 12px;
  }
}

/* 平板及以下（< 1024px）*/
@media (max-width: 1023px) {
  html {
    font-size: 14px; /* 移动端保持可读性 */
  }
}
</style>
