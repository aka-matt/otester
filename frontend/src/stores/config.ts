import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ConfigView, EndpointView, Variable } from '../types'

declare global {
  interface Window {
    go?: {
      app: {
        LoadConfig: () => Promise<ConfigView>
        ReloadConfig: () => Promise<ConfigView>
        ValidateConfig: () => Promise<{ valid: boolean; errors: string[]; warnings: string[] }>
        GetAppInfo: () => Promise<{ version: string; name: string }>
      }
    }
  }
}

export const useConfigStore = defineStore('config', () => {
  const config = ref<ConfigView | null>(null)
  const selectedEndpointId = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const selectedEndpoint = computed(() => {
    if (!config.value || !selectedEndpointId.value) return null
    return config.value.endpoints.find(ep => ep.id === selectedEndpointId.value) || null
  })

  async function loadConfig() {
    loading.value = true
    error.value = null
    try {
      config.value = await window.go?.app.LoadConfig()
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }

  async function reloadConfig() {
    return loadConfig()
  }

  function selectEndpoint(id: string | null) {
    selectedEndpointId.value = id
  }

  function substituteVariables(text: string): string {
    if (!config.value) return text
    let result = text
    for (const v of config.value.variables) {
      result = result.replace(new RegExp(`{{${v.id}}}`, 'g'), v.base_url)
    }
    return result
  }

  return {
    config,
    selectedEndpointId,
    selectedEndpoint,
    loading,
    error,
    loadConfig,
    reloadConfig,
    selectEndpoint,
    substituteVariables,
  }
})
