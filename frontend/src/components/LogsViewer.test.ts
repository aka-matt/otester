import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import LogsViewer from './LogsViewer.vue'
import { useLogsStore } from '../stores/logs'

const { clearLogs } = vi.hoisted(() => ({ clearLogs: vi.fn(() => Promise.resolve()) }))
vi.mock('../../wailsjs/go/app/App', () => ({ ClearLogs: clearLogs }))

vi.mock('naive-ui', () => ({
  NModal: { name: 'NModal', props: ['show'], template: '<div><slot /></div>' },
  NInput: { name: 'NInput', props: ['value', 'readonly'], template: '<textarea :readonly="readonly" :value="value"></textarea>' },
  NButton: { name: 'NButton', emits: ['click'], template: '<button @click="$emit(\'click\')"><slot /></button>' },
  NSpace: { name: 'NSpace', template: '<div><slot /></div>' },
  useMessage: () => ({ success: vi.fn(), error: vi.fn() }),
}))

describe('LogsViewer', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders log text read-only', () => {
    const store = useLogsStore()
    store.append({ time: '1', level: 'INFO', message: 'hello-log' })
    const wrapper = mount(LogsViewer, { props: { show: true } })
    const ta = wrapper.find('textarea')
    expect(ta.attributes('readonly')).toBeDefined()
    expect(ta.element.getAttribute('value')).toContain('hello-log')
  })

  it('Clear button calls backend and empties the store', async () => {
    const store = useLogsStore()
    store.append({ time: '1', level: 'INFO', message: 'x' })
    const wrapper = mount(LogsViewer, { props: { show: true } })
    await wrapper.findAll('button')[0].trigger('click')
    expect(clearLogs).toHaveBeenCalled()
    expect(store.entries.length).toBe(0)
  })
})