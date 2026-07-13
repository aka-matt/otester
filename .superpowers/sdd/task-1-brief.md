### Task 1: Backend log bus (`internal/logbus`)

**Files:**
- Create: `internal/logbus/logbus.go`
- Test: `internal/logbus/logbus_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type Entry struct { Time string; Level string; Message string }` with JSON tags `time`,`level`,`message`.
  - `type Level string` with consts `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError` (values `"DEBUG"`,`"INFO"`,`"WARN"`,`"ERROR"`).
  - `type Logger interface { Log(Level, string, ...any); Debugf(string, ...any); Infof(string, ...any); Warnf(string, ...any); Errorf(string, ...any) }`
  - `func New(capacity int) *Bus` — `*Bus` implements `Logger`.
  - `func Nop() Logger` — no-op logger (safe when unused).
  - Methods on `*Bus`: `SetEmit(fn func(Entry))`, `Snapshot() []Entry`, `Clear()`, `SetClock(fn func() string)` (test seam for timestamps).
  - `Bus` also implements the four `…f` helpers by delegating to `Log`.

- [ ] **Step 1: Write the failing test**

```go
package logbus

import "testing"

func TestRingBufferDropsOldestBeyondCapacity(t *testing.T) {
	b := New(2)
	b.SetClock(func() string { return "T" })
	b.Infof("first")
	b.Infof("second")
	b.Infof("third")

	got := b.Snapshot()
	if len(got) != 2 {
		t.Fatalf("want 2 entries, got %d", len(got))
	}
	if got[0].Message != "second" || got[1].Message != "third" {
		t.Fatalf("want [second third], got [%s %s]", got[0].Message, got[1].Message)
	}
	if got[0].Level != "INFO" || got[0].Time != "T" {
		t.Fatalf("unexpected level/time: %+v", got[0])
	}
}

func TestSnapshotReturnsCopy(t *testing.T) {
	b := New(4)
	b.Infof("one")
	snap := b.Snapshot()
	snap[0].Message = "mutated"
	if b.Snapshot()[0].Message != "one" {
		t.Fatal("Snapshot must return a copy, not the backing slice")
	}
}

func TestClearEmptiesBuffer(t *testing.T) {
	b := New(4)
	b.Infof("one")
	b.Clear()
	if len(b.Snapshot()) != 0 {
		t.Fatal("Clear must empty the buffer")
	}
}

func TestEmitCalledPerEntry(t *testing.T) {
	b := New(4)
	var emitted []Entry
	b.SetEmit(func(e Entry) { emitted = append(emitted, e) })
	b.Warnf("hi %d", 7)
	if len(emitted) != 1 || emitted[0].Message != "hi 7" || emitted[0].Level != "WARN" {
		t.Fatalf("emit not called correctly: %+v", emitted)
	}
}

func TestNopLoggerIsSafe(t *testing.T) {
	l := Nop()
	l.Infof("no panic")
	l.Log(LevelError, "still no panic")
}

func TestEmitNilIsNoop(t *testing.T) {
	b := New(4) // no SetEmit
	b.Infof("buffered but not emitted") // must not panic
	if len(b.Snapshot()) != 1 {
		t.Fatal("entry should still buffer when emit is nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/logbus/`
Expected: FAIL — `undefined: New` / package has no test target yet.

- [ ] **Step 3: Write minimal implementation**

```go
package logbus

import (
	"fmt"
	"sync"
	"time"
)

type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

type Entry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type Logger interface {
	Log(level Level, format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type Bus struct {
	mu       sync.Mutex
	capacity int
	entries  []Entry
	emit     func(Entry)
	clock    func() string
}

func New(capacity int) *Bus {
	if capacity < 1 {
		capacity = 1
	}
	return &Bus{
		capacity: capacity,
		entries:  make([]Entry, 0, capacity),
		clock:    func() string { return time.Now().Format("15:04:05.000") },
	}
}

func (b *Bus) SetEmit(fn func(Entry)) { b.mu.Lock(); b.emit = fn; b.mu.Unlock() }
func (b *Bus) SetClock(fn func() string) { b.mu.Lock(); b.clock = fn; b.mu.Unlock() }

func (b *Bus) Log(level Level, format string, args ...any) {
	e := Entry{Level: string(level), Message: fmt.Sprintf(format, args...)}
	b.mu.Lock()
	e.Time = b.clock()
	b.entries = append(b.entries, e)
	if len(b.entries) > b.capacity {
		b.entries = b.entries[len(b.entries)-b.capacity:]
	}
	emit := b.emit
	b.mu.Unlock()
	if emit != nil {
		emit(e)
	}
}

func (b *Bus) Debugf(f string, a ...any) { b.Log(LevelDebug, f, a...) }
func (b *Bus) Infof(f string, a ...any)  { b.Log(LevelInfo, f, a...) }
func (b *Bus) Warnf(f string, a ...any)  { b.Log(LevelWarn, f, a...) }
func (b *Bus) Errorf(f string, a ...any) { b.Log(LevelError, f, a...) }

func (b *Bus) Snapshot() []Entry {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Entry, len(b.entries))
	copy(out, b.entries)
	return out
}

func (b *Bus) Clear() {
	b.mu.Lock()
	b.entries = b.entries[:0]
	b.mu.Unlock()
}

type nop struct{}

func Nop() Logger { return nop{} }

func (nop) Log(Level, string, ...any) {}
func (nop) Debugf(string, ...any)     {}
func (nop) Infof(string, ...any)      {}
func (nop) Warnf(string, ...any)      {}
func (nop) Errorf(string, ...any)     {}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/logbus/`
Expected: PASS (all tests).

- [ ] **Step 5: Commit**

```bash
git add internal/logbus/logbus.go internal/logbus/logbus_test.go
git commit -m "feat(logbus): in-memory ring-buffer logger with emit hook"
```

---

