<template>
  <div class="endpoint-list">
    <div class="variable-selector" aria-label="Configuration variable">
      <button
        v-for="(variable, index) in configStore.config?.variables ?? []"
        :key="`${variable.id}-${variable.environment}`"
        type="button"
        class="variable-option"
        :class="{ selected: configStore.selectedVariableIndex === index }"
        :aria-pressed="configStore.selectedVariableIndex === index"
        data-testid="variable-option"
        @click="selectVariable(index)"
      >
        {{ variable.environment ? `${variable.environment} (${variable.id})` : variable.id }}
      </button>
      <p v-if="(configStore.config?.variables ?? []).length === 0" class="variable-empty">
        No configuration variables available.
      </p>
    </div>

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
      <section
        v-for="group in filteredGroups"
        :key="group.id"
        class="group endpoint-group-card"
        :style="{ '--group-accent': groupAccent(group.id) }"
        data-testid="endpoint-group-card"
      >
        <button type="button" class="group-header" @click="toggleGroup(group.id)">
          <span class="group-name">{{ group.name }}</span>
          <span class="group-toggle">{{ collapsedGroups[group.id] ? '+' : '-' }}</span>
        </button>

        <div v-if="!collapsedGroups[group.id]" class="endpoints">
          <button
            v-for="endpoint in getEndpointsByGroup(group.id)"
            :key="endpoint.id"
            type="button"
            class="endpoint-item"
            :class="{ selected: selectedEndpointId === endpoint.id, disabled: !endpoint.enabled }"
            :disabled="!endpoint.enabled"
            @click="selectEndpoint(endpoint)"
          >
            <span class="method-badge" :class="endpoint.method.toLowerCase()">
              {{ endpoint.method }}
            </span>
            <span class="endpoint-name">{{ endpoint.name }}</span>
            <span v-if="endpoint.hasOAuth" class="oauth-icon">🔒</span>
          </button>
        </div>
      </section>

      <section
        v-if="ungroupedEndpoints.length > 0"
        class="group endpoint-group-card"
        :style="{ '--group-accent': 'var(--border-color)' }"
        data-testid="ungrouped-card"
      >
        <div class="group-header">
          <span class="group-name">Ungrouped</span>
        </div>
        <div class="endpoints">
          <button
            v-for="endpoint in ungroupedEndpoints"
            :key="endpoint.id"
            type="button"
            class="endpoint-item"
            :class="{ selected: selectedEndpointId === endpoint.id, disabled: !endpoint.enabled }"
            :disabled="!endpoint.enabled"
            @click="selectEndpoint(endpoint)"
          >
            <span class="method-badge" :class="endpoint.method.toLowerCase()">
              {{ endpoint.method }}
            </span>
            <span class="endpoint-name">{{ endpoint.name }}</span>
            <span v-if="endpoint.hasOAuth" class="oauth-icon">🔒</span>
          </button>
        </div>
      </section>

      <p v-if="filteredEndpoints.length === 0" class="empty-search">No endpoints match your search.</p>
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
const groupPalette = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ec4899', '#06b6d4']

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

const knownGroupIds = computed(() => new Set((configStore.config?.endpointGroups ?? []).map(g => g.id)))

const ungroupedEndpoints = computed(() =>
  filteredEndpoints.value.filter(ep => !knownGroupIds.value.has(ep.groupId))
)

function getEndpointsByGroup(groupId: string) {
  return filteredEndpoints.value.filter(ep => ep.groupId === groupId)
}

function toggleGroup(groupId: string) {
  collapsedGroups.value[groupId] = !collapsedGroups.value[groupId]
}

function groupAccent(groupId: string) {
  const index = (configStore.config?.endpointGroups ?? []).findIndex(g => g.id === groupId)
  return groupPalette[index % groupPalette.length]
}

function loadEndpoint(endpoint: EndpointView) {
  requestStore.loadFromEndpoint({ ...endpoint, url: configStore.substituteVariables(endpoint.url) })
}

function selectEndpoint(endpoint: EndpointView) {
  if (!endpoint.enabled) return
  configStore.selectEndpoint(endpoint.id)
  loadEndpoint(endpoint)
  responseStore.clear()
}

function selectVariable(index: number) {
  configStore.selectVariable(index)
  if (configStore.selectedEndpoint) loadEndpoint(configStore.selectedEndpoint)
}
</script>

<style scoped>
.endpoint-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 12px;
}

.variable-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.variable-option {
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 5px 8px;
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
}

.variable-option:hover:not(:disabled) {
  border-color: var(--accent-color);
}

.variable-option.selected {
  background: var(--accent-color);
  border-color: var(--accent-color);
  color: white;
}

.variable-option:focus-visible,
.group-header:focus-visible,
.endpoint-item:focus-visible {
  outline: 2px solid var(--accent-color);
  outline-offset: 2px;
}

.variable-option:disabled,
.endpoint-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.variable-empty {
  margin: 0;
  color: var(--text-secondary);
  font-size: 13px;
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
  border: 1px solid var(--group-accent);
  border-left-width: 4px;
  border-radius: 6px;
  overflow: hidden;
}

.group-header {
  display: flex;
  width: 100%;
  justify-content: space-between;
  padding: 6px 8px;
  cursor: pointer;
  border: 0;
  background: var(--bg-primary);
  color: var(--text-primary);
  margin-bottom: 4px;
  text-align: left;
}

.group-header:hover {
  background: var(--border-color);
}

.group-name {
  font-weight: 600;
  font-size: 13px;
  color: var(--group-accent);
}

.endpoints {
  padding-left: 8px;
}

.endpoint-item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 8px;
  padding: 8px;
  cursor: pointer;
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: var(--text-primary);
  text-align: left;
  transition: background 0.15s;
}

.endpoint-item:hover:not(:disabled) {
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

.empty-search {
  margin: 16px 8px;
  color: var(--text-secondary);
  font-size: 13px;
  text-align: center;
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
