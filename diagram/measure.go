package diagram

import (
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
)

// A small box-model calculator over Tailwind classes.
//
// Every loom component emits complete Tailwind class strings, and Tailwind's
// values are fixed and known, so a node's body can be measured rather than
// guessed: padding, borders, gaps, fixed sizes, font sizes, line heights and
// flex direction all read straight off the markup. What the classes cannot
// give us is glyph advance — that lives in the font file. Monospace is exact
// (IBM Plex Mono advances a constant 0.6em); proportional text uses a
// per-character width table, which is an estimate but a far closer one than a
// single average.
//
// This understands the subset of CSS loom actually emits. Anything it does not
// recognise contributes nothing, so an exotic body may still want Size(w, h).

// Tailwind v4 defaults (no overrides in the generated theme).
const spacingUnit = 4.0 // --spacing: 0.25rem

// textSizes maps text-* to its font-size and paired line-height.
var textSizes = map[string][2]float64{
	"xs":   {12, 16},
	"sm":   {14, 20},
	"base": {16, 24},
	"lg":   {18, 28},
	"xl":   {20, 28},
	"2xl":  {24, 32},
	"3xl":  {30, 36},
}

var leadings = map[string]float64{
	"none": 1, "tight": 1.25, "snug": 1.375, "normal": 1.5, "relaxed": 1.625, "loose": 2,
}

var iconSizes = map[string]float64{"xs": 16, "sm": 20, "base": 24}

// style is the resolved layout-affecting CSS for one element.
type style struct {
	padT, padR, padB, padL float64
	brdT, brdR, brdB, brdL float64
	gap                    float64
	w, h                   float64 // fixed box size; 0 means auto
	row                    bool    // children flow horizontally
	hidden                 bool
	fontSize, lineHeight   float64
	mono                   bool
}

// inherit seeds a child's style with the inheritable text properties.
func (s style) inherit() style {
	return style{fontSize: s.fontSize, lineHeight: s.lineHeight, mono: s.mono}
}

// parseStyle resolves an element's classes against its inherited text context.
func parseStyle(n *html.Node, parent style) style {
	s := parent.inherit()
	var flex, col bool
	for _, c := range strings.Fields(dom.GetAttr(n, "class")) {
		// Variant-prefixed utilities (dark:, hover:, md:) never apply to the
		// base measurement; arbitrary-variant wrappers are stripped so
		// icon's "[:where(&)]:size-5" still reads as size-5.
		if i := strings.LastIndex(c, ":"); i >= 0 {
			if strings.HasPrefix(c, "[") {
				c = c[i+1:]
			} else {
				continue
			}
		}
		switch {
		case c == "hidden", c == "sr-only":
			s.hidden = true
		case c == "flex", c == "inline-flex":
			flex = true
		case c == "flex-col":
			col = true
		case c == "font-mono":
			s.mono = true
		case c == "font-sans":
			s.mono = false
		case c == "border":
			s.brdT, s.brdR, s.brdB, s.brdL = 1, 1, 1, 1
		case strings.HasPrefix(c, "text-"):
			if fs, lh, ok := parseText(c[5:]); ok {
				s.fontSize, s.lineHeight = fs, lh
			}
		case strings.HasPrefix(c, "leading-"):
			if m, ok := leadings[c[8:]]; ok && s.fontSize > 0 {
				s.lineHeight = s.fontSize * m
			}
		default:
			applySpacing(&s, c)
		}
	}
	s.row = flex && !col
	if s.fontSize == 0 {
		s.fontSize, s.lineHeight = fontSize, fontSize*1.35
	}
	if s.lineHeight == 0 {
		s.lineHeight = s.fontSize * 1.35
	}
	// loom icons carry their size as data, which is more reliable than
	// unpicking the arbitrary-variant class they ship.
	if n.DataAtom == atom.Svg {
		if px, ok := iconSizes[dom.GetAttr(n, "data-size")]; ok {
			s.w, s.h = px, px
		} else if s.w == 0 {
			s.w, s.h = 24, 24
		}
	}
	return s
}

// applySpacing handles the numeric utilities: padding, border widths, gap and
// fixed sizes.
func applySpacing(s *style, c string) {
	prefix, val, ok := strings.Cut(c, "-")
	if !ok {
		return
	}
	switch prefix {
	case "p", "px", "py", "pt", "pr", "pb", "pl":
		v, ok := spacing(val)
		if !ok {
			return
		}
		switch prefix {
		case "p":
			s.padT, s.padR, s.padB, s.padL = v, v, v, v
		case "px":
			s.padR, s.padL = v, v
		case "py":
			s.padT, s.padB = v, v
		case "pt":
			s.padT = v
		case "pr":
			s.padR = v
		case "pb":
			s.padB = v
		case "pl":
			s.padL = v
		}
	case "gap":
		if v, ok := spacing(val); ok {
			s.gap = v
		}
	case "border":
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return // border-<color>, not a width
		}
		s.brdT, s.brdR, s.brdB, s.brdL = v, v, v, v
	case "w":
		if v, ok := spacing(val); ok {
			s.w = v
		}
	case "h":
		if v, ok := spacing(val); ok {
			s.h = v
		}
	case "size":
		if v, ok := spacing(val); ok {
			s.w, s.h = v, v
		}
	}
}

// spacing resolves a Tailwind spacing step ("2", "1.5") or an arbitrary pixel
// value ("[10px]") to pixels. Relative units (full, auto, %) report false —
// they size against the parent and so contribute no intrinsic width.
func spacing(v string) (float64, bool) {
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		return parsePx(v[1 : len(v)-1])
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return n * spacingUnit, true
}

func parseText(v string) (fs, lh float64, ok bool) {
	if p, found := textSizes[v]; found {
		return p[0], p[1], true
	}
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		if px, ok := parsePx(v[1 : len(v)-1]); ok {
			return px, px * 1.35, true
		}
	}
	return 0, 0, false
}

func parsePx(v string) (float64, bool) {
	v = strings.TrimSuffix(v, "px")
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// measure returns the intrinsic size of an element's content.
func measure(n *html.Node, parent style) (w, h float64) {
	s := parseStyle(n, parent)
	if s.hidden {
		return 0, 0
	}
	if s.w > 0 && s.h > 0 {
		return s.w, s.h
	}

	cw, ch := measureChildren(n, s)
	if s.w > 0 {
		cw = s.w - (s.padL + s.padR + s.brdL + s.brdR)
	}
	if s.h > 0 {
		ch = s.h - (s.padT + s.padB + s.brdT + s.brdB)
	}
	return cw + s.padL + s.padR + s.brdL + s.brdR,
		ch + s.padT + s.padB + s.brdT + s.brdB
}

// measureChildren lays children out as lines: consecutive inline content
// accumulates horizontally, block children start a new line. A flex row keeps
// everything on one line.
func measureChildren(n *html.Node, s style) (w, h float64) {
	var lineW, lineH float64
	var count int
	flush := func() {
		if lineW > w {
			w = lineW
		}
		h += lineH
		lineW, lineH, count = 0, 0, 0
	}
	add := func(cw, ch float64) {
		if count > 0 {
			lineW += s.gap
		}
		lineW += cw
		if ch > lineH {
			lineH = ch
		}
		count++
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch {
		case c.Type == html.TextNode:
			t := strings.Join(strings.Fields(c.Data), " ")
			if t == "" {
				continue
			}
			add(textWidth(t, s.fontSize, s.mono), s.lineHeight)
		case c.Type != html.ElementNode:
			continue
		case c.DataAtom == atom.Br:
			flush()
		case !s.row && isBlock(c):
			flush()
			cw, ch := measure(c, s)
			add(cw, ch)
			flush()
		default:
			cw, ch := measure(c, s)
			add(cw, ch)
		}
	}
	flush()
	return w, h
}

// textWidth sums glyph advances. Monospace is exact; proportional text uses a
// per-character ratio table.
func textWidth(t string, fontSize float64, mono bool) float64 {
	if mono {
		return float64(len([]rune(t))) * fontSize * 0.6
	}
	var total float64
	for _, r := range t {
		total += fontSize * ratio(r)
	}
	return total
}

func ratio(r rune) float64 {
	switch {
	case strings.ContainsRune(" ", r):
		return 0.28
	case strings.ContainsRune("ijltfr.,'!|:;()[]-", r):
		return 0.31
	case strings.ContainsRune("mwMW@", r):
		return 0.86
	case r >= '0' && r <= '9':
		return 0.56
	case r >= 'A' && r <= 'Z':
		return 0.67
	case r >= 'a' && r <= 'z':
		return 0.53
	default:
		return 0.55
	}
}
