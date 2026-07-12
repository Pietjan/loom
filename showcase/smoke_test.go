package main

// Browser smoke tests: drive the showcase in a real Chromium and assert
// the zero-JS interactions actually work — the modal opens via invoker
// commands, the dropdown popover toggles, <details name> is exclusive,
// the sidebar responds to viewport width.
//
// The browser runs in a Playwright server container; playwright-go
// connects over websocket:
//
//	make browser   # start the container (once)
//	make smoke     # run these tests
//	make browser-stop
//
// Without the container the tests skip. Run `make css` first so
// static/styles.css exists — the assertions depend on the compiled CSS.
//
// Env overrides: LOOM_PW_WS (websocket url, default ws://127.0.0.1:43110/),
// LOOM_PW_HOST (hostname the containerized browser uses to reach this
// process, default host.docker.internal).

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

var (
	pwBrowser playwright.Browser
	baseURL   string
)

func TestMain(m *testing.M) {
	wsURL := envOr("LOOM_PW_WS", "ws://127.0.0.1:43110/")
	browserHost := envOr("LOOM_PW_HOST", "host.docker.internal")

	// No container → skip the whole suite quietly.
	if conn, err := net.DialTimeout("tcp", hostPort(wsURL), 2*time.Second); err != nil {
		fmt.Println("smoke: playwright server not reachable — start it with `make browser` (skipping)")
		os.Exit(0)
	} else {
		conn.Close()
	}

	// The showcase must listen on all interfaces: the browser lives in the
	// container and reaches us through the docker host gateway.
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(err)
	}
	srv := &http.Server{Handler: handler()}
	go srv.Serve(ln)
	baseURL = fmt.Sprintf("http://%s:%d/", browserHost, ln.Addr().(*net.TCPAddr).Port)

	// playwright.Run needs the local driver; install it on first use
	// (browsers stay in the container — SkipInstallBrowsers).
	pw, err := playwright.Run()
	if err != nil {
		if err := playwright.Install(&playwright.RunOptions{SkipInstallBrowsers: true}); err != nil {
			panic(fmt.Sprintf("smoke: installing playwright driver: %v", err))
		}
		if pw, err = playwright.Run(); err != nil {
			panic(err)
		}
	}
	pwBrowser, err = pw.Chromium.Connect(wsURL)
	if err != nil {
		panic(fmt.Sprintf("smoke: connecting to playwright server at %s: %v", wsURL, err))
	}

	code := m.Run()

	pwBrowser.Close()
	pw.Stop()
	srv.Close()
	os.Exit(code)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func hostPort(wsURL string) string {
	// ws://127.0.0.1:43110/ -> 127.0.0.1:43110
	trimmed := wsURL
	for _, prefix := range []string{"ws://", "wss://"} {
		if len(trimmed) > len(prefix) && trimmed[:len(prefix)] == prefix {
			trimmed = trimmed[len(prefix):]
		}
	}
	if i := len(trimmed) - 1; trimmed[i] == '/' {
		trimmed = trimmed[:i]
	}
	return trimmed
}

func newPage(t *testing.T, width, height int) playwright.Page {
	t.Helper()
	page, err := pwBrowser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: width, Height: height},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { page.Close() })
	if _, err := page.Goto(baseURL); err != nil {
		t.Fatal(err)
	}
	return page
}

func waitVisible(t *testing.T, page playwright.Page, selector string) {
	t.Helper()
	err := page.Locator(selector).First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("waiting for %q: %v", selector, err)
	}
}

func waitHidden(t *testing.T, page playwright.Page, selector string) {
	t.Helper()
	err := page.Locator(selector).First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		t.Fatalf("waiting for %q to hide: %v", selector, err)
	}
}

func evalBool(t *testing.T, page playwright.Page, expr string) bool {
	t.Helper()
	v, err := page.Evaluate(expr)
	if err != nil {
		t.Fatalf("evaluate %q: %v", expr, err)
	}
	b, ok := v.(bool)
	if !ok {
		t.Fatalf("evaluate %q: got %T(%v), want bool", expr, v, v)
	}
	return b
}

func evalInt(t *testing.T, page playwright.Page, expr string) int {
	t.Helper()
	v, err := page.Evaluate(expr)
	if err != nil {
		t.Fatalf("evaluate %q: %v", expr, err)
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	t.Fatalf("evaluate %q: got %T, want number", expr, v)
	return 0
}

func TestModalOpensAndCloses(t *testing.T) {
	page := newPage(t, 1280, 900)

	if err := page.Locator(`button[command="show-modal"]`).Click(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `dialog[data-ui="modal"][open]`)
	if !evalBool(t, page, `document.querySelector('dialog[data-ui="modal"]').matches(':modal')`) {
		t.Fatal("dialog did not open as a modal (focus trap + backdrop)")
	}

	if err := page.Locator(`button[command="close"]`).Click(); err != nil {
		t.Fatal(err)
	}
	waitHidden(t, page, `dialog[data-ui="modal"][open]`)
}

func TestModalEscCloses(t *testing.T) {
	page := newPage(t, 1280, 900)
	if err := page.Locator(`button[command="show-modal"]`).Click(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `dialog[data-ui="modal"][open]`)
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatal(err)
	}
	waitHidden(t, page, `dialog[data-ui="modal"][open]`)
}

func TestDropdownPopoverToggles(t *testing.T) {
	page := newPage(t, 1280, 900)

	trigger := page.Locator(`button[command="toggle-popover"][commandfor*="dropdown"]`)
	if err := trigger.Click(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `[data-ui="dropdown-menu"]:popover-open`)

	// Light dismiss via Esc — platform behavior, no JS of ours.
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatal(err)
	}
	waitHidden(t, page, `[data-ui="dropdown-menu"]:popover-open`)
}

func TestAccordionIsExclusive(t *testing.T) {
	page := newPage(t, 1280, 900)

	// The first item starts open (accordion.Open()); opening the second
	// must close it — the platform's <details name> exclusivity.
	second := page.Locator(`[data-ui="accordion"] [data-ui="accordion-item"] summary`).Nth(1)
	if err := second.Click(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `[data-ui="accordion-item"][open]`)
	if n := evalInt(t, page, `document.querySelectorAll('[data-ui="accordion-item"][open]').length`); n != 1 {
		t.Fatalf("exclusive accordion has %d open items, want 1", n)
	}
}

func TestSidebarRespondsToViewport(t *testing.T) {
	// Narrow: hidden popover until the toggle opens it.
	page := newPage(t, 500, 900)
	waitVisible(t, page, `[data-ui="sidebar-toggle"]`)
	if evalBool(t, page, `document.querySelector('[data-ui="sidebar"]').matches(':popover-open')`) {
		t.Fatal("sidebar must start closed on narrow viewport")
	}
	if err := page.Locator(`[data-ui="sidebar-toggle"]`).Click(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `[data-ui="sidebar"]:popover-open`)

	// Wide: statically visible, toggle gone.
	if err := page.SetViewportSize(1400, 900); err != nil {
		t.Fatal(err)
	}
	if evalBool(t, page, `getComputedStyle(document.querySelector('[data-ui="sidebar"]')).display === 'none'`) {
		t.Fatal("sidebar must be statically visible on wide viewport")
	}
	if !evalBool(t, page, `getComputedStyle(document.querySelector('[data-ui="sidebar-toggle"]')).display === 'none'`) {
		t.Fatal("toggle must hide on wide viewport")
	}
}

func TestPopoverToggles(t *testing.T) {
	page := newPage(t, 1280, 900)

	if visible, _ := page.Locator(`[data-ui="popover-content"]`).First().IsVisible(); visible {
		t.Fatal("popover must start closed")
	}
	if err := page.Locator(`button[commandfor*="popover"]`).First().Click(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `[data-ui="popover-content"]:popover-open`)
	// Esc = light dismiss, from the Popover API.
	if err := page.Keyboard().Press("Escape"); err != nil {
		t.Fatal(err)
	}
	waitHidden(t, page, `[data-ui="popover-content"]:popover-open`)
}

func TestFlashDismiss(t *testing.T) {
	page := newPage(t, 1280, 900)

	flash := page.Locator(`[data-ui="flash"]`).First()
	if visible, _ := flash.IsVisible(); !visible {
		t.Fatal("flash should be visible on load")
	}
	// The close control is a label wrapping a hidden checkbox; clicking it
	// checks the box and has-[:checked]:hidden collapses the flash — no JS.
	if err := flash.Locator("label").Click(); err != nil {
		t.Fatal(err)
	}
	waitHidden(t, page, `[data-ui="flash"]:first-of-type`)
}

func TestTooltipShowsOnHoverAndFocus(t *testing.T) {
	page := newPage(t, 1280, 900)

	tip := page.Locator(`[data-ui="tooltip-content"]`).First()
	if visible, _ := tip.IsVisible(); visible {
		t.Fatal("tooltip must start hidden")
	}
	if err := page.Locator(`[data-ui="tooltip"] button`).First().Hover(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `[data-ui="tooltip-content"]`)
}

func TestDetailsTabsSwitch(t *testing.T) {
	page := newPage(t, 1280, 900)

	// Exactly one section open initially.
	if n := evalInt(t, page, `document.querySelectorAll('[data-ui="tabs-section"][open]').length`); n != 1 {
		t.Fatalf("expected 1 open section, got %d", n)
	}

	// Click the second handle: it opens, the first closes (details name).
	if err := page.Locator(`[data-ui="tabs-section"] summary`).Nth(1).Click(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `[data-ui="tabs-section"]:nth-of-type(2)[open]`)
	if n := evalInt(t, page, `document.querySelectorAll('[data-ui="tabs-section"][open]').length`); n != 1 {
		t.Fatalf("expected exclusive switching, got %d open", n)
	}

	// The open handle is mouse-inert (can't click the visible tab shut).
	if !evalBool(t, page, `getComputedStyle(document.querySelector('[data-ui="tabs-section"][open] > summary')).pointerEvents === 'none'`) {
		t.Fatal("open summary should be mouse-inert")
	}
}

func TestChartTooltipOnHover(t *testing.T) {
	page := newPage(t, 1280, 900)

	// Charts reuse the ordinary tooltip component in an HTML overlay.
	tip := page.Locator(`[data-ui="chart"] [data-ui="tooltip-content"]`).First()
	if visible, _ := tip.IsVisible(); visible {
		t.Fatal("chart tooltip must start hidden")
	}
	if err := page.Locator(`[data-ui="chart"] [data-ui="tooltip"]`).First().Hover(); err != nil {
		t.Fatal(err)
	}
	waitVisible(t, page, `[data-ui="chart"] [data-ui="tooltip-content"]`)

	text, err := tip.TextContent()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, ":") {
		t.Fatalf("tooltip should carry series name and value, got %q", text)
	}
}

func TestPickerDegradesOrEnhances(t *testing.T) {
	page := newPage(t, 1280, 900)

	supported := evalBool(t, page, `CSS.supports('appearance', 'base-select')`)
	options := evalInt(t, page, `document.querySelectorAll('select[data-ui="picker"] option').length`)
	if options < 3 {
		t.Fatalf("picker lost options: %d", options)
	}
	t.Logf("base-select supported: %v, options intact: %d", supported, options)
}

func TestNoConsoleErrors(t *testing.T) {
	page, err := pwBrowser.NewPage(playwright.BrowserNewPageOptions{
		Viewport: &playwright.Size{Width: 1280, Height: 900},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { page.Close() })

	var errors []string
	page.OnConsole(func(msg playwright.ConsoleMessage) {
		if msg.Type() == "error" {
			errors = append(errors, msg.Text())
		}
	})
	page.OnPageError(func(e error) {
		errors = append(errors, e.Error())
	})
	if _, err := page.Goto(baseURL); err != nil {
		t.Fatal(err)
	}
	if len(errors) > 0 {
		t.Fatalf("console errors: %v", errors)
	}
}
