<template>
  <div class="endpoint-list">
    <div class="search-box">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Search endpoints..."
        class="acrylic-input search-input"
      />
    </div>

    <div class="filter-row">
      <select v-model="methodFilter" class="method-filter">
        <option value="">All Methods</option>
        <option value="GET">GET</option>
        <option value="POST">POST</option>
        <option value="PUT">PUT</option>
        <option value="PATCH">PATCH</option>
        <option value="DELETE">DELETE</option>
      </select>
    </div>

    <div class="groups">
      <div
        v-for="group in filteredGroups"
        :key="group.id"
        class="group"
      >
        <div class="group-header" @click="toggleGroup(group.id)">
          <span class="group-name">{{ group.name }}</span>
          <span class="group-toggle">{{ collapsedGroups[group.id] ? '+' : '-' }}</span>
        </div>

        <div v-if="!collapsedGroups[group.id]" class="endpoints">
          <div
            v-for="endpoint in getEndpointsByGroup(group.id)"
            :key="endpoint.id"
            class="endpoint-item"
            :class="{ selected: selectedEndpointId === endpoint.id, disabled: !endpoint.enabled }"
            @click="selectEndpoint(endpoint)"
          >
            <span class="method-badge" :class="endpoint.method.toLowerCase()">
              {{ endpoint.method }}
            </span>
            <span class="endpoint-name">{{ endpoint.name }}</span>
            <span v-if="endpoint.hasOAuth" class="oauth-icon">🔒</span>
          </div>
        </div>
      </div>

      <div v-if="ungroupedEndpoints.length > 0" class="group">
        <div class="group-header">
          <span class="group-name">Ungrouped</span>
        </div>
        <div class="endpoints">
          <div
            v-for="endpoint in ungroupedEndpoints"
            :key="endpoint.id"
            class="endpoint-item"
            :class="{ selected: selectedEndpointId === endpoint.id, disabled: !endpoint.enabled }"
            @click="selectEndpoint(endpoint)"
          >
            <span class="method-badge" :class="endpoint.method.toLowerCase()">
              {{ endpoint.method }}
            </span>
            <span class="endpoint-name">{{ endpoint.name }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useConfigStore } from '../stores/config'
import { useRequestStore } from '../stores/request'
import { useResponseStore } from '../stores/response'
import type { EndpointView } from '../types'

const configStore = useConfigStore()
const requestStore = useRequestStore()
const responseStore = useResponseStore()

const searchQuery = ref('')
const methodFilter = ref('')
const collapsedGroups = ref<Record<string, boolean>>({})

const selectedEndpointId = computed(() => configStore.selectedEndpointId)

const filteredEndpoints = computed(() => {
  let eps = configStore.config?.endpoints || []

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    eps = eps.filter(ep =>
      ep.name.toLowerCase().includes(q) ||
      ep.url.toLowerCase().includes(q)
    )
  }

  if (methodFilter.value) {
    eps = eps.filter(ep => ep.method === methodFilter.value)
  }

  return eps
})

const filteredGroups = computed(() => {
  const groupIds = new Set(filteredEndpoints.value.map(ep => ep.groupId))
  return (configStore.config?.endpointGroups || []).filter(g => groupIds.has(g.id))
})

const ungroupedEndpoints = computed(() =>
  filteredEndpoints.value.filter(ep => !ep.groupId)
)

function getEndpointsByGroup(groupId: string) {
  return filteredEndpoints.value.filter(ep => ep.groupId === groupId)
}

function toggleGroup(groupId: string) {
  collapsedGroups.value[groupId] = !collapsedGroups.value[groupId]
}

function selectEndpoint(endpoint: EndpointView) {
  if (!endpoint.enabled) return
  configStore.selectEndpoint(endpoint.id)
  requestStore.loadFromEndpoint(endpoint)
  responseStore.clear()
}
</script>

<style scoped>
.endpoint-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 12px;
}

.search-input {
  width: 100%;
  margin-bottom: 8px;
}

.method-filter {
  width: 100%;
  margin-bottom: 12px;
  padding: 6px 8px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--input-bg);
  color: var(--text-primary);
}

.groups {
  flex: 1;
  overflow-y: auto;
}

.group {
  margin-bottom: 8px;
}

.group-header {
  display: flex;
  justify-content: space-between;
  padding: 6px 8px;
  cursor: pointer;
  border-radius: 6px;
  background: var(--bg-primary);
  margin-bottom: 4px;
}

.group-header:hover {
  background: var(--border-color);
}

.group-name {
  font-weight: 600;
  font-size: 13px;
}

.endpoints {
  padding-left: 8px;
}

.endpoint-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  cursor: pointer;
  border-radius: 6px;
  transition: background 0.15s;
}

.endpoint-item:hover {
  background: var(--bg-primary);
}

.endpoint-item.selected {
  background: var(--accent-color);
  color: white;
}

.endpoint-item.selected .method-badge {
  background: rgba(255,255,255,0.2);
  color: white;
}

.endpoint-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.method-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.method-badge.get { color: #61affe; }
.method-badge.post { color: #49cc90; }
.method-badge.put { color: #fca130; }
.method-badge.patch { color: #50e3c2; }
.method-badge.delete { color: #f93e3e; }
.method-badge.head { color: #9012fe; }
.method-badge.options { color: #0d5aa7; }

.endpoint-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.oauth-icon {
  font-size: 12px;
}
</style>
