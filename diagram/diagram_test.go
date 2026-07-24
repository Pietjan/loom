package diagram_test

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html"

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

// TestRichBody: a node body may be any markup, rendered untouched.
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

// TestParentCentersOverChildren: coordinate assignment puts a parent exactly
// on the midpoint of its children — including a root whose two children are a
// two-child subtree and a wider childless sibling, which the earlier averaging
// sweep left visibly off-center.
func TestParentCentersOverChildren(t *testing.T) {
	out := diag(t,
		[]templ.Component{
			node("root", "Root"), node("l", "Left"), node("r", "Right"),
			node("ll", "L1"), node("lr", "L2"),
		},
		diagram.Edge("root", "l"), diagram.Edge("root", "r"),
		diagram.Edge("l", "ll"), diagram.Edge("l", "lr"),
	)
	c := shapeCenters(t, out) // document order: root, l, r, ll, lr
	if len(c) != 5 {
		t.Fatalf("want 5 node shapes, got %d", len(c))
	}
	for _, tc := range []struct {
		what         string
		parent, a, b int
	}{
		{"root over Left and Right", 0, 1, 2},
		{"Left over its two children", 1, 3, 4},
	} {
		if want := (c[tc.a] + c[tc.b]) / 2; math.Abs(c[tc.parent]-want) > 0.01 {
			t.Errorf("%s: center = %v, want %v", tc.what, c[tc.parent], want)
		}
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

// TestEdgeLineStyles: Dashed and Dotted break an edge's shaft via a
// stroke-dasharray while leaving other edges (and every arrowhead) solid.
func TestEdgeLineStyles(t *testing.T) {
	out := diag(t,
		[]templ.Component{node("a", "A"), node("b", "B"), node("c", "C"), node("d", "D")},
		diagram.Edge("a", "b", diagram.Dashed()),
		diagram.Edge("a", "c", diagram.Dotted()),
		diagram.Edge("a", "d"), // solid
	)
	tree := testutil.NewTree(t, out)

	edges := dom.FindAll(tree.Root, dom.ByMarker("diagram-edge"))
	if len(edges) != 3 {
		t.Fatalf("edges = %d, want 3", len(edges))
	}
	var dashed, dotted, solid int
	for _, e := range edges {
		switch dom.GetAttr(e, "stroke-dasharray") {
		case "":
			solid++
		case "6 4":
			dashed++
			if lc := dom.GetAttr(e, "stroke-linecap"); lc != "" {
				t.Errorf("dashed edge should not set a linecap, got %q", lc)
			}
		case "0 6":
			dotted++
			if lc := dom.GetAttr(e, "stroke-linecap"); lc != "round" {
				t.Errorf("dotted edge needs round linecap for dots, got %q", lc)
			}
		default:
			t.Errorf("unexpected stroke-dasharray %q", dom.GetAttr(e, "stroke-dasharray"))
		}
	}
	if dashed != 1 || dotted != 1 || solid != 1 {
		t.Errorf("dashed=%d dotted=%d solid=%d, want 1 each", dashed, dotted, solid)
	}

	// Arrowheads stay solid regardless of the shaft's pattern.
	for _, a := range dom.FindAll(tree.Root, dom.ByMarker("diagram-arrow")) {
		if dom.GetAttr(a, "stroke-dasharray") != "" {
			t.Error("arrowhead should never be dashed")
		}
	}
}

// TestArrowEnds: NoArrow draws a plain connector, BiDirectional draws a head at
// each end, and a default edge keeps its single head at the target.
func TestArrowEnds(t *testing.T) {
	count := func(opts ...diagram.EdgeOption) int {
		out := diag(t,
			[]templ.Component{node("a", "A"), node("b", "B")},
			diagram.Edge("a", "b", opts...),
		)
		return len(dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-arrow")))
	}
	if n := count(); n != 1 {
		t.Errorf("default edge arrows = %d, want 1", n)
	}
	if n := count(diagram.NoArrow()); n != 0 {
		t.Errorf("NoArrow edge arrows = %d, want 0", n)
	}
	if n := count(diagram.BiDirectional()); n != 2 {
		t.Errorf("BiDirectional edge arrows = %d, want 2", n)
	}
}

// TestBiDirectionalHeadsPointOutward: the two heads of a bidirectional edge sit
// at opposite ends and point away from each other, not both the same way.
func TestBiDirectionalHeadsPointOutward(t *testing.T) {
	out := diag(t,
		[]templ.Component{node("a", "A"), node("b", "B")},
		diagram.Edge("a", "b", diagram.BiDirectional()),
	)
	arrows := dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-arrow"))
	if len(arrows) != 2 {
		t.Fatalf("arrows = %d, want 2", len(arrows))
	}
	// Top-bottom flow: the arrowhead's tip is its first polygon point, so the
	// two tips should sit at opposite ends of the axis, not on top of each other.
	tips := make([]float64, len(arrows))
	for i, a := range arrows {
		first := strings.Fields(dom.GetAttr(a, "points"))[0]
		_, y, _ := strings.Cut(first, ",")
		tips[i] = atof(t, y)
	}
	if math.Abs(tips[0]-tips[1]) < 1 {
		t.Errorf("bidirectional heads should sit at opposite ends, tips at %v and %v", tips[0], tips[1])
	}
}

// TestArrowShapes: each arrowhead variant renders the expected SVG element and
// fill treatment, and the default stays a filled triangle.
func TestArrowShapes(t *testing.T) {
	arrow := func(opts ...diagram.EdgeOption) *html.Node {
		out := diag(t,
			[]templ.Component{node("a", "A"), node("b", "B")},
			diagram.Edge("a", "b", opts...),
		)
		got := dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-arrow"))
		if len(got) != 1 {
			t.Fatalf("arrows = %d, want 1", len(got))
		}
		return got[0]
	}

	for _, tc := range []struct {
		name     string
		opt      diagram.EdgeOption
		tag      string // expected SVG element
		classSub string // a substring the class must contain
	}{
		{"solid", nil, "polygon", "fill-base-300"},
		{"hollow", diagram.HollowArrow(), "polygon", "fill-white"},
		{"open", diagram.OpenArrow(), "polyline", "fill-none"},
		{"diamond", diagram.DiamondArrow(), "polygon", "fill-base-300"},
		{"dot", diagram.DotArrow(), "circle", "fill-base-300"},
	} {
		var a *html.Node
		if tc.opt == nil {
			a = arrow()
		} else {
			a = arrow(tc.opt)
		}
		if a.Data != tc.tag {
			t.Errorf("%s: element = %q, want %q", tc.name, a.Data, tc.tag)
		}
		if cls := dom.GetAttr(a, "class"); !strings.Contains(cls, tc.classSub) {
			t.Errorf("%s: class %q missing %q", tc.name, cls, tc.classSub)
		}
	}

	// The diamond has four points (tip, two sides, back); the solid triangle
	// three — a quick guard that the diamond geometry is actually distinct.
	if got := len(strings.Fields(dom.GetAttr(arrow(diagram.DiamondArrow()), "points"))); got != 4 {
		t.Errorf("diamond points = %d, want 4", got)
	}
}

// TestShaftTrimming: a closed arrowhead pulls the shaft back to its base, while
// an open barb (or no head) lets the shaft run to the node border.
func TestShaftTrimming(t *testing.T) {
	end := func(opts ...diagram.EdgeOption) (x, y float64) {
		out := diag(t, []templ.Component{node("a", "A"), node("b", "B")},
			diagram.Edge("a", "b", opts...))
		edge := dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-edge"))[0]
		// The path ends with "... L x y"; grab its final two numbers.
		nums := regexp.MustCompile(`[\d.]+`).FindAllString(dom.GetAttr(edge, "d"), -1)
		return atof(t, nums[len(nums)-2]), atof(t, nums[len(nums)-1])
	}
	gap := func(a, b [2]float64) float64 { return math.Hypot(a[0]-b[0], a[1]-b[1]) }

	sx, sy := end()                         // solid triangle: trimmed a full length
	ox, oy := end(diagram.OpenArrow())      // barb: reaches the border
	nx, ny := end(diagram.NoArrow())        // no head: reaches the border
	dmx, dmy := end(diagram.DiamondArrow()) // diamond: trimmed less, overlaps in
	dtx, dty := end(diagram.DotArrow())     // dot: trimmed less, overlaps in

	if d := gap([2]float64{ox, oy}, [2]float64{sx, sy}); math.Abs(d-8) > 0.5 {
		t.Errorf("solid shaft should stop ~8px (one arrow length) short of the border, gap=%v", d)
	}
	if d := gap([2]float64{ox, oy}, [2]float64{nx, ny}); d > 0.5 {
		t.Errorf("open and no-arrow shafts should both reach the border, gap=%v", d)
	}
	// Pointed heads trim less than the triangle, so the shaft runs a touch
	// further into their bodies: 8 − 2 = 6px short of the border.
	for _, tc := range []struct {
		name string
		x, y float64
	}{{"diamond", dmx, dmy}, {"dot", dtx, dty}} {
		if d := gap([2]float64{ox, oy}, [2]float64{tc.x, tc.y}); math.Abs(d-6) > 0.5 {
			t.Errorf("%s shaft should stop ~6px short of the border, gap=%v", tc.name, d)
		}
		if d := gap([2]float64{ox, oy}, [2]float64{sx, sy}); d <= gap([2]float64{ox, oy}, [2]float64{tc.x, tc.y}) {
			t.Errorf("%s should be trimmed less than the solid triangle", tc.name)
		}
	}
}

// pathEnds returns the first and last point of an SVG path's "d" attribute.
func pathEnds(t *testing.T, d string) (first, last [2]float64) {
	t.Helper()
	n := regexp.MustCompile(`[\d.]+`).FindAllString(d, -1)
	if len(n) < 4 {
		t.Fatalf("path too short: %q", d)
	}
	return [2]float64{atof(t, n[0]), atof(t, n[1])},
		[2]float64{atof(t, n[len(n)-2]), atof(t, n[len(n)-1])}
}

// TestPortsSeparateOppositeEdges: the reported bug — a 2-cycle A⇄B anchored both
// edges at the same face point, drawing one shaft over the other's arrowhead.
// Distributed ports must give them distinct anchors on each node.
func TestPortsSeparateOppositeEdges(t *testing.T) {
	out := diag(t, []templ.Component{node("a", "A"), node("b", "B")},
		diagram.Edge("a", "b"), diagram.Edge("b", "a"))
	edges := dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-edge"))
	if len(edges) != 2 {
		t.Fatalf("edges = %d, want 2", len(edges))
	}
	// Edge 0 is A→B (starts at A); edge 1 is B→A, drawn reversed (ends at A).
	aStart, _ := pathEnds(t, dom.GetAttr(edges[0], "d"))
	_, aEnd := pathEnds(t, dom.GetAttr(edges[1], "d"))
	if math.Hypot(aStart[0]-aEnd[0], aStart[1]-aEnd[1]) < 1 {
		t.Errorf("opposite edges still share node A's anchor near %v", aStart)
	}
}

// TestPortsShareFanOut: sibling edges leaving one node go the same direction, so
// they share a single exit port and fan out from one point.
func TestPortsShareFanOut(t *testing.T) {
	out := diag(t, []templ.Component{node("a", "A"), node("b", "B"), node("c", "C")},
		diagram.Edge("a", "b"), diagram.Edge("a", "c"))
	edges := dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-edge"))
	toB, _ := pathEnds(t, dom.GetAttr(edges[0], "d")) // a→b exit
	toC, _ := pathEnds(t, dom.GetAttr(edges[1], "d")) // a→c exit
	if toB != toC {
		t.Errorf("same-direction fan-out edges should share one exit port, got %v and %v", toB, toC)
	}
}

// TestPortsSpreadOption: diagram.Ports(PortsSpread) reverts to a distinct port
// per edge, so the same fan-out no longer shares an exit.
func TestPortsSpreadOption(t *testing.T) {
	nodes := []templ.Component{node("a", "A"), node("b", "B"), node("c", "C")}
	edges := []diagram.Option{diagram.Edge("a", "b"), diagram.Edge("a", "c")}

	exits := func(opts ...diagram.Option) (b, c [2]float64) {
		out := diag(t, nodes, append(edges, opts...)...)
		e := dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-edge"))
		b, _ = pathEnds(t, dom.GetAttr(e[0], "d"))
		c, _ = pathEnds(t, dom.GetAttr(e[1], "d"))
		return b, c
	}

	if b, c := exits(); b != c { // default PortsShared
		t.Errorf("default should share the fan-out exit, got %v and %v", b, c)
	}
	if b, c := exits(diagram.Ports(diagram.PortsSpread)); b == c {
		t.Errorf("PortsSpread should give distinct exits, both at %v", b)
	}
}

// --- edge crossing detection ---

type xyPt struct{ x, y float64 }

var pathNums = regexp.MustCompile(`-?[\d.]+`)
var pathCorners = regexp.MustCompile(`Q (-?[\d.]+) (-?[\d.]+)`)

// edgePolyline reconstructs an edge's corner points from its rounded path: the
// start, each bend's control point (which is the corner itself), and the end.
func edgePolyline(t *testing.T, d string) []xyPt {
	t.Helper()
	n := pathNums.FindAllString(d, -1)
	out := []xyPt{{atof(t, n[0]), atof(t, n[1])}}
	for _, m := range pathCorners.FindAllStringSubmatch(d, -1) {
		out = append(out, xyPt{atof(t, m[1]), atof(t, m[2])})
	}
	return append(out, xyPt{atof(t, n[len(n)-2]), atof(t, n[len(n)-1])})
}

// properCrossing reports a genuine X between two segments; merely touching at a
// shared point (as edges sharing a port do) does not count.
func properCrossing(a, b, c, d xyPt) bool {
	turn := func(o, p, q xyPt) int {
		v := (p.x-o.x)*(q.y-o.y) - (p.y-o.y)*(q.x-o.x)
		switch {
		case v > 1e-9:
			return 1
		case v < -1e-9:
			return -1
		}
		return 0
	}
	return turn(a, b, c)*turn(a, b, d) < 0 && turn(c, d, a)*turn(c, d, b) < 0
}

// edgeCrossings counts places where two different edges cross.
func edgeCrossings(t *testing.T, out string) int {
	t.Helper()
	var polys [][]xyPt
	for _, e := range dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-edge")) {
		polys = append(polys, edgePolyline(t, dom.GetAttr(e, "d")))
	}
	n := 0
	for i := range polys {
		for j := i + 1; j < len(polys); j++ {
			for a := 0; a+1 < len(polys[i]); a++ {
				for b := 0; b+1 < len(polys[j]); b++ {
					if properCrossing(polys[i][a], polys[i][a+1], polys[j][b], polys[j][b+1]) {
						t.Logf("edge %d crosses edge %d near %v", i, j, polys[i][a])
						n++
					}
				}
			}
		}
	}
	return n
}

// TestAntiparallelEdgesDoNotCross: a forward edge and the back edge sharing its
// corridor must nest in their own lanes, not weave over each other. The lane
// side alone isn't enough — the back edge's elbow has to fall on the far side
// of the forward edge's, which flips with the direction the corridor runs, so
// both a leftward and a rightward pair are checked.
//
// Scoped to pairs between adjacent layers. Layered layout reduces crossings
// heuristically rather than eliminating them, and once a cycle is long enough
// to route through dummies the pair's shape is decided by the ordering phase
// and by dummy snapping too, which can still leave a weave.
func TestAntiparallelEdgesDoNotCross(t *testing.T) {
	for _, tc := range []struct {
		name  string
		nodes []templ.Component
		opts  []diagram.Option
	}{
		{"plain 2-cycle",
			[]templ.Component{node("a", "A"), node("b", "B")},
			[]diagram.Option{diagram.Edge("a", "b"), diagram.Edge("b", "a")}},
		// Target left of the source.
		{"cycle with a wider head",
			[]templ.Component{node("a", "Stopped"), node("b", "Go")},
			[]diagram.Option{diagram.Edge("a", "b"), diagram.Edge("b", "a")}},
		// Decision fanning right, with the back edge returning up the corridor.
		{"decision with back edge",
			[]templ.Component{
				node("start", "Start", diagram.Stadium()),
				node("ok", "OK?", diagram.Diamond()),
				node("yes", "Ship"), node("no", "Fix"),
			},
			[]diagram.Option{
				diagram.Edge("start", "ok"), diagram.Edge("ok", "yes", diagram.Label("yes")),
				diagram.Edge("ok", "no", diagram.Label("no")), diagram.Edge("no", "ok"),
			}},
	} {
		if n := edgeCrossings(t, diag(t, tc.nodes, tc.opts...)); n != 0 {
			t.Errorf("%s: %d edge crossing(s), want 0", tc.name, n)
		}
	}
}

// TestNoDegenerateBends: an edge never emits a "corner" too small to draw. A
// bend pulls back by its radius on both sides, so a step under twice that has
// no straight run left between the two curves and reads as a wobble; such steps
// must be flattened into a straight line instead.
func TestNoDegenerateBends(t *testing.T) {
	out := diag(t, []templ.Component{
		node("draft", "Draft", diagram.Stadium()), node("review", "In review"),
		node("check", "Approved?", diagram.Diamond()),
		node("publish", "Published"), node("revise", "Revising"),
	},
		diagram.Ports(diagram.PortsSpread),
		diagram.Edge("draft", "review"), diagram.Edge("review", "check"),
		diagram.Edge("check", "publish", diagram.Label("approved")),
		diagram.Edge("check", "revise", diagram.Label("changes")),
		diagram.Edge("revise", "review", diagram.Dashed()),
	)
	for i, e := range dom.FindAll(testutil.NewTree(t, out).Root, dom.ByMarker("diagram-edge")) {
		pts := edgePolyline(t, dom.GetAttr(e, "d"))
		for k := 0; k+1 < len(pts); k++ {
			d := math.Hypot(pts[k+1].x-pts[k].x, pts[k+1].y-pts[k].y)
			// Interior runs join two bends; the first and last meet a port.
			if k > 0 && k+2 < len(pts) && d < 12 {
				t.Errorf("edge %d: %vpx run between bends is too short to render as a corner", i, d)
			}
		}
	}
}

// TestFanPortsStaySymmetric: straightening a near-aligned edge must not drag a
// port that belongs to a fan — its branches would then sit lopsided about the
// node's centreline. Only a face's sole attachment is free to slide, so here
// the branch to Published is straightened by moving its far end instead.
func TestFanPortsStaySymmetric(t *testing.T) {
	out := diag(t, []templ.Component{
		node("draft", "Draft", diagram.Stadium()), node("review", "In review"),
		node("check", "Approved?", diagram.Diamond()),
		node("publish", "Published"), node("revise", "Revising"),
	},
		diagram.Ports(diagram.PortsSpread),
		diagram.Edge("draft", "review"), diagram.Edge("review", "check"),
		diagram.Edge("check", "publish", diagram.Label("approved")),
		diagram.Edge("check", "revise", diagram.Label("changes")),
		diagram.Edge("revise", "review", diagram.Dashed()),
	)
	tree := testutil.NewTree(t, out)

	// The decision diamond's centre is the x of its top vertex.
	var cx float64
	for _, s := range dom.FindAll(tree.Root, dom.ByMarker("diagram-shape")) {
		if s.Data == "polygon" {
			cx = atof(t, strings.Split(strings.Fields(dom.GetAttr(s, "points"))[0], ",")[0])
		}
	}
	edges := dom.FindAll(tree.Root, dom.ByMarker("diagram-edge"))
	pubStart, pubEnd := pathEnds(t, dom.GetAttr(edges[2], "d"))
	revStart, _ := pathEnds(t, dom.GetAttr(edges[3], "d"))

	if l, r := cx-pubStart[0], revStart[0]-cx; math.Abs(l-r) > 0.01 {
		t.Errorf("fan lopsided about centre %v: exits at %v and %v sit %v and %v away",
			cx, pubStart[0], revStart[0], l, r)
	}
	// ...and the straightening it gave up at that end still happened at the other.
	if pubStart[0] != pubEnd[0] {
		t.Errorf("branch to Published should run straight, got x %v → %v", pubStart[0], pubEnd[0])
	}
}

// TestGapAndMargin: spacing is configurable, and a tight layer gap clamps the
// edge corner radius rather than letting the two bends merge.
func TestGapAndMargin(t *testing.T) {
	nodes := []templ.Component{node("a", "A"), node("b", "B")}
	edges := []diagram.Option{diagram.Edge("a", "b")}

	def := viewBox(t, diag(t, nodes, edges...))
	roomy := viewBox(t, diag(t, nodes, append(edges, diagram.Gap(120, 28), diagram.Margin(40))...))

	// One layer gap + two margins grew: 80 more gap, 48 more margin.
	if got, want := roomy.h-def.h, 80.0+48.0; got != want {
		t.Errorf("height grew by %v, want %v", got, want)
	}
	if got, want := roomy.w-def.w, 48.0; got != want {
		t.Errorf("width grew by %v (margin only), want %v", got, want)
	}

	// A tight gap must not emit corners wider than the space between layers.
	tight := diag(t, []templ.Component{node("a", "A"), node("b", "B"), node("c", "C")},
		diagram.Gap(8, 28), diagram.Edge("a", "b"), diagram.Edge("a", "c"))
	for _, d := range curveRadii(t, tight) {
		if d > 2.0 {
			t.Errorf("corner radius %v exceeds a quarter of the 8px layer gap", d)
		}
	}
}

// curveRadii measures how far each rounded corner pulls back from its bend.
func curveRadii(t *testing.T, out string) []float64 {
	t.Helper()
	var radii []float64
	for _, d := range regexp.MustCompile(`d="([^"]*)"`).FindAllStringSubmatch(out, -1) {
		// "L x y Q cx cy x2 y2" — the gap between the L point and the control
		// point is the pull-back distance.
		for _, m := range regexp.MustCompile(`L ([\d.]+) ([\d.]+) Q ([\d.]+) ([\d.]+)`).FindAllStringSubmatch(d[1], -1) {
			lx, ly := atof(t, m[1]), atof(t, m[2])
			qx, qy := atof(t, m[3]), atof(t, m[4])
			radii = append(radii, math.Hypot(qx-lx, qy-ly))
		}
	}
	return radii
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

// shapeCenters reads the horizontal center of each node's rect chrome.
func shapeCenters(t *testing.T, out string) []float64 {
	t.Helper()
	xs, ws := shapeAttr(t, out, "x"), shapeAttr(t, out, "width")
	c := make([]float64, len(xs))
	for i := range xs {
		c[i] = xs[i] + ws[i]/2
	}
	return c
}

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
