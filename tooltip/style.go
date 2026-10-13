package tooltip

import "github.com/pietjan/loom/internal/styles"

// classes is the wrapper's. The tip's z-10 only counts inside the nearest
// stacking context, and a wrapper the caller positions or transforms - a
// pin centred on a map point - is one, so siblings drawn after it paint
// over the tip. Raising the wrapper while it is hovered or focused lifts
// the tip with it, among whatever it is placed in.
func classes(_ Config) string {
	var b styles.Builder
	b.Add("group/tooltip relative inline-flex")
	b.Add("hover:z-30 focus-within:z-30")
	return b.String()
}

func tipClasses(cfg Config) string {
	var b styles.Builder
	if cfg.Side == "right" {
		b.Add("pointer-events-none absolute left-full top-1/2 z-10 ml-1.5 -translate-y-1/2")
	} else {
		b.Add("pointer-events-none absolute bottom-full left-1/2 z-10 mb-1.5 -translate-x-1/2")
	}
	b.Add("whitespace-nowrap rounded-md px-2 py-1 text-xs font-medium")
	// b.Add("bg-base-800 text-white dark:bg-base-100 dark:text-base-800")
	b.Add("text-white bg-base-800 dark:bg-base-700 dark:border dark:border-white/10")
	b.Add("invisible opacity-0 transition-opacity duration-100")
	b.Add("group-hover/tooltip:visible group-hover/tooltip:opacity-100")
	b.Add("group-focus-within/tooltip:visible group-focus-within/tooltip:opacity-100")
	return b.String()
}
