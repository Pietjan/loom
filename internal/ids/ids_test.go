package ids_test

import (
	"context"
	"strings"
	"testing"

	"github.com/pietjan/loom/internal/ids"
)

func TestSequentialWithCounter(t *testing.T) {
	ctx := ids.WithCounter(context.Background())
	if got := ids.New(ctx, "field"); got != "loom-field-1" {
		t.Fatalf("got %q", got)
	}
	if got := ids.New(ctx, "modal"); got != "loom-modal-2" {
		t.Fatalf("got %q", got)
	}
}

func TestFreshCounterPerContext(t *testing.T) {
	a := ids.New(ids.WithCounter(context.Background()), "x")
	b := ids.New(ids.WithCounter(context.Background()), "x")
	if a != b {
		t.Fatalf("fresh counters should restart: %q vs %q", a, b)
	}
}

func TestRandomFallback(t *testing.T) {
	ctx := context.Background()
	a := ids.New(ctx, "field")
	b := ids.New(ctx, "field")
	if a == b {
		t.Fatalf("random IDs collided: %q", a)
	}
	if !strings.HasPrefix(a, "loom-field-") {
		t.Fatalf("unexpected shape: %q", a)
	}
}
