# Groundation: Advanced Go Programming Drills

Groundation is a self-paced curriculum designed to help experienced Go programmers
re-familiarize themselves with modern Go features and edge cases. Each level is a
small scenario with accompanying tests that you can run to measure your progress.

## How the curriculum works

- The repository is organized into themed levels inside [`levels/`](levels/).
- Every level contains:
  - A `challenge.md` file that describes the scenario, the learning goals, and
    additional reading recommendations.
  - A Go package with starter code in `*.go` files. Many implementations include
    `TODO` markers; you are encouraged to replace the naive implementations with
    idiomatic solutions.
  - A `*_test.go` file that defines the acceptance tests for the level. Tests are
    initially skipped with `t.Skip(...)` calls so that the suite passes out of the
    box. When you are ready to attempt the challenge, remove the relevant `t.Skip`
    lines and implement the required functionality until the tests pass.
- Levels are independent; you can tackle them in any order.

## Getting started

1. Install Go 1.22 or newer.
2. Inspect the level directory you want to play through and read its
   `challenge.md` file.
3. Remove the `t.Skip` calls from the tests you want to activate.
4. Run the test suite for that level. For example, to attempt the generics level:

   ```bash
   go test ./levels/01-generics
   ```

5. Iterate on the implementation until all activated tests pass.

## Suggested progression

| Level | Theme | Highlights |
| ----- | ----- | ---------- |
| 01 | Generics & type inference | Type constraints, generic algorithm design, zero-value pitfalls |
| 02 | Concurrency primitives | Context cancellation, worker pools, channel ownership |
| 03 | Advanced standard library usage | `io`, `encoding/json`, `errors`, functional options |

Feel free to extend the curriculum with your own levels as your practice needs evolve.

## Tips for success

- Treat each level as a kata. Start with the naive solution and refactor toward idiomatic Go.
- Explore the Go 1.18+ release notes for more context on generics, fuzzing, and runtime improvements.
- Use tools such as `go test -run` to focus on a specific case and `go test -race` to uncover data races in concurrency-oriented levels.
- Experiment with benchmarks (`go test -bench .`) when performance is a concern.

Happy hacking and welcome back to Go!
