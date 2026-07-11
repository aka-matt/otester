import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useConfigStore } from './config'
import type { ConfigView } from '../types'

const fixture = (): ConfigView => ({
  app: {} as ConfigView['app'],
  variables: [
    { id: 'base_url', base_url: 'https://test.example', environment: 'test' },
    { id: 'base_url', base_url: 'https://prod.example', environment: 'production' },
  ], oauthProfiles: [], endpointGroups: [], endpoints: [], configPath: '/tmp/config.json',
})

describe('config variable selection', () => {
  beforeEach(() => { setActivePinia(createPinia()) })

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
})
