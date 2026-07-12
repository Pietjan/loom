// Package theme generates the CSS custom properties loom components are
// styled against: a base neutral palette (--color-base-50..950) and an
// accent triple (--color-accent, --color-accent-content,
// --color-accent-foreground) with dark-mode overrides.
//
// Each accent has a curated matching neutral base (blue pairs with slate,
// orange with neutral, ...). cmd/css embeds the output in a full Tailwind
// entry file; use Generate directly if you assemble your own.
package theme

import (
	"bytes"
	"fmt"
	"text/template"
)

const tmpl = `/*
 * Design tokens for loom components: the base neutral palette
 * (--color-base-*) and the accent triple (--color-accent*).
 *
 * @theme is a native Tailwind v4 directive — it emits these as :root
 * custom properties AND registers them as utilities (bg-accent,
 * text-base-500, ...). The dark-mode accent overrides below are plain
 * CSS in @layer theme, not part of @theme.
 */
@theme {
    --color-base-50: var(--color-{{.Base}}-50);
    --color-base-100: var(--color-{{.Base}}-100);
    --color-base-200: var(--color-{{.Base}}-200);
    --color-base-300: var(--color-{{.Base}}-300);
    --color-base-400: var(--color-{{.Base}}-400);
    --color-base-500: var(--color-{{.Base}}-500);
    --color-base-600: var(--color-{{.Base}}-600);
    --color-base-700: var(--color-{{.Base}}-700);
    --color-base-800: var(--color-{{.Base}}-800);
    --color-base-900: var(--color-{{.Base}}-900);
    --color-base-950: var(--color-{{.Base}}-950);

    --color-accent: var({{.Accent.Accent}});
    --color-accent-content: var({{.Accent.AccentContent}});
    --color-accent-foreground: var({{.Accent.AccentForeground}});
}

@layer theme {
    /* Dark accent tracks Tailwind's default dark: variant (media query),
     * and also a manual .dark class for apps using a class toggle. */
    @media (prefers-color-scheme: dark) {
        :root {
            --color-accent: var({{.Accent.Dark.Accent}});
            --color-accent-content: var({{.Accent.Dark.AccentContent}});
            --color-accent-foreground: var({{.Accent.Dark.AccentForeground}});
        }
    }
    .dark {
        --color-accent: var({{.Accent.Dark.Accent}});
        --color-accent-content: var({{.Accent.Dark.AccentContent}});
        --color-accent-foreground: var({{.Accent.Dark.AccentForeground}});
    }
}

`

// Color is a Tailwind palette name.
type Color string

// Variant returns the CSS variable for a shade, e.g. --color-blue-500.
func (c Color) Variant(i int) string {
	return fmt.Sprintf("--color-%s-%d", c, i)
}

const colorWhite string = "--color-white"

// Palette names.
const (
	Slate   Color = "slate"
	Gray    Color = "gray"
	Zinc    Color = "zinc"
	Neutral Color = "neutral"
	Stone   Color = "stone"
	Red     Color = "red"
	Orange  Color = "orange"
	Amber   Color = "amber"
	Yellow  Color = "yellow"
	Lime    Color = "lime"
	Green   Color = "green"
	Emerald Color = "emerald"
	Teal    Color = "teal"
	Cyan    Color = "cyan"
	Sky     Color = "sky"
	Blue    Color = "blue"
	Indigo  Color = "indigo"
	Violet  Color = "violet"
	Purple  Color = "purple"
	Fuchsia Color = "fuchsia"
	Pink    Color = "pink"
	Rose    Color = "rose"
)

// Accent maps an accent color to its light and dark CSS variables.
type Accent struct {
	Accent           string
	AccentContent    string
	AccentForeground string
	Dark             AccentDark
}

// AccentDark holds the dark-mode override variables.
type AccentDark struct {
	Accent           string
	AccentContent    string
	AccentForeground string
}

// Accents is the curated accent table: per color, which shades work as
// accent surface, accent-colored text, and text on the accent surface —
// in light and dark mode.
var Accents = map[Color]Accent{
	Slate:   monochromeAccent(Slate),
	Gray:    monochromeAccent(Gray),
	Zinc:    monochromeAccent(Zinc),
	Neutral: monochromeAccent(Neutral),
	Stone:   monochromeAccent(Stone),
	Red: {
		Accent: Red.Variant(500), AccentContent: Red.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Red.Variant(500), AccentContent: Red.Variant(400), AccentForeground: colorWhite},
	},
	Orange: {
		Accent: Orange.Variant(500), AccentContent: Orange.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Orange.Variant(400), AccentContent: Orange.Variant(400), AccentForeground: Orange.Variant(950)},
	},
	Amber: {
		Accent: Amber.Variant(400), AccentContent: Amber.Variant(600), AccentForeground: Amber.Variant(950),
		Dark: AccentDark{Accent: Amber.Variant(400), AccentContent: Amber.Variant(400), AccentForeground: Amber.Variant(950)},
	},
	Yellow: {
		Accent: Yellow.Variant(400), AccentContent: Yellow.Variant(600), AccentForeground: Yellow.Variant(950),
		Dark: AccentDark{Accent: Yellow.Variant(400), AccentContent: Yellow.Variant(400), AccentForeground: Yellow.Variant(950)},
	},
	Lime: {
		Accent: Lime.Variant(400), AccentContent: Lime.Variant(600), AccentForeground: Lime.Variant(950),
		Dark: AccentDark{Accent: Lime.Variant(400), AccentContent: Lime.Variant(400), AccentForeground: Lime.Variant(950)},
	},
	Green: {
		Accent: Green.Variant(600), AccentContent: Green.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Green.Variant(600), AccentContent: Green.Variant(400), AccentForeground: colorWhite},
	},
	Emerald: {
		Accent: Emerald.Variant(600), AccentContent: Emerald.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Emerald.Variant(600), AccentContent: Emerald.Variant(400), AccentForeground: colorWhite},
	},
	Teal: {
		Accent: Teal.Variant(600), AccentContent: Teal.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Teal.Variant(600), AccentContent: Teal.Variant(400), AccentForeground: colorWhite},
	},
	Cyan: {
		Accent: Cyan.Variant(600), AccentContent: Cyan.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Cyan.Variant(600), AccentContent: Cyan.Variant(400), AccentForeground: colorWhite},
	},
	Sky: {
		Accent: Sky.Variant(600), AccentContent: Sky.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Sky.Variant(600), AccentContent: Sky.Variant(400), AccentForeground: colorWhite},
	},
	Blue: {
		Accent: Blue.Variant(500), AccentContent: Blue.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Blue.Variant(500), AccentContent: Blue.Variant(400), AccentForeground: colorWhite},
	},
	Indigo: {
		Accent: Indigo.Variant(500), AccentContent: Indigo.Variant(500), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Indigo.Variant(500), AccentContent: Indigo.Variant(300), AccentForeground: colorWhite},
	},
	Violet: {
		Accent: Violet.Variant(500), AccentContent: Violet.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Violet.Variant(500), AccentContent: Violet.Variant(400), AccentForeground: colorWhite},
	},
	Purple: {
		Accent: Purple.Variant(500), AccentContent: Purple.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Purple.Variant(500), AccentContent: Purple.Variant(300), AccentForeground: colorWhite},
	},
	Fuchsia: {
		Accent: Fuchsia.Variant(600), AccentContent: Fuchsia.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Fuchsia.Variant(600), AccentContent: Fuchsia.Variant(400), AccentForeground: colorWhite},
	},
	Pink: {
		Accent: Pink.Variant(600), AccentContent: Pink.Variant(600), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Pink.Variant(600), AccentContent: Pink.Variant(400), AccentForeground: colorWhite},
	},
	Rose: {
		Accent: Rose.Variant(500), AccentContent: Rose.Variant(500), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: Rose.Variant(500), AccentContent: Rose.Variant(400), AccentForeground: colorWhite},
	},
}

// monochromeAccent builds the neutral accents (dark surface in light mode,
// white surface in dark mode).
func monochromeAccent(c Color) Accent {
	return Accent{
		Accent: c.Variant(800), AccentContent: c.Variant(800), AccentForeground: colorWhite,
		Dark: AccentDark{Accent: colorWhite, AccentContent: colorWhite, AccentForeground: c.Variant(800)},
	}
}

// pairs curates a matching neutral base per accent.
var pairs = map[Color]Color{
	Slate: Slate, Gray: Gray, Zinc: Zinc, Neutral: Neutral, Stone: Stone,
	Red: Zinc, Orange: Neutral, Amber: Neutral, Yellow: Stone,
	Lime: Zinc, Green: Zinc, Emerald: Zinc, Teal: Gray,
	Cyan: Gray, Sky: Gray, Blue: Slate, Indigo: Slate,
	Violet: Gray, Purple: Gray, Fuchsia: Zinc, Pink: Zinc, Rose: Zinc,
}

// Theme is a resolved accent + base combination.
type Theme struct {
	Accent Accent
	Base   Color
}

// Option configures a theme.
type Option func(*Theme) error

// WithAccent selects the accent color and its curated base palette.
func WithAccent(accent string) Option {
	return func(t *Theme) error {
		base, ok := pairs[Color(accent)]
		if !ok {
			return fmt.Errorf("theme: unknown accent %q", accent)
		}
		t.Accent = Accents[Color(accent)]
		t.Base = base
		return nil
	}
}

// WithBase overrides the neutral base palette (must be one of slate, gray,
// zinc, neutral, stone).
func WithBase(base string) Option {
	return func(t *Theme) error {
		switch c := Color(base); c {
		case Slate, Gray, Zinc, Neutral, Stone:
			t.Base = c
			return nil
		default:
			return fmt.Errorf("theme: base must be a neutral palette, got %q", base)
		}
	}
}

// Generate renders the theme CSS (@theme variables + dark overrides).
func Generate(opts ...Option) ([]byte, error) {
	t := Theme{Accent: Accents[Zinc], Base: Zinc}
	for _, opt := range opts {
		if err := opt(&t); err != nil {
			return nil, err
		}
	}

	parsed, err := template.New("theme").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := parsed.Execute(&buf, t); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
