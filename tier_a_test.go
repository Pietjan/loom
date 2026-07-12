package loom_test

// Goldens + behavioral assertions for the Tier A markup/CSS components.

import (
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/pietjan/loom/avatar"
	"github.com/pietjan/loom/badge"
	"github.com/pietjan/loom/breadcrumbs"
	"github.com/pietjan/loom/description"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/kbd"
	"github.com/pietjan/loom/pagination"
	"github.com/pietjan/loom/progress"
	"github.com/pietjan/loom/skeleton"
	"github.com/pietjan/loom/stat"
	"github.com/pietjan/loom/timeline"
)

func TestAvatar(t *testing.T) {
	// Image avatar renders an <img> with the alt text.
	img := testutil.NewTree(t, testutil.Render(t,
		avatar.New(avatar.Src("/u/olivia.jpg"), avatar.Alt("Olivia"))))
	pic := dom.Find(img.One("avatar"), func(n *html.Node) bool { return n.Data == "img" })
	if pic == nil || dom.GetAttr(pic, "alt") != "Olivia" {
		t.Fatal("image avatar should render an <img> with alt")
	}

	// Initials fallback carries an accessible name.
	ini := testutil.NewTree(t, testutil.Render(t,
		avatar.New(avatar.Initials("OM"), avatar.Alt("Olivia Martin"), avatar.Online)))
	a := ini.One("avatar")
	if dom.GetAttr(a, "role") != "img" || dom.GetAttr(a, "aria-label") != "Olivia Martin" {
		t.Fatalf("initials avatar needs role=img + aria-label, got role=%q label=%q",
			dom.GetAttr(a, "role"), dom.GetAttr(a, "aria-label"))
	}
	if !strings.Contains(testutil.Render(t, avatar.New(avatar.Initials("OM"), avatar.Online)), "bg-green-500") {
		t.Fatal("online status dot missing")
	}

	testutil.Golden(t, "avatar", avatar.New(avatar.Initials("OM"), avatar.Alt("Olivia Martin"), avatar.Online))
}

func TestProgressValueAndAria(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, progress.New(progress.Value(3, progress.Of(10)))))
	bar := tree.One("progress")
	if dom.GetAttr(bar, "role") != "progressbar" {
		t.Fatal("missing role=progressbar")
	}
	if dom.GetAttr(bar, "aria-valuenow") != "3" || dom.GetAttr(bar, "aria-valuemax") != "10" {
		t.Fatalf("aria values: now=%q max=%q", dom.GetAttr(bar, "aria-valuenow"), dom.GetAttr(bar, "aria-valuemax"))
	}
	if fill := tree.One("progress-bar"); !strings.Contains(dom.GetAttr(fill, "style"), "width: 30%") {
		t.Fatalf("bar width: %q", dom.GetAttr(fill, "style"))
	}

	// Indeterminate: no value, animation hook present.
	ind := testutil.NewTree(t, testutil.Render(t, progress.New(progress.Indeterminate())))
	if !dom.HasAttr(ind.One("progress"), "data-indeterminate") {
		t.Fatal("indeterminate marker missing")
	}
}

func TestProgressClamps(t *testing.T) {
	out := testutil.Render(t, progress.New(progress.Value(250, progress.Of(100))))
	if !strings.Contains(out, "width: 100%") {
		t.Fatalf("over-max value should clamp to 100%%: %s", out)
	}
}

func TestPaginationCurrentAndDisabled(t *testing.T) {
	c := testutil.WithChildren(pagination.New(), testutil.Sequence(
		pagination.Prev(""), // disabled
		testutil.WithChildren(pagination.Item("?page=1", pagination.Current()), testutil.Text("1")),
		testutil.WithChildren(pagination.Item("?page=2"), testutil.Text("2")),
		pagination.Gap(),
		testutil.WithChildren(pagination.Item("?page=9"), testutil.Text("9")),
		pagination.Next("?page=2"),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))

	items := dom.FindAll(tree.Root, dom.ByMarker("pagination-item"))
	current := 0
	for _, it := range items {
		if dom.GetAttr(it, "aria-current") == "page" {
			current++
		}
	}
	if current != 1 {
		t.Fatalf("expected exactly one current page, got %d", current)
	}
	// Disabled Prev is an inert <span>, not a link.
	prev := items[0]
	if prev.Data != "span" || dom.GetAttr(prev, "aria-disabled") != "true" {
		t.Fatalf("disabled prev should be an inert span, got <%s aria-disabled=%q>", prev.Data, dom.GetAttr(prev, "aria-disabled"))
	}
	if dom.Find(tree.Root, dom.ByMarker("pagination-gap")) == nil {
		t.Fatal("gap missing")
	}
}

func TestBreadcrumbsCurrent(t *testing.T) {
	c := testutil.WithChildren(breadcrumbs.New(), testutil.Sequence(
		testutil.WithChildren(breadcrumbs.Item("/"), testutil.Text("Home")),
		testutil.WithChildren(breadcrumbs.Item("", breadcrumbs.Current()), testutil.Text("Loom")),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))
	last := dom.FindAll(tree.Root, dom.ByMarker("breadcrumbs-item"))[1]
	span := dom.Find(last, func(n *html.Node) bool { return n.Data == "span" })
	if span == nil || dom.GetAttr(span, "aria-current") != "page" {
		t.Fatal("current crumb should be a span with aria-current=page")
	}
	if dom.Find(last, func(n *html.Node) bool { return n.Data == "a" }) != nil {
		t.Fatal("current crumb should not be a link")
	}
}

func TestGoldens(t *testing.T) {
	t.Run("breadcrumbs", func(t *testing.T) {
		testutil.Golden(t, "breadcrumbs", testutil.WithChildren(breadcrumbs.New(), testutil.Sequence(
			testutil.WithChildren(breadcrumbs.Item("/"), testutil.Text("Home")),
			testutil.WithChildren(breadcrumbs.Item("/projects"), testutil.Text("Projects")),
			testutil.WithChildren(breadcrumbs.Item("", breadcrumbs.Current()), testutil.Text("Loom")),
		)))
	})
	t.Run("progress", func(t *testing.T) {
		testutil.Golden(t, "progress", progress.New(progress.Value(70)))
	})
	t.Run("skeleton", func(t *testing.T) {
		testutil.Golden(t, "skeleton", skeleton.New(skeleton.Circle, skeleton.Class("size-10")))
	})
	t.Run("stat", func(t *testing.T) {
		testutil.Golden(t, "stat", testutil.WithChildren(
			stat.New(stat.Label("Revenue"), stat.Value("$48.2k")),
			testutil.WithChildren(badge.New(badge.Green), testutil.Text("+12%"))))
	})
	t.Run("timeline", func(t *testing.T) {
		testutil.Golden(t, "timeline", testutil.WithChildren(timeline.New(), testutil.Sequence(
			testutil.WithChildren(timeline.Item(timeline.Title("Deployed"), timeline.Time("2h ago")), testutil.Text("Shipped.")),
			testutil.WithChildren(timeline.Item(timeline.Title("Merged")), testutil.Text("Landed.")),
		)))
	})
	t.Run("description", func(t *testing.T) {
		testutil.Golden(t, "description", testutil.WithChildren(description.New(), testutil.Sequence(
			testutil.WithChildren(description.Term(), testutil.Text("Plan")),
			testutil.WithChildren(description.Detail(), testutil.Text("Pro")),
		)))
	})
	t.Run("kbd", func(t *testing.T) {
		testutil.Golden(t, "kbd", testutil.WithChildren(kbd.New(), testutil.Text("K")))
	})
}
