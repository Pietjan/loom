// Package styles assembles Tailwind class strings.
//
// Each component keeps its class recipe in a single classes(cfg) function
// in its style.go, built with Builder. Class strings must be complete
// literals in Go source — never assembled with fmt.Sprintf fragments — or
// the Tailwind CLI scanner will not see them and the utilities will be
// missing from the compiled CSS.
package styles

import (
	"strings"

	twmerge "github.com/Oudwins/tailwind-merge-go"
)

// Builder accumulates class fragments for a component recipe.
type Builder struct {
	parts []string
}

// Add appends class fragments unconditionally.
func (b *Builder) Add(classes ...string) {
	b.parts = append(b.parts, classes...)
}

// If appends a class fragment when cond holds.
func (b *Builder) If(cond bool, classes string) {
	if cond {
		b.parts = append(b.parts, classes)
	}
}

// Match appends the fragment mapped to key, if any — the variant/size
// switch replacement.
func Match[K comparable](b *Builder, key K, m map[K]string) {
	if classes, ok := m[key]; ok {
		b.parts = append(b.parts, classes)
	}
}

// String returns the joined recipe, sorted into canonical order.
func (b *Builder) String() string {
	return Sort(strings.Join(b.parts, " "))
}

// Merge resolves Tailwind conflicts between the component recipe and
// user-supplied classes — user always wins — and returns the result in
// canonical order (stable golden files).
func Merge(component, user string) string {
	if user == "" {
		return Sort(component)
	}
	return Sort(twmerge.Merge(component, user))
}
