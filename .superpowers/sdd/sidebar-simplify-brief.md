### Task 1: Simplify endpoint sidebar

**Files:**
- Modify: `frontend/src/components/EndpointList.vue`
- Modify: `frontend/src/components/EndpointList.test.ts`

- [ ] Write a failing component test asserting no input with placeholder `Search endpoints...`, no select containing `All Methods`, a separator element with `data-testid="endpoint-divider"`, and all fixture endpoints rendered regardless of method.
- [ ] Run `npm test -- src/components/EndpointList.test.ts`; expected failure because old controls remain and no divider exists.
- [ ] Remove `searchQuery`, `methodFilter`, `filteredEndpoints`, search/filter markup/styles and empty-search copy. Replace all filtered endpoint/group uses with a computed `endpoints` list from config. Add `<hr data-testid="endpoint-divider" class="endpoint-divider">` after `.variable-selector`; render a no-endpoints message only when this full list is empty.
- [ ] Run `npm test -- src/components/EndpointList.test.ts && npm run build`; expected PASS.
- [ ] Commit with `git add frontend/src/components/EndpointList.vue frontend/src/components/EndpointList.test.ts && git commit -m "feat: simplify endpoint sidebar"`.

## Plan Self-Review

- The task covers removal, separator insertion, direct endpoint rendering, retained card behavior and regression tests.
- No incomplete requirements or inconsistent interface names remain.
