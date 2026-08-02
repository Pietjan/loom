package loom_test

// Every component that takes a color draws from internal/palette. These
// tests are what stops the four enums and the one table from drifting:
// adding a hue to a component without a row (or renaming a row out from
// under a component) renders an element with no color at all, which is
// easy to miss by eye and impossible to miss here.

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/badge"
	"github.com/pietjan/loom/chart"
	"github.com/pietjan/loom/diagram"
	"github.com/pietjan/loom/internal/palette"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/progress"
	"github.com/pietjan/loom/timeline"
)

// hues is the palette as the components spell it - the string values line
// up, which is why a component converts its Color rather than mapping it.
var hues = func() []string {
	out := make([]string, 0, len(palette.All))
	for _, c := range palette.All {
		out = append(out, string(c))
	}
	return out
}()

func TestEveryHueResolves(t *testing.T) {
	for _, hue := range hues {
		t.Run(hue, func(t *testing.T) {
			// A tinted badge and the timeline indicator share the Tint
			// row, so both must carry the hue's background and text.
			b := testutil.Render(t, testutil.WithChildren(
				badge.New(badge.WithColor(badge.Color(hue))), testutil.Text("x"),
			))
			wantTint := []string{"bg-" + hue + "-400/", "text-" + hue + "-", "dark:bg-" + hue + "-400/"}
			for _, want := range wantTint {
				if !strings.Contains(b, want) {
					t.Errorf("tinted badge %s missing %q: %s", hue, want, b)
				}
			}

			solid := testutil.Render(t, testutil.WithChildren(
				badge.New(badge.WithColor(badge.Color(hue)), badge.Solid()), testutil.Text("x"),
			))
			if !strings.Contains(solid, "bg-"+hue+"-") {
				t.Errorf("solid badge %s has no fill: %s", hue, solid)
			}

			ind := testutil.Render(t, testutil.WithChildren(
				timeline.Indicator(timeline.WithColor(timeline.Color(hue))), testutil.Text("1"),
			))
			if !strings.Contains(ind, "bg-"+hue+"-400/") {
				t.Errorf("timeline indicator %s not tinted: %s", hue, ind)
			}

			bar := testutil.Render(t, progress.New(
				progress.Value(50), progress.WithColor(progress.Color(hue)),
			))
			if !strings.Contains(bar, "bg-"+hue+"-500") {
				t.Errorf("progress bar %s not at mark strength: %s", hue, bar)
			}

			node := testutil.Render(t, testutil.WithChildren(
				diagram.New(), testutil.WithChildren(
					diagram.Node("n", diagram.WithColor(diagram.Color(hue))), testutil.Text("n"),
				),
			))
			if !strings.Contains(node, "stroke-"+hue+"-500") {
				t.Errorf("diagram node %s outline not colored: %s", hue, node)
			}

			line := testutil.Render(t, chart.New(chart.Area(), chart.Legend(),
				chart.Series("s", []float64{1, 2}, chart.Colored(chart.Color(hue)))))
			for _, want := range []string{"stroke-" + hue + "-500", "fill-" + hue + "-500/15", "bg-" + hue + "-500"} {
				if !strings.Contains(line, want) {
					t.Errorf("chart series %s missing %q: %s", hue, want, line)
				}
			}
		})
	}
}

// Accent is a theme token, not a hue: it has no palette row, and the
// components that offer it must fall through to the accent utilities
// rather than render uncolored.
func TestAccentIsNotAHue(t *testing.T) {
	if _, ok := palette.Of("accent"); ok {
		t.Fatal("accent must not have a palette row - it is a theme token")
	}

	bar := testutil.Render(t, progress.New(progress.Value(50)))
	if !strings.Contains(bar, "bg-accent") {
		t.Errorf("default progress bar is not accent: %s", bar)
	}

	line := testutil.Render(t, chart.New(chart.Legend(),
		chart.Series("s", []float64{1, 2}, chart.Accent)))
	if !strings.Contains(line, "stroke-accent") {
		t.Errorf("accent series is not accent: %s", line)
	}

	// A diagram node must always land on exactly one stroke: an accent, a
	// hue, or the neutral chrome - never two, whose winner would be decided
	// by the order Tailwind happened to emit them in.
	for _, tc := range []struct {
		name  string
		color diagram.Color
		want  string
	}{
		{"accent", diagram.ColorAccent, "stroke-accent"},
		{"default", diagram.ColorDefault, "stroke-base-300"},
		{"unknown", diagram.Color("chartreuse"), "stroke-base-300"},
	} {
		out := testutil.Render(t, testutil.WithChildren(
			diagram.New(), testutil.WithChildren(
				diagram.Node("n", diagram.WithColor(tc.color)), testutil.Text("n"),
			),
		))
		shape := out[strings.Index(out, `data-ui="diagram-shape"`):]
		shape = shape[:strings.Index(shape, ">")]
		// One stroke color per variant. stroke-1 is the width, not a color.
		var light, dark int
		for class := range strings.FieldsSeq(shape) {
			switch {
			case class == "stroke-1":
			case strings.HasPrefix(class, "stroke-"):
				light++
			case strings.HasPrefix(class, "dark:stroke-"):
				dark++
			}
		}
		if light != 1 || dark > 1 {
			t.Errorf("diagram node (%s) wears %d light and %d dark strokes, want 1 and at most 1: %s",
				tc.name, light, dark, shape)
		}
		if !strings.Contains(shape, tc.want) {
			t.Errorf("diagram node (%s) missing %q: %s", tc.name, tc.want, shape)
		}
	}
}
