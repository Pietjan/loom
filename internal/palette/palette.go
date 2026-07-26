// Package palette holds the one curated set of hue shades that every loom
// component tints with.
//
// The shades started in badge, where each row was picked by eye rather
// than derived: which tint opacity a hue needs to read at all (amber,
// yellow and lime carry more than the rest), which text step stays legible
// on it, and which hues invert their foreground when filled (amber and
// yellow are light enough that white text fails on them). Copying that
// table into each component is how the shades drift apart, so it lives
// here once and components look their hue up.
//
// Roles, not hues, are what a component chooses between:
//
//   - Tint is a translucent surface with a foreground that reads on it -
//     a badge chip, a timeline indicator.
//   - Solid fills with the hue and puts a foreground on top - a solid
//     badge.
//   - Bg, Stroke, Fill and Wash paint the hue itself, for elements that
//     ARE the color rather than sitting on it - a progress bar, a chart
//     line, an area under it.
//
// Class strings are complete literals here, as everywhere: the Tailwind
// scanner reads source text, so a hue assembled at runtime compiles to
// nothing.
package palette

// Color names a hue. The values match the Color enums components expose,
// so a component converts its own to this one rather than mapping.
type Color string

// The palette. Zinc is the neutral; the rest run the color wheel in
// Tailwind's order.
const (
	Zinc    Color = "zinc"
	Red     Color = "red"
	Orange  Color = "orange"
	Amber   Color = "amber"
	Yellow  Color = "yellow"
	Lime    Color = "lime"
	Green   Color = "green"
	Emerald Color = "emerald"
	Teal    Color = "teal"
	Cyan    Color = "cyan"
	Sky     Color = "sky"
	Blue    Color = "blue"
	Indigo  Color = "indigo"
	Violet  Color = "violet"
	Purple  Color = "purple"
	Fuchsia Color = "fuchsia"
	Pink    Color = "pink"
	Rose    Color = "rose"
)

// All lists the hues in palette order, for tests and documentation that
// walk the set.
var All = []Color{
	Zinc, Red, Orange, Amber, Yellow, Lime, Green, Emerald, Teal,
	Cyan, Sky, Blue, Indigo, Violet, Purple, Fuchsia, Pink, Rose,
}

// Shades are one hue's class recipes, one per role.
type Shades struct {
	Tint   string // translucent surface + foreground
	Solid  string // filled surface + foreground
	Bg     string // the hue as a background, at mark strength
	Stroke string // the hue as an SVG stroke
	Fill   string // the hue as an SVG fill
	Wash   string // Fill, transparent enough to sit under a line
}

// Of returns a hue's shades. The bool reports whether the color is one of
// ours, which lets a caller fall through to a component-specific default
// (an accent-colored series, an untinted indicator) rather than paint a
// blank.
func Of(c Color) (Shades, bool) {
	s, ok := shades[c]
	return s, ok
}

// Mark strength is the 500 step the Solid row fills with in light mode.
// Only yellow is carried over brighter for dark, the one hue Solid also
// treats that way - the rest read on either background at 500.
var shades = map[Color]Shades{
	Zinc: {
		Tint:   "bg-zinc-400/15 text-zinc-700 dark:bg-zinc-400/25 dark:text-zinc-300",
		Solid:  "text-white dark:text-white bg-zinc-600 dark:bg-zinc-600",
		Bg:     "bg-zinc-500",
		Stroke: "stroke-zinc-500",
		Fill:   "fill-zinc-500",
		Wash:   "fill-zinc-500/15",
	},
	Red: {
		Tint:   "bg-red-400/15 text-red-700 dark:bg-red-400/25 dark:text-red-300",
		Solid:  "text-white dark:text-white bg-red-500 dark:bg-red-600",
		Bg:     "bg-red-500",
		Stroke: "stroke-red-500",
		Fill:   "fill-red-500",
		Wash:   "fill-red-500/15",
	},
	Orange: {
		Tint:   "bg-orange-400/15 text-orange-700 dark:bg-orange-400/25 dark:text-orange-300",
		Solid:  "text-white dark:text-white bg-orange-500 dark:bg-orange-600",
		Bg:     "bg-orange-500",
		Stroke: "stroke-orange-500",
		Fill:   "fill-orange-500",
		Wash:   "fill-orange-500/15",
	},
	Amber: {
		Tint:   "bg-amber-400/20 text-amber-800 dark:bg-amber-400/25 dark:text-amber-300",
		Solid:  "text-white dark:text-zinc-950 bg-amber-500 dark:bg-amber-500",
		Bg:     "bg-amber-500",
		Stroke: "stroke-amber-500",
		Fill:   "fill-amber-500",
		Wash:   "fill-amber-500/15",
	},
	Yellow: {
		Tint:   "bg-yellow-400/20 text-yellow-800 dark:bg-yellow-400/25 dark:text-yellow-200",
		Solid:  "text-white dark:text-zinc-950 bg-yellow-500 dark:bg-yellow-400",
		Bg:     "bg-yellow-500 dark:bg-yellow-400",
		Stroke: "stroke-yellow-500 dark:stroke-yellow-400",
		Fill:   "fill-yellow-500 dark:fill-yellow-400",
		Wash:   "fill-yellow-500/15 dark:fill-yellow-400/15",
	},
	Lime: {
		Tint:   "bg-lime-400/20 text-lime-800 dark:bg-lime-400/25 dark:text-lime-300",
		Solid:  "text-white dark:text-white bg-lime-500 dark:bg-lime-600",
		Bg:     "bg-lime-500",
		Stroke: "stroke-lime-500",
		Fill:   "fill-lime-500",
		Wash:   "fill-lime-500/15",
	},
	Green: {
		Tint:   "bg-green-400/15 text-green-800 dark:bg-green-400/25 dark:text-green-300",
		Solid:  "text-white dark:text-white bg-green-500 dark:bg-green-600",
		Bg:     "bg-green-500",
		Stroke: "stroke-green-500",
		Fill:   "fill-green-500",
		Wash:   "fill-green-500/15",
	},
	Emerald: {
		Tint:   "bg-emerald-400/15 text-emerald-800 dark:bg-emerald-400/25 dark:text-emerald-300",
		Solid:  "text-white dark:text-white bg-emerald-500 dark:bg-emerald-600",
		Bg:     "bg-emerald-500",
		Stroke: "stroke-emerald-500",
		Fill:   "fill-emerald-500",
		Wash:   "fill-emerald-500/15",
	},
	Teal: {
		Tint:   "bg-teal-400/15 text-teal-800 dark:bg-teal-400/25 dark:text-teal-300",
		Solid:  "text-white dark:text-white bg-teal-500 dark:bg-teal-600",
		Bg:     "bg-teal-500",
		Stroke: "stroke-teal-500",
		Fill:   "fill-teal-500",
		Wash:   "fill-teal-500/15",
	},
	Cyan: {
		Tint:   "bg-cyan-400/15 text-cyan-800 dark:bg-cyan-400/25 dark:text-cyan-300",
		Solid:  "text-white dark:text-white bg-cyan-500 dark:bg-cyan-600",
		Bg:     "bg-cyan-500",
		Stroke: "stroke-cyan-500",
		Fill:   "fill-cyan-500",
		Wash:   "fill-cyan-500/15",
	},
	Sky: {
		Tint:   "bg-sky-400/15 text-sky-800 dark:bg-sky-400/25 dark:text-sky-300",
		Solid:  "text-white dark:text-white bg-sky-500 dark:bg-sky-600",
		Bg:     "bg-sky-500",
		Stroke: "stroke-sky-500",
		Fill:   "fill-sky-500",
		Wash:   "fill-sky-500/15",
	},
	Blue: {
		Tint:   "bg-blue-400/15 text-blue-800 dark:bg-blue-400/25 dark:text-blue-300",
		Solid:  "text-white dark:text-white bg-blue-500 dark:bg-blue-600",
		Bg:     "bg-blue-500",
		Stroke: "stroke-blue-500",
		Fill:   "fill-blue-500",
		Wash:   "fill-blue-500/15",
	},
	Indigo: {
		Tint:   "bg-indigo-400/15 text-indigo-700 dark:bg-indigo-400/25 dark:text-indigo-300",
		Solid:  "text-white dark:text-white bg-indigo-500 dark:bg-indigo-600",
		Bg:     "bg-indigo-500",
		Stroke: "stroke-indigo-500",
		Fill:   "fill-indigo-500",
		Wash:   "fill-indigo-500/15",
	},
	Violet: {
		Tint:   "bg-violet-400/15 text-violet-700 dark:bg-violet-400/25 dark:text-violet-300",
		Solid:  "text-white dark:text-white bg-violet-500 dark:bg-violet-600",
		Bg:     "bg-violet-500",
		Stroke: "stroke-violet-500",
		Fill:   "fill-violet-500",
		Wash:   "fill-violet-500/15",
	},
	Purple: {
		Tint:   "bg-purple-400/15 text-purple-700 dark:bg-purple-400/25 dark:text-purple-300",
		Solid:  "text-white dark:text-white bg-purple-500 dark:bg-purple-600",
		Bg:     "bg-purple-500",
		Stroke: "stroke-purple-500",
		Fill:   "fill-purple-500",
		Wash:   "fill-purple-500/15",
	},
	Fuchsia: {
		Tint:   "bg-fuchsia-400/15 text-fuchsia-700 dark:bg-fuchsia-400/25 dark:text-fuchsia-300",
		Solid:  "text-white dark:text-white bg-fuchsia-500 dark:bg-fuchsia-600",
		Bg:     "bg-fuchsia-500",
		Stroke: "stroke-fuchsia-500",
		Fill:   "fill-fuchsia-500",
		Wash:   "fill-fuchsia-500/15",
	},
	Pink: {
		Tint:   "bg-pink-400/15 text-pink-700 dark:bg-pink-400/25 dark:text-pink-300",
		Solid:  "text-white dark:text-white bg-pink-500 dark:bg-pink-600",
		Bg:     "bg-pink-500",
		Stroke: "stroke-pink-500",
		Fill:   "fill-pink-500",
		Wash:   "fill-pink-500/15",
	},
	Rose: {
		Tint:   "bg-rose-400/15 text-rose-700 dark:bg-rose-400/25 dark:text-rose-300",
		Solid:  "text-white dark:text-white bg-rose-500 dark:bg-rose-600",
		Bg:     "bg-rose-500",
		Stroke: "stroke-rose-500",
		Fill:   "fill-rose-500",
		Wash:   "fill-rose-500/15",
	},
}
