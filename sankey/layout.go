package sankey

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	// defaultNodeWidth is chart's data point across: r="3.5", so a bar is
	// 7 units wide and its capsule cap has the point's own radius.
	defaultNodeWidth = 7.0
	defaultGap       = 14.0 // vertical space between two bars in a column
	labelGap         = 6.0  // space between a bar and its label
	labelPad         = 2.0  // slack in the gutter, against the width estimate
	labelSize        = 10.0 // label font size, matching style.go
	stageMargin      = 10.0 // room above and below for a label's half-line

	// relaxSweeps is how many times every node is pulled toward the
	// weighted center of its neighbours. Each sweep runs both ways and
	// resolves overlaps after each; six is where the arrangement stops
	// visibly changing on the graphs this draws well.
	relaxSweeps = 6

	// gapShare caps what the gaps between bars may take of a column, so a
	// crowded column loses breathing room instead of its ribbons.
	gapShare = 0.5
)

// node is one endpoint, laid out as a bar.
type node struct {
	id string
	mark
	text    string // label as drawn, with the total when Values is on
	in, out []int  // flow indices, ordered by the other end's position
	depth   int
	value   float64
	x0, x1  float64
	y0, y1  float64
}

func (n *node) center() float64 { return (n.y0 + n.y1) / 2 }

func (n *node) shift(dy float64) { n.y0, n.y1 = n.y0+dy, n.y1+dy }

// flow is one ribbon: a value carried from one node to another.
type flow struct {
	from, to string
	value    float64
	mark
	src, dst int     // node indices
	sy, ty   float64 // the ribbon's top edge where it leaves / arrives
	w        float64 // its thickness
}

// graph is the resolved drawing: nodes in order of first appearance, flows
// in declaration order.
type graph struct {
	nodes    []*node
	flows    []*flow
	maxDepth int
}

// buildGraph interns the flows' endpoints, applies the Node overrides, and
// derives everything that does not depend on the drawing's size.
func buildGraph(cfg Config) (*graph, error) {
	if len(cfg.flows) == 0 {
		return nil, ErrNoFlows
	}

	g := &graph{}
	index := map[string]int{}
	intern := func(id string) int {
		if i, ok := index[id]; ok {
			return i
		}
		i := len(g.nodes)
		index[id] = i
		g.nodes = append(g.nodes, &node{id: id, mark: mark{label: id}})
		return i
	}

	for _, src := range cfg.flows {
		if src.from == src.to {
			return nil, fmt.Errorf("%w: %q", ErrSelfLoop, src.from)
		}
		if !(src.value > 0) {
			return nil, fmt.Errorf("%w: %q → %q carries %v", ErrNonPositiveValue, src.from, src.to, src.value)
		}
		f := src
		f.src, f.dst = intern(f.from), intern(f.to)
		i := len(g.flows)
		g.flows = append(g.flows, &f)
		g.nodes[f.src].out = append(g.nodes[f.src].out, i)
		g.nodes[f.dst].in = append(g.nodes[f.dst].in, i)
	}

	for _, o := range cfg.overrides {
		i, ok := index[o.id]
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrUnknownNode, o.id)
		}
		n := g.nodes[i]
		if o.label != "" {
			n.label = o.label
		}
		if o.colorSet {
			n.color, n.colorSet = o.color, true
		}
	}

	if err := g.assignDepths(); err != nil {
		return nil, err
	}
	g.assignValues()
	g.assignColors()
	for _, n := range g.nodes {
		n.text = n.label
		if cfg.totals {
			n.text += " " + cfg.format(n.value)
		}
	}
	return g, nil
}

// assignDepths puts each node in the column one past its furthest
// predecessor, walking a topological order - which is also how a cycle is
// caught, since a cycle leaves nodes the walk never reaches.
func (g *graph) assignDepths() error {
	indegree := make([]int, len(g.nodes))
	for _, f := range g.flows {
		indegree[f.dst]++
	}
	queue := make([]int, 0, len(g.nodes))
	for i, d := range indegree {
		if d == 0 {
			queue = append(queue, i)
		}
	}

	var seen int
	for len(queue) > 0 {
		n := g.nodes[queue[0]]
		queue, seen = queue[1:], seen+1
		for _, fi := range n.out {
			f := g.flows[fi]
			if d := n.depth + 1; d > g.nodes[f.dst].depth {
				g.nodes[f.dst].depth = d
			}
			if indegree[f.dst]--; indegree[f.dst] == 0 {
				queue = append(queue, f.dst)
			}
		}
	}
	if seen != len(g.nodes) {
		return ErrCycle
	}

	for _, n := range g.nodes {
		g.maxDepth = max(g.maxDepth, n.depth)
	}
	// Sinks belong in the last column: a flow that ends early would
	// otherwise stop mid-drawing with empty space beside it. Every column
	// it leaves still holds the node that gave it its depth, which has an
	// outgoing flow and so stays put.
	for _, n := range g.nodes {
		if len(n.out) == 0 {
			n.depth = g.maxDepth
		}
	}
	return nil
}

// assignValues sizes each bar. The two sides of a node need not agree - a
// funnel loses people between steps - so the bar is the larger.
func (g *graph) assignValues() {
	for _, n := range g.nodes {
		var in, out float64
		for _, fi := range n.in {
			in += g.flows[fi].value
		}
		for _, fi := range n.out {
			out += g.flows[fi].value
		}
		n.value = math.Max(in, out)
	}
}

// assignColors walks the palette in first-appearance order, then hands each
// ribbon its source's hue, so everything leaving one node reads as one flow.
func (g *graph) assignColors() {
	for i, n := range g.nodes {
		if !n.colorSet {
			n.color = defaultOrder[i%len(defaultOrder)]
		}
	}
	for _, f := range g.flows {
		if !f.colorSet {
			f.color = g.nodes[f.src].color
		}
	}
}

// columns groups the nodes by depth. None is empty: a node at depth d > 0
// sits on a longest path that passes through every depth below it.
func (g *graph) columns() [][]*node {
	cols := make([][]*node, g.maxDepth+1)
	for _, n := range g.nodes {
		cols[n.depth] = append(cols[n.depth], n)
	}
	return cols
}

// geometry is the box the columns are placed in, in viewBox units.
type geometry struct {
	left, top, right, bottom float64
	nodeWidth, gap, step     float64
}

// plotGeometry reserves a gutter for the labels of the two end columns,
// each sized by the widest it holds but capped, so one long label squeezes
// itself rather than the ribbons.
func plotGeometry(cfg Config, g *graph) geometry {
	var gutterL, gutterR float64
	for _, n := range g.nodes {
		w := textWidth(n.text) + labelGap + labelPad
		if n.depth == 0 {
			gutterL = math.Max(gutterL, w)
		} else if n.depth == g.maxDepth {
			gutterR = math.Max(gutterR, w)
		}
	}
	limit := cfg.W * 0.3
	gutterL, gutterR = math.Min(gutterL, limit), math.Min(gutterR, limit)

	gm := geometry{
		left:      gutterL,
		right:     cfg.W - gutterR,
		top:       stageMargin,
		bottom:    cfg.H - stageMargin,
		nodeWidth: cfg.nodeWidth,
		gap:       cfg.gap,
	}
	if g.maxDepth > 0 {
		gm.step = (gm.right - gm.left - gm.nodeWidth) / float64(g.maxDepth)
	}

	// A column with more nodes than the gap leaves room for gets a smaller
	// one; without this the bars would be scaled away to nothing.
	var crowd int
	for _, col := range g.columns() {
		crowd = max(crowd, len(col))
	}
	if crowd > 1 {
		if fit := (gm.bottom - gm.top) * gapShare / float64(crowd-1); gm.gap > fit {
			gm.gap = fit
		}
	}
	return gm
}

// layout places every bar and every ribbon end.
func (g *graph) layout(cfg Config) {
	gm := plotGeometry(cfg, g)
	cols := g.columns()

	for d, col := range cols {
		x0 := gm.left + float64(d)*gm.step
		for _, n := range col {
			n.x0, n.x1 = x0, x0+gm.nodeWidth
		}
	}

	ky := g.scale(cols, gm)
	g.stack(cols, gm, ky)
	for i := range relaxSweeps {
		// The pull eases off so late sweeps settle rather than swing.
		alpha := math.Pow(0.97, float64(i))
		g.alignToTargets(cols, alpha)
		resolveOverlaps(cols, gm)
		g.alignToSources(cols, alpha)
		resolveOverlaps(cols, gm)
	}
	g.orderLinks(ky)
}

// scale is the one value-to-pixel factor every column shares: the tightest
// any of them can take, so none overflows the stage.
func (g *graph) scale(cols [][]*node, gm geometry) float64 {
	ky := math.Inf(1)
	for _, col := range cols {
		var sum float64
		for _, n := range col {
			sum += n.value
		}
		if sum <= 0 {
			continue
		}
		room := (gm.bottom - gm.top) - float64(len(col)-1)*gm.gap
		ky = math.Min(ky, room/sum)
	}
	if math.IsInf(ky, 1) || ky <= 0 {
		return 1
	}
	return ky
}

// stack is the starting arrangement: each column's bars in declaration
// order, centered vertically.
func (g *graph) stack(cols [][]*node, gm geometry, ky float64) {
	for _, col := range cols {
		var height float64
		for i, n := range col {
			if i > 0 {
				height += gm.gap
			}
			height += n.value * ky
		}
		y := gm.top + ((gm.bottom-gm.top)-height)/2
		for _, n := range col {
			n.y0, n.y1 = y, y+n.value*ky
			y = n.y1 + gm.gap
		}
	}
}

// alignToTargets pulls each node toward the weighted center of what its
// flows lead to, sweeping right to left.
func (g *graph) alignToTargets(cols [][]*node, alpha float64) {
	for d := len(cols) - 2; d >= 0; d-- {
		for _, n := range cols[d] {
			if y, ok := g.weightedCenter(n.out, func(f *flow) *node { return g.nodes[f.dst] }); ok {
				n.shift((y - n.center()) * alpha)
			}
		}
	}
}

// alignToSources is the same pull in the other direction, toward what flows
// into each node.
func (g *graph) alignToSources(cols [][]*node, alpha float64) {
	for d := 1; d < len(cols); d++ {
		for _, n := range cols[d] {
			if y, ok := g.weightedCenter(n.in, func(f *flow) *node { return g.nodes[f.src] }); ok {
				n.shift((y - n.center()) * alpha)
			}
		}
	}
}

// weightedCenter averages the centers of the nodes at the other end of the
// given flows, weighted by what each carries.
func (g *graph) weightedCenter(flows []int, other func(*flow) *node) (float64, bool) {
	var sum, weight float64
	for _, fi := range flows {
		f := g.flows[fi]
		sum += other(f).center() * f.value
		weight += f.value
	}
	if weight == 0 {
		return 0, false
	}
	return sum / weight, true
}

// resolveOverlaps separates the bars a sweep pushed into each other: down
// from the top, then back up from the bottom for anything driven past it.
func resolveOverlaps(cols [][]*node, gm geometry) {
	for _, col := range cols {
		sort.SliceStable(col, func(i, j int) bool { return col[i].y0 < col[j].y0 })

		y := gm.top
		for _, n := range col {
			if d := y - n.y0; d > 0 {
				n.shift(d)
			}
			y = n.y1 + gm.gap
		}
		y = gm.bottom
		for i := len(col) - 1; i >= 0; i-- {
			n := col[i]
			if d := n.y1 - y; d > 0 {
				n.shift(-d)
			}
			y = n.y0 - gm.gap
		}
	}
}

// orderLinks sorts the ribbons at each bar by where their other end sits -
// which is what keeps them from crossing on the bar's own face - and then
// stacks them from its top edge.
func (g *graph) orderLinks(ky float64) {
	for _, n := range g.nodes {
		sort.SliceStable(n.out, func(i, j int) bool {
			return g.nodes[g.flows[n.out[i]].dst].center() < g.nodes[g.flows[n.out[j]].dst].center()
		})
		sort.SliceStable(n.in, func(i, j int) bool {
			return g.nodes[g.flows[n.in[i]].src].center() < g.nodes[g.flows[n.in[j]].src].center()
		})

		y := n.y0
		for _, fi := range n.out {
			f := g.flows[fi]
			f.w, f.sy = f.value*ky, y
			y += f.w
		}
		y = n.y0
		for _, fi := range n.in {
			f := g.flows[fi]
			f.w, f.ty = f.value*ky, y
			y += f.w
		}
	}
}

// textWidth estimates a label's width by summing per-character advances.
// The server cannot measure rendered text, and this is only sizing a
// gutter, but a single average is wrong by a third between "Illinois" and
// "WOMBAT" - far enough to clip a label against the edge of the drawing.
// (diagram's measurer has the full table; a gutter does not need it.)
func textWidth(s string) float64 {
	var total float64
	for _, r := range s {
		total += labelSize * advance(r)
	}
	return total
}

func advance(r rune) float64 {
	switch {
	case r == ' ':
		return 0.28
	case strings.ContainsRune("ijltfr.,'!|:;()[]-", r):
		return 0.31
	case strings.ContainsRune("mwMW@", r):
		return 0.86
	case r >= '0' && r <= '9':
		return 0.56
	case r >= 'A' && r <= 'Z':
		return 0.67
	case r >= 'a' && r <= 'z':
		return 0.53
	default:
		return 0.55
	}
}
