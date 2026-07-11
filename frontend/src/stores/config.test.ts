import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useConfigStore } from './config'
import type { ConfigView } from '../types'

const { loadConfig } = vi.hoisted(() => ({
  loadConfig: vi.fn<[], Promise<ConfigView | null>>(),
}))

vi.mock('../../wailsjs/go/app/App', () => ({
  LoadConfig: loadConfig,
}))

const fixture = (): ConfigView => ({
  app: {} as ConfigView['app'],
  variables: [
    { id: 'base_url', base_url: 'https://test.example', environment: 'test' },
    { id: 'base_url', base_url: 'https://prod.example', environment: 'production' },
  ], oauthProfiles: [], endpointGroups: [], endpoints: [], configPath: '/tmp/config.json',
})

describe('config variable selection', () => {
  beforeEach(() => { setActivePinia(createPinia()); vi.clearAllMocks() })

  it('defaults to the first variable and substitutes it', () => {
    const store = useConfigStore(); store.config = fixture(); store.ensureSelectedVariable()
    expect(store.selectedVariable?.environment).toBe('test')
    expect(store.substituteVariables('{{base_url}}/health')).toBe('https://test.example/health')
  })

  it('uses a user-selected variable for substitutions', () => {
    const store = useConfigStore(); store.config = fixture(); store.selectVariable(1)
    expect(store.selectedVariable?.environment).toBe('production')
    expect(store.substituteVariables('{{base_url}}/health')).toBe('https://prod.example/health')
  })

  it('falls back after reload removes the selection', () => {
    const store = useConfigStore(); store.config = fixture(); store.selectVariable(1)
    store.config = { ...fixture(), variables: [fixture().variables[0]] }; store.ensureSelectedVariable()
    expect(store.selectedVariableIndex).toBe(0)
  })

  it('retains the selected variable when reload reorders variables', async () => {
    const store = useConfigStore(); store.config = fixture(); store.selectVariable(1)
    loadConfig.mockResolvedValue({ ...fixture(), variables: [fixture().variables[1], fixture().variables[0]] })

    await store.loadConfig()

    expect(loadConfig).toHaveBeenCalledOnce()
    expect(store.selectedVariableIndex).toBe(0)
    expect(store.selectedVariable?.environment).toBe('production')
  })

  it('substitutes a selected variable literally', () => {
    const store = useConfigStore()
    store.config = {
      ...fixture(),
      variables: [{ id: 'api.url', base_url: 'https://test.example/$&/api', environment: 'test' }],
    }
    store.selectVariable(0)

    expect(store.substituteVariables('{{api.url}} {{apiXurl}}')).toBe('https://test.example/$&/api {{apiXurl}}')
  })

  it('treats {{base_url}} as an alias for the selected variable', () => {
    // Real configs (e.g. the bundled sample) often pair a variable whose id is
    // "host" or "dev" with a URL like {{base_url}}/endpoint. The substitution
    // must resolve {{base_url}} to the currently selected variable's base_url.
    const store = useConfigStore()
    store.config = {
      ...fixture(),
      variables: [
        { id: 'host', base_url: 'https://dev.example', environment: 'dev' },
        { id: 'host', base_url: 'https://prod.example', environment: 'prod' },
      ],
    }
    store.selectVariable(0)
    expect(store.substituteVariables('{{base_url}}/users')).toBe('https://dev.example/users')

    store.selectVariable(1)
    expect(store.substituteVariables('{{base_url}}/users')).toBe('https://prod.example/users')
  })

  it('substitutes {{environment}} with the selected variable environment', () => {
    const store = useConfigStore()
    store.config = {
      ...fixture(),
      variables: [{ id: 'host', base_url: 'https://dev.example', environment: 'staging' }],
    }
    store.selectVariable(0)
    expect(store.substituteVariables('{{base_url}}/endpoint-{{environment}}')).toBe('https://dev.example/endpoint-staging')
  })
})
