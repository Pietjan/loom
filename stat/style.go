package stat

import "github.com/pietjan/loom/internal/styles"

func classes(background bool) string {
	var b styles.Builder
	b.Add("flex flex-col gap-1 rounded-lg border border-base-200 bg-white p-5 shadow-xs")
	b.Add("dark:border-base-600 dark:bg-base-700")
	// Only when there is a background layer: it is positioned against the
	// tile and clipped by its rounded corners. Unconditional overflow
	// clipping would trap a tooltip or menu dropped into the tile.
	b.If(background, "relative overflow-hidden")
	return b.String()
}

// backgroundClasses positions the decorative layer: full-bleed along the
// bottom edge, its height coming from the component's own aspect ratio
// (a chart.Sparkline() is 4:1, so a quarter of the tile's width - pass
// chart.Size to make the band shorter). The mask fades it out upwards so
// a peak crossing the value stays behind the text rather than through it.
func backgroundClasses() string {
	var b styles.Builder
	b.Add("pointer-events-none absolute inset-x-0 bottom-0 opacity-70")
	b.Add("mask-t-from-50%")
	return b.String()
}

// The label and the value row take `relative` only when a background
// layer exists: a positioned element paints over in-flow content whatever
// the source order, so the content has to be positioned to stay on top.
func labelClasses(background bool) string {
	var b styles.Builder
	b.Add("text-sm font-medium text-base-500 dark:text-base-400")
	b.If(background, "relative")
	return b.String()
}

func rowClasses(background bool) string {
	var b styles.Builder
	b.Add("flex items-center gap-2")
	b.If(background, "relative")
	return b.String()
}

func valueClasses() string {
	var b styles.Builder
	b.Add("text-2xl font-semibold tracking-tight text-base-800 dark:text-white")
	return b.String()
}
