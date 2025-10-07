package stdlib

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestReadAllWithLimit(t *testing.T) {
	t.Skip("TODO: remove this skip to attempt ReadAllWithLimit")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	data := strings.Repeat("a", 1024)

	out, err := ReadAllWithLimit(ctx, strings.NewReader(data), 2048)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != len(data) {
		t.Fatalf("expected %d bytes, got %d", len(data), len(out))
	}

	if _, err := ReadAllWithLimit(ctx, strings.NewReader(data), 512); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}

	cancel()
	if _, err := ReadAllWithLimit(ctx, slowReader{delay: 50 * time.Millisecond}, 2048); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}

type slowReader struct {
	delay time.Duration
}

func (s slowReader) Read(p []byte) (int, error) {
	time.Sleep(s.delay)
	return copy(p, []byte("slow")), io.EOF
}

func TestDecodeJSONStrict(t *testing.T) {
	t.Skip("TODO: remove this skip to work on DecodeJSONStrict")

	type payload struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	data := `{"id": 1, "name": "gopher"}`
	got, err := DecodeJSONStrict[payload]([]byte(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "gopher" {
		t.Fatalf("expected name gopher, got %q", got.Name)
	}

	bad := `{"id": 2, "nickname": "go"}`
	if _, err := DecodeJSONStrict[payload]([]byte(bad)); err == nil {
		t.Fatalf("expected unknown field error")
	}
}

func TestRetry(t *testing.T) {
	t.Skip("TODO: remove this skip to explore Retry")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	attempts := 0
	err := Retry(ctx, 3, 10*time.Millisecond, func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("boom")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected retry to eventually succeed, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}

	attempts = 0
	err = Retry(context.Background(), 2, 5*time.Millisecond, func(context.Context) error {
		attempts++
		return errors.New("still failing")
	})
	if err == nil {
		t.Fatalf("expected aggregated error from retry")
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestWalkDirFiltered(t *testing.T) {
	t.Skip("TODO: remove this skip to take on WalkDirFiltered")

	fsys := fstest.MapFS{
		"root/a.txt":        {Data: []byte("a")},
		"root/b/b.txt":      {Data: []byte("b")},
		"root/b/ignore.tmp": {Data: []byte("tmp")},
	}

	ctx := context.Background()
	paths, err := WalkDirFiltered(ctx, fsys, "root", func(d fs.DirEntry) bool {
		return strings.HasSuffix(d.Name(), ".txt")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}
}
