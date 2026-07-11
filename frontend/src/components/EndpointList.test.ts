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
  let pinia: ReturnType<typeof createPinia>

  function mountWithConfig(config: ConfigView) {
    configStore.config = config
    configStore.ensureSelectedVariable()
    wrapper = mount(EndpointList, { global: { plugins: [pinia] } })
  }

  beforeEach(() => {
    pinia = createPinia()
    setActivePinia(pinia)
    configStore = useConfigStore()
    requestStore = useRequestStore()
    vi.spyOn(useResponseStore(), 'clear')
  })

  it('selects a variable and lets the URL bar recompute the substituted URL', async () => {
    mountWithConfig(fixtureConfig)
    configStore.selectEndpoint('users')
    await wrapper.findAll('.endpoint-item')[0].trigger('click')

    await wrapper.findAll('[data-testid="variable-option"]')[1].trigger('click')

    expect(configStore.selectedVariable?.environment).toBe('production')
    // The request store keeps the raw template so any later variable change
    // propagates through the URL bar's computed without reloading the endpoint.
    expect(requestStore.url).toBe('{{api}}/users')
    expect(configStore.substituteVariables(requestStore.url)).toBe('https://prod.example/users')
  })

  it('stores the raw template on click so the URL bar can show the substituted URL', async () => {
    // Regression: clicking an endpoint used to write the substituted URL into
    // requestStore.url, which meant Send could send "{{base_url}}/..." if no
    // variable was selected. The URL bar now always displays the substituted
    // form, while requestStore.url keeps the template.
    const somethingEndpoint = endpoint({
      id: 'something',
      name: 'Get something',
      url: '{{base_url}}/api/v2/something/one',
    })
    configStore.config = {
      ...fixtureConfig,
      variables: [{ id: 'base_url', base_url: 'https://sample.co', environment: 'test' }],
      endpoints: [somethingEndpoint],
    }
    configStore.ensureSelectedVariable()

    await wrapper.findAll('.endpoint-item')[0].trigger('click')

    expect(requestStore.url).toBe('{{base_url}}/api/v2/something/one')
    expect(configStore.substituteVariables(requestStore.url)).toBe('https://sample.co/api/v2/something/one')
  })

  it('regenerates the URL bar when a different variable is picked', async () => {
    configStore.config = {
      ...fixtureConfig,
      variables: [
        { id: 'base_url', base_url: 'https://sample.co', environment: 'test' },
        { id: 'base_url', base_url: 'https://staging.pokeapi.co', environment: 'staging' },
      ],
      endpoints: [
        endpoint({ id: 'something', name: 'Get something', url: '{{base_url}}/api/v2/something/one' }),
      ],
    }
    configStore.ensureSelectedVariable()

    await wrapper.findAll('.endpoint-item')[0].trigger('click')
    expect(configStore.substituteVariables(requestStore.url)).toBe('https://sample.co/api/v2/something/one')

    await wrapper.findAll('[data-testid="variable-option"]')[1].trigger('click')

    expect(configStore.selectedVariable?.environment).toBe('staging')
    expect(configStore.substituteVariables(requestStore.url)).toBe('https://staging.pokeapi.co/api/v2/something/one')
  })

  it('substitutes {{base_url}} even when the variable id is different', async () => {
    // Real-world configs (including the sample) name the variable id something
    // like "host" or "dev" but use {{base_url}} in endpoint URLs. The
    // substitution helper must handle that as an alias for the selected
    // variable's base_url field.
    configStore.config = {
      ...fixtureConfig,
      variables: [{ id: 'host', base_url: 'https://sample.co', environment: 'test' }],
      endpoints: [
        endpoint({ id: 'something', name: 'Get something', url: '{{base_url}}/api/v2/something/one' }),
      ],
    }
    configStore.ensureSelectedVariable()

    await wrapper.findAll('.endpoint-item')[0].trigger('click')

    expect(configStore.substituteVariables(requestStore.url)).toBe('https://sample.co/api/v2/something/one')
  })

  it('auto-selects the first variable when an endpoint is clicked without one selected', async () => {
    // If the user somehow never had a variable selected (e.g. config reloaded
    // to a state where selectedVariableIndex ended up null), clicking an
    // endpoint must still produce a substituted URL — not the raw template.
    configStore.config = {
      ...fixtureConfig,
      variables: [{ id: 'base_url', base_url: 'https://sample.co', environment: 'test' }],
      endpoints: [
        endpoint({ id: 'something', name: 'Get something', url: '{{base_url}}/api/v2/something/one' }),
      ],
    }
    configStore.selectedVariableIndex = null

    await wrapper.findAll('.endpoint-item')[0].trigger('click')

    expect(configStore.selectedVariableIndex).toBe(0)
    expect(configStore.substituteVariables(requestStore.url)).toBe('https://sample.co/api/v2/something/one')
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
