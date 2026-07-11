<template>
  <div class="request-editor">
    <div class="url-bar">
      <select v-model="method" class="method-select">
        <option value="GET">GET</option>
        <option value="POST">POST</option>
        <option value="PUT">PUT</option>
        <option value="PATCH">PATCH</option>
        <option value="DELETE">DELETE</option>
        <option value="HEAD">HEAD</option>
        <option value="OPTIONS">OPTIONS</option>
      </select>

      <input
        v-model="url"
        type="text"
        placeholder="Enter request URL..."
        class="url-input acrylic-input"
        @keydown.enter="sendRequest"
      />

      <button
        class="acrylic-btn send-btn"
        :disabled="sending || !url"
        @click="sendRequest"
      >
        {{ sending ? 'Cancel' : 'Send' }}
      </button>
    </div>

    <n-tabs type="line" animated>
      <n-tab-pane name="headers" tab="Headers">
        <KeyValueEditor v-model="headers" />
      </n-tab-pane>

      <n-tab-pane name="params" tab="Query Params">
        <KeyValueEditor v-model="queryParams" />
      </n-tab-pane>

      <n-tab-pane name="body" tab="Body">
        <div class="body-editor">
          <select v-model="bodyType" class="body-type-select">
            <option value="none">None</option>
            <option value="json">JSON</option>
            <option value="text">Text</option>
            <option value="x-www-form-urlencoded">x-www-form-urlencoded</option>
            <option value="raw">Raw</option>
          </select>

          <div v-if="bodyType !== 'none'" class="body-content">
            <codemirror
              v-model="body"
              :style="{ height: '200px' }"
              :extensions="codeMirrorExtensions"
            />
          </div>
        </div>
      </n-tab-pane>

      <n-tab-pane name="auth" tab="Auth">
        <div class="auth-panel">
          <label class="auth-checkbox">
            <input type="checkbox" v-model="useOAuth" />
            <span>Use OAuth 2.0</span>
          </label>

          <div v-if="useOAuth" class="oauth-config">
            <label>
              <span>OAuth Profile</span>
              <select v-model="oauthProfileId" class="acrylic-input">
                <option value="">Select profile...</option>
                <option
                  v-for="profile in oauthProfiles"
                  :key="profile.id"
                  :value="profile.id"
                >
                  {{ profile.name }}
                </option>
              </select>
            </label>

            <div v-if="tokenStatus" class="token-status">
              <span v-if="tokenStatus.hasToken" class="token-valid">
                Token valid{{ tokenStatus.fromCache ? ' (cached)' : '' }}
              </span>
              <span v-else class="token-missing">
                No token cached
              </span>
            </div>
          </div>
        </div>
      </n-tab-pane>
    </n-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NTabs, NTabPane } from 'naive-ui'
import { useRequestStore } from '../stores/request'
import { useConfigStore } from '../stores/config'
import { useResponseStore } from '../stores/response'
import { useOAuthStore } from '../stores/oauth'
import KeyValueEditor from './KeyValueEditor.vue'
import type { KeyValue } from '../types'

const requestStore = useRequestStore()
const configStore = useConfigStore()
const responseStore = useResponseStore()
const oauthStore = useOAuthStore()

const method = computed({
  get: () => requestStore.method,
  set: (v) => { requestStore.method = v }
})

const url = computed({
  get: () => requestStore.url,
  set: (v) => { requestStore.url = v }
})

const headers = computed({
  get: () => requestStore.headers,
  set: (v) => { requestStore.headers = v }
})

const queryParams = computed({
  get: () => requestStore.queryParams,
  set: (v) => { requestStore.queryParams = v }
})

const bodyType = computed({
  get: () => requestStore.bodyType,
  set: (v) => { requestStore.bodyType = v }
})

const body = computed({
  get: () => requestStore.body,
  set: (v) => { requestStore.body = v }
})

const sending = computed(() => requestStore.sending)

const useOAuth = computed({
  get: () => requestStore.useOAuth,
  set: (v) => { requestStore.useOAuth = v }
})

const oauthProfileId = computed({
  get: () => requestStore.oauthProfileId,
  set: (v) => { requestStore.oauthProfileId = v }
})

const oauthProfiles = computed(() => configStore.config?.oauthProfiles || [])

const tokenStatus = computed(() =>
  oauthProfileId.value ? oauthStore.tokenStatuses[oauthProfileId.value] : null
)

// Placeholder extensions for CodeMirror - actual extensions need codemirror task
const codeMirrorExtensions = ref([])

async function sendRequest() {
  if (sending.value) {
    // Cancel
    if (requestStore.requestId) {
      await window.go?.app.CancelRequest(requestStore.requestId)
    }
    requestStore.sending = false
    return
  }

  const requestId = requestStore.generateRequestId()
  requestStore.requestId = requestId
  requestStore.sending = true
  responseStore.loading = true

  try {
    const result = await window.go?.app.SendRequest({
      requestId,
      endpointId: configStore.selectedEndpointId || '',
      method: method.value,
      url: configStore.substituteVariables(url.value),
      headers: headers.value.filter(h => h.enabled),
      queryParams: queryParams.value.filter(q => q.enabled),
      bodyType: bodyType.value,
      body: body.value,
      timeoutSeconds: requestStore.timeoutSeconds,
      useOAuth: useOAuth.value,
      oauthProfileId: oauthProfileId.value,
    })

    responseStore.setResponse(result)

    if (useOAuth.value && oauthProfileId.value) {
      const status = await window.go?.app.GetTokenStatus(oauthProfileId.value)
      if (status) {
        oauthStore.updateStatus(oauthProfileId.value, status)
      }
    }
  } catch (e) {
    console.error('Request failed:', e)
  } finally {
    requestStore.sending = false
  }
}
</script>

<style scoped>
.request-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 16px;
  gap: 12px;
}

.url-bar {
  display: flex;
  gap: 8px;
}

.method-select {
  width: 100px;
  padding: 8px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--input-bg);
  color: var(--text-primary);
}

.url-input {
  flex: 1;
}

.send-btn {
  min-width: 80px;
}

.body-editor {
  padding: 12px 0;
}

.body-type-select {
  margin-bottom: 8px;
  padding: 6px;
  border-radius: 6px;
  border: 1px solid var(--border-color);
  background: var(--input-bg);
  color: var(--text-primary);
}

.auth-panel {
  padding: 12px 0;
}

.auth-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.oauth-config {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.oauth-config label {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.token-status {
  font-size: 13px;
  padding: 8px;
  border-radius: 6px;
  background: var(--bg-primary);
}

.token-valid {
  color: var(--success-color);
}

.token-missing {
  color: var(--text-secondary);
}
</style>
