package navbar_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/navbar"
)

func TestCurrentAndBadge(t *testing.T) {
	c := testutil.WithChildren(navbar.New(navbar.Label("Main")), testutil.Sequence(
		testutil.WithChildren(navbar.Item("/", navbar.Current()),
			testutil.Sequence(icon.New(icon.Home, icon.Mini), testutil.Text("Home"))),
		testutil.WithChildren(navbar.Item("/inbox", navbar.Badge("12")),
			testutil.Sequence(icon.New(icon.Inbox, icon.Mini), testutil.Text("Inbox"))),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))

	nav := tree.One("navbar")
	if dom.GetAttr(nav, "aria-label") != "Main" {
		t.Fatalf("nav aria-label: %q", dom.GetAttr(nav, "aria-label"))
	}

	items := dom.FindAll(tree.Root, dom.ByMarker("navbar-item"))
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if dom.GetAttr(items[0], "aria-current") != "page" {
		t.Fatal("first item should be current")
	}
	if dom.HasAttr(items[1], "aria-current") {
		t.Fatal("second item should not be current")
	}

	out := testutil.Render(t, c)
	if !strings.Contains(out, ">12<") {
		t.Fatalf("badge not rendered: %s", out)
	}
}

func TestGolden(t *testing.T) {
	c := testutil.WithChildren(navbar.New(navbar.Label("Main")), testutil.Sequence(
		testutil.WithChildren(navbar.Item("/", navbar.Current()), testutil.Text("Home")),
		testutil.WithChildren(navbar.Item("/inbox"), testutil.Text("Inbox")),
	))
	testutil.Golden(t, "navbar", c)
}
