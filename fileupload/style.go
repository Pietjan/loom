package fileupload

import "github.com/pietjan/loom/internal/styles"

// classes style the input shell; the "choose file" button itself is
// styled via ::file-selector-button in cmd/css/loom.css (Tailwind can't
// target that pseudo-element).
func classes() string {
	var b styles.Builder
	b.Add("block w-full text-sm rounded-lg cursor-pointer")
	b.Add("text-base-500 dark:text-base-400")
	b.Add("border border-base-200 border-b-base-300/80 shadow-xs")
	b.Add("bg-white dark:bg-base-700 dark:border-base-600")
	b.Add("disabled:opacity-75 disabled:cursor-not-allowed")
	return b.String()
}
