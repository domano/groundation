package generics

import (
	"cmp"
	"context"
	"errors"
	"sync"
)

// ErrNoCandidates is returned by Max when no candidates were provided.
var ErrNoCandidates = errors.New("generics: no candidates provided")

// MapSlice applies fn to each element of in and returns a new slice with the results.
// The current implementation is intentionally naive; optimize memory usage and
// safety as part of the exercise.
func MapSlice[T any, R any](in []T, fn func(T) R) []R {
	if fn == nil {
		return nil
	}
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

// Reduce folds the slice from left to right using fn, starting at init.
// Customize this to behave well with nil slices and zero values.
func Reduce[T any, R any](in []T, init R, fn func(R, T) R) R {
	acc := init
	for _, v := range in {
		acc = fn(acc, v)
	}
	return acc
}

// Partition splits the input slice into two slices based on the predicate result.
func Partition[T any](in []T, pred func(T) bool) (matches, rest []T) {
	for _, v := range in {
		if pred(v) {
			matches = append(matches, v)
		} else {
			rest = append(rest, v)
		}
	}
	return matches, rest
}

// Max returns the largest value among the provided candidates. When candidates is
// empty an error is returned so callers can differentiate the case from the zero value.
func Max[T cmp.Ordered](candidates ...T) (T, error) {
	var zero T
	if len(candidates) == 0 {
		return zero, ErrNoCandidates
	}
	max := candidates[0]
	for _, v := range candidates[1:] {
		if v > max {
			max = v
		}
	}
	return max, nil
}

// Memoizer caches the results of invoking fn and is safe for concurrent use.
type Memoizer[K comparable, V any] struct {
	mu    sync.RWMutex
	fn    func(context.Context, K) (V, error)
	cache map[K]V
	errs  map[K]error
}

// NewMemoizer constructs a Memoizer that wraps fn.
func NewMemoizer[K comparable, V any](fn func(context.Context, K) (V, error)) *Memoizer[K, V] {
	return &Memoizer[K, V]{
		fn:    fn,
		cache: make(map[K]V),
		errs:  make(map[K]error),
	}
}

// Get retrieves the cached value for key or computes it if not present. The context
// can be used by fn to respect cancellation or deadlines. Duplicate computations for
// the same key should be avoided by your enhanced implementation.
func (m *Memoizer[K, V]) Get(ctx context.Context, key K) (V, error) {
	m.mu.RLock()
	if v, ok := m.cache[key]; ok {
		m.mu.RUnlock()
		return v, nil
	}
	if err, ok := m.errs[key]; ok {
		m.mu.RUnlock()
		var zero V
		return zero, err
	}
	m.mu.RUnlock()

	v, err := m.fn(ctx, key)
	m.mu.Lock()
	defer m.mu.Unlock()
	if err != nil {
		m.errs[key] = err
		var zero V
		return zero, err
	}
	m.cache[key] = v
	return v, nil
}

// RingBuffer is a fixed-size circular buffer suitable for generic data.
type RingBuffer[T any] struct {
	data []T
	head int
	size int
}

// NewRingBuffer allocates a buffer that can hold capacity elements.
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity <= 0 {
		panic("generics: capacity must be positive")
	}
	return &RingBuffer[T]{
		data: make([]T, capacity),
	}
}

// Push inserts v into the buffer and returns the element that was evicted, along
// with a boolean indicating whether an eviction took place.
func (r *RingBuffer[T]) Push(v T) (evicted T, evictedOK bool) {
	if len(r.data) == 0 {
		return evicted, false
	}
	if r.size == len(r.data) {
		evicted = r.data[r.head]
		evictedOK = true
	}
	r.data[r.head] = v
	r.head = (r.head + 1) % len(r.data)
	if r.size < len(r.data) {
		r.size++
	}
	return evicted, evictedOK
}

// Snapshot returns the current contents of the buffer in FIFO order.
func (r *RingBuffer[T]) Snapshot() []T {
	if r.size == 0 {
		return nil
	}
	out := make([]T, r.size)
	for i := 0; i < r.size; i++ {
		idx := (r.head - r.size + i + len(r.data)) % len(r.data)
		out[i] = r.data[idx]
	}
	return out
}
