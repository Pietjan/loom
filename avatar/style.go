package avatar

import "github.com/pietjan/loom/internal/styles"

func classes(c Config) string {
	var b styles.Builder
	b.Add("relative inline-flex shrink-0 items-center justify-center overflow-visible rounded-full")
	b.Add("bg-base-200 text-base-700 font-medium uppercase")
	b.Add("dark:bg-base-600 dark:text-base-100")
	styles.Match(&b, c.Size, map[Size]string{
		SizeSmall: "size-6 text-xs",
		SizeBase:  "size-8 text-sm",
		SizeLarge: "size-12 text-base",
	})
	return b.String()
}

func imageClasses() string {
	var b styles.Builder
	b.Add("size-full rounded-full object-cover")
	return b.String()
}

var statusColor = map[Status]string{
	StatusOnline:  "bg-green-500",
	StatusOffline: "bg-base-400",
	StatusBusy:    "bg-red-500",
	StatusAway:    "bg-amber-500",
}

func statusClasses(s Status) string {
	var b styles.Builder
	b.Add("absolute bottom-0 end-0 size-2.5 rounded-full ring-2 ring-white dark:ring-base-800")
	styles.Match(&b, s, statusColor)
	return b.String()
}

func groupClasses() string {
	var b styles.Builder
	// Overlap avatars and give each a ring so the stack reads clearly.
	b.Add("flex items-center -space-x-2")
	b.Add("[&>[data-ui=avatar]]:ring-2 [&>[data-ui=avatar]]:ring-white dark:[&>[data-ui=avatar]]:ring-base-800")
	return b.String()
}
