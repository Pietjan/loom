package styles_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/internal/styles"
)

func TestBuilder(t *testing.T) {
	var b styles.Builder
	b.Add("inline-flex items-center")
	b.If(true, "gap-2")
	b.If(false, "hidden")
	styles.Match(&b, "primary", map[string]string{
		"primary": "bg-accent text-accent-content",
		"ghost":   "bg-transparent",
	})

	got := b.String()
	for _, want := range []string{"inline-flex", "items-center", "gap-2", "bg-accent", "text-accent-content"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "hidden") || strings.Contains(got, "bg-transparent") {
		t.Fatalf("unexpected classes in %q", got)
	}
}

func TestMergeUserWins(t *testing.T) {
	got := styles.Merge("px-4 py-2 bg-accent", "bg-red-500")
	if strings.Contains(got, "bg-accent") {
		t.Fatalf("component class should lose the conflict: %q", got)
	}
	if !strings.Contains(got, "bg-red-500") {
		t.Fatalf("user class missing: %q", got)
	}
	if !strings.Contains(got, "px-4") || !strings.Contains(got, "py-2") {
		t.Fatalf("non-conflicting classes must survive: %q", got)
	}
}

func TestMergeEmptyUser(t *testing.T) {
	if got := styles.Merge("px-4", ""); got != "px-4" {
		t.Fatalf("got %q", got)
	}
}

func TestSortStable(t *testing.T) {
	// Same set, different input order -> same output.
	a := styles.Sort("text-sm bg-accent flex px-4 hover:bg-accent/90")
	b := styles.Sort("hover:bg-accent/90 px-4 flex bg-accent text-sm")
	if a != b {
		t.Fatalf("sort not canonical:\n%q\n%q", a, b)
	}
	// Layout before spacing before background.
	if !(strings.Index(a, "flex") < strings.Index(a, "px-4") && strings.Index(a, "px-4") < strings.Index(a, "bg-accent")) {
		t.Fatalf("unexpected order: %q", a)
	}
	// Unvarianted before varianted.
	if strings.Index(a, "hover:bg-accent/90") < strings.Index(a, "bg-accent") {
		t.Fatalf("variant should sort after plain utilities: %q", a)
	}
}

func TestSortDeduplicates(t *testing.T) {
	if got := styles.Sort("flex flex px-4"); got != "flex px-4" {
		t.Fatalf("got %q", got)
	}
}

// TestSortIsTotalOrder guards against golden flakiness: tailwind-merge's
// output order is not stable across processes, so Sort must produce the
// same result for any input permutation — including classes that match no
// sort prefix (negative -translate-*) or share an arbitrary variant.
func TestSortIsTotalOrder(t *testing.T) {
	perms := [][]string{
		{"-translate-x-1/2", "-translate-y-1/2", "absolute", "size-5"},
		{"-translate-y-1/2", "size-5", "-translate-x-1/2", "absolute"},
		{"size-5", "absolute", "-translate-y-1/2", "-translate-x-1/2"},
		{"absolute", "-translate-x-1/2", "size-5", "-translate-y-1/2"},
	}
	want := styles.Sort(strings.Join(perms[0], " "))
	for _, p := range perms[1:] {
		if got := styles.Sort(strings.Join(p, " ")); got != want {
			t.Fatalf("Sort not permutation-invariant:\n%q\n%q", want, got)
		}
	}
}

func TestSortKeepsArbitraryValuesIntact(t *testing.T) {
	in := "[&_svg]:size-4 grid-cols-[1fr_auto] supports-[anchor-name:--a]:absolute"
	got := styles.Sort(in)
	for _, tok := range strings.Fields(in) {
		if !strings.Contains(got, tok) {
			t.Fatalf("token %q mangled; got %q", tok, got)
		}
	}
}
