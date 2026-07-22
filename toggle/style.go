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
	b.Add("pointer-events-none absolute top-0.5 start-0.5 size-4 rounded-full shadow-sm transition-transform")
	// White on the unchecked track; on the checked (accent) track the thumb
	// takes --color-accent-foreground, same reasoning as the checkbox glyph:
	// neutral accents invert in dark mode (white surface) and amber/yellow/
	// lime are light in both — a white thumb disappears on those.
	b.Add("bg-white peer-checked:bg-accent-foreground")
	b.Add("peer-checked:translate-x-4 rtl:peer-checked:-translate-x-4")
	return b.String()
}

func wrapperClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-2 text-sm text-base-800 select-none cursor-pointer dark:text-base-100")
	b.Add("has-[input:disabled]:opacity-75 has-[input:disabled]:cursor-not-allowed")
	return b.String()
}
