package pagination

import "github.com/pietjan/loom/internal/styles"

func navClasses() string {
	var b styles.Builder
	b.Add("flex items-center gap-1")
	return b.String()
}

func itemClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center justify-center min-w-9 h-9 px-2 rounded-md text-sm font-medium")
	b.Add("text-base-500 dark:text-base-400")
	b.Add("hover:bg-base-800/5 hover:text-base-800 dark:hover:bg-white/10 dark:hover:text-white")
	b.Add("aria-[current=page]:bg-accent aria-[current=page]:text-accent-foreground aria-[current=page]:hover:bg-accent")
	b.Add("aria-disabled:opacity-40 aria-disabled:pointer-events-none")
	b.Add("**:data-[ui=icon]:size-4")
	return b.String()
}

func gapClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center justify-center min-w-9 h-9 text-sm text-base-400")
	return b.String()
}
