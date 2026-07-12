// Package dom provides the node construction, attribute, and query
// primitives every loom component builds on.
//
// Composition rules (the contract between components):
//
//   - ctx scopes (internal/scope) change how a child renders itself.
//   - Post-passes use the query API (Find, FindAll, FindShallow) and only
//     set attributes on marker-identified nodes.
//   - Post-passes never restructure the tree: no moving, wrapping, or
//     removing nodes. A component that needs a wrapper renders it itself.
package dom

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// MarkerAttr is the attribute every component root (and named part)
// carries, e.g. data-ui="button" or data-ui="field-label". It powers
// post-processing queries, structural CSS, tests, and user CSS hooks.
const MarkerAttr = "data-ui"

// El constructs an element node.
func El(a atom.Atom, attrs ...html.Attribute) *html.Node {
	return &html.Node{
		Type:     html.ElementNode,
		Data:     a.String(),
		DataAtom: a,
		Attr:     attrs,
	}
}

// Text constructs a text node.
func Text(text string) *html.Node {
	return &html.Node{Type: html.TextNode, Data: text}
}

// CustomEl constructs an element the atom table doesn't know — new
// platform elements like <selectedcontent>.
func CustomEl(tag string, attrs ...html.Attribute) *html.Node {
	return &html.Node{Type: html.ElementNode, Data: tag, Attr: attrs}
}

// Attr constructs an attribute; multiple values are space-joined.
//
// Attribute keys match case-insensitively throughout this package but
// preserve their original spelling: the x/net/html parser case-adjusts
// foreign-content attributes (SVG's viewBox), which must survive
// round-tripping.
func Attr(key string, val ...string) html.Attribute {
	return html.Attribute{Key: key, Val: strings.Join(val, " ")}
}

// Marker constructs the data-ui marker attribute for a component root or part.
func Marker(name string) html.Attribute {
	return Attr(MarkerAttr, name)
}

// MarkerName returns the node's data-ui value, or "" if it has none.
func MarkerName(n *html.Node) string {
	return GetAttr(n, MarkerAttr)
}

// GetAttr returns the value of the attribute, or "" if absent.
func GetAttr(n *html.Node, key string) string {
	if n == nil {
		return ""
	}
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

// HasAttr reports whether the attribute is present (even if empty).
func HasAttr(n *html.Node, key string) bool {
	if n == nil {
		return false
	}
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return true
		}
	}
	return false
}

// SetAttr sets or replaces an attribute on the node.
func SetAttr(n *html.Node, key, val string) {
	n.Attr = SetAttribute(n.Attr, key, val)
}

// AddAttr appends to an existing attribute value (space-separated) or sets it.
func AddAttr(n *html.Node, key, val string) {
	n.Attr = AddAttribute(n.Attr, key, val)
}

// DelAttr removes an attribute from the node.
func DelAttr(n *html.Node, key string) {
	n.Attr = DeleteAttribute(n.Attr, key)
}

// SetAttribute sets or replaces an attribute in a list.
func SetAttribute(attrs []html.Attribute, key, value string) []html.Attribute {
	for i, a := range attrs {
		if strings.EqualFold(a.Key, key) {
			attrs[i] = html.Attribute{Key: a.Key, Val: value}
			return attrs
		}
	}
	return append(attrs, html.Attribute{Key: key, Val: value})
}

// AddAttribute appends to an existing attribute in a list (space-separated) or adds it.
func AddAttribute(attrs []html.Attribute, key, value string) []html.Attribute {
	for i, a := range attrs {
		if strings.EqualFold(a.Key, key) {
			if a.Val == "" {
				attrs[i] = html.Attribute{Key: a.Key, Val: value}
			} else {
				attrs[i] = html.Attribute{Key: a.Key, Val: a.Val + " " + value}
			}
			return attrs
		}
	}
	return append(attrs, html.Attribute{Key: key, Val: value})
}

// GetAttribute returns an attribute value from a list, or "" if absent.
func GetAttribute(attrs []html.Attribute, key string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

// DeleteAttribute removes an attribute from a list.
func DeleteAttribute(attrs []html.Attribute, key string) []html.Attribute {
	out := attrs[:0]
	for _, a := range attrs {
		if !strings.EqualFold(a.Key, key) {
			out = append(out, a)
		}
	}
	return out
}
