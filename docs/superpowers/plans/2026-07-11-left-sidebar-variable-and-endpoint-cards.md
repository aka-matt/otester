# Left Sidebar Variable and Endpoint Cards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users select one configuration variable in the left sidebar and browse endpoints as color-coded endpoint-group cards.

**Architecture:** The Pinia config store keeps the selected variable as an array index so environment entries may reuse a variable ID. The sidebar resolves selected endpoint URLs before loading the request store. It owns filtering and stable group-color presentation only.

**Tech Stack:** Vue 3, TypeScript, Pinia, Vite, Vitest, Vue Test Utils, Go test.

## Global Constraints

- Do not alter `config.json`, the Go backend API, or write configuration from the UI.
- Default to the first variable; retain an existing selection after reload when the same `id` and `environment` exists, otherwise use the first variable.
- Replace only `{{<selected variable id>}}` with the selected entry's `base_url`; without selection leave text unchanged.
- Switching variables immediately reloads a selected endpoint URL.
- Group accents are stable by configured group order; ungrouped and unknown groups use neutral styling.
- Preserve search, method filter, collapsing, disabled endpoints, OAuth indicator, and selected-state contrast.

---

## File Structure

- `frontend/package.json`: Add test script/dependencies.
- `frontend/vite.config.ts`: Configure Vitest jsdom.
- `frontend/src/test/setup.ts`: Test cleanup.
- `frontend/src/stores/config.ts`: Selected-variable state and substitution.
- `frontend/src/stores/config.test.ts`: Store tests.
- `frontend/src/components/EndpointList.vue`: Variable buttons and card UI.
- `frontend/src/components/EndpointList.test.ts`: Sidebar tests.

### Task 1: Establish frontend test harness and selected-variable store

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/vite.config.ts`
- Create: `frontend/src/test/setup.ts`
- Modify: `frontend/src/stores/config.ts`
- Create: `frontend/src/stores/config.test.ts`

**Interfaces:**
- Produces `selectedVariableIndex: Ref<number | null>`, `selectedVariable: ComputedRef<Variable | null>`, `ensureSelectedVariable(): void`, and `selectVariable(index: number): void` from `useConfigStore()`.
- Produces `substituteVariables(text: string): string` using only the selected variable.

- [ ] **Step 1: Add test configuration**

Update `frontend/package.json`:

```json
"scripts": { "test": "vitest run" },
"devDependencies": {
  "@vue/test-utils": "^2.4.6",
  "jsdom": "^24.1.1",
  "vitest": "^1.6.0"
}
```

Keep existing scripts/dependencies. Add this to the Vite config:

```ts
test: {
  environment: 'jsdom',
  setupFiles: ['./src/test/setup.ts'],
  globals: true,
},
```

Create `frontend/src/test/setup.ts`:

```ts
import { afterEach } from 'vitest'

afterEach(() => { document.body.innerHTML = '' })
```

- [ ] **Step 2: Install and verify the harness**

Run: `npm install`

Expected: dependencies install and `frontend/package-lock.json` updates.

Run: `npm test -- --passWithNoTests`

Expected: PASS with no test files found.

- [ ] **Step 3: Write failing config-store tests**

Create `frontend/src/stores/config.test.ts`:

```ts
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
  beforeEach(() => setActivePinia(createPinia()))
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
```

- [ ] **Step 4: Verify RED**

Run: `npm test -- src/stores/config.test.ts`

Expected: FAIL because the new store members are not exposed.

- [ ] **Step 5: Implement minimal store support**

In `frontend/src/stores/config.ts` add:

```ts
const selectedVariableIndex = ref<number | null>(null)
const selectedVariable = computed(() => selectedVariableIndex.value === null
  ? null : config.value?.variables[selectedVariableIndex.value] ?? null)
function ensureSelectedVariable() {
  const variables = config.value?.variables ?? []
  const current = selectedVariable.value
  if (current && variables.some(v => v.id === current.id && v.environment === current.environment)) return
  selectedVariableIndex.value = variables.length ? 0 : null
}
function selectVariable(index: number) {
  if (config.value?.variables[index]) selectedVariableIndex.value = index
}
function substituteVariables(text: string) {
  const variable = selectedVariable.value
  return variable ? text.replace(new RegExp(`{{${variable.id}}}`, 'g'), variable.base_url) : text
}
```

Call `ensureSelectedVariable()` after loading config and return all four new members. Remove the existing loop-based substitution implementation.

- [ ] **Step 6: Verify GREEN**

Run: `npm test -- src/stores/config.test.ts && npm run build`

Expected: three store tests PASS and the production build PASS.

- [ ] **Step 7: Commit Task 1**

```bash
git add frontend/package.json frontend/package-lock.json frontend/vite.config.ts frontend/src/test/setup.ts frontend/src/stores/config.ts frontend/src/stores/config.test.ts
git commit -m "feat: add selectable config variable state"
```

### Task 2: Render variable buttons and color-coded group cards

**Files:**
- Modify: `frontend/src/components/EndpointList.vue`
- Create: `frontend/src/components/EndpointList.test.ts`

**Interfaces:**
- Consumes the Task 1 config-store API plus `requestStore.loadFromEndpoint(endpoint)`.
- Produces `[data-testid="variable-option"]`, `[data-testid="endpoint-group-card"]`, and `[data-testid="ungrouped-card"]` sidebar elements.

- [ ] **Step 1: Write failing component tests**

Create `frontend/src/components/EndpointList.test.ts`, mounting with an active Pinia and a fixture containing two variables, `Users`/`Orders` groups, and an endpoint referencing an unknown group. Stub `responseStore.clear`. Implement:

```ts
it('selects a variable and reloads the active endpoint URL', async () => {
  configStore.config = fixtureConfig; configStore.selectEndpoint('users')
  await wrapper.findAll('[data-testid="variable-option"]')[1].trigger('click')
  expect(configStore.selectedVariable?.environment).toBe('production')
  expect(requestStore.url).toBe('https://prod.example/users')
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
```

- [ ] **Step 2: Verify RED**

Run: `npm test -- src/components/EndpointList.test.ts`

Expected: FAIL because the controls/cards and their test IDs do not exist.

- [ ] **Step 3: Implement the sidebar**

In `EndpointList.vue`:

```ts
const groupPalette = ['#3b82f6', '#10b981', '#8b5cf6', '#f59e0b', '#ec4899', '#06b6d4']
const knownGroupIds = computed(() => new Set((configStore.config?.endpointGroups ?? []).map(g => g.id)))
const ungroupedEndpoints = computed(() => filteredEndpoints.value.filter(ep => !knownGroupIds.value.has(ep.groupId)))
function groupAccent(groupId: string) {
  const index = (configStore.config?.endpointGroups ?? []).findIndex(g => g.id === groupId)
  return groupPalette[index % groupPalette.length]
}
function loadEndpoint(endpoint: EndpointView) {
  requestStore.loadFromEndpoint({ ...endpoint, url: configStore.substituteVariables(endpoint.url) })
}
function selectVariable(index: number) {
  configStore.selectVariable(index)
  if (configStore.selectedEndpoint) loadEndpoint(configStore.selectedEndpoint)
}
```

Render a `variable-selector` above search, with button label `variable.environment || variable.id` plus the ID, `aria-pressed`, selected class, and `data-testid="variable-option"`. Update existing endpoint selection to call `loadEndpoint`. Wrap each group in a card with `:style="{ '--group-accent': groupAccent(group.id) }"` and `data-testid="endpoint-group-card"`; use neutral `--group-accent` and `data-testid="ungrouped-card"` for unknown/no group IDs. Add an empty-search message. Style buttons/cards with keyboard focus, accent border/title, hover, disabled opacity, and unchanged high-contrast selected endpoint styling.

- [ ] **Step 4: Verify GREEN**

Run: `npm test -- src/components/EndpointList.test.ts`

Expected: all three component tests PASS.

- [ ] **Step 5: Run complete verification**

Run: `npm test && npm run build && go test ./...`

Expected: all frontend tests, frontend production build, and Go packages PASS.

- [ ] **Step 6: Commit Task 2**

```bash
git add frontend/src/components/EndpointList.vue frontend/src/components/EndpointList.test.ts
git commit -m "feat: add variable selector and endpoint group cards"
```

## Plan Self-Review

- Coverage: Task 1 implements default/retained selection, fallback, and selected-variable replacement. Task 2 implements variable buttons, immediate URL refresh, stable group cards/colors, ungrouped handling, empty state, and preserved list behavior.
- Placeholder scan: no incomplete or unspecified implementation action remains.
- Type consistency: later tasks use exactly the Task 1 store API.
