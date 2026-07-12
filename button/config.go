package button

import "github.com/pietjan/loom/internal/opts"

// Type is the button's type attribute.
type Type string

const (
	TypeButton Type = "button" // default
	TypeSubmit Type = "submit"
	TypeReset  Type = "reset"
)

// Variant is the visual style of the button.
type Variant string

const (
	VariantOutline Variant = "outline" // default
	VariantPrimary Variant = "primary"
	VariantFilled  Variant = "filled"
	VariantDanger  Variant = "danger"
	VariantGhost   Variant = "ghost"
	VariantSubtle  Variant = "subtle"
)

// Size is the button size.
type Size string

const (
	SizeBase       Size = "base" // default
	SizeSmall      Size = "sm"
	SizeExtraSmall Size = "xs"
)

// Config holds button options.
type Config struct {
	opts.Common
	Type     Type
	Variant  Variant
	Size     Size
	Disabled bool
	Label    string // accessible name, required for icon-only buttons

	square bool // set during Node when the only content is an icon
}

// Option configures a button.
type Option = func(*Config)

// Common options, instantiated from the shared generics.
var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// WithType sets the button type.
func WithType(t Type) Option { return func(c *Config) { c.Type = t } }

// WithVariant sets the visual style.
func WithVariant(v Variant) Option { return func(c *Config) { c.Variant = v } }

// WithSize sets the button size.
func WithSize(s Size) Option { return func(c *Config) { c.Size = s } }

// Disabled disables the button.
func Disabled() Option { return func(c *Config) { c.Disabled = true } }

// Label sets the accessible name (aria-label). Required when the button's
// only content is an icon.
func Label(label string) Option { return func(c *Config) { c.Label = label } }

// Pre-baked options.
var (
	Submit  = WithType(TypeSubmit)
	Reset   = WithType(TypeReset)
	Outline = WithVariant(VariantOutline)
	Primary = WithVariant(VariantPrimary)
	Filled  = WithVariant(VariantFilled)
	Danger  = WithVariant(VariantDanger)
	Ghost   = WithVariant(VariantGhost)
	Subtle  = WithVariant(VariantSubtle)
	Small   = WithSize(SizeSmall)
	Tiny    = WithSize(SizeExtraSmall)
)
