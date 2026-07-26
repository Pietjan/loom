package timeline

import (
	"github.com/pietjan/loom/internal/palette"
	"github.com/pietjan/loom/internal/styles"
)

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
	tinted, isHue := palette.Of(palette.Color(c.Color))
	switch {
	case isHue:
		// An explicit hue wins over the status tint: the same chip
		// recipe a badge of that color wears.
		b.Add(tinted.Tint)
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
