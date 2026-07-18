package markdown

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
)

// highlightInto fills a <code> element with the highlighted source:
// chroma token spans carrying chroma's standard short classes (.k, .s,
// ...), colored by the [data-ui="markdown-code"] rules in loom.css. No
// language, an unknown language, and a tokenizer error all degrade to a
// plain text child — highlighting never breaks a page render.
func highlightInto(code *html.Node, src, lang string) {
	if lang == "" {
		code.AppendChild(dom.Text(src))
		return
	}
	lexer := lexers.Get(lang)
	if lexer == nil {
		code.AppendChild(dom.Text(src))
		return
	}
	it, err := chroma.Coalesce(lexer).Tokenise(nil, src)
	if err != nil {
		code.AppendChild(dom.Text(src))
		return
	}
	for _, tok := range it.Tokens() {
		cls := shortClass(tok.Type)
		if tok.Type == chroma.TextWhitespace {
			cls = "" // whitespace needs no span of its own
		}
		if cls == "" {
			code.AppendChild(dom.Text(tok.Value))
			continue
		}
		span := dom.El(atom.Span, dom.Attr("class", cls))
		span.AppendChild(dom.Text(tok.Value))
		code.AppendChild(span)
	}
}

// shortClass resolves a token type to chroma's standard short class,
// walking up the token category hierarchy the way chroma's own HTML
// formatter does.
func shortClass(t chroma.TokenType) string {
	for t != 0 {
		if cls, ok := chroma.StandardTypes[t]; ok {
			return cls
		}
		t = t.Parent()
	}
	return ""
}
