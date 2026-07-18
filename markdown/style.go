package markdown

import (
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/pietjan/loom/internal/styles"
)

// loom deliberately styles no raw tags globally, so every element the
// walker emits carries its full recipe — a per-element "prose". Block
// spacing is mt-* with first:mt-0 rather than a uniform space-y-*.

func classes(Config) string {
	var b styles.Builder
	b.Add("text-sm text-base-800 dark:text-base-100")
	return b.String()
}

// headingClass returns the per-level override passed as heading.Class:
// tw-merge lets it win over heading's base size and weight.
func headingClass(level int) string {
	switch level {
	case 1:
		return "mt-10 text-2xl font-semibold first:mt-0"
	case 2:
		return "mt-10 text-xl font-semibold first:mt-0"
	case 3:
		return "mt-8 text-lg font-semibold first:mt-0"
	case 4:
		return "mt-8 text-base font-semibold first:mt-0"
	case 5:
		return "mt-6 text-sm font-semibold first:mt-0"
	default:
		return "mt-6 text-sm font-medium first:mt-0"
	}
}

// paragraphClass overrides text.Node's muted default tone: markdown body
// copy reads at full contrast.
func paragraphClass() string {
	var b styles.Builder
	b.Add("mt-4 leading-relaxed first:mt-0")
	b.Add("text-base-800 dark:text-base-100")
	return b.String()
}

func listClass(ordered, nested bool) string {
	var b styles.Builder
	if ordered {
		b.Add("list-decimal")
	} else {
		b.Add("list-disc")
	}
	b.Add("ps-6")
	if nested {
		b.Add("mt-1")
	} else {
		b.Add("mt-4 first:mt-0")
	}
	return b.String()
}

func itemClass(task bool) string {
	var b styles.Builder
	b.Add("mt-1 leading-relaxed")
	// Task items drop the bullet and pull back into the list padding,
	// GitHub style.
	b.If(task, "-ms-6 list-none")
	return b.String()
}

func checkboxClass() string {
	var b styles.Builder
	b.Add("me-1.5 size-3.5 rounded accent-accent align-[-2px]")
	return b.String()
}

func blockquoteClass() string {
	var b styles.Builder
	b.Add("mt-4 border-s-4 border-base-200 ps-4 italic first:mt-0")
	b.Add("text-base-500 dark:border-base-600 dark:text-base-300")
	return b.String()
}

func codeBlockClass() string {
	var b styles.Builder
	b.Add("mt-4 overflow-x-auto rounded-lg border p-4 first:mt-0")
	b.Add("font-mono text-xs leading-relaxed")
	b.Add("border-base-200 bg-base-50 dark:border-base-600 dark:bg-base-800")
	return b.String()
}

func inlineCodeClass() string {
	var b styles.Builder
	b.Add("rounded bg-base-800/5 px-1.5 py-0.5 font-mono text-[0.85em] dark:bg-white/10")
	return b.String()
}

func imageClass() string {
	var b styles.Builder
	b.Add("mt-4 max-w-full rounded-lg first:mt-0")
	return b.String()
}

// Table recipes mirror table/style.go so markdown tables match the table
// component.

func tableWrapperClass() string {
	var b styles.Builder
	b.Add("mt-4 w-full overflow-x-auto rounded-lg border first:mt-0")
	b.Add("border-base-200 dark:border-base-600")
	return b.String()
}

func tableClass() string {
	var b styles.Builder
	b.Add("w-full text-sm")
	return b.String()
}

func tableHeadClass() string {
	var b styles.Builder
	b.Add("bg-base-50 dark:bg-base-800")
	return b.String()
}

func tableBodyClass() string {
	var b styles.Builder
	b.Add("divide-y divide-base-200 dark:divide-base-600")
	return b.String()
}

func tableCellClass(header bool, align extast.Alignment) string {
	var b styles.Builder
	if header {
		b.Add("px-4 py-2.5 text-xs font-semibold uppercase tracking-wide")
		b.Add("text-base-500 dark:text-base-400")
		b.Add("border-b border-base-200 dark:border-base-600")
	} else {
		b.Add("px-4 py-3 text-base-800 dark:text-base-100")
	}
	b.If(header && align == extast.AlignNone, "text-start")
	styles.Match(&b, align, map[extast.Alignment]string{
		extast.AlignLeft:   "text-left",
		extast.AlignCenter: "text-center",
		extast.AlignRight:  "text-right",
	})
	return b.String()
}
