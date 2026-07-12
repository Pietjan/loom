package loom_test

// Golden coverage for the small foundation components; behavioral tests
// live in each package.

import (
	"testing"

	"github.com/pietjan/loom/callout"
	"github.com/pietjan/loom/card"
	"github.com/pietjan/loom/heading"
	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/testutil"
	"github.com/pietjan/loom/link"
	"github.com/pietjan/loom/separator"
	"github.com/pietjan/loom/text"
)

func TestFoundationGoldens(t *testing.T) {
	t.Run("heading", func(t *testing.T) {
		testutil.Golden(t, "heading-h1-xl",
			testutil.WithChildren(heading.New(heading.Level(1), heading.XL), testutil.Text("Dashboard")))
	})
	t.Run("text", func(t *testing.T) {
		testutil.Golden(t, "text-subtle",
			testutil.WithChildren(text.New(text.Subtle), testutil.Text("Secondary line")))
	})
	t.Run("separator", func(t *testing.T) {
		testutil.Golden(t, "separator", separator.New())
		testutil.Golden(t, "separator-vertical", separator.New(separator.Vertical))
	})
	t.Run("link", func(t *testing.T) {
		testutil.Golden(t, "link-external",
			testutil.WithChildren(link.New("https://example.com", link.External()), testutil.Text("Docs")))
	})
	t.Run("card", func(t *testing.T) {
		testutil.Golden(t, "card",
			testutil.WithChildren(card.New(), testutil.Text("Content")))
	})
	t.Run("callout", func(t *testing.T) {
		testutil.Golden(t, "callout-warning",
			testutil.WithChildren(callout.New(callout.Warning), testutil.Sequence(
				icon.New(icon.ExclamationTriangle),
				testutil.WithChildren(callout.Heading(), testutil.Text("Heads up")),
				testutil.WithChildren(callout.Text(), testutil.Text("Something needs attention.")),
			)))
	})
}
