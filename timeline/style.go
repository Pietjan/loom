package timeline

import "github.com/pietjan/loom/internal/styles"

func listClasses() string {
	var b styles.Builder
	b.Add("flex flex-col")
	return b.String()
}

func itemClasses() string {
	var b styles.Builder
	// The inline-start border is the connector; the last item's border
	// stops at the dot so the line doesn't dangle.
	b.Add("relative border-s border-base-200 ps-6 pb-6 dark:border-base-600")
	b.Add("last:border-transparent last:pb-0")
	return b.String()
}

func dotClasses() string {
	var b styles.Builder
	// Centered on the connector line.
	b.Add("absolute -start-[4.5px] top-1 size-2.5 rounded-full bg-accent ring-4 ring-white dark:ring-base-800")
	return b.String()
}

func titleClasses() string {
	var b styles.Builder
	b.Add("flex items-baseline gap-2 font-medium text-base-800 dark:text-white")
	return b.String()
}

func timeClasses() string {
	var b styles.Builder
	b.Add("text-xs font-normal text-base-400")
	return b.String()
}
