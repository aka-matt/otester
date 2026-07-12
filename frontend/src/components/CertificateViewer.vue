<template>
  <div class="certificate-viewer">
    <!-- 1. empty -->
    <div v-if="info === null" class="empty-state">
      <p>No request sent yet.</p>
    </div>

    <!-- 2. no TLS attempted -->
    <div v-else-if="info.status === 'no_tls_attempted'" class="info-panel acrylic-card">
      <div class="panel-title">No TLS</div>
      <div class="panel-body">{{ info.error || 'This request did not use HTTPS — certificate info is not applicable.' }}</div>
      <div class="panel-meta">Target: {{ info.targetHost }}</div>
    </div>

    <!-- 3. handshake failed -->
    <div v-else-if="info.status === 'handshake_failed'" class="error-panel">
      <div class="panel-title">⚠ TLS handshake failed</div>
      <div class="panel-body">{{ info.error }}</div>
      <div class="panel-meta">
        Target: {{ info.targetHost }}
        <span v-if="info.attemptedServerName"> · SNI attempted: {{ info.attemptedServerName }}</span>
      </div>
    </div>

    <!-- 4. ok -->
    <div v-else class="ok-content">
      <div v-if="info.validationSkipped" class="warning-panel">
        <div class="panel-title">⚠ TLS certificate validation was skipped</div>
        <div class="panel-body">
          <div class="warning-reason">
            <code>allow_insecure_tls: true</code> in <code>config.json</code> caused this request to fall back to an insecure connection.
          </div>
          <div v-if="info.originalError" class="warning-original-error">
            Original error: <code>{{ info.originalError }}</code>
          </div>
        </div>
      </div>

      <!-- TLS connection block -->
      <div class="connection-block acrylic-card">
        <div class="block-title">TLS Connection</div>
        <dl class="kv-grid">
          <dt>Version</dt>
          <dd>{{ info.connection?.version }}</dd>

          <dt>Cipher suite</dt>
          <dd>
            <code>{{ info.connection?.cipherSuite }}</code>
            <span class="muted">({{ info.connection?.cipherSuiteName }})</span>
          </dd>

          <dt>Negotiated protocol (ALPN)</dt>
          <dd>{{ info.connection?.negotiatedProtocol || '—' }}</dd>

          <dt>Server name (SNI)</dt>
          <dd>{{ info.connection?.serverName || '—' }}</dd>

          <dt>Resumed session</dt>
          <dd>{{ info.connection?.resumed ? 'Yes' : 'No' }}</dd>

          <dt>Peer certificates</dt>
          <dd>{{ info.connection?.peerCertificates }}</dd>

          <dt>OCSP stapled</dt>
          <dd>{{ info.connection?.ocspStapled ? 'Yes' : 'No' }}</dd>

          <dt>SCTs</dt>
          <dd>{{ info.connection?.scts?.length ?? 0 }}</dd>
        </dl>
      </div>

      <!-- Certificate chain -->
      <div class="chain-block">
        <div class="block-title">Certificate Chain ({{ info.certificates.length }})</div>
        <n-collapse :default-expanded-names="defaultExpanded">
          <n-collapse-item
            v-for="(cert, idx) in info.certificates"
            :key="idx"
            :name="idx"
            :title="cert.position + ' — ' + commonName(cert.subject)"
          >
            <div class="cert-panel">
              <section class="cert-section">
                <h4>Subject & Issuer</h4>
                <dl class="kv-grid">
                  <dt>Subject</dt><dd class="dn">{{ cert.subject }}</dd>
                  <dt>Issuer</dt><dd class="dn">{{ cert.issuer }}</dd>
                  <dt>Serial</dt><dd><code>{{ cert.serialNumber }}</code></dd>
                  <dt>Version</dt><dd>{{ cert.version }}</dd>
                  <dt>Signature</dt><dd>{{ cert.signatureAlgorithm }}</dd>
                </dl>
              </section>

              <section class="cert-section">
                <h4>Validity</h4>
                <dl class="kv-grid">
                  <dt>Not before</dt><dd>{{ cert.notBefore }}</dd>
                  <dt>Not after</dt><dd>{{ cert.notAfter }}</dd>
                  <dt>Days to expiry</dt><dd>{{ cert.daysToExpiry }}</dd>
                  <dt>Status</dt>
                  <dd>
                    <span class="pill" :class="validityClass(cert)">{{ validityLabel(cert) }}</span>
                  </dd>
                </dl>
              </section>

              <section class="cert-section">
                <h4>Subject Alternative Names</h4>
                <ul class="san-list">
                  <li v-for="(san, i) in cert.sans" :key="i">{{ san }}</li>
                </ul>
              </section>
              <section class="cert-section">
                <h4>Key</h4>
                <dl class="kv-grid">
                  <dt>Algorithm</dt><dd>{{ cert.keyAlgorithm }}</dd>
                  <dt>Size</dt><dd>{{ cert.keySize || '—' }} bits</dd>
                </dl>
                <details>
                  <summary>Public key (PEM)</summary>
                  <pre class="pem-block">{{ cert.publicKeyPem }}</pre>
                  <button class="copy-btn" :aria-label="'Copy public key PEM'" @click="copy(cert.publicKeyPem, 'Public key copied')">Copy Public Key</button>
                </details>
              </section>

              <section class="cert-section">
                <h4>Fingerprints</h4>
                <dl class="kv-grid">
                  <dt>SHA-256</dt>
                  <dd>
                    <code>{{ cert.fingerprintSha256 }}</code>
                    <button class="copy-btn" :aria-label="'Copy SHA-256 fingerprint'" @click="copy(cert.fingerprintSha256, 'SHA-256 copied')">Copy SHA-256</button>
                  </dd>
                  <dt>SHA-1</dt>
                  <dd>
                    <code>{{ cert.fingerprintSha1 }}</code>
                    <button class="copy-btn" :aria-label="'Copy SHA-1 fingerprint'" @click="copy(cert.fingerprintSha1, 'SHA-1 copied')">Copy SHA-1</button>
                  </dd>
                </dl>
              </section>

              <section class="cert-section">
                <h4>Extensions</h4>
                <dl class="kv-grid">
                  <dt>Is CA</dt><dd>{{ cert.isCa ? 'Yes' : 'No' }}</dd>
                  <dt>Max path length</dt><dd>{{ cert.maxPathLength < 0 ? '—' : cert.maxPathLength }}</dd>
                  <dt>Key usage</dt>
                  <dd>
                    <span v-for="ku in cert.keyUsage" :key="ku" class="badge">{{ ku }}</span>
                  </dd>
                  <dt>Extended key usage</dt>
                  <dd>
                    <span v-for="eku in cert.extendedKeyUsage" :key="eku" class="badge">{{ eku }}</span>
                  </dd>
                  <template v-if="cert.crlDistributionPoints?.length">
                    <dt>CRL distribution points</dt>
                    <dd>
                      <ul class="plain-list">
                        <li v-for="(p, i) in cert.crlDistributionPoints" :key="i">{{ p }}</li>
                      </ul>
                    </dd>
                  </template>
                  <template v-if="cert.policies?.length">
                    <dt>Policies</dt>
                    <dd>
                      <span v-for="p in cert.policies" :key="p" class="badge">{{ p }}</span>
                    </dd>
                  </template>
                </dl>
              </section>

              <details class="raw-section">
                <summary>Raw (PEM + signature)</summary>
                <h5>PEM</h5>
                <pre class="pem-block">{{ cert.pem }}</pre>
                <button class="copy-btn" :aria-label="'Copy certificate PEM'" @click="copy(cert.pem, 'Certificate PEM copied')">Copy PEM</button>
                <h5>Signature (base64)</h5>
                <pre class="pem-block">{{ cert.signatureBytes }}</pre>
                <h5>DER (base64)</h5>
                <pre class="pem-block">{{ cert.rawDer }}</pre>
              </details>
            </div>
          </n-collapse-item>
        </n-collapse>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { NCollapse, NCollapseItem, useMessage } from 'naive-ui'
import type { CertificateView, TLSInfo } from '../types'

const props = defineProps<{ info: TLSInfo | null }>()

const message = useMessage()

const defaultExpanded = computed(() => {
  if (props.info?.status !== 'ok') return []
  // Expand only the first (leaf) cert by default.
  return props.info.certificates.length > 0 ? [0] : []
})

function commonName(subject: string): string {
  const m = subject.match(/CN=([^,]+)/)
  return m ? m[1] : subject
}

function validityLabel(cert: CertificateView): string {
  if (cert.isNotYetValid) return 'Not yet valid'
  if (cert.isExpired) return 'Expired'
  if (cert.daysToExpiry <= 30) return 'Expires soon'
  return 'Valid'
}

function validityClass(cert: CertificateView): string {
  if (cert.isNotYetValid) return 'pill-warning'
  if (cert.isExpired) return 'pill-danger'
  if (cert.daysToExpiry <= 30) return 'pill-warning'
  return 'pill-success'
}

async function copy(text: string, msg: string) {
  try {
    await navigator.clipboard.writeText(text)
    message.success(msg)
  } catch {
    message.error('Copy failed')
  }
}
</script>
<style scoped>
.certificate-viewer {
  padding: 8px 0;
}

.empty-state {
  text-align: center;
  padding: 40px 16px;
  color: var(--text-secondary);
}

.info-panel,
.error-panel {
  padding: 16px;
  border-radius: 8px;
  margin-bottom: 12px;
}

.info-panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
}

.error-panel {
  background: rgba(211, 47, 47, 0.1);
  border: 1px solid var(--danger-color);
}

.panel-title {
  font-weight: 600;
  margin-bottom: 8px;
}

.panel-body {
  font-size: 13px;
  color: var(--text-primary);
  word-break: break-all;
}

.panel-meta {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.connection-block {
  padding: 12px;
  margin-bottom: 16px;
}

.block-title {
  font-weight: 600;
  margin-bottom: 8px;
}

.kv-grid {
  display: grid;
  grid-template-columns: max-content 1fr;
  gap: 6px 16px;
  margin: 0;
}

.kv-grid dt {
  color: var(--text-secondary);
  font-size: 12px;
}

.kv-grid dd {
  margin: 0;
  font-size: 13px;
  word-break: break-all;
}

.muted {
  color: var(--text-secondary);
  font-size: 12px;
}

.dn {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
}
.cert-panel {
  padding: 8px 4px;
}

.cert-section {
  margin-bottom: 16px;
}

.cert-section h4 {
  margin: 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.san-list,
.plain-list {
  list-style: disc;
  margin: 4px 0 4px 16px;
  padding: 0;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
}

.pem-block {
  background: var(--bg-primary);
  padding: 8px;
  border-radius: 6px;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  margin: 8px 0;
}

.copy-btn {
  padding: 4px 10px;
  font-size: 12px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
  color: var(--text-primary);
  cursor: pointer;
  margin-left: 8px;
}

.copy-btn:hover {
  background: var(--border-color);
}

.warning-panel {
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
  background: rgba(255, 152, 0, 0.1);
  border: 1px solid var(--warning-color);
}

.warning-panel .panel-title {
  color: var(--warning-color);
}

.warning-panel .warning-reason,
.warning-panel .warning-original-error {
  font-size: 13px;
  color: var(--text-primary);
  margin-top: 6px;
  word-break: break-all;
}

.warning-panel code {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  background: var(--bg-primary);
  padding: 1px 6px;
  border-radius: 4px;
}

.pill {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.pill-success {
  background: rgba(76, 175, 80, 0.15);
  color: var(--success-color);
}

.pill-warning {
  background: rgba(255, 152, 0, 0.15);
  color: var(--warning-color);
}

.pill-danger {
  background: rgba(211, 47, 47, 0.15);
  color: var(--danger-color);
}

.badge {
  display: inline-block;
  padding: 2px 8px;
  margin: 2px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 11px;
  font-family: 'Consolas', 'Monaco', monospace;
}

.raw-section {
  border-top: 1px solid var(--border-color);
  padding-top: 12px;
}

.raw-section h5 {
  margin: 8px 0 4px;
  font-size: 12px;
  color: var(--text-secondary);
  text-transform: uppercase;
}

code {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
}
</style>
