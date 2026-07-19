package code_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/code"
	"github.com/pietjan/loom/internal/testutil"
)

// loom ships its own templ lexer because chroma's built-in collapses a whole
// component call into a single token. These tests pin the finer tokenisation:
// the call path, the Go arguments, string and numeric literals, and HTML.
func TestTemplHighlighting(t *testing.T) {
	const src = `@button.New(button.Primary, button.Label("Save")) {
	Save
}`
	out := testutil.Render(t, code.New(src, code.Language("templ")))

	cases := map[string]string{
		"@ marker is a keyword":                  `<span class="k">@</span>`,
		"qualifier stays default like a package": `<span class="n">button.</span>`,
		"only the final identifier is a function": `<span class="nf">New</span>`,
		"nested call name is highlighted":          `<span class="nf">Label</span>`,
		"string argument is a literal":             `<span class="s">&#34;Save&#34;</span>`,
	}
	for name, want := range cases {
		if !strings.Contains(out, want) {
			t.Errorf("%s: missing %q in:\n%s", name, want, out)
		}
	}
}

func TestTemplHighlightsCompositeLiterals(t *testing.T) {
	const src = `@chart.New(chart.Series("Signups", []float64{40, 60}))`
	out := testutil.Render(t, code.New(src, code.Language("templ")))

	if !strings.Contains(out, `<span class="kt">float64</span>`) {
		t.Errorf("builtin type not highlighted:\n%s", out)
	}
	if !strings.Contains(out, `<span class="m">40</span>`) {
		t.Errorf("numeric literal not highlighted:\n%s", out)
	}
}

func TestTemplHighlightsMarkup(t *testing.T) {
	const src = `<div class="grid">
	@icon.New(icon.Home)
</div>`
	out := testutil.Render(t, code.New(src, code.Language("templ")))

	if !strings.Contains(out, `<span class="nt">div</span>`) {
		t.Errorf("HTML tag not highlighted:\n%s", out)
	}
	if !strings.Contains(out, `<span class="na">class</span>`) {
		t.Errorf("HTML attribute not highlighted:\n%s", out)
	}
}
