package field

import "github.com/pietjan/loom/internal/styles"

func classes(c Config) string {
	var b styles.Builder
	b.Add("grid gap-2")
	return b.String()
}

func labelClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-1 text-sm font-medium text-base-800 select-none dark:text-white")
	return b.String()
}

func descriptionClasses() string {
	var b styles.Builder
	b.Add("text-sm text-base-500 dark:text-base-400")
	return b.String()
}

func errorClasses() string {
	var b styles.Builder
	b.Add("text-sm font-medium text-red-600 dark:text-red-400")
	return b.String()
}
