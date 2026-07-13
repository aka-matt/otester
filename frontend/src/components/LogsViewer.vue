<template>
  <n-modal
    :show="show"
    preset="card"
    title="Logs"
    class="logs-modal"
    :style="{ width: '80vw', maxWidth: '1100px' }"
    :bordered="false"
    size="huge"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <div class="logs-body">
      <n-input
        :value="logsStore.text"
        type="textarea"
        class="logs-textarea"
        readonly
        placeholder="HTTP and auth activity appears here…"
        :autosize="false"
      />
      <n-space class="logs-actions">
        <n-button size="small" @click="onClear">Clear</n-button>
        <n-button size="small" @click="onCopy">Copy</n-button>
      </n-space>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { NModal, NInput, NButton, NSpace, useMessage } from 'naive-ui'
import { useLogsStore } from '../stores/logs'
import { ClearLogs } from '../../wailsjs/go/app/App'

defineProps<{ show: boolean }>()
const emit = defineEmits<{ (e: 'update:show', value: boolean): void }>()

const logsStore = useLogsStore()
const message = useMessage()

async function onClear() {
  try {
    await ClearLogs()
  } catch {
    /* backend may be unavailable in dev; still clear the view */
  }
  logsStore.clear()
}

async function onCopy() {
  try {
    await navigator.clipboard.writeText(logsStore.text)
    message.success('Logs copied')
  } catch {
    message.error('Copy failed')
  }
}
</script>

<style scoped>
.logs-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 60vh;
}

.logs-textarea {
  flex: 1;
}

.logs-textarea :deep(.n-input__textarea-el) {
  height: 100%;
  resize: none;
  font-family: 'Consolas', 'Menlo', monospace;
  font-size: 12px;
}

.logs-actions {
  justify-content: flex-end;
}
</style>