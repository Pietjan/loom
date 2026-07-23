package loom_test

// Tier C: popover, carousel, flash.

import (
	"errors"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/button"
	"github.com/pietjan/loom/carousel"
	"github.com/pietjan/loom/flash"
	"github.com/pietjan/loom/heading"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/popover"
	"github.com/pietjan/loom/text"
)

func fullPopover() templ.Component {
	return testutil.WithChildren(popover.Root(), testutil.Sequence(
		testutil.WithChildren(popover.Trigger(),
			testutil.WithChildren(button.New(button.Ghost), testutil.Text("Details"))),
		testutil.WithChildren(popover.Content(), testutil.Sequence(
			testutil.WithChildren(heading.New(), testutil.Text("Storage")),
			testutil.WithChildren(text.New(), testutil.Text("8.2 GB of 10 GB used.")),
		)),
	))
}

func init() {
	composites["popover"] = fullPopover
}

func TestPopoverWiring(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, fullPopover()))

	panel := tree.One("popover-content")
	if !dom.HasAttr(panel, "popover") {
		t.Fatal("content is not a popover")
	}
	btn := dom.Find(tree.Root, dom.ByAttr("commandfor"))
	if btn == nil || dom.GetAttr(btn, "command") != "toggle-popover" {
		t.Fatal("trigger not wired to toggle-popover")
	}
	if dom.GetAttr(btn, "commandfor") != dom.GetAttr(panel, "id") {
		t.Fatal("commandfor does not match panel id")
	}
	// Anchor pair.
	name := strings.TrimSpace(strings.TrimPrefix(dom.GetAttr(btn, "style"), "anchor-name:"))
	if !strings.Contains(dom.GetAttr(panel, "style"), "position-anchor: "+name) {
		t.Fatalf("panel style %q does not reference anchor %q", dom.GetAttr(panel, "style"), name)
	}
}

func TestPopoverOutsideRootFails(t *testing.T) {
	if err := testutil.RenderErr(testutil.WithChildren(popover.Content(), testutil.Text("x"))); !errors.Is(err, popover.ErrNoScope) {
		t.Fatalf("expected ErrNoScope, got %v", err)
	}
}

func TestCarouselDotsTargetSlides(t *testing.T) {
	c := testutil.WithChildren(carousel.New(), testutil.Sequence(
		testutil.WithChildren(carousel.Slide(), testutil.Text("One")),
		testutil.WithChildren(carousel.Slide(), testutil.Text("Two")),
		testutil.WithChildren(carousel.Slide(), testutil.Text("Three")),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))

	slides := dom.FindAll(tree.Root, dom.ByMarker("carousel-slide"))
	if len(slides) != 3 {
		t.Fatalf("expected 3 slides, got %d", len(slides))
	}
	// Every slide has an id, and every dot links to one that exists.
	ids := map[string]bool{}
	for _, s := range slides {
		id := dom.GetAttr(s, "id")
		if id == "" {
			t.Fatal("slide missing id")
		}
		ids[id] = true
	}
	dots := dom.FindAll(dom.Find(tree.Root, dom.ByMarker("carousel-dots")), dom.ByTag(atom.A))
	if len(dots) != 3 {
		t.Fatalf("expected 3 dots, got %d", len(dots))
	}
	for _, d := range dots {
		href := strings.TrimPrefix(dom.GetAttr(d, "href"), "#")
		if !ids[href] {
			t.Fatalf("dot targets #%s which is not a slide", href)
		}
	}
}

func TestCarouselSingleSlideNoDots(t *testing.T) {
	c := testutil.WithChildren(carousel.New(),
		testutil.WithChildren(carousel.Slide(), testutil.Text("Only")))
	tree := testutil.NewTree(t, testutil.Render(t, c))
	if dom.Find(tree.Root, dom.ByMarker("carousel-dots")) != nil {
		t.Fatal("a single-slide carousel should not render dots")
	}
}

func TestFlashDismissibleAndRole(t *testing.T) {
	// Danger is assertive.
	d := testutil.NewTree(t, testutil.Render(t,
		testutil.WithChildren(flash.New(flash.Danger), testutil.Text("Boom"))))
	if got := dom.GetAttr(d.One("flash"), "role"); got != "alert" {
		t.Fatalf("danger flash role=%q, want alert", got)
	}

	// Dismissible flash carries the checkbox-hack machinery.
	c := testutil.WithChildren(flash.New(flash.Success, flash.Dismissible()), testutil.Text("Saved"))
	out := testutil.Render(t, c)
	tree := testutil.NewTree(t, out)
	root := tree.One("flash")
	if got := dom.GetAttr(root, "role"); got != "status" {
		t.Fatalf("success flash role=%q, want status", got)
	}
	if !strings.Contains(dom.GetAttr(root, "class"), "has-checked:hidden") {
		t.Fatal("dismissible flash needs the has-checked hide hook")
	}
	if dom.Find(root, func(n *html.Node) bool { return n.Data == "input" && dom.GetAttr(n, "type") == "checkbox" }) == nil {
		t.Fatal("dismissible flash needs a checkbox")
	}
}

func TestFlashAutohide(t *testing.T) {
	out := testutil.Render(t, testutil.WithChildren(flash.New(flash.Info, flash.Autohide()), testutil.Text("Heads up")))
	if !strings.Contains(out, "data-autohide") {
		t.Fatalf("autohide flash missing data-autohide: %s", out)
	}
}

func TestTierCGoldens(t *testing.T) {
	testutil.Golden(t, "popover", fullPopover())
	testutil.Golden(t, "carousel", testutil.WithChildren(carousel.New(), testutil.Sequence(
		testutil.WithChildren(carousel.Slide(), testutil.Text("One")),
		testutil.WithChildren(carousel.Slide(), testutil.Text("Two")),
	)))
	testutil.Golden(t, "flash", testutil.WithChildren(flash.New(flash.Success, flash.Dismissible()), testutil.Text("Your changes were saved.")))
}
