# Level 01 – Generics & Type Inference

This level focuses on writing expressive, type-safe APIs with generics. You'll
work with type constraints, zero values, and higher-order functions that mirror
capabilities from other languages while staying idiomatic in Go.

## Learning goals

- Design reusable generic functions and data structures.
- Understand how type inference interacts with composite literals and function
  arguments.
- Build constraints that compose comparability and ordered semantics without
  relying on external packages.
- Handle zero values safely when transforming or aggregating slices and maps.

## Exercises

1. Implement functional-style helpers such as `MapSlice`, `Reduce`, and
   `Partition` while avoiding unnecessary allocations.
2. Build a `Max` function that uses Go 1.21's `cmp.Ordered` constraint and
   returns an error when no candidates are provided.
3. Implement a `Memoize` helper that stores computation results in a
   concurrency-safe cache without leaking goroutines.
4. Explore how Go handles generic methods by finishing the `RingBuffer[T]`
   implementation.

Remove the `t.Skip` lines in `generics_test.go` to activate the tests. Each test
suite represents a stage; progress through them sequentially or jump to the
scenarios that feel most relevant.

## Bonus challenges

- Extend the memoization helper to support cache invalidation.
- Write benchmarks comparing the performance of generic helpers versus type-
  specific implementations.
- Experiment with [`constraints.Float`](https://pkg.go.dev/golang.org/x/exp/constraints)
  and compare it with Go's standard library `cmp.Ordered`.

## Recommended reading

- [Go blog: "When to Use Generics"](https://go.dev/blog/when-generics)
- [Go 1.18 Release Notes](https://go.dev/doc/go1.18)
- [`cmp` package documentation](https://pkg.go.dev/cmp)
- [`sync.Map` documentation](https://pkg.go.dev/sync#Map)
