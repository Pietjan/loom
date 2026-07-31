package input_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/input"
	"github.com/pietjan/loom/inputgroup"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

func render(t *testing.T, options ...input.Option) string {
	t.Helper()
	tree := testutil.Parse(t, testutil.Render(t, input.New(options...)))
	el := dom.Find(tree, dom.ByMarker("input"))
	if el == nil {
		t.Fatal("no input marker")
	}
	return dom.GetAttr(el, "class")
}

// TestSizes. Small exists so a field and a button can sit on one row, which
// means its height has to be button.Small's - h-8 - and not merely smaller.
func TestSizes(t *testing.T) {
	base := render(t)
	small := render(t, input.Small)

	for _, want := range []string{"h-10", "rounded-lg"} {
		if !strings.Contains(base, want) {
			t.Errorf("default input has no %q: %q", want, base)
		}
	}
	for _, want := range []string{"h-8", "rounded-md"} {
		if !strings.Contains(small, want) {
			t.Errorf("small input has no %q: %q", want, small)
		}
	}
	if strings.Contains(small, "h-10") {
		t.Errorf("small input kept the default height: %q", small)
	}
}

// TestSizeLeavesEverythingElseAlone: a size is a size, not a variant.
func TestSizeLeavesEverythingElseAlone(t *testing.T) {
	small := render(t, input.Small)
	for _, want := range []string{"bg-white", "border", "shadow-xs", "placeholder:text-base-400"} {
		if !strings.Contains(small, want) {
			t.Errorf("small input lost %q: %q", want, small)
		}
	}
}

// TestGroupedIgnoresSize. Inside an input group the group draws the shell
// and sets the height, and the input fills it - so a size here would be an
// input arguing with its own container.
func TestGroupedIgnoresSize(t *testing.T) {
	tree := testutil.Parse(t, testutil.Render(t,
		testutil.WithChildren(inputgroup.New(), input.New(input.Small))))
	el := dom.Find(tree, dom.ByMarker("input"))
	if el == nil {
		t.Fatal("no input marker")
	}

	class := dom.GetAttr(el, "class")
	for _, unwanted := range []string{"h-8", "h-10", "rounded-md", "rounded-lg"} {
		if strings.Contains(class, unwanted) {
			t.Errorf("grouped input states its own %q: %q", unwanted, class)
		}
	}
	if !strings.Contains(class, "h-full") {
		t.Errorf("grouped input does not fill the group: %q", class)
	}
}
