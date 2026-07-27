package pages

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The picker lists theme names, the stylesheet defines them, and the two are
// generated from different places - the Go table in cmd/css/internal/theme and
// a hand-ordered list here, because hue order is presentation. A swatch with
// no rule behind it is a dead chip that silently applies nothing, and a rule
// with no swatch is a theme nobody can reach.
func TestPickerMatchesStylesheet(t *testing.T) {
	css, err := os.ReadFile(stylesheet)
	if err != nil {
		t.Skipf("%s not built; run make site/css", stylesheet)
	}

	for _, tc := range []struct {
		axis    string
		listed  []string
		pattern *regexp.Regexp
	}{
		{"accent", accentNames, regexp.MustCompile(`html\[data-accent=["']?([a-z]+)["']?\]`)},
		{"base", baseNames, regexp.MustCompile(`html\[data-base=["']?([a-z]+)["']?\]`)},
	} {
		defined := map[string]bool{}
		for _, m := range tc.pattern.FindAllStringSubmatch(string(css), -1) {
			defined[m[1]] = true
		}
		if len(defined) == 0 {
			t.Fatalf("no %s rules in %s; the themes step of make site/css did not run", tc.axis, stylesheet)
		}

		listed := map[string]bool{}
		for _, name := range tc.listed {
			listed[name] = true
			if !defined[name] {
				t.Errorf("picker offers %s %q with no rule in the stylesheet", tc.axis, name)
			}
			// A swatch draws itself with the generated class, so a missing
			// one is an invisible chip rather than a wrong color.
			class := fmt.Sprintf(".loom-swatch-%s", name)
			if tc.axis == "base" {
				class = fmt.Sprintf(".loom-swatch-base-%s", name)
			}
			if !strings.Contains(string(css), class+"{") {
				t.Errorf("no %s swatch color for %q", tc.axis, name)
			}
			// The Base chip borrows the monochrome accent's swatch for
			// whichever neutral is in force, so those need one too.
			if tc.axis == "base" && !strings.Contains(string(css), fmt.Sprintf(".loom-swatch-%s{", name)) {
				t.Errorf("no monochrome accent swatch for base %q; the Base chip cannot draw it", name)
			}
		}
		for name := range defined {
			// The neutral accents are reachable as Base plus a base chip
			// rather than as accent chips of their own.
			if tc.axis == "accent" && contains(baseNames, name) {
				continue
			}
			if !listed[name] {
				t.Errorf("stylesheet defines %s %q that the picker does not offer", tc.axis, name)
			}
		}
	}
}

func contains(list []string, want string) bool {
	for _, got := range list {
		if got == want {
			return true
		}
	}
	return false
}

// The picker reads the theme in force out of the stylesheet rather than
// tracking it - without these two, no chip can be marked on a page nobody has
// themed yet, and the Base chip has no neutral to draw.
func TestThemeNamesAreExposed(t *testing.T) {
	css, err := os.ReadFile(stylesheet)
	if err != nil {
		t.Skipf("%s not built; run make site/css", stylesheet)
	}
	for _, v := range []string{"--loom-accent-name", "--loom-base-name"} {
		if !strings.Contains(string(css), v+":") {
			t.Errorf("%s missing from the stylesheet", v)
		}
	}
}

// stylesheet is the compiled output, the same oracle the library's class test
// uses - the picker is only correct against the CSS that actually shipped.
// Relative to this package, which is where go test runs it from.
const stylesheet = "../static/styles.css"
