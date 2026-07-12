package chart_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/pietjan/loom/chart"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

func TestLineChart(t *testing.T) {
	c := chart.New(
		chart.Title("Visitors per month"),
		chart.Labels("Jan", "Feb", "Mar", "Apr"),
		chart.Series("Visitors", []float64{120, 190, 170, 260}),
		chart.Series("Signups", []float64{40, 60, 55, 100}, chart.Colored(chart.Emerald)),
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

	// The accessible name lives on aria-label only — no <title>, which
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
	// overlay — one per datum, positioned in percentages.
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
