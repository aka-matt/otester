<template>
  <n-config-provider :theme="theme">
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

const themeStore = useThemeStore()
const configStore = useConfigStore()

onMounted(async () => {
  themeStore.init()
  await configStore.loadConfig()
})
</script>

<style>
@import './themes/acrylic.css';
</style>
