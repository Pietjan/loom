package sankey

import (
	"github.com/pietjan/loom/internal/palette"
	"github.com/pietjan/loom/internal/styles"
)

// Color selects a node's hue, and with it the hue of every ribbon leaving
// it: the theme accent, or any color from the shared palette (the same set
// badge offers). Nodes with no color of their own walk defaultOrder.
type Color string

const (
	ColorAccent  Color = "accent"
	ColorZinc    Color = "zinc"
	ColorRed     Color = "red"
	ColorOrange  Color = "orange"
	ColorAmber   Color = "amber"
	ColorYellow  Color = "yellow"
	ColorLime    Color = "lime"
	ColorGreen   Color = "green"
	ColorEmerald Color = "emerald"
	ColorTeal    Color = "teal"
	ColorCyan    Color = "cyan"
	ColorSky     Color = "sky"
	ColorBlue    Color = "blue"
	ColorIndigo  Color = "indigo"
	ColorViolet  Color = "violet"
	ColorPurple  Color = "purple"
	ColorFuchsia Color = "fuchsia"
	ColorPink    Color = "pink"
	ColorRose    Color = "rose"
)

// defaultOrder is the walk for uncolored nodes - the same hand-picked
// subset chart uses for its series: eight hues far enough apart to stay
// distinguishable beside each other, which the full eighteen are not.
var defaultOrder = []Color{ColorAccent, ColorIndigo, ColorBlue, ColorEmerald, ColorAmber, ColorRose, ColorViolet, ColorCyan}

// A sankey is two kinds of mark in one hue, which is the weighting chart
// gives an area chart: the bar is the mark, at the palette's mark strength,
// and the ribbon is a surface under it, at the palette's Wash. Ribbons
// cross and overlap, and a drawing holds far more of them than a chart
// holds area fills, so the translucent row is what keeps a graph of any
// size from turning into a block of color - and it is a row of the shared
// palette rather than an opacity this component invented.

// barFill resolves a color to the one fill utility a node's bar paints
// with. The three cases are exhaustive and each contributes exactly one
// fill, so no mark ever wears conflicting fill-* classes whose CSS source
// order would decide the winner. The hue comes from the shared palette at
// mark strength, so a sankey bar, a chart point and a solid badge of one
// hue are the same color.
func barFill(c Color) string {
	if c == ColorAccent {
		return "fill-accent" // a theme token, with no palette row
	}
	if s, ok := palette.Of(palette.Color(c)); ok {
		return s.Fill
	}
	return "fill-base-300 dark:fill-base-600"
}

// stopColor resolves a color to the utility that paints a gradient stop.
// A ribbon runs from its source node's hue into its target's, so neither
// hue can sit in a fill utility on the path - the pair lives on the two
// stops instead, and the path carries only the alpha.
func stopColor(c Color) string {
	if c == ColorAccent {
		return "[stop-color:var(--color-accent)]" // a theme token, with no palette row
	}
	if s, ok := palette.Of(palette.Color(c)); ok {
		return s.Stop
	}
	return "[stop-color:var(--color-base-300)] dark:[stop-color:var(--color-base-600)]"
}

// washAlpha is what the palette's Wash row carries, said as an opacity
// because a gradient's hues arrive through its stops rather than through a
// fill utility that could carry the alpha with them. A ribbon is the same
// strength as the area chart fill it is modelled on.
const washAlpha = "0.15"

func rootClasses() string {
	var b styles.Builder
	b.Add("w-full")
	return b.String()
}

func svgClasses() string {
	var b styles.Builder
	b.Add("w-full h-auto")
	return b.String()
}

// nodeClasses styles a bar, following chart's data point exactly: the hue
// at mark strength, ringed in the surface color so the ribbons read as
// stopping short of it rather than running into it. The ring is why a
// point sits on a line without the line crossing it, and it does the same
// job here.
func nodeClasses(c Color) string {
	var b styles.Builder
	b.Add("stroke-white dark:stroke-base-800")
	b.Add(barFill(c))
	return b.String()
}

// stopClasses styles one end of a ribbon's gradient.
func stopClasses(c Color) string {
	var b styles.Builder
	b.Add(stopColor(c))
	return b.String()
}

// labelClasses is chart's tick-label treatment. A node's name sits at the
// same weight as a value read off an axis, which is what it is, and at this
// weight it stays legible over a washed ribbon without a halo behind it.
func labelClasses() string {
	var b styles.Builder
	b.Add("text-[10px] fill-base-400 select-none")
	return b.String()
}
