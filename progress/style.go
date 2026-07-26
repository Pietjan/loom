package progress

import (
	"github.com/pietjan/loom/internal/palette"
	"github.com/pietjan/loom/internal/styles"
)

func trackClasses() string {
	var b styles.Builder
	b.Add("relative h-2 w-full overflow-hidden rounded-full bg-base-200 dark:bg-base-700")
	return b.String()
}

func barClasses(c Config) string {
	var b styles.Builder
	b.Add("h-full rounded-full transition-[width] duration-300")
	// The bar is the color rather than something sitting on it, so it
	// takes the palette's mark strength; Accent is the theme token and has
	// no hue row.
	if s, ok := palette.Of(palette.Color(c.Color)); ok {
		b.Add(s.Bg)
	} else {
		b.Add("bg-accent")
	}
	// For indeterminate, width and the slide are structural CSS (keyed on
	// the track's data-indeterminate) in cmd/css/loom.css.
	b.If(!c.indeterminate, "w-0")
	return b.String()
}
