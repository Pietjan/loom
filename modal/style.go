package modal

import "github.com/pietjan/loom/internal/styles"

// classes style the dialog box; the backdrop and open/close transitions
// are structural CSS in css/loom.css ([data-ui=modal]).
func classes(c Config) string {
	var b styles.Builder
	b.Add("m-auto w-full max-w-lg rounded-xl p-6")
	b.Add("bg-white text-base-800 shadow-xl border border-base-200")
	b.Add("dark:bg-base-800 dark:text-base-100 dark:border-base-600")
	b.Add("space-y-4")
	return b.String()
}

func titleClasses() string {
	var b styles.Builder
	b.Add("text-lg font-semibold text-base-800 dark:text-white")
	return b.String()
}
