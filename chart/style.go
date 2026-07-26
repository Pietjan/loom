package chart

import (
	"github.com/pietjan/loom/internal/palette"
	"github.com/pietjan/loom/internal/styles"
)

// Color selects a series color: the theme accent, or any hue from the
// shared palette (the same set badge offers). The first series defaults
// to ColorAccent, later ones walk defaultOrder.
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

// defaultOrder is the walk for unnamed series - a hand-picked subset, not
// the whole palette: eight hues far enough apart to stay distinguishable
// beside each other, which the full eighteen are not.
var defaultOrder = []Color{ColorAccent, ColorIndigo, ColorBlue, ColorEmerald, ColorAmber, ColorRose, ColorViolet, ColorCyan}

// seriesStyle is the set of classes one series wears, one per SVG part.
type seriesStyle struct {
	line string // stroke on the line path
	fill string // translucent fill on the area path
	dot  string // point markers, bars
	swat string // legend swatch
}

// accentStyle is the theme accent, which is a token rather than a hue and
// so has no row in the shared palette.
var accentStyle = seriesStyle{
	line: "stroke-accent",
	fill: "fill-accent/15",
	dot:  "fill-accent",
	swat: "bg-accent",
}

// styleFor resolves a series color to its classes. The hue rows come from
// the shared palette at mark strength, so a chart line, a progress bar and
// a solid badge of one hue are the same color.
func styleFor(c Color) seriesStyle {
	s, ok := palette.Of(palette.Color(c))
	if !ok {
		return accentStyle
	}
	return seriesStyle{line: s.Stroke, fill: s.Wash, dot: s.Fill, swat: s.Bg}
}

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

func lineClasses(s seriesStyle) string {
	var b styles.Builder
	b.Add("fill-none stroke-2")
	b.Add(s.line)
	return b.String()
}

func gridClasses() string {
	var b styles.Builder
	b.Add("stroke-base-200 dark:stroke-base-600")
	return b.String()
}

func tickLabelClasses() string {
	var b styles.Builder
	b.Add("text-[10px] fill-base-400 select-none")
	return b.String()
}

func dotClasses(s seriesStyle) string {
	var b styles.Builder
	b.Add("stroke-white dark:stroke-base-800")
	b.Add(s.dot)
	return b.String()
}

func legendClasses() string {
	var b styles.Builder
	b.Add("mt-2 flex flex-wrap items-center gap-4 text-sm text-base-500 dark:text-base-300")
	return b.String()
}

func legendSwatchClasses(s seriesStyle) string {
	var b styles.Builder
	b.Add("size-2.5 rounded-full")
	b.Add(s.swat)
	return b.String()
}
