// Package assets embeds the Phosphor SVG sources consumed by the icon
// package. The SVG files are self-consistent (every weight paints with
// fill="currentColor" on a 0 0 256 256 viewBox, with no width/height, so
// CSS sizing wins) — loaders must not override their presentation
// attributes.
package assets

import (
	"embed"
	"errors"
	"fmt"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//go:embed icon/*/*.svg
var files embed.FS

// errNoSVGRoot marks a vendored asset whose root element is not <svg>.
var errNoSVGRoot = errors.New("assets: no <svg> root")

// LoadIcon parses the named icon SVG for a variant (regular, fill) and
// returns its root <svg> node.
func LoadIcon(name, variant string) (*html.Node, error) {
	file, err := files.Open(fmt.Sprintf("icon/%s/%s.svg", variant, name))
	if err != nil {
		return nil, fmt.Errorf("unknown icon %q (variant %q): %w", name, variant, err)
	}
	defer func() { _ = file.Close() }()

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
	return nil, fmt.Errorf("%w: icon %q (variant %q)", errNoSVGRoot, name, variant)
}
