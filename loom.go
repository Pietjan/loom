// Package loom is a library of pre-styled, accessible UI components for Go
// and templ (github.com/a-h/templ).
//
// Components are pure Go: they build an *html.Node tree and implement
// templ.Component, so they drop straight into .templ templates:
//
//	@button.New(button.Primary) { Save }
//
// Interactivity comes from modern web platform primitives — the Popover
// API, invoker commands (commandfor/command), <dialog>, <details name> —
// not JavaScript. Styling is Tailwind; compile your CSS with the Tailwind
// CLI using the entry file written by:
//
//	go run github.com/pietjan/loom/cmd/css -accent indigo -base zinc -o assets/css/input.css
package loom

import (
	"context"
	"net/http"

	"github.com/pietjan/loom/internal/ids"
)

// NewContext prepares a context for one render pass: components generate
// sequential, deterministic element IDs (loom-field-1, loom-modal-2, …)
// instead of random ones. Use one prepared context per rendered page.
func NewContext(ctx context.Context) context.Context {
	return ids.WithCounter(ctx)
}

// Middleware installs a fresh ID counter per request, so full-page renders
// get deterministic sequential IDs. Handlers rendering fragments (HTMX
// partials) work without it — IDs fall back to random, collision-free
// values.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(NewContext(r.Context())))
	})
}
