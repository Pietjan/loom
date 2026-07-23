package diagram_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/pietjan/loom/diagram"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

func TestFlowchartBasics(t *testing.T) {
	out := testutil.Render(t, diagram.New(
		diagram.Title("Signup flow"),
		diagram.Node("a", "Start"),
		diagram.Node("b", "Verify"),
		diagram.Node("c", "Done"),
		diagram.Edge("a", "b"),
		diagram.Edge("b", "c"),
	))

	for _, want := range []string{
		`data-ui="diagram"`,
		`role="img"`,
		`aria-label="Signup flow"`,
		`viewBox="0 0 `,
		`data-ui="diagram-node"`,
		`data-ui="diagram-edge"`,
		`data-ui="diagram-arrow"`,
		`>Start<`, `>Verify<`, `>Done<`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}

	// The accessible name is on aria-label only — a <title> would show as a
	// native hover tooltip (same rationale as chart).
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

// TestLinearFlowsDownward checks the core layout invariant: in a top-bottom
// chain, each node sits strictly below the previous one (increasing y).
func TestLinearFlowsDownward(t *testing.T) {
	out := testutil.Render(t, diagram.New(
		diagram.Node("a", "A"),
		diagram.Node("b", "B"),
		diagram.Node("c", "C"),
		diagram.Edge("a", "b"),
		diagram.Edge("b", "c"),
	))
	ys := rectYs(t, out)
	if len(ys) != 3 {
		t.Fatalf("want 3 node rects, got %d", len(ys))
	}
	if !(ys[0] < ys[1] && ys[1] < ys[2]) {
		t.Errorf("linear chain not top-to-bottom: %v", ys)
	}
}

// TestLeftRightFlow: the same chain laid left-to-right is wider than tall,
// where top-bottom is taller than wide.
func TestLeftRightFlow(t *testing.T) {
	nodes := []diagram.Option{
		diagram.Node("a", "A"), diagram.Node("b", "B"), diagram.Node("c", "C"),
		diagram.Edge("a", "b"), diagram.Edge("b", "c"),
	}
	tb := viewBox(t, testutil.Render(t, diagram.New(nodes...)))
	lr := viewBox(t, testutil.Render(t, diagram.New(append(nodes, diagram.Dir(diagram.LeftRight))...)))

	if !(tb.h > tb.w) {
		t.Errorf("top-bottom chain should be taller than wide, got %vx%v", tb.w, tb.h)
	}
	if !(lr.w > lr.h) {
		t.Errorf("left-right chain should be wider than tall, got %vx%v", lr.w, lr.h)
	}
}

// TestBranchSpreads: a branch (A→B, A→C) puts B and C on the same layer but at
// different cross positions, so their boxes don't coincide.
func TestBranchSpreads(t *testing.T) {
	out := testutil.Render(t, diagram.New(
		diagram.Node("a", "A"),
		diagram.Node("b", "B"),
		diagram.Node("c", "C"),
		diagram.Node("d", "D"),
		diagram.Edge("a", "b"),
		diagram.Edge("a", "c"),
		diagram.Edge("b", "d"),
		diagram.Edge("c", "d"),
	))
	xs := rectXs(t, out)
	// B and C are the 2nd and 3rd rects; they must have different x.
	if xs[1] == xs[2] {
		t.Errorf("branch nodes overlap at x=%v", xs[1])
	}
}

// TestCycleTerminates: a cycle must lay out (cycle-breaking) rather than
// recurse forever, and still draw an arrow per edge.
func TestCycleTerminates(t *testing.T) {
	out := testutil.Render(t, diagram.New(
		diagram.Node("a", "A"),
		diagram.Node("b", "B"),
		diagram.Node("c", "C"),
		diagram.Edge("a", "b"),
		diagram.Edge("b", "c"),
		diagram.Edge("c", "a"),
	))
	tree := testutil.NewTree(t, out)
	if n := len(dom.FindAll(tree.Root, dom.ByMarker("diagram-arrow"))); n != 3 {
		t.Errorf("arrows = %d, want 3", n)
	}
}

// TestDeterministic: identical configs render byte-identical output (guards
// against accidental map iteration in the layout).
func TestDeterministic(t *testing.T) {
	build := func() string {
		return testutil.Render(t, diagram.New(
			diagram.Node("a", "A"), diagram.Node("b", "B"), diagram.Node("c", "C"), diagram.Node("d", "D"),
			diagram.Edge("a", "b"), diagram.Edge("a", "c"), diagram.Edge("b", "d"), diagram.Edge("c", "d"),
		))
	}
	if build() != build() {
		t.Error("diagram output is not deterministic")
	}
}

func TestFailsLoudly(t *testing.T) {
	if err := testutil.RenderErr(diagram.New()); !errors.Is(err, diagram.ErrNoNodes) {
		t.Fatalf("expected ErrNoNodes, got %v", err)
	}
	err := testutil.RenderErr(diagram.New(
		diagram.Node("a", "A"),
		diagram.Edge("a", "ghost"),
	))
	if err == nil || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("expected unknown-node error naming ghost, got %v", err)
	}
	err = testutil.RenderErr(diagram.New(
		diagram.Node("a", "A"),
		diagram.Node("a", "again"),
	))
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate id error, got %v", err)
	}
}

func TestGolden(t *testing.T) {
	testutil.Golden(t, "diagram-linear", diagram.New(
		diagram.Node("a", "Start"),
		diagram.Node("b", "Middle"),
		diagram.Node("c", "End"),
		diagram.Edge("a", "b"),
		diagram.Edge("b", "c"),
	))
	testutil.Golden(t, "diagram-branch", diagram.New(
		diagram.Node("a", "Open"),
		diagram.Node("b", "Review"),
		diagram.Node("c", "Test"),
		diagram.Node("d", "Merge"),
		diagram.Edge("a", "b"),
		diagram.Edge("a", "c"),
		diagram.Edge("b", "d"),
		diagram.Edge("c", "d"),
	))
	testutil.Golden(t, "diagram-tree", diagram.New(
		diagram.Node("root", "Root"),
		diagram.Node("l", "Left"),
		diagram.Node("r", "Right"),
		diagram.Node("ll", "L1"),
		diagram.Node("lr", "L2"),
		diagram.Edge("root", "l"),
		diagram.Edge("root", "r"),
		diagram.Edge("l", "ll"),
		diagram.Edge("l", "lr"),
	))
	testutil.Golden(t, "diagram-decision", diagram.New(
		diagram.Node("start", "Start", diagram.Stadium()),
		diagram.Node("ok", "OK?", diagram.Diamond(), diagram.WithTone(diagram.ToneAccent)),
		diagram.Node("yes", "Ship", diagram.WithTone(diagram.ToneEmerald)),
		diagram.Node("no", "Fix", diagram.WithTone(diagram.ToneRose)),
		diagram.Edge("start", "ok"),
		diagram.Edge("ok", "yes", diagram.Label("yes")),
		diagram.Edge("ok", "no", diagram.Label("no")),
		diagram.Edge("no", "ok"),
	))
	testutil.Golden(t, "diagram-lr", diagram.New(
		diagram.Dir(diagram.LeftRight),
		diagram.Node("a", "Build"),
		diagram.Node("b", "Test"),
		diagram.Node("c", "Deploy"),
		diagram.Edge("a", "b"),
		diagram.Edge("b", "c"),
	))
	testutil.Golden(t, "diagram-cycle", diagram.New(
		diagram.Node("a", "A"),
		diagram.Node("b", "B"),
		diagram.Node("c", "C"),
		diagram.Edge("a", "b"),
		diagram.Edge("b", "c"),
		diagram.Edge("c", "a"),
	))
}

// --- helpers ---

type dims struct{ w, h float64 }

func viewBox(t *testing.T, out string) dims {
	t.Helper()
	tree := testutil.NewTree(t, out)
	svg := tree.One("diagram")
	f := strings.Fields(dom.GetAttr(svg, "viewBox"))
	if len(f) != 4 {
		t.Fatalf("bad viewBox %q", dom.GetAttr(svg, "viewBox"))
	}
	return dims{w: atof(t, f[2]), h: atof(t, f[3])}
}

func rectYs(t *testing.T, out string) []float64 { return rectAttr(t, out, "y") }
func rectXs(t *testing.T, out string) []float64 { return rectAttr(t, out, "x") }

func rectAttr(t *testing.T, out, attr string) []float64 {
	t.Helper()
	tree := testutil.NewTree(t, out)
	var vals []float64
	for _, g := range dom.FindAll(tree.Root, dom.ByMarker("diagram-node")) {
		for c := g.FirstChild; c != nil; c = c.NextSibling {
			if c.Data == "rect" {
				vals = append(vals, atof(t, dom.GetAttr(c, attr)))
			}
		}
	}
	return vals
}

func atof(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return v
}
