// Package assets embeds the heroicons SVG sources consumed by the icon
// package. The SVG files are self-consistent (outline strokes with
// currentColor, solid/mini/micro fill with currentColor, correct viewBox
// per variant) — loaders must not override their presentation attributes.
package assets

import (
	"embed"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//go:embed icon/*/*.svg
var files embed.FS

// LoadIcon parses the named icon SVG for a variant (outline, solid, mini,
// micro) and returns its root <svg> node.
func LoadIcon(name, variant string) (*html.Node, error) {
	file, err := files.Open(fmt.Sprintf("icon/%s/%s.svg", variant, name))
	if err != nil {
		return nil, fmt.Errorf("unknown icon %q (variant %q): %w", name, variant, err)
	}
	defer file.Close()

	container := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(file, container)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Type == html.ElementNode && n.DataAtom == atom.Svg {
			return n, nil
		}
	}
	return nil, fmt.Errorf("no <svg> root in icon %q (variant %q)", name, variant)
}

// IconNames lists the icon names available for a variant, without the .svg
// suffix. Used by cmd/icons for code generation.
func IconNames(variant string) ([]string, error) {
	entries, err := files.ReadDir("icon/" + variant)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if n, ok := strings.CutSuffix(e.Name(), ".svg"); ok {
			names = append(names, n)
		}
	}
	return names, nil
}
