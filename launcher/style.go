package launcher

import "github.com/pietjan/loom/internal/styles"

func rootClasses() string {
	var b styles.Builder
	b.Add("flex w-full flex-col gap-2")
	return b.String()
}

// inputClasses deliberately mirror the input component's recipe rather
// than composing it: a launcher's query field is the same control with the
// native search decorations suppressed, and reaching into another
// component's package for its classes is the coupling arch_test.go exists
// to stop.
func inputClasses() string {
	var b styles.Builder
	b.Add("block w-full rounded-lg border border-base-200 bg-white px-3 py-2 text-sm")
	b.Add("text-base-800 placeholder:text-base-400")
	b.Add("focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30")
	b.Add("dark:border-base-600 dark:bg-base-700 dark:text-base-100")
	b.Add("[&::-webkit-search-cancel-button]:appearance-none")
	return b.String()
}

// listClasses set no display utility on purpose. Whether the panel is
// showing is derived in css/loom.css from focus and from whether any row
// survived the filter, and a utility would sit in the utilities layer and
// beat those rules - the one thing a visibility toggle cannot survive.
func listClasses() string {
	var b styles.Builder
	b.Add("max-h-80 overflow-y-auto rounded-lg border border-base-200 p-1")
	b.Add("dark:border-base-600")
	return b.String()
}

// itemClasses need no selected state: arrow keys move real focus between
// rows, so focus-visible is already the cursor.
func itemClasses() string {
	var b styles.Builder
	b.Add("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-start text-sm cursor-pointer")
	b.Add("text-base-800 dark:text-base-100")
	b.Add("hover:bg-base-800/5 dark:hover:bg-white/10")
	b.Add("focus-visible:bg-base-800/5 dark:focus-visible:bg-white/10")
	b.Add("focus-visible:outline-none")
	b.Add("**:data-[ui=icon]:size-4 **:data-[ui=icon]:text-base-400")
	return b.String()
}

func emptyClasses() string {
	var b styles.Builder
	b.Add("px-2 py-8 text-center text-sm text-base-500 dark:text-base-400")
	return b.String()
}
