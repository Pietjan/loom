package highlight

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// templLexer is loom's lexer for the templ templating language
// (github.com/a-h/templ). Chroma ships a "Templ" lexer, but its component-call
// rule swallows a whole `@button.New(button.Primary) {` into one NameFunction
// token, so the Go arguments and string literals lose their own colors. Since
// loom snippets are dominated by such calls, we register a richer lexer here:
// component calls are split into the `@`, the dotted path, and an argument
// state that tokenises the Go expressions inside — including nested calls and
// composite literals like []float64{...}.
var templLexer = buildTemplLexer()

// goLexer is delegated to for embedded Go (top-level declarations, control-flow
// conditions, { ... } interpolations, and component arguments). It is captured
// once so UsingLexer needs no LexerRegistry at tokenise time.
var goLexer = lexers.Get("Go")

func buildTemplLexer() chroma.Lexer {
	using := chroma.UsingLexer(goLexer)

	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      "templ",
			Aliases:   []string{"templ"},
			Filenames: []string{"*.templ"},
			MimeTypes: []string{"text/x-templ"},
			DotAll:    true,
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {
					{Pattern: `//[^\n\r]*`, Type: chroma.CommentSingle},
					{Pattern: `/\*.*?\*/`, Type: chroma.CommentMultiline},
					// Top-level Go declarations, one line at a time.
					{Pattern: `(?m)^(\s*)(package|import|const|var|type|func)([^\n]*)`, Type: chroma.ByGroups(chroma.TextWhitespace, using, using)},
					// templ / css / script declarations.
					{Pattern: `\b(templ|css|script)(\s+)([A-Za-z_]\w*)(\s*)(\([^)]*\))(\s*)(\{)`, Type: chroma.ByGroups(chroma.KeywordDeclaration, chroma.TextWhitespace, chroma.NameFunction, chroma.TextWhitespace, using, chroma.TextWhitespace, chroma.Punctuation)},
					// Component call: the @ marker is a keyword, the qualifier
					// path stays default like a package (matching how
					// code.Language renders in the args), only the final
					// identifier is the function, and the Go argument list is
					// tokenised by the args state.
					{Pattern: `(@)((?:[A-Za-z_]\w*\.)*)([A-Za-z_]\w*)(\()`, Type: chroma.ByGroups(chroma.Keyword, chroma.Name, chroma.NameFunction, chroma.Punctuation), Mutator: chroma.Push("args")},
					// Component reference without a call.
					{Pattern: `(@)((?:[A-Za-z_]\w*\.)*)([A-Za-z_]\w*)`, Type: chroma.ByGroups(chroma.Keyword, chroma.Name, chroma.NameFunction)},
					// Control flow with a Go condition.
					{Pattern: `(?m)^(\s*)(if|for|switch|select)(\s+)([^{}\n]*)(\s*)(\{)`, Type: chroma.ByGroups(chroma.TextWhitespace, chroma.Keyword, chroma.TextWhitespace, using, chroma.TextWhitespace, chroma.Punctuation)},
					{Pattern: `(\})(\s*)(else)(\s*)(if)?(\s*)([^{}\n]*)(\s*)(\{)?`, Type: chroma.ByGroups(chroma.Punctuation, chroma.TextWhitespace, chroma.Keyword, chroma.TextWhitespace, chroma.Keyword, chroma.TextWhitespace, using, chroma.TextWhitespace, chroma.Punctuation)},
					// { goExpr } interpolation.
					{Pattern: `(\{)([^{}\n]+)(\})`, Type: chroma.ByGroups(chroma.Punctuation, using, chroma.Punctuation)},
					// HTML comment.
					{Pattern: `<!--`, Type: chroma.Comment, Mutator: chroma.Push("comment")},
					// HTML tags.
					{Pattern: `(</)([A-Za-z][\w:.-]*)(\s*)(>)`, Type: chroma.ByGroups(chroma.Punctuation, chroma.NameTag, chroma.Text, chroma.Punctuation)},
					{Pattern: `(<)([A-Za-z][\w:.-]*)`, Type: chroma.ByGroups(chroma.Punctuation, chroma.NameTag), Mutator: chroma.Push("tag")},
					{Pattern: "`[^`]*`", Type: chroma.LiteralStringBacktick},
					{Pattern: `&\S*?;`, Type: chroma.NameEntity},
					{Pattern: `\s+`, Type: chroma.TextWhitespace},
					{Pattern: "[^<&@{}`\\s]+", Type: chroma.Text},
					{Pattern: `[@{}<&]`, Type: chroma.Punctuation},
				},
				// args tokenises a Go argument list, balancing brackets so it
				// pops back to root exactly when the opening "(" closes.
				"args": {
					{Pattern: "`[^`]*`", Type: chroma.LiteralStringBacktick},
					{Pattern: `"(\\.|[^"\\])*"`, Type: chroma.LiteralString},
					{Pattern: `'(\\.|[^'\\])*'`, Type: chroma.LiteralStringChar},
					{Pattern: `//[^\n\r]*`, Type: chroma.CommentSingle},
					{Pattern: `/\*.*?\*/`, Type: chroma.CommentMultiline},
					{Pattern: `\b\d+(\.\d+)?\b`, Type: chroma.LiteralNumber},
					{Pattern: `\b(func|range|map|chan|struct|interface|return|nil)\b`, Type: chroma.Keyword},
					{Pattern: `\b(true|false|iota)\b`, Type: chroma.KeywordConstant},
					{Pattern: `\b(string|bool|byte|rune|error|u?int(?:8|16|32|64)?|float(?:32|64)|complex(?:64|128)|uintptr|any)\b`, Type: chroma.KeywordType},
					// Identifier immediately before "(" is a call.
					{Pattern: `[A-Za-z_]\w*(?=\s*\()`, Type: chroma.NameFunction},
					{Pattern: `[A-Za-z_]\w*`, Type: chroma.Name},
					{Pattern: `[([{]`, Type: chroma.Punctuation, Mutator: chroma.Push("args")},
					{Pattern: `[)\]}]`, Type: chroma.Punctuation, Mutator: chroma.Pop(1)},
					{Pattern: `[+\-*/%<>=!&|^~:]+`, Type: chroma.Operator},
					{Pattern: `[.,;]`, Type: chroma.Punctuation},
					{Pattern: `\s+`, Type: chroma.TextWhitespace},
				},
				"tag": {
					{Pattern: `\s+`, Type: chroma.TextWhitespace},
					{Pattern: `([\w:.-]+)(\s*)(=)(\s*)`, Type: chroma.ByGroups(chroma.NameAttribute, chroma.Text, chroma.Operator, chroma.Text), Mutator: chroma.Push("attr")},
					{Pattern: `[\w:.-]+`, Type: chroma.NameAttribute},
					{Pattern: `(/?)(\s*)(>)`, Type: chroma.ByGroups(chroma.Punctuation, chroma.Text, chroma.Punctuation), Mutator: chroma.Pop(1)},
				},
				"attr": {
					{Pattern: `\{[^{}]*\}`, Type: using, Mutator: chroma.Pop(1)},
					{Pattern: `"(\\.|[^"\\])*"`, Type: chroma.LiteralString, Mutator: chroma.Pop(1)},
					{Pattern: `'(\\.|[^'\\])*'`, Type: chroma.LiteralString, Mutator: chroma.Pop(1)},
					{Pattern: `[^\s>]+`, Type: chroma.LiteralString, Mutator: chroma.Pop(1)},
				},
				"comment": {
					{Pattern: `[^-]+`, Type: chroma.Comment},
					{Pattern: `-->`, Type: chroma.Comment, Mutator: chroma.Pop(1)},
					{Pattern: `-`, Type: chroma.Comment},
				},
			}
		},
	)
}
