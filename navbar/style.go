package navbar

import "github.com/pietjan/loom/internal/styles"

func navClasses() string {
	var b styles.Builder
	b.Add("flex items-center gap-1")
	return b.String()
}

func itemClasses() string {
	var b styles.Builder
	b.Add("relative inline-flex items-center gap-2 rounded-md px-3 h-14 text-sm font-medium")
	b.Add("text-base-500 dark:text-base-300")
	b.Add("hover:text-base-800 dark:hover:text-white")
	// Current page: accent text + an underline flush with the header's
	// bottom border (the persistent-tab look).
	b.Add("aria-[current=page]:text-accent-content")
	b.Add("aria-[current=page]:after:absolute aria-[current=page]:after:inset-x-3 aria-[current=page]:after:-bottom-px aria-[current=page]:after:h-0.5 aria-[current=page]:after:bg-accent")
	b.Add("**:data-[ui=icon]:size-5 **:data-[ui=icon]:text-base-400")
	b.Add("aria-[current=page]:**:data-[ui=icon]:text-accent-content")
	return b.String()
}

func badgeClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center justify-center min-w-5 h-5 px-1.5 rounded-full")
	b.Add("text-xs font-medium bg-base-800/5 text-base-500")
	b.Add("dark:bg-white/10 dark:text-base-300")
	return b.String()
}
