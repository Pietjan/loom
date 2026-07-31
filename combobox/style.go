package combobox

import "github.com/pietjan/loom/internal/styles"

// rootClasses make the wrapper a positioning context, which is what the
// list falls back to where CSS anchor positioning is missing. inline-block
// rather than inline-flex: the wrapper holds a field, and a field should be
// able to fill its column.
func rootClasses() string {
	var b styles.Builder
	b.Add("relative block w-full")
	return b.String()
}

// inputClasses deliberately mirror the input component's recipe rather than
// composing it: a combobox's query field is the same control, and reaching
// into another component's package for its classes is the coupling
// arch_test.go exists to stop.
func inputClasses() string {
	var b styles.Builder
	b.Add("block h-10 w-full rounded-lg border px-3 text-sm shadow-xs")
	b.Add("border-base-200 border-b-base-300/80 bg-white text-base-800 placeholder:text-base-400")
	b.Add("focus:border-accent focus:outline-none focus:ring-2 focus:ring-accent/30")
	b.Add("dark:border-base-600 dark:bg-base-700 dark:text-base-100")
	return b.String()
}

// listClasses set no display utility, and that is load-bearing: the panel's
// visibility is the popover's, and a utility would sit in the utilities
// layer and beat the [popover] rules in css/loom.css - the one thing a
// visibility toggle cannot survive. Width tracks the field through
// anchor-size where it is supported; the fallback is the wrapper's width.
func listClasses() string {
	var b styles.Builder
	b.Add("max-h-80 overflow-y-auto rounded-lg border p-1 shadow-lg")
	b.Add("border-base-200 bg-white")
	b.Add("dark:border-base-600 dark:bg-base-700")
	return b.String()
}

// itemClasses need no selected state for the cursor: arrow keys move real
// focus between rows, so focus-visible is already it. data-chosen is the
// value that was picked, which is a different thing and outlives the panel.
func itemClasses() string {
	var b styles.Builder
	b.Add("flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-start text-sm cursor-pointer")
	b.Add("text-base-800 dark:text-base-100")
	b.Add("hover:bg-base-800/5 dark:hover:bg-white/10")
	b.Add("focus-visible:bg-base-800/5 dark:focus-visible:bg-white/10")
	b.Add("focus-visible:outline-none")
	b.Add("data-chosen:font-medium data-chosen:text-accent-content")
	b.Add("disabled:cursor-not-allowed disabled:opacity-50")
	b.Add("**:data-[ui=icon]:size-4 **:data-[ui=icon]:text-base-400")
	return b.String()
}

func emptyClasses() string {
	var b styles.Builder
	b.Add("px-2 py-6 text-center text-sm text-base-500 dark:text-base-400")
	return b.String()
}
