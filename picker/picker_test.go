package picker_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/picker"
)

// Raw string assertions: the golden Parse/Format path would strip the
// customizable-select internals, because x/net/html's in-select parsing
// mode predates them. Serialization (what browsers receive) is correct —
// that is what we assert on.
func TestCustomizableSelectMarkup(t *testing.T) {
	c := testutil.WithChildren(
		picker.New(picker.Name("pet"), picker.Placeholder("Choose a pet")),
		testutil.Sequence(
			testutil.WithChildren(picker.Item("cat"), testutil.Sequence(
				icon.New(icon.Heart, icon.ExtraSmall), testutil.Text("Cat"))),
			testutil.WithChildren(picker.Item("dog"), testutil.Text("Dog")),
		),
	)
	out := testutil.Render(t, c)

	for _, want := range []string{
		`data-ui="picker"`,
		`name="pet"`,
		`<button`,
		`<selectedcontent></selectedcontent>`,
		`<option value="" disabled="" hidden="" selected=""`, // placeholder selected: nothing else claimed it
		`<option data-ui="picker-option" value="cat"`,
		`data-ui="icon"`,       // rich option content survives
		`rotate-180`,           // icon arrow replaces ::picker-icon
		`data-picker-check=""`, // icon checkmark replaces ::checkmark
		`in-checked:visible`,   // ...and shows only while checked
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestPlaceholderYieldsToSelectedItem(t *testing.T) {
	c := testutil.WithChildren(
		picker.New(picker.Name("pet"), picker.Placeholder("Choose")),
		testutil.WithChildren(picker.Item("dog", picker.Selected()), testutil.Text("Dog")),
	)
	out := testutil.Render(t, c)

	if strings.Contains(out, `hidden="" selected=""`) || strings.Contains(out, `selected="" hidden=""`) {
		t.Fatalf("placeholder must not be selected when an item is:\n%s", out)
	}
	if !strings.Contains(out, `value="dog" selected=""`) {
		t.Fatalf("selected item lost its selection:\n%s", out)
	}
}
