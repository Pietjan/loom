// Package chart renders data as SVG on the server — line, area, and bar
// charts plus sparklines, zero JavaScript:
//
//	@chart.New(
//		chart.Title("Visitors per month"),
//		chart.Labels("Jan", "Feb", "Mar", "Apr", "May", "Jun"),
//		chart.Series("Visitors", []float64{120, 190, 170, 220, 300, 260}),
//		chart.Series("Signups", []float64{40, 60, 55, 90, 120, 100}, chart.Colored(chart.Emerald)),
//		chart.Area(), chart.Smooth(), chart.Legend(),
//	)
//
// Hovering a data point or bar reveals its value in a regular loom
// tooltip — still no JS. HTML can't live inside SVG, but the SVG scales
// proportionally, so viewBox coordinates map 1:1 to percentage positions
// in an HTML overlay; each point gets a small anchored hit area wrapped
// by the tooltip component. Sparklines are the same chart with chrome
// stripped:
//
//	@chart.New(chart.Sparkline(), chart.Series("Trend", data))
//
// Honest limitation, by design: a synced crosshair cursor with a live
// multi-value tooltip (à la Flux's chart.cursor) requires JavaScript, and
// loom ships none — per-point tooltips are the offering.
package chart

import (
	"context"
	"errors"
	"fmt"

	"github.com/a-h/templ"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/opts"
	"github.com/pietjan/loom/internal/render"
	"github.com/pietjan/loom/tooltip"
)

// ErrNoSeries is returned when the chart has no data.
var ErrNoSeries = errors.New("chart: at least one chart.Series(...) is required")

type series struct {
	name   string
	values []float64
	color  Color
}

// SeriesOption configures one series.
type SeriesOption = func(*series)

// Colored sets the series palette color.
func Colored(c Color) SeriesOption { return func(s *series) { s.color = c } }

// Config holds chart options.
type Config struct {
	opts.Common
	TitleText string
	W, H      float64
	labels    []string
	series    []series
	area      bool
	smooth    bool
	bars      bool
	sparkline bool
	legend    bool
	ticks     int
	format    func(float64) string
}

// Option configures a chart.
type Option = func(*Config)

var (
	Class = opts.Class[*Config]
	ID    = opts.ID[*Config]
	Attr  = opts.Attr[*Config]
)

// Title labels the chart for assistive tech (aria-label + <title>).
func Title(title string) Option { return func(c *Config) { c.TitleText = title } }

// Size sets the SVG viewBox (drawing coordinates; the rendered chart
// scales to its container). Default 600×240, sparklines 160×40.
func Size(w, h int) Option {
	return func(c *Config) { c.W, c.H = float64(w), float64(h) }
}

// Labels set the categorical x-axis labels. When given, every series must
// have exactly one value per label.
func Labels(labels ...string) Option { return func(c *Config) { c.labels = labels } }

// Series adds a data series. The first series defaults to the accent
// color; later ones walk the palette.
func Series(name string, values []float64, options ...SeriesOption) Option {
	return func(c *Config) {
		s := series{name: name, values: values}
		for _, opt := range options {
			opt(&s)
		}
		if s.color == "" {
			s.color = defaultOrder[len(c.series)%len(defaultOrder)]
		}
		c.series = append(c.series, s)
	}
}

// Area fills the region under each line.
func Area() Option { return func(c *Config) { c.area = true } }

// Smooth draws curved lines (Catmull-Rom) instead of straight segments.
func Smooth() Option { return func(c *Config) { c.smooth = true } }

// Bars renders grouped bars instead of lines.
func Bars() Option { return func(c *Config) { c.bars = true } }

// Sparkline strips axes, grid, labels, and points for inline use.
func Sparkline() Option { return func(c *Config) { c.sparkline = true } }

// Legend renders series names with color swatches under the chart.
func Legend() Option { return func(c *Config) { c.legend = true } }

// Ticks sets the approximate y-axis tick count (default 4).
func Ticks(n int) Option { return func(c *Config) { c.ticks = n } }

// Format overrides value formatting for ticks and point labels.
func Format(fn func(float64) string) Option { return func(c *Config) { c.format = fn } }

// New renders a chart as a templ component.
func New(options ...Option) templ.Component {
	return render.Component(func(ctx context.Context) (*html.Node, error) {
		return Node(ctx, options...)
	})
}

// Node builds the chart node: a <figure> holding the SVG (with its
// tooltip overlay) and an optional legend.
func Node(ctx context.Context, options ...Option) (*html.Node, error) {
	cfg := Config{ticks: 4, format: formatValue}
	for _, opt := range options {
		opt(&cfg)
	}
	if len(cfg.series) == 0 {
		return nil, ErrNoSeries
	}
	n := len(cfg.series[0].values)
	if n == 0 {
		return nil, fmt.Errorf("chart: series %q has no values", cfg.series[0].name)
	}
	for _, s := range cfg.series {
		if len(s.values) != n {
			return nil, fmt.Errorf("chart: series %q has %d values, want %d (all series and labels must align)", s.name, len(s.values), n)
		}
	}
	if cfg.labels != nil && len(cfg.labels) != n {
		return nil, fmt.Errorf("chart: %d labels for %d values per series", len(cfg.labels), n)
	}
	if cfg.W == 0 {
		cfg.W, cfg.H = 600, 240
		if cfg.sparkline {
			cfg.W, cfg.H = 160, 40
		}
	}

	svg := draw(cfg, n)

	// The overlay must cover exactly the SVG's box (not the legend), so
	// they share a positioning wrapper.
	plot := dom.El(atom.Div, dom.Attr("class", "relative"))
	plot.AppendChild(svg)
	if !cfg.sparkline {
		if err := appendTooltips(ctx, plot, cfg, n); err != nil {
			return nil, err
		}
	}

	fig := dom.El(atom.Figure, dom.Marker("chart"))
	fig.AppendChild(plot)
	if cfg.legend {
		fig.AppendChild(legend(cfg))
	}
	cfg.Apply(fig, rootClasses())
	return fig, nil
}

// appendTooltips adds the HTML hover layer: one loom tooltip per datum,
// anchored over the datum's viewBox geometry expressed as percentages of
// the plot box. Line/area points get a small centered hit spot; bars are
// hoverable across their whole rect, and the bubble still rises from the
// bar's top-center because the tooltip's bubble anchors to its wrapper's
// top edge.
func appendTooltips(ctx context.Context, plot *html.Node, cfg Config, n int) error {
	g := plotGeometry(cfg)
	lo, hi := domain(cfg.series)
	ticks := niceTicks(lo, hi, cfg.ticks)
	g.lo, g.hi = ticks[0], ticks[len(ticks)-1]

	band := (g.right - g.left) / float64(n)
	group := band * 0.7
	bar := group / float64(len(cfg.series))
	baseline := g.y(max(g.lo, 0))

	// The hoverable surface fills whatever box its tooltip wrapper defines.
	fill := render.Component(func(context.Context) (*html.Node, error) {
		return dom.El(atom.Span, dom.Attr("class", "block size-full")), nil
	})

	for si, s := range cfg.series {
		for i, v := range s.values {
			text := cfg.format(v)
			if s.name != "" {
				text = s.name + ": " + text
			}

			var class, style string
			if cfg.bars {
				// Cover the whole bar rect — matches drawBars geometry.
				x := g.left + band*float64(i) + (band-group)/2 + bar*float64(si)
				top, height := g.y(v), baseline-g.y(v)
				if height < 0 {
					top, height = baseline, -height
				}
				class = "absolute"
				style = fmt.Sprintf("left: %s%%; top: %s%%; width: %s%%; height: %s%%",
					fmtCoord(x/cfg.W*100), fmtCoord(top/cfg.H*100),
					fmtCoord((bar-2)/cfg.W*100), fmtCoord(height/cfg.H*100))
			} else {
				// A centered hit spot over the point.
				x, y := g.x(i, n), g.y(v)
				class = "absolute size-5 -translate-x-1/2 -translate-y-1/2"
				style = fmt.Sprintf("left: %s%%; top: %s%%", fmtCoord(x/cfg.W*100), fmtCoord(y/cfg.H*100))
			}

			tip, err := tooltip.Node(templ.WithChildren(ctx, fill),
				tooltip.Text(text),
				tooltip.Class(class),
				tooltip.Attr("style", style))
			if err != nil {
				return err
			}
			plot.AppendChild(tip)
		}
	}
	return nil
}

// geometry maps data space to SVG user units.
type geometry struct {
	left, top, right, bottom float64 // inner plot edges
	lo, hi                   float64 // y domain
}

func (g geometry) x(i, n int) float64 {
	if n == 1 {
		return (g.left + g.right) / 2
	}
	return g.left + (g.right-g.left)*float64(i)/float64(n-1)
}

func (g geometry) y(v float64) float64 {
	return g.bottom - (g.bottom-g.top)*(v-g.lo)/(g.hi-g.lo)
}

func plotGeometry(cfg Config) geometry {
	if cfg.sparkline {
		return geometry{left: 2, top: 2, right: cfg.W - 2, bottom: cfg.H - 2}
	}
	return geometry{left: 36, top: 10, right: cfg.W - 8, bottom: cfg.H - 22}
}

func draw(cfg Config, n int) *html.Node {
	svg := dom.El(atom.Svg,
		dom.Marker("chart-svg"),
		dom.Attr("viewBox", fmt.Sprintf("0 0 %s %s", fmtCoord(cfg.W), fmtCoord(cfg.H))),
		dom.Attr("role", "img"),
		dom.Attr("class", svgClasses()))

	label := cfg.TitleText
	if label == "" {
		for i, s := range cfg.series {
			if i > 0 {
				label += ", "
			}
			label += s.name
		}
	}
	// aria-label on role="img" is the accessible name screen readers
	// announce. A <title> would name it too but also renders as a native
	// browser tooltip on hover — noise over our own point/bar tooltips —
	// so we deliberately omit it.
	dom.SetAttr(svg, "aria-label", label)

	g := plotGeometry(cfg)

	lo, hi := domain(cfg.series)
	ticks := niceTicks(lo, hi, cfg.ticks)
	g.lo, g.hi = ticks[0], ticks[len(ticks)-1]

	if !cfg.sparkline {
		drawGrid(svg, cfg, g, ticks)
		drawXLabels(svg, cfg, g, n)
	}
	if cfg.bars {
		drawBars(svg, cfg, g, n)
	} else {
		drawLines(svg, cfg, g, n)
		if !cfg.sparkline {
			drawPoints(svg, cfg, g, n)
		}
	}
	return svg
}

func drawGrid(svg *html.Node, cfg Config, g geometry, ticks []float64) {
	grid := dom.CustomEl("g", dom.Marker("chart-grid"), dom.Attr("class", gridClasses()))
	for _, t := range ticks {
		y := fmtCoord(g.y(t))
		line := dom.CustomEl("line",
			dom.Attr("x1", fmtCoord(g.left)), dom.Attr("y1", y),
			dom.Attr("x2", fmtCoord(g.right)), dom.Attr("y2", y))
		grid.AppendChild(line)
	}
	svg.AppendChild(grid)

	for _, t := range ticks {
		label := dom.CustomEl("text",
			dom.Attr("x", fmtCoord(g.left-6)),
			dom.Attr("y", fmtCoord(g.y(t)+3)),
			dom.Attr("text-anchor", "end"),
			dom.Attr("class", tickLabelClasses()))
		label.AppendChild(dom.Text(cfg.format(t)))
		svg.AppendChild(label)
	}
}

func drawXLabels(svg *html.Node, cfg Config, g geometry, n int) {
	for i, text := range cfg.labels {
		x := g.x(i, n)
		if cfg.bars {
			band := (g.right - g.left) / float64(n)
			x = g.left + band*(float64(i)+0.5)
		}
		label := dom.CustomEl("text",
			dom.Attr("x", fmtCoord(x)),
			dom.Attr("y", fmtCoord(g.bottom+16)),
			dom.Attr("text-anchor", "middle"),
			dom.Attr("class", tickLabelClasses()))
		label.AppendChild(dom.Text(text))
		svg.AppendChild(label)
	}
}

func drawLines(svg *html.Node, cfg Config, g geometry, n int) {
	for _, s := range cfg.series {
		style := palette[s.color]
		pts := make([]xy, n)
		for i, v := range s.values {
			pts[i] = xy{g.x(i, n), g.y(v)}
		}

		path := linePath(pts)
		if cfg.smooth {
			path = smoothPath(pts)
		}

		if cfg.area {
			area := dom.CustomEl("path",
				dom.Attr("d", areaPath(path, pts[0], pts[n-1], g.y(g.lo))),
				dom.Attr("class", style.fill))
			svg.AppendChild(area)
		}
		line := dom.CustomEl("path",
			dom.Attr("d", path),
			dom.Attr("stroke-linecap", "round"),
			dom.Attr("stroke-linejoin", "round"),
			dom.Attr("class", lineClasses(style)))
		svg.AppendChild(line)
	}
}

// drawPoints marks each datum with a dot. Purely visual — the hover
// tooltips live in the HTML overlay (appendTooltips).
func drawPoints(svg *html.Node, cfg Config, g geometry, n int) {
	for _, s := range cfg.series {
		style := palette[s.color]
		for i, v := range s.values {
			dot := dom.CustomEl("circle", dom.Marker("chart-point"),
				dom.Attr("cx", fmtCoord(g.x(i, n))), dom.Attr("cy", fmtCoord(g.y(v))),
				dom.Attr("r", "3.5"),
				dom.Attr("stroke-width", "1.5"),
				dom.Attr("class", dotClasses(style)))
			svg.AppendChild(dot)
		}
	}
}

func drawBars(svg *html.Node, cfg Config, g geometry, n int) {
	band := (g.right - g.left) / float64(n)
	group := band * 0.7
	bar := group / float64(len(cfg.series))
	baseline := g.y(max(g.lo, 0))

	for si, s := range cfg.series {
		style := palette[s.color]
		for i, v := range s.values {
			x := g.left + band*float64(i) + (band-group)/2 + bar*float64(si)
			y := g.y(v)
			top, height := y, baseline-y
			if height < 0 { // negative value: bar hangs below the baseline
				top, height = baseline, -height
			}
			rect := dom.CustomEl("rect", dom.Marker("chart-bar"),
				dom.Attr("x", fmtCoord(x)),
				dom.Attr("y", fmtCoord(top)),
				dom.Attr("width", fmtCoord(bar-2)),
				dom.Attr("height", fmtCoord(height)),
				dom.Attr("rx", "2"),
				dom.Attr("class", style.dot))
			svg.AppendChild(rect)
		}
	}
}

func legend(cfg Config) *html.Node {
	cap := dom.El(atom.Figcaption, dom.Marker("chart-legend"),
		dom.Attr("class", legendClasses()))
	for _, s := range cfg.series {
		item := dom.El(atom.Span, dom.Attr("class", "inline-flex items-center gap-1.5"))
		swatch := dom.El(atom.Span,
			dom.Attr("class", legendSwatchClasses(palette[s.color])),
			dom.Attr("aria-hidden", "true"))
		item.AppendChild(swatch)
		item.AppendChild(dom.Text(s.name))
		cap.AppendChild(item)
	}
	return cap
}
