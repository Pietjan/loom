package button_test

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/button"
	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

var (
	withChildren = testutil.WithChildren
	textChild    = testutil.Text
)

func TestGoldenVariants(t *testing.T) {
	for name, opt := range map[string]button.Option{
		"outline": button.Outline,
		"primary": button.Primary,
		"danger":  button.Danger,
		"ghost":   button.Ghost,
	} {
		t.Run(name, func(t *testing.T) {
			testutil.Golden(t, "button-"+name, withChildren(button.New(opt), textChild("Save")))
		})
	}
}

func TestDefaults(t *testing.T) {
	tree := testutil.Parse(t, testutil.Render(t, withChildren(button.New(), textChild("Go"))))
	btn := dom.Find(tree, dom.ByMarker("button"))
	if btn == nil {
		t.Fatal("no button marker")
	}
	if got := dom.GetAttr(btn, "type"); got != "button" {
		t.Fatalf("default type: %q", got)
	}
}

func TestUserClassWinsConflict(t *testing.T) {
	out := testutil.Render(t, withChildren(button.New(button.Class("px-8")), textChild("Wide")))
	if strings.Contains(out, "px-4") {
		t.Fatalf("component px-4 should lose to user px-8: %s", out)
	}
	if !strings.Contains(out, "px-8") {
		t.Fatalf("user class missing: %s", out)
	}
}

func TestIconOnlyRequiresAccessibleName(t *testing.T) {
	iconChild := icon.New(icon.X)

	err := testutil.RenderErr(withChildren(button.New(), iconChild))
	if !errors.Is(err, button.ErrNoAccessibleName) {
		t.Fatalf("expected ErrNoAccessibleName, got %v", err)
	}

	out := testutil.Render(t, withChildren(button.New(button.Label("Close")), iconChild))
	tree := testutil.Parse(t, out)
	btn := dom.Find(tree, dom.ByMarker("button"))
	if got := dom.GetAttr(btn, "aria-label"); got != "Close" {
		t.Fatalf("aria-label: %q", got)
	}
	if !strings.Contains(dom.GetAttr(btn, "class"), "w-10") {
		t.Fatalf("icon-only button should be square: %s", dom.GetAttr(btn, "class"))
	}
}

func TestButtonWithTextAndIconIsNotSquare(t *testing.T) {
	both := testutil.Sequence(icon.New(icon.Plus), testutil.Text("Add"))
	out := testutil.Render(t, withChildren(button.New(), both))
	tree := testutil.Parse(t, out)
	btn := dom.Find(tree, dom.ByMarker("button"))
	if strings.Contains(dom.GetAttr(btn, "class"), "w-10") {
		t.Fatal("button with text must not be square")
	}
}

func TestDisabled(t *testing.T) {
	out := testutil.Render(t, withChildren(button.New(button.Disabled()), textChild("Nope")))
	tree := testutil.Parse(t, out)
	btn := dom.Find(tree, dom.ByMarker("button"))
	if !dom.HasAttr(btn, "disabled") {
		t.Fatal("expected disabled attribute")
	}
}

func TestHrefRendersAnchor(t *testing.T) {
	out := testutil.Render(t, withChildren(button.New(button.Primary, button.Href("/signup")), textChild("Join")))
	tree := testutil.Parse(t, out)
	a := dom.Find(tree, dom.ByMarker("button"))
	if a.DataAtom != atom.A {
		t.Fatalf("expected <a>, got <%s>", a.Data)
	}
	if got := dom.GetAttr(a, "href"); got != "/signup" {
		t.Fatalf("href: %q", got)
	}
	if dom.HasAttr(a, "type") {
		t.Fatal("anchor must not carry a type attribute")
	}
	// Variant styling is shared with the <button> form.
	if !strings.Contains(dom.GetAttr(a, "class"), "bg-accent") {
		t.Fatalf("primary styling missing: %s", dom.GetAttr(a, "class"))
	}
}

func TestHrefExternal(t *testing.T) {
	c := button.New(button.Href("https://example.com"), button.External())
	out := testutil.Render(t, withChildren(c, textChild("Docs")))
	tree := testutil.Parse(t, out)
	a := dom.Find(tree, dom.ByMarker("button"))
	if got := dom.GetAttr(a, "target"); got != "_blank" {
		t.Fatalf("target: %q", got)
	}
	if got := dom.GetAttr(a, "rel"); got != "noopener noreferrer" {
		t.Fatalf("rel: %q", got)
	}
}

func TestDisabledHrefIsInertSpan(t *testing.T) {
	c := button.New(button.Href("/signup"), button.Disabled())
	out := testutil.Render(t, withChildren(c, textChild("Join")))
	tree := testutil.Parse(t, out)
	n := dom.Find(tree, dom.ByMarker("button"))
	if n.DataAtom != atom.Span {
		t.Fatalf("expected inert <span>, got <%s>", n.Data)
	}
	if dom.HasAttr(n, "href") {
		t.Fatal("disabled link button must not keep its href")
	}
	if got := dom.GetAttr(n, "aria-disabled"); got != "true" {
		t.Fatalf("aria-disabled: %q", got)
	}
}

func TestHrefWithTypeIsRejected(t *testing.T) {
	c := button.New(button.Href("/signup"), button.Submit)
	if err := testutil.RenderErr(withChildren(c, textChild("Join"))); !errors.Is(err, button.ErrHrefWithType) {
		t.Fatalf("expected ErrHrefWithType, got %v", err)
	}
}

func TestHrefIconOnlyRequiresAccessibleName(t *testing.T) {
	c := button.New(button.Href("/close"))
	if err := testutil.RenderErr(withChildren(c, icon.New(icon.X))); !errors.Is(err, button.ErrNoAccessibleName) {
		t.Fatalf("expected ErrNoAccessibleName, got %v", err)
	}
}

func TestGroup(t *testing.T) {
	buttons := testutil.Sequence(
		withChildren(button.New(), textChild("One")),
		withChildren(button.New(), textChild("Two")),
	)
	out := testutil.Render(t, withChildren(button.Group(), buttons))
	tree := testutil.Parse(t, out)
	g := dom.Find(tree, dom.ByMarker("button-group"))
	if g == nil {
		t.Fatal("no group marker")
	}
	if got := dom.GetAttr(g, "role"); got != "group" {
		t.Fatalf("role: %q", got)
	}
	if got := len(dom.FindAll(g, dom.ByTag(atom.Button))); got != 2 {
		t.Fatalf("expected 2 buttons in group, got %d", got)
	}
}
