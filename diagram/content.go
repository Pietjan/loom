package diagram

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
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

// inferSize estimates a node's box from its rendered content: the widest text
// line sets the width, the line count sets the height, and any inline icons add
// a little width. Deterministic but approximate — no server-side font metrics —
// so Size(w,h) overrides it when exactness matters.
func inferSize(lines []string, content *html.Node) (w, h float64) {
	maxRunes := 0
	for _, ln := range lines {
		if r := len([]rune(ln)); r > maxRunes {
			maxRunes = r
		}
	}
	icons := len(dom.FindAll(content, dom.ByTag(atom.Svg)))
	w = float64(maxRunes)*glyphAdv + 2*padX + float64(icons)*(fontSize+6)
	if w < minNodeW {
		w = minNodeW
	}
	h = float64(len(lines))*lineHeight + 2*padY
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
