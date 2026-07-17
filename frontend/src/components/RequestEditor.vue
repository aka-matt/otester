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
            <option value="xml">XML</option>
            <option value="x-www-form-urlencoded">x-www-form-urlencoded</option>
            <option value="raw">Raw</option>
          </select>

          <div v-if="bodyType !== 'none'" class="body-content">
            <textarea
              v-model="body"
              class="body-textarea"
              :placeholder="bodyPlaceholder"
              spellcheck="false"
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
import { computed, onBeforeUnmount, onMounted, watch } from 'vue'
import { NTabs, NTabPane } from 'naive-ui'
import { useRequestStore } from '../stores/request'
import { useConfigStore } from '../stores/config'
import { useResponseStore } from '../stores/response'
import { useOAuthStore } from '../stores/oauth'
import KeyValueEditor from './KeyValueEditor.vue'
import { SendRequest, CancelRequest } from '../../wailsjs/go/app/App'
import { model } from '../../wailsjs/go/models'

const requestStore = useRequestStore()
const configStore = useConfigStore()
const responseStore = useResponseStore()
const oauthStore = useOAuthStore()

const method = computed({
  get: () => requestStore.method,
  set: (v) => { requestStore.method = v }
})

// The URL bar shows the live-substituted URL (selected variable applied).
// The underlying requestStore.url keeps the raw template so variable
// changes propagate automatically and the user can still type their own URL.
const url = computed({
  get: () => configStore.substituteVariables(requestStore.url),
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

const bodyPlaceholder = computed(() => {
  switch (bodyType.value) {
    case 'json':
      return '{\n  key: value\n}'
    case 'xml':
      return '<?xml version="1.0"?>\n<root>\n  <item/>\n</root>'
    default:
      return 'Enter request body...'
  }
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

async function sendRequest() {
  console.log('[DEBUG] sendRequest called', { sending: sending.value, url: url.value, method: method.value })
  if (sending.value) {
    // Cancel
    if (requestStore.requestId) {
      await CancelRequest(requestStore.requestId)
    }
    requestStore.sending = false
    return
  }

  const requestId = requestStore.generateRequestId()
  requestStore.requestId = requestId
  requestStore.sending = true
  responseStore.loading = true

  try {
    console.log('[DEBUG] calling SendRequest', { requestId, url: url.value })
    const result = await SendRequest(model.RequestInput.createFrom({
      requestId,
      endpointId: configStore.selectedEndpointId || '',
      method: method.value,
      url: configStore.substituteVariables(requestStore.url),
      headers: headers.value.filter(h => h.enabled),
      queryParams: queryParams.value.filter(q => q.enabled),
      bodyType: bodyType.value,
      body: body.value,
      timeoutSeconds: requestStore.timeoutSeconds,
      useOAuth: useOAuth.value,
      oauthProfileId: oauthProfileId.value,
    }))
    console.log('[DEBUG] SendRequest result', result)

    if (result) {
      responseStore.setResponse(result)
    }

    if (useOAuth.value && oauthProfileId.value) {
      await oauthStore.refreshStatus(oauthProfileId.value)
      scheduleExpiryRefresh()
    }
  } catch (e) {
    console.error('[DEBUG] Request failed:', e)
  } finally {
    requestStore.sending = false
    console.log('[DEBUG] sendRequest done')
  }
}

// Periodic refresh so the UI flips back to "No token cached" once the
// refresh_before_expiry_seconds window starts (the backend enforces this;
// we just re-poll to keep the indicator accurate).
let expiryTimer: ReturnType<typeof setTimeout> | null = null
function scheduleExpiryRefresh() {
  if (expiryTimer) {
    clearTimeout(expiryTimer)
    expiryTimer = null
  }
  const status = tokenStatus.value
  if (!status || !status.hasToken || !status.expiresAt) return
  const ms = new Date(status.expiresAt).getTime() - Date.now()
  // Cap to 5 minutes so we don't park a long timer forever; the watcher on
  // oauthProfileId and any post-send refresh will re-arm if needed.
  const delay = Math.max(1000, Math.min(ms, 5 * 60 * 1000))
  expiryTimer = setTimeout(async () => {
    if (oauthProfileId.value) {
      await oauthStore.refreshStatus(oauthProfileId.value)
      scheduleExpiryRefresh()
    }
  }, delay)
}

watch(oauthProfileId, async (id) => {
  if (id) await oauthStore.refreshStatus(id)
  scheduleExpiryRefresh()
})

watch(useOAuth, async (enabled) => {
  if (enabled && oauthProfileId.value) {
    await oauthStore.refreshStatus(oauthProfileId.value)
    scheduleExpiryRefresh()
  }
})

onMounted(() => {
  if (useOAuth.value && oauthProfileId.value) {
    oauthStore.refreshStatus(oauthProfileId.value).then(scheduleExpiryRefresh)
  }
})

onBeforeUnmount(() => {
  if (expiryTimer) {
    clearTimeout(expiryTimer)
    expiryTimer = null
  }
})
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

.body-textarea {
  width: 100%;
  min-height: 200px;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--border-color);
  background: var(--input-bg);
  color: var(--text-primary);
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  resize: vertical;
}

.body-textarea:focus {
  outline: none;
  border-color: var(--accent-color);
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
