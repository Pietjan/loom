package chart_test

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/pietjan/loom/chart"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

func TestLineChart(t *testing.T) {
	c := chart.New(
		chart.Title("Visitors per month"),
		chart.Labels("Jan", "Feb", "Mar", "Apr"),
		chart.Series("Visitors", []float64{120, 190, 170, 260}),
		chart.Series("Signups", []float64{40, 60, 55, 100}, chart.Emerald),
		chart.Area(), chart.Smooth(), chart.Legend(),
	)
	out := testutil.Render(t, c)

	for _, want := range []string{
		`data-ui="chart"`,
		`data-ui="chart-svg"`,
		`role="img"`,
		`aria-label="Visitors per month"`,
		`data-ui="chart-grid"`,
		`data-ui="chart-point"`,
		`data-ui="chart-legend"`,
		`stroke-accent`,
		`stroke-emerald-500`,
		`fill-accent/15`, // area fill
		` C `,            // smooth: cubic segments
		`>Jan<`, `>Apr<`, // x labels
		`>Visitors<`, `>Signups<`, // legend
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q", want)
		}
	}

	// The accessible name lives on aria-label only - no <title>, which
	// browsers would surface as a native hover tooltip over our own.
	if strings.Contains(out, "<title") {
		t.Errorf("chart should not emit <title> (native hover tooltip); accessible name is on aria-label")
	}

	tree := testutil.NewTree(t, out)
	points := dom.FindAll(tree.Root, dom.ByMarker("chart-point"))
	if len(points) != 8 {
		t.Fatalf("expected 8 dots (2 series × 4), got %d", len(points))
	}

	// Hover values reuse the ordinary tooltip component in an HTML
	// overlay - one per datum, positioned in percentages.
	tips := dom.FindAll(tree.Root, dom.ByMarker("tooltip-content"))
	if len(tips) != 8 {
		t.Fatalf("expected 8 tooltips, got %d", len(tips))
	}
	if tips[0].FirstChild == nil || tips[0].FirstChild.Data != "Visitors: 120" {
		t.Fatalf("tooltip content: %v", tips[0].FirstChild)
	}
	anchors := dom.FindAll(tree.Root, dom.ByMarker("tooltip"))
	if got := dom.GetAttr(anchors[0], "style"); !strings.Contains(got, "left: ") || !strings.Contains(got, "%") {
		t.Fatalf("tooltip anchor not percentage-positioned: %q", got)
	}
}

func TestBarChart(t *testing.T) {
	c := chart.New(
		chart.Bars(),
		chart.Labels("Q1", "Q2"),
		chart.Series("Revenue", []float64{100, 150}),
	)
	tree := testutil.NewTree(t, testutil.Render(t, c))
	bars := dom.FindAll(tree.Root, dom.ByMarker("chart-bar"))
	if len(bars) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(bars))
	}
	if dom.Find(tree.Root, dom.ByMarker("chart-point")) != nil {
		t.Fatal("bar charts must not render dots")
	}
	// Bars get hover tooltips too.
	if got := len(dom.FindAll(tree.Root, dom.ByMarker("tooltip"))); got != 2 {
		t.Fatalf("expected 2 bar tooltips, got %d", got)
	}
}

// Stacked bars share their label's column: one x and the group's full
// width, each segment sitting on the one before it with the bar gap
// between them. A negative value hangs from zero instead of cutting into
// the positives, and the y-axis reaches the stack totals.
func TestStackedBarChart(t *testing.T) {
	out := testutil.Render(t, chart.New(
		chart.Bars(), chart.Stacked(),
		chart.Labels("A", "B"),
		chart.Series("x", []float64{10, 20}),
		chart.Series("y", []float64{30, -10}),
		chart.Series("z", []float64{20, 5}),
	))
	tree := testutil.NewTree(t, out)
	bars := dom.FindAll(tree.Root, dom.ByMarker("chart-bar"))
	if len(bars) != 6 {
		t.Fatalf("expected 6 segments (3 series × 2 labels), got %d", len(bars))
	}
	// Series-major: x(A), x(B), y(A), y(B), z(A), z(B).
	xA, xB, yA, yB, zA, zB := box(t, bars[0]), box(t, bars[1]), box(t, bars[2]), box(t, bars[3]), box(t, bars[4]), box(t, bars[5])

	for _, b := range []rect{yA, zA} {
		if b.x != xA.x || b.w != xA.w {
			t.Errorf("segment %+v not in A's column %+v", b, xA)
		}
	}
	// 600-wide default: band 278, group 194.6, minus the 2-unit gap.
	if !near(xA.w, 192.6) {
		t.Errorf("stacked bar width = %v, want the full group (192.6)", xA.w)
	}

	// A climbs x, y, z from the baseline, a 2-unit gap at each junction.
	if gap := xA.y - (yA.y + yA.h); !near(gap, 2) {
		t.Errorf("gap between x and y = %v, want 2", gap)
	}
	if gap := yA.y - (zA.y + zA.h); !near(gap, 2) {
		t.Errorf("gap between y and z = %v, want 2", gap)
	}

	// A has nothing negative, so its stack stands right on zero. B's
	// negative y hangs from zero, and the two sides meet there like any
	// other junction: the gap split across the zero line.
	zero := xA.y + xA.h
	if bottom := xB.y + xB.h; !near(bottom, zero-1) {
		t.Errorf("x(B) bottom = %v, want 1 above zero (%v)", bottom, zero)
	}
	if !near(yB.y, zero+1) {
		t.Errorf("negative segment top = %v, want 1 below zero (%v)", yB.y, zero)
	}
	// z(B) stacks on x(B), not on the negative y.
	if gap := xB.y - (zB.y + zB.h); !near(gap, 2) {
		t.Errorf("z(B) does not sit on x(B): gap %v", gap)
	}

	// A's stack totals 60, beyond any single value: the axis runs to 100.
	if !strings.Contains(out, ">100<") {
		t.Errorf("y-axis does not reach the stack total: %s", out)
	}

	// Each segment's hover area is its own rect.
	tips := dom.FindAll(tree.Root, dom.ByMarker("tooltip"))
	if len(tips) != 6 {
		t.Fatalf("expected 6 tooltips, got %d", len(tips))
	}
	var tip rect
	style := dom.GetAttr(tips[4], "style")
	if _, err := fmt.Sscanf(style, "left: %f%%; top: %f%%; width: %f%%; height: %f%%",
		&tip.x, &tip.y, &tip.w, &tip.h); err != nil {
		t.Fatalf("z(A) tooltip style %q: %v", style, err)
	}
	// Percentages of the 600×240 box.
	if !near(tip.x, zA.x/6) || !near(tip.y, zA.y/2.4) || !near(tip.w, zA.w/6) || !near(tip.h, zA.h/2.4) {
		t.Errorf("z(A) tooltip %+v does not cover its segment %+v", tip, zA)
	}
}

// Stacking is a bar layout; a line chart renders as if it were not asked.
func TestStackedIgnoredWithoutBars(t *testing.T) {
	options := []chart.Option{
		chart.Labels("A", "B"),
		chart.Series("x", []float64{10, 20}),
		chart.Series("y", []float64{30, 40}),
	}
	plain := testutil.Render(t, chart.New(options...))
	stacked := testutil.Render(t, chart.New(append(options, chart.Stacked())...))
	if plain != stacked {
		t.Errorf("Stacked() changed a line chart:\n%s\nvs\n%s", plain, stacked)
	}
}

type rect struct{ x, y, w, h float64 }

func box(t *testing.T, n *html.Node) rect {
	t.Helper()
	f := func(name string) float64 {
		v, err := strconv.ParseFloat(dom.GetAttr(n, name), 64)
		if err != nil {
			t.Fatalf("bar %s: %v", name, err)
		}
		return v
	}
	return rect{f("x"), f("y"), f("width"), f("height")}
}

// near compares coordinates that were rounded to two decimals on output.
func near(a, b float64) bool { return math.Abs(a-b) < 0.02 }

func TestSparklineIsBare(t *testing.T) {
	out := testutil.Render(t, chart.New(chart.Sparkline(),
		chart.Series("Trend", []float64{1, 3, 2, 5})))
	for _, forbidden := range []string{"chart-grid", "chart-point", "chart-legend", "tooltip"} {
		if strings.Contains(out, forbidden) {
			t.Errorf("sparkline must not render %s", forbidden)
		}
	}
	if !strings.Contains(out, `viewBox="0 0 160 40"`) {
		t.Errorf("sparkline default size missing: %s", out)
	}
}

// Inset(0) makes a sparkline draw edge to edge - line and fill together,
// which is what a backdrop (stat.Background) needs so it does not float a
// hairline off its container's border. The default inset keeps the stroke
// off the box edge instead.
func TestSparklineInset(t *testing.T) {
	series := chart.Series("Trend", []float64{1, 3, 2, 5})

	bleed := testutil.Render(t, chart.New(chart.Sparkline(), chart.Area(), chart.Inset(0), series))
	if !strings.Contains(bleed, `L 160 40 L 0 40 Z`) {
		t.Errorf("Inset(0) area does not close on the box corners: %s", bleed)
	}
	if !strings.Contains(bleed, `d="M 0 33.33`) || !strings.Contains(bleed, `L 160 6.67"`) {
		t.Errorf("Inset(0) line does not span the box edge to edge: %s", bleed)
	}

	def := testutil.Render(t, chart.New(chart.Sparkline(), chart.Area(), series))
	if !strings.Contains(def, `L 158 38 L 2 38 Z`) {
		t.Errorf("default sparkline lost its 2-unit inset: %s", def)
	}

	// A full chart ignores it: those margins hold the tick labels.
	full := testutil.Render(t, chart.New(chart.Area(), chart.Inset(0), series))
	if !strings.Contains(full, `L 592 218 L 36 218 Z`) {
		t.Errorf("Inset moved a full chart's plot box: %s", full)
	}
}

func TestFailsLoudly(t *testing.T) {
	if err := testutil.RenderErr(chart.New()); !errors.Is(err, chart.ErrNoSeries) {
		t.Fatalf("expected ErrNoSeries, got %v", err)
	}

	err := testutil.RenderErr(chart.New(
		chart.Labels("a", "b", "c"),
		chart.Series("x", []float64{1, 2}),
	))
	if err == nil || !strings.Contains(err.Error(), "labels") {
		t.Fatalf("expected label mismatch error, got %v", err)
	}

	err = testutil.RenderErr(chart.New(
		chart.Series("x", []float64{1, 2}),
		chart.Series("y", []float64{1, 2, 3}),
	))
	if err == nil || !strings.Contains(err.Error(), "align") {
		t.Fatalf("expected series mismatch error, got %v", err)
	}
}

func TestGolden(t *testing.T) {
	testutil.Golden(t, "chart-line", chart.New(
		chart.Title("Traffic"),
		chart.Labels("Mon", "Tue", "Wed"),
		chart.Series("Hits", []float64{10, 40, 25}),
	))
}
