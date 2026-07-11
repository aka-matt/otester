import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

type ThemeMode = 'light' | 'dark' | 'system'

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>('system')

  function setMode(newMode: ThemeMode) {
    mode.value = newMode
    localStorage.setItem('otester-theme', newMode)
    applyTheme()
  }

  function applyTheme() {
    const root = document.documentElement
    if (mode.value === 'system') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      root.setAttribute('data-theme', prefersDark ? 'dark' : 'light')
    } else {
      root.setAttribute('data-theme', mode.value)
    }
  }

  function init() {
    const saved = localStorage.getItem('otester-theme') as ThemeMode | null
    if (saved) {
      mode.value = saved
    }
    applyTheme()

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (mode.value === 'system') {
        applyTheme()
      }
    })
  }

  return { mode, setMode, init }
})
