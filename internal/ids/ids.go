// Package ids generates element IDs for wiring (label for=, commandfor=,
// aria-describedby=).
//
// When the request context was prepared with loom.NewContext (or
// loom.Middleware), IDs are sequential per render pass: deterministic for
// golden tests, unique within a page. Without it — e.g. HTMX fragments
// rendered with a bare context — IDs fall back to crypto/rand, which is
// never deterministic but also never collides across fragments.
package ids

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"sync/atomic"
)

type counterKey struct{}

// WithCounter installs a fresh sequential ID counter.
func WithCounter(ctx context.Context) context.Context {
	return context.WithValue(ctx, counterKey{}, new(atomic.Uint64))
}

// New returns a unique element ID like "loom-modal-3" (counter installed)
// or "loom-modal-a1b2c3d4e5f6" (fallback).
func New(ctx context.Context, prefix string) string {
	if c, ok := ctx.Value(counterKey{}).(*atomic.Uint64); ok {
		return "loom-" + prefix + "-" + strconv.FormatUint(c.Add(1), 10)
	}
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "loom-" + prefix + "-" + hex.EncodeToString(b)
}
