package diagram

import "math"

// The gap between two layers is a routing channel. Every edge crossing it makes
// one sideways move, and if they all make it at the same flow coordinate an
// edge travelling far sideways cuts straight through the vertical stub of one
// that stops short — a crossing the graph never asked for. So the runs are
// ordered by which must sit above which and each is given its own track.
//
// A segment runs from an upper stop at cross a down to a lower stop at cross b:
// a vertical stub at a from the top of the band down to its track, the cross-run
// along the track from a to b, then a stub at b down to the bottom. j's run cuts
// i's upper stub only when it passes above i's turn, and i's lower stub only
// when it passes below, so for two segments i and j:
//
//	a_i strictly inside span_j ⇒ track_i ≤ track_j
//	b_i strictly inside span_j ⇒ track_i ≥ track_j
//
// Note those are not strict. Two runs on one track touch where a stub meets
// them and never cross, and that is the shape a fan wants: every edge leaving a
// port turns at the same height and drops off its bar one at a time. Ordering
// alone is what the drawing needs — the old routing settled each elbow from its
// own two endpoints, so an edge bound for a routing dummy (an 8px box against a
// 28px node) turned a few pixels lower than its neighbours and sliced through
// them on the way past.
//
// A pair does need its own tracks when it is constrained one way only and
// shares neither endpoint: their runs would otherwise overlap along a stretch
// that belongs to both, drawing two edges as one line. That is the case an edge
// and the back edge beside it fall into — different port lanes at both ends —
// so the jog that used to be hand-picked per direction now falls out of the
// same rule as everything else. Constrained both ways is the opposite: one run
// nested strictly inside another can only avoid cutting it by sharing its
// track, so those must stay equal.
//
// Two edges that genuinely have to cross make the constraints contradict, so
// the relation is a graph rather than an order and cycles have to be broken —
// dropping a constraint just accepts a crossing that was unavoidable anyway.

const (
	// align is the cross-axis offset below which a segment is drawn straight:
	// an elbow for a sub-pixel misalignment reads as a wobble, not a corner.
	align = 2.0

	// trackEps is how far inside a run's span a stub has to fall to count as
	// cutting it. Ports are placed by arithmetic on box widths, so runs that
	// mean to share an endpoint can miss it by a hair; without the slack that
	// near-miss would buy a whole extra track.
	trackEps = 0.5
)

// route is one edge's resolved geometry in cross space: where each stop sits on
// the cross axis, and which track carries the cross-run of each segment between
// them (-1 where the segment runs straight and needs none).
type route struct {
	cross []float64
	track []int
}

// channels is the resolved routing for the whole drawing.
type channels struct {
	routes []route
	count  []int       // tracks in use per gap, indexed by the upper layer
	flow   [][]float64 // flow coordinate of each track, filled in by assignFlow
}

// segment is one edge segment needing a cross-run in some gap.
type segment struct {
	ei, k int     // which edge, and which of its segments
	a, b  float64 // cross of the upper and lower stop
}

// planChannels resolves every edge's stops on the cross axis and allocates the
// gaps' tracks. It runs before the flow axis is settled, because both only need
// the cross coordinates Brandes–Köpf has already fixed — and because the number
// of tracks a gap ends up with decides how much room it needs.
func planChannels(verts []vertex, edgeChains [][]int, plan portPlan, layerCount int) channels {
	ch := channels{
		routes: make([]route, len(edgeChains)),
		count:  make([]int, max(layerCount-1, 0)),
	}
	perGap := make([][]segment, len(ch.count))

	for ei, chain := range edgeChains {
		cross := make([]float64, len(chain))
		for k, vi := range chain {
			cross[k] = verts[vi].cross
		}
		src, dst := verts[chain[0]], verts[chain[len(chain)-1]]
		ex, en := straighten(src.cross, dst.cross,
			plan.exit[ei], plan.entry[ei], plan.exitFree[ei], plan.entryFree[ei])
		cross[0] = src.cross + ex
		cross[len(cross)-1] = dst.cross + en

		cross, bends := resolveStops(cross)
		track := make([]int, len(bends))
		for k, bend := range bends {
			track[k] = -1
			if bend {
				g := verts[chain[k]].layer
				perGap[g] = append(perGap[g], segment{ei: ei, k: k, a: cross[k], b: cross[k+1]})
			}
		}
		ch.routes[ei] = route{cross: cross, track: track}
	}

	for g, segs := range perGap {
		for i, t := range allocate(segs) {
			ch.routes[segs[i].ei].track[segs[i].k] = t
			ch.count[g] = max(ch.count[g], t+1)
		}
	}
	return ch
}

// resolveStops fixes an edge's stops on the cross axis and reports which of its
// segments step across. Offsets under align are flattened rather than drawn as
// a corner, which moves the stop — including a port, whose arrowhead follows —
// so this has to settle before spans are measured: whether a segment bends at
// all, and how wide its run is, must read the same to the allocator and to the
// router.
func resolveStops(cross []float64) ([]float64, []bool) {
	out := snapStops(cross)
	bends := make([]bool, len(out)-1)
	for i := 1; i < len(out); i++ {
		if math.Abs(out[i]-out[i-1]) > align {
			bends[i-1] = true
		} else {
			out[i] = out[i-1]
		}
	}
	return out, bends
}

// snapStops flattens a long edge's routing dummies onto their neighbours where
// the offset between them is under minBend — too small to draw as a corner.
// Only the intermediate stops move; the first and last are the edge's ports and
// stay put, so arrowheads keep their place on the box.
func snapStops(cross []float64) []float64 {
	out := append([]float64(nil), cross...)
	if len(out) < 3 {
		return out
	}
	for i := 1; i < len(out)-1; i++ { // pull each dummy onto the one before it
		if math.Abs(out[i]-out[i-1]) <= minBend {
			out[i] = out[i-1]
		}
	}
	for i := len(out) - 2; i >= 1; i-- { // then let the fixed tail port win
		if math.Abs(out[i]-out[i+1]) <= minBend {
			out[i] = out[i+1]
		}
	}
	return out
}

// above is a constraint "u sits on a track at least step above v" — step 0 for
// an ordering that a shared track satisfies, 1 where the two need their own.
type above struct {
	u, v, step int
}

// allocate gives every segment in one gap a track, low tracks sitting nearer
// the upper layer. Segments nothing separates share track 0, which is what lets
// a fan read as a single bar rather than a staircase.
func allocate(segs []segment) []int {
	n := len(segs)
	span := func(s segment) (float64, float64) {
		return math.Min(s.a, s.b), math.Max(s.a, s.b)
	}
	inside := func(v float64, s segment) bool {
		lo, hi := span(s)
		return v > lo+trackEps && v < hi-trackEps
	}
	// A one-way constraint costs a track unless the two runs meet at a shared
	// port or dummy: those belong on one bar, the rest would overlap along a
	// stretch that is only theirs and read as a single line.
	step := func(i, j segment) int {
		if math.Abs(i.a-j.a) <= trackEps || math.Abs(i.b-j.b) <= trackEps {
			return 0
		}
		return 1
	}

	var cons []above
	for i := range n {
		for j := i + 1; j < n; j++ {
			iFirst := inside(segs[i].a, segs[j]) || inside(segs[j].b, segs[i])
			jFirst := inside(segs[j].a, segs[i]) || inside(segs[i].b, segs[j])
			switch {
			case iFirst && jFirst:
				// Nested: neither can turn clear of the other, so pin them level.
				cons = append(cons, above{i, j, 0}, above{j, i, 0})
			case iFirst:
				cons = append(cons, above{i, j, step(segs[i], segs[j])})
			case jFirst:
				cons = append(cons, above{j, i, step(segs[i], segs[j])})
			}
		}
	}
	return layerTracks(n, breakConstraintCycles(n, cons))
}

// breakConstraintCycles drops the constraints that close a cycle — those pairs
// cross whatever we do. Constraints are visited in the order they were built,
// which is by segment index, so the choice stays deterministic like every other
// phase of the pipeline.
func breakConstraintCycles(n int, cons []above) []above {
	out := make([][]above, n)
	for _, c := range cons {
		out[c.u] = append(out[c.u], c)
	}
	const (
		white = iota
		gray
		black
	)
	state := make([]int, n)
	var kept []above
	var dfs func(u int)
	dfs = func(u int) {
		state[u] = gray
		for _, c := range out[u] {
			if state[c.v] != gray {
				kept = append(kept, c)
			}
			if state[c.v] == white {
				dfs(c.v)
			}
		}
		state[u] = black
	}
	for u := range n {
		if state[u] == white {
			dfs(u)
		}
	}
	return kept
}

// layerTracks solves the surviving constraints by longest path: a segment's
// track is the tallest stack of segments that has to sit above it.
func layerTracks(n int, cons []above) []int {
	over := make([][]above, n)
	for _, c := range cons {
		over[c.v] = append(over[c.v], c)
	}
	track := make([]int, n)
	done := make([]bool, n)
	var depth func(v int) int
	depth = func(v int) int {
		if done[v] {
			return track[v]
		}
		done[v] = true // the constraints are acyclic, so this cannot recurse back
		d := 0
		for _, c := range over[v] {
			if e := depth(c.u) + c.step; e > d {
				d = e
			}
		}
		track[v] = d
		return d
	}
	for v := range n {
		depth(v)
	}
	return track
}
