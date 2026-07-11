<template>
  <div class="key-value-editor">
    <div class="kv-header">
      <span class="col-enabled"></span>
      <span class="col-key">Key</span>
      <span class="col-value">Value</span>
      <span class="col-actions"></span>
    </div>

    <div class="kv-rows">
      <div v-for="(item, index) in modelValue" :key="index" class="kv-row">
        <input
          type="checkbox"
          v-model="item.enabled"
          class="col-enabled"
        />
        <input
          v-model="item.key"
          type="text"
          placeholder="Key"
          class="col-key acrylic-input"
        />
        <input
          v-model="item.value"
          type="text"
          placeholder="Value"
          class="col-value acrylic-input"
        />
        <button class="col-actions delete-btn" @click="removeRow(index)">×</button>
      </div>
    </div>

    <button class="acrylic-btn-secondary add-btn" @click="addRow">
      + Add
    </button>
  </div>
</template>

<script setup lang="ts">
import type { KeyValue } from '../types'

const props = defineProps<{
  modelValue: KeyValue[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: KeyValue[]): void
}>()

function addRow() {
  emit('update:modelValue', [
    ...props.modelValue,
    { key: '', value: '', enabled: true }
  ])
}

function removeRow(index: number) {
  const newValue = [...props.modelValue]
  newValue.splice(index, 1)
  emit('update:modelValue', newValue)
}
</script>

<style scoped>
.key-value-editor {
  padding: 12px 0;
}

.kv-header {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}

.kv-rows {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.kv-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.col-enabled {
  width: 24px;
}

.col-key {
  flex: 1;
}

.col-value {
  flex: 2;
}

.col-actions {
  width: 28px;
}

.delete-btn {
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.delete-btn:hover {
  background: var(--bg-primary);
  color: var(--danger-color);
}

.add-btn {
  margin-top: 8px;
}
</style>
