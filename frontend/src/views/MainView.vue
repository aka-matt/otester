<template>
  <div class="main-view">
    <header class="top-bar">
      <div class="menu-section">
        <n-dropdown trigger="click" :options="fileMenuOptions" @select="handleMenuSelect">
          <span class="menu-item">File</span>
        </n-dropdown>
        <n-dropdown trigger="click" :options="viewMenuOptions" @select="handleMenuSelect">
          <span class="menu-item">View</span>
        </n-dropdown>
        <n-dropdown trigger="click" :options="toolsMenuOptions" @select="handleMenuSelect">
          <span class="menu-item">Tools</span>
        </n-dropdown>
        <n-dropdown trigger="click" :options="helpMenuOptions" @select="handleMenuSelect">
          <span class="menu-item">Help</span>
        </n-dropdown>
      </div>
      <ThemeSwitcher />
    </header>

    <div class="content-area">
      <aside class="left-panel acrylic-card">
        <EndpointList />
      </aside>

      <main class="center-panel">
        <RequestEditor />
      </main>

      <aside class="right-panel acrylic-card">
        <ResponseViewer />
      </aside>
    </div>

    <StatusBar />
  </div>
</template>

<script setup lang="ts">
import { NDropdown, useMessage } from 'naive-ui'
import EndpointList from '../components/EndpointList.vue'
import RequestEditor from '../components/RequestEditor.vue'
import ResponseViewer from '../components/ResponseViewer.vue'
import ThemeSwitcher from '../components/ThemeSwitcher.vue'
import StatusBar from '../components/StatusBar.vue'
import { useConfigStore } from '../stores/config'
import { useResponseStore } from '../stores/response'
import {
  OpenConfigFile,
  OpenLogDirectory,
  ValidateConfig,
  ClearTokenCache,
  GetAppInfo,
} from '../../wailsjs/go/app/App'
import type { ConfigView } from '../types'

const configStore = useConfigStore()
const responseStore = useResponseStore()
const message = useMessage()

const fileMenuOptions = [
  { label: 'Reload Config', key: 'reload-config' },
  { type: 'divider', key: 'd1' },
  { label: 'Open Config File', key: 'open-config-file' },
  { label: 'Open Log Directory', key: 'open-log-dir' },
  { type: 'divider', key: 'd2' },
  { label: 'Exit', key: 'exit' },
]

const viewMenuOptions = [
  { label: 'Toggle Left Panel', key: 'toggle-left' },
  { label: 'Toggle Right Panel', key: 'toggle-right' },
]

const toolsMenuOptions = [
  { label: 'Validate Config', key: 'validate-config' },
  { label: 'Clear Token Cache', key: 'clear-tokens' },
]

const helpMenuOptions = [
  { label: 'About', key: 'about' },
]

async function handleMenuSelect(key: string) {
  switch (key) {
    case 'reload-config':
      await configStore.loadConfig()
      message.success('Config reloaded')
      break
    case 'open-config-file':
      try {
        const config = await OpenConfigFile()
        if (config) {
          configStore.setConfig(config as unknown as ConfigView)
          responseStore.clear()
        }
      } catch (error) {
        message.error('Failed to open config file: ' + String(error))
      }
      break
    case 'open-log-dir':
      await OpenLogDirectory()
      break
    case 'validate-config':
      const result = await ValidateConfig()
      if (result?.valid) {
        message.success('Config is valid')
      } else {
        message.error('Config error: ' + (result?.errors?.join(', ') || 'Unknown'))
      }
      break
    case 'clear-tokens':
      await ClearTokenCache()
      message.success('Token cache cleared')
      break
    case 'about':
      const info = await GetAppInfo()
      message.info(`${info?.Name || 'otester'} v${info?.Version || '1.0.0'}`)
      break
  }
}
</script>

<style scoped>
.main-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.top-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
}

.menu-section {
  display: flex;
  gap: 16px;
}

.menu-item {
  cursor: pointer;
  color: var(--text-primary);
  padding: 4px 8px;
  border-radius: 4px;
}

.menu-item:hover {
  background: var(--border-color);
}

.content-area {
  display: flex;
  flex: 1;
  overflow: hidden;
  gap: 16px;
  padding: 16px;
}

.left-panel {
  width: 280px;
  min-width: 200px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.center-panel {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.right-panel {
  width: 400px;
  min-width: 300px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
