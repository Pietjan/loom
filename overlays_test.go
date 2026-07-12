package loom_test

// Phase 3 composites: contract registrations + behavioral tests for the
// invoker-command wiring and the fail-loud pairing rules.

import (
	"errors"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/accordion"
	"github.com/pietjan/loom/badge"
	"github.com/pietjan/loom/button"
	"github.com/pietjan/loom/dropdown"
	"github.com/pietjan/loom/header"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/modal"
	"github.com/pietjan/loom/navbar"
	"github.com/pietjan/loom/navlist"
	"github.com/pietjan/loom/sidebar"
	"github.com/pietjan/loom/table"
	"github.com/pietjan/loom/tabs"
	"github.com/pietjan/loom/tooltip"
)

func fullModal() templ.Component {
	return testutil.WithChildren(modal.Root(), testutil.Sequence(
		testutil.WithChildren(modal.Trigger(),
			testutil.WithChildren(button.New(), testutil.Text("Delete"))),
		testutil.WithChildren(modal.Content(), testutil.Sequence(
			testutil.WithChildren(modal.Title(), testutil.Text("Are you sure?")),
			testutil.WithChildren(modal.Close(),
				testutil.WithChildren(button.New(), testutil.Text("Cancel"))),
		)),
	))
}

func fullDropdown() templ.Component {
	return testutil.WithChildren(dropdown.Root(), testutil.Sequence(
		testutil.WithChildren(dropdown.Trigger(),
			testutil.WithChildren(button.New(), testutil.Text("Options"))),
		testutil.WithChildren(dropdown.Menu(), testutil.Sequence(
			testutil.WithChildren(dropdown.Item("/profile"), testutil.Text("Profile")),
			dropdown.Divider(),
			testutil.WithChildren(dropdown.ItemButton(), testutil.Text("Log out")),
		)),
	))
}

func init() {
	composites["modal"] = fullModal
	composites["modal-named"] = func() templ.Component {
		return testutil.Sequence(
			testutil.WithChildren(modal.Trigger(modal.For("confirm")),
				testutil.WithChildren(button.New(), testutil.Text("Open"))),
			testutil.WithChildren(modal.Content(modal.Name("confirm")),
				testutil.WithChildren(modal.Title(), testutil.Text("Named"))),
		)
	}
	composites["dropdown"] = fullDropdown
	composites["tooltip"] = func() templ.Component {
		return testutil.WithChildren(tooltip.New(tooltip.Text("Delete forever")),
			testutil.WithChildren(button.New(button.Label("Delete")), testutil.Text("D")))
	}
	composites["accordion"] = func() templ.Component {
		return testutil.WithChildren(accordion.Root(accordion.Exclusive()), testutil.Sequence(
			testutil.WithChildren(accordion.Item(accordion.Title("Shipping")), testutil.Text("3-5 days")),
			testutil.WithChildren(accordion.Item(accordion.Title("Returns"), accordion.Open()), testutil.Text("30 days")),
		))
	}
	composites["navlist"] = func() templ.Component {
		return testutil.WithChildren(navlist.New(navlist.Label("Main")), testutil.Sequence(
			testutil.WithChildren(navlist.Item("/", navlist.Current()), testutil.Text("Home")),
			testutil.WithChildren(navlist.Group(navlist.Title("Settings")), testutil.Sequence(
				testutil.WithChildren(navlist.Item("/profile"), testutil.Text("Profile")),
			)),
		))
	}
	composites["header"] = func() templ.Component {
		return testutil.WithChildren(header.New(header.Sticky()), testutil.Sequence(
			testutil.WithChildren(navbar.New(navbar.Label("Main")), testutil.Sequence(
				testutil.WithChildren(navbar.Item("/", navbar.Current()), testutil.Text("Home")),
				testutil.WithChildren(navbar.Item("/inbox", navbar.Badge("12")), testutil.Text("Inbox")),
			)),
			header.Spacer(),
			testutil.WithChildren(button.New(button.Ghost), testutil.Text("Sign out")),
		))
	}
	composites["sidebar"] = func() templ.Component {
		return testutil.Sequence(
			sidebar.Toggle(),
			testutil.WithChildren(sidebar.New(),
				testutil.WithChildren(navlist.New(navlist.Label("Main")),
					testutil.WithChildren(navlist.Item("/"), testutil.Text("Home")))),
		)
	}
	composites["tabs"] = func() templ.Component {
		return testutil.WithChildren(tabs.New(tabs.Label("Sections")), testutil.Sequence(
			testutil.WithChildren(tabs.Section(tabs.Title("Profile")), testutil.Text("Profile panel")),
			testutil.WithChildren(tabs.Section(tabs.Title("Billing"), tabs.Open()), testutil.Text("Billing panel")),
			testutil.WithChildren(tabs.Section(tabs.Title("Team")), testutil.Text("Team panel")),
		))
	}
	composites["table"] = func() templ.Component {
		return testutil.WithChildren(table.New(), testutil.Sequence(
			testutil.WithChildren(table.Header(), testutil.Sequence(
				testutil.WithChildren(table.Column(), testutil.Text("Name")),
				testutil.WithChildren(table.Column(), testutil.Text("Status")),
			)),
			testutil.WithChildren(table.Body(),
				testutil.WithChildren(table.Row(), testutil.Sequence(
					testutil.WithChildren(table.Cell(), testutil.Text("Ada")),
					testutil.WithChildren(table.Cell(),
						testutil.WithChildren(badge.New(badge.Green), testutil.Text("Active"))),
				))),
		))
	}
}

func TestModalWiring(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, fullModal()))

	dialog := tree.One("modal")
	if dom.GetAttr(dialog, "closedby") != "any" {
		t.Fatal("dialog missing closedby=any")
	}

	buttons := dom.FindAll(tree.Root, dom.ByAttr("commandfor"))
	if len(buttons) != 2 {
		t.Fatalf("expected trigger + close wired, got %d", len(buttons))
	}
	for _, b := range buttons {
		if got := dom.GetAttr(b, "commandfor"); got != dom.GetAttr(dialog, "id") {
			t.Fatalf("commandfor=%q, dialog id=%q", got, dom.GetAttr(dialog, "id"))
		}
	}
	if dom.GetAttr(buttons[0], "command") != "show-modal" || dom.GetAttr(buttons[1], "command") != "close" {
		t.Fatalf("unexpected commands: %q, %q", dom.GetAttr(buttons[0], "command"), dom.GetAttr(buttons[1], "command"))
	}

	title := tree.One("modal-title")
	if dom.GetAttr(dialog, "aria-labelledby") != dom.GetAttr(title, "id") {
		t.Fatal("dialog not labelled by its title")
	}
}

func TestModalFailsLoudlyWithoutTarget(t *testing.T) {
	err := testutil.RenderErr(testutil.WithChildren(modal.Trigger(),
		testutil.WithChildren(button.New(), testutil.Text("Open"))))
	if !errors.Is(err, modal.ErrNoTarget) {
		t.Fatalf("expected ErrNoTarget, got %v", err)
	}

	err = testutil.RenderErr(testutil.WithChildren(modal.Root(),
		testutil.WithChildren(modal.Trigger(), testutil.Text("no button here"))))
	if !errors.Is(err, modal.ErrNoButton) {
		t.Fatalf("expected ErrNoButton, got %v", err)
	}
}

func TestModalTriggerIgnoresNestedComponentButtons(t *testing.T) {
	// A button inside another component (a badge here, standing in for any
	// composite) must not be hijacked; the trigger's own button must be.
	err := testutil.RenderErr(testutil.WithChildren(modal.Root(),
		testutil.WithChildren(modal.Trigger(),
			testutil.WithChildren(badge.New(), // opaque component root
				testutil.WithChildren(button.New(), testutil.Text("nested"))))))
	if !errors.Is(err, modal.ErrNoButton) {
		t.Fatalf("trigger should not reach into other components; got %v", err)
	}
}

func TestDropdownWiring(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, fullDropdown()))

	menu := tree.One("dropdown-menu")
	if !dom.HasAttr(menu, "popover") {
		t.Fatal("menu is not a popover")
	}

	btn := dom.Find(tree.Root, dom.ByAttr("commandfor"))
	if btn == nil {
		t.Fatal("trigger button not wired")
	}
	if dom.GetAttr(btn, "command") != "toggle-popover" {
		t.Fatalf("command=%q", dom.GetAttr(btn, "command"))
	}
	if dom.GetAttr(btn, "commandfor") != dom.GetAttr(menu, "id") {
		t.Fatal("commandfor does not match menu id")
	}

	// Anchor pair: trigger declares the name the menu positions against.
	anchor := dom.GetAttr(btn, "style")
	positionAnchor := dom.GetAttr(menu, "style")
	if !strings.Contains(anchor, "anchor-name: --") {
		t.Fatalf("trigger style=%q", anchor)
	}
	name := strings.TrimSpace(strings.TrimPrefix(anchor, "anchor-name:"))
	if !strings.Contains(positionAnchor, "position-anchor: "+name) {
		t.Fatalf("menu style=%q does not reference %q", positionAnchor, name)
	}
}

func TestDropdownOutsideRootFails(t *testing.T) {
	err := testutil.RenderErr(testutil.WithChildren(dropdown.Menu(), testutil.Text("x")))
	if !errors.Is(err, dropdown.ErrNoScope) {
		t.Fatalf("expected ErrNoScope, got %v", err)
	}
}

func TestTooltipDescribesItsTarget(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, composites["tooltip"]()))
	tip := tree.One("tooltip-content")
	btn := tree.One("button")

	if dom.GetAttr(tip, "role") != "tooltip" {
		t.Fatal("missing role=tooltip")
	}
	if dom.GetAttr(btn, "aria-describedby") != dom.GetAttr(tip, "id") {
		t.Fatal("target not described by tooltip")
	}
	if dom.GetAttr(btn, "interestfor") != dom.GetAttr(tip, "id") {
		t.Fatal("button missing interestfor enhancement")
	}
}

func TestAccordionExclusiveSharesName(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, composites["accordion"]()))
	items := dom.FindAll(tree.Root, dom.ByMarker("accordion-item"))
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	name := dom.GetAttr(items[0], "name")
	if name == "" || name != dom.GetAttr(items[1], "name") {
		t.Fatalf("exclusive items must share a details name: %q vs %q",
			name, dom.GetAttr(items[1], "name"))
	}
	if !dom.HasAttr(items[1], "open") {
		t.Fatal("Open() item not open")
	}
}

func TestSidebarPairsWithToggle(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, composites["sidebar"]()))
	aside := tree.One("sidebar")
	toggle := tree.One("sidebar-toggle")

	if !dom.HasAttr(aside, "popover") {
		t.Fatal("sidebar must carry popover")
	}
	if dom.GetAttr(toggle, "commandfor") != dom.GetAttr(aside, "id") {
		t.Fatal("toggle not wired to sidebar")
	}
	if dom.GetAttr(toggle, "command") != "toggle-popover" {
		t.Fatal("toggle missing command")
	}
}

func TestDetailsTabs(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, composites["tabs"]()))

	group := tree.One("tabs-group")
	sections := dom.FindAll(tree.Root, dom.ByMarker("tabs-section"))
	if len(sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(sections))
	}

	// Exclusive switching: all sections share the group's details name.
	name := dom.GetAttr(sections[0], "name")
	for _, s := range sections {
		if dom.GetAttr(s, "name") != name {
			t.Fatalf("sections must share a details name")
		}
	}

	// The explicitly opened section is the only open one.
	var open []int
	for i, s := range sections {
		if dom.HasAttr(s, "open") {
			open = append(open, i)
		}
	}
	if len(open) != 1 || open[0] != 1 {
		t.Fatalf("expected only section 1 open, got %v", open)
	}

	// One explicit grid column per handle, so panels can span 1/-1.
	if got := dom.GetAttr(group, "style"); !strings.Contains(got, "repeat(3, max-content)") {
		t.Fatalf("group style %q missing column template", got)
	}
}

func TestDetailsTabsDefaultOpen(t *testing.T) {
	c := testutil.WithChildren(tabs.New(), testutil.Sequence(
		testutil.WithChildren(tabs.Section(tabs.Title("A")), testutil.Text("a")),
		testutil.WithChildren(tabs.Section(tabs.Title("B")), testutil.Text("b")),
	))
	tree := testutil.NewTree(t, testutil.Render(t, c))
	sections := dom.FindAll(tree.Root, dom.ByMarker("tabs-section"))
	if !dom.HasAttr(sections[0], "open") {
		t.Fatal("first section must open by default")
	}
	if dom.HasAttr(sections[1], "open") {
		t.Fatal("only the first section should open by default")
	}
}

func TestSectionOutsideGroupFails(t *testing.T) {
	err := testutil.RenderErr(testutil.WithChildren(tabs.Section(tabs.Title("X")), testutil.Text("x")))
	if !errors.Is(err, tabs.ErrNoGroup) {
		t.Fatalf("expected ErrNoGroup, got %v", err)
	}
}

func TestTableStructureSurvivesParsing(t *testing.T) {
	tree := testutil.NewTree(t, testutil.Render(t, composites["table"]()))
	if len(dom.FindAll(tree.Root, dom.ByTag(atom.Th))) != 2 {
		t.Fatal("columns lost")
	}
	if len(dom.FindAll(tree.Root, dom.ByTag(atom.Td))) != 2 {
		t.Fatal("cells lost")
	}
	if dom.Find(tree.Root, dom.ByMarker("badge")) == nil {
		t.Fatal("component inside cell lost")
	}
}

func TestPhase3Goldens(t *testing.T) {
	testutil.Golden(t, "modal", fullModal())
	testutil.Golden(t, "dropdown", fullDropdown())
	testutil.Golden(t, "navlist", composites["navlist"]())
	testutil.Golden(t, "accordion", composites["accordion"]())
	testutil.Golden(t, "table", composites["table"]())
	testutil.Golden(t, "tabs", composites["tabs"]())
}
