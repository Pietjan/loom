package markdown

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gast "github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/code"
	"github.com/pietjan/loom/heading"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/link"
	"github.com/pietjan/loom/separator"
	"github.com/pietjan/loom/text"
)

// walker builds the *html.Node tree from a goldmark AST. Block and inline
// handlers append themselves to the parent they are given; an unmapped
// node kind is a loom bug and fails the render loudly.
type walker struct {
	ctx    context.Context
	source []byte
	cfg    Config
}

// blocks appends n's block children to parent.
func (w *walker) blocks(n gast.Node, parent *html.Node) error {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if err := w.block(c, parent); err != nil {
			return err
		}
	}
	return nil
}

func (w *walker) block(n gast.Node, parent *html.Node) error {
	switch n := n.(type) {
	case *gast.Heading:
		h, err := heading.Node(w.ctx, heading.Level(n.Level), heading.Class(headingClass(n.Level)))
		if err != nil {
			return err
		}
		if err := w.inlines(n, h); err != nil {
			return err
		}
		parent.AppendChild(h)
	case *gast.Paragraph:
		p, err := text.Node(w.ctx, text.Class(paragraphClass()))
		if err != nil {
			return err
		}
		if err := w.inlines(n, p); err != nil {
			return err
		}
		parent.AppendChild(p)
	case *gast.TextBlock:
		// Tight list items carry a TextBlock instead of a Paragraph:
		// inline its content straight into the parent <li>.
		return w.inlines(n, parent)
	case *gast.Blockquote:
		q := dom.El(atom.Blockquote, dom.Marker("markdown-blockquote"), dom.Attr("class", blockquoteClass()))
		if err := w.blocks(n, q); err != nil {
			return err
		}
		parent.AppendChild(q)
	case *gast.List:
		return w.list(n, parent)
	case *gast.FencedCodeBlock:
		var lang string
		if l := n.Language(w.source); l != nil {
			lang = string(l)
		}
		return w.codeBlock(w.blockLines(n), lang, parent)
	case *gast.CodeBlock:
		return w.codeBlock(w.blockLines(n), "", parent)
	case *gast.ThematicBreak:
		hr, err := separator.Node(w.ctx, separator.Class("my-8"))
		if err != nil {
			return err
		}
		parent.AppendChild(hr)
	case *gast.HTMLBlock:
		if !w.cfg.unsafe {
			return nil
		}
		var sb strings.Builder
		sb.WriteString(w.blockLines(n))
		if n.HasClosure() {
			sb.Write(n.ClosureLine.Value(w.source))
		}
		return appendFragment(parent, sb.String())
	case *extast.Table:
		return w.table(n, parent)
	default:
		return fmt.Errorf("markdown: unsupported block kind %s", n.Kind())
	}
	return nil
}

func (w *walker) list(n *gast.List, parent *html.Node) error {
	// A list nested in a list item sits tighter than a top-level one.
	nested := n.Parent() != nil && n.Parent().Kind() == gast.KindListItem
	tag := atom.Ul
	if n.IsOrdered() {
		tag = atom.Ol
	}
	list := dom.El(tag, dom.Attr("class", listClass(n.IsOrdered(), nested)))
	if n.IsOrdered() && n.Start > 1 {
		dom.SetAttr(list, "start", strconv.Itoa(n.Start))
	}
	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		li := dom.El(atom.Li, dom.Attr("class", itemClass(isTask(item))))
		if err := w.blocks(item, li); err != nil {
			return err
		}
		list.AppendChild(li)
	}
	parent.AppendChild(list)
	return nil
}

// isTask reports whether a list item starts with a GFM task checkbox.
func isTask(item gast.Node) bool {
	first := item.FirstChild()
	if first == nil {
		return false
	}
	_, ok := first.FirstChild().(*extast.TaskCheckBox)
	return ok
}

// codeBlock renders a fence through the code component (embeddable, like
// heading and text), adding only markdown's block spacing on top.
func (w *walker) codeBlock(src, lang string, parent *html.Node) error {
	pre, err := code.Node(w.ctx, src, code.Language(lang), code.Class("mt-4 first:mt-0"))
	if err != nil {
		return err
	}
	parent.AppendChild(pre)
	return nil
}

// blockLines concatenates a block node's raw source lines.
func (w *walker) blockLines(n gast.Node) string {
	var sb strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		sb.Write(line.Value(w.source))
	}
	return sb.String()
}

func (w *walker) table(n *extast.Table, parent *html.Node) error {
	wrap := dom.El(atom.Div, dom.Marker("markdown-table"), dom.Attr("class", tableWrapperClass()))
	tbl := dom.El(atom.Table, dom.Attr("class", tableClass()))
	wrap.AppendChild(tbl)
	var tbody *html.Node
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		switch row := row.(type) {
		case *extast.TableHeader:
			thead := dom.El(atom.Thead, dom.Attr("class", tableHeadClass()))
			tr := dom.El(atom.Tr)
			if err := w.tableCells(row, tr, atom.Th); err != nil {
				return err
			}
			thead.AppendChild(tr)
			tbl.AppendChild(thead)
		case *extast.TableRow:
			if tbody == nil {
				tbody = dom.El(atom.Tbody, dom.Attr("class", tableBodyClass()))
				tbl.AppendChild(tbody)
			}
			tr := dom.El(atom.Tr)
			if err := w.tableCells(row, tr, atom.Td); err != nil {
				return err
			}
			tbody.AppendChild(tr)
		default:
			return fmt.Errorf("markdown: unsupported table child kind %s", row.Kind())
		}
	}
	parent.AppendChild(wrap)
	return nil
}

func (w *walker) tableCells(row gast.Node, tr *html.Node, tag atom.Atom) error {
	for c := row.FirstChild(); c != nil; c = c.NextSibling() {
		cell, ok := c.(*extast.TableCell)
		if !ok {
			return fmt.Errorf("markdown: unsupported table cell kind %s", c.Kind())
		}
		td := dom.El(tag, dom.Attr("class", tableCellClass(tag == atom.Th, cell.Alignment)))
		if tag == atom.Th {
			dom.SetAttr(td, "scope", "col")
		}
		if err := w.inlines(cell, td); err != nil {
			return err
		}
		tr.AppendChild(td)
	}
	return nil
}

// inlines appends n's inline children to parent.
func (w *walker) inlines(n gast.Node, parent *html.Node) error {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if err := w.inline(c, parent); err != nil {
			return err
		}
	}
	return nil
}

func (w *walker) inline(n gast.Node, parent *html.Node) error {
	switch n := n.(type) {
	case *gast.Text:
		parent.AppendChild(dom.Text(string(n.Segment.Value(w.source))))
		if n.HardLineBreak() {
			parent.AppendChild(dom.El(atom.Br))
		} else if n.SoftLineBreak() {
			parent.AppendChild(dom.Text("\n"))
		}
	case *gast.String:
		parent.AppendChild(dom.Text(string(n.Value)))
	case *gast.CodeSpan:
		code := dom.El(atom.Code, dom.Attr("class", inlineCodeClass()))
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			switch c := c.(type) {
			case *gast.Text:
				code.AppendChild(dom.Text(string(c.Segment.Value(w.source))))
			case *gast.String:
				code.AppendChild(dom.Text(string(c.Value)))
			}
		}
		parent.AppendChild(code)
	case *gast.Emphasis:
		el := dom.El(atom.Em)
		if n.Level >= 2 {
			el = dom.El(atom.Strong, dom.Attr("class", "font-semibold"))
		}
		if err := w.inlines(n, el); err != nil {
			return err
		}
		parent.AppendChild(el)
	case *gast.Link:
		a, err := link.Node(w.ctx, string(n.Destination))
		if err != nil {
			return err
		}
		if len(n.Title) > 0 {
			dom.SetAttr(a, "title", string(n.Title))
		}
		if err := w.inlines(n, a); err != nil {
			return err
		}
		parent.AppendChild(a)
	case *gast.AutoLink:
		url := string(n.URL(w.source))
		if n.AutoLinkType == gast.AutoLinkEmail && !strings.HasPrefix(url, "mailto:") {
			url = "mailto:" + url
		}
		a, err := link.Node(w.ctx, url)
		if err != nil {
			return err
		}
		a.AppendChild(dom.Text(string(n.Label(w.source))))
		parent.AppendChild(a)
	case *gast.Image:
		img := dom.El(atom.Img,
			dom.Attr("src", string(n.Destination)),
			dom.Attr("alt", plainText(n, w.source)),
			dom.Attr("class", imageClass()))
		if len(n.Title) > 0 {
			dom.SetAttr(img, "title", string(n.Title))
		}
		parent.AppendChild(img)
	case *gast.RawHTML:
		if !w.cfg.unsafe {
			return nil
		}
		var sb strings.Builder
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			sb.Write(seg.Value(w.source))
		}
		return appendFragment(parent, sb.String())
	case *extast.Strikethrough:
		del := dom.El(atom.Del, dom.Attr("class", "line-through"))
		if err := w.inlines(n, del); err != nil {
			return err
		}
		parent.AppendChild(del)
	case *extast.TaskCheckBox:
		box := dom.El(atom.Input,
			dom.Attr("type", "checkbox"),
			dom.Attr("disabled", ""),
			dom.Attr("class", checkboxClass()))
		if n.IsChecked {
			dom.SetAttr(box, "checked", "")
		}
		parent.AppendChild(box)
	default:
		return fmt.Errorf("markdown: unsupported inline kind %s", n.Kind())
	}
	return nil
}

// plainText flattens a node's descendant text — the spec-correct alt for
// images.
func plainText(n gast.Node, source []byte) string {
	var sb strings.Builder
	collectText(n, source, &sb)
	return sb.String()
}

func collectText(n gast.Node, source []byte, sb *strings.Builder) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c := c.(type) {
		case *gast.Text:
			sb.Write(c.Segment.Value(source))
		case *gast.String:
			sb.Write(c.Value)
		default:
			collectText(c, source, sb)
		}
	}
}

// appendFragment parses raw HTML against the real parent — the same
// parser-context rule internal/render documents: a neutral context would
// silently drop table-structure elements.
func appendFragment(parent *html.Node, raw string) error {
	nodes, err := html.ParseFragment(strings.NewReader(raw), parent)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		parent.AppendChild(n)
	}
	return nil
}
