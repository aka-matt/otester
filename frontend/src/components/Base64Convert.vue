<template>
  <n-modal
    :show="show"
    preset="card"
    title="Base64 Convert"
    class="base64-modal"
    :style="{ width: '80vw', maxWidth: '1100px' }"
    :bordered="false"
    size="huge"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <div class="base64-columns">
      <div class="base64-col">
        <div class="col-label">Input</div>
        <n-input
          v-model:value="input"
          type="textarea"
          class="base64-textarea"
          placeholder="Type or paste text here…"
          :autosize="false"
        />
      </div>
      <div class="base64-col">
        <div class="col-label">Base64</div>
        <n-input
          :value="encoded"
          type="textarea"
          class="base64-textarea"
          readonly
          placeholder="Base64 output appears here…"
          :autosize="false"
        />
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { NModal, NInput } from 'naive-ui'

defineProps<{ show: boolean }>()
const emit = defineEmits<{ (e: 'update:show', value: boolean): void }>()

const input = ref('')

// UTF-8 safe base64 encoding.
const encoded = computed(() => {
  if (!input.value) return ''
  try {
    const bytes = new TextEncoder().encode(input.value)
    let binary = ''
    for (const b of bytes) binary += String.fromCharCode(b)
    return btoa(binary)
  } catch {
    return ''
  }
})
</script>

<style scoped>
.base64-columns {
  display: flex;
  gap: 16px;
  height: 60vh;
}

.base64-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.col-label {
  margin-bottom: 8px;
  font-weight: 600;
  color: var(--text-primary);
}

.base64-textarea {
  flex: 1;
}

.base64-textarea :deep(.n-input__textarea-el) {
  height: 100%;
  resize: none;
  font-family: 'Consolas', 'Menlo', monospace;
}
</style>
