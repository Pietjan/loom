package pages

import "github.com/pietjan/loom/navlist"

// navItemCurrent marks a sidebar item as the active page.
func navItemCurrent(current bool) []navlist.Option {
	if current {
		return []navlist.Option{navlist.Current()}
	}
	return nil
}
