# Simplify Endpoint Sidebar Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove endpoint search and HTTP-method filtering, separating variables from endpoint cards with a plain horizontal rule.

**Architecture:** `EndpointList.vue` will render configured endpoints directly, removing filter state and derived filtering. The existing card, group, selection and variable behavior remains unchanged.

**Tech Stack:** Vue 3, TypeScript, Vitest, Vue Test Utils.

## Global Constraints

- Remove both search and method-filter controls, their reactive state, and filtering logic.
- Insert a pure horizontal separator after Variables; do not add text.
- Preserve group cards, collapse, colors, selection, disabled state, OAuth indicator and empty endpoint state.

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
