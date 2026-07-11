import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { KeyValue } from '../types'

export const useRequestStore = defineStore('request', () => {
  const method = ref('GET')
  const url = ref('')
  const headers = ref<KeyValue[]>([])
  const queryParams = ref<KeyValue[]>([])
  const bodyType = ref<'none' | 'json' | 'text' | 'x-www-form-urlencoded' | 'raw'>('none')
  const body = ref('')
  const timeoutSeconds = ref(30)
  const useOAuth = ref(false)
  const oauthProfileId = ref('')
  const sending = ref(false)
  const requestId = ref<string | null>(null)

  function loadFromEndpoint(endpoint: any) {
    method.value = endpoint.method || 'GET'
    url.value = endpoint.url || ''
    headers.value = endpoint.headers || []
    queryParams.value = endpoint.queryParams || []
    bodyType.value = endpoint.body?.type || 'none'
    body.value = endpoint.body?.content || ''
    timeoutSeconds.value = endpoint.timeoutSeconds || 30
    useOAuth.value = endpoint.auth?.type === 'oauth2'
    oauthProfileId.value = endpoint.auth?.profileId || ''
  }

  function reset() {
    method.value = 'GET'
    url.value = ''
    headers.value = []
    queryParams.value = []
    bodyType.value = 'none'
    body.value = ''
    timeoutSeconds.value = 30
    useOAuth.value = false
    oauthProfileId.value = ''
    sending.value = false
    requestId.value = null
  }

  function generateRequestId(): string {
    return `req-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
  }

  return {
    method,
    url,
    headers,
    queryParams,
    bodyType,
    body,
    timeoutSeconds,
    useOAuth,
    oauthProfileId,
    sending,
    requestId,
    loadFromEndpoint,
    reset,
    generateRequestId,
  }
})
