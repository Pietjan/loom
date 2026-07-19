package picker

import "github.com/pietjan/loom/internal/styles"

// classes style the <select> itself — its closed appearance in both
// worlds (base-select and classic). Flex layout keeps the browser's
// ::picker-icon on the same row as the value (block would wrap it);
// classic selects ignore inner layout, so it is harmless there.
func classes(c Config) string {
	var b styles.Builder
	b.Add("flex items-center justify-between gap-2 w-full h-10 px-3 text-sm rounded-lg cursor-pointer")
	b.Add("bg-white text-base-800")
	b.Add("border shadow-xs")
	b.Add("dark:bg-base-700 dark:text-base-100")
	b.Add("disabled:opacity-75 disabled:cursor-not-allowed")
	b.If(!c.invalid, "border-base-200 border-b-base-300/80 dark:border-base-600")
	b.If(c.invalid, "border-red-500 dark:border-red-400")
	return b.String()
}

// buttonClasses style the customizable-select closed button (Chromium
// only; inert elsewhere). It takes the free space; the ::picker-icon sits
// after it.
func buttonClasses() string {
	var b styles.Builder
	b.Add("flex flex-1 items-center gap-2 min-w-0")
	return b.String()
}

// optionClasses style each option inside the base-select picker panel.
// Non-supporting browsers ignore option classes — harmless.
func optionClasses() string {
	var b styles.Builder
	b.Add("flex items-center gap-2 rounded-md px-2 py-1.5 text-sm cursor-pointer")
	b.Add("text-base-800 dark:text-base-100")
	b.Add("hover:bg-base-800/5 dark:hover:bg-white/10")
	b.Add("checked:font-medium")
	b.Add("disabled:opacity-50 disabled:cursor-not-allowed")
	b.Add("**:data-[ui=icon]:size-4")
	return b.String()
}

// arrowClasses style the icon replacing ::picker-icon; :open on the
// ancestor select rotates it. Scoped to select:open — a bare in-open:
// would also match an open <details> ancestor (accordion, navlist group).
func arrowClasses() string {
	var b styles.Builder
	b.Add("ms-auto shrink-0 text-base-400 transition-transform duration-150 in-[select:open]:rotate-180")
	return b.String()
}

// checkClasses style the icon replacing option::checkmark: pushed to
// the row end, visible only while the option is checked.
func checkClasses() string {
	var b styles.Builder
	b.Add("ms-auto shrink-0 invisible text-accent-content in-checked:visible")
	return b.String()
}
