import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import ResponseViewer from './ResponseViewer.vue'
import { useResponseStore } from '../stores/response'
import type { ResponseOutput, TLSInfo } from '../types'

const message = { success: vi.fn(), error: vi.fn(), warning: vi.fn(), info: vi.fn(), loading: vi.fn() }

vi.mock('naive-ui', async () => {
  const actual = await vi.importActual<typeof import('naive-ui')>('naive-ui')
  return {
    ...actual,
    useMessage: () => message,
  }
})

const baseResponse = (overrides: Partial<ResponseOutput> = {}): ResponseOutput => ({
  requestId: 'r1',
  statusCode: 200,
  status: '200 OK',
  headers: { 'Content-Type': ['application/json'] },
  body: '{"ok":true}',
  bodyTruncated: false,
  durationMs: 100,
  sizeBytes: 12,
  contentType: 'application/json',
  usedOAuth: false,
  tokenFromCache: false,
  errorCode: '',
  errorMessage: '',
  ...overrides,
})

describe('ResponseViewer Certificate tab', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('renders the Certificate tab pane', () => {
    const store = useResponseStore()
    store.setResponse(baseResponse())
    const w = mount(ResponseViewer)
    expect(w.text()).toContain('Certificate')
  })

  it('Certificate tab is visible when response has no tls field', () => {
    const store = useResponseStore()
    store.setResponse(baseResponse({ tls: undefined as any }))
    const w = mount(ResponseViewer)
    // The tab title is always rendered; clicking it shows the no-TLS panel.
    expect(w.text()).toContain('Certificate')
  })

  it('Certificate tab is visible when response has tls info', () => {
    const store = useResponseStore()
    const tls: TLSInfo = {
      status: 'ok',
      targetHost: 'example.com:443',
      connection: {
        version: 'TLS 1.3',
        cipherSuite: '0x1301',
        cipherSuiteName: 'TLS_AES_128_GCM_SHA256',
        serverName: 'example.com',
        resumed: false,
        ocspStapled: false,
        peerCertificates: 1,
      },
      certificates: [
        {
          position: 'leaf',
          subject: 'CN=example.com',
          issuer: 'CN=Test CA',
          serialNumber: '1',
          version: 3,
          signatureAlgorithm: 'SHA256-RSA',
          notBefore: '2026-01-01T00:00:00Z',
          notAfter: '2027-01-01T00:00:00Z',
          isExpired: false,
          isNotYetValid: false,
          daysToExpiry: 200,
          subjectKeyId: '',
          authorityKeyId: '',
          sans: ['DNS:example.com'],
          keyAlgorithm: 'RSA',
          keySize: 2048,
          publicKeyPem: '',
          fingerprintSha1: '',
          fingerprintSha256: '',
          isCa: false,
          maxPathLength: -1,
          keyUsage: [],
          extendedKeyUsage: [],
          rawDer: '',
          pem: '',
          signatureBytes: '',
        },
      ],
    }
    store.setResponse(baseResponse({ tls }))
    const w = mount(ResponseViewer)
    expect(w.text()).toContain('Certificate')
    expect(w.text()).toContain('TLS 1.3')
  })
})
