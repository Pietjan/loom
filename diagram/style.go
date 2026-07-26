package diagram

import (
	"github.com/pietjan/loom/internal/palette"
	"github.com/pietjan/loom/internal/styles"
)

// outlineStroke resolves a node's accent color to its outline classes. The
// three cases are exhaustive and each contributes exactly one stroke
// utility, so no node ever wears conflicting stroke-* classes whose CSS
// source order would decide the winner.
func outlineStroke(c Color) string {
	if c == ColorAccent {
		return "stroke-accent" // a theme token, with no palette row
	}
	if s, ok := palette.Of(palette.Color(c)); ok {
		return s.Stroke
	}
	return "stroke-base-300 dark:stroke-base-600"
}

// rootClasses styles the stage: a positioned box of the diagram's natural
// size that the SVG layer and the node bodies are placed inside.
func rootClasses() string {
	var b styles.Builder
	b.Add("relative")
	return b.String()
}

// canvasClasses pins the edge/chrome SVG to the stage, behind the bodies.
func canvasClasses() string {
	var b styles.Builder
	b.Add("absolute inset-0 overflow-visible")
	return b.String()
}

// Strokes are 1px to match loom's borders, which are `border` everywhere.
func nodeShapeClasses(c Color) string {
	var b styles.Builder
	b.Add("fill-white stroke-1 dark:fill-base-800")
	b.Add(outlineStroke(c))
	return b.String()
}

// contentClasses positions and centers a node's body over its box. Bare nodes
// get only layout - their body brings its own chrome and type.
func contentClasses(bare bool) string {
	var b styles.Builder
	b.Add("absolute flex items-center justify-center text-center leading-tight")
	b.If(!bare, "px-2 text-xs font-medium text-base-700 dark:text-base-100")
	return b.String()
}

// Connectors are stroke-2, matching chart's data lines: an edge is content
// (the relationship) rather than chrome, and the edge dot is sized against a
// 2px line the same way chart's data points are. Node outlines stay at 1px
// like loom's borders, giving the same content-over-chrome weighting chart
// uses for its lines over its grid.
func edgeClasses() string {
	var b styles.Builder
	b.Add("fill-none stroke-base-300 stroke-2 dark:stroke-base-600")
	return b.String()
}

// dashPattern returns the SVG stroke-dasharray (and linecap) for a line style;
// an empty pattern means a solid line, drawn without either attribute. The
// values are tuned against the 2px edge stroke: a clearly broken dash, and
// round dots produced by zero-length segments under a round cap.
func dashPattern(s lineStyle) (dash, cap string) {
	switch s {
	case lineDashed:
		return "6 4", ""
	case lineDotted:
		return "0 6", "round"
	default:
		return "", ""
	}
}

func arrowClasses() string {
	var b styles.Builder
	b.Add("fill-base-300 dark:fill-base-600")
	return b.String()
}

// arrowHollowClasses styles an outlined arrowhead: filled with the surface
// colour so the shaft can't show through, outlined in the line's colour.
func arrowHollowClasses() string {
	var b styles.Builder
	b.Add("fill-white stroke-1 stroke-base-300 dark:fill-base-800 dark:stroke-base-600")
	return b.String()
}

// arrowOpenClasses styles an open barb: no fill, stroked like the line it caps.
func arrowOpenClasses() string {
	var b styles.Builder
	b.Add("fill-none stroke-2 stroke-base-300 dark:stroke-base-600")
	return b.String()
}

// edgeDotClasses styles the marker sitting on a labelled edge, following
// chart's data points: a solid dot in the line's own colour, ringed in the
// surface colour so it reads as sitting on the line rather than crossing it.
// It is a little under chart's r=3.5 point, which suits a marker you hover for
// a label rather than a value to read off the line.
func edgeDotClasses() string {
	var b styles.Builder
	b.Add("fill-base-300 stroke-white dark:fill-base-600 dark:stroke-base-800")
	return b.String()
}

// edgeHitClasses styles the target sitting over that dot. It is transparent -
// the dot itself is painted in the canvas - and exists to carry the hover and
// focus the tooltip opens on, so it is sized a little larger than the dot to
// stay comfortable to hit. Focus has to show here rather than on the dot,
// which is in the SVG and out of this element's reach.
func edgeHitClasses() string {
	var b styles.Builder
	b.Add("block size-2.5 rounded-full")
	b.Add("focus-visible:outline-2 focus-visible:outline-current")
	return b.String()
}
