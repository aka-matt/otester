<template>
  <div class="status-bar">
    <span class="status-item">
      <span class="label">Config:</span>
      <span class="value">{{ configPath || 'Not loaded' }}</span>
    </span>
    <span class="status-item">
      <span class="label">Status:</span>
      <span class="value" :class="statusClass">{{ statusText }}</span>
    </span>
    <span class="status-item">
      <span class="label">Version:</span>
      <span class="value">{{ version }}</span>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useConfigStore } from '../stores/config'
import { useRequestStore } from '../stores/request'

const configStore = useConfigStore()
const requestStore = useRequestStore()

const configPath = computed(() => configStore.config?.configPath || '')
const version = computed(() => configStore.config?.app?.name || '1.0.0')

const statusText = computed(() => {
  if (requestStore.sending) return 'Sending...'
  if (configStore.error) return 'Error'
  return 'Ready'
})

const statusClass = computed(() => {
  if (requestStore.sending) return 'status-sending'
  if (configStore.error) return 'status-error'
  return 'status-ready'
})
</script>

<style scoped>
.status-bar {
  display: flex;
  gap: 24px;
  padding: 6px 16px;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border-color);
  font-size: 12px;
}

.status-item {
  display: flex;
  gap: 6px;
}

.label {
  color: var(--text-secondary);
}

.value {
  color: var(--text-primary);
}

.status-sending {
  color: var(--accent-color);
}

.status-error {
  color: var(--danger-color);
}

.status-ready {
  color: var(--success-color);
}
</style>
