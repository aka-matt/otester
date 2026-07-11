# Task 1 Report: Frontend test harness and selected-variable store

## Implementation

- Added a Vitest script plus `@vue/test-utils`, `jsdom`, and `vitest` development dependencies; refreshed `frontend/package-lock.json` with `npm install`.
- Configured Vitest for jsdom with a DOM-reset setup file.
- Added selected-variable index/state, selection validation, selection API, and selected-variable-only substitution to the config Pinia store. `loadConfig` now validates the selection after loading.
- Added three Pinia store tests covering default selection/substitution, explicit selection/substitution, and fallback when a reload removes the selection.

## Files

- Modified `frontend/package.json`
- Modified `frontend/package-lock.json`
- Modified `frontend/vite.config.ts`
- Created `frontend/src/test/setup.ts`
- Modified `frontend/src/stores/config.ts`
- Created `frontend/src/stores/config.test.ts`

## RED/GREEN evidence

- RED: `npm test -- src/stores/config.test.ts` failed all three tests because `ensureSelectedVariable` and `selectVariable` were not exposed by the existing store.
- GREEN: after the minimal store implementation, `npm test -- src/stores/config.test.ts` passed all 3 tests.
- A first production build revealed a Vitest TypeScript error caused by returning the Pinia instance from the single-expression `beforeEach` callback. The callback was changed to block form, returning `void`; this preserved test behavior and resolved the type error.

## Verification

- `npm test -- --passWithNoTests` — passed (harness verification before tests existed).
- `npm test -- src/stores/config.test.ts && npm run build && npm test` — passed: 3/3 focused tests, production build, and 3/3 full-suite tests.
- `git diff --check -- [Task 1 files]` — passed.

## Self-review

- Checked each required store member is returned from `useConfigStore`.
- Confirmed substitution now operates only on `selectedVariable` and leaves input unchanged with no selection.
- Confirmed config loading invokes `ensureSelectedVariable` after assignment.
- Confirmed only the six requested Task 1 files will be staged and committed.
- Independent review found no Critical, Important, or Minor issues; it also confirmed manifest/lockfile consistency and the required test coverage.

## Concerns

- `npm install` reports 4 dependency audit findings (2 moderate, 1 high, 1 critical); no remediation was applied because it is outside the requested Task 1 dependency additions.
- Workspace-wide `git diff --check` reports a pre-existing trailing whitespace issue in unrelated `otester-PRD.md`; the Task 1 file-specific check is clean.

## Review follow-up

- The selection reconciliation now saves the pre-load selected variable identity, replaces the config, and resolves that identity to its new index. If it is absent, it falls back to the first variable or no selection.
- Substitution now escapes variable IDs before constructing the exact placeholder regular expression and uses a callback replacement, preserving literal `$&` sequences in base URLs.
- RED: the added reorder and literal-substitution tests failed against the initial implementation: reordering retained index `1` instead of production at index `0`, and `api.url` incorrectly matched `apiXurl` while `$&` was expanded.
- GREEN: `npm test -- src/stores/config.test.ts` passed all 5 tests and `npm run build` passed after the scoped store fix.
