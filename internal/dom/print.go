package dom

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// voidElements render without a closing tag.
var voidElements = map[atom.Atom]bool{
	atom.Area: true, atom.Base: true, atom.Br: true, atom.Col: true,
	atom.Embed: true, atom.Hr: true, atom.Img: true, atom.Input: true,
	atom.Link: true, atom.Meta: true, atom.Source: true, atom.Track: true,
	atom.Wbr: true,
}

// Format renders a node tree as indented HTML for readable golden files
// and test diffs. It is NOT a faithful serializer: whitespace-only text is
// dropped and remaining text is trimmed. Use html.Render for real output.
func Format(n *html.Node) string {
	var sb strings.Builder
	format(&sb, n, 0)
	return sb.String()
}

func format(sb *strings.Builder, n *html.Node, depth int) {
	indent := strings.Repeat("  ", depth)
	switch n.Type {
	case html.TextNode:
		if t := strings.TrimSpace(n.Data); t != "" {
			sb.WriteString(indent)
			sb.WriteString(html.EscapeString(t))
			sb.WriteString("\n")
		}
	case html.ElementNode:
		sb.WriteString(indent)
		sb.WriteString("<")
		sb.WriteString(n.Data)
		for _, a := range n.Attr {
			sb.WriteString(" ")
			sb.WriteString(a.Key)
			if a.Val != "" {
				sb.WriteString(`="`)
				sb.WriteString(html.EscapeString(a.Val))
				sb.WriteString(`"`)
			}
		}
		sb.WriteString(">")
		if voidElements[n.DataAtom] {
			sb.WriteString("\n")
			return
		}
		if onlyText(n) {
			sb.WriteString(html.EscapeString(strings.TrimSpace(n.FirstChild.Data)))
			sb.WriteString("</")
			sb.WriteString(n.Data)
			sb.WriteString(">\n")
			return
		}
		sb.WriteString("\n")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			format(sb, c, depth+1)
		}
		sb.WriteString(indent)
		sb.WriteString("</")
		sb.WriteString(n.Data)
		sb.WriteString(">\n")
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			format(sb, c, depth)
		}
	}
}

// onlyText reports whether the node has exactly one child and it is text.
func onlyText(n *html.Node) bool {
	return n.FirstChild != nil && n.FirstChild == n.LastChild && n.FirstChild.Type == html.TextNode
}
