package styles

// Canonical Tailwind class ordering, ported from pulseui pkg/twsort with
// two tokenizer fixes: leading whitespace no longer leaks into the next
// token, and whitespace inside [arbitrary values] is preserved instead of
// dropped. Sorting is purely cosmetic (conflicts are resolved by tw-merge
// before sorting) but keeps rendered class attributes — and therefore
// golden files — stable.

import (
	"regexp"
	"sort"
	"strings"
)

var classOrder = []string{
	// Layout
	"box-border", "box-content", "block", "inline-block", "inline", "flex", "inline-flex", "table", "inline-table",
	"table-caption", "table-cell", "table-column", "table-column-group", "table-footer-group", "table-header-group",
	"table-row-group", "table-row", "flow-root", "grid", "inline-grid", "contents", "list-item", "hidden", "float-",
	"clear-", "isolate", "isolation-auto", "object-", "overflow-", "overscroll-", "static", "fixed", "absolute",
	"relative", "sticky", "top-", "right-", "bottom-", "left-", "inset-", "visible", "invisible", "z-",

	// Flexbox & Grid
	"flex-basis-", "flex-direction-", "flex-wrap-", "flex-", "flex-grow", "flex-shrink", "order-", "grid-cols-",
	"grid-col-", "grid-rows-", "grid-row-", "grid-flow-", "gap-", "justify-", "justify-items-", "justify-self-",
	"items-", "align-", "place-content-", "place-items-", "place-self-",

	// Spacing
	"p-", "px-", "py-", "ps-", "pe-", "pt-", "pr-", "pb-", "pl-", "m-", "mx-", "my-", "ms-", "me-", "mt-", "mr-", "mb-", "ml-", "space-",

	// Sizing
	"size-", "w-", "min-w-", "max-w-", "h-", "min-h-", "max-h-",

	// Typography
	"font-", "text-", "italic", "not-italic", "font-weight-", "font-variant-numeric-", "letter-spacing-",
	"line-clamp-", "line-height-", "list-", "text-align-", "text-color-", "text-decoration-",
	"text-decoration-color-", "text-decoration-style-", "text-decoration-thickness-", "text-underline-offset-",
	"text-transform-", "text-overflow-", "text-indent-", "vertical-align-", "whitespace-", "break-",
	"content-",

	// Backgrounds
	"bg-", "bg-opacity-", "bg-origin-", "bg-position-", "bg-repeat-", "bg-size-", "bg-image-", "gradient-to-",
	"from-", "via-", "to-",

	// Borders
	"rounded-", "border", "border-", "border-opacity-", "border-style-", "divide-", "divide-opacity-",
	"divide-style-", "outline-", "outline-offset-", "outline-style-", "ring-", "ring-offset-", "ring-opacity-",

	// Effects
	"shadow-", "opacity-", "mix-blend-", "bg-blend-",

	// Filters
	"filter", "blur-", "brightness-", "contrast-", "drop-shadow-", "grayscale-", "hue-rotate-", "invert-",
	"saturate-", "sepia-", "backdrop-",

	// Tables
	"border-collapse", "border-spacing-", "table-layout-", "caption-side-",

	// Transitions & Animation
	"transition", "duration-", "ease-", "delay-", "animate-",

	// Transforms
	"transform", "scale-", "rotate-", "translate-", "skew-", "transform-origin-",

	// Interactivity
	"accent-", "appearance-", "cursor-", "caret-", "pointer-events-", "resize", "scroll-", "scroll-snap-",
	"touch-", "select-", "will-change-",

	// SVG
	"fill-", "stroke-", "stroke-width-",

	// Screen Readers
	"sr-only", "not-sr-only",
}

var variantOrder = map[string]int{
	"sm": 0, "md": 1, "lg": 2, "xl": 3, "2xl": 4, "dark": 10,
	"motion-safe": 20, "motion-reduce": 21, "portrait": 22, "landscape": 23,
	"first": 30, "last": 31, "odd": 32, "even": 33, "visited": 34, "checked": 35,
	"disabled": 36, "enabled": 37, "hover": 40, "focus": 41, "focus-within": 42,
	"focus-visible": 43, "active": 44,
}

var arbitraryVariantRegex = regexp.MustCompile(`^\[.+?\]`)

type variantProperty struct {
	Order int
	Name  string
}

type classProperty struct {
	Variants     []variantProperty
	UtilityOrder int
}

func getClassProperty(className string) classProperty {
	parts := splitVariants(className)
	variants := make([]variantProperty, 0)

	utilityIndex := len(parts) - 1
	for idx, part := range parts {
		if arbitraryVariantRegex.MatchString(part) {
			variants = append(variants, variantProperty{Order: 99, Name: part})
			continue
		}
		if order, ok := variantOrder[part]; ok {
			variants = append(variants, variantProperty{Order: order, Name: part})
			continue
		}
		utilityIndex = idx
		break
	}
	utility := strings.Join(parts[utilityIndex:], ":")

	sort.Slice(variants, func(i, j int) bool {
		if variants[i].Order != variants[j].Order {
			return variants[i].Order < variants[j].Order
		}
		return variants[i].Name < variants[j].Name
	})

	utilityOrder := len(classOrder)
	for idx, prefix := range classOrder {
		if strings.HasPrefix(utility, prefix) {
			utilityOrder = idx
			break
		}
	}

	return classProperty{Variants: variants, UtilityOrder: utilityOrder}
}

// splitVariants splits on ':' outside brackets, so arbitrary variants like
// [&_svg]:size-4 or supports-[anchor-name:--a]:hidden stay intact.
func splitVariants(className string) []string {
	var parts []string
	var current strings.Builder
	bracketLevel := 0
	for _, r := range className {
		switch r {
		case '[':
			bracketLevel++
			current.WriteRune(r)
		case ']':
			bracketLevel--
			current.WriteRune(r)
		case ':':
			if bracketLevel == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	parts = append(parts, current.String())
	return parts
}

// tokenize splits a class string on whitespace, keeping whitespace inside
// [arbitrary values] as part of the token.
func tokenize(classString string) []string {
	var tokens []string
	var current strings.Builder
	bracketLevel := 0

	for _, r := range classString {
		switch r {
		case '[':
			bracketLevel++
			current.WriteRune(r)
		case ']':
			bracketLevel--
			current.WriteRune(r)
		case ' ', '\t', '\n', '\r':
			if bracketLevel > 0 {
				current.WriteRune(r)
				continue
			}
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// Sort deduplicates and orders a Tailwind class string canonically.
func Sort(class string) string {
	fields := tokenize(class)
	if len(fields) == 0 {
		return ""
	}

	seen := make(map[string]struct{}, len(fields))
	unique := make([]string, 0, len(fields))
	for _, c := range fields {
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			unique = append(unique, c)
		}
	}

	props := make(map[string]classProperty, len(unique))
	for _, c := range unique {
		props[c] = getClassProperty(c)
	}

	sort.Slice(unique, func(i, j int) bool {
		pi, pj := props[unique[i]], props[unique[j]]
		if len(pi.Variants) != len(pj.Variants) {
			return len(pi.Variants) < len(pj.Variants)
		}
		for idx := range pi.Variants {
			if pi.Variants[idx].Order != pj.Variants[idx].Order {
				return pi.Variants[idx].Order < pj.Variants[idx].Order
			}
		}
		if pi.UtilityOrder != pj.UtilityOrder {
			return pi.UtilityOrder < pj.UtilityOrder
		}
		// Total-order tiebreak on the class text. Without this, classes
		// with equal sort keys (e.g. negative -translate-* that match no
		// prefix, or two arbitrary variants) keep their input order — and
		// that input is tailwind-merge's output, which is NOT stable
		// across processes. A lexical tiebreak makes Sort deterministic
		// regardless, keeping golden files reproducible.
		return unique[i] < unique[j]
	})

	return strings.Join(unique, " ")
}
