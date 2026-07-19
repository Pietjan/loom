package loom_test

// Tier B form controls: slider, file upload, input group — plus the
// field-through-input-group wiring.

import (
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/pietjan/loom/field"
	"github.com/pietjan/loom/fileupload"
	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/input"
	"github.com/pietjan/loom/inputgroup"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/slider"
)

func TestSlider(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t,
		slider.New(slider.Name("zoom"), slider.Min(1), slider.Max(10), slider.Step(0.5), slider.Value(3))))
	s := tree.One("slider")
	if dom.GetAttr(s, "type") != "range" {
		t.Fatalf("type=%q", dom.GetAttr(s, "type"))
	}
	for k, want := range map[string]string{"name": "zoom", "min": "1", "max": "10", "step": "0.5", "value": "3"} {
		if got := dom.GetAttr(s, k); got != want {
			t.Errorf("%s=%q, want %q", k, got, want)
		}
	}
}

func TestSliderDefaultsRange(t *testing.T) {
	s := testutil.NewTree(t, testutil.Render(t, slider.New(slider.Name("v")))).One("slider")
	if dom.GetAttr(s, "min") != "0" || dom.GetAttr(s, "max") != "100" {
		t.Fatalf("default range min=%q max=%q", dom.GetAttr(s, "min"), dom.GetAttr(s, "max"))
	}
}

func TestFileUpload(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t,
		fileupload.New(fileupload.Name("docs"), fileupload.Accept(".pdf"), fileupload.Multiple())))
	f := tree.One("file")
	if dom.GetAttr(f, "type") != "file" {
		t.Fatalf("type=%q", dom.GetAttr(f, "type"))
	}
	if dom.GetAttr(f, "accept") != ".pdf" || !dom.HasAttr(f, "multiple") {
		t.Fatalf("accept=%q multiple=%v", dom.GetAttr(f, "accept"), dom.HasAttr(f, "multiple"))
	}
}

func TestInputGroupLeadingTrailing(t *testing.T) {
	c := testutil.WithChildren(inputgroup.New(), testutil.Sequence(
		testutil.WithChildren(inputgroup.Addon(), testutil.Text("https://")),
		input.New(input.Name("site")),
		testutil.WithChildren(inputgroup.Addon(), testutil.Text(".com")),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))

	grp := tree.One("input-group")
	addons := dom.FindAll(grp, dom.ByMarker("input-addon"))
	if len(addons) != 2 {
		t.Fatalf("expected 2 addons, got %d", len(addons))
	}
	if dom.Find(grp, dom.ByMarker("input")) == nil {
		t.Fatal("input missing from group")
	}
	// Addon order is source order: leading then trailing.
	if !strings.Contains(textOf(addons[0]), "https://") || !strings.Contains(textOf(addons[1]), ".com") {
		t.Fatalf("addon order wrong: %q, %q", textOf(addons[0]), textOf(addons[1]))
	}
}

// TestFieldWiresThroughInputGroup is the composition test: a field must
// wire its label/aria to the input even when it's wrapped in an
// input-group.
func TestFieldWiresThroughInputGroup(t *testing.T) {
	c := testutil.WithChildren(field.Root(field.Error("bad")), testutil.Sequence(
		testutil.WithChildren(field.Label(), testutil.Text("Website")),
		testutil.WithChildren(inputgroup.New(), testutil.Sequence(
			testutil.WithChildren(inputgroup.Addon(), testutil.Text("https://")),
			input.New(input.Name("site")),
		)),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))

	control := tree.One("input")
	label := tree.One("field-label")
	id := dom.GetAttr(control, "id")
	if id == "" {
		t.Fatal("control (inside input-group) did not get an id from the field")
	}
	if dom.GetAttr(label, "for") != id {
		t.Fatalf("label for=%q, control id=%q — field did not wire through the group", dom.GetAttr(label, "for"), id)
	}
	if dom.GetAttr(control, "aria-invalid") != "true" {
		t.Fatal("error state did not reach the control inside the group")
	}
	if dom.GetAttr(control, "aria-describedby") == "" {
		t.Fatal("describedby did not reach the control inside the group")
	}
}

func TestTierBGoldens(t *testing.T) {
	t.Run("slider", func(t *testing.T) {
		testutil.Golden(t, "slider", slider.New(slider.Name("volume"), slider.Value(50)))
	})
	t.Run("fileupload", func(t *testing.T) {
		testutil.Golden(t, "fileupload", fileupload.New(fileupload.Name("avatar"), fileupload.Accept("image/*")))
	})
	t.Run("inputgroup", func(t *testing.T) {
		testutil.Golden(t, "inputgroup", testutil.WithChildren(inputgroup.New(), testutil.Sequence(
			testutil.WithChildren(inputgroup.Addon(), icon.New(icon.MagnifyingGlass, icon.Small)),
			input.New(input.Name("q"), input.Placeholder("Search")),
		)))
	})
}

func textOf(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}
