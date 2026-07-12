package breadcrumbs

import "github.com/pietjan/loom/internal/styles"

func listClasses() string {
	var b styles.Builder
	b.Add("flex flex-wrap items-center gap-1.5 text-sm")
	return b.String()
}

func itemClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-1.5")
	// A chevron separator before every item except the first.
	b.Add("[&:not(:first-child)]:before:content-['/'] [&:not(:first-child)]:before:text-base-300 dark:[&:not(:first-child)]:before:text-base-600")
	return b.String()
}

func linkClasses() string {
	var b styles.Builder
	b.Add("text-base-500 hover:text-base-800 dark:text-base-400 dark:hover:text-white")
	return b.String()
}

func currentClasses() string {
	var b styles.Builder
	b.Add("font-medium text-base-800 dark:text-white")
	return b.String()
}
