package diagram

import (
	"fmt"
	"math"
	"sort"
)

// Layout tuning (SVG user units).
const (
	layerGap = 56.0 // gap between layers along the flow axis
	nodeGap  = 36.0 // gap between nodes within a layer along the cross axis
	margin   = 20.0 // padding around the whole drawing
	sweeps   = 8    // crossing-reduction and coordinate-alignment iterations
)

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
	W, H  float64
	boxes []box    // one per node, in node order
	edges []routed // one per edge, in edge order
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
func layout(nodes []layoutNode, edges []edge, dir Direction) (laid, error) {
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
	assignCoords(dir, layers, verts, up, down)

	// 5. Map abstract (flow, cross) to (x, y) per direction.
	for vi := range verts {
		if dir == LeftRight {
			verts[vi].x, verts[vi].y = verts[vi].flow, verts[vi].cross
		} else {
			verts[vi].x, verts[vi].y = verts[vi].cross, verts[vi].flow
		}
	}

	return assemble(len(nodes), verts, edgeChains, reversed, edges), nil
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
// coordinate (packed, then iteratively aligned toward neighbors without
// overlap).
func assignCoords(dir Direction, layers [][]int, verts []vertex, up, down [][]int) {
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
		flow += thick + layerGap
	}

	// Cross: initial left-to-right packing within each layer.
	for L := range layers {
		c := 0.0
		for _, vi := range layers[L] {
			c += crossExtent(vi) / 2
			verts[vi].cross = c
			c += crossExtent(vi)/2 + nodeGap
		}
	}

	// Alignment: pull each vertex toward the average of the given neighbors,
	// enforce the min gap, then recenter the layer's block onto the centroid
	// of the desired positions. The recentring cancels the directional bias
	// of the min-gap sweep, so two siblings that both want the same slot end
	// up straddling it symmetrically (a parent lands centered over them).
	align := func(ls []int, nbr [][]int) {
		desired := make([]float64, len(ls))
		for i, vi := range ls {
			desired[i] = verts[vi].cross
			if ns := nbr[vi]; len(ns) > 0 {
				sum := 0.0
				for _, n := range ns {
					sum += verts[n].cross
				}
				desired[i] = sum / float64(len(ns))
			}
			verts[vi].cross = desired[i]
		}
		for i := 1; i < len(ls); i++ {
			min := verts[ls[i-1]].cross + crossExtent(ls[i-1])/2 + nodeGap + crossExtent(ls[i])/2
			if verts[ls[i]].cross < min {
				verts[ls[i]].cross = min
			}
		}
		var wantSum, gotSum float64
		for i, vi := range ls {
			wantSum += desired[i]
			gotSum += verts[vi].cross
		}
		if n := len(ls); n > 0 {
			shift := (wantSum - gotSum) / float64(n)
			for _, vi := range ls {
				verts[vi].cross += shift
			}
		}
	}

	both := make([][]int, len(verts))
	for vi := range verts {
		both[vi] = append(append([]int(nil), up[vi]...), down[vi]...)
	}
	for it := 0; it < sweeps; it++ {
		if it%2 == 0 {
			for L := 1; L < len(layers); L++ {
				align(layers[L], up)
			}
		} else {
			for L := len(layers) - 2; L >= 0; L-- {
				align(layers[L], down)
			}
		}
	}
	for i := 0; i < sweeps; i++ {
		for L := range layers {
			align(layers[L], both)
		}
	}
}

// assemble translates the drawing to the origin, computes the viewBox, and
// builds the boxes and border-trimmed edge polylines.
func assemble(nodeCount int, verts []vertex, edgeChains [][]int, reversed []bool, edges []edge) laid {
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for i := range verts {
		v := verts[i]
		minX, maxX = math.Min(minX, v.x-v.bw/2), math.Max(maxX, v.x+v.bw/2)
		minY, maxY = math.Min(minY, v.y-v.bh/2), math.Max(maxY, v.y+v.bh/2)
	}
	dx, dy := margin-minX, margin-minY
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
		pts[0] = borderPoint(src.x, src.y, src.bw, src.bh, pts[1])
		pts[len(pts)-1] = borderPoint(dst.x, dst.y, dst.bw, dst.bh, pts[len(pts)-2])
		if reversed[ei] {
			for l, r := 0, len(pts)-1; l < r; l, r = l+1, r-1 {
				pts[l], pts[r] = pts[r], pts[l]
			}
		}
		routedEdges[ei] = routed{pts: pts, label: edges[ei].label}
	}

	return laid{
		W:     maxX - minX + 2*margin,
		H:     maxY - minY + 2*margin,
		boxes: boxes,
		edges: routedEdges,
	}
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
