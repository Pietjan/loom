package sankey

import (
	"fmt"
	"math"
	"strconv"
)

// ribbonPath builds a flow's closed outline: a cubic along each edge of the
// band, joined across the node faces. Both control points sit at the
// horizontal midpoint, which gives the S-curve - and puts the curve's own
// midpoint at the mean of its ends, where the hover target goes.
func ribbonPath(x0, x1, sy, ty, w float64) string {
	xm := fmtCoord((x0 + x1) / 2)
	return fmt.Sprintf("M %s %s C %s %s, %s %s, %s %s L %s %s C %s %s, %s %s, %s %s Z",
		fmtCoord(x0), fmtCoord(sy),
		xm, fmtCoord(sy), xm, fmtCoord(ty), fmtCoord(x1), fmtCoord(ty),
		fmtCoord(x1), fmtCoord(ty+w),
		xm, fmtCoord(ty+w), xm, fmtCoord(sy+w), fmtCoord(x0), fmtCoord(sy+w))
}

// formatValue renders a value compactly: integers without decimals, large
// values as 1.2k / 3.4M.
func formatValue(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 1_000_000:
		return trimZero(v/1_000_000) + "M"
	case abs >= 1_000:
		return trimZero(v/1_000) + "k"
	default:
		return trimZero(v)
	}
}

func trimZero(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}

// fmtCoord renders an SVG coordinate with sub-pixel precision but no
// float-drift noise.
func fmtCoord(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}
