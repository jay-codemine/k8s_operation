import { ref, watch } from 'vue'
import { defineStore } from 'pinia'

export const useThemeStore = defineStore('theme', () => {
  const stored = localStorage.getItem('theme') || 'light'
  const mode = ref(stored)
  const isDark = ref(stored === 'dark')

  const apply = (m) => {
    document.documentElement.setAttribute('data-theme', m)
    isDark.value = m === 'dark'
  }

  const toggle = () => {
    mode.value = mode.value === 'dark' ? 'light' : 'dark'
    localStorage.setItem('theme', mode.value)
    apply(mode.value)
  }

  // init on load
  apply(mode.value)

  return { mode, isDark, toggle }
})
