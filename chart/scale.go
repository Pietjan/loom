package chart

import (
	"math"
	"strconv"
)

// niceTicks returns ~count ascending tick values covering [min, max],
// stepped at a "nice" interval (1, 2, or 5 × 10ⁿ). The returned ticks
// extend to the step boundaries enclosing the domain.
func niceTicks(min, max float64, count int) []float64 {
	if count < 2 {
		count = 2
	}
	if min == max {
		max = min + 1
	}

	step := niceStep((max - min) / float64(count-1))
	lo := math.Floor(min/step) * step
	hi := math.Ceil(max/step) * step

	var ticks []float64
	// Guard float drift with a half-step epsilon.
	for v := lo; v <= hi+step/2; v += step {
		// Normalize -0 and drift like 0.30000000000000004.
		ticks = append(ticks, round(v, step))
	}
	return ticks
}

// niceStep rounds a raw step up to 1, 2, or 5 × 10ⁿ.
func niceStep(raw float64) float64 {
	if raw <= 0 {
		return 1
	}
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	switch norm := raw / mag; {
	case norm <= 1:
		return mag
	case norm <= 2:
		return 2 * mag
	case norm <= 5:
		return 5 * mag
	default:
		return 10 * mag
	}
}

// round snaps v to the step's decimal precision.
func round(v, step float64) float64 {
	decimals := 0
	if step < 1 {
		decimals = int(math.Ceil(-math.Log10(step)))
	}
	pow := math.Pow(10, float64(decimals))
	r := math.Round(v*pow) / pow
	if r == 0 {
		return 0 // never -0
	}
	return r
}

// domain returns the y range covering all series: zero-based when all
// values are non-negative (bars and most dashboards read better anchored
// at zero), nice-floored otherwise.
func domain(series []series) (float64, float64) {
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, s := range series {
		for _, v := range s.values {
			lo = math.Min(lo, v)
			hi = math.Max(hi, v)
		}
	}
	if math.IsInf(lo, 1) {
		return 0, 1
	}
	if lo >= 0 {
		lo = 0
	}
	if lo == hi {
		hi = lo + 1
	}
	return lo, hi
}

// formatValue renders a tick or point value compactly: integers without
// decimals, large values as 1.2k / 3.4M.
func formatValue(v float64) string {
	abs := math.Abs(v)
	switch {
	case abs >= 1_000_000:
		return trimZero(v/1_000_000) + "M"
	case abs >= 1_000:
		return trimZero(v/1_000) + "k"
	default:
		return trimZero(v)
	}
}

func trimZero(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// fmtCoord renders an SVG coordinate with sub-pixel precision but no
// float-drift noise.
func fmtCoord(v float64) string {
	return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64)
}
