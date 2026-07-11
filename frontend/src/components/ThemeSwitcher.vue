<template>
  <div class="theme-switcher">
    <select v-model="currentMode" class="theme-select" @change="onThemeChange">
      <option value="light">Light</option>
      <option value="dark">Dark</option>
      <option value="system">System</option>
    </select>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useThemeStore } from '../stores/theme'

const themeStore = useThemeStore()

const currentMode = computed({
  get: () => themeStore.mode,
  set: (v) => themeStore.setMode(v as 'light' | 'dark' | 'system')
})

function onThemeChange() {
  themeStore.setMode(currentMode.value)
}
</script>

<style scoped>
.theme-switcher {
  display: flex;
  align-items: center;
}

.theme-select {
  padding: 4px 8px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--input-bg);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
}
</style>
