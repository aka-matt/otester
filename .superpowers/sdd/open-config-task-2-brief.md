### Task 2: Frontend menu integration and generated bindings

**Files:**
- Modify: `frontend/src/views/MainView.vue`, `frontend/src/stores/config.ts`
- Regenerate: `frontend/wailsjs/go/app/App.js`, `frontend/wailsjs/go/app/App.d.ts`
- Create/Modify: `frontend/src/views/MainView.test.ts`

- [ ] Write a failing menu test that expects `Open Config File`, mocks `OpenConfigFile`, and verifies a returned config replaces `configStore.config`; test cancellation leaves it unchanged and errors show Naive UI error feedback.
- [ ] Run `npm test -- src/views/MainView.test.ts` and verify RED.
- [ ] Add `OpenConfigFile(): Promise<ConfigView | null>` to frontend Wails typing and regenerate bindings with the project Wails generator rather than hand-editing generated files.
- [ ] Rename menu option/key, call `window.go.app.OpenConfigFile()`, assign successful config through a store method that reconciles selected variables, and clear response; cancellation does nothing and catch reports error.
- [ ] Run `npm test && npm run build`, then `go test ./...`; report the known hard-coded-path failures separately if unchanged. Commit `feat: open config file from menu`.

## Plan Self-Review

- Task 1 covers atomic backend path replacement, cancellation, failure retention, and reload semantics.
- Task 2 covers the visible menu rename, frontend state update, user feedback, and regenerated bindings.
- All interfaces use `OpenConfigFile() (*config.ConfigView, error)` / `Promise<ConfigView | null>` consistently.
