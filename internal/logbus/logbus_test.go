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
