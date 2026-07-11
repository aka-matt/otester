import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ResponseOutput } from '../types'

export const useResponseStore = defineStore('response', () => {
  const response = ref<ResponseOutput | null>(null)
  const loading = ref(false)

  function setResponse(resp: ResponseOutput) {
    response.value = resp
    loading.value = false
  }

  function clear() {
    response.value = null
    loading.value = false
  }

  return {
    response,
    loading,
    setResponse,
    clear,
  }
})
