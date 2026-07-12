package dom

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Matcher reports whether a node is the one a query is looking for.
type Matcher func(*html.Node) bool

// ByMarker matches element nodes whose data-ui value is one of names.
func ByMarker(names ...string) Matcher {
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		m := MarkerName(n)
		for _, name := range names {
			if m == name {
				return true
			}
		}
		return false
	}
}

// ByTag matches element nodes with one of the given tags.
func ByTag(atoms ...atom.Atom) Matcher {
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		for _, a := range atoms {
			if n.DataAtom == a {
				return true
			}
		}
		return false
	}
}

// ByAttr matches element nodes that carry the attribute, regardless of value.
func ByAttr(key string) Matcher {
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && HasAttr(n, key)
	}
}

// Any matches when any of the given matchers match.
func Any(ms ...Matcher) Matcher {
	return func(n *html.Node) bool {
		for _, m := range ms {
			if m(n) {
				return true
			}
		}
		return false
	}
}

// Find returns the first descendant of root (depth-first, excluding root
// itself) matched by m, or nil.
func Find(root *html.Node, m Matcher) *html.Node {
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if m(c) {
			return c
		}
		if found := Find(c, m); found != nil {
			return found
		}
	}
	return nil
}

// FindAll returns all descendants of root (depth-first, excluding root
// itself) matched by m.
func FindAll(root *html.Node, m Matcher) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if m(c) {
				out = append(out, c)
			}
			walk(c)
		}
	}
	walk(root)
	return out
}

// FindShallow returns the first descendant of root matched by m, but does
// not descend into other components: a non-matching element that carries a
// data-ui marker is treated as opaque and its subtree is skipped.
//
// This is the query post-passes should use for wiring, so that e.g. a
// trigger looking for its button never rewrites a button nested inside a
// card placed in the same block.
func FindShallow(root *html.Node, m Matcher) *html.Node {
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if m(c) {
			return c
		}
		if c.Type == html.ElementNode && HasAttr(c, MarkerAttr) {
			continue // another component's root: opaque
		}
		if found := FindShallow(c, m); found != nil {
			return found
		}
	}
	return nil
}

// FindAllShallow returns all matches with the same descent rule as FindShallow.
func FindAllShallow(root *html.Node, m Matcher) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if m(c) {
				out = append(out, c)
				continue
			}
			if c.Type == html.ElementNode && HasAttr(c, MarkerAttr) {
				continue
			}
			walk(c)
		}
	}
	walk(root)
	return out
}
