<template>
  <div class="main-view">
    <header class="top-bar">
      <div class="menu-section">
        <span class="menu-item">File</span>
        <span class="menu-item">View</span>
        <span class="menu-item">Tools</span>
        <span class="menu-item">Help</span>
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
import EndpointList from '../components/EndpointList.vue'
import RequestEditor from '../components/RequestEditor.vue'
import ResponseViewer from '../components/ResponseViewer.vue'
import ThemeSwitcher from '../components/ThemeSwitcher.vue'
import StatusBar from '../components/StatusBar.vue'
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
