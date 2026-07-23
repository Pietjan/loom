package diagram

import (
	"math"
	"strconv"
	"strings"
)

// xy is a point in SVG user units.
type xy struct {
	x, y float64
}

// Node box metrics, kept on loom's compact scale: text-xs labels in a box
// about as tall as a small button (h-8), padded like a medium badge.
const (
	fontSize  = 12.0 // label font-size (matches content text-xs)
	padX      = 12.0 // horizontal box padding
	padY      = 8.0  // vertical box padding
	minNodeW  = 48.0 // floor so short labels aren't cramped
	dummySize = 8.0  // routing-dummy box extent
)

// fmtCoord renders an SVG coordinate with sub-pixel precision but no
// float-drift noise (mirrors chart's fmtCoord).
func fmtCoord(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

// cornerRadius rounds the bends in a routed edge, matching the node radius.
const cornerRadius = 6.0

// roundedPath draws a polyline with its corners rounded: each interior vertex
// becomes a quadratic curve whose control point is the corner itself, pulled
// back along both segments by the radius (clamped so short segments can't
// overshoot). A two-point path is just a line.
func roundedPath(pts []xy, r float64) string {
	pts = simplify(dedupe(pts))
	if len(pts) < 3 {
		return polyline(pts)
	}
	var b strings.Builder
	b.WriteString("M ")
	writePoint(&b, pts[0])
	for i := 1; i < len(pts)-1; i++ {
		prev, corner, next := pts[i-1], pts[i], pts[i+1]
		rr := math.Min(r, math.Min(dist(prev, corner), dist(corner, next))/2)
		b.WriteString(" L ")
		writePoint(&b, toward(corner, prev, rr))
		b.WriteString(" Q ")
		writePoint(&b, corner)
		b.WriteString(" ")
		writePoint(&b, toward(corner, next, rr))
	}
	b.WriteString(" L ")
	writePoint(&b, pts[len(pts)-1])
	return b.String()
}

// toward returns the point d units from p in the direction of t.
func toward(p, t xy, d float64) xy {
	l := dist(p, t)
	if l == 0 {
		return p
	}
	return xy{p.x + (t.x-p.x)/l*d, p.y + (t.y-p.y)/l*d}
}

func dist(a, b xy) float64 { return math.Hypot(b.x-a.x, b.y-a.y) }

// simplify drops interior points the path runs straight through — a routing
// waypoint that introduces no turn would otherwise be rounded into a
// pointless curve.
func simplify(pts []xy) []xy {
	if len(pts) < 3 {
		return pts
	}
	out := []xy{pts[0]}
	for i := 1; i < len(pts)-1; i++ {
		a, p, b := out[len(out)-1], pts[i], pts[i+1]
		cross := (p.x-a.x)*(b.y-p.y) - (p.y-a.y)*(b.x-p.x)
		if math.Abs(cross) > 0.01 {
			out = append(out, p)
		}
	}
	return append(out, pts[len(pts)-1])
}

// dedupe drops consecutive duplicate points, which would otherwise produce
// zero-length segments and degenerate corners.
func dedupe(pts []xy) []xy {
	out := pts[:0:0]
	for i, p := range pts {
		if i == 0 || dist(out[len(out)-1], p) > 0.01 {
			out = append(out, p)
		}
	}
	return out
}

func writePoint(b *strings.Builder, p xy) {
	b.WriteString(fmtCoord(p.x))
	b.WriteByte(' ')
	b.WriteString(fmtCoord(p.y))
}

// polyline builds a straight multi-segment path ("M x y L x y ...").
func polyline(pts []xy) string {
	var b strings.Builder
	for i, p := range pts {
		if i == 0 {
			b.WriteString("M ")
		} else {
			b.WriteString(" L ")
		}
		b.WriteString(fmtCoord(p.x))
		b.WriteByte(' ')
		b.WriteString(fmtCoord(p.y))
	}
	return b.String()
}

// borderPoint returns where the segment from box center toward t crosses the
// box's rectangular border, so an edge meets the box edge instead of its center.
func borderPoint(cx, cy, w, h float64, t xy) xy {
	dx, dy := t.x-cx, t.y-cy
	if dx == 0 && dy == 0 {
		return xy{cx, cy}
	}
	s := math.Inf(1)
	if dx != 0 {
		s = math.Min(s, (w/2)/math.Abs(dx))
	}
	if dy != 0 {
		s = math.Min(s, (h/2)/math.Abs(dy))
	}
	return xy{cx + dx*s, cy + dy*s}
}

// arrowhead returns the three-point polygon (tip + two base corners) for an
// arrow at the end of the polyline, pointing along its final segment.
func arrowhead(pts []xy) string {
	const length, halfWidth = 8.0, 4.0
	tip := pts[len(pts)-1]
	prev := pts[len(pts)-2]
	dx, dy := tip.x-prev.x, tip.y-prev.y
	d := math.Hypot(dx, dy)
	if d == 0 {
		d = 1
	}
	ux, uy := dx/d, dy/d                       // unit vector along the segment
	bx, by := tip.x-ux*length, tip.y-uy*length // base center
	px, py := -uy, ux                          // perpendicular
	return strings.Join([]string{
		fmtCoord(tip.x) + "," + fmtCoord(tip.y),
		fmtCoord(bx+px*halfWidth) + "," + fmtCoord(by+py*halfWidth),
		fmtCoord(bx-px*halfWidth) + "," + fmtCoord(by-py*halfWidth),
	}, " ")
}

// diamondPoints returns the four-point polygon for a decision node.
func diamondPoints(cx, cy, w, h float64) string {
	return strings.Join([]string{
		fmtCoord(cx) + "," + fmtCoord(cy-h/2), // top
		fmtCoord(cx+w/2) + "," + fmtCoord(cy), // right
		fmtCoord(cx) + "," + fmtCoord(cy+h/2), // bottom
		fmtCoord(cx-w/2) + "," + fmtCoord(cy), // left
	}, " ")
}

// midpoint returns the point halfway along a polyline by arc length, used to
// place an edge label.
func midpoint(pts []xy) xy {
	total := 0.0
	for i := 1; i < len(pts); i++ {
		total += math.Hypot(pts[i].x-pts[i-1].x, pts[i].y-pts[i-1].y)
	}
	half := total / 2
	acc := 0.0
	for i := 1; i < len(pts); i++ {
		seg := math.Hypot(pts[i].x-pts[i-1].x, pts[i].y-pts[i-1].y)
		if acc+seg >= half && seg > 0 {
			t := (half - acc) / seg
			return xy{pts[i-1].x + (pts[i].x-pts[i-1].x)*t, pts[i-1].y + (pts[i].y-pts[i-1].y)*t}
		}
		acc += seg
	}
	return pts[len(pts)/2]
}
