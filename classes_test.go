package loom_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// stylesheet is the compiled Tailwind output, which this test uses as its
// oracle for what a utility class actually declares.
const stylesheet = "site/static/styles.css"

// unmappedBudget caps the classes in golden output that the selector parser
// below cannot attribute to a rule — arbitrary variants like
// `**:data-[ui=icon]:size-4` and `has-[:disabled]:opacity-75`. Those go
// unchecked, so this is a ratchet: it may fall, never rise.
const unmappedBudget = 45

// TestNoConflictingClasses: within one variant, two utilities must not set the
// same CSS property to different values. One of them is doing nothing, and
// which one wins depends on the order Tailwind happens to emit its rules in.
//
// The oracle is the compiled stylesheet rather than tailwind-merge, which is
// what let the original instance through: its group model treats `outline` as
// style-only, true in Tailwind v3 but not v4, where `outline-2` carries the
// style as well. styles.Merge also skips tailwind-merge entirely when a
// component renders without user classes, so a recipe's own conflicts were
// never examined by anything.
func TestNoConflictingClasses(t *testing.T) {
	css, err := os.ReadFile(stylesheet)
	if err != nil {
		t.Skipf("%s not built; run `make site/css` (audit does)", stylesheet)
	}
	table := utilityTable(string(css))
	if len(table) < 100 {
		t.Fatalf("parsed only %d utilities from %s; the stylesheet format changed", len(table), stylesheet)
	}

	recipes := goldenClassAttrs(t)
	if len(recipes) == 0 {
		t.Fatal("no class attributes found in golden files")
	}

	var unmapped []string
	for _, recipe := range recipes {
		classes := strings.Fields(recipe)
		// Declarations of one property, within one variant.
		type decl struct{ class, value string }
		seen := map[string][]decl{}

		for _, class := range classes {
			props, ok := table[class]
			if !ok {
				unmapped = append(unmapped, class)
				continue
			}
			for prop, value := range props {
				// A value of var(--tw-*) is a deferral, not a declaration:
				// text-xs yields its line-height to whichever leading-*
				// sets --tw-leading. Composition, not conflict.
				if strings.HasPrefix(value, "var(--tw-") {
					continue
				}
				key := variantOf(class) + "|" + prop
				seen[key] = append(seen[key], decl{class, value})
			}
		}

		for key, decls := range seen {
			values := map[string]bool{}
			for _, d := range decls {
				values[d.value] = true
			}
			if len(values) < 2 {
				continue
			}
			variant, prop, _ := strings.Cut(key, "|")
			if variant == "" {
				variant = "(no variant)"
			}
			sort.Slice(decls, func(i, j int) bool { return decls[i].class < decls[j].class })
			var lines []string
			for _, d := range decls {
				lines = append(lines, fmt.Sprintf("      %s sets it to %s", d.class, d.value))
			}
			t.Errorf("conflicting classes in one recipe:\n    [%s] %s is set twice\n%s\n    recipe: %s",
				variant, prop, strings.Join(lines, "\n"), recipe)
		}
	}

	if n := len(distinct(unmapped)); n > unmappedBudget {
		t.Errorf("%d classes could not be attributed to a rule, budget is %d: %v",
			n, unmappedBudget, distinct(unmapped))
	}
}

// utilityTable maps each utility class to the properties it declares.
func utilityTable(css string) map[string]map[string]string {
	table := map[string]map[string]string{}
	eachRule(css, func(selector, body string) {
		class := classOf(selector)
		if class == "" {
			return
		}
		decls := declarationsOf(body)
		if len(decls) == 0 {
			return
		}
		if table[class] == nil {
			table[class] = map[string]string{}
		}
		// A class can be emitted more than once (light and dark, say); the
		// declarations are collected together, which is what a browser
		// resolves anyway.
		for prop, value := range decls {
			table[class][prop] = value
		}
	})
	return table
}

// eachRule calls fn for every style rule, descending through at-rules so that
// utilities nested in @media or @supports are seen too.
func eachRule(css string, fn func(selector, body string)) {
	var selector strings.Builder
	for i := 0; i < len(css); {
		switch css[i] {
		case '{':
			sel := strings.TrimSpace(selector.String())
			selector.Reset()
			depth, j := 1, i+1
			for ; j < len(css) && depth > 0; j++ {
				switch css[j] {
				case '{':
					depth++
				case '}':
					depth--
				}
			}
			body := css[i+1 : j-1]
			if strings.HasPrefix(sel, "@") {
				eachRule(body, fn)
			} else {
				fn(sel, body)
			}
			i = j
		case '}':
			selector.Reset()
			i++
		default:
			selector.WriteByte(css[i])
			i++
		}
	}
}

// classOf returns the utility class a selector targets. Tailwind v4 puts the
// variant in the selector itself — `.focus-visible\:outline-2:focus-visible` —
// so the class name ends at the first unescaped pseudo or combinator.
func classOf(selector string) string {
	if !strings.HasPrefix(selector, ".") {
		return ""
	}
	var b strings.Builder
	for i := 1; i < len(selector); i++ {
		c := selector[i]
		if c == '\\' && i+1 < len(selector) {
			b.WriteByte(selector[i+1])
			i++
			continue
		}
		if strings.IndexByte(" >+~,([:", c) >= 0 {
			break
		}
		b.WriteByte(c)
	}
	return b.String()
}

// declarationsOf splits a rule body into property/value pairs, respecting the
// parentheses in values like var(--x, y) and skipping any nested rule.
func declarationsOf(body string) map[string]string {
	out := map[string]string{}
	depth, start := 0, 0
	flush := func(end int) {
		part := strings.TrimSpace(body[start:end])
		start = end + 1
		if part == "" || strings.Contains(part, "{") {
			return
		}
		prop, value, ok := strings.Cut(part, ":")
		if !ok {
			return
		}
		prop, value = strings.TrimSpace(prop), strings.TrimSpace(value)
		if prop != "" && value != "" {
			out[prop] = value
		}
	}
	for i := 0; i < len(body); i++ {
		switch body[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ';':
			if depth == 0 {
				flush(i)
			}
		}
	}
	flush(len(body))
	return out
}

// variantOf returns a class's variant prefix, so that dark:p-2 and p-2 are not
// mistaken for rivals. Colons inside brackets belong to arbitrary values.
func variantOf(class string) string {
	depth, cut := 0, -1
	for i := 0; i < len(class); i++ {
		switch class[i] {
		case '[':
			depth++
		case ']':
			depth--
		case ':':
			if depth == 0 {
				cut = i
			}
		}
	}
	if cut < 0 {
		return ""
	}
	return class[:cut]
}

var classAttr = regexp.MustCompile(`class="([^"]*)"`)

// goldenClassAttrs returns every distinct class attribute recorded in a golden
// file. Goldens are the rendered truth — post-merge, post-wiring — and cover
// every component without this test having to render anything itself.
func goldenClassAttrs(t *testing.T) []string {
	t.Helper()
	seen := map[string]bool{}
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "site") {
			return fs.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(path, ".golden.html") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, m := range classAttr.FindAllStringSubmatch(string(b), -1) {
			seen[m[1]] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking golden files: %v", err)
	}
	out := make([]string, 0, len(seen))
	for attr := range seen {
		out = append(out, attr)
	}
	sort.Strings(out)
	return out
}

func distinct(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
