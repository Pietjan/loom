package chart

import "github.com/pietjan/loom/internal/styles"

// Color selects a series palette. The first series defaults to Accent,
// later ones walk defaultOrder.
type Color string

const (
	Accent  Color = "accent"
	Indigo  Color = "indigo"
	Blue    Color = "blue"
	Emerald Color = "emerald"
	Amber   Color = "amber"
	Rose    Color = "rose"
	Violet  Color = "violet"
	Cyan    Color = "cyan"
)

var defaultOrder = []Color{Accent, Indigo, Blue, Emerald, Amber, Rose, Violet, Cyan}

// seriesStyle maps a palette color to the classes for each SVG part.
// Complete literals, so the Tailwind scanner sees them.
type seriesStyle struct {
	line string // stroke on the line path
	fill string // translucent fill on the area path / bars
	dot  string // point markers
	swat string // legend swatch
}

var palette = map[Color]seriesStyle{
	Accent: {
		line: "stroke-accent",
		fill: "fill-accent/15",
		dot:  "fill-accent",
		swat: "bg-accent",
	},
	Indigo: {
		line: "stroke-indigo-500",
		fill: "fill-indigo-500/15",
		dot:  "fill-indigo-500",
		swat: "bg-indigo-500",
	},
	Blue: {
		line: "stroke-blue-500",
		fill: "fill-blue-500/15",
		dot:  "fill-blue-500",
		swat: "bg-blue-500",
	},
	Emerald: {
		line: "stroke-emerald-500",
		fill: "fill-emerald-500/15",
		dot:  "fill-emerald-500",
		swat: "bg-emerald-500",
	},
	Amber: {
		line: "stroke-amber-500",
		fill: "fill-amber-500/15",
		dot:  "fill-amber-500",
		swat: "bg-amber-500",
	},
	Rose: {
		line: "stroke-rose-500",
		fill: "fill-rose-500/15",
		dot:  "fill-rose-500",
		swat: "bg-rose-500",
	},
	Violet: {
		line: "stroke-violet-500",
		fill: "fill-violet-500/15",
		dot:  "fill-violet-500",
		swat: "bg-violet-500",
	},
	Cyan: {
		line: "stroke-cyan-500",
		fill: "fill-cyan-500/15",
		dot:  "fill-cyan-500",
		swat: "bg-cyan-500",
	},
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
