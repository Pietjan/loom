package stat

import "github.com/pietjan/loom/internal/styles"

func classes() string {
	var b styles.Builder
	b.Add("flex flex-col gap-1 rounded-lg border border-base-200 bg-white p-5 shadow-xs")
	b.Add("dark:border-base-600 dark:bg-base-700")
	return b.String()
}

func labelClasses() string {
	var b styles.Builder
	b.Add("text-sm font-medium text-base-500 dark:text-base-400")
	return b.String()
}

func valueClasses() string {
	var b styles.Builder
	b.Add("text-2xl font-semibold tracking-tight text-base-800 dark:text-white")
	return b.String()
}
