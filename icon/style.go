package icon

import "github.com/pietjan/loom/internal/styles"

// The size utilities are wrapped in :where() so any composite's icon
// sizing (e.g. a badge shrinking its icon) wins without specificity games.
func classes(c Config) string {
	var b styles.Builder
	b.Add("shrink-0")
	styles.Match(&b, c.Variant, map[Variant]string{
		VariantOutline: "[:where(&)]:size-6",
		VariantSolid:   "[:where(&)]:size-6",
		VariantMini:    "[:where(&)]:size-5",
		VariantMicro:   "[:where(&)]:size-4",
	})
	return b.String()
}
