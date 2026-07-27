package theme

import (
	"fmt"
	"strings"
	"testing"
)

// The generated CSS is what runtime theme switching stands on, and it fails
// silently: a missing accent or a block in the wrong order still parses, and
// only shows up as a page that keeps the theme it was asked to leave.
func TestGenerateAll(t *testing.T) {
	out, err := GenerateAll()
	if err != nil {
		t.Fatal(err)
	}
	css := string(out)

	for name := range Accents {
		for _, selector := range []string{
			fmt.Sprintf("html[data-accent=%q]", name),
			fmt.Sprintf("html:not(.light)[data-accent=%q]", name),
			fmt.Sprintf("html.dark[data-accent=%q]", name),
		} {
			if !strings.Contains(css, selector) {
				t.Errorf("missing %s", selector)
			}
		}
	}

	// An accent carries the base ramp it pairs with, so a standalone base
	// block ties with it on specificity - source order is the only thing
	// that lets an explicit base win.
	lastAccent := strings.LastIndex(css, "[data-accent=")
	firstBase := strings.Index(css, "[data-base=")
	if firstBase < lastAccent {
		t.Error("base blocks are emitted before the accents they must override")
	}

	for _, base := range neutrals {
		if !strings.Contains(css, fmt.Sprintf("html[data-base=%q]", base)) {
			t.Errorf("missing base %s", base)
		}
	}
}
