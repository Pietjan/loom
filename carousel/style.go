package carousel

import "github.com/pietjan/loom/internal/styles"

func rootClasses() string {
	var b styles.Builder
	b.Add("w-full")
	return b.String()
}

func trackClasses() string {
	var b styles.Builder
	// Horizontal scroll-snap track. scroll-smooth makes the dot links
	// glide; snap-mandatory locks each slide into place.
	b.Add("flex snap-x snap-mandatory overflow-x-auto scroll-smooth rounded-lg")
	b.Add("[scrollbar-width:none] [&::-webkit-scrollbar]:hidden")
	return b.String()
}

func slideClasses() string {
	var b styles.Builder
	b.Add("w-full shrink-0 snap-start")
	return b.String()
}

func dotsClasses() string {
	var b styles.Builder
	b.Add("mt-3 flex items-center justify-center gap-2")
	return b.String()
}

func dotClasses() string {
	var b styles.Builder
	b.Add("size-2 rounded-full bg-base-300 transition-colors hover:bg-base-400")
	b.Add("dark:bg-base-600 dark:hover:bg-base-500")
	return b.String()
}
