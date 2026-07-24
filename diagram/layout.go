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

// layoutNode is a node reduced to what layout needs: its id and box size.
type layoutNode struct {
	id   string
	w, h float64
}

// box is a positioned node box.
type box struct {
	x, y, w, h float64
}

// routed is a drawn edge: a border-to-border polyline (arrow at the last
// point) plus an optional label.
type routed struct {
	pts   []xy
	label string
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
	flow, cross float64 // abstract center coordinates
	x, y        float64 // final center after direction mapping
}

// layout runs the full Sugiyama pipeline. Every phase iterates slices in
// insertion order and tie-breaks by index — never map iteration — so the
// output is deterministic and golden files don't flake.
func layout(nodes []layoutNode, edges []edge, dir Direction, direct bool, sp gaps) (laid, error) {
	if len(nodes) == 0 {
		return laid{}, ErrNoNodes
	}

	// 1. Index and validate.
	index := make(map[string]int, len(nodes))
	for i, n := range nodes {
		if _, dup := index[n.id]; dup {
			return laid{}, fmt.Errorf("diagram: duplicate node id %q", n.id)
		}
		index[n.id] = i
	}
	for _, e := range edges {
		if _, ok := index[e.from]; !ok {
			return laid{}, fmt.Errorf("diagram: edge references unknown node %q", e.from)
		}
		if _, ok := index[e.to]; !ok {
			return laid{}, fmt.Errorf("diagram: edge references unknown node %q", e.to)
		}
		if e.from == e.to {
			return laid{}, fmt.Errorf("diagram: self-loop on node %q not supported", e.from)
		}
	}

	reversed := breakCycles(edges, index, len(nodes))
	layer := assignLayers(edges, index, reversed, len(nodes))

	// 2. Build vertices for real nodes with their declared sizes.
	verts := make([]vertex, len(nodes))
	for i := range nodes {
		verts[i] = vertex{node: i, layer: layer[i], bw: nodes[i].w, bh: nodes[i].h}
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
	assignCoords(dir, sp, layers, verts, up, down)

	// 5. Map abstract (flow, cross) to (x, y) per direction.
	for vi := range verts {
		if dir == LeftRight {
			verts[vi].x, verts[vi].y = verts[vi].flow, verts[vi].cross
		} else {
			verts[vi].x, verts[vi].y = verts[vi].cross, verts[vi].flow
		}
	}

	return assemble(dir, direct, sp, len(nodes), verts, edgeChains, reversed, edges), nil
}

// breakCycles marks the back edges of a greedy DFS (in node insertion order)
// so the graph layers as a DAG; their drawn direction is restored later.
func breakCycles(edges []edge, index map[string]int, n int) []bool {
	out := make([][]int, n) // out[u] = edge indices leaving u
	for ei, e := range edges {
		out[index[e.from]] = append(out[index[e.from]], ei)
	}
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
	for u := 0; u < n; u++ {
		if state[u] == white {
			dfs(u)
		}
	}
	return reversed
}

// assignLayers gives each node a longest-path layer on the acyclic graph
// (edges flipped where reversed points them backward).
func assignLayers(edges []edge, index map[string]int, reversed []bool, n int) []int {
	adj := make([][]int, n)
	indeg := make([]int, n)
	for ei, e := range edges {
		s, d := index[e.from], index[e.to]
		if reversed[ei] {
			s, d = d, s
		}
		adj[s] = append(adj[s], d)
		indeg[d]++
	}
	layer := make([]int, n)
	var queue []int
	for u := 0; u < n; u++ {
		if indeg[u] == 0 {
			queue = append(queue, u)
		}
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range adj[u] {
			if layer[u]+1 > layer[v] {
				layer[v] = layer[u] + 1
			}
			if indeg[v]--; indeg[v] == 0 {
				queue = append(queue, v)
			}
		}
	}
	return layer
}

// reduceCrossings runs median-heuristic sweeps, keeping the best ordering seen.
func reduceCrossings(layers [][]int, verts []vertex, up, down [][]int) {
	best := snapshot(layers)
	bestCount := countCrossings(layers, verts, down)
	for it := 0; it < sweeps; it++ {
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

// assignCoords sets each vertex's flow coordinate (from its layer) and cross
// coordinate (Brandes–Köpf).
func assignCoords(dir Direction, sp gaps, layers [][]int, verts []vertex, up, down [][]int) {
	flowExtent := func(vi int) float64 {
		if dir == LeftRight {
			return verts[vi].bw
		}
		return verts[vi].bh
	}
	crossExtent := func(vi int) float64 {
		if dir == LeftRight {
			return verts[vi].bh
		}
		return verts[vi].bw
	}

	// Flow: stack layers by their thickest box.
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
		flow += thick + sp.layer
	}

	brandesKopf(layers, verts, up, down, crossExtent, sp.node)
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
func assemble(dir Direction, direct bool, sp gaps, nodeCount int, verts []vertex, edgeChains [][]int, reversed []bool, edges []edge) laid {
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
	for i := 0; i < nodeCount; i++ {
		v := verts[i]
		boxes[i] = box{x: v.x, y: v.y, w: v.bw, h: v.bh}
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
			// Leave and arrive square to the box face, then step between
			// waypoints with right angles.
			pts[0] = faceExit(src, dir)
			pts[len(pts)-1] = faceEntry(dst, dir)
			pts = orthoRoute(pts, dir)
		}
		if reversed[ei] {
			for l, r := 0, len(pts)-1; l < r; l, r = l+1, r-1 {
				pts[l], pts[r] = pts[r], pts[l]
			}
		}
		routedEdges[ei] = routed{pts: pts, label: edges[ei].label}
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

// faceExit is the centre of the box face an edge leaves along the flow axis.
func faceExit(v vertex, dir Direction) xy {
	if dir == LeftRight {
		return xy{v.x + v.bw/2, v.y}
	}
	return xy{v.x, v.y + v.bh/2}
}

// faceEntry is the centre of the box face an edge arrives at.
func faceEntry(v vertex, dir Direction) xy {
	if dir == LeftRight {
		return xy{v.x - v.bw/2, v.y}
	}
	return xy{v.x, v.y - v.bh/2}
}

// orthoRoute turns a list of waypoints into an axis-aligned path: where two
// consecutive stops differ on the cross axis, it inserts an elbow halfway
// along the flow axis, so the edge steps across with right angles.
func orthoRoute(stops []xy, dir Direction) []xy {
	// Offsets below this are snapped straight: an elbow for a sub-pixel
	// misalignment reads as a wobble, not a corner.
	const align = 2.0

	out := []xy{stops[0]}
	for _, b := range stops[1:] {
		a := out[len(out)-1]
		if dir == LeftRight {
			if math.Abs(a.y-b.y) > align {
				mid := (a.x + b.x) / 2
				out = append(out, xy{mid, a.y}, xy{mid, b.y})
			} else {
				b.y = a.y
			}
		} else if math.Abs(a.x-b.x) > align {
			mid := (a.y + b.y) / 2
			out = append(out, xy{a.x, mid}, xy{b.x, mid})
		} else {
			b.x = a.x
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
