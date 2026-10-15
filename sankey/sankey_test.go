package sankey_test

import (
	"errors"
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/sankey"
)

// funnel is the worked example: two sources joining, then a split, then a
// split of that split - enough shape to exercise columns, sinks and the
// shared scale.
func funnel(extra ...sankey.Option) []sankey.Option {
	return append([]sankey.Option{
		sankey.Flow("Search", "Landing", 1200),
		sankey.Flow("Direct", "Landing", 600),
		sankey.Flow("Landing", "Signup", 1100),
		sankey.Flow("Landing", "Bounce", 700),
		sankey.Flow("Signup", "Paid", 420),
		sankey.Flow("Signup", "Free", 680),
	}, extra...)
}

func TestSankey(t *testing.T) {
	out := testutil.Render(t, sankey.New(funnel(sankey.Title("Signup funnel"))...))

	for _, want := range []string{
		`data-ui="sankey"`,
		`data-ui="sankey-svg"`,
		`data-ui="sankey-link"`,
		`data-ui="sankey-node"`,
		`data-ui="sankey-label"`,
		`role="img"`,
		`aria-label="Signup funnel"`,
		`viewBox="0 0 640 320"`,
		`<linearGradient`, `fill-opacity="0.15"`, // a ribbon blends source hue into target, at wash strength
		`>Search<`, `>Landing<`, `>Paid<`, // labels, including an inner column
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}

	// Like chart, the accessible name lives on aria-label only - a <title>
	// would surface as a native hover tooltip over our own.
	if strings.Contains(out, "<title") {
		t.Error("sankey should not emit <title>; the accessible name is on aria-label")
	}

	tree := testutil.NewTree(t, out)
	if got := len(dom.FindAll(tree.Root, dom.ByMarker("sankey-link"))); got != 6 {
		t.Errorf("expected 6 ribbons, got %d", got)
	}
	if got := len(dom.FindAll(tree.Root, dom.ByMarker("sankey-node"))); got != 7 {
		t.Errorf("expected 7 bars, got %d", got)
	}

	// Hover values reuse the ordinary tooltip component in an HTML overlay:
	// one per ribbon and one per bar, positioned in percentages.
	tips := dom.FindAll(tree.Root, dom.ByMarker("tooltip-content"))
	if len(tips) != 13 {
		t.Fatalf("expected 13 tooltips (6 ribbons + 7 bars), got %d", len(tips))
	}
	if got := text(tips[0]); got != "Search → Landing: 1.2k" {
		t.Errorf("ribbon tooltip = %q", got)
	}
	if got := text(tips[6]); got != "Search: 1.2k" {
		t.Errorf("bar tooltip = %q", got)
	}
	anchors := dom.FindAll(tree.Root, dom.ByMarker("tooltip"))
	if got := dom.GetAttr(anchors[0], "style"); !strings.Contains(got, "left: ") || !strings.Contains(got, "%") {
		t.Errorf("tooltip anchor not percentage-positioned: %q", got)
	}
}

// Without a title the node labels stand in, so the drawing is never
// announced as an unnamed image.
func TestAccessibleNameFallsBackToLabels(t *testing.T) {
	out := testutil.Render(t, sankey.New(sankey.Flow("a", "b", 1)))
	if !strings.Contains(out, `aria-label="a, b"`) {
		t.Errorf("missing label-derived accessible name: %s", out)
	}
}

// A node's column is its longest path from a source, and a node nothing
// leaves is moved out to the last column rather than stopping mid-drawing.
func TestColumnsAndSinks(t *testing.T) {
	bars := barsByLabel(t, funnel()...)

	cols := map[string]float64{}
	for label, b := range bars {
		cols[label] = b.x
	}
	for _, pair := range [][2]string{{"Search", "Direct"}, {"Paid", "Free"}} {
		if cols[pair[0]] != cols[pair[1]] {
			t.Errorf("%s and %s should share a column: %v vs %v", pair[0], pair[1], cols[pair[0]], cols[pair[1]])
		}
	}
	if !(cols["Search"] < cols["Landing"] && cols["Landing"] < cols["Signup"] && cols["Signup"] < cols["Paid"]) {
		t.Errorf("columns out of order: %v", cols)
	}
	// Bounce leaves Landing, one column in, but nothing leaves Bounce.
	if cols["Bounce"] != cols["Paid"] {
		t.Errorf("sink Bounce not pushed to the last column: %v vs %v", cols["Bounce"], cols["Paid"])
	}
}

// One value-to-pixel scale serves every column, so a bar's height and a
// ribbon's thickness mean the same thing wherever they are read.
func TestOneScaleForEveryColumn(t *testing.T) {
	bars := barsByLabel(t, funnel()...)
	totals := map[string]float64{
		"Search": 1200, "Direct": 600, "Landing": 1800,
		"Signup": 1100, "Bounce": 700, "Paid": 420, "Free": 680,
	}

	var want float64
	for label, b := range bars {
		ky := b.h / totals[label]
		if want == 0 {
			want = ky
		}
		if math.Abs(ky-want) > 0.001 {
			t.Errorf("%s scales at %v, others at %v", label, ky, want)
		}
	}
}

// Bars in a column never overlap, and none is driven off the stage.
func TestBarsStayApartAndOnStage(t *testing.T) {
	// A column of many small nodes beside one large one is the case that
	// makes the gaps fight the scale.
	options := []sankey.Option{sankey.Flow("in", "hub", 100)}
	for _, id := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		options = append(options, sankey.Flow("hub", id, 12.5))
	}
	bars := barsByLabel(t, options...)

	byColumn := map[float64][]bar{}
	for _, b := range bars {
		byColumn[b.x] = append(byColumn[b.x], b)
	}
	for x, col := range byColumn {
		for i, a := range col {
			if a.y < 0 || a.y+a.h > 320 {
				t.Errorf("bar at x=%v runs off the stage: y=%v h=%v", x, a.y, a.h)
			}
			for _, b := range col[i+1:] {
				if a.y < b.y+b.h && b.y < a.y+a.h {
					t.Errorf("bars overlap in column x=%v: %v and %v", x, a, b)
				}
			}
		}
	}
}

// Ribbons leave and arrive inside the bar they attach to: a flow drawn off
// the end of its own node is the layout bug this catches.
func TestRibbonsAttachWithinTheirBars(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, sankey.New(funnel()...)))

	// Bars by column, since a ribbon end only tells us which column it
	// touches - the bar it belongs to is the one whose span holds it.
	columns := map[float64][]bar{}
	for _, n := range dom.FindAll(tree.Root, dom.ByMarker("sankey-node")) {
		b := bar{
			x: num(t, dom.GetAttr(n, "x")),
			y: num(t, dom.GetAttr(n, "y")),
			h: num(t, dom.GetAttr(n, "height")),
		}
		columns[b.x] = append(columns[b.x], b)
	}
	within := func(x, y0, y1 float64) bool {
		for _, b := range columns[x] {
			if y0 >= b.y-0.01 && y1 <= b.y+b.h+0.01 {
				return true
			}
		}
		return false
	}

	for _, p := range dom.FindAll(tree.Root, dom.ByMarker("sankey-link")) {
		x0, sy, x1, ty, w := ribbon(t, dom.GetAttr(p, "d"))
		// A ribbon leaves a bar's right edge and arrives at another's left.
		if !within(x0-nodeWidth, sy, sy+w) {
			t.Errorf("ribbon leaves [%v,%v] at x=%v, off every bar there", sy, sy+w, x0)
		}
		if !within(x1, ty, ty+w) {
			t.Errorf("ribbon arrives [%v,%v] at x=%v, off every bar there", ty, ty+w, x1)
		}
	}
}

// Every hue reaches the drawing the same way: a node's bar wears it at
// mark strength, and a ribbon runs from its source node's hue into its
// target's at wash strength - unless the ribbon names a hue of its own,
// which it then wears end to end.
func TestColors(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, sankey.New(
		sankey.Flow("a", "b", 10),
		sankey.Flow("a", "c", 10, sankey.Rose),
		sankey.Node("a", sankey.Teal),
	)))

	bar := dom.FindAll(tree.Root, dom.ByMarker("sankey-node"))[0]
	if !hasClass(bar, "fill-teal-500") {
		t.Errorf("the bar is not a solid teal: %q", dom.GetAttr(bar, "class"))
	}
	// The ring is what reads as space between a bar and the ribbons
	// meeting it - chart puts the same one around a data point.
	if !hasClass(bar, "stroke-white") || !hasClass(bar, "dark:stroke-base-800") {
		t.Errorf("the bar has no surface ring: %q", dom.GetAttr(bar, "class"))
	}

	links := dom.FindAll(tree.Root, dom.ByMarker("sankey-link"))
	if got, want := ribbonStops(t, tree, links[0]), []string{"teal", "indigo"}; !slices.Equal(got, want) {
		t.Errorf("ribbon runs %v, want %v", got, want)
	}
	if got, want := ribbonStops(t, tree, links[1]), []string{"rose", "rose"}; !slices.Equal(got, want) {
		t.Errorf("the recolored ribbon runs %v, want it rose end to end", got)
	}

	// The accent is a theme token, not a palette hue, and must still paint.
	accent := testutil.Render(t, sankey.New(sankey.Flow("a", "b", 1), sankey.Node("a", sankey.Accent)))
	if !strings.Contains(accent, "[stop-color:var(--color-accent)]") {
		t.Errorf("accent node's ribbon is not accent: %s", accent)
	}
}

// Ribbons sharing a hue pair share one definition, and every ribbon's
// fill resolves to a gradient that is actually in the defs.
func TestGradientsAreSharedAndResolve(t *testing.T) {
	// Two flows into one node from two nodes of the same hue: one pair,
	// so one gradient.
	shared := testutil.NewTree(t, testutil.Render(t, sankey.New(
		sankey.Flow("a", "c", 1), sankey.Flow("b", "c", 1),
		sankey.Node("a", sankey.Teal), sankey.Node("b", sankey.Teal),
	)))
	if got := len(findGradients(shared.Root)); got != 1 {
		t.Errorf("two ribbons of one hue pair built %d gradients, want 1", got)
	}

	tree := testutil.NewTree(t, testutil.Render(t, sankey.New(funnel()...)))
	defined := map[string]bool{}
	for _, g := range findGradients(tree.Root) {
		defined[dom.GetAttr(g, "id")] = true
	}
	for _, link := range dom.FindAll(tree.Root, dom.ByMarker("sankey-link")) {
		fill := dom.GetAttr(link, "fill")
		id := strings.TrimSuffix(strings.TrimPrefix(fill, "url(#"), ")")
		if !defined[id] {
			t.Errorf("ribbon fills with %q, which no gradient defines", fill)
		}
	}
}

// ribbonStops names the hues at each end of a ribbon's gradient.
func ribbonStops(t *testing.T, tree *testutil.Tree, link *html.Node) []string {
	t.Helper()
	id := strings.TrimSuffix(strings.TrimPrefix(dom.GetAttr(link, "fill"), "url(#"), ")")
	for _, g := range findGradients(tree.Root) {
		if dom.GetAttr(g, "id") != id {
			continue
		}
		var hues []string
		for c := g.FirstChild; c != nil; c = c.NextSibling {
			hues = append(hues, hueOf(dom.GetAttr(c, "class")))
		}
		return hues
	}
	t.Fatalf("no gradient %q in defs", id)
	return nil
}

// findGradients walks for gradient elements by name, which the HTML parser
// has lowercased on the way back in.
func findGradients(root *html.Node) []*html.Node {
	var out []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "linearGradient") {
			out = append(out, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(root)
	return out
}

// hueOf picks the hue out of a stop's utility, e.g.
// "[stop-color:var(--color-teal-500)]" -> "teal".
func hueOf(class string) string {
	_, rest, ok := strings.Cut(class, "var(--color-")
	if !ok {
		return class
	}
	name, _, _ := strings.Cut(rest, ")")
	hue, _, _ := strings.Cut(name, "-")
	return hue
}

func hasClass(n *html.Node, want string) bool {
	return slices.Contains(strings.Fields(dom.GetAttr(n, "class")), want)
}

// Label renames a node; on a flow it replaces the text its tooltip reveals.
func TestLabelsAndValues(t *testing.T) {
	out := testutil.Render(t, sankey.New(
		sankey.Totals(),
		sankey.Flow("a", "b", 1500, sankey.Label("most of them")),
		sankey.Node("a", sankey.Label("Visitors")),
	))
	if !strings.Contains(out, ">Visitors 1.5k<") {
		t.Errorf("relabelled node missing its name or total: %s", out)
	}
	tree := testutil.NewTree(t, out)
	if got := text(dom.FindAll(tree.Root, dom.ByMarker("tooltip-content"))[0]); got != "most of them" {
		t.Errorf("flow tooltip = %q, want the flow's own label", got)
	}
}

func TestFormat(t *testing.T) {
	out := testutil.Render(t, sankey.New(
		sankey.Format(func(v float64) string { return "€" + strconv.FormatFloat(v, 'f', 0, 64) }),
		sankey.Totals(),
		sankey.Flow("a", "b", 1500),
	))
	if !strings.Contains(out, ">a €1500<") {
		t.Errorf("Format not applied to labels: %s", out)
	}
}

func TestFailsLoudly(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  error
		want error
	}{
		{"no flows", testutil.RenderErr(sankey.New()), sankey.ErrNoFlows},
		{"self loop", testutil.RenderErr(sankey.New(sankey.Flow("a", "a", 1))), sankey.ErrSelfLoop},
		{"zero value", testutil.RenderErr(sankey.New(sankey.Flow("a", "b", 0))), sankey.ErrNonPositiveValue},
		{"negative value", testutil.RenderErr(sankey.New(sankey.Flow("a", "b", -3))), sankey.ErrNonPositiveValue},
		{
			"cycle",
			testutil.RenderErr(sankey.New(
				sankey.Flow("a", "b", 1), sankey.Flow("b", "c", 1), sankey.Flow("c", "a", 1),
			)),
			sankey.ErrCycle,
		},
		{
			"node nothing flows through",
			testutil.RenderErr(sankey.New(sankey.Flow("a", "b", 1), sankey.Node("typo"))),
			sankey.ErrUnknownNode,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.got, tc.want) {
				t.Fatalf("got %v, want %v", tc.got, tc.want)
			}
		})
	}
}

func TestGolden(t *testing.T) {
	testutil.Golden(t, "sankey-funnel", sankey.New(funnel(sankey.Title("Signup funnel"))...))
}

// nodeWidth is the default bar thickness, which the geometry assertions
// need to step from one side of a bar to the other.
const nodeWidth = 7

// bar is one rendered node bar, in viewBox units.
type bar struct{ x, y, h float64 }

// barsByLabel pairs each bar with its label by vertical center, which is
// where the label is drawn.
func barsByLabel(t *testing.T, options ...sankey.Option) map[string]bar {
	t.Helper()
	tree := testutil.NewTree(t, testutil.Render(t, sankey.New(options...)))

	bars := make(map[string]bar)
	labels := dom.FindAll(tree.Root, dom.ByMarker("sankey-label"))
	for i, n := range dom.FindAll(tree.Root, dom.ByMarker("sankey-node")) {
		b := bar{
			x: num(t, dom.GetAttr(n, "x")),
			y: num(t, dom.GetAttr(n, "y")),
			h: num(t, dom.GetAttr(n, "height")),
		}
		// Bars and labels are emitted in the same order.
		bars[text(labels[i])] = b
	}
	return bars
}

// ribbon reads back a flow's outline: where it leaves, where it arrives,
// and how thick it is.
func ribbon(t *testing.T, d string) (x0, sy, x1, ty, w float64) {
	t.Helper()
	f := strings.FieldsFunc(d, func(r rune) bool { return r == ' ' || r == ',' })
	// M x0 sy C xm sy xm ty x1 ty L x1 ty1 …
	if len(f) < 13 || f[0] != "M" || f[10] != "L" {
		t.Fatalf("unrecognized ribbon path: %q", d)
	}
	return num(t, f[1]), num(t, f[2]), num(t, f[8]), num(t, f[9]), num(t, f[12]) - num(t, f[9])
}

func num(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatalf("not a number: %q", s)
	}
	return v
}

func text(n *html.Node) string {
	if n == nil || n.FirstChild == nil {
		return ""
	}
	return n.FirstChild.Data
}
