<template>
  <div class="response-viewer">
    <div v-if="!response" class="empty-state">
      <p>Send a request to see the response</p>
    </div>

    <div v-else class="response-content">
      <div class="response-overview acrylic-card">
        <div class="status-display">
          <span class="status-badge" :class="statusClass">{{ response.statusCode }}</span>
          <span class="status-text">{{ statusText }}</span>
        </div>
        <div class="meta-info">
          <span class="meta-item">
            <span class="label">Time:</span>
            <span class="value">{{ response.durationMs }}ms</span>
          </span>
          <span class="meta-item">
            <span class="label">Size:</span>
            <span class="value">{{ formatSize(response.sizeBytes) }}</span>
          </span>
          <span v-if="response.usedOAuth" class="meta-item">
            <span class="label">OAuth:</span>
            <span class="value">{{ response.tokenFromCache ? 'Cached' : 'Fresh' }}</span>
          </span>
        </div>
        <div v-if="response.bodyTruncated" class="truncation-warning">
          ⚠ Response truncated (exceeds 10MB limit)
        </div>
      </div>

      <n-tabs type="line" animated>
        <n-tab-pane name="body" tab="Body">
          <div class="body-content">
            <pre class="body-pre" v-html="formattedBody"></pre>
          </div>
        </n-tab-pane>

        <n-tab-pane name="headers" tab="Headers">
          <div class="headers-table">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Value</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(values, name) in response.headers" :key="name">
                  <td class="header-name">{{ name }}</td>
                  <td class="header-value">{{ values.join(', ') }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </n-tab-pane>

        <n-tab-pane name="certificate" tab="Certificate" display-directive="show">
          <CertificateViewer :info="response.tls ?? null" />
        </n-tab-pane>
      </n-tabs>

      <div v-if="response.errorCode" class="error-panel">
        <div class="error-code">{{ response.errorCode }}</div>
        <div class="error-message">{{ response.errorMessage }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NTabs, NTabPane } from 'naive-ui'
import { useResponseStore } from '../stores/response'
import CertificateViewer from './CertificateViewer.vue'

const responseStore = useResponseStore()
const response = computed(() => responseStore.response)

const statusClass = computed(() => {
  if (!response.value) return ''
  const code = response.value.statusCode
  if (code >= 200 && code < 300) return 'status-success'
  if (code >= 300 && code < 400) return 'status-redirect'
  if (code >= 400 && code < 500) return 'status-client-error'
  return 'status-server-error'
})

// Strip the leading "200 " / "404 " numeric prefix from response.status (which
// comes from net/http as e.g. "200 OK") so we don't render the code twice.
const statusText = computed(() => {
  if (!response.value) return ''
  return response.value.status.replace(/^\d+\s+/, '')
})

const formattedBody = computed(() => {
  if (!response.value) return ''
  const body = response.value.body
  if (isJSON(body)) {
    try {
      return syntaxHighlightJSON(JSON.stringify(JSON.parse(body), null, 2))
    } catch {
      return escapeHtml(body)
    }
  }
  return escapeHtml(body)
})

function isJSON(str: string): boolean {
  try {
    JSON.parse(str)
    return true
  } catch {
    return false
  }
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

function syntaxHighlightJSON(json: string): string {
  return json
    .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g,
      (match) => {
        let cls = 'json-number'
        if (/^"/.test(match)) {
          cls = /:$/.test(match) ? 'json-key' : 'json-string'
        } else if (/true|false/.test(match)) {
          cls = 'json-boolean'
        } else if (/null/.test(match)) {
          cls = 'json-null'
        }
        return '<span class="' + cls + '">' + match + '</span>'
      }
    )
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>

<style scoped>
.response-viewer {
  height: 100%;
  padding: 16px;
  overflow-y: auto;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

.response-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}

/* Make n-tabs stretch to fill remaining vertical space */
.response-content :deep(.n-tabs) {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  min-height: 0;
}

.response-content :deep(.n-tabs-nav) {
  flex-shrink: 0;
}

.response-content :deep(.n-tabs-pane-wrapper),
.response-content :deep(.n-tab-pane) {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  min-height: 0;
}

.response-overview {
  padding: 12px;
  flex-shrink: 0;
}

.status-display {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.status-badge {
  display: inline-block;
  min-width: 48px;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  text-align: center;
  font-family: 'Consolas', 'Monaco', monospace;
  letter-spacing: 0.5px;
}

.status-text {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
}

.status-success   { background: var(--success-color); }
.status-redirect   { background: var(--accent-color); }
.status-client-error { background: var(--warning-color); }
.status-server-error { background: var(--danger-color); }

.meta-info {
  display: flex;
  gap: 16px;
  font-size: 13px;
}

.meta-item .label {
  color: var(--text-secondary);
  margin-right: 4px;
}

.truncation-warning {
  margin-top: 8px;
  font-size: 12px;
  color: var(--warning-color);
}

.body-pre {
  background: var(--bg-primary);
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
}

.headers-table {
  flex: 1 1 0;
  overflow-y: auto;
}

.headers-table table {
  width: 100%;
  border-collapse: collapse;
}

.headers-table th,
.headers-table td {
  text-align: left;
  padding: 8px;
  border-bottom: 1px solid var(--border-color);
}

.header-name {
  font-weight: 500;
  color: var(--accent-color);
}

.header-value {
  word-break: break-all;
}

.error-panel {
  padding: 12px;
  border-radius: 8px;
  background: rgba(211, 47, 47, 0.1);
  border: 1px solid var(--danger-color);
}

.error-code {
  font-weight: 600;
  color: var(--danger-color);
  margin-bottom: 4px;
}

.error-message {
  font-size: 13px;
  color: var(--text-secondary);
}
</style>
