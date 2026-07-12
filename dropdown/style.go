package dropdown

import "github.com/pietjan/loom/internal/styles"

func rootClasses() string {
	var b styles.Builder
	// relative: containing block for the non-anchor-positioning fallback.
	b.Add("relative inline-flex")
	return b.String()
}

// menuClasses style the panel surface; its positioning (anchor vs
// fallback) is structural CSS in css/loom.css ([data-ui=dropdown-menu]).
func menuClasses() string {
	var b styles.Builder
	b.Add("min-w-48 rounded-lg border border-base-200 bg-white p-1 shadow-lg")
	b.Add("dark:border-base-600 dark:bg-base-700")
	return b.String()
}

func itemClasses() string {
	var b styles.Builder
	b.Add("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-start text-sm cursor-pointer")
	b.Add("text-base-800 dark:text-base-100")
	b.Add("hover:bg-base-800/5 dark:hover:bg-white/10")
	b.Add("focus-visible:bg-base-800/5 dark:focus-visible:bg-white/10")
	b.Add("[&_[data-ui=icon]]:size-4 [&_[data-ui=icon]]:text-base-400")
	return b.String()
}

func dividerClasses() string {
	var b styles.Builder
	b.Add("my-1 border-0 h-px bg-base-200 dark:bg-base-600")
	return b.String()
}
