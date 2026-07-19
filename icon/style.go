package icon

import "github.com/pietjan/loom/internal/styles"

// The size utilities are wrapped in :where() so any composite's icon
// sizing (e.g. a badge shrinking its icon) wins without specificity games.
func classes(c Config) string {
	var b styles.Builder
	b.Add("shrink-0")
	styles.Match(&b, c.Size, map[Size]string{
		SizeBase:       "[:where(&)]:size-6",
		SizeSmall:      "[:where(&)]:size-5",
		SizeExtraSmall: "[:where(&)]:size-4",
	})
	return b.String()
}
