<template>
  <n-config-provider>
    <n-message-provider>
      <n-dialog-provider>
        <div class="app-container">
          <h1>otester</h1>
        </div>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { NConfigProvider, NMessageProvider, NDialogProvider } from 'naive-ui'
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  width: 100%;
  height: 100%;
}

.app-container {
  width: 100%;
  height: 100%;
  padding: 16px;
}
</style>
