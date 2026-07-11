import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ConfigView, RequestInput, ResponseOutput, TokenStatus } from '../types'

declare global {
  interface Window {
    go?: {
      app: {
        LoadConfig: () => Promise<ConfigView>
        ReloadConfig: () => Promise<ConfigView>
        ValidateConfig: () => Promise<{ valid: boolean; errors: string[]; warnings: string[] }>
        GetAppInfo: () => Promise<{ version: string; name: string }>
        SendRequest: (input: RequestInput) => Promise<ResponseOutput>
        CancelRequest: (requestId: string) => Promise<void>
        GetTokenStatus: (profileId: string, profile: any) => Promise<TokenStatus>
        ClearTokenCache: () => Promise<void>
        OpenConfigDirectory: () => Promise<void>
        OpenLogDirectory: () => Promise<void>
      }
    }
  }
}

export const useConfigStore = defineStore('config', () => {
  const config = ref<ConfigView | null>(null)
  const selectedEndpointId = ref<string | null>(null)
  const selectedVariableIndex = ref<number | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const selectedEndpoint = computed(() => {
    if (!config.value || !selectedEndpointId.value) return null
    return config.value.endpoints.find(ep => ep.id === selectedEndpointId.value) || null
  })

  const selectedVariable = computed(() => selectedVariableIndex.value === null
    ? null : config.value?.variables[selectedVariableIndex.value] ?? null)

  function ensureSelectedVariable() {
    const variables = config.value?.variables ?? []
    const current = selectedVariable.value
    if (current && variables.some(v => v.id === current.id && v.environment === current.environment)) return
    selectedVariableIndex.value = variables.length ? 0 : null
  }

  function selectVariable(index: number) {
    if (config.value?.variables[index]) selectedVariableIndex.value = index
  }

  async function loadConfig() {
    loading.value = true
    error.value = null
    try {
      config.value = await window.go?.app.LoadConfig() ?? null
      ensureSelectedVariable()
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
    const variable = selectedVariable.value
    return variable ? text.replace(new RegExp(`{{${variable.id}}}`, 'g'), variable.base_url) : text
  }

  return {
    config,
    selectedEndpointId,
    selectedEndpoint,
    selectedVariableIndex,
    selectedVariable,
    loading,
    error,
    loadConfig,
    reloadConfig,
    selectEndpoint,
    ensureSelectedVariable,
    selectVariable,
    substituteVariables,
  }
})
