package concurrency

import (
	"context"
	"errors"
	"sync"
	"time"
)

// FanIn multiplexes the provided input channels into a single channel. The returned
// channel is closed when the context is done or all producers have closed.
func FanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(inputs))

	forward := func(ch <-chan T) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	for _, ch := range inputs {
		if ch == nil {
			wg.Done()
			continue
		}
		go forward(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// RateLimitedPool processes tasks with a bounded level of concurrency and a
// simple rate limit defined in requests-per-second. The first encountered error is
// returned along with the partial results collected up to that point.
func RateLimitedPool[T any, R any](ctx context.Context, workers int, qps int, tasks []T, fn func(context.Context, T) (R, error)) ([]R, error) {
	if workers <= 0 {
		return nil, errors.New("concurrency: workers must be positive")
	}
	if qps <= 0 {
		return nil, errors.New("concurrency: qps must be positive")
	}
	if fn == nil {
		return nil, errors.New("concurrency: fn must be provided")
	}

	results := make([]R, 0, len(tasks))
	var mu sync.Mutex
	var firstErr error

	ticker := time.NewTicker(time.Second / time.Duration(qps))
	defer ticker.Stop()

	taskCh := make(chan T)
	go func() {
		defer close(taskCh)
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			select {
			case taskCh <- task:
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for task := range taskCh {
				if ctx.Err() != nil {
					return
				}
				res, err := fn(ctx, task)
				mu.Lock()
				if err != nil && firstErr == nil {
					firstErr = err
				}
				if err == nil {
					results = append(results, res)
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	if firstErr != nil {
		return results, firstErr
	}
	if err := ctx.Err(); err != nil {
		return results, err
	}
	return results, nil
}

// WaitForAll waits for the provided functions to complete. Each function is invoked
// in its own goroutine. Any panics should be recovered and returned as errors.
func WaitForAll(funcs ...func() error) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, fn := range funcs {
		if fn == nil {
			continue
		}
		wg.Add(1)
		go func(f func() error) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = errors.New("panic recovered")
					}
					mu.Unlock()
				}
			}()
			if err := f(); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(fn)
	}

	wg.Wait()
	return firstErr
}

// BoundedParallelMap consumes values from the input channel, processes them with fn
// using at most workers goroutines, and emits the transformed values while preserving
// order. Closing the input channel signals completion. The returned channel is
// closed after all work finishes or the context is canceled.
func BoundedParallelMap[T any, R any](ctx context.Context, workers int, in <-chan T, fn func(context.Context, T) (R, error)) (<-chan R, <-chan error) {
	out := make(chan R)
	errCh := make(chan error, 1)

	if workers <= 0 {
		close(out)
		errCh <- errors.New("concurrency: workers must be positive")
		close(errCh)
		return out, errCh
	}
	if fn == nil {
		close(out)
		errCh <- errors.New("concurrency: fn must be provided")
		close(errCh)
		return out, errCh
	}

	type task[T any] struct {
		index int
		value T
	}
	type result[R any] struct {
		index int
		value R
		err   error
	}

	go func() {
		defer close(out)
		defer close(errCh)

		tasks := make(chan task[T])
		results := make(chan result[R])

		var wg sync.WaitGroup
		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for task := range tasks {
					if ctx.Err() != nil {
						return
					}
					res, err := fn(ctx, task.value)
					select {
					case results <- result[R]{index: task.index, value: res, err: err}:
					case <-ctx.Done():
						return
					}
				}
			}()
		}

		idx := 0
		for v := range in {
			select {
			case <-ctx.Done():
				close(tasks)
				wg.Wait()
				close(results)
				return
			case tasks <- task[T]{index: idx, value: v}:
				idx++
			}
		}
		close(tasks)
		go func() {
			wg.Wait()
			close(results)
		}()

		next := 0
		buffer := make(map[int]result[R])
		for {
			select {
			case <-ctx.Done():
				return
			case res, ok := <-results:
				if !ok {
					for next < idx {
						if stored, ok := buffer[next]; ok {
							if stored.err != nil {
								select {
								case errCh <- stored.err:
								default:
								}
							} else {
								select {
								case out <- stored.value:
								case <-ctx.Done():
									return
								}
							}
						}
						next++
					}
					return
				}
				if res.err != nil {
					select {
					case errCh <- res.err:
					default:
					}
					buffer[res.index] = res
					continue
				}
				buffer[res.index] = res
				for {
					stored, ok := buffer[next]
					if !ok {
						break
					}
					if stored.err != nil {
						select {
						case errCh <- stored.err:
						default:
						}
					} else {
						select {
						case out <- stored.value:
						case <-ctx.Done():
							return
						}
					}
					delete(buffer, next)
					next++
				}
			}
		}
	}()

	return out, errCh
}
