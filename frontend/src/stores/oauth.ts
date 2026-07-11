import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { TokenStatus } from '../types'

export const useOAuthStore = defineStore('oauth', () => {
  const tokenStatuses = ref<Record<string, TokenStatus>>({})

  function updateStatus(profileId: string, status: TokenStatus) {
    tokenStatuses.value[profileId] = status
  }

  function clearStatus(profileId: string) {
    delete tokenStatuses.value[profileId]
  }

  function clearAll() {
    tokenStatuses.value = {}
  }

  return {
    tokenStatuses,
    updateStatus,
    clearStatus,
    clearAll,
  }
})
