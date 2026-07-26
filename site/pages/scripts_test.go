package pages

import (
	"context"
	"strings"
	"testing"
)

// renderPage renders one page the way the static build does.
func renderPage(t *testing.T, slug string) string {
	t.Helper()
	for _, p := range All() {
		if p.Slug != slug {
			continue
		}
		var sb strings.Builder
		if err := p.Body().Render(context.Background(), &sb); err != nil {
			t.Fatal(err)
		}
		return sb.String()
	}
	t.Fatalf("no page %q", slug)
	return ""
}

// Scripts ship in the page, not beside it. A <script src> pointing at
// static/ would work in the dev server and 404 in the built site, since
// static/ no longer holds any JavaScript.
func TestScriptsAreInlined(t *testing.T) {
	for _, slug := range []string{"badge", iconSlug} {
		out := renderPage(t, slug)
		if strings.Contains(out, ".js\"") || strings.Contains(out, ".js'") {
			t.Errorf("%s links a script instead of inlining it", slug)
		}
		if !strings.Contains(out, "<script>") {
			t.Errorf("%s carries no inline script", slug)
		}
	}

	// theme.js has to be inline in <head>, ahead of the first paint.
	out := renderPage(t, "badge")
	head := out[:strings.Index(out, "</head>")]
	if !strings.Contains(head, "root.classList.toggle('dark'") {
		t.Error("theme script is not inline in <head>; a light frame will show on a dark OS")
	}

	// The icon browser's script must follow the grid it queries.
	icons := renderPage(t, iconSlug)
	if grid, script := strings.Index(icons, "data-icon-browser"), strings.Index(icons, "data-icon-search]"); grid > script {
		t.Error("icon browser script runs before the grid it queries is parsed")
	}
}

// The embeds must actually carry the files - an empty string still compiles
// and renders, and only fails in the browser.
func TestEmbeddedScriptsAreNotEmpty(t *testing.T) {
	for name, js := range map[string]string{"theme.js": themeJS, "icon-browser.js": iconBrowserJS} {
		if len(strings.TrimSpace(js)) < 100 {
			t.Errorf("%s embedded as %d bytes", name, len(js))
		}
		// A script tag ends at the first "</script>" in its text, whatever
		// it is nested in, so the source must never contain one.
		if strings.Contains(js, "</script") {
			t.Errorf("%s contains a closing script tag and would truncate the page", name)
		}
	}
}
