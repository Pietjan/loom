package badge_test

import (
	"strings"
	"testing"

	"github.com/pietjan/loom/badge"
	"github.com/pietjan/loom/internal/testutil"
)

func TestGolden(t *testing.T) {
	testutil.Golden(t, "badge-green",
		testutil.WithChildren(badge.New(badge.Green), testutil.Text("Active")))
}

func TestPill(t *testing.T) {
	out := testutil.Render(t, testutil.WithChildren(badge.New(badge.Pill()), testutil.Text("42")))
	if !strings.Contains(out, "rounded-full") {
		t.Fatalf("expected rounded-full: %s", out)
	}
	if strings.Contains(out, "rounded-md") {
		t.Fatalf("rounded-md should be replaced by pill: %s", out)
	}
}
