package icon_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

func TestGolden(t *testing.T) {
	testutil.Golden(t, "icon-bell-mini", icon.New(icon.Bell, icon.Mini))
}

func TestVariantsKeepTheirViewBox(t *testing.T) {
	for _, tc := range []struct {
		variant icon.Option
		viewBox string
	}{
		{icon.Outline, "0 0 24 24"},
		{icon.Mini, "0 0 20 20"},
		{icon.Micro, "0 0 16 16"},
	} {
		variant, viewBox := tc.variant, tc.viewBox
		tree := testutil.Parse(t, testutil.Render(t, icon.New(icon.Bell, variant)))
		svg := dom.Find(tree, dom.ByMarker("icon"))
		if svg == nil {
			t.Fatal("no icon marker")
		}
		if got := dom.GetAttr(svg, "viewbox"); got != viewBox {
			t.Fatalf("viewBox %q, want %q", got, viewBox)
		}
	}
}

func TestDecorativeByDefault(t *testing.T) {
	tree := testutil.Parse(t, testutil.Render(t, icon.New(icon.Bell)))
	svg := dom.Find(tree, dom.ByMarker("icon"))
	if got := dom.GetAttr(svg, "aria-hidden"); got != "true" {
		t.Fatalf("aria-hidden: %q", got)
	}
}

func TestUnknownIconFails(t *testing.T) {
	err := testutil.RenderErr(icon.New(icon.Name("no-such-icon")))
	if err == nil || !strings.Contains(err.Error(), "no-such-icon") {
		t.Fatalf("expected loud failure, got %v", err)
	}
}
