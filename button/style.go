package button

import "github.com/pietjan/loom/internal/styles"

func classes(c Config) string {
	var b styles.Builder

	b.Add("relative inline-flex items-center justify-center gap-2 font-medium whitespace-nowrap")
	// no-underline because an Href button renders as an <a>, which the
	// surrounding prose styles may underline; harmless on a <button>.
	b.Add("no-underline")
	// aria-disabled mirrors the :disabled rules for the link form — a
	// <span>/<a> never matches :disabled.
	b.Add("disabled:opacity-75 disabled:cursor-default disabled:pointer-events-none")
	b.Add("aria-disabled:opacity-75 aria-disabled:cursor-default aria-disabled:pointer-events-none")

	size(&b, c)
	background(&b, c)
	text(&b, c)
	border(&b, c)
	shadow(&b, c)
	group(&b, c)

	return b.String()
}

// size also pins the icon size, overriding whatever variant the caller
// passed. Icon's own size utility is wrapped in :where(), so these win
// without specificity games. The 24px box of an extra-small button has no
// room for a 20px glyph, hence the step down.
func size(b *styles.Builder, c Config) {
	switch c.Size {
	case SizeSmall:
		b.Add("h-8 text-sm rounded-md **:data-[ui=icon]:size-5")
		b.If(c.square, "w-8")
		b.If(!c.square, "px-3")
	case SizeExtraSmall:
		b.Add("h-6 text-xs rounded-md **:data-[ui=icon]:size-4")
		b.If(c.square, "w-6")
		b.If(!c.square, "px-2")
	default:
		b.Add("h-10 text-sm rounded-lg **:data-[ui=icon]:size-5")
		b.If(c.square, "w-10")
		b.If(!c.square, "px-4")
	}
}

func background(b *styles.Builder, c Config) {
	switch c.Variant {
	case VariantPrimary:
		b.Add("bg-accent hover:bg-[color-mix(in_oklab,var(--color-accent),transparent_10%)]")
	case VariantFilled:
		b.Add("bg-base-800/5 hover:bg-base-800/10 dark:bg-white/10 dark:hover:bg-white/20")
	case VariantDanger:
		b.Add("bg-red-500 hover:bg-red-600 dark:bg-red-600 dark:hover:bg-red-500")
	case VariantGhost, VariantSubtle:
		b.Add("bg-transparent hover:bg-base-800/5 dark:hover:bg-white/15")
	default:
		b.Add("bg-white hover:bg-base-50 dark:bg-base-700 dark:hover:bg-base-600/75")
	}
}

func text(b *styles.Builder, c Config) {
	switch c.Variant {
	case VariantPrimary:
		b.Add("text-accent-foreground")
	case VariantFilled, VariantGhost:
		b.Add("text-base-800 dark:text-white")
	case VariantDanger:
		b.Add("text-white")
	case VariantSubtle:
		b.Add("text-base-400 hover:text-base-800 dark:text-base-400 dark:hover:text-white")
	default:
		b.Add("text-base-800 dark:text-base-400 dark:hover:text-white")
	}
}

func border(b *styles.Builder, c Config) {
	switch c.Variant {
	case VariantPrimary:
		b.Add("border border-black/10 dark:border-0")
	case VariantFilled, VariantDanger, VariantGhost, VariantSubtle:
		// borderless
	default:
		b.Add("border border-base-200 border-b-base-300/80 dark:border-base-600")
	}
}

func shadow(b *styles.Builder, c Config) {
	switch c.Variant {
	case VariantPrimary:
		b.Add("shadow-[inset_0px_1px_--theme(--color-white/.2)]")
	case VariantDanger:
		b.Add("shadow-[inset_0px_1px_var(--color-red-500),inset_0px_2px_--theme(--color-white/.15)] dark:shadow-none")
	case VariantGhost, VariantSubtle:
		// no shadow
	default:
		b.If(c.Size == SizeExtraSmall, "shadow-none")
		b.If(c.Size != SizeExtraSmall, "shadow-xs")
	}
}

// group styles apply only when the button sits inside a button.Group,
// expressed as relational selectors keyed on the group marker.
func group(b *styles.Builder, c Config) {
	b.Add("in-data-[ui=button-group]:rounded-none")
	b.Add("in-data-[ui=button-group]:first:rounded-s-lg")
	b.Add("in-data-[ui=button-group]:last:rounded-e-lg")
	switch c.Variant {
	case VariantOutline:
		b.Add("in-data-[ui=button-group]:border-s-0")
		b.Add("in-data-[ui=button-group]:first:border-s")
	case VariantFilled:
		b.Add("in-data-[ui=button-group]:border-e in-data-[ui=button-group]:last:border-e-0")
		b.Add("in-data-[ui=button-group]:border-base-200/80 dark:in-data-[ui=button-group]:border-base-900/50")
	case VariantDanger:
		b.Add("in-data-[ui=button-group]:border-e in-data-[ui=button-group]:last:border-e-0")
		b.Add("in-data-[ui=button-group]:border-red-600 dark:in-data-[ui=button-group]:border-red-900/25")
	case VariantPrimary:
		b.Add("[[data-ui=button-group]_&:not(:first-child)]:border-s-[color-mix(in_srgb,var(--color-accent-foreground),transparent_85%)]")
	}
}

func groupClasses() string {
	var b styles.Builder
	b.Add("inline-flex isolate")
	return b.String()
}
