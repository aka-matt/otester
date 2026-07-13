import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useLogsStore } from './logs'

describe('logs store', () => {
  beforeEach(() => { setActivePinia(createPinia()) })

  it('primes, appends, and joins text', () => {
    const s = useLogsStore()
    s.prime([{ time: '1', level: 'INFO', message: 'a' }])
    s.append({ time: '2', level: 'WARN', message: 'b' })
    expect(s.entries.length).toBe(2)
    expect(s.text).toContain('INFO')
    expect(s.text).toContain('a')
    expect(s.text).toContain('b')
  })

  it('clears', () => {
    const s = useLogsStore()
    s.append({ time: '1', level: 'INFO', message: 'a' })
    s.clear()
    expect(s.entries.length).toBe(0)
  })

  it('caps at 2000 entries', () => {
    const s = useLogsStore()
    for (let i = 0; i < 2100; i++) s.append({ time: String(i), level: 'INFO', message: 'm' })
    expect(s.entries.length).toBe(2000)
    expect(s.entries[0].time).toBe('100')
  })
})
