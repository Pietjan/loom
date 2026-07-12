package chart

import (
	"reflect"
	"testing"
)

func TestNiceTicks(t *testing.T) {
	for _, tc := range []struct {
		min, max float64
		count    int
		want     []float64
	}{
		{0, 300, 4, []float64{0, 100, 200, 300}},
		{0, 260, 4, []float64{0, 100, 200, 300}},
		{0, 5, 4, []float64{0, 2, 4, 6}},
		{0, 0.9, 4, []float64{0, 0.5, 1}},
		// Step rounds UP to the next nice unit: fewer, rounder ticks.
		{-50, 120, 4, []float64{-100, 0, 100, 200}},
	} {
		got := niceTicks(tc.min, tc.max, tc.count)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("niceTicks(%v, %v, %d) = %v, want %v", tc.min, tc.max, tc.count, got, tc.want)
		}
	}
}

func TestNiceTicksCoverDomain(t *testing.T) {
	ticks := niceTicks(3, 987, 5)
	if ticks[0] > 3 || ticks[len(ticks)-1] < 987 {
		t.Fatalf("ticks %v do not cover [3, 987]", ticks)
	}
}

func TestFormatValue(t *testing.T) {
	for v, want := range map[float64]string{
		0:         "0",
		42:        "42",
		1500:      "1.5k",
		2_400_000: "2.4M",
		0.5:       "0.5",
	} {
		if got := formatValue(v); got != want {
			t.Errorf("formatValue(%v) = %q, want %q", v, got, want)
		}
	}
}
