package sidebar

import "github.com/pietjan/loom/internal/styles"

// classes style the sidebar surface. Its responsive popover-vs-static
// behavior (display/position/inset overrides, slide-in transition,
// backdrop) is structural CSS in css/loom.css ([data-ui=sidebar]) — it
// must override popover user-agent styles, which utility classes must not
// try to fight.
func classes() string {
	var b styles.Builder
	b.Add("w-64 flex-col gap-4 p-4")
	b.Add("bg-base-50 text-base-800 border-e border-base-200")
	b.Add("dark:bg-base-800 dark:text-base-100 dark:border-base-600")
	return b.String()
}

func toggleClasses() string {
	var b styles.Builder
	b.Add("inline-flex items-center justify-center size-10 rounded-lg cursor-pointer")
	b.Add("text-base-800 hover:bg-base-800/5 dark:text-base-100 dark:hover:bg-white/10")
	b.Add("lg:hidden")
	return b.String()
}
