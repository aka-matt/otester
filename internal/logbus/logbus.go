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

func (b *Bus) SetEmit(fn func(Entry))    { b.mu.Lock(); b.emit = fn; b.mu.Unlock() }
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
