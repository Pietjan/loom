package diagram_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/pietjan/loom/diagram"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

// node builds a text-bodied node component.
func node(id, text string, opts ...diagram.NodeOption) templ.Component {
	return testutil.WithChildren(diagram.Node(id, opts...), testutil.Text(text))
}

// diag renders a diagram with the given node children and options.
func diag(t *testing.T, nodes []templ.Component, opts ...diagram.Option) string {
	t.Helper()
	return testutil.Render(t, testutil.WithChildren(diagram.New(opts...), testutil.Sequence(nodes...)))
}

func TestFlowchartBasics(t *testing.T) {
	out := diag(t,
		[]templ.Component{node("a", "Start"), node("b", "Verify"), node("c", "Done")},
		diagram.Title("Signup flow"),
		diagram.Edge("a", "b"),
		diagram.Edge("b", "c"),
	)

	for _, want := range []string{
		`data-ui="diagram"`,
		`role="img"`,
		`aria-label="Signup flow"`,
		`viewBox="0 0 `,
		`data-ui="diagram-canvas"`,
		`data-ui="diagram-node"`,
		`data-ui="diagram-shape"`,
		`data-ui="diagram-edge"`,
		`data-ui="diagram-arrow"`,
		`>Start<`, `>Verify<`, `>Done<`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}

	// The helper attrs a Node stashes for the post-pass must not leak.
	if strings.Contains(out, "data-node-id") {
		t.Errorf("internal data-node-* attribute leaked into output")
	}
	if strings.Contains(out, "<title") {
		t.Errorf("diagram should not emit <title>")
	}

	tree := testutil.NewTree(t, out)
	if n := len(dom.FindAll(tree.Root, dom.ByMarker("diagram-node"))); n != 3 {
		t.Errorf("nodes = %d, want 3", n)
	}
	if n := len(dom.FindAll(tree.Root, dom.ByMarker("diagram-edge"))); n != 2 {
		t.Errorf("edges = %d, want 2", n)
	}
	if n := len(dom.FindAll(tree.Root, dom.ByMarker("diagram-arrow"))); n != 2 {
		t.Errorf("arrows = %d, want 2", n)
	}
}

// TestRichBody: a node body may be any markup; it renders inside the
// foreignObject untouched.
func TestRichBody(t *testing.T) {
	out := diag(t, []templ.Component{
		testutil.WithChildren(diagram.Node("a"), testutil.Text("Bold")),
		node("b", "Plain"),
	}, diagram.Edge("a", "b"))

	if !strings.Contains(out, ">Bold<") {
		t.Errorf("node body not rendered: %s", out)
	}
	// Bodies must be plain HTML, never wrapped in a foreignObject (whose
	// overflow handling clips content in some browsers).
	if strings.Contains(out, "foreignObject") {
		t.Errorf("node bodies must not render inside a foreignObject")
	}
}

// TestLinearFlowsDownward: in a top-bottom chain each node sits strictly below
// the previous one.
func TestLinearFlowsDownward(t *testing.T) {
	out := diag(t,
		[]templ.Component{node("a", "A"), node("b", "B"), node("c", "C")},
		diagram.Edge("a", "b"), diagram.Edge("b", "c"),
	)
	ys := shapeYs(t, out)
	if len(ys) != 3 {
		t.Fatalf("want 3 node shapes, got %d", len(ys))
	}
	if !(ys[0] < ys[1] && ys[1] < ys[2]) {
		t.Errorf("linear chain not top-to-bottom: %v", ys)
	}
}

// TestLeftRightFlow: the same chain is wider than tall left-to-right, taller
// than wide top-to-bottom.
func TestLeftRightFlow(t *testing.T) {
	nodes := []templ.Component{node("a", "A"), node("b", "B"), node("c", "C")}
	edges := []diagram.Option{diagram.Edge("a", "b"), diagram.Edge("b", "c")}
	tb := viewBox(t, diag(t, nodes, edges...))
	lr := viewBox(t, diag(t, nodes, append(edges, diagram.Dir(diagram.LeftRight))...))

	if !(tb.h > tb.w) {
		t.Errorf("top-bottom chain should be taller than wide, got %vx%v", tb.w, tb.h)
	}
	if !(lr.w > lr.h) {
		t.Errorf("left-right chain should be wider than tall, got %vx%v", lr.w, lr.h)
	}
}

// TestBranchSpreads: a branch puts siblings on the same layer at different
// cross positions.
func TestBranchSpreads(t *testing.T) {
	out := diag(t,
		[]templ.Component{node("a", "A"), node("b", "B"), node("c", "C"), node("d", "D")},
		diagram.Edge("a", "b"), diagram.Edge("a", "c"),
		diagram.Edge("b", "d"), diagram.Edge("c", "d"),
	)
	xs := shapeXs(t, out)
	if xs[1] == xs[2] {
		t.Errorf("branch nodes overlap at x=%v", xs[1])
	}
}

// TestCycleTerminates: a cycle lays out (via cycle-breaking) and still draws an
// arrow per edge.
func TestCycleTerminates(t *testing.T) {
	out := diag(t,
		[]templ.Component{node("a", "A"), node("b", "B"), node("c", "C")},
		diagram.Edge("a", "b"), diagram.Edge("b", "c"), diagram.Edge("c", "a"),
	)
	tree := testutil.NewTree(t, out)
	if n := len(dom.FindAll(tree.Root, dom.ByMarker("diagram-arrow"))); n != 3 {
		t.Errorf("arrows = %d, want 3", n)
	}
}

// TestBareOmitsChrome: Bare() suppresses the shape behind the body.
func TestBareOmitsChrome(t *testing.T) {
	plain := diag(t, []templ.Component{node("a", "A"), node("b", "B")}, diagram.Edge("a", "b"))
	bare := diag(t, []templ.Component{
		node("a", "A", diagram.Bare()),
		node("b", "B", diagram.Bare()),
	}, diagram.Edge("a", "b"))

	if !strings.Contains(plain, "<rect") {
		t.Errorf("default node should draw a rect")
	}
	if strings.Contains(bare, "<rect") {
		t.Errorf("bare node should not draw a rect: %s", bare)
	}
}

// TestMeasuresTailwindBoxModel: a bare node's box is computed from the
// Tailwind classes on its body — padding, border and monospace advance are all
// known values, so the result is exact rather than estimated.
func TestMeasuresTailwindBoxModel(t *testing.T) {
	// 10 mono chars at 11px (0.6em advance) = 66, + px-2.5 (20) + border (2).
	body := `<div class="border px-2.5 py-1 font-mono text-[11px]">abcdefghij</div>`
	out := diag(t, []templ.Component{
		testutil.WithChildren(diagram.Node("a", diagram.Bare()), testutil.Text(body)),
		node("b", "B"),
	}, diagram.Edge("a", "b"))

	w := nodeStyleVal(t, out, 0, "width")
	if w != 88 {
		t.Errorf("measured width = %v, want 88 (66 text + 20 padding + 2 border)", w)
	}
}

// TestMonospaceScalesExactly: monospace advance is a constant, so doubling the
// characters doubles the text contribution.
func TestMonospaceScalesExactly(t *testing.T) {
	box := func(text string) float64 {
		body := `<div class="font-mono text-[11px]">` + text + `</div>`
		out := diag(t, []templ.Component{
			testutil.WithChildren(diagram.Node("a", diagram.Bare()), testutil.Text(body)),
			node("b", "B"),
		}, diagram.Edge("a", "b"))
		return nodeStyleVal(t, out, 0, "width")
	}
	five, ten := box("aaaaa"), box("aaaaaaaaaa")
	if ten-five != five {
		t.Errorf("mono width not linear: 5 chars = %v, 10 chars = %v", five, ten)
	}
}

// TestEdgeLabelsAreHoverDots: a labelled edge gets a dot on the line whose
// label is revealed by loom's CSS-only tooltip, not printed onto the diagram.
func TestEdgeLabelsAreHoverDots(t *testing.T) {
	out := diag(t,
		[]templ.Component{node("a", "A"), node("b", "B"), node("c", "C")},
		diagram.Edge("a", "b", diagram.Label("yes")),
		diagram.Edge("a", "c"), // unlabelled: no dot
	)
	tree := testutil.NewTree(t, out)

	dots := dom.FindAll(tree.Root, dom.ByMarker("diagram-edge-dot"))
	if len(dots) != 1 {
		t.Fatalf("dots = %d, want 1 (only the labelled edge)", len(dots))
	}
	// Reachable by keyboard, and described by the tooltip it opens.
	if dom.GetAttr(dots[0], "tabindex") != "0" {
		t.Error("dot must be focusable so the tooltip's focus path works")
	}
	if dom.GetAttr(dots[0], "aria-describedby") == "" {
		t.Error("dot should be described by its tooltip")
	}

	tips := dom.FindAll(tree.Root, dom.ByMarker("tooltip-content"))
	if len(tips) != 1 || tips[0].FirstChild == nil || tips[0].FirstChild.Data != "yes" {
		t.Errorf("tooltip should carry the label, got %v", tips)
	}
	// The label must not also be painted onto the canvas.
	if strings.Contains(out, "diagram-edge-label") {
		t.Error("label should not be drawn as SVG text any more")
	}
}

// TestSizeOverride: an explicit Size wins over inference.
func TestSizeOverride(t *testing.T) {
	out := diag(t, []templ.Component{
		node("a", "x", diagram.Size(240, 120)),
		node("b", "y"),
	}, diagram.Edge("a", "b"))
	tree := testutil.NewTree(t, out)
	found := false
	for _, s := range dom.FindAll(tree.Root, dom.ByMarker("diagram-shape")) {
		if dom.GetAttr(s, "width") == "240" && dom.GetAttr(s, "height") == "120" {
			found = true
		}
	}
	if !found {
		t.Errorf("explicit Size(240,120) not applied")
	}
}

// TestDeterministic: identical inputs render byte-identical output.
func TestDeterministic(t *testing.T) {
	build := func() string {
		return diag(t,
			[]templ.Component{node("a", "A"), node("b", "B"), node("c", "C"), node("d", "D")},
			diagram.Edge("a", "b"), diagram.Edge("a", "c"),
			diagram.Edge("b", "d"), diagram.Edge("c", "d"),
		)
	}
	if build() != build() {
		t.Error("diagram output is not deterministic")
	}
}

func TestFailsLoudly(t *testing.T) {
	if err := testutil.RenderErr(diagram.New()); !errors.Is(err, diagram.ErrNoNodes) {
		t.Fatalf("expected ErrNoNodes, got %v", err)
	}
	unknown := testutil.WithChildren(
		diagram.New(diagram.Edge("a", "ghost")),
		node("a", "A"),
	)
	if err := testutil.RenderErr(unknown); err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected unknown-node error naming ghost, got %v", err)
	}
	dup := testutil.WithChildren(
		diagram.New(),
		testutil.Sequence(node("a", "A"), node("a", "again")),
	)
	if err := testutil.RenderErr(dup); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestGolden(t *testing.T) {
	g := func(name string, nodes []templ.Component, opts ...diagram.Option) {
		testutil.Golden(t, name, testutil.WithChildren(diagram.New(opts...), testutil.Sequence(nodes...)))
	}
	g("diagram-linear",
		[]templ.Component{node("a", "Start"), node("b", "Middle"), node("c", "End")},
		diagram.Edge("a", "b"), diagram.Edge("b", "c"))
	g("diagram-branch",
		[]templ.Component{node("a", "Open"), node("b", "Review"), node("c", "Test"), node("d", "Merge")},
		diagram.Edge("a", "b"), diagram.Edge("a", "c"),
		diagram.Edge("b", "d"), diagram.Edge("c", "d"))
	// A tree exercises parent-centering: each parent should settle over the
	// midpoint of its children, which no other golden covers.
	g("diagram-tree",
		[]templ.Component{
			node("root", "Root"), node("l", "Left"), node("r", "Right"),
			node("ll", "L1"), node("lr", "L2"),
		},
		diagram.Edge("root", "l"), diagram.Edge("root", "r"),
		diagram.Edge("l", "ll"), diagram.Edge("l", "lr"))
	g("diagram-decision",
		[]templ.Component{
			node("start", "Start", diagram.Stadium()),
			node("ok", "OK?", diagram.Diamond(), diagram.WithTone(diagram.ToneAccent)),
			node("yes", "Ship", diagram.WithTone(diagram.ToneEmerald)),
			node("no", "Fix", diagram.WithTone(diagram.ToneRose)),
		},
		diagram.Edge("start", "ok"),
		diagram.Edge("ok", "yes", diagram.Label("yes")),
		diagram.Edge("ok", "no", diagram.Label("no")),
		diagram.Edge("no", "ok"))
	g("diagram-lr",
		[]templ.Component{node("a", "Build"), node("b", "Test"), node("c", "Deploy")},
		diagram.Dir(diagram.LeftRight),
		diagram.Edge("a", "b"), diagram.Edge("b", "c"))
	g("diagram-cycle",
		[]templ.Component{node("a", "A"), node("b", "B"), node("c", "C")},
		diagram.Edge("a", "b"), diagram.Edge("b", "c"), diagram.Edge("c", "a"))
}

// --- helpers ---

type dims struct{ w, h float64 }

func viewBox(t *testing.T, out string) dims {
	t.Helper()
	svg := testutil.NewTree(t, out).One("diagram-canvas")
	f := strings.Fields(dom.GetAttr(svg, "viewBox"))
	if len(f) != 4 {
		t.Fatalf("bad viewBox %q", dom.GetAttr(svg, "viewBox"))
	}
	return dims{w: atof(t, f[2]), h: atof(t, f[3])}
}

func shapeYs(t *testing.T, out string) []float64 { return shapeAttr(t, out, "y") }
func shapeXs(t *testing.T, out string) []float64 { return shapeAttr(t, out, "x") }

// shapeAttr reads the x/y of each node's drawn rect chrome, in document order.
func shapeAttr(t *testing.T, out, attr string) []float64 {
	t.Helper()
	tree := testutil.NewTree(t, out)
	var vals []float64
	for _, s := range dom.FindAll(tree.Root, dom.ByMarker("diagram-shape")) {
		if s.Data == "rect" {
			vals = append(vals, atof(t, dom.GetAttr(s, attr)))
		}
	}
	return vals
}

// nodeStyleVal reads a numeric px value out of the nth node body's style.
func nodeStyleVal(t *testing.T, out string, n int, prop string) float64 {
	t.Helper()
	nodes := dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-node"))
	if n >= len(nodes) {
		t.Fatalf("node %d out of range (%d nodes)", n, len(nodes))
	}
	for _, decl := range strings.Split(dom.GetAttr(nodes[n], "style"), ";") {
		k, v, ok := strings.Cut(decl, ":")
		if ok && strings.TrimSpace(k) == prop {
			return atof(t, strings.TrimSuffix(strings.TrimSpace(v), "px"))
		}
	}
	t.Fatalf("no %s in style %q", prop, dom.GetAttr(nodes[n], "style"))
	return 0
}

func atof(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}
