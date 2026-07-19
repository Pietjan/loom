package timeline

import "github.com/pietjan/loom/internal/styles"

// Tinted circle palette — the same recipe as badge, so colored indicators
// and badges on one page read as a single system.
var colorClasses = map[Color]string{
	ColorZinc:    "bg-zinc-400/15 text-zinc-700 dark:bg-zinc-400/25 dark:text-zinc-300",
	ColorRed:     "bg-red-400/15 text-red-700 dark:bg-red-400/25 dark:text-red-300",
	ColorOrange:  "bg-orange-400/15 text-orange-700 dark:bg-orange-400/25 dark:text-orange-300",
	ColorAmber:   "bg-amber-400/20 text-amber-800 dark:bg-amber-400/25 dark:text-amber-300",
	ColorYellow:  "bg-yellow-400/20 text-yellow-800 dark:bg-yellow-400/25 dark:text-yellow-200",
	ColorLime:    "bg-lime-400/20 text-lime-800 dark:bg-lime-400/25 dark:text-lime-300",
	ColorGreen:   "bg-green-400/15 text-green-800 dark:bg-green-400/25 dark:text-green-300",
	ColorEmerald: "bg-emerald-400/15 text-emerald-800 dark:bg-emerald-400/25 dark:text-emerald-300",
	ColorTeal:    "bg-teal-400/15 text-teal-800 dark:bg-teal-400/25 dark:text-teal-300",
	ColorCyan:    "bg-cyan-400/15 text-cyan-800 dark:bg-cyan-400/25 dark:text-cyan-300",
	ColorSky:     "bg-sky-400/15 text-sky-800 dark:bg-sky-400/25 dark:text-sky-300",
	ColorBlue:    "bg-blue-400/15 text-blue-800 dark:bg-blue-400/25 dark:text-blue-300",
	ColorIndigo:  "bg-indigo-400/15 text-indigo-700 dark:bg-indigo-400/25 dark:text-indigo-300",
	ColorViolet:  "bg-violet-400/15 text-violet-700 dark:bg-violet-400/25 dark:text-violet-300",
	ColorPurple:  "bg-purple-400/15 text-purple-700 dark:bg-purple-400/25 dark:text-purple-300",
	ColorFuchsia: "bg-fuchsia-400/15 text-fuchsia-700 dark:bg-fuchsia-400/25 dark:text-fuchsia-300",
	ColorPink:    "bg-pink-400/15 text-pink-700 dark:bg-pink-400/25 dark:text-pink-300",
	ColorRose:    "bg-rose-400/15 text-rose-700 dark:bg-rose-400/25 dark:text-rose-300",
}

func listClasses(c Config) string {
	var b styles.Builder
	b.Add("flex")
	b.If(c.horizontal, "flex-row")
	b.If(!c.horizontal, "flex-col")
	return b.String()
}

func itemClasses(c Config) string {
	var b styles.Builder
	// group/tl lets the connector segment and content padding switch off
	// on the last item.
	b.Add("group/tl grid")
	b.If(!c.horizontal, "grid-cols-[auto_1fr] gap-x-3")
	b.If(c.horizontal, "min-w-0 flex-1 grid-rows-[auto_1fr] gap-y-3")
	return b.String()
}

// railClasses is the indicator column/row: circle first, connector after.
func railClasses(c Config) string {
	var b styles.Builder
	b.Add("flex items-center")
	b.If(!c.horizontal, "flex-col")
	return b.String()
}

func circleClasses(c Config) string {
	var b styles.Builder
	if c.bare {
		b.Add("flex shrink-0 items-center justify-center")
		return b.String()
	}
	b.Add("flex shrink-0 items-center justify-center rounded-full font-medium")
	// The circle runs a size step ahead of its glyph so icons get a ring of
	// breathing room rather than filling the disc edge to edge.
	styles.Match(&b, c.Size, map[Size]string{
		SizeBase:  "size-8 text-xs **:data-[ui=icon]:size-5",
		SizeLarge: "size-10 text-sm **:data-[ui=icon]:size-6",
	})
	switch {
	case c.Color != "":
		styles.Match(&b, c.Color, colorClasses)
	case c.Status == StatusComplete:
		b.Add("bg-accent text-accent-foreground")
	case c.Status == StatusCurrent:
		b.Add("bg-accent text-accent-foreground ring-4 ring-accent/25")
	case c.Status == StatusIncomplete:
		b.Add("bg-base-100 text-base-400 dark:bg-base-700 dark:text-base-500")
	default:
		b.Add("bg-base-100 text-base-500 dark:bg-base-700 dark:text-base-300")
	}
	return b.String()
}

// lineClasses is the connector segment from this indicator to the next
// item; the last item's segment disappears.
func lineClasses(c Config) string {
	var b styles.Builder
	b.Add("group-last/tl:hidden")
	b.If(!c.horizontal, "w-px grow")
	b.If(c.horizontal, "h-px grow")
	b.If(c.Status == StatusComplete, "bg-accent")
	b.If(c.Status != StatusComplete, "bg-base-200 dark:bg-base-600")
	return b.String()
}

func dotClasses() string {
	var b styles.Builder
	b.Add("size-2 rounded-full bg-current")
	return b.String()
}

func contentClasses(c Config) string {
	var b styles.Builder
	b.Add("min-w-0")
	if c.horizontal {
		b.Add("pe-4 group-last/tl:pe-0")
		return b.String()
	}
	b.Add("pb-6 group-last/tl:pb-0")
	// Nudge the first text line level with the indicator circle.
	styles.Match(&b, c.Size, map[Size]string{
		SizeBase:  "pt-1.5",
		SizeLarge: "pt-2.5",
	})
	return b.String()
}
