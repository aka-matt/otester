import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import CertificateViewer from './CertificateViewer.vue'
import type { CertificateView, TLSConnectionView, TLSInfo } from '../types'

const message = { success: vi.fn(), error: vi.fn(), warning: vi.fn(), info: vi.fn(), loading: vi.fn() }

vi.mock('naive-ui', async () => {
  const actual = await vi.importActual<typeof import('naive-ui')>('naive-ui')
  return {
    ...actual,
    useMessage: () => message,
  }
})

const fixtureConnection = (): TLSConnectionView => ({
  version: 'TLS 1.3',
  cipherSuite: '0x1301',
  cipherSuiteName: 'TLS_AES_128_GCM_SHA256',
  negotiatedProtocol: 'h2',
  serverName: 'example.com',
  resumed: false,
  ocspStapled: true,
  peerCertificates: 2,
  scts: ['AQID'],
})

const fixtureCert = (overrides: Partial<CertificateView> = {}): CertificateView => ({
  position: 'leaf',
  subject: 'CN=example.com',
  issuer: "CN=R3, O=Let's Encrypt",
  serialNumber: 'ABCDEF',
  version: 3,
  signatureAlgorithm: 'SHA256-RSA',
  notBefore: '2026-01-01T00:00:00Z',
  notAfter: '2027-01-01T00:00:00Z',
  isExpired: false,
  isNotYetValid: false,
  daysToExpiry: 365,
  subjectKeyId: '',
  authorityKeyId: '',
  sans: ['DNS:example.com', 'IP:1.2.3.4'],
  keyAlgorithm: 'RSA',
  keySize: 2048,
  publicKeyPem: '-----BEGIN PUBLIC KEY-----\nABCD\n-----END PUBLIC KEY-----',
  fingerprintSha1: 'AA:BB:CC',
  fingerprintSha256: '11:22:33:44:55',
  isCa: false,
  maxPathLength: -1,
  keyUsage: ['DigitalSignature'],
  extendedKeyUsage: ['ServerAuth'],
  rawDer: '',
  pem: '-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----',
  signatureBytes: '',
  ...overrides,
})

const fixtureTLS = (overrides: Partial<TLSInfo> = {}): TLSInfo => ({
  status: 'ok',
  targetHost: 'example.com:443',
  connection: fixtureConnection(),
  certificates: [fixtureCert(), fixtureCert({ position: 'root', subject: 'CN=Root CA' })],
  ...overrides,
})

describe('CertificateViewer', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.stubGlobal('navigator', {
      clipboard: { writeText: vi.fn().mockResolvedValue(undefined) },
    })
  })

  it('renders empty state when info is null', () => {
    const w = mount(CertificateViewer, { props: { info: null } })
    expect(w.text()).toContain('No request sent yet')
  })

  it('renders No TLS panel for no_tls_attempted', () => {
    const w = mount(CertificateViewer, {
      props: { info: { status: 'no_tls_attempted', targetHost: 'example.com:80', certificates: [] } },
    })
    expect(w.text()).toContain('No TLS')
    expect(w.text()).toContain('example.com:80')
  })

  it('renders handshake-failed panel with error', () => {
    const w = mount(CertificateViewer, {
      props: {
        info: {
          status: 'handshake_failed',
          targetHost: 'example.com:443',
          attemptedServerName: 'example.com',
          error: 'x509: certificate is valid for other.com',
          certificates: [],
        },
      },
    })
    expect(w.text()).toContain('TLS handshake failed')
    expect(w.text()).toContain('x509: certificate is valid for other.com')
    expect(w.text()).toContain('example.com')
  })

  it('renders connection block for ok status', () => {
    const w = mount(CertificateViewer, { props: { info: fixtureTLS() } })
    expect(w.text()).toContain('TLS 1.3')
    expect(w.text()).toContain('TLS_AES_128_GCM_SHA256')
    expect(w.text()).toContain('h2')
    expect(w.text()).toContain('example.com')
  })
  it('renders accordion with one item per cert', () => {
    const w = mount(CertificateViewer, {
      props: { info: fixtureTLS({ certificates: [fixtureCert(), fixtureCert({ position: 'intermediate' }), fixtureCert({ position: 'root' })] }) },
    })
    expect(w.text()).toContain('leaf')
    expect(w.text()).toContain('intermediate')
    expect(w.text()).toContain('root')
  })

  it('shows expired pill with danger styling', () => {
    const w = mount(CertificateViewer, {
      props: { info: fixtureTLS({ certificates: [fixtureCert({ isExpired: true, daysToExpiry: -10 })] }) },
    })
    expect(w.text()).toContain('Expired')
  })

  it('shows valid pill for unexpired cert', () => {
    const w = mount(CertificateViewer, {
      props: { info: fixtureTLS({ certificates: [fixtureCert({ daysToExpiry: 60 })] }) },
    })
    expect(w.text()).toContain('Valid')
  })

  it('shows "Expires soon" pill for cert within 30 days', () => {
    const w = mount(CertificateViewer, {
      props: { info: fixtureTLS({ certificates: [fixtureCert({ daysToExpiry: 15 })] }) },
    })
    expect(w.text()).toContain('Expires soon')
  })

  it('shows "Not yet valid" pill when isNotYetValid is true', () => {
    const w = mount(CertificateViewer, {
      props: { info: fixtureTLS({ certificates: [fixtureCert({ isNotYetValid: true, daysToExpiry: 365 })] }) },
    })
    expect(w.text()).toContain('Not yet valid')
  })

  it('renders SANs with type prefixes', () => {
    const w = mount(CertificateViewer, {
      props: { info: fixtureTLS({ certificates: [fixtureCert()] }) },
    })
    expect(w.text()).toContain('DNS:example.com')
    expect(w.text()).toContain('IP:1.2.3.4')
  })

  it('copy SHA-256 button calls clipboard.writeText', async () => {
    const w = mount(CertificateViewer, { props: { info: fixtureTLS() } })
    const buttons = w.findAll('button')
    const sha256Button = buttons.find(b => b.text().toLowerCase().includes('sha-256'))
    expect(sha256Button).toBeTruthy()
    await sha256Button!.trigger('click')
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('11:22:33:44:55')
  })

  it('copy SHA-1 button calls clipboard.writeText', async () => {
    const w = mount(CertificateViewer, { props: { info: fixtureTLS() } })
    const buttons = w.findAll('button')
    const sha1Button = buttons.find(b => b.text().toLowerCase().includes('sha-1'))
    expect(sha1Button).toBeTruthy()
    await sha1Button!.trigger('click')
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('AA:BB:CC')
  })

  it('copy PEM button calls clipboard.writeText', async () => {
    const w = mount(CertificateViewer, { props: { info: fixtureTLS() } })
    const buttons = w.findAll('button')
    const pemButton = buttons.find(b => b.text().toLowerCase().includes('copy pem'))
    expect(pemButton).toBeTruthy()
    await pemButton!.trigger('click')
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      expect.stringContaining('BEGIN CERTIFICATE')
    )
  })

  it('renders warning banner when status=ok and validationSkipped=true', () => {
    const info: TLSInfo = {
      status: 'ok',
      targetHost: 'self-signed.example:443',
      connection: fixtureConnection(),
      certificates: [fixtureCert()],
      validationSkipped: true,
      originalError: 'x509: certificate signed by unknown authority',
    }
    const w = mount(CertificateViewer, { props: { info } })
    expect(w.text()).toContain('TLS certificate validation was skipped')
    expect(w.text()).toContain('allow_insecure_tls: true')
    expect(w.text()).toContain('x509: certificate signed by unknown authority')
  })

  it('does NOT render warning banner when validationSkipped is false', () => {
    const info: TLSInfo = {
      status: 'ok',
      targetHost: 'example.com:443',
      connection: fixtureConnection(),
      certificates: [fixtureCert()],
      validationSkipped: false,
    }
    const w = mount(CertificateViewer, { props: { info } })
    expect(w.text()).not.toContain('TLS certificate validation was skipped')
    expect(w.text()).not.toContain('allow_insecure_tls')
  })
})
