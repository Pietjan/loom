// Package scope is the parent→child coordination channel.
//
// A parent component installs a typed Scope value into the context before
// rendering its children (via render.Children's enrich functions); any
// component rendered inside the block reads it with From. This works
// because templ child blocks receive the context passed at render time,
// not the one present at declaration time — pinned by a test in
// internal/render.
//
// Scopes carry state that changes how a child renders itself (IDs to
// adopt, invalid/required/disabled flags, size inheritance). They must not
// carry sibling-relational wiring — that is what post-passes over the
// materialized tree are for (see internal/dom).
package scope

import "context"

// key is a distinct context key per scope type T.
type key[T any] struct{}

// With returns a context enricher that installs v as the T scope.
// The signature matches render.Children's enrich parameter.
func With[T any](v T) func(context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, key[T]{}, v)
	}
}

// Set installs v as the T scope directly.
func Set[T any](ctx context.Context, v T) context.Context {
	return context.WithValue(ctx, key[T]{}, v)
}

// From returns the T scope installed by the nearest enclosing parent.
func From[T any](ctx context.Context) (T, bool) {
	v, ok := ctx.Value(key[T]{}).(T)
	return v, ok
}
