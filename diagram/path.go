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

// Text/box sizing. There is no DOM to measure against on the server, so box
// widths come from a deterministic estimate: a fixed average glyph advance
// per rune plus padding. Approximate for proportional fonts and unwrapped in
// v1, but stable and reproducible — which is what golden tests require.
const (
	fontSize  = 14.0 // label font-size (matches nodeLabelClasses text-[14px])
	glyphAdv  = 8.0  // ~0.57em average advance
	padX      = 16.0 // horizontal box padding
	padY      = 11.0 // vertical box padding
	minNodeW  = 56.0 // floor so short labels aren't cramped
	dummySize = 8.0  // routing-dummy box extent
)

// nodeSize estimates the drawn box size for a node's label and shape.
func nodeSize(n node) (w, h float64) {
	w = float64(len([]rune(n.label)))*glyphAdv + 2*padX
	if w < minNodeW {
		w = minNodeW
	}
	h = fontSize + 2*padY
	if n.shape == diamond {
		// Text inscribed in a rhombus needs the box grown to keep the label
		// clear of the slanted edges.
		w *= 1.5
		h *= 1.7
	}
	return w, h
}

// fmtCoord renders an SVG coordinate with sub-pixel precision but no
// float-drift noise (mirrors chart's fmtCoord).
func fmtCoord(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
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
	const length, halfWidth = 9.0, 5.0
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
