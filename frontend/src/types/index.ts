export interface KeyValue {
  key: string
  value: string
  enabled: boolean
}

export interface AppInfo {
  version: string
  name: string
}

export interface RequestInput {
  requestId: string
  endpointId: string
  method: string
  url: string
  headers: KeyValue[]
  queryParams: KeyValue[]
  bodyType: 'none' | 'json' | 'text' | 'x-www-form-urlencoded' | 'raw'
  body: string
  timeoutSeconds: number
  useOAuth: boolean
  oauthProfileId: string
}

export interface ResponseOutput {
  requestId: string
  statusCode: number
  status: string
  headers: Record<string, string[]>
  body: string
  bodyTruncated: boolean
  durationMs: number
  sizeBytes: number
  contentType: string
  usedOAuth: boolean
  tokenFromCache: boolean
  errorCode: string
  errorMessage: string
  tls?: TLSInfo
}

export interface AppConfig {
  name: string
  title: string
  defaultTimeoutSeconds: number
  maxResponseBodyBytes: number
  allowInsecureTLS: boolean
  persistRequestHistory: boolean
}

export interface Variable {
  id: string
  base_url: string
  environment: string
}

export interface OAuthProfileView {
  id: string
  name: string
  type: string
  org_id_uuid: string
  client_id: string
  clientSecretMasked: string
  scope: string
  token_url: string
  refreshBeforeExpirySeconds: number
}

export interface EndpointGroup {
  id: string
  name: string
}

export interface AuthConfig {
  type: 'none' | 'oauth2'
  profileId: string
  allowAuthorizationHeaderOverride: boolean
}

export interface BodyConfig {
  type: 'none' | 'json' | 'text' | 'x-www-form-urlencoded' | 'raw'
  content: string
}

export interface EndpointView {
  id: string
  name: string
  description: string
  groupId: string
  enabled: boolean
  method: string
  url: string
  timeoutSeconds: number
  auth: AuthConfig
  headers: KeyValue[]
  queryParams: KeyValue[]
  body: BodyConfig
  hasOAuth: boolean
}

export interface ConfigView {
  app: AppConfig
  variables: Variable[]
  oauthProfiles: OAuthProfileView[]
  endpointGroups: EndpointGroup[]
  endpoints: EndpointView[]
  configPath: string
}

export interface ValidationResult {
  valid: boolean
  errors: string[]
  warnings: string[]
}

export interface TokenStatus {
  profileId: string
  hasToken: boolean
  expiresAt: string | null
  fromCache: boolean
}

export interface TLSConnectionView {
  version: string
  cipherSuite: string
  cipherSuiteName: string
  negotiatedProtocol?: string
  serverName: string
  resumed: boolean
  scts?: string[]
  ocspStapled: boolean
  peerCertificates: number
}

export interface CertificateView {
  position: string
  subject: string
  issuer: string
  serialNumber: string
  version: number
  signatureAlgorithm: string
  notBefore: string
  notAfter: string
  isExpired: boolean
  isNotYetValid: boolean
  daysToExpiry: number
  subjectKeyId: string
  authorityKeyId: string
  sans: string[]
  keyAlgorithm: string
  keySize: number
  publicKeyPem: string
  fingerprintSha1: string
  fingerprintSha256: string
  isCa: boolean
  maxPathLength: number
  keyUsage: string[]
  extendedKeyUsage: string[]
  crlDistributionPoints?: string[]
  policies?: string[]
  rawDer: string
  pem: string
  signatureBytes: string
}

export interface TLSInfo {
  status: string
  error?: string
  targetHost: string
  attemptedServerName?: string
  connection?: TLSConnectionView
  certificates: CertificateView[]
  validationSkipped?: boolean
  originalError?: string
}

export type LogEntry = {
  time: string
  level: string
  message: string
}
