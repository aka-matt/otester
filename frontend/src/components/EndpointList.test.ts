import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, type VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import EndpointList from './EndpointList.vue'
import { useConfigStore } from '../stores/config'
import { useRequestStore } from '../stores/request'
import { useResponseStore } from '../stores/response'
import type { ConfigView, EndpointView } from '../types'

const endpoint = (overrides: Partial<EndpointView>): EndpointView => ({
  id: 'endpoint',
  name: 'Endpoint',
  description: '',
  groupId: '',
  enabled: true,
  method: 'GET',
  url: 'https://example.test',
  timeoutSeconds: 30,
  auth: { type: 'none', profileId: '', allowAuthorizationHeaderOverride: false },
  headers: [],
  queryParams: [],
  body: { type: 'none', content: '' },
  hasOAuth: false,
  ...overrides,
})

const fixtureConfig: ConfigView = {
  app: {
    name: 'test', title: 'Test', defaultTimeoutSeconds: 30,
    maxResponseBodyBytes: 1024, allowInsecureTLS: false, persistRequestHistory: false,
  },
  variables: [
    { id: 'api', base_url: 'https://staging.example', environment: 'staging' },
    { id: 'api', base_url: 'https://prod.example', environment: 'production' },
    { id: 'fallback', base_url: 'https://fallback.example', environment: '' },
  ],
  oauthProfiles: [],
  endpointGroups: [
    { id: 'users', name: 'Users' },
    { id: 'orders', name: 'Orders' },
  ],
  endpoints: [
    endpoint({ id: 'users', name: 'Users endpoint', groupId: 'users', url: '{{api}}/users' }),
    endpoint({ id: 'orders', name: 'Orders endpoint', groupId: 'orders' }),
    endpoint({ id: 'orphan', name: 'Orphan endpoint', groupId: 'missing', hasOAuth: true }),
  ],
  configPath: '/tmp/config.json',
}

describe('EndpointList', () => {
  let wrapper: VueWrapper
  let configStore: ReturnType<typeof useConfigStore>
  let requestStore: ReturnType<typeof useRequestStore>

  beforeEach(() => {
    const pinia = createPinia()
    setActivePinia(pinia)
    configStore = useConfigStore()
    requestStore = useRequestStore()
    configStore.config = fixtureConfig
    configStore.ensureSelectedVariable()
    vi.spyOn(useResponseStore(), 'clear')
    wrapper = mount(EndpointList, { global: { plugins: [pinia] } })
  })

  it('selects a variable and reloads the active endpoint URL', async () => {
    configStore.selectEndpoint('users')

    await wrapper.findAll('[data-testid="variable-option"]')[1].trigger('click')

    expect(configStore.selectedVariable?.environment).toBe('production')
    expect(requestStore.url).toBe('https://prod.example/users')
  })

  it('renders an empty-state message when no variables are configured', async () => {
    configStore.config = { ...fixtureConfig, variables: [] }
    await nextTick()

    expect(wrapper.text()).toContain('No configuration variables available.')
  })

  it('uses only the variable ID when its environment is absent', () => {
    expect(wrapper.findAll('[data-testid="variable-option"]')[2].text()).toBe('fallback')
  })

  it('renders every endpoint without search or method filtering controls', () => {
    expect(wrapper.find('input[placeholder="Search endpoints..."]').exists()).toBe(false)
    expect(wrapper.findAll('select').some(select => select.text().includes('All Methods'))).toBe(false)
    expect(wrapper.find('[data-testid="endpoint-divider"]').element.tagName).toBe('HR')

    for (const fixtureEndpoint of fixtureConfig.endpoints) {
      expect(wrapper.text()).toContain(fixtureEndpoint.name)
    }
  })

  it('renders configured groups as differently accented endpoint cards', () => {
    const cards = wrapper.findAll('[data-testid="endpoint-group-card"]')
    expect(cards).toHaveLength(2)
    expect(cards[0].attributes('style')).not.toBe(cards[1].attributes('style'))
    expect(cards[0].text()).toContain('Users')
  })

  it('puts unknown-group endpoints in the ungrouped card', () => {
    expect(wrapper.get('[data-testid="ungrouped-card"]').text()).toContain('Orphan endpoint')
  })

  it('renders OAuth endpoint locks in the ungrouped card', () => {
    expect(wrapper.get('[data-testid="ungrouped-card"]').text()).toContain('🔒')
  })
})
