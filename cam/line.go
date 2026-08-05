package cam

import (
	"fmt"
	"math"
	"strings"
)

// LineParams configures contour laser cutting.
type LineParams struct {
	FeedRate  float64 // cutting speed, mm/min
	Power     float64 // laser power 0–1000 (S value)
	Passes    int     // number of passes (≥1)
	SafeZ     float64 // retract height, mm (0 = 2D only)
	LaserMode bool    // true = M4 dynamic power, false = M3 constant
	AirAssist bool    // M7 air/CO₂ assist
}

// LineCut generates G-code for cutting along 2D contour paths.
// Each path is traced at the configured speed/power for the given
// number of passes. Closed paths return to their start point.
func LineCut(paths []InputPath, params LineParams) []string {
	passes := params.Passes
	if passes < 1 {
		passes = 1
	}

	var lines []string
	w := func(s string) { lines = append(lines, s) }

	w("; Line cut")
	w("G90 G21")
	w("G17")
	if params.LaserMode {
		w("M4")
	} else {
		w("M3")
	}
	w(fmt.Sprintf("S%.0f", params.Power))

	if params.AirAssist {
		w("M7 ; air assist on")
	}

	if params.SafeZ > 0 {
		w(fmt.Sprintf("G0 Z%.3f", params.SafeZ))
	}

	for pass := 0; pass < passes; pass++ {
		if passes > 1 {
			w(fmt.Sprintf("; pass %d/%d", pass+1, passes))
		}
		for _, path := range paths {
			if len(path.Points) < 2 {
				continue
			}
			p0 := path.Points[0]
			w(fmt.Sprintf("G0 X%.3f Y%.3f", p0.X, p0.Y))

			if params.SafeZ > 0 {
				w(fmt.Sprintf("G1 Z0 F%.0f", params.FeedRate))
			}

			for _, p := range path.Points[1:] {
				w(fmt.Sprintf("G1 X%.3f Y%.3f F%.0f", p.X, p.Y, params.FeedRate))
			}

			if path.Closed && len(path.Points) > 2 {
				w(fmt.Sprintf("G1 X%.3f Y%.3f F%.0f", p0.X, p0.Y, params.FeedRate))
			}

			if params.SafeZ > 0 {
				w(fmt.Sprintf("G0 Z%.3f", params.SafeZ))
			}
		}
	}

	if params.AirAssist {
		w("M9 ; air assist off")
	}
	w("S0")
	w("M5")
	w("G0 X0 Y0")
	w("M2")

	return lines
}

// LineCutString returns the full G-code as a single string.
func LineCutString(paths []InputPath, params LineParams) string {
	return strings.Join(LineCut(paths, params), "\n") + "\n"
}

// OptimizePaths reorders paths to minimize rapid travel distance
// using a nearest-neighbor heuristic from the origin.
func OptimizePaths(paths []InputPath) []InputPath {
	if len(paths) <= 1 {
		return paths
	}

	result := make([]InputPath, 0, len(paths))
	used := make([]bool, len(paths))
	cur := Point2D{0, 0}

	for range paths {
		best := -1
		bestDist := math.MaxFloat64

		for i, p := range paths {
			if used[i] || len(p.Points) == 0 {
				continue
			}
			d := dist2D(cur, p.Points[0])
			if d < bestDist {
				bestDist = d
				best = i
			}
			if p.Closed && len(p.Points) > 1 {
				for j, pt := range p.Points[1:] {
					d = dist2D(cur, pt)
					if d < bestDist {
						bestDist = d
						best = i
						_ = j
					}
				}
			}
		}

		if best < 0 {
			break
		}
		used[best] = true

		p := paths[best]
		if p.Closed && len(p.Points) > 1 {
			startIdx := nearestPointIndex(cur, p.Points)
			if startIdx > 0 {
				p = rotatePath(p, startIdx)
			}
		}

		result = append(result, p)
		last := p.Points[len(p.Points)-1]
		if p.Closed {
			last = p.Points[0]
		}
		cur = last
	}

	return result
}

func dist2D(a, b Point2D) float64 {
	dx := b.X - a.X
	dy := b.Y - a.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func nearestPointIndex(from Point2D, pts []Point2D) int {
	best := 0
	bestDist := dist2D(from, pts[0])
	for i, p := range pts[1:] {
		d := dist2D(from, p)
		if d < bestDist {
			bestDist = d
			best = i + 1
		}
	}
	return best
}

func rotatePath(p InputPath, startIdx int) InputPath {
	n := len(p.Points)
	rotated := make([]Point2D, n)
	for i := 0; i < n; i++ {
		rotated[i] = p.Points[(startIdx+i)%n]
	}
	return InputPath{
		Points: rotated,
		Closed: p.Closed,
	}
}
