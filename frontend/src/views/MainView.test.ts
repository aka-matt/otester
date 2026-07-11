import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import MainView from './MainView.vue'
import { useConfigStore } from '../stores/config'
import { useResponseStore } from '../stores/response'
import type { ConfigView } from '../types'

const message = { error: vi.fn(), success: vi.fn(), info: vi.fn() }

vi.mock('naive-ui', () => ({
  NDropdown: {
    name: 'NDropdown',
    props: ['options'],
    emits: ['select'],
    template: '<div><slot /></div>',
  },
  useMessage: () => message,
}))

const config = (path: string, variableID = 'base_url'): ConfigView => ({
  app: {} as ConfigView['app'],
  variables: [{ id: variableID, base_url: 'https://example.test', environment: 'test' }],
  oauthProfiles: [],
  endpointGroups: [],
  endpoints: [],
  configPath: path,
})

describe('MainView config file menu', () => {
  const openConfigFile = vi.fn<[], Promise<ConfigView | null>>()

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    window.go = { app: { OpenConfigFile: openConfigFile } } as unknown as Window['go']
  })

  async function selectFileMenuItem(key: string) {
    const wrapper = mount(MainView, {
      global: {
        stubs: {
          EndpointList: true,
          RequestEditor: true,
          ResponseViewer: true,
          ThemeSwitcher: true,
          StatusBar: true,
        },
      },
    })
    const dropdown = wrapper.findAllComponents({ name: 'NDropdown' })[0]
    dropdown.vm.$emit('select', key)
    await nextTick()
    await Promise.resolve()
    return wrapper
  }

  it('shows Open Config File and replaces the config after a selected file loads', async () => {
    const configStore = useConfigStore()
    configStore.config = config('/current.json')
    configStore.ensureSelectedVariable()
    const responseStore = useResponseStore()
    responseStore.setResponse({
      requestId: 'request', statusCode: 200, status: 'OK', headers: {}, body: '',
      bodyTruncated: false, durationMs: 1, sizeBytes: 0, contentType: '', usedOAuth: false,
      tokenFromCache: false, errorCode: '', errorMessage: '',
    })
    openConfigFile.mockResolvedValue(config('/selected.json', 'replacement'))

    const wrapper = await selectFileMenuItem('open-config-file')

    expect(wrapper.findAllComponents({ name: 'NDropdown' })[0].props('options')).toContainEqual({
      label: 'Open Config File', key: 'open-config-file',
    })
    expect(openConfigFile).toHaveBeenCalledOnce()
    expect(configStore.config?.configPath).toBe('/selected.json')
    expect(configStore.selectedVariable?.id).toBe('replacement')
    expect(responseStore.response).toBeNull()
  })

  it('leaves the config unchanged when selecting a file is cancelled', async () => {
    const configStore = useConfigStore()
    configStore.config = config('/current.json')
    openConfigFile.mockResolvedValue(null)

    await selectFileMenuItem('open-config-file')

    expect(configStore.config?.configPath).toBe('/current.json')
  })

  it('shows an error when opening a config file fails', async () => {
    openConfigFile.mockRejectedValue(new Error('invalid config'))

    await selectFileMenuItem('open-config-file')

    expect(message.error).toHaveBeenCalledWith('Failed to open config file: Error: invalid config')
  })
})
