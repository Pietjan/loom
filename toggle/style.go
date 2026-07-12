package toggle

import "github.com/pietjan/loom/internal/styles"

func trackClasses() string {
	var b styles.Builder
	b.Add("peer appearance-none w-9 h-5 rounded-full cursor-pointer transition-colors")
	b.Add("bg-base-300 dark:bg-base-600 checked:bg-accent dark:checked:bg-accent")
	b.Add("disabled:opacity-75 disabled:cursor-not-allowed")
	return b.String()
}

func thumbClasses() string {
	var b styles.Builder
	b.Add("pointer-events-none absolute top-0.5 start-0.5 size-4 rounded-full bg-white shadow-sm transition-transform")
	b.Add("peer-checked:translate-x-4 rtl:peer-checked:-translate-x-4")
	return b.String()
}

func wrapperClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-2 text-sm text-base-800 select-none cursor-pointer dark:text-base-100")
	b.Add("has-[input:disabled]:opacity-75 has-[input:disabled]:cursor-not-allowed")
	return b.String()
}
