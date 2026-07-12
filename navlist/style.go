package navlist

import "github.com/pietjan/loom/internal/styles"

func navClasses() string {
	var b styles.Builder
	b.Add("flex flex-col gap-0.5")
	return b.String()
}

func itemClasses() string {
	var b styles.Builder
	b.Add("flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium")
	b.Add("text-base-500 dark:text-base-300")
	b.Add("hover:bg-base-800/5 hover:text-base-800 dark:hover:bg-white/10 dark:hover:text-white")
	b.Add("aria-[current=page]:bg-base-800/5 aria-[current=page]:text-base-800")
	b.Add("dark:aria-[current=page]:bg-white/10 dark:aria-[current=page]:text-white")
	b.Add("**:data-[ui=icon]:size-5 **:data-[ui=icon]:text-base-400")
	b.Add("[&[aria-current=page]_[data-ui=icon]]:text-accent-content")
	return b.String()
}

func groupClasses() string {
	var b styles.Builder
	b.Add("group/navgroup")
	return b.String()
}

func summaryClasses() string {
	var b styles.Builder
	b.Add("flex cursor-pointer items-center justify-between gap-2 rounded-md px-2 py-1.5")
	b.Add("text-sm font-medium text-base-500 select-none dark:text-base-300")
	b.Add("hover:bg-base-800/5 hover:text-base-800 dark:hover:bg-white/10 dark:hover:text-white")
	b.Add("list-none [&::-webkit-details-marker]:hidden")
	return b.String()
}

func chevronClasses() string {
	var b styles.Builder
	b.Add("size-4 shrink-0 text-base-400 transition-transform duration-150")
	b.Add("group-open/navgroup:rotate-180")
	return b.String()
}

func panelClasses() string {
	var b styles.Builder
	b.Add("ms-3 mt-0.5 flex flex-col gap-0.5 border-s border-base-200 ps-3 dark:border-base-600")
	return b.String()
}

func headingClasses() string {
	var b styles.Builder
	b.Add("px-2 pt-4 pb-1 text-xs font-semibold uppercase tracking-wide text-base-400")
	return b.String()
}
