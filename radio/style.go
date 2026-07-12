package radio

import "github.com/pietjan/loom/internal/styles"

func classes(c Config) string {
	var b styles.Builder
	b.Add("appearance-none shrink-0 size-4.5 rounded-full border shadow-xs cursor-pointer")
	b.Add("bg-white dark:bg-base-700")
	// The dot itself is structural CSS in css/loom.css ([data-ui=radio]:checked).
	b.Add("checked:bg-accent checked:border-transparent")
	b.Add("disabled:opacity-75 disabled:cursor-not-allowed")
	b.If(!c.invalid, "border-base-300 dark:border-base-600")
	b.If(c.invalid, "border-red-500 dark:border-red-400")
	return b.String()
}

func wrapperClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-2 text-sm text-base-800 select-none cursor-pointer dark:text-base-100")
	b.Add("has-[input:disabled]:opacity-75 has-[input:disabled]:cursor-not-allowed")
	return b.String()
}

func groupClasses() string {
	var b styles.Builder
	b.Add("grid gap-2 border-0 p-0 m-0")
	return b.String()
}

func legendClasses() string {
	var b styles.Builder
	b.Add("mb-1 p-0 text-sm font-medium text-base-800 dark:text-white")
	return b.String()
}
