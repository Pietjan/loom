package progress

import "github.com/pietjan/loom/internal/styles"

func trackClasses() string {
	var b styles.Builder
	b.Add("relative h-2 w-full overflow-hidden rounded-full bg-base-200 dark:bg-base-700")
	return b.String()
}

var barColor = map[Color]string{
	Accent:  "bg-accent",
	Emerald: "bg-emerald-500",
	Amber:   "bg-amber-500",
	Red:     "bg-red-500",
}

func barClasses(c Config) string {
	var b styles.Builder
	b.Add("h-full rounded-full transition-[width] duration-300")
	styles.Match(&b, c.Color, barColor)
	// For indeterminate, width and the slide are structural CSS (keyed on
	// the track's data-indeterminate) in cmd/css/loom.css.
	b.If(!c.indeterminate, "w-0")
	return b.String()
}
