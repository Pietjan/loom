package diagram

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// contentLines flattens a node's rendered body into visual text lines: text is
// gathered within inline context and broken at block-level elements and <br>.
// It drives both size inference and the accessible name.
func contentLines(n *html.Node) []string {
	var lines []string
	var cur strings.Builder
	flush := func() {
		if s := strings.Join(strings.Fields(cur.String()), " "); s != "" {
			lines = append(lines, s)
		}
		cur.Reset()
	}
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			switch {
			case c.Type == html.TextNode:
				cur.WriteString(c.Data)
				cur.WriteByte(' ')
			case c.Type == html.ElementNode && c.DataAtom == atom.Br:
				flush()
			case c.Type == html.ElementNode && isBlock(c):
				flush()
				walk(c)
				flush()
			case c.Type == html.ElementNode:
				walk(c) // inline element: keep accumulating the line
			}
		}
	}
	walk(n)
	flush()
	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

// inferSize measures a node's box from its rendered content by walking the
// markup and resolving the Tailwind classes on it (see measure.go) — padding,
// borders, gaps, fixed sizes, font sizes and line heights are all known
// values, so this is a real box-model calculation rather than a guess. Only
// proportional glyph advances remain estimated; monospace is exact.
//
// A bare node's body brings its own chrome, so the box is exactly the measured
// content. A default node is padded so its label isn't flush to the border.
func inferSize(content *html.Node, bare bool) (w, h float64) {
	base := style{fontSize: fontSize, lineHeight: fontSize * 1.25} // matches contentClasses
	w, h = measureChildren(content, base)
	if bare {
		return w, h
	}
	w += 2 * padX
	h += 2 * padY
	if w < minNodeW {
		w = minNodeW
	}
	if min := fontSize + 2*padY; h < min {
		h = min
	}
	return w, h
}

var blockAtoms = map[atom.Atom]bool{
	atom.Div: true, atom.P: true, atom.Ul: true, atom.Ol: true, atom.Li: true,
	atom.H1: true, atom.H2: true, atom.H3: true, atom.H4: true, atom.H5: true, atom.H6: true,
	atom.Section: true, atom.Header: true, atom.Footer: true, atom.Article: true,
	atom.Table: true, atom.Tr: true, atom.Pre: true, atom.Blockquote: true, atom.Figure: true,
}

func isBlock(n *html.Node) bool { return blockAtoms[n.DataAtom] }
