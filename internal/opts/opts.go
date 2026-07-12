// Package opts provides the common options every component package
// re-exports — Class, ID, Attr — implemented once, generically, instead of
// hand-copied into every package.
//
// A component config embeds Common and instantiates the generics as
// package vars:
//
//	type Config struct {
//		opts.Common
//		Variant Variant
//	}
//	type Option = func(*Config)
//
//	var (
//		Class = opts.Class[*Config]
//		ID    = opts.ID[*Config]
//		Attr  = opts.Attr[*Config]
//	)
package opts

import (
	"golang.org/x/net/html"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/styles"
)

// Common holds the user-supplied attributes shared by all components.
type Common struct {
	Attrs []html.Attribute
}

func (c *Common) common() *Common { return c }

// HasCommon is satisfied by any config that embeds Common.
type HasCommon interface{ common() *Common }

// Class appends user classes; they win over the component recipe on
// Tailwind conflicts (resolved in Apply via tw-merge).
func Class[T HasCommon](classes string) func(T) {
	return func(t T) {
		c := t.common()
		c.Attrs = dom.AddAttribute(c.Attrs, "class", classes)
	}
}

// ID sets the element id. Composites that generate IDs respect a
// user-supplied one.
func ID[T HasCommon](id string) func(T) {
	return func(t T) {
		c := t.common()
		c.Attrs = dom.SetAttribute(c.Attrs, "id", id)
	}
}

// Attr sets an arbitrary attribute; multiple values are space-joined.
func Attr[T HasCommon](key string, val ...string) func(T) {
	return func(t T) {
		c := t.common()
		a := dom.Attr(key, val...)
		c.Attrs = dom.SetAttribute(c.Attrs, a.Key, a.Val)
	}
}

// UserClass returns the user-supplied classes accumulated so far.
func (c *Common) UserClass() string {
	return dom.GetAttribute(c.Attrs, "class")
}

// UserID returns the user-supplied id, or "".
func (c *Common) UserID() string {
	return dom.GetAttribute(c.Attrs, "id")
}

// Apply finishes a component's root node: user attributes are copied onto
// the node (overriding component defaults of the same name, except the
// data-ui marker), and the class attribute is set to the tw-merge of the
// component recipe with user classes — user classes win.
func (c *Common) Apply(n *html.Node, componentClasses string) {
	for _, a := range c.Attrs {
		if a.Key == "class" || a.Key == dom.MarkerAttr {
			continue
		}
		dom.SetAttr(n, a.Key, a.Val)
	}
	if merged := styles.Merge(componentClasses, c.UserClass()); merged != "" {
		dom.SetAttr(n, "class", merged)
	}
}
