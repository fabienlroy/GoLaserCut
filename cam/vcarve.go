package cam

import (
	"fmt"
	"math"
	"strings"
)

// Point2D is a 2D point for toolpath input.
type Point2D struct {
	X, Y float64
}

// Point3D is a 3D toolpath point.
type Point3D struct {
	X, Y, Z float64
}

// InputPath is a 2D path for V-carving input.
type InputPath struct {
	Points []Point2D
	Closed bool
}

// Toolpath is a sequence of 3D points the tool follows.
type Toolpath struct {
	Points []Point3D
	Closed bool
}

// VCarveParams configures V-carving toolpath generation.
type VCarveParams struct {
	Tool       VBit
	FeedRate   float64 // XY cutting speed, mm/min
	PlungeRate float64 // Z plunge speed, mm/min
	SafeZ      float64 // safe retract height, mm (positive)
	MaxDepth   float64 // max cut depth, mm (positive)
	StepDown   float64 // depth per pass, mm (0 = single pass)
	CO2Assist  bool
}

// VGroove generates constant-depth grooves along paths.
// Depth is computed from targetWidth using the VBit geometry.
func VGroove(paths []InputPath, targetWidth float64, params VCarveParams) []Toolpath {
	depth := params.Tool.DepthForWidth(targetWidth)
	if params.MaxDepth > 0 && depth > params.MaxDepth {
		depth = params.MaxDepth
	}
	z := -depth

	var result []Toolpath
	for _, path := range paths {
		if len(path.Points) < 2 {
			continue
		}
		tp := Toolpath{Closed: path.Closed}
		for _, p := range path.Points {
			tp.Points = append(tp.Points, Point3D{p.X, p.Y, z})
		}
		result = append(result, tp)
	}
	return result
}

// VCarve generates adaptive-depth V-carving for closed paths using
// inward polygon offsets. Each offset level produces a contour at the
// corresponding Z depth, creating the classic V-carved profile.
func VCarve(paths []InputPath, params VCarveParams) []Toolpath {
	tanHalf := math.Tan(params.Tool.halfAngleRad())
	stepDown := params.StepDown
	if stepDown <= 0 {
		stepDown = params.MaxDepth
	}
	if stepDown <= 0 {
		return nil
	}

	var result []Toolpath
	for _, path := range paths {
		if !path.Closed || len(path.Points) < 3 {
			continue
		}

		for depth := stepDown; depth <= params.MaxDepth+1e-9; depth += stepDown {
			if depth > params.MaxDepth {
				depth = params.MaxDepth
			}
			offset := depth * tanHalf
			if params.Tool.TipWidth > 0 {
				offset += params.Tool.TipWidth / 2
			}

			offsetPaths := offsetPolygon(path.Points, offset)
			for _, op := range offsetPaths {
				tp := Toolpath{Closed: true}
				for _, p := range op {
					tp.Points = append(tp.Points, Point3D{p.X, p.Y, -depth})
				}
				result = append(result, tp)
			}

			if depth >= params.MaxDepth {
				break
			}
		}
	}
	return result
}

// VCarveGCode generates G-code for V-carve toolpaths.
func VCarveGCode(toolpaths []Toolpath, params VCarveParams) []string {
	var lines []string
	w := func(s string) { lines = append(lines, s) }

	w(fmt.Sprintf("; V-carve: %.1f° V-bit, %.2fmm tip", params.Tool.Angle, params.Tool.TipWidth))
	w("G90 G21")
	w("G17")
	if params.CO2Assist {
		w("M7 ; CO2 assist on")
	}
	w(fmt.Sprintf("G0 Z%.3f", params.SafeZ))

	for _, tp := range toolpaths {
		if len(tp.Points) == 0 {
			continue
		}
		p0 := tp.Points[0]
		w(fmt.Sprintf("G0 X%.3f Y%.3f", p0.X, p0.Y))
		w(fmt.Sprintf("G1 Z%.3f F%.0f", p0.Z, params.PlungeRate))

		for _, p := range tp.Points[1:] {
			w(fmt.Sprintf("G1 X%.3f Y%.3f Z%.3f F%.0f", p.X, p.Y, p.Z, params.FeedRate))
		}

		if tp.Closed && len(tp.Points) > 1 {
			w(fmt.Sprintf("G1 X%.3f Y%.3f Z%.3f F%.0f", p0.X, p0.Y, p0.Z, params.FeedRate))
		}

		w(fmt.Sprintf("G0 Z%.3f", params.SafeZ))
	}

	if params.CO2Assist {
		w("M9 ; CO2 assist off")
	}
	w("M5")
	w("M2")

	return lines
}

// VCarveGCodeString returns the full G-code as a single string.
func VCarveGCodeString(toolpaths []Toolpath, params VCarveParams) string {
	return strings.Join(VCarveGCode(toolpaths, params), "\n") + "\n"
}

// Polygon offset (inward shrink)

func signedArea(pts []Point2D) float64 {
	n := len(pts)
	a := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		a += pts[i].X*pts[j].Y - pts[j].X*pts[i].Y
	}
	return a / 2
}

// offsetPolygon shrinks a closed polygon inward by dist.
// Returns nil if the polygon collapses. May return multiple polygons
// if the offset causes the shape to split.
func offsetPolygon(pts []Point2D, dist float64) [][]Point2D {
	n := len(pts)
	if n < 3 || dist <= 0 {
		return nil
	}

	area := signedArea(pts)
	if math.Abs(area) < 1e-12 {
		return nil
	}
	sign := 1.0
	if area < 0 {
		sign = -1.0
	}

	type offsetEdge struct {
		px, py float64 // point on offset line
		dx, dy float64 // direction (normalized)
	}
	edges := make([]offsetEdge, n)

	for i := 0; i < n; i++ {
		j := (i + 1) % n
		ex := pts[j].X - pts[i].X
		ey := pts[j].Y - pts[i].Y
		l := math.Sqrt(ex*ex + ey*ey)
		if l < 1e-12 {
			edges[i] = edges[(i-1+n)%n]
			continue
		}
		// Inward normal: sign * (-ey, ex) / l
		nx := sign * (-ey / l)
		ny := sign * (ex / l)
		edges[i] = offsetEdge{
			px: pts[i].X + nx*dist,
			py: pts[i].Y + ny*dist,
			dx: ex / l,
			dy: ey / l,
		}
	}

	result := make([]Point2D, 0, n)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		pt, ok := lineIntersect(
			edges[i].px, edges[i].py, edges[i].dx, edges[i].dy,
			edges[j].px, edges[j].py, edges[j].dx, edges[j].dy,
		)
		if !ok {
			result = append(result, Point2D{
				X: (edges[i].px + edges[j].px) / 2,
				Y: (edges[i].py + edges[j].py) / 2,
			})
			continue
		}
		result = append(result, pt)
	}

	if len(result) < 3 {
		return nil
	}

	// Detect collapse: when offset exceeds inradius, result edges
	// reverse direction relative to their corresponding offset edges.
	// Result edge i→i+1 lies on offset edge (i+1)%n.
	for i := 0; i < len(result); i++ {
		j := (i + 1) % len(result)
		edgeIdx := (i + 1) % n
		rdx := result[j].X - result[i].X
		rdy := result[j].Y - result[i].Y
		dot := rdx*edges[edgeIdx].dx + rdy*edges[edgeIdx].dy
		if dot < 0 {
			return nil
		}
	}

	resultArea := signedArea(result)
	if (area > 0 && resultArea <= 0) || (area < 0 && resultArea >= 0) {
		return nil
	}

	cleaned := removeSelfintersections(result)
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}

func lineIntersect(px1, py1, dx1, dy1, px2, py2, dx2, dy2 float64) (Point2D, bool) {
	cross := dx1*dy2 - dy1*dx2
	if math.Abs(cross) < 1e-12 {
		return Point2D{}, false
	}
	dpx := px2 - px1
	dpy := py2 - py1
	t := (dpx*dy2 - dpy*dx2) / cross
	return Point2D{
		X: px1 + t*dx1,
		Y: py1 + t*dy1,
	}, true
}

// removeSelfintersections checks for self-intersections and returns
// valid sub-polygons. For simple cases, returns the polygon as-is.
func removeSelfintersections(pts []Point2D) [][]Point2D {
	n := len(pts)
	if n < 3 {
		return nil
	}

	for i := 0; i < n; i++ {
		a1 := pts[i]
		a2 := pts[(i+1)%n]
		for j := i + 2; j < n; j++ {
			if j == n-1 && i == 0 {
				continue // adjacent edges
			}
			b1 := pts[j]
			b2 := pts[(j+1)%n]
			if _, ok := segmentIntersect(a1, a2, b1, b2); ok {
				return nil // collapsed, discard
			}
		}
	}

	return [][]Point2D{pts}
}

func segmentIntersect(a1, a2, b1, b2 Point2D) (Point2D, bool) {
	dx1 := a2.X - a1.X
	dy1 := a2.Y - a1.Y
	dx2 := b2.X - b1.X
	dy2 := b2.Y - b1.Y

	cross := dx1*dy2 - dy1*dx2
	if math.Abs(cross) < 1e-12 {
		return Point2D{}, false
	}

	dpx := b1.X - a1.X
	dpy := b1.Y - a1.Y
	t := (dpx*dy2 - dpy*dx2) / cross
	u := (dpx*dy1 - dpy*dx1) / cross

	if t < 0 || t > 1 || u < 0 || u > 1 {
		return Point2D{}, false
	}
	return Point2D{a1.X + t*dx1, a1.Y + t*dy1}, true
}
