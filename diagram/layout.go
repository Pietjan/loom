package diagram

import (
	"fmt"
	"math"
	"sort"
)

// Layout defaults (SVG user units); Gap and Margin override the spacing ones.
const (
	defaultLayerGap = 40.0 // gap between layers along the flow axis
	defaultNodeGap  = 28.0 // gap between nodes within a layer along the cross axis
	defaultMargin   = 16.0 // padding around the whole drawing
	sweeps          = 8    // crossing-reduction iterations
)

// gaps is the resolved spacing configuration for one render.
type gaps struct {
	layer, node, margin float64
}

// layoutNode is a node reduced to what layout needs: its id, box size, and
// silhouette (so ports can follow a non-rectangular outline).
type layoutNode struct {
	id    string
	w, h  float64
	shape shape
}

// box is a positioned node box.
type box struct {
	x, y, w, h float64
}

// routed is a drawn edge: a border-to-border polyline (arrow at the last
// point) plus an optional label, stroke style, and arrowhead placement/shape.
type routed struct {
	pts    []xy
	label  string
	style  lineStyle
	arrows arrowEnds
	head   arrowShape
}

// laid is the finished layout ready for SVG emission.
type laid struct {
	W, H   float64
	radius float64  // corner radius for routed edges
	boxes  []box    // one per node, in node order
	edges  []routed // one per edge, in edge order
}

// vertex is a layout node: a real node (node >= 0) or a routing dummy.
type vertex struct {
	node        int // index into the node slice, or -1 for a dummy
	layer       int
	order       int
	bw, bh      float64 // box size in screen space
	shape       shape   // silhouette (rounded for dummies), for port geometry
	flow, cross float64 // abstract center coordinates
	x, y        float64 // final center after direction mapping
}

// layout runs the full Sugiyama pipeline. Every phase iterates slices in
// insertion order and tie-breaks by index — never map iteration — so the
// output is deterministic and golden files don't flake.
func layout(nodes []layoutNode, edges []edge, dir Direction, direct bool, ports PortMode, sp gaps) (laid, error) {
	if len(nodes) == 0 {
		return laid{}, ErrNoNodes
	}

	// 1. Index and validate.
	index := make(map[string]int, len(nodes))
	for i, n := range nodes {
		if _, dup := index[n.id]; dup {
			return laid{}, fmt.Errorf("%w: %q", ErrDuplicateNode, n.id)
		}
		index[n.id] = i
	}
	for _, e := range edges {
		if _, ok := index[e.from]; !ok {
			return laid{}, fmt.Errorf("%w: %q", ErrUnknownNode, e.from)
		}
		if _, ok := index[e.to]; !ok {
			return laid{}, fmt.Errorf("%w: %q", ErrUnknownNode, e.to)
		}
		if e.from == e.to {
			return laid{}, fmt.Errorf("%w: %q", ErrSelfLoop, e.from)
		}
	}

	reversed := breakCycles(edges, index, len(nodes))
	layer := assignLayers(edges, index, reversed, len(nodes))

	// 2. Build vertices for real nodes with their declared sizes.
	verts := make([]vertex, len(nodes))
	for i := range nodes {
		verts[i] = vertex{node: i, layer: layer[i], bw: nodes[i].w, bh: nodes[i].h, shape: nodes[i].shape}
	}

	// 3. Dummy nodes: split every edge into single-layer segments so ordering
	// and routing only ever compare adjacent layers. edgeChains[ei] lists the
	// vertex indices from the lower layer to the higher layer.
	edgeChains := make([][]int, len(edges))
	var segs [][2]int // {upper, lower} vertex indices, upper.layer+1 == lower.layer
	for ei, e := range edges {
		lo, hi := index[e.from], index[e.to]
		if reversed[ei] {
			lo, hi = hi, lo
		}
		chain := []int{lo}
		for L := layer[lo] + 1; L < layer[hi]; L++ {
			verts = append(verts, vertex{node: -1, layer: L, bw: dummySize, bh: dummySize})
			chain = append(chain, len(verts)-1)
		}
		chain = append(chain, hi)
		edgeChains[ei] = chain
		for k := 1; k < len(chain); k++ {
			segs = append(segs, [2]int{chain[k-1], chain[k]})
		}
	}

	maxLayer := 0
	for i := range verts {
		if verts[i].layer > maxLayer {
			maxLayer = verts[i].layer
		}
	}

	// 4. Group vertices by layer; initial order is insertion order (real
	// nodes first in node order, then dummies in edge order).
	layers := make([][]int, maxLayer+1)
	for vi := range verts {
		layers[verts[vi].layer] = append(layers[verts[vi].layer], vi)
	}
	for L := range layers {
		for oi, vi := range layers[L] {
			verts[vi].order = oi
		}
	}

	up := make([][]int, len(verts))
	down := make([][]int, len(verts))
	for _, s := range segs {
		down[s[0]] = append(down[s[0]], s[1])
		up[s[1]] = append(up[s[1]], s[0])
	}

	reduceCrossings(layers, verts, up, down)
	brandesKopf(layers, verts, up, down, crossExtent(dir, verts), sp.node)

	// 5. Resolve the routing on the cross axis, then stack the layers along the
	// flow axis — in that order, because a gap crowded with cross-runs needs
	// more room between its layers than one an edge crosses straight.
	//
	// Both steps are for orthogonal routing only: Direct mode draws each edge
	// as one straight line anchored radially at the node border, so it has
	// neither ports to distribute nor cross-runs to keep apart.
	var ch channels
	if !direct {
		plan := assignPorts(dir, ports, verts, edgeChains, reversed)
		ch = planChannels(verts, edgeChains, plan, len(layers))
	}
	assignFlow(dir, sp, layers, verts, &ch)

	// 6. Map abstract (flow, cross) to (x, y) per direction.
	for vi := range verts {
		if dir == LeftRight {
			verts[vi].x, verts[vi].y = verts[vi].flow, verts[vi].cross
		} else {
			verts[vi].x, verts[vi].y = verts[vi].cross, verts[vi].flow
		}
	}

	return assemble(dir, direct, sp, len(nodes), verts, edgeChains, reversed, edges, ch), nil
}

// breakCycles marks the back edges of a greedy DFS so the graph layers as a
// DAG; their drawn direction is restored later.
//
// Which edge of a cycle ends up reversed decides the whole shape of a cyclic
// drawing — the node that keeps its outgoing edge lays out above the one that
// gives its up — so where the traversal starts matters a great deal. Starting
// at whichever node happened to be declared first meant moving two
// @diagram.Node blocks past each other could silently redraw the diagram.
//
// So begin at the nodes nothing points at. Those are the graph's real entry
// points, they are a property of the edges rather than of the source layout,
// and a traversal rooted there descends the flow instead of picking it up
// somewhere in the middle and pushing part of it backwards. Declaration order
// only breaks ties between several such sources.
func breakCycles(edges []edge, index map[string]int, n int) []bool {
	out := make([][]int, n) // out[u] = edge indices leaving u
	indeg := make([]int, n)
	for ei, e := range edges {
		out[index[e.from]] = append(out[index[e.from]], ei)
		indeg[index[e.to]]++
	}
	roots := make([]int, n)
	for i := range roots {
		roots[i] = i
	}
	sort.SliceStable(roots, func(a, b int) bool {
		return indeg[roots[a]] == 0 && indeg[roots[b]] != 0
	})

	const (
		white = iota
		gray
		black
	)
	state := make([]int, n)
	reversed := make([]bool, len(edges))
	var dfs func(u int)
	dfs = func(u int) {
		state[u] = gray
		for _, ei := range out[u] {
			switch v := index[edges[ei].to]; state[v] {
			case white:
				dfs(v)
			case gray:
				reversed[ei] = true // edge points back onto the DFS stack
			}
		}
		state[u] = black
	}
	for _, u := range roots {
		if state[u] == white {
			dfs(u)
		}
	}
	return reversed
}

// reduceCrossings runs median-heuristic sweeps, keeping the best ordering seen.
func reduceCrossings(layers [][]int, verts []vertex, up, down [][]int) {
	best := snapshot(layers)
	bestCount := countCrossings(layers, verts, down)
	for it := range sweeps {
		if it%2 == 0 {
			for L := 1; L < len(layers); L++ {
				medianSort(layers[L], verts, up)
			}
		} else {
			for L := len(layers) - 2; L >= 0; L-- {
				medianSort(layers[L], verts, down)
			}
		}
		if c := countCrossings(layers, verts, down); c < bestCount {
			bestCount = c
			best = snapshot(layers)
		}
	}
	restore(layers, best, verts)
}

// medianSort reorders one layer by the weighted median of each vertex's
// neighbor positions in the adjacent (fixed) layer. Vertices with no
// neighbors keep their place; ties break by current order.
func medianSort(ls []int, verts []vertex, nbr [][]int) {
	type vk struct {
		vi  int
		key float64
		ord int
	}
	arr := make([]vk, len(ls))
	for i, vi := range ls {
		arr[i] = vk{vi: vi, key: float64(verts[vi].order), ord: verts[vi].order}
		if ns := nbr[vi]; len(ns) > 0 {
			pos := make([]float64, len(ns))
			for j, n := range ns {
				pos[j] = float64(verts[n].order)
			}
			sort.Float64s(pos)
			arr[i].key = weightedMedian(pos)
		}
	}
	sort.SliceStable(arr, func(a, b int) bool {
		if arr[a].key != arr[b].key {
			return arr[a].key < arr[b].key
		}
		return arr[a].ord < arr[b].ord
	})
	for i := range arr {
		ls[i] = arr[i].vi
		verts[arr[i].vi].order = i
	}
}

// weightedMedian is the Eades–Wells median position of sorted neighbor slots.
func weightedMedian(pos []float64) float64 {
	m := len(pos) / 2
	if len(pos)%2 == 1 {
		return pos[m]
	}
	left := pos[m-1] - pos[0]
	right := pos[len(pos)-1] - pos[m]
	if left+right == 0 {
		return (pos[m-1] + pos[m]) / 2
	}
	return (pos[m-1]*right + pos[m]*left) / (left + right)
}

// countCrossings totals edge crossings across all adjacent layer pairs by
// counting inversions in each pair's target order.
func countCrossings(layers [][]int, verts []vertex, down [][]int) int {
	total := 0
	for L := 0; L+1 < len(layers); L++ {
		var targets []int
		for _, vi := range layers[L] {
			ds := append([]int(nil), down[vi]...)
			sort.Slice(ds, func(a, b int) bool { return verts[ds[a]].order < verts[ds[b]].order })
			for _, w := range ds {
				targets = append(targets, verts[w].order)
			}
		}
		for i := 0; i < len(targets); i++ {
			for j := i + 1; j < len(targets); j++ {
				if targets[i] > targets[j] {
					total++
				}
			}
		}
	}
	return total
}

// crossExtent measures a vertex across the flow — the axis Brandes–Köpf lays
// out and the ports spread along.
func crossExtent(dir Direction, verts []vertex) func(int) float64 {
	return func(vi int) float64 {
		if dir == LeftRight {
			return verts[vi].bh
		}
		return verts[vi].bw
	}
}

// assignFlow stacks the layers along the flow axis and places each gap's
// tracks inside the band between them.
//
// A gap holds its tracks at even spacing, and grows past the configured layer
// gap only when it has more of them than will fit — so an ordinary drawing
// keeps exactly the spacing it asks for, and a crowded one stops folding its
// corners into each other. The band runs from the bottom of the thickest box in
// the layer above to the top of the thickest in the layer below, so every stub
// reaches its track: a dummy's face, or a diamond's port receded by flowInset,
// sits above the band rather than in it.
func assignFlow(dir Direction, sp gaps, layers [][]int, verts []vertex, ch *channels) {
	flowExtent := func(vi int) float64 {
		if dir == LeftRight {
			return verts[vi].bw
		}
		return verts[vi].bh
	}

	ch.flow = make([][]float64, max(len(layers)-1, 0))
	flow := 0.0
	for L := range layers {
		thick := 0.0
		for _, vi := range layers[L] {
			if e := flowExtent(vi); e > thick {
				thick = e
			}
		}
		center := flow + thick/2
		for _, vi := range layers[L] {
			verts[vi].flow = center
		}
		flow += thick
		if L+1 == len(layers) {
			break
		}

		// Direct mode allocates no tracks at all, so its gaps stay as configured.
		n := 0
		if L < len(ch.count) {
			n = ch.count[L]
		}
		height := sp.layer
		if n > 0 {
			// A track needs a corner's worth of clearance on each side, but
			// never more than half the gap it was given — a deliberately tight
			// gap draws tight corners rather than being overruled.
			if need := float64(n+1) * math.Min(minBend, sp.layer/2); need > height {
				height = need
			}
			ch.flow[L] = make([]float64, n)
			for k := range ch.flow[L] {
				ch.flow[L][k] = flow + height*float64(k+1)/float64(n+1)
			}
		}
		flow += height
	}
}

// brandesKopf assigns cross coordinates with Brandes and Köpf's algorithm
// ("Fast and Simple Horizontal Coordinate Assignment", GD 2002).
//
// Each vertex is aligned with a median neighbor, chaining vertices into blocks
// that share one coordinate; the blocks are then compacted against the minimum
// gap. That runs four times — aligning against the layer above and against the
// one below, packing toward each side of the cross axis — and the four
// candidates are averaged, so no single pass's directional bias survives. Long
// edges come out straight and a parent settles on the true midpoint of its
// children, which the naive averaging sweep only approximated.
func brandesKopf(layers [][]int, verts []vertex, up, down [][]int, size func(int) float64, gap float64) {
	// Two nodes can be joined by several edges (a 2-cycle leaves a pair of
	// parallel segments). That is one alignment candidate, not several, or the
	// duplicate drags the median onto itself.
	up, down = distinct(up), distinct(down)
	marked := markConflicts(layers, verts, up)

	var cand [4][]float64
	for i := range cand {
		upward, rightward := i&1 != 0, i&2 != 0

		// Feed the core one normalized view: layers ordered from the fixed
		// end, each layer ordered from the side we pack toward.
		ls := make([][]int, len(layers))
		for L, layer := range layers {
			at := L
			if upward {
				at = len(layers) - 1 - L
			}
			ls[at] = append([]int(nil), layer...)
			if rightward {
				reverseInts(ls[at])
			}
		}
		nbr := up
		if upward {
			nbr = down
		}

		cand[i] = alignAndCompact(ls, nbr, verts, marked, upward, size, gap)
		if rightward {
			for v := range cand[i] {
				cand[i][v] = -cand[i][v] // mirror back into the real axis
			}
		}
	}
	balance(cand, verts, size)
}

// markConflicts finds type-1 conflicts: a crossing between an inner segment
// (both ends routing dummies, i.e. a long edge passing through) and an
// ordinary one. The ordinary segment is marked so alignment never follows it,
// which is what keeps long edges drawn straight.
func markConflicts(layers [][]int, verts []vertex, up [][]int) map[[2]int]bool {
	pos := make([]int, len(verts))
	for _, layer := range layers {
		for k, v := range layer {
			pos[v] = k
		}
	}
	dummy := func(v int) bool { return verts[v].node < 0 }
	innerUpper := func(v int) int {
		if !dummy(v) {
			return -1
		}
		for _, u := range up[v] {
			if dummy(u) {
				return u
			}
		}
		return -1
	}

	marked := map[[2]int]bool{}
	for i := 1; i+1 < len(layers); i++ {
		lower := layers[i+1]
		k0, l := 0, 0
		for l1, v := range lower {
			inner := innerUpper(v)
			// Scan up to each inner segment (and to the layer's end): every
			// segment landing outside the window those two bound crosses one.
			if inner < 0 && l1 != len(lower)-1 {
				continue
			}
			k1 := len(layers[i]) - 1
			if inner >= 0 {
				k1 = pos[inner]
			}
			for ; l <= l1; l++ {
				for _, u := range up[lower[l]] {
					if k := pos[u]; k < k0 || k > k1 {
						marked[[2]int{u, lower[l]}] = true
					}
				}
			}
			k0 = k1
		}
	}
	return marked
}

// alignAndCompact runs one of the four passes. ls holds the layers in
// processing order, each ordered from the side being packed toward; nbr[v]
// lists v's neighbors in the preceding processing layer. It returns one
// coordinate per vertex, still in the mirrored frame when packing rightward.
func alignAndCompact(ls [][]int, nbr [][]int, verts []vertex, marked map[[2]int]bool, upward bool, size func(int) float64, gap float64) []float64 {
	n := len(verts)
	pos := make([]int, n)  // index within the vertex's layer
	home := make([]int, n) // which layer of ls the vertex is in
	for L, layer := range ls {
		for k, v := range layer {
			pos[v], home[v] = k, L
		}
	}
	// marked is keyed in the graph's own direction, so flip when aligning up.
	conflict := func(u, v int) bool {
		if upward {
			return marked[[2]int{v, u}]
		}
		return marked[[2]int{u, v}]
	}

	// Alignment: each vertex tries to join its median neighbor's block, but
	// only if that keeps the aligned pairs non-crossing (r) and skips conflicts.
	root, align := make([]int, n), make([]int, n)
	for v := range verts {
		root[v], align[v] = v, v
	}
	for L := 1; L < len(ls); L++ {
		r := -1
		for _, v := range ls[L] {
			ns := append([]int(nil), nbr[v]...)
			if len(ns) == 0 {
				continue
			}
			sort.Slice(ns, func(a, b int) bool { return pos[ns[a]] < pos[ns[b]] })
			for _, m := range [2]int{(len(ns) - 1) / 2, len(ns) / 2} {
				if align[v] != v {
					break
				}
				u := ns[m]
				if conflict(u, v) || r >= pos[u] {
					continue
				}
				align[u], root[v], align[v] = v, root[u], root[u]
				r = pos[u]
			}
		}
	}

	// Compaction: place each block as far toward the packing side as its
	// in-layer predecessors allow. Blocks that end up in separate classes are
	// reconciled afterwards by their sink's accumulated shift.
	sink, shift := make([]int, n), make([]float64, n)
	x, placed := make([]float64, n), make([]bool, n)
	for v := range verts {
		sink[v], shift[v] = v, math.Inf(1)
	}
	var place func(v int)
	place = func(v int) {
		if placed[v] {
			return
		}
		placed[v] = true
		for w := v; ; {
			if p := pos[w]; p > 0 {
				pred := ls[home[w]][p-1]
				u := root[pred]
				place(u)
				if sink[v] == v {
					sink[v] = sink[u]
				}
				delta := (size(pred)+size(w))/2 + gap
				if sink[v] != sink[u] {
					shift[sink[u]] = math.Min(shift[sink[u]], x[v]-x[u]-delta)
				} else {
					x[v] = math.Max(x[v], x[u]+delta)
				}
			}
			if w = align[w]; w == v {
				break
			}
		}
	}
	for v := range verts {
		if root[v] == v {
			place(v)
		}
	}

	out := make([]float64, n)
	for v := range verts {
		out[v] = x[root[v]]
		if s := shift[sink[root[v]]]; !math.IsInf(s, 1) {
			out[v] += s
		}
	}
	return out
}

// balance shifts the four candidates onto the narrowest one and gives each
// vertex the average of its two middle values. Averaging order statistics of
// assignments that each respect the min gap preserves the min gap.
func balance(cand [4][]float64, verts []vertex, size func(int) float64) {
	span := func(c []float64, pad bool) (lo, hi float64) {
		lo, hi = math.Inf(1), math.Inf(-1)
		for v, x := range c {
			half := 0.0
			if pad {
				half = size(v) / 2
			}
			lo, hi = math.Min(lo, x-half), math.Max(hi, x+half)
		}
		return lo, hi
	}

	best, narrowest := 0, math.Inf(1)
	for i, c := range cand {
		lo, hi := span(c, true)
		if hi-lo < narrowest {
			narrowest, best = hi-lo, i
		}
	}

	lo, hi := span(cand[best], false)
	for i, c := range cand {
		l, h := span(c, false)
		// A pass packed toward the high side keeps its far edge, one packed
		// toward the low side its near edge.
		d := lo - l
		if i&2 != 0 {
			d = hi - h
		}
		for v := range c {
			c[v] += d
		}
	}

	for v := range verts {
		vals := [4]float64{cand[0][v], cand[1][v], cand[2][v], cand[3][v]}
		sort.Float64s(vals[:])
		verts[v].cross = (vals[1] + vals[2]) / 2
	}
}

// distinct copies an adjacency list with repeated neighbors collapsed,
// keeping first-seen order.
func distinct(adj [][]int) [][]int {
	out := make([][]int, len(adj))
	for v, ns := range adj {
		seen := make(map[int]bool, len(ns))
		for _, n := range ns {
			if !seen[n] {
				seen[n] = true
				out[v] = append(out[v], n)
			}
		}
	}
	return out
}

func reverseInts(s []int) {
	for l, r := 0, len(s)-1; l < r; l, r = l+1, r-1 {
		s[l], s[r] = s[r], s[l]
	}
}

// assemble translates the drawing to the origin, computes the viewBox, and
// builds the boxes and border-trimmed edge polylines.
func assemble(dir Direction, direct bool, sp gaps, nodeCount int, verts []vertex, edgeChains [][]int, reversed []bool, edges []edge, ch channels) laid {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for i := range verts {
		v := verts[i]
		minX, maxX = math.Min(minX, v.x-v.bw/2), math.Max(maxX, v.x+v.bw/2)
		minY, maxY = math.Min(minY, v.y-v.bh/2), math.Max(maxY, v.y+v.bh/2)
	}
	dx, dy := sp.margin-minX, sp.margin-minY
	for i := range verts {
		verts[i].x += dx
		verts[i].y += dy
	}

	boxes := make([]box, nodeCount)
	for i := range nodeCount {
		v := verts[i]
		boxes[i] = box{x: v.x, y: v.y, w: v.bw, h: v.bh}
	}

	// The routing was resolved on the cross axis before the layers were stacked;
	// translating the drawing shifts it by the same amount.
	crossShift, flowShift := dx, dy
	if dir == LeftRight {
		crossShift, flowShift = dy, dx
	}

	routedEdges := make([]routed, len(edges))
	for ei := range edges {
		chain := edgeChains[ei]
		pts := make([]xy, len(chain))
		for k, vi := range chain {
			pts[k] = xy{verts[vi].x, verts[vi].y}
		}
		src, dst := verts[chain[0]], verts[chain[len(chain)-1]]
		if direct {
			pts[0] = borderPoint(src.x, src.y, src.bw, src.bh, pts[1])
			pts[len(pts)-1] = borderPoint(dst.x, dst.y, dst.bw, dst.bh, pts[len(pts)-2])
		} else {
			// Leave and arrive square to the box face at the edge's port, then
			// step between waypoints with right angles, each step riding the
			// track its gap allocated it.
			r := ch.routes[ei]
			for k := range pts {
				setCross(&pts[k], dir, r.cross[k]+crossShift)
			}
			pts[0] = faceExit(src, dir, r.cross[0]-src.cross)
			pts[len(pts)-1] = faceEntry(dst, dir, r.cross[len(r.cross)-1]-dst.cross)

			mid := make([]float64, len(r.track))
			for k, t := range r.track {
				if t >= 0 {
					mid[k] = ch.flow[verts[chain[k]].layer][t] + flowShift
				}
			}
			pts = orthoRoute(dir, pts, r.track, mid)
		}
		if reversed[ei] {
			for l, r := 0, len(pts)-1; l < r; l, r = l+1, r-1 {
				pts[l], pts[r] = pts[r], pts[l]
			}
		}
		routedEdges[ei] = routed{pts: pts, label: edges[ei].label, style: edges[ei].style, arrows: edges[ei].arrows, head: edges[ei].head}
	}

	return laid{
		W: maxX - minX + 2*sp.margin,
		H: maxY - minY + 2*sp.margin,
		// A tight layer gap leaves no room for two rounded corners between
		// layers, so clamp the radius rather than let them merge.
		radius: math.Min(cornerRadius, sp.layer/4),
		boxes:  boxes,
		edges:  routedEdges,
	}
}

// faceExit is the point on the box outline an edge leaves along the flow axis,
// off units along the cross axis from the face centre (its port).
func faceExit(v vertex, dir Direction, off float64) xy {
	if dir == LeftRight {
		return xy{v.x + v.bw/2 - flowInset(v, dir, off), v.y + off}
	}
	return xy{v.x + off, v.y + v.bh/2 - flowInset(v, dir, off)}
}

// faceEntry is the point on the box outline an edge arrives at, off units along
// the cross axis from the face centre (its port).
func faceEntry(v vertex, dir Direction, off float64) xy {
	if dir == LeftRight {
		return xy{v.x - v.bw/2 + flowInset(v, dir, off), v.y + off}
	}
	return xy{v.x + off, v.y - v.bh/2 + flowInset(v, dir, off)}
}

// flowInset is how far a port at cross-offset off recedes along the flow axis
// from a box's flat face, so ports track a non-rectangular outline instead of
// floating past it. A diamond's flow-axis faces are single vertices, so a port
// slides up the slanted edge as it moves off-centre (reaching the side vertex
// at the full half-width); rounded and stadium boxes have a flat face and don't
// recede.
func flowInset(v vertex, dir Direction, off float64) float64 {
	if v.shape != diamond || off == 0 {
		return 0
	}
	crossHalf, flowHalf := v.bw/2, v.bh/2
	if dir == LeftRight {
		crossHalf, flowHalf = v.bh/2, v.bw/2
	}
	if crossHalf == 0 {
		return 0
	}
	return math.Min(flowHalf*math.Abs(off)/crossHalf, flowHalf)
}

// portPlan is the resolved attachment for every edge: the cross-axis offset of
// its exit port (at the source) and entry port (at the destination), plus
// whether each port is its face's only attachment. A lone port sits at the face
// centre and is free to slide to keep its edge straight; one that shares a face
// is placed relative to its neighbours, and moving it would throw the fan's
// spacing off centre.
type portPlan struct {
	exit, entry         []float64
	exitFree, entryFree []bool
}

// assignPorts gives each box face at most two attachment points ("floating
// ports") instead of stacking every edge on the face centre.
//
// The bug it fixes is an outgoing shaft drawn over an incoming arrowhead when
// both share a point (e.g. a 2-cycle). Edges are sorted into two lanes by
// whether they run along the flow or were reversed to break a cycle, and the
// lanes keep the same order on every face: forward edges on the low side, back
// edges on the high side. Using the same convention at both ends is what stops
// an edge and the back edge beside it from swapping sides and crossing over —
// ordering the lanes per face (by drawn direction, or by where the neighbours
// sit) puts the back edge on the left at one node and the right at the other,
// which is exactly the X it is meant to avoid.
//
// Within a lane the edges share one port, so fans and joins read as one point.
// PortsSpread instead gives every edge its own port, ordered by its neighbour —
// busier, but each connection stays visually distinct.
func assignPorts(dir Direction, mode PortMode, verts []vertex, edgeChains [][]int, reversed []bool) portPlan {
	// Ports sit this far apart, capped so a wide box doesn't fling them to the
	// corners.
	const spread, gapMax = 0.7, 14.0

	n := len(edgeChains)
	p := portPlan{
		exit: make([]float64, n), entry: make([]float64, n),
		exitFree: make([]bool, n), entryFree: make([]bool, n),
	}

	crossExtent := func(vi int) float64 {
		if dir == LeftRight {
			return verts[vi].bh
		}
		return verts[vi].bw
	}

	// One edge incident to a node's face: which edge, whether it was reversed
	// (so which lane it belongs in), and where the next stop along its chain
	// sits on the cross axis (to order the ports in spread mode).
	type incident struct {
		ei       int
		rev      bool
		nbrCross float64
	}
	byFace := func(node, nbr func([]int) int) map[int][]incident {
		buckets := map[int][]incident{}
		for ei, chain := range edgeChains {
			buckets[node(chain)] = append(buckets[node(chain)],
				incident{ei, reversed[ei], verts[nbr(chain)].cross})
		}
		return buckets
	}
	// chain runs lower layer → higher: the node at chain[0] attaches on its exit
	// face, the one at chain[len-1] on its entry face.
	exitFaces := byFace(func(c []int) int { return c[0] }, func(c []int) int { return c[1] })
	entryFaces := byFace(func(c []int) int { return c[len(c)-1] }, func(c []int) int { return c[len(c)-2] })

	place := func(faces map[int][]incident, off []float64, free []bool) {
		for vi, inc := range faces {
			lone := len(inc) == 1 // the face's only attachment, so free to slide
			var slots [][]int     // edges per port, in cross order
			if mode == PortsSpread {
				sort.Slice(inc, func(a, b int) bool {
					if inc[a].nbrCross != inc[b].nbrCross {
						return inc[a].nbrCross < inc[b].nbrCross
					}
					return inc[a].ei < inc[b].ei // deterministic
				})
				for _, in := range inc {
					slots = append(slots, []int{in.ei})
				}
			} else {
				// Two lanes: edges running along the flow, and back edges.
				// Order them by where they actually head — the mean position of
				// the next stop along each chain — so a lane sits on the side
				// its edges travel toward. That matters for a long edge, whose
				// next stop is a routing dummy the ordering phase has already
				// placed: put the lane on the other side and the edge weaves
				// across its own dummy.
				var lane [2][]int
				var sum [2]float64
				for _, in := range inc {
					g := 0
					if in.rev {
						g = 1
					}
					lane[g] = append(lane[g], in.ei)
					sum[g] += in.nbrCross
				}
				type laneInfo struct {
					mean float64
					rev  int
					eis  []int
				}
				var lanes []laneInfo
				for g, l := range lane {
					if len(l) > 0 {
						lanes = append(lanes, laneInfo{sum[g] / float64(len(l)), g, l})
					}
				}
				sort.Slice(lanes, func(a, b int) bool {
					if lanes[a].mean != lanes[b].mean {
						return lanes[a].mean < lanes[b].mean
					}
					// Both lanes head for the same place — an antiparallel pair
					// between adjacent layers — so position says nothing. Fall
					// back to the flow lane first. This tie-break must resolve
					// the same way at both ends of an edge, which is why it keys
					// off the reversal flag and not the direction the arrow
					// points: "outgoing" swaps meaning between the two faces,
					// and that swap is exactly what makes the pair cross.
					return lanes[a].rev < lanes[b].rev
				})
				for _, l := range lanes {
					slots = append(slots, l.eis)
				}
			}
			k := len(slots)
			step := math.Min(crossExtent(vi)*spread/float64(k), gapMax)
			for i, eis := range slots {
				o := step * (float64(i) - float64(k-1)/2) // centred; k==1 ⇒ 0
				for _, ei := range eis {
					off[ei], free[ei] = o, lone
				}
			}
		}
	}
	place(exitFaces, p.exit, p.exitFree)
	place(entryFaces, p.entry, p.entryFree)
	return p
}

// straighten pulls an edge's two ports onto a common cross coordinate when they
// sit closer than minBend — a step that small cannot render as a corner (see
// minBend), so it would come out as a wobble.
//
// Only a port that is its face's sole attachment may move: one that belongs to
// a fan is spaced against its neighbours, and sliding it to suit its own edge
// would leave the fan visibly lopsided about the box's centreline. So when just
// one end is free it travels the whole way to meet the fixed port, when both
// are free they split the difference, and when neither is the step stays.
func straighten(srcCross, dstCross, ex, en float64, srcFree, dstFree bool) (float64, float64) {
	a, b := srcCross+ex, dstCross+en
	if math.Abs(a-b) > minBend {
		return ex, en
	}
	// A free port starts centred, so it never travels more than minBend from
	// the centre and stays comfortably on the face.
	switch {
	case srcFree && dstFree:
		mid := (a + b) / 2
		return mid - srcCross, mid - dstCross
	case srcFree:
		return b - srcCross, en
	case dstFree:
		return ex, a - dstCross
	}
	return ex, en
}

// minBend is the smallest cross-axis step that can be drawn as a real corner.
// Rounding a bend pulls the path back by the radius on both sides, so a step
// under twice that leaves no straight run between the two curves: the "corner"
// degenerates into a wobble. Steps below it are flattened instead.
const minBend = 2 * cornerRadius

// setCross moves a point along the cross axis, leaving its flow coordinate be.
func setCross(p *xy, dir Direction, c float64) {
	if dir == LeftRight {
		p.y = c
	} else {
		p.x = c
	}
}

// orthoRoute turns an edge's stops into an axis-aligned path. A segment the
// channel plan gave a track (track[k] >= 0) steps across at that track's flow
// coordinate mid[k], with a right angle at each end; the rest run straight,
// their stops already having been pulled onto a common cross coordinate by
// resolveStops.
func orthoRoute(dir Direction, stops []xy, track []int, mid []float64) []xy {
	out := []xy{stops[0]}
	for k, b := range stops[1:] {
		if track[k] >= 0 {
			a := out[len(out)-1]
			if dir == LeftRight {
				out = append(out, xy{mid[k], a.y}, xy{mid[k], b.y})
			} else {
				out = append(out, xy{a.x, mid[k]}, xy{b.x, mid[k]})
			}
		}
		out = append(out, b)
	}
	return dedupe(out)
}

func snapshot(layers [][]int) [][]int {
	cp := make([][]int, len(layers))
	for i := range layers {
		cp[i] = append([]int(nil), layers[i]...)
	}
	return cp
}

func restore(layers, best [][]int, verts []vertex) {
	for i := range layers {
		copy(layers[i], best[i])
		for oi, vi := range layers[i] {
			verts[vi].order = oi
		}
	}
}
