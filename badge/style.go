package badge

import "github.com/pietjan/loom/internal/styles"

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

func classes(c Config) string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-1 font-medium whitespace-nowrap")
	// Icons shrink to fit the badge — CSS, not tree surgery.
	b.Add("**:data-[ui=icon]:size-4")
	styles.Match(&b, c.Size, map[Size]string{
		SizeSmall: "text-xs px-1.5 py-0.5 **:data-[ui=icon]:size-3",
		SizeBase:  "text-sm px-2 py-1 **:data-[ui=icon]:size-4",
		SizeLarge: "text-sm px-2.5 py-1.5 **:data-[ui=icon]:size-5",
	})
	b.If(c.pill, "rounded-full")
	b.If(!c.pill, "rounded-md")
	styles.Match(&b, c.Color, colorClasses)
	return b.String()
}
