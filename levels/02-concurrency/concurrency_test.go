package concurrency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestFanIn(t *testing.T) {
	t.Skip("TODO: remove this skip to attempt the FanIn exercises")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	producers := []<-chan int{
		produce(ctx, []int{1, 3, 5}, 5*time.Millisecond),
		produce(ctx, []int{2, 4, 6}, 7*time.Millisecond),
	}

	seen := map[int]bool{}
	for v := range FanIn(ctx, producers...) {
		seen[v] = true
	}
	for i := 1; i <= 6; i++ {
		if !seen[i] {
			t.Fatalf("missing value %d", i)
		}
	}
}

func produce(ctx context.Context, values []int, delay time.Duration) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range values {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}
			select {
			case <-ctx.Done():
				return
			case ch <- v:
			}
		}
	}()
	return ch
}

func TestRateLimitedPool(t *testing.T) {
	t.Skip("TODO: remove this skip to work on RateLimitedPool")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var calls atomic.Int32
	tasks := []int{1, 2, 3, 4}
	start := time.Now()
	results, err := RateLimitedPool(ctx, 2, 10, tasks, func(ctx context.Context, v int) (int, error) {
		calls.Add(1)
		return v * 2, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != int32(len(tasks)) {
		t.Fatalf("expected %d calls, got %d", len(tasks), calls.Load())
	}
	if time.Since(start) < 300*time.Millisecond/10 {
		t.Fatalf("rate limiter did not appear to engage")
	}
	if len(results) != len(tasks) {
		t.Fatalf("expected %d results, got %d", len(tasks), len(results))
	}
}

func TestWaitForAll(t *testing.T) {
	t.Skip("TODO: remove this skip to activate WaitForAll")

	err := WaitForAll(
		func() error { return nil },
		func() error { return errors.New("boom") },
		func() error { panic("nope") },
	)
	if err == nil {
		t.Fatalf("expected an error from WaitForAll")
	}
}

func TestBoundedParallelMap(t *testing.T) {
	t.Skip("TODO: remove this skip to take on BoundedParallelMap")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	in := make(chan int)
	go func() {
		defer close(in)
		for i := 0; i < 6; i++ {
			in <- i
		}
	}()

	out, errs := BoundedParallelMap(ctx, 3, in, func(ctx context.Context, v int) (int, error) {
		time.Sleep(5 * time.Millisecond)
		if v == 4 {
			return 0, errors.New("bad value")
		}
		return v * v, nil
	})

	var results []int
	for v := range out {
		results = append(results, v)
	}
	expected := []int{0, 1, 4, 9, 25}
	if len(results) != len(expected) {
		t.Fatalf("expected %d results, got %d", len(expected), len(results))
	}
	for i, v := range expected {
		if results[i] != v {
			t.Fatalf("results[%d] = %d, want %d", i, results[i], v)
		}
	}

	select {
	case err, ok := <-errs:
		if !ok || err == nil {
			t.Fatalf("expected to observe an error")
		}
	default:
		t.Fatalf("expected an error to be reported")
	}
}
