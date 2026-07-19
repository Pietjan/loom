package icon_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/icon"
	"github.com/pietjan/loom/internal/dom"
	"github.com/pietjan/loom/internal/testutil"
)

func TestGolden(t *testing.T) {
	testutil.Golden(t, "icon-bell-small", icon.New(icon.Bell, icon.Small))
}

// Phosphor draws every weight on one grid, so sizing is CSS-only: the
// viewBox must survive untouched whatever variant or size is asked for.
func TestSizesShareOneViewBox(t *testing.T) {
	for _, option := range []icon.Option{icon.Regular, icon.Fill, icon.Small, icon.ExtraSmall} {
		tree := testutil.Parse(t, testutil.Render(t, icon.New(icon.Bell, option)))
		svg := dom.Find(tree, dom.ByMarker("icon"))
		if svg == nil {
			t.Fatal("no icon marker")
		}
		if got := dom.GetAttr(svg, "viewbox"); got != "0 0 256 256" {
			t.Fatalf("viewBox %q, want %q", got, "0 0 256 256")
		}
	}
}

func TestSizeSetsTheSizeUtility(t *testing.T) {
	for _, tc := range []struct {
		option icon.Option
		class  string
	}{
		{nil, "size-6"},
		{icon.Small, "size-5"},
		{icon.ExtraSmall, "size-4"},
	} {
		options := []icon.Option{}
		if tc.option != nil {
			options = append(options, tc.option)
		}
		tree := testutil.Parse(t, testutil.Render(t, icon.New(icon.Bell, options...)))
		svg := dom.Find(tree, dom.ByMarker("icon"))
		if got := dom.GetAttr(svg, "class"); !strings.Contains(got, tc.class) {
			t.Fatalf("class %q, want it to contain %q", got, tc.class)
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
