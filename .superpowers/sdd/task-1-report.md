# Task 1 Report — Backend log bus (`internal/logbus`)

## What was built

A new, self-contained Go package `internal/logbus` providing a thread-safe
in-memory ring-buffer logger used by later tasks to record HTTP and OAuth
activity for the in-app log window.

### Files created

- `/mnt/c/dev/GitHub/otester/internal/logbus/logbus.go` — implementation.
- `/mnt/c/dev/GitHub/otester/internal/logbus/logbus_test.go` — tests (verbatim from brief).

### Public surface (verbatim from brief)

- `type Level string` + consts `LevelDebug`/`LevelInfo`/`LevelWarn`/`LevelError`.
- `type Entry struct { Time, Level, Message string }` with JSON tags.
- `type Logger interface { Log, Debugf, Infof, Warnf, Errorf }`.
- `func New(capacity int) *Bus` — `*Bus` satisfies `Logger`.
- `func Nop() Logger`.
- `(*Bus).SetEmit(func(Entry))`, `Snapshot() []Entry`, `Clear()`,
  `SetClock(fn func() string)` (test seam).

### Behaviour

- `Log` formats with `fmt.Sprintf`, stamps `e.Time` while holding the mutex,
  then appends to a fixed-capacity slice. Overflow drops oldest entries
  via slice re-slice. `SetEmit` callback fires **after** the mutex is
  released to keep call sites deadlock-free and avoid holding the lock
  during arbitrary frontend-bridge calls.
- `Snapshot` returns a defensive copy; `Clear` reuses the slice header.
- `Nop()` returns a zero-cost `nop{}` value; a logger with nil emit hook
  is safe.

## TDD evidence

### RED — initial test run

Command: `go test ./internal/logbus/`

```
# otester/internal/logbus [otester/internal/logbus.test]
internal/logbus/logbus_test.go:6:7: undefined: New
internal/logbus/logbus_test.go:25:7: undefined: New
internal/logbus/logbus_test.go:35:7: undefined: New
internal/logbus/logbus_test.go:44:7: undefined: New
internal/logbus/logbus_test.go:45:16: undefined: Entry
internal/logbus/logbus_test.go:46:19: undefined: Entry
internal/logbus/logbus_test.go:54:7: undefined: Nop
internal/logbus/logbus_test.go:56:8: undefined: LevelError
internal/logbus/logbus_test.go:60:7: undefined: New
FAIL	otester/internal/logbus [build failed]
FAIL
```

Build fails for the exact reason the brief predicted
(`undefined: New` / `Entry` / `Nop` / `LevelError`).

### GREEN — after implementation

Command: `go test ./internal/logbus/ -v`

```
=== RUN   TestRingBufferDropsOldestBeyondCapacity
--- PASS: TestRingBufferDropsOldestBeyondCapacity (0.00s)
=== RUN   TestSnapshotReturnsCopy
--- PASS: TestSnapshotReturnsCopy (0.00s)
=== RUN   TestClearEmptiesBuffer
--- PASS: TestClearEmptiesBuffer (0.00s)
=== RUN   TestEmitCalledPerEntry
--- PASS: TestEmitCalledPerEntry (0.00s)
=== RUN   TestNopLoggerIsSafe
--- PASS: TestNopLoggerIsSafe (0.00s)
=== RUN   TestEmitNilIsNoop
--- PASS: TestEmitNilIsNoop (0.00s)
PASS
ok  	otester/internal/logbus	0.004s
```

All 6 tests pass; output is pristine (no stray prints, no skipped tests).

### Build + vet

```
$ go build ./...             # silent — clean
$ go vet ./internal/logbus/  # silent — clean
```

## Commit

```
dc70034 feat(logbus): in-memory ring-buffer logger with emit hook
 2 files changed, 161 insertions(+)
```

Only the two new files were staged; the pre-existing modifications in
`internal/config/loader.go` and `internal/config/loader_test.go` remain
untracked and untouched, per the line-ending-churn notice.

## Self-review findings

- **Completeness** — Every symbol and behaviour listed in the brief's
  `Produces` block is implemented; brief tests pass without modification.
- **Quality** — Implementation mirrors the brief verbatim, so there is
  no drift to justify. Mutex is held only for the in-memory mutation;
  the optional emit callback runs outside the lock so a slow or blocking
  frontend bridge cannot stall other loggers calling into the same bus.
- **YAGNI** — Nothing was added beyond the brief (no extra levels, no
  slog adapter, no file sink). Disk persistence is forbidden by the
  global constraints; this package is ring-only.
- **Pristine test output** — `go test -v` produces only `--- PASS` per
  test followed by a clean `PASS`/`ok` line; no test prints, no
  `t.Logf` leakage.
- **Secrets** — `Log` formats with `fmt.Sprintf` only on caller-supplied
  data; it never touches HTTP headers or tokens. Redaction of
  `client_secret` / `Authorization: Bearer …` is the responsibility of
  the call sites introduced by Tasks 2 and 3.

## Concerns

None. Implementation is verbatim from the brief, tests pass cleanly, and
the commit is scoped to exactly the two new files.
