import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ConfigView } from '../types'
import { LoadConfig } from '../../wailsjs/go/app/App'

declare global {
  interface Window {
    go?: {
      app: {
        App: {
          LoadConfig: () => Promise<ConfigView>
        }
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

  function ensureSelectedVariable(previousSelection = selectedVariable.value) {
    const variables = config.value?.variables ?? []
    const selectedIndex = previousSelection
      ? variables.findIndex(v => v.id === previousSelection.id && v.environment === previousSelection.environment)
      : -1
    if (selectedIndex !== -1) {
      selectedVariableIndex.value = selectedIndex
      return
    }
    selectedVariableIndex.value = variables.length ? 0 : null
  }

  function selectVariable(index: number) {
    if (config.value?.variables[index]) selectedVariableIndex.value = index
  }

  function setConfig(nextConfig: ConfigView) {
    const previousSelection = selectedVariable.value
    config.value = nextConfig
    ensureSelectedVariable(previousSelection)
  }

  async function loadConfig() {
    loading.value = true
    error.value = null
    try {
      const nextConfig = await LoadConfig() ?? null
      if (nextConfig) setConfig(nextConfig as unknown as ConfigView)
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
    if (!variable) return text
    const escapedID = variable.id.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    return text.replace(new RegExp(`{{${escapedID}}}`, 'g'), () => variable.base_url)
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
    setConfig,
    substituteVariables,
  }
})
