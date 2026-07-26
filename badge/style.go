package badge

import (
	"github.com/pietjan/loom/internal/palette"
	"github.com/pietjan/loom/internal/styles"
)

func classes(c Config) string {
	var b styles.Builder
	b.Add("inline-flex items-center gap-1 font-medium whitespace-nowrap")
	// Icons shrink to fit the badge - CSS, not tree surgery.
	b.Add("**:data-[ui=icon]:size-4")
	styles.Match(&b, c.Size, map[Size]string{
		SizeSmall: "text-xs px-1.5 py-0.5 **:data-[ui=icon]:size-3",
		SizeBase:  "text-sm px-2 py-1 **:data-[ui=icon]:size-4",
		SizeLarge: "text-sm px-2.5 py-1.5 **:data-[ui=icon]:size-5",
	})
	b.If(c.pill, "rounded-full")
	b.If(!c.pill, "rounded-md")
	if s, ok := palette.Of(palette.Color(c.Color)); ok {
		if c.solid {
			b.Add(s.Solid)
		} else {
			b.Add(s.Tint)
		}
	}
	return b.String()
}
