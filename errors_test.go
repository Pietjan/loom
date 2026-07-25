package loom_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/a-h/templ"

	"github.com/pietjan/loom/chart"
	"github.com/pietjan/loom/diagram"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/modal"
)

// TestErrorsAreMatchable: a component that fails loudly must do so with an
// error a caller can identify. Several of these carry detail (which series,
// which node id) around a wrapped sentinel, so only errors.Is can see through
// to the cause — a plain equality check would pass today and rot the moment
// the message gains context.
func TestErrorsAreMatchable(t *testing.T) {
	node := func(id string) templ.Component {
		return testutil.WithChildren(diagram.Node(id), testutil.Text(id))
	}

	for _, tc := range []struct {
		name string
		got  error
		want error
	}{
		{
			"chart with no series",
			testutil.RenderErr(chart.New()),
			chart.ErrNoSeries,
		},
		{
			"chart series with no values",
			testutil.RenderErr(chart.New(chart.Series("visits", nil))),
			chart.ErrNoValues,
		},
		{
			"chart series of differing lengths",
			testutil.RenderErr(chart.New(
				chart.Series("a", []float64{1, 2}),
				chart.Series("b", []float64{1}),
			)),
			chart.ErrMisalignedSeries,
		},
		{
			"chart labels not matching values",
			testutil.RenderErr(chart.New(
				chart.Series("a", []float64{1, 2}),
				chart.Labels("only-one"),
			)),
			chart.ErrMisalignedSeries,
		},
		{
			"diagram with no nodes",
			testutil.RenderErr(diagram.New()),
			diagram.ErrNoNodes,
		},
		{
			"diagram with a duplicate node id",
			testutil.RenderErr(testutil.WithChildren(
				diagram.New(), testutil.Sequence(node("a"), node("a")),
			)),
			diagram.ErrDuplicateNode,
		},
		{
			"diagram edge naming an undeclared node",
			testutil.RenderErr(testutil.WithChildren(
				diagram.New(diagram.Edge("a", "nowhere")), node("a"),
			)),
			diagram.ErrUnknownNode,
		},
		{
			"diagram self-loop",
			testutil.RenderErr(testutil.WithChildren(
				diagram.New(diagram.Edge("a", "a")), node("a"),
			)),
			diagram.ErrSelfLoop,
		},
		{
			"modal trigger with no dialog",
			testutil.RenderErr(testutil.WithChildren(modal.Trigger(), testutil.Text("x"))),
			modal.ErrNoTarget,
		},
		{
			"modal content with no dialog",
			testutil.RenderErr(testutil.WithChildren(modal.Content(), testutil.Text("x"))),
			modal.ErrNoContentID,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.got, tc.want) {
				t.Errorf("got %v, want errors.Is(err, %v)", tc.got, tc.want)
			}
		})
	}
}

// TestModalFailuresStayDistinct: a trigger and a piece of content both fail
// for the same underlying reason — no dialog id in scope — and each wraps that
// cause with the option that would fix it. Sharing the cause must not make the
// two indistinguishable, or a caller cannot tell which end is misconfigured.
func TestModalFailuresStayDistinct(t *testing.T) {
	trigger := testutil.RenderErr(testutil.WithChildren(modal.Trigger(), testutil.Text("x")))
	content := testutil.RenderErr(testutil.WithChildren(modal.Content(), testutil.Text("x")))

	if errors.Is(trigger, modal.ErrNoContentID) {
		t.Error("a trigger failure should not match ErrNoContentID")
	}
	if errors.Is(content, modal.ErrNoTarget) {
		t.Error("a content failure should not match ErrNoTarget")
	}
	// The advice has to match the end that is broken.
	if !strings.Contains(trigger.Error(), "modal.For") {
		t.Errorf("trigger error should point at modal.For: %v", trigger)
	}
	if !strings.Contains(content.Error(), "modal.Name") {
		t.Errorf("content error should point at modal.Name: %v", content)
	}
}
