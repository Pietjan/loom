package diagram

import "github.com/pietjan/loom/internal/styles"

// toneStroke maps a node tone to its complete outline classes. The default
// tone is in the map too, so exactly one stroke utility is emitted per node
// (no conflicting stroke-* classes whose CSS source order would decide the
// winner). Complete literals, so the Tailwind scanner sees them.
var toneStroke = map[Tone]string{
	ToneDefault: "stroke-base-300 dark:stroke-base-600",
	ToneAccent:  "stroke-accent",
	ToneIndigo:  "stroke-indigo-400 dark:stroke-indigo-500",
	ToneEmerald: "stroke-emerald-400 dark:stroke-emerald-500",
	ToneAmber:   "stroke-amber-400 dark:stroke-amber-500",
	ToneRose:    "stroke-rose-400 dark:stroke-rose-500",
}

func rootClasses() string {
	var b styles.Builder
	b.Add("block w-full h-auto")
	return b.String()
}

func nodeShapeClasses(t Tone) string {
	var b styles.Builder
	b.Add("fill-white stroke-[1.5] dark:fill-base-800")
	styles.Match(&b, t, toneStroke)
	return b.String()
}

// contentClasses styles the HTML wrapper inside a node's foreignObject. Bare
// nodes get only centering — their body brings its own chrome and type.
func contentClasses(bare bool) string {
	var b styles.Builder
	b.Add("flex h-full w-full items-center justify-center text-center leading-tight")
	b.If(!bare, "px-2 text-[14px] font-medium text-base-700 dark:text-base-100")
	return b.String()
}

func edgeClasses() string {
	var b styles.Builder
	b.Add("fill-none stroke-base-300 stroke-[1.5] dark:stroke-base-600")
	return b.String()
}

func arrowClasses() string {
	var b styles.Builder
	b.Add("fill-base-300 dark:fill-base-600")
	return b.String()
}

func edgeLabelClasses() string {
	var b styles.Builder
	b.Add("fill-base-500 text-[12px] select-none dark:fill-base-400")
	return b.String()
}
