package flash

import "github.com/pietjan/loom/internal/styles"

var toneClasses = map[Tone]string{
	ToneNeutral: "border-base-200 bg-white text-base-800 dark:border-base-600 dark:bg-base-700 dark:text-base-100",
	ToneInfo:    "border-blue-200 bg-blue-50 text-blue-900 dark:border-blue-400/30 dark:bg-blue-400/10 dark:text-blue-200",
	ToneSuccess: "border-green-200 bg-green-50 text-green-900 dark:border-green-400/30 dark:bg-green-400/10 dark:text-green-200",
	ToneWarning: "border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-400/30 dark:bg-amber-400/10 dark:text-amber-200",
	ToneDanger:  "border-red-200 bg-red-50 text-red-900 dark:border-red-400/30 dark:bg-red-400/10 dark:text-red-200",
}

func classes(c Config) string {
	var b styles.Builder
	b.Add("flex items-start gap-3 rounded-lg border p-4 shadow-xs")
	// The hidden-checkbox dismiss: when it's checked, hide the flash.
	b.Add("has-checked:hidden")
	styles.Match(&b, c.Tone, toneClasses)
	return b.String()
}

func closeClasses() string {
	var b styles.Builder
	b.Add("-m-1 inline-flex shrink-0 cursor-pointer items-center justify-center rounded p-1 opacity-70")
	b.Add("hover:opacity-100 has-focus-visible:outline has-focus-visible:outline-2 has-focus-visible:outline-current")
	b.Add("**:data-[ui=icon]:size-4")
	return b.String()
}
