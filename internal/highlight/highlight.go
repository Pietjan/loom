// Package highlight fills html nodes with chroma token spans carrying
// chroma's standard short classes (.k, .s, ...), colored by the
// [data-ui="code"] rules in loom.css.
package highlight

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
)

// Into fills parent with the highlighted source. No language, an unknown
// language, and a tokenizer error all degrade to a plain text child —
// highlighting never breaks a page render.
func Into(parent *html.Node, src, lang string) {
	if lang == "" {
		parent.AppendChild(dom.Text(src))
		return
	}
	lexer := lexerFor(lang)
	if lexer == nil {
		parent.AppendChild(dom.Text(src))
		return
	}
	it, err := chroma.Coalesce(lexer).Tokenise(nil, src)
	if err != nil {
		parent.AppendChild(dom.Text(src))
		return
	}
	for _, tok := range it.Tokens() {
		cls := shortClass(tok.Type)
		if tok.Type == chroma.TextWhitespace {
			cls = "" // whitespace needs no span of its own
		}
		if cls == "" {
			parent.AppendChild(dom.Text(tok.Value))
			continue
		}
		span := dom.El(atom.Span, dom.Attr("class", cls))
		span.AppendChild(dom.Text(tok.Value))
		parent.AppendChild(span)
	}
}

// lexerFor resolves a language to a lexer, preferring loom's richer templ
// lexer over chroma's coarser built-in for templ and its aliases.
func lexerFor(lang string) chroma.Lexer {
	switch lang {
	case "templ", "text/x-templ":
		return templLexer
	default:
		return lexers.Get(lang)
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
