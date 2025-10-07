# Level 03 – Advanced Standard Library Practices

This level highlights nuanced uses of the Go standard library. The exercises
focus on robust I/O, JSON decoding edge cases, error wrapping, and iterating over
filesystem state.

## Learning goals

- Implement defensive readers and writers that enforce size limits and deadlines.
- Decode JSON streams with strict schema validation using `encoding/json`.
- Combine `errors.Join` and `errors.Is` to build ergonomic error reporting.
- Explore directory traversal with `fs.WalkDir` and contextual cancellation.

## Exercises

1. Harden `ReadAllWithLimit` so it fails fast on oversized payloads and respects
   context cancellation.
2. Implement `DecodeJSONStrict` to reject unknown fields while still supporting
   optional values via pointers.
3. Finish `Retry` to back off with jitter and wrap errors with attempt metadata.
4. Complete `WalkDirFiltered` to combine file-system iteration with predicate
   filtering and aggregated errors.

See `stdlib_test.go` for testable requirements. Remove the `t.Skip` calls when
you're ready for the challenge.

## Bonus challenges

- Extend `Retry` with circuit-breaker semantics.
- Support streaming JSON decoding via `json.Decoder.Token`.
- Implement file hashing using `crypto/sha256` inside the directory walker.

## Recommended reading

- [`io` package documentation](https://pkg.go.dev/io)
- [`encoding/json` documentation](https://pkg.go.dev/encoding/json)
- [Go 1.20 Release Notes](https://go.dev/doc/go1.20) for details on `errors.Join`.
- [`io/fs` package documentation](https://pkg.go.dev/io/fs)
