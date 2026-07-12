package field_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/input"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

func render(t *testing.T, opts ...field.Option) *testutil.Tree {
	t.Helper()
	c := testutil.WithChildren(field.Root(opts...), testutil.Sequence(
		testutil.WithChildren(field.Label(), testutil.Text("Email")),
		input.New(input.Type("email"), input.Name("email")),
		testutil.WithChildren(field.Description(), testutil.Text("We never share it.")),
	))
	return testutil.NewTree(t, testutil.Render(t, c))
}

func TestLabelForMatchesControlID(t *testing.T) {
	tree := render(t)
	label := tree.One("field-label")
	control := tree.One("input")

	id := dom.GetAttr(control, "id")
	if id == "" {
		t.Fatal("control did not adopt an id")
	}
	if got := dom.GetAttr(label, "for"); got != id {
		t.Fatalf("label for=%q, control id=%q", got, id)
	}
}

func TestDescribedByListsRenderedPartsOnly(t *testing.T) {
	// With description, without error.
	tree := render(t)
	control := tree.One("input")
	desc := tree.One("field-description")

	got := dom.GetAttr(control, "aria-describedby")
	if got != dom.GetAttr(desc, "id") {
		t.Fatalf("aria-describedby=%q, want just description id %q", got, dom.GetAttr(desc, "id"))
	}

	// With description AND error.
	tree = render(t, field.Error("Invalid address"))
	control = tree.One("input")
	ids := strings.Fields(dom.GetAttr(control, "aria-describedby"))
	if len(ids) != 2 {
		t.Fatalf("expected 2 describedby ids, got %v", ids)
	}
	if ids[0] != dom.GetAttr(tree.One("field-description"), "id") ||
		ids[1] != dom.GetAttr(tree.One("field-error"), "id") {
		t.Fatalf("describedby ids %v don't match part ids", ids)
	}
}

func TestErrorStateFlowsToControl(t *testing.T) {
	tree := render(t, field.Error("Invalid address"))
	control := tree.One("input")

	if got := dom.GetAttr(control, "aria-invalid"); got != "true" {
		t.Fatalf("aria-invalid=%q", got)
	}
	if !strings.Contains(dom.GetAttr(control, "class"), "border-red-500") {
		t.Fatal("control missing invalid styling")
	}
	if errEl := tree.One("field-error"); errEl.FirstChild == nil || errEl.FirstChild.Data != "Invalid address" {
		t.Fatal("error message not rendered")
	}
}

func TestRequiredAndDisabledFlowToControl(t *testing.T) {
	tree := render(t, field.Required(), field.Disabled())
	control := tree.One("input")
	if !dom.HasAttr(control, "required") {
		t.Fatal("control not required")
	}
	if !dom.HasAttr(control, "disabled") {
		t.Fatal("control not disabled")
	}
}

func TestUserSuppliedControlIDWins(t *testing.T) {
	c := testutil.WithChildren(field.Root(), testutil.Sequence(
		testutil.WithChildren(field.Label(), testutil.Text("Email")),
		input.New(input.ID("my-email")),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))
	if got := dom.GetAttr(tree.One("input"), "id"); got != "my-email" {
		t.Fatalf("control id=%q, want user-supplied my-email", got)
	}
	if got := dom.GetAttr(tree.One("field-label"), "for"); got != "my-email" {
		t.Fatalf("label for=%q, must follow user-supplied id", got)
	}
}

func TestGolden(t *testing.T) {
	c := testutil.WithChildren(field.Root(field.Error("Invalid address"), field.Required()), testutil.Sequence(
		testutil.WithChildren(field.Label(), testutil.Text("Email")),
		input.New(input.Type("email"), input.Name("email")),
		testutil.WithChildren(field.Description(), testutil.Text("We never share it.")),
	))
	testutil.Golden(t, "field-full", c)
}
