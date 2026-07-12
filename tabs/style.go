package tabs

import "github.com/pietjan/loom/internal/styles"

// groupClasses style the tabs container; its grid layout is structural
// CSS in css/loom.css (needs ::details-content + @supports fallback), the
// column template is a per-instance inline style. The column gap sets the
// spacing between tab handles; rows keep no gap — the panel's own padding
// spaces it.
func groupClasses() string {
	var b styles.Builder
	b.Add("relative gap-x-4")
	return b.String()
}

func sectionClasses() string {
	// display: contents comes from loom.css under @supports, so older
	// browsers keep a plain vertical accordion.
	return ""
}

// sectionPanelClasses carry the full-width line under the tab bar as the
// panel's own top border (the panel spans the whole grid row; a pseudo
// grid item would steal row-1 cells from the auto-placed handles).
func sectionPanelClasses() string {
	var b styles.Builder
	b.Add("pt-4 border-t border-base-200 dark:border-base-600")
	return b.String()
}

// sectionTabClasses style a section's summary as a tab handle. Active
// state is the direct parent's [open] — deliberately not in-open:, which
// would also match any open <details> further up (accordion, navlist).
// relative z-10: the -mb-px overlap must paint over the panel's top
// border, and the later-in-DOM panel wins without it.
func sectionTabClasses() string {
	var b styles.Builder
	b.Add("relative z-10 -mb-px inline-flex items-center gap-2 border-b-2 px-1 py-2.5 text-sm font-medium")
	b.Add("cursor-pointer select-none list-none [&::-webkit-details-marker]:hidden")
	b.Add("border-transparent text-base-500 dark:text-base-400")
	b.Add("hover:border-base-300 hover:text-base-800 dark:hover:text-white")
	b.Add("[details[open]>&]:border-accent [details[open]>&]:text-accent-content")
	return b.String()
}
