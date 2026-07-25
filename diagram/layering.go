package diagram

// Layering decides which layer each node sits on — its position along the flow
// axis — before any of the ordering, coordinate or routing work.
//
// The obvious method is longest path: give every node a layer one past its
// deepest predecessor. It is quick and it never breaks a constraint, but it
// pins every source at layer 0, which stretches edges. A node dragged down by
// one parent leaves its other parents behind, and those edges then span the
// layers in between and have to route around whatever sits there. For a
// component whose whole promise is placing things automatically, the worse
// problem is that the result depends on input order: which edge a cycle
// happens to break on decides which node counts as a source, and the drawing
// changes shape around it.
//
// Network simplex instead finds the ranking of minimum total edge length
// (Gansner, Koutsofios, North and Vo, "A Technique for Drawing Directed
// Graphs", IEEE TSE 19(3), 1993, section 2). Nodes settle next to the
// neighbours they are joined to, long edges survive only where the graph
// really demands them, and the drawing stops moving about when the node
// declarations are reordered.
//
// The formulation: minimise the sum over edges of (rank[v] - rank[u] - 1)
// subject to rank[v] - rank[u] >= 1. Start from a feasible ranking, build a
// spanning tree of tight edges, then repeatedly swap out a tree edge whose cut
// value is negative for the least-slack edge crossing back the other way. Each
// swap shortens the total, and the process ends when no cut value is negative.

// rankInf stands in for "no candidate yet" when minimising slack.
const rankInf = int(^uint(0) >> 1)

// layerEdge is an edge of the acyclic graph layering works on: u sits on an
// earlier layer than v.
type layerEdge struct{ u, v int }

// layerComponent is one weakly connected piece of that graph. Layering is
// solved per component, since ranks in one place no constraint on another.
type layerComponent struct {
	nodes []int
	edges []int // indices into the edge slice
}

// assignLayers gives each node a layer on the acyclic graph (edges flipped
// where reversed points them backward), minimising total edge length.
func assignLayers(edges []edge, index map[string]int, reversed []bool, n int) []int {
	es := make([]layerEdge, 0, len(edges))
	for ei, e := range edges {
		s, d := index[e.from], index[e.to]
		if reversed[ei] {
			s, d = d, s
		}
		es = append(es, layerEdge{s, d})
	}

	rank := longestPathRanks(es, n)
	for _, c := range weakComponents(es, n) {
		if len(c.nodes) > 1 {
			networkSimplex(es, c, rank)
		}
		normalizeRanks(c, rank)
	}
	// Every phase downstream indexes layers directly, so a ranking that broke a
	// constraint would corrupt the drawing rather than merely look worse. The
	// longest-path ranking is always feasible, so fall back to it.
	if !feasibleRanks(es, rank) {
		return longestPathRanks(es, n)
	}
	return rank
}

// longestPathRanks is the simple feasible ranking: one past the deepest
// predecessor. Network simplex starts from it, and it is the fallback.
func longestPathRanks(es []layerEdge, n int) []int {
	adj := make([][]int, n)
	indeg := make([]int, n)
	for _, e := range es {
		adj[e.u] = append(adj[e.u], e.v)
		indeg[e.v]++
	}
	rank := make([]int, n)
	var queue []int
	for u := range n {
		if indeg[u] == 0 {
			queue = append(queue, u)
		}
	}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range adj[u] {
			if rank[u]+1 > rank[v] {
				rank[v] = rank[u] + 1
			}
			if indeg[v]--; indeg[v] == 0 {
				queue = append(queue, v)
			}
		}
	}
	return rank
}

func feasibleRanks(es []layerEdge, rank []int) bool {
	for _, e := range es {
		if rank[e.v]-rank[e.u] < 1 {
			return false
		}
	}
	return true
}

// weakComponents groups nodes joined by edges in either direction. Nodes are
// visited in index order and edges in insertion order, so the grouping — and
// with it every choice made from it — is deterministic.
func weakComponents(es []layerEdge, n int) []layerComponent {
	inc := make([][]int, n)
	for i, e := range es {
		inc[e.u] = append(inc[e.u], i)
		inc[e.v] = append(inc[e.v], i)
	}
	seenNode := make([]bool, n)
	seenEdge := make([]bool, len(es))
	var out []layerComponent
	for s := range n {
		if seenNode[s] {
			continue
		}
		c := layerComponent{nodes: []int{s}}
		seenNode[s] = true
		for qi := 0; qi < len(c.nodes); qi++ {
			for _, i := range inc[c.nodes[qi]] {
				if !seenEdge[i] {
					seenEdge[i] = true
					c.edges = append(c.edges, i)
				}
				for _, w := range [2]int{es[i].u, es[i].v} {
					if !seenNode[w] {
						seenNode[w] = true
						c.nodes = append(c.nodes, w)
					}
				}
			}
		}
		out = append(out, c)
	}
	return out
}

// normalizeRanks shifts a component so its earliest layer is 0. Layers index
// slices downstream, so they must be non-negative, and each component reads
// best starting at the top.
func normalizeRanks(c layerComponent, rank []int) {
	low := rank[c.nodes[0]]
	for _, v := range c.nodes {
		if rank[v] < low {
			low = rank[v]
		}
	}
	for _, v := range c.nodes {
		rank[v] -= low
	}
}

// networkSimplex rewrites one component's ranks with ones of minimum total
// edge length, leaving them feasible throughout.
func networkSimplex(es []layerEdge, c layerComponent, rank []int) {
	slack := func(i int) int { return rank[es[i].v] - rank[es[i].u] - 1 }

	n := len(rank)
	inTree := make([]bool, n)
	tree := make([]bool, len(es))

	// growTight grows a maximal tree of tight (zero-slack) edges from the
	// component's first node, reporting how many nodes it reached.
	growTight := func() int {
		for _, v := range c.nodes {
			inTree[v] = false
		}
		for _, i := range c.edges {
			tree[i] = false
		}
		inTree[c.nodes[0]] = true
		count := 1
		for grew := true; grew; {
			grew = false
			for _, i := range c.edges {
				if tree[i] || slack(i) != 0 {
					continue
				}
				u, v := es[i].u, es[i].v
				if inTree[u] == inTree[v] {
					continue
				}
				if inTree[u] {
					inTree[v] = true
				} else {
					inTree[u] = true
				}
				tree[i], count, grew = true, count+1, true
			}
		}
		return count
	}

	// The tight edges alone may not span the component. Slide the tree along
	// the least-slack edge leaving it until they do; shifting a whole tree
	// keeps its own edges tight and cannot break any other constraint.
	for growTight() < len(c.nodes) {
		best, least := -1, rankInf
		for _, i := range c.edges {
			if inTree[es[i].u] == inTree[es[i].v] {
				continue
			}
			if s := slack(i); s < least {
				best, least = i, s
			}
		}
		if best < 0 {
			return // no edge bridges the gap; leave the feasible ranking alone
		}
		d := least
		if inTree[es[best].v] {
			d = -d
		}
		for _, v := range c.nodes {
			if inTree[v] {
				rank[v] += d
			}
		}
	}

	// Now optimise. A tree edge's cut value is the weight crossing its split in
	// its own direction minus the weight crossing back; where that is negative,
	// the total shortens by swapping it for an edge going the other way. The
	// iteration cap is belt and braces — cut values strictly decrease the
	// objective, so this terminates on its own.
	head := make([]bool, n)
	for iter := 0; iter < 4*len(c.edges)+16; iter++ {
		leave := -1
		for _, i := range c.edges {
			if !tree[i] {
				continue
			}
			markHeadSide(es, c, tree, i, head)
			if cutValue(es, c, head) < 0 {
				leave = i
				break
			}
		}
		if leave < 0 {
			break // optimal
		}
		// head still describes the split that leaving edge makes.
		enter, least := -1, rankInf
		for _, j := range c.edges {
			if tree[j] || !head[es[j].u] || head[es[j].v] {
				continue
			}
			if s := slack(j); s < least {
				enter, least = j, s
			}
		}
		if enter < 0 {
			break
		}
		tree[leave], tree[enter] = false, true
		retighten(es, c, tree, rank)
	}
}

// markHeadSide marks the nodes that stay attached to the head of tree edge
// skip once that edge is removed.
func markHeadSide(es []layerEdge, c layerComponent, tree []bool, skip int, head []bool) {
	for _, v := range c.nodes {
		head[v] = false
	}
	start := es[skip].v
	head[start] = true
	stack := []int{start}
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, i := range c.edges {
			if !tree[i] || i == skip {
				continue
			}
			var w int
			switch {
			case es[i].u == u:
				w = es[i].v
			case es[i].v == u:
				w = es[i].u
			default:
				continue
			}
			if !head[w] {
				head[w] = true
				stack = append(stack, w)
			}
		}
	}
}

// cutValue totals the edges crossing a split: those running into the head side
// count for it, those running back out count against.
func cutValue(es []layerEdge, c layerComponent, head []bool) int {
	cv := 0
	for _, i := range c.edges {
		switch {
		case !head[es[i].u] && head[es[i].v]:
			cv++
		case head[es[i].u] && !head[es[i].v]:
			cv--
		}
	}
	return cv
}

// retighten re-derives ranks from the spanning tree so every tree edge is
// tight again after a swap.
func retighten(es []layerEdge, c layerComponent, tree []bool, rank []int) {
	seen := make([]bool, len(rank))
	start := c.nodes[0]
	rank[start], seen[start] = 0, true
	stack := []int{start}
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, i := range c.edges {
			if !tree[i] {
				continue
			}
			switch {
			case es[i].u == u && !seen[es[i].v]:
				rank[es[i].v], seen[es[i].v] = rank[u]+1, true
				stack = append(stack, es[i].v)
			case es[i].v == u && !seen[es[i].u]:
				rank[es[i].u], seen[es[i].u] = rank[u]-1, true
				stack = append(stack, es[i].u)
			}
		}
	}
}
