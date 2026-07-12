package table

import "github.com/pietjan/loom/internal/styles"

func wrapperClasses() string {
	var b styles.Builder
	b.Add("w-full overflow-x-auto rounded-lg border border-base-200 dark:border-base-600")
	return b.String()
}

func tableClasses() string {
	var b styles.Builder
	b.Add("w-full text-sm")
	return b.String()
}

func headerClasses() string {
	var b styles.Builder
	b.Add("bg-base-50 dark:bg-base-800")
	return b.String()
}

func bodyClasses() string {
	var b styles.Builder
	b.Add("divide-y divide-base-200 dark:divide-base-600")
	return b.String()
}

func rowClasses() string {
	var b styles.Builder
	b.Add("hover:bg-base-800/2.5 dark:hover:bg-white/2.5")
	return b.String()
}

func columnClasses() string {
	var b styles.Builder
	b.Add("px-4 py-2.5 text-start text-xs font-semibold uppercase tracking-wide")
	b.Add("text-base-500 dark:text-base-400")
	b.Add("border-b border-base-200 dark:border-base-600")
	return b.String()
}

func cellClasses() string {
	var b styles.Builder
	b.Add("px-4 py-3 text-base-800 dark:text-base-100")
	return b.String()
}
