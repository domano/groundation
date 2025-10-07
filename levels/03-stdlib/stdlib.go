package stdlib

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"path/filepath"
	"sync"
	"time"
)

// ErrLimitExceeded is returned when ReadAllWithLimit encounters more data than allowed.
var ErrLimitExceeded = errors.New("stdlib: read limit exceeded")

// ReadAllWithLimit reads from r until EOF or the limit is reached. If the context is
// canceled before the read completes an error is returned. When the limit would be
// exceeded, ErrLimitExceeded is returned.
func ReadAllWithLimit(ctx context.Context, r io.Reader, limit int64) ([]byte, error) {
	if limit <= 0 {
		return nil, ErrLimitExceeded
	}
	buf := make([]byte, 0, int(min64(limit, 32*1024)))
	tmp := make([]byte, 32*1024)
	var total int64
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, err := r.Read(tmp)
		if n > 0 {
			total += int64(n)
			if total > limit {
				return nil, ErrLimitExceeded
			}
			buf = append(buf, tmp[:n]...)
		}
		if errors.Is(err, io.EOF) {
			return buf, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

// DecodeJSONStrict decodes data into a value of type T while disallowing unknown
// fields. Pointer fields may remain nil when absent.
func DecodeJSONStrict[T any](data []byte) (T, error) {
	var zero T
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var out T
	if err := dec.Decode(&out); err != nil {
		return zero, err
	}
	if dec.More() {
		return zero, errors.New("stdlib: trailing data")
	}
	return out, nil
}

// Retry invokes fn up to attempts times using exponential backoff with jitter. When
// the context is canceled the retry loop stops immediately. The returned error wraps
// all attempt errors via errors.Join.
func Retry(ctx context.Context, attempts int, baseDelay time.Duration, fn func(context.Context) error) error {
	if attempts <= 0 {
		return errors.New("stdlib: attempts must be positive")
	}
	if baseDelay <= 0 {
		baseDelay = 10 * time.Millisecond
	}
	if fn == nil {
		return errors.New("stdlib: fn must be provided")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var errs []error

	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			errs = append(errs, err)
			break
		}
		attemptCtx, cancel := context.WithCancel(ctx)
		err := fn(attemptCtx)
		cancel()
		if err == nil {
			return nil
		}
		errs = append(errs, fmt.Errorf("stdlib: attempt %d failed: %w", i+1, err))
		select {
		case <-ctx.Done():
			errs = append(errs, ctx.Err())
			return errors.Join(errs...)
		case <-time.After(backoffWithJitter(baseDelay, i, rng)):
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func backoffWithJitter(base time.Duration, attempt int, rng *rand.Rand) time.Duration {
	max := base << attempt
	if max <= 0 {
		max = base
	}
	if max > time.Second {
		max = time.Second
	}
	jitter := time.Duration(rng.Int63n(int64(max/2 + 1)))
	return max + jitter
}

// WalkDirFiltered walks fsys rooted at root collecting files for which filter returns true.
// Errors encountered during traversal are aggregated via errors.Join.
func WalkDirFiltered(ctx context.Context, fsys fs.FS, root string, filter func(fs.DirEntry) bool) ([]string, error) {
	if fsys == nil {
		return nil, errors.New("stdlib: fsys must be provided")
	}

	var mu sync.Mutex
	var matches []string
	var errs []error

	walkErr := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
			return nil
		}
		if d == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if filter != nil && !filter(d) {
			return nil
		}
		if d.Type().IsRegular() {
			mu.Lock()
			matches = append(matches, filepath.ToSlash(path))
			mu.Unlock()
		}
		return nil
	})

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		errs = append(errs, walkErr)
	}
	if err := ctx.Err(); err != nil {
		errs = append(errs, err)
	}

	var err error
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return matches, err
}

func min64(a int64, b int) int64 {
	if a < int64(b) {
		return a
	}
	return int64(b)
}
