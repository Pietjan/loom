package markdown_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/markdown"
)

func TestGolden(t *testing.T) {
	testutil.Golden(t, "markdown-document", markdown.New(strings.Join([]string{
		"# Title",
		"",
		"A paragraph with **strong**, *emphasis*, `code`, ~~gone~~, and",
		"a [link](https://example.com \"Example\").",
		"",
		"## Section",
		"",
		"---",
		"",
		"### Sub",
		"",
		"Autolink: https://loom.dev and mail@example.com.",
	}, "\n")))

	testutil.Golden(t, "markdown-lists", markdown.New(strings.Join([]string{
		"- one",
		"- two",
		"  - two point one",
		"  - two point two",
		"",
		"3. three",
		"",
		"4. four",
	}, "\n")))

	testutil.Golden(t, "markdown-code-go", markdown.New(strings.Join([]string{
		"```go",
		`func main() { fmt.Println("hi") } // greet`,
		"```",
	}, "\n")))

	testutil.Golden(t, "markdown-code-unknown", markdown.New(strings.Join([]string{
		"```nosuchlang",
		"plain as day",
		"```",
	}, "\n")))

	testutil.Golden(t, "markdown-table", markdown.New(strings.Join([]string{
		"| Name | Amount | Note |",
		"| :--- | ---: | :---: |",
		"| Alpha | 10 | first |",
		"| Beta | 20 | second |",
	}, "\n")))

	testutil.Golden(t, "markdown-blockquote", markdown.New(strings.Join([]string{
		"> Quoted wisdom",
		">",
		"> - even a list",
	}, "\n")))

	testutil.Golden(t, "markdown-tasklist", markdown.New(strings.Join([]string{
		"- [x] done",
		"- [ ] pending",
	}, "\n")))
}

func TestHeadingLevels(t *testing.T) {
	out := testutil.Render(t, markdown.New("## Section"))
	tree := testutil.NewTree(t, out)
	h := tree.One("heading")
	if h.Data != "h2" {
		t.Fatalf("expected <h2>, got <%s>", h.Data)
	}
	cls := dom.GetAttr(h, "class")
	if !strings.Contains(cls, "text-xl") {
		t.Fatalf("h2 missing text-xl override: %q", cls)
	}
	// heading's own default size must have been merged away.
	if strings.Contains(cls, "text-sm") {
		t.Fatalf("heading default text-sm survived tw-merge: %q", cls)
	}
}

func TestCodeHighlighting(t *testing.T) {
	out := testutil.Render(t, markdown.New("```go\nfunc main() {}\n```"))
	if !strings.Contains(out, "<span class=") {
		t.Fatalf("go fence not highlighted: %s", out)
	}

	out = testutil.Render(t, markdown.New("```nosuchlang\nfunc main() {}\n```"))
	if strings.Contains(out, "<span") {
		t.Fatalf("unknown language must render plain: %s", out)
	}
}

func TestTightAndLooseLists(t *testing.T) {
	tight := testutil.Render(t, markdown.New("- one\n- two"))
	if strings.Contains(tight, "<p") {
		t.Fatalf("tight list items must not wrap in <p>: %s", tight)
	}
	loose := testutil.Render(t, markdown.New("- one\n\n- two"))
	if !strings.Contains(loose, "<p") {
		t.Fatalf("loose list items must wrap in <p>: %s", loose)
	}
}

func TestRawHTML(t *testing.T) {
	src := "before\n\n<div class=\"raw\">html</div>\n\nafter"
	out := testutil.Render(t, markdown.New(src))
	if strings.Contains(out, "raw") {
		t.Fatalf("raw HTML must be dropped by default: %s", out)
	}
	out = testutil.Render(t, markdown.New(src, markdown.Unsafe()))
	if !strings.Contains(out, `<div class="raw">`) {
		t.Fatalf("Unsafe() must pass raw HTML through: %s", out)
	}
}

func TestTextEscaped(t *testing.T) {
	out := testutil.Render(t, markdown.New(`paragraph with \<script>alert(1)\</script> inline`))
	if strings.Contains(out, "<script>") {
		t.Fatalf("script tag must be escaped: %s", out)
	}
	if !strings.Contains(out, "&lt;script&gt;") {
		t.Fatalf("escaped script text missing: %s", out)
	}
}

func TestTaskList(t *testing.T) {
	out := testutil.Render(t, markdown.New("- [x] done\n- [ ] pending"))
	if !strings.Contains(out, `type="checkbox"`) {
		t.Fatalf("task checkbox missing: %s", out)
	}
	if !strings.Contains(out, "checked") {
		t.Fatalf("checked state missing: %s", out)
	}
	if !strings.Contains(out, "disabled") {
		t.Fatalf("task checkbox must be disabled: %s", out)
	}
}

func TestHardBreak(t *testing.T) {
	out := testutil.Render(t, markdown.New("line one  \nline two"))
	if !strings.Contains(out, "<br") {
		t.Fatalf("hard break must render <br>: %s", out)
	}
}

func TestUserClassMerges(t *testing.T) {
	out := testutil.Render(t, markdown.New("hi", markdown.Class("text-base")))
	tree := testutil.NewTree(t, out)
	root := tree.One("markdown")
	cls := dom.GetAttr(root, "class")
	if !strings.Contains(cls, "text-base") || strings.Contains(cls, "text-sm") {
		t.Fatalf("user class must win over root text-sm: %q", cls)
	}
}

func TestEmptySource(t *testing.T) {
	out := testutil.Render(t, markdown.New(""))
	tree := testutil.NewTree(t, out)
	root := tree.One("markdown")
	if root.FirstChild != nil {
		t.Fatalf("empty source must render an empty root: %s", out)
	}
}

func TestChildBlockNotInjected(t *testing.T) {
	out := testutil.Render(t, testutil.WithChildren(markdown.New("# hi"), testutil.Text("INJECTED")))
	if strings.Contains(out, "INJECTED") {
		t.Fatalf("child block must not leak into markdown output: %s", out)
	}
}
