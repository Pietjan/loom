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
	"maps"
	"slices"
	"text/template"
)

const tmpl = `/*
 * Design tokens for loom components: the base neutral palette
 * (--color-base-*) and the accent triple (--color-accent*).
 *
 * @theme is a native Tailwind v4 directive - it emits these as :root
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
    /* Which theme this is. The variables above are colors and say nothing
     * about where they came from; UI that offers a choice has to be able to
     * show which one is in force, including when nothing has been picked. */
    :root {
        --loom-accent-name: "{{.Name}}";
        --loom-base-name: "{{.Base}}";
    }

    /* Dark accent tracks Tailwind's default dark: variant (media query),
     * and also a manual .dark class for apps using a class toggle. */
    @media (prefers-color-scheme: dark) {
        /* :root:not(.light) - a real 0,2,0 selector, not :where(), so it
         * beats the @theme :root light default; an explicit .light opts out. */
        :root:not(.light) {
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

// allTmpl renders every accent and base as an attribute-scoped override, for
// pages that switch theme at runtime rather than recompiling.
const allTmpl = `/*
 * Every curated accent and neutral base, scoped to a data attribute, so a page
 * can switch theme by setting data-accent / data-base on <html>. Appended to
 * the entry file by cmd/css/themes; cmd/css itself still writes one theme.
 *
 * html[data-...] is a 0,1,1 selector - it outranks both the :root defaults
 * @theme emits and the plain .dark block above, without !important. The dark
 * blocks repeat the .dark / :not(.light) pair the single theme uses.
 */
@layer theme {
{{- range .Accents}}
    html[data-accent="{{.Name}}"] {
{{- range .Light}}
        {{.Name}}: var({{.Value}});
{{- end}}
        --loom-accent-name: "{{.Name}}";
        --loom-base-name: "{{.Base}}";
    }
    @media (prefers-color-scheme: dark) {
        html:not(.light)[data-accent="{{.Name}}"] {
{{- range .Dark}}
            {{.Name}}: var({{.Value}});
{{- end}}
        }
    }
    html.dark[data-accent="{{.Name}}"] {
{{- range .Dark}}
        {{.Name}}: var({{.Value}});
{{- end}}
    }
{{end}}
    /* Bases last: an explicit data-base ties on specificity with the base ramp
     * an accent brings along, so source order is what lets it win. */
{{- range .Bases}}
    html[data-base="{{.Name}}"] {
{{- range .Vars}}
        {{.Name}}: var({{.Value}});
{{- end}}
        --loom-base-name: "{{.Name}}";
    }
{{- end}}
}

/* Swatches: what each theme looks like before it is applied, for a picker that
 * has to draw the choice. Chrome rather than tokens, but the colors belong to
 * the table above - a picker that names its own shades drifts from the theme
 * it is selecting.
 *
 * The five monochrome accents all invert to the same white in dark mode, so a
 * picker that offers them as colors has five identical swatches. Pair this
 * with --loom-base-name and offer one "base" choice per neutral instead,
 * which is the distinction that survives.
 *
 * @layer components, not theme: Tailwind's preflight sets border-color on
 * every element in @layer base, and a later layer wins whatever the
 * specificity - in the theme layer a swatch got its fill and not its edge.
 * Components still sits below utilities, so a picker can override any of this
 * with a plain class. */
@layer components {
{{- range .Accents}}
    .loom-swatch-{{.Name}} { background-color: var({{.Swatch}}); }
{{- end}}
    @media (prefers-color-scheme: dark) {
{{- range .Accents}}
        html:not(.light) .loom-swatch-{{.Name}} { background-color: var({{.SwatchDark}}); }
{{- end}}
    }
{{- range .Accents}}
    html.dark .loom-swatch-{{.Name}} { background-color: var({{.SwatchDark}}); }
{{- end}}
{{- range .Bases}}
    .loom-swatch-base-{{.Name}} {
        background-color: var({{.Swatch}});
        border-color: var({{.SwatchEdge}});
        --loom-swatch-edge-strong: var({{.SwatchEdgeStrong}});
    }
{{- end}}
    /* The emphasised edge inverts: a darker shade defines a swatch on a light
     * page, and disappears into a dark one. Left as a variable rather than a
     * rule so whoever draws the picker decides what emphasis means - hover,
     * selection, focus - without the palette knowing about any of them. */
    @media (prefers-color-scheme: dark) {
{{- range .Bases}}
        html:not(.light) .loom-swatch-base-{{.Name}} { --loom-swatch-edge-strong: var({{.SwatchEdgeStrongDark}}); }
{{- end}}
    }
{{- range .Bases}}
    html.dark .loom-swatch-base-{{.Name}} { --loom-swatch-edge-strong: var({{.SwatchEdgeStrongDark}}); }
{{- end}}
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
	Mauve   Color = "mauve"
	Olive   Color = "olive"
	Mist    Color = "mist"
	Taupe   Color = "taupe"
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
// accent surface, accent-colored text, and text on the accent surface -
// in light and dark mode.
var Accents = map[Color]Accent{
	Slate:   monochromeAccent(Slate),
	Gray:    monochromeAccent(Gray),
	Zinc:    monochromeAccent(Zinc),
	Neutral: monochromeAccent(Neutral),
	Stone:   monochromeAccent(Stone),
	Mauve:   monochromeAccent(Mauve),
	Olive:   monochromeAccent(Olive),
	Mist:    monochromeAccent(Mist),
	Taupe:   monochromeAccent(Taupe),
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

// neutrals are the palettes allowed as a base ramp, in the order they are
// emitted. All nine of Tailwind's - the four tinted ones (mauve, olive, mist,
// taupe) arrived after the original five.
var neutrals = []Color{Slate, Gray, Zinc, Neutral, Stone, Mauve, Olive, Mist, Taupe}

// shades are the steps of the base ramp.
var shades = []int{50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950}

// pairs curates a matching neutral base per accent.
var pairs = map[Color]Color{
	Slate: Slate, Gray: Gray, Zinc: Zinc, Neutral: Neutral, Stone: Stone,
	Mauve: Mauve, Olive: Olive, Mist: Mist, Taupe: Taupe,
	Red: Zinc, Orange: Neutral, Amber: Neutral, Yellow: Stone,
	Lime: Zinc, Green: Zinc, Emerald: Zinc, Teal: Gray,
	Cyan: Gray, Sky: Gray, Blue: Slate, Indigo: Slate,
	Violet: Gray, Purple: Gray, Fuchsia: Zinc, Pink: Zinc, Rose: Zinc,
}

// Theme is a resolved accent + base combination. Name is the accent's palette
// name, kept alongside its shades so the output can say which theme it is.
type Theme struct {
	Name   Color
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
		t.Name = Color(accent)
		t.Accent = Accents[Color(accent)]
		t.Base = base
		return nil
	}
}

// WithBase overrides the neutral base palette (must be one of slate, gray,
// zinc, neutral, stone).
func WithBase(base string) Option {
	return func(t *Theme) error {
		c := Color(base)
		if !slices.Contains(neutrals, c) {
			return fmt.Errorf("theme: base must be a neutral palette, got %q", base)
		}
		t.Base = c
		return nil
	}
}

// Generate renders the theme CSS (@theme variables + dark overrides).
func Generate(opts ...Option) ([]byte, error) {
	t := Theme{Name: Zinc, Accent: Accents[Zinc], Base: Zinc}
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

// decl is one custom property and the palette variable it aliases.
type decl struct{ Name, Value string }

// accentBlock is an accent's rules: the light declarations (its triple plus
// the base ramp it pairs with), the dark overrides for the triple, and the
// surface color on its own, for swatches.
type accentBlock struct {
	Name       Color
	Base       Color
	Light      []decl
	Dark       []decl
	Swatch     string
	SwatchDark string
}

// baseBlock is a neutral base ramp on its own, for overriding the one an
// accent brings along. The swatch carries its own edge as well as its fill:
// a picker drawing nine ramps has to draw each one in its own color, and a
// border taken from --color-base-* would be the ramp currently in force
// rather than the one on offer.
type baseBlock struct {
	Name                 Color
	Vars                 []decl
	Swatch               string
	SwatchEdge           string
	SwatchEdgeStrong     string
	SwatchEdgeStrongDark string
}

// GenerateAll renders every accent and base as an attribute-scoped override
// (html[data-accent], html[data-base]), so a page can switch theme by setting
// attributes rather than compiling a stylesheet per combination. Generate
// still supplies the defaults these override.
func GenerateAll() ([]byte, error) {
	var data struct {
		Accents []accentBlock
		Bases   []baseBlock
	}

	for _, name := range slices.Sorted(maps.Keys(Accents)) {
		base, ok := pairs[name]
		if !ok {
			return nil, fmt.Errorf("theme: accent %q has no paired base", name)
		}
		a := Accents[name]
		data.Accents = append(data.Accents, accentBlock{
			Name: name,
			Base: base,
			Light: append([]decl{
				{"--color-accent", a.Accent},
				{"--color-accent-content", a.AccentContent},
				{"--color-accent-foreground", a.AccentForeground},
			}, baseDecls(base)...),
			Dark: []decl{
				{"--color-accent", a.Dark.Accent},
				{"--color-accent-content", a.Dark.AccentContent},
				{"--color-accent-foreground", a.Dark.AccentForeground},
			},
			Swatch:     a.Accent,
			SwatchDark: a.Dark.Accent,
		})
	}
	for _, base := range neutrals {
		data.Bases = append(data.Bases, baseBlock{
			Name: base, Vars: baseDecls(base),
			Swatch: base.Variant(400), SwatchEdge: base.Variant(500),
			SwatchEdgeStrong: base.Variant(700), SwatchEdgeStrongDark: base.Variant(200),
		})
	}

	parsed, err := template.New("themes").Parse(allTmpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := parsed.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// baseDecls aliases the base ramp to a neutral palette.
func baseDecls(base Color) []decl {
	out := make([]decl, 0, len(shades))
	for _, s := range shades {
		out = append(out, decl{fmt.Sprintf("--color-base-%d", s), base.Variant(s)})
	}
	return out
}
