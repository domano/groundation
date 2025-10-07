# Level 02 – Concurrency Primitives & Patterns

This level revisits the building blocks of concurrent Go applications. Practice
coordinating goroutines, managing cancellation, and designing APIs that make
ownership explicit.

## Learning goals

- Combine multiple producer channels into a single stream while respecting
  cancellation semantics.
- Design worker pools that propagate context errors and surface partial results.
- Integrate ticker/rate-limiting logic to protect downstream services.
- Encapsulate goroutine lifecycles to prevent leaks.

## Exercises

1. Rework `FanIn` to fairly multiplex values from N input channels until the
   parent context is canceled.
2. Finish `RateLimitedPool` so that it schedules work respecting the provided
   QPS limit while capturing the first error encountered.
3. Implement `WaitForAll` to join background goroutines and surface the most
   relevant error.
4. Polish `BoundedParallelMap` to process a stream without unbounded buffering.

Refer to `concurrency_test.go` for specific acceptance criteria. Remove the
`t.Skip` calls to activate the stages.

## Bonus challenges

- Swap the rate limiter for a token bucket and compare throughput.
- Add metrics hooks (e.g., callbacks) to inspect worker lifecycle events.
- Try running the tests with the race detector enabled: `go test -race ./...`.

## Recommended reading

- [Go blog: "Go Concurrency Patterns"](https://go.dev/blog/pipelines)
- [`context` package documentation](https://pkg.go.dev/context)
- [`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup)
- [Go blog: "First Class Functions in Go"](https://go.dev/blog/facode)
