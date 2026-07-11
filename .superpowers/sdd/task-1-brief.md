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

