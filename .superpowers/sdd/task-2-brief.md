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
