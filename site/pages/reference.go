package pages

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pietjan/loom/site/internal/apidoc"
)

// apis holds the parsed component APIs. main installs it at startup so the
// reference sections render from source rather than from hand-written tables.
var apis map[string]*apidoc.API

// SetAPI installs the parsed package reference data.
func SetAPI(parsed map[string]*apidoc.API) { apis = parsed }

// apiFor returns the reference for a component page, or nil when the page has
// no matching package - the reference sections are then skipped.
func apiFor(slug string) *apidoc.API { return apis[slug] }

// goGetSnippet is the one-time module install, shown above every import.
const goGetSnippet = "go get " + apidoc.ImportBase

// maxValues caps how many constants a value table lists. Beyond this a table
// stops being a reference and becomes a wall.
const maxValues = 24

// valueNote replaces a value table on pages that document the same constants
// better elsewhere, returning the note and the section to link to. Only the
// icon set qualifies: 1248 names belong in the browser, not a table.
func valueNote(slug, group string) (note, anchor string, ok bool) {
	if slug == iconSlug && group == "Name" {
		return "1248 generated constants, one per Phosphor icon.", "browse-icons", true
	}
	return "", "", false
}

// cappedValues returns the values a table should list, and how many it left
// out, so a long enum degrades to a sample plus a count.
func cappedValues(decls []apidoc.Decl) ([]apidoc.Decl, int) {
	if len(decls) <= maxValues {
		return decls, 0
	}
	return decls[:maxValues], len(decls) - maxValues
}

func importSnippet(api *apidoc.API) string {
	return fmt.Sprintf("import %q", api.ImportPath)
}

// optionsTitle names an option group: the plain Option type is simply
// "Options", while a narrower one - chart's SeriesOption, diagram's
// EdgeOption - is titled after what it configures.
func optionsTitle(g apidoc.Group) string {
	if g.Name == "Option" {
		return "Options"
	}
	return strings.TrimSuffix(g.Name, "Option") + " options"
}

// iconNames returns every generated icon constant, for the browser on the
// icon page. It reads the same parsed source the reference does, so the grid
// cannot list an icon the package does not export.
func iconNames() []apidoc.Decl {
	api := apiFor("icon")
	if api == nil {
		return nil
	}
	for _, group := range api.Values {
		if group.Name == "Name" {
			return group.Decls
		}
	}
	return nil
}

// iconSearchKey is what the browser's filter matches against: the Go constant
// and the icon's own name, so "ArrowLeft" and "arrow-left" both find it.
func iconSearchKey(d apidoc.Decl) string { return d.Name + " " + d.Value }

// anchorID turns a section title into a stable fragment id, so a reader can
// link straight to a component's reference.
func anchorID(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
