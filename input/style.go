package input

import "github.com/pietjan/loom/internal/styles"

func classes(c Config) string {
	var b styles.Builder

	// Inside an input group: no shell of its own — fill the row, stay
	// transparent, let the group draw the border/background/ring.
	if c.grouped {
		b.Add("min-w-0 flex-1 h-full px-3 text-sm bg-transparent text-base-800 placeholder:text-base-400")
		b.Add("border-0 outline-none disabled:cursor-not-allowed dark:text-base-100")
		return b.String()
	}

	b.Add("block w-full h-10 px-3 text-sm rounded-lg")
	b.Add("bg-white text-base-800 placeholder:text-base-400")
	b.Add("border shadow-xs")
	b.Add("dark:bg-base-700 dark:text-base-100")
	b.Add("disabled:opacity-75 disabled:cursor-not-allowed disabled:bg-base-50 dark:disabled:bg-base-800")
	b.If(!c.invalid, "border-base-200 border-b-base-300/80 dark:border-base-600")
	b.If(c.invalid, "border-red-500 dark:border-red-400")
	return b.String()
}
