// Package pages holds the site's hand-written documentation pages and the
// registry that drives routing, static rendering, and sidebar navigation.
package pages

import (
	"context"
	"strings"

	"github.com/a-h/templ"
)

// Page is one documentation page, addressed as /components/<slug>/.
type Page struct {
	Slug        string
	Title       string
	Description string
	Body        func() templ.Component
}

// Category groups pages in the sidebar; order is presentation order.
type Category struct {
	Name  string
	Pages []Page
}

// Categories is the single source of truth: it drives the dev-server routes,
// the static-render loop, and the sidebar navigation. Populated in init
// because the page bodies render the sidebar from Categories, which the
// compiler would reject as an initialization cycle in a var initializer.
var Categories []Category

func init() {
	Categories = []Category{
		{Name: "Foundation", Pages: []Page{
			{Slug: "button", Title: "Button", Description: "Clickable actions in seven variants, three sizes, and groupable rows.", Body: Button},
			{Slug: "badge", Title: "Badge", Description: "Small status labels in color variants, pill and small forms.", Body: Badge},
			{Slug: "icon", Title: "Icon", Description: "Phosphor icons in regular and fill variants at three sizes, with generated name constants.", Body: Icon},
			{Slug: "heading", Title: "Heading", Description: "Semantic headings with visual size decoupled from heading level.", Body: Heading},
			{Slug: "text", Title: "Text", Description: "Body copy with strong and subtle fragments.", Body: Text},
			{Slug: "link", Title: "Link", Description: "Styled anchors: default, ghost, subtle, and external.", Body: Link},
			{Slug: "separator", Title: "Separator", Description: "A horizontal rule that adapts to the theme.", Body: Separator},
			{Slug: "card", Title: "Card", Description: "A bordered content container in regular and small padding.", Body: Card},
			{Slug: "callout", Title: "Callout", Description: "Inline notices with heading, text, and severity variants.", Body: Callout},
			{Slug: "avatar", Title: "Avatar", Description: "Initials avatars with sizes, status dots, and stacked groups.", Body: Avatar},
			{Slug: "kbd", Title: "Kbd", Description: "Keyboard key caps for shortcut hints.", Body: Kbd},
			{Slug: "skeleton", Title: "Skeleton", Description: "Loading placeholders in bar and circle shapes.", Body: Skeleton},
			{Slug: "tooltip", Title: "Tooltip", Description: "CSS-only tooltips on hover and focus.", Body: Tooltip},
		}},
		{Name: "Data display", Pages: []Page{
			{Slug: "table", Title: "Table", Description: "Semantic data tables with styled header and rows.", Body: Table},
			{Slug: "stat", Title: "Stat", Description: "KPI tiles with label, value, and a trend slot.", Body: Stat},
			{Slug: "chart", Title: "Chart", Description: "Server-rendered SVG line, area, bar, and sparkline charts.", Body: Chart},
			{Slug: "diagram", Title: "Diagram", Description: "Server-rendered SVG flowcharts with automatic layered layout — no JavaScript.", Body: Diagram},
			{Slug: "timeline", Title: "Timeline", Description: "Vertical and horizontal event timelines with composable indicators.", Body: Timeline},
			{Slug: "kanban", Title: "Kanban", Description: "Static kanban board markup: columns, cards, headers, footers.", Body: Kanban},
			{Slug: "description", Title: "Description", Description: "Definition lists for label and value pairs.", Body: Description},
			{Slug: "progress", Title: "Progress", Description: "Determinate and indeterminate progress bars.", Body: Progress},
			{Slug: "pagination", Title: "Pagination", Description: "Link-based pagination with gaps, previous and next.", Body: Pagination},
			{Slug: "breadcrumbs", Title: "Breadcrumbs", Description: "Hierarchical navigation trails.", Body: Breadcrumbs},
		}},
		{Name: "Content", Pages: []Page{
			{Slug: "markdown", Title: "Markdown", Description: "Render a markdown string as loom-styled markup, GFM included.", Body: Markdown},
			{Slug: "code", Title: "Code", Description: "Chroma-highlighted code blocks with optional diff rendering.", Body: Code},
			{Slug: "accordion", Title: "Accordion", Description: "Expandable sections on the details element, optionally exclusive.", Body: Accordion},
			{Slug: "carousel", Title: "Carousel", Description: "CSS scroll-snap slides — no JavaScript.", Body: Carousel},
		}},
		{Name: "Forms", Pages: []Page{
			{Slug: "field", Title: "Field", Description: "Label, control, description, and error wiring for form fields.", Body: Field},
			{Slug: "fieldset", Title: "Fieldset", Description: "Grouped fields with a legend; disable them all at once.", Body: Fieldset},
			{Slug: "input", Title: "Input", Description: "Text inputs pre-styled for every type.", Body: Input},
			{Slug: "inputgroup", Title: "Input group", Description: "Inputs with leading and trailing addons.", Body: Inputgroup},
			{Slug: "textarea", Title: "Textarea", Description: "Multi-line text entry.", Body: Textarea},
			{Slug: "checkbox", Title: "Checkbox", Description: "Checkboxes with attached labels.", Body: Checkbox},
			{Slug: "radio", Title: "Radio", Description: "Radio groups sharing name and legend through context.", Body: Radio},
			{Slug: "toggle", Title: "Toggle", Description: "Switch-style boolean input.", Body: Toggle},
			{Slug: "slider", Title: "Slider", Description: "Range input with min, max, and step.", Body: Slider},
			{Slug: "picker", Title: "Picker", Description: "A customizable select element with rich options.", Body: Picker},
			{Slug: "fileupload", Title: "File upload", Description: "Styled file inputs with accept filters.", Body: Fileupload},
		}},
		{Name: "Overlays", Pages: []Page{
			{Slug: "modal", Title: "Modal", Description: "Dialogs on the native dialog element and invoker commands — no JavaScript.", Body: Modal},
			{Slug: "dropdown", Title: "Dropdown", Description: "Popover-powered menus with anchor positioning.", Body: Dropdown},
			{Slug: "popover", Title: "Popover", Description: "Freeform popover panels anchored to a trigger.", Body: Popover},
			{Slug: "flash", Title: "Flash", Description: "Flash messages with CSS-only dismiss and auto-hide.", Body: Flash},
		}},
		{Name: "Layout & navigation", Pages: []Page{
			{Slug: "header", Title: "Header", Description: "A top app bar with spacer and sticky option.", Body: Header},
			{Slug: "sidebar", Title: "Sidebar", Description: "Responsive sidebar: popover on small screens, pinned on wide.", Body: Sidebar},
			{Slug: "navbar", Title: "Navbar", Description: "Horizontal navigation with current item and badges.", Body: Navbar},
			{Slug: "navlist", Title: "Navlist", Description: "Vertical navigation lists with labels and collapsible groups.", Body: Navlist},
			{Slug: "tabs", Title: "Tabs", Description: "Tabs on exclusive details — the platform switches panels.", Body: Tabs},
		}},
	}
}

// All flattens Categories in presentation order.
func All() []Page {
	var out []Page
	for _, c := range Categories {
		out = append(out, c.Pages...)
	}
	return out
}

// Find returns the page registered under slug.
func Find(slug string) (Page, bool) {
	for _, p := range All() {
		if p.Slug == slug {
			return p, true
		}
	}
	return Page{}, false
}

type baseKey struct{}

// WithBase stores the site's base path prefix (e.g. "/loom") in ctx; every
// internal link and asset URL is rendered through href so the same pages work
// at "/" in dev and under a project path on GitHub Pages.
func WithBase(ctx context.Context, base string) context.Context {
	return context.WithValue(ctx, baseKey{}, strings.TrimSuffix(base, "/"))
}

// href prefixes an absolute site path with the base path from ctx.
func href(ctx context.Context, path string) string {
	base, _ := ctx.Value(baseKey{}).(string)
	return base + path
}
