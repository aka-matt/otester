<template>
  <n-config-provider :theme="themeStore.theme">
    <n-message-provider>
      <n-dialog-provider>
        <div class="app-container">
          <MainView />
        </div>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { NConfigProvider, NMessageProvider, NDialogProvider } from 'naive-ui'
import MainView from './views/MainView.vue'
import { useThemeStore } from './stores/theme'
import { useConfigStore } from './stores/config'
import { useLogsStore } from './stores/logs'
import { GetLogs } from '../wailsjs/go/app/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import type { LogEntry } from './types'

const themeStore = useThemeStore()
const configStore = useConfigStore()
const logsStore = useLogsStore()

onMounted(async () => {
  themeStore.init()
  EventsOn('log:entry', (e: LogEntry) => logsStore.append(e))
  try {
    const existing = await GetLogs()
    if (existing) logsStore.prime(existing as LogEntry[])
  } catch {
    /* backend not ready in dev/test; ignore */
  }
  await configStore.loadConfig()
})
</script>

<style>
@import './themes/acrylic.css';
</style>
