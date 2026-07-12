package tooltip

import "github.com/pietjan/loom/internal/styles"

func classes(c Config) string {
	var b styles.Builder
	b.Add("group/tooltip relative inline-flex")
	return b.String()
}

func tipClasses() string {
	var b styles.Builder
	b.Add("pointer-events-none absolute bottom-full left-1/2 z-10 mb-1.5 -translate-x-1/2")
	b.Add("whitespace-nowrap rounded-md px-2 py-1 text-xs font-medium")
	// b.Add("bg-base-800 text-white dark:bg-base-100 dark:text-base-800")
	b.Add("text-white bg-base-800 dark:bg-base-700 dark:border dark:border-white/10")
	b.Add("invisible opacity-0 transition-opacity duration-100")
	b.Add("group-hover/tooltip:visible group-hover/tooltip:opacity-100")
	b.Add("group-focus-within/tooltip:visible group-focus-within/tooltip:opacity-100")
	return b.String()
}
