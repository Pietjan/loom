package kanban

import "github.com/pietjan/loom/internal/styles"

func boardClasses() string {
	var b styles.Builder
	b.Add("flex items-start gap-4 overflow-x-auto")
	return b.String()
}

func columnClasses() string {
	var b styles.Builder
	// Translucent panel so the column reads on any page background.
	b.Add("flex max-h-full w-72 shrink-0 flex-col rounded-xl bg-base-100 p-2 dark:bg-base-800")
	return b.String()
}

func headerClasses() string {
	var b styles.Builder
	b.Add("flex items-center gap-1 px-1.5 pt-1 pb-2.5")
	return b.String()
}

// headerTitleClasses is the heading block; flex-1 pushes actions to the end.
func headerTitleClasses() string {
	var b styles.Builder
	b.Add("min-w-0 flex-1")
	return b.String()
}

func headerRowClasses() string {
	var b styles.Builder
	b.Add("flex items-center gap-2")
	return b.String()
}

func headingClasses() string {
	var b styles.Builder
	b.Add("truncate text-sm font-medium text-base-800 dark:text-white")
	return b.String()
}

// cardHeadingClasses wraps instead of truncating — card titles run long.
func cardHeadingClasses() string {
	var b styles.Builder
	b.Add("text-sm font-medium text-base-800 dark:text-white")
	return b.String()
}

func countClasses() string {
	var b styles.Builder
	b.Add("rounded-full bg-base-800/5 px-1.5 py-0.5 text-xs font-medium text-base-500 dark:bg-white/10 dark:text-base-300")
	return b.String()
}

func subheadingClasses() string {
	var b styles.Builder
	b.Add("truncate text-xs text-base-400")
	return b.String()
}

func cardsClasses() string {
	var b styles.Builder
	b.Add("flex min-h-0 flex-col gap-2 overflow-y-auto")
	return b.String()
}

func footerClasses() string {
	var b styles.Builder
	b.Add("flex flex-col gap-2 pt-2")
	return b.String()
}

func cardClasses() string {
	var b styles.Builder
	// A real border, not a ring: rings draw outside the box and get
	// clipped by the Cards container's overflow.
	b.Add("rounded-lg border border-base-200 bg-white p-3 shadow-xs")
	b.Add("dark:border-base-600 dark:bg-base-700")
	return b.String()
}

func cardHeaderClasses() string {
	var b styles.Builder
	b.Add("mb-2 flex flex-wrap items-center gap-2")
	return b.String()
}

func cardFooterClasses() string {
	var b styles.Builder
	b.Add("mt-2 flex items-center gap-2 text-base-400")
	return b.String()
}
