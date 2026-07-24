package diagram

import "github.com/pietjan/loom/internal/styles"

// toneStroke maps a node tone to its complete outline classes. The default
// tone is in the map too, so exactly one stroke utility is emitted per node
// (no conflicting stroke-* classes whose CSS source order would decide the
// winner). Complete literals, so the Tailwind scanner sees them.
var toneStroke = map[Tone]string{
	ToneDefault: "stroke-base-300 dark:stroke-base-600",
	ToneAccent:  "stroke-accent",
	ToneIndigo:  "stroke-indigo-400 dark:stroke-indigo-500",
	ToneEmerald: "stroke-emerald-400 dark:stroke-emerald-500",
	ToneAmber:   "stroke-amber-400 dark:stroke-amber-500",
	ToneRose:    "stroke-rose-400 dark:stroke-rose-500",
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
func nodeShapeClasses(t Tone) string {
	var b styles.Builder
	b.Add("fill-white stroke-1 dark:fill-base-800")
	styles.Match(&b, t, toneStroke)
	return b.String()
}

// contentClasses positions and centers a node's body over its box. Bare nodes
// get only layout — their body brings its own chrome and type.
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

func arrowClasses() string {
	var b styles.Builder
	b.Add("fill-base-300 dark:fill-base-600")
	return b.String()
}

// dotClasses styles the marker sitting on a labelled edge, matching chart's
// data points: a solid dot in the line's own colour, ringed in the surface
// colour so it reads as sitting on the line rather than crossing it. Chart
// draws r=3.5 with a 1.5 stroke; this is the HTML equivalent.
func dotClasses() string {
	var b styles.Builder
	b.Add("block size-[7px] rounded-full bg-base-300 ring-[1.5px] ring-white")
	b.Add("dark:bg-base-600 dark:ring-base-800")
	return b.String()
}
