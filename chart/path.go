package chart

import "strings"

// xy is a point in SVG user units.
type xy struct {
	x, y float64
}

// linePath builds a straight polyline path ("M x y L x y ...").
func linePath(pts []xy) string {
	var b strings.Builder
	for i, p := range pts {
		if i == 0 {
			b.WriteString("M ")
		} else {
			b.WriteString(" L ")
		}
		b.WriteString(fmtCoord(p.x))
		b.WriteString(" ")
		b.WriteString(fmtCoord(p.y))
	}
	return b.String()
}

// smoothPath builds a Catmull-Rom spline through the points, converted to
// cubic Béziers — passes through every data point, no overshoot drama at
// typical dashboard densities.
func smoothPath(pts []xy) string {
	if len(pts) < 3 {
		return linePath(pts)
	}
	var b strings.Builder
	b.WriteString("M ")
	b.WriteString(fmtCoord(pts[0].x))
	b.WriteString(" ")
	b.WriteString(fmtCoord(pts[0].y))

	for i := 0; i < len(pts)-1; i++ {
		p0 := pts[max(i-1, 0)]
		p1 := pts[i]
		p2 := pts[i+1]
		p3 := pts[min(i+2, len(pts)-1)]

		c1 := xy{p1.x + (p2.x-p0.x)/6, p1.y + (p2.y-p0.y)/6}
		c2 := xy{p2.x - (p3.x-p1.x)/6, p2.y - (p3.y-p1.y)/6}

		b.WriteString(" C ")
		b.WriteString(fmtCoord(c1.x))
		b.WriteString(" ")
		b.WriteString(fmtCoord(c1.y))
		b.WriteString(", ")
		b.WriteString(fmtCoord(c2.x))
		b.WriteString(" ")
		b.WriteString(fmtCoord(c2.y))
		b.WriteString(", ")
		b.WriteString(fmtCoord(p2.x))
		b.WriteString(" ")
		b.WriteString(fmtCoord(p2.y))
	}
	return b.String()
}

// areaPath closes a line path down to the baseline for the filled region.
func areaPath(line string, first, last xy, baseline float64) string {
	var b strings.Builder
	b.WriteString(line)
	b.WriteString(" L ")
	b.WriteString(fmtCoord(last.x))
	b.WriteString(" ")
	b.WriteString(fmtCoord(baseline))
	b.WriteString(" L ")
	b.WriteString(fmtCoord(first.x))
	b.WriteString(" ")
	b.WriteString(fmtCoord(baseline))
	b.WriteString(" Z")
	return b.String()
}
