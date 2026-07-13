import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type { LogEntry } from '../types'

const MAX_ENTRIES = 2000

export const useLogsStore = defineStore('logs', () => {
  const entries = ref<LogEntry[]>([])

  function cap() {
    if (entries.value.length > MAX_ENTRIES) {
      entries.value.splice(0, entries.value.length - MAX_ENTRIES)
    }
  }

  function append(e: LogEntry) {
    entries.value.push(e)
    cap()
  }

  function prime(list: LogEntry[]) {
    entries.value = list.slice(-MAX_ENTRIES)
  }

  function clear() {
    entries.value = []
  }

  const text = computed(() =>
    entries.value.map((e) => `${e.time} [${e.level}] ${e.message}`).join('\n'),
  )

  return { entries, text, append, prime, clear }
})
