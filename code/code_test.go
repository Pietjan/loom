package code_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/code"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

const goSample = "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}\n"

const diffSample = `diff --git a/main.go b/main.go
index 1234567..89abcde 100644
--- a/main.go
+++ b/main.go
@@ -1,4 +1,5 @@
 package main

 func main() {
-	println("hi")
+	println("hello")
+	println("world")
 }
`

func TestGolden(t *testing.T) {
	testutil.Golden(t, "code-go", code.New(goSample, code.Language("go")))
	testutil.Golden(t, "code-plain", code.New(goSample))
	testutil.Golden(t, "code-unknown-language", code.New(goSample, code.Language("nosuchlang")))
	testutil.Golden(t, "code-diff", code.New(diffSample, code.Diff()))
	testutil.Golden(t, "code-diff-go", code.New(diffSample, code.Language("go"), code.Diff()))
}

func TestHighlighting(t *testing.T) {
	out := testutil.Render(t, code.New(goSample, code.Language("go")))
	if !strings.Contains(out, "<span class=") {
		t.Fatalf("go source not highlighted: %s", out)
	}

	out = testutil.Render(t, code.New(goSample))
	if strings.Contains(out, "<span") {
		t.Fatalf("no language must render plain: %s", out)
	}

	out = testutil.Render(t, code.New(goSample, code.Language("nosuchlang")))
	if strings.Contains(out, "<span") {
		t.Fatalf("unknown language must render plain: %s", out)
	}
}

func TestDiffLines(t *testing.T) {
	out := testutil.Render(t, code.New(diffSample, code.Diff()))
	tree := testutil.NewTree(t, out)
	codeEl := tree.One("code").FirstChild
	if codeEl == nil || codeEl.Data != "code" {
		t.Fatalf("diff pre must wrap a <code>: %s", out)
	}

	lines := 0
	for line := codeEl.FirstChild; line != nil; line = line.NextSibling {
		lines++
	}
	if want := strings.Count(diffSample, "\n"); lines != want {
		t.Fatalf("expected %d line spans, got %d: %s", want, lines, out)
	}

	classOfLine := func(prefix string) string {
		for line := codeEl.FirstChild; line != nil; line = line.NextSibling {
			if line.FirstChild != nil && strings.HasPrefix(line.FirstChild.Data, prefix) {
				return dom.GetAttr(line, "class")
			}
		}
		t.Fatalf("no line starting with %q: %s", prefix, out)
		return ""
	}
	if cls := classOfLine("+\tprintln"); !strings.Contains(cls, "bg-green") {
		t.Fatalf("added line missing green background: %q", cls)
	}
	if cls := classOfLine("-\tprintln"); !strings.Contains(cls, "bg-red") {
		t.Fatalf("removed line missing red background: %q", cls)
	}
	if cls := classOfLine("@@"); !strings.Contains(cls, "bg-blue") {
		t.Fatalf("hunk line missing blue background: %q", cls)
	}
	// File headers are meta, not add/del.
	if cls := classOfLine("+++"); strings.Contains(cls, "bg-green") {
		t.Fatalf("+++ header must not read as an added line: %q", cls)
	}
	if cls := classOfLine("---"); strings.Contains(cls, "bg-red") {
		t.Fatalf("--- header must not read as a removed line: %q", cls)
	}
}

func TestDiffWithLanguage(t *testing.T) {
	out := testutil.Render(t, code.New(diffSample, code.Language("go"), code.Diff()))
	tree := testutil.NewTree(t, out)
	codeEl := tree.One("code").FirstChild

	var hunkHasSpan, changeHasSpan bool
	for line := codeEl.FirstChild; line != nil; line = line.NextSibling {
		text := line.FirstChild
		if text == nil {
			continue
		}
		hasSpan := text.NextSibling != nil
		switch {
		case strings.HasPrefix(text.Data, "@@"), strings.HasPrefix(text.Data, "diff "):
			hunkHasSpan = hunkHasSpan || hasSpan
		case strings.HasPrefix(text.Data, "+"), strings.HasPrefix(text.Data, "-"):
			changeHasSpan = changeHasSpan || hasSpan
		}
	}
	if !changeHasSpan {
		t.Fatalf("changed lines must carry token spans: %s", out)
	}
	if hunkHasSpan {
		t.Fatalf("hunk/header lines must not be highlighted: %s", out)
	}
}

func TestUserClassMerges(t *testing.T) {
	out := testutil.Render(t, code.New("hi", code.Class("p-8")))
	tree := testutil.NewTree(t, out)
	root := tree.One("code")
	cls := dom.GetAttr(root, "class")
	if !strings.Contains(cls, "p-8") || strings.Contains(cls, "p-4") {
		t.Fatalf("user class must win over base p-4: %q", cls)
	}
}

func TestChildBlockNotInjected(t *testing.T) {
	out := testutil.Render(t, testutil.WithChildren(code.New("hi"), testutil.Text("INJECTED")))
	if strings.Contains(out, "INJECTED") {
		t.Fatalf("child block must not leak into code output: %s", out)
	}
}
