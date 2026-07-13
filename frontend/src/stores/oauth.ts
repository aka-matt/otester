import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { TokenStatus } from '../types'
import { GetTokenStatus } from '../../wailsjs/go/app/App'

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

  // Refresh the cached status for a profile from the backend. Safe to call
  // whenever the UI needs an up-to-date view (on profile selection, on Auth
  // tab open, after a successful OAuth send, periodically when a token is
  // cached so it expires from view shortly before the real lifetime ends).
  async function refreshStatus(profileId: string): Promise<void> {
    if (!profileId) {
      return
    }
    try {
      const status = await GetTokenStatus(profileId, null as any)
      if (status) {
        updateStatus(profileId, status as any)
      }
    } catch (e) {
      // Best-effort refresh; the previous status (if any) stays visible.
      console.warn('[oauth] refreshStatus failed', e)
    }
  }

  return {
    tokenStatuses,
    updateStatus,
    clearStatus,
    clearAll,
    refreshStatus,
  }
})
