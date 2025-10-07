package generics

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMapReducePartition(t *testing.T) {
	t.Skip("TODO: remove this skip to attempt the Map/Reduce/Partition exercises")

	squares := MapSlice([]int{1, 2, 3, 4}, func(v int) int { return v * v })
	if len(squares) != 4 {
		t.Fatalf("expected 4 squares, got %d", len(squares))
	}
	expected := []int{1, 4, 9, 16}
	for i, v := range expected {
		if squares[i] != v {
			t.Fatalf("squares[%d] = %d, want %d", i, squares[i], v)
		}
	}

	sum := Reduce(squares, 0, func(acc, next int) int { return acc + next })
	if sum != 30 {
		t.Fatalf("expected sum 30, got %d", sum)
	}

	evens, odds := Partition(squares, func(v int) bool { return v%2 == 0 })
	if len(evens) != 2 || len(odds) != 2 {
		t.Fatalf("expected 2 evens and 2 odds, got %d/%d", len(evens), len(odds))
	}
}

func TestMax(t *testing.T) {
	t.Skip("TODO: remove this skip to activate the Max challenge")

	max, err := Max(1, 4, 2, 9, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if max != 9 {
		t.Fatalf("expected max 9, got %d", max)
	}

	_, err = Max[int]()
	if !errors.Is(err, ErrNoCandidates) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestMemoizer(t *testing.T) {
	t.Skip("TODO: remove this skip to explore the Memoizer")

	var calls atomic.Int32
	memo := NewMemoizer(func(ctx context.Context, key string) (string, error) {
		calls.Add(1)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(10 * time.Millisecond):
			return key + "!", nil
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			got, err := memo.Get(ctx, "echo")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != "echo!" {
				t.Errorf("expected echo!, got %q", got)
			}
		}()
	}
	wg.Wait()

	if calls.Load() != 1 {
		t.Fatalf("expected underlying function to run once, ran %d times", calls.Load())
	}
}

func TestRingBuffer(t *testing.T) {
	t.Skip("TODO: remove this skip to work on the RingBuffer")

	buf := NewRingBuffer[int](3)
	if snapshot := buf.Snapshot(); len(snapshot) != 0 {
		t.Fatalf("expected empty snapshot, got %v", snapshot)
	}

	if _, evicted := buf.Push(1); evicted {
		t.Fatalf("did not expect eviction on first push")
	}
	buf.Push(2)
	buf.Push(3)
	if snapshot := buf.Snapshot(); len(snapshot) != 3 {
		t.Fatalf("expected 3 entries, got %v", snapshot)
	}

	evicted, ok := buf.Push(4)
	if !ok || evicted != 1 {
		t.Fatalf("expected to evict 1, got (%d, %v)", evicted, ok)
	}
	expected := []int{2, 3, 4}
	if snapshot := buf.Snapshot(); len(snapshot) != len(expected) {
		t.Fatalf("snapshot length mismatch: got %d want %d", len(snapshot), len(expected))
	} else {
		for i, v := range expected {
			if snapshot[i] != v {
				t.Fatalf("snapshot[%d] = %d, want %d", i, snapshot[i], v)
			}
		}
	}
}
