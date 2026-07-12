package accordion

import "github.com/pietjan/loom/internal/styles"

func rootClasses() string {
	var b styles.Builder
	b.Add("divide-y divide-base-200 rounded-lg border border-base-200")
	b.Add("dark:divide-base-600 dark:border-base-600")
	return b.String()
}

func itemClasses() string {
	var b styles.Builder
	b.Add("group/accordion")
	return b.String()
}

func summaryClasses() string {
	var b styles.Builder
	b.Add("flex cursor-pointer items-center justify-between gap-2 px-4 py-3")
	b.Add("text-sm font-medium text-base-800 select-none dark:text-white")
	b.Add("list-none [&::-webkit-details-marker]:hidden")
	b.Add("hover:bg-base-800/2.5 dark:hover:bg-white/5")
	return b.String()
}

func chevronClasses() string {
	var b styles.Builder
	b.Add("size-4 shrink-0 text-base-400 transition-transform duration-150")
	b.Add("group-open/accordion:rotate-180")
	return b.String()
}

func panelClasses() string {
	var b styles.Builder
	b.Add("px-4 pb-4 pt-1 text-sm text-base-600 dark:text-base-300")
	return b.String()
}
