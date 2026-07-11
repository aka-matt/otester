import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { lightTheme, darkTheme } from 'naive-ui'

type ThemeMode = 'light' | 'dark' | 'system'

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>('system')

  const theme = computed(() => {
    if (mode.value === 'system') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      return prefersDark ? darkTheme : lightTheme
    }
    return mode.value === 'dark' ? darkTheme : lightTheme
  })

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

  return { mode, theme, setMode, init }
})
