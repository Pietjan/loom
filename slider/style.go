package slider

import "github.com/pietjan/loom/internal/styles"

// classes reset the native control; the track and thumb are drawn in
// cmd/css/loom.css ([data-ui=slider], keyed on the ::-webkit/::-moz
// pseudo-elements), because Tailwind can't target those.
func classes() string {
	var b styles.Builder
	b.Add("w-full h-5 cursor-pointer appearance-none bg-transparent")
	b.Add("disabled:opacity-75 disabled:cursor-not-allowed")
	return b.String()
}
