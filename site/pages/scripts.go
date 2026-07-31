package pages

// The site's scripts are authored as real .js files - syntax highlighting, a
// formatter, no escaping games - and embedded rather than served, so a page
// carries its own script instead of paying a request for it. It is the trade
// cmd/css already makes with loom.css: edit a file, ship its contents.
//
// It earns its keep on theme.js, which has to run before the first paint or a
// light frame shows on a dark OS; as an external script that meant a
// render-blocking round trip on every page. icon-browser.js is inlined for
// consistency rather than for speed - it is deferred and on one page.
//
// The cost is that both are re-sent with every page instead of being cached
// once, which is the right way round at ~1.5KB each, and that a CSP would need
// a hash or nonce for them. Past a few KB, or if a script starts appearing on
// most pages, serve it from static/ again.

import (
	"context"
	"io"

	_ "embed"

	"github.com/a-h/templ"
)

//go:embed scripts/theme.js
var themeJS string

//go:embed scripts/icon-browser.js
var iconBrowserJS string

//go:embed scripts/theme-picker.js
var themePickerJS string

//go:embed scripts/launcher.js
var launcherJS string

// inlineScript writes an embedded script into the page verbatim. The content
// is our own source, not user input, so it is written raw - templ would
// otherwise escape it into something that no longer parses as JavaScript.
func inlineScript(js string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "<script>"); err != nil {
			return err
		}
		if _, err := io.WriteString(w, js); err != nil {
			return err
		}
		_, err := io.WriteString(w, "</script>")
		return err
	})
}
