package cam

import (
	"math"
	"strings"
	"testing"
)

func square(size float64) InputPath {
	return InputPath{
		Points: []Point2D{
			{0, 0}, {size, 0}, {size, size}, {0, size},
		},
		Closed: true,
	}
}

func TestVGrooveConstantDepth(t *testing.T) {
	paths := []InputPath{{
		Points: []Point2D{{0, 0}, {100, 0}, {100, 50}},
	}}
	params := VCarveParams{
		Tool:       VBit{Angle: 90, TipWidth: 0},
		FeedRate:   1000,
		PlungeRate: 300,
		SafeZ:      5,
		MaxDepth:   10,
	}
	tps := VGroove(paths, 2.0, params)
	if len(tps) != 1 {
		t.Fatalf("got %d toolpaths, want 1", len(tps))
	}
	// 90° sharp: depth for 2mm width = 1mm
	for _, p := range tps[0].Points {
		if !approx(p.Z, -1.0) {
			t.Errorf("Z = %f, want -1.0", p.Z)
		}
	}
}

func TestVGrooveMaxDepthClamp(t *testing.T) {
	paths := []InputPath{{
		Points: []Point2D{{0, 0}, {100, 0}},
	}}
	params := VCarveParams{
		Tool:     VBit{Angle: 90, TipWidth: 0},
		MaxDepth: 0.5,
	}
	tps := VGroove(paths, 10.0, params) // wants 5mm deep, clamped to 0.5
	for _, p := range tps[0].Points {
		if !approx(p.Z, -0.5) {
			t.Errorf("Z = %f, want -0.5 (clamped)", p.Z)
		}
	}
}

func TestVGrooveEmptyPath(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}}}}
	tps := VGroove(paths, 2.0, VCarveParams{Tool: VBit{Angle: 90}})
	if len(tps) != 0 {
		t.Errorf("single-point path should produce no toolpaths")
	}
}

func TestVCarveSquare(t *testing.T) {
	paths := []InputPath{square(10)}
	params := VCarveParams{
		Tool:     VBit{Angle: 90, TipWidth: 0},
		MaxDepth: 3,
		StepDown: 1,
		FeedRate: 1000,
	}
	tps := VCarve(paths, params)
	if len(tps) == 0 {
		t.Fatal("expected at least one toolpath for square")
	}

	// Each toolpath should be at the correct Z
	for _, tp := range tps {
		z := tp.Points[0].Z
		if z > 0 || z < -3.0-0.01 {
			t.Errorf("Z = %f, should be between -3 and 0", z)
		}
	}
}

func TestVCarveSquareDepths(t *testing.T) {
	paths := []InputPath{square(20)}
	params := VCarveParams{
		Tool:     VBit{Angle: 90, TipWidth: 0},
		MaxDepth: 3,
		StepDown: 1,
	}
	tps := VCarve(paths, params)

	depths := map[float64]bool{}
	for _, tp := range tps {
		depths[math.Round(tp.Points[0].Z*100)/100] = true
	}
	for _, d := range []float64{-1, -2, -3} {
		if !depths[d] {
			t.Errorf("missing toolpath at depth %f", d)
		}
	}
}

func TestVCarveSinglePass(t *testing.T) {
	paths := []InputPath{square(20)}
	params := VCarveParams{
		Tool:     VBit{Angle: 90, TipWidth: 0},
		MaxDepth: 2,
		StepDown: 0, // single pass
	}
	tps := VCarve(paths, params)
	if len(tps) == 0 {
		t.Fatal("expected toolpath for single pass")
	}
	for _, tp := range tps {
		if !approx(tp.Points[0].Z, -2.0) {
			t.Errorf("single pass Z = %f, want -2.0", tp.Points[0].Z)
		}
	}
}

func TestVCarveOpenPathSkipped(t *testing.T) {
	paths := []InputPath{{
		Points: []Point2D{{0, 0}, {10, 0}, {10, 10}},
		Closed: false,
	}}
	tps := VCarve(paths, VCarveParams{
		Tool:     VBit{Angle: 90},
		MaxDepth: 2,
		StepDown: 1,
	})
	if len(tps) != 0 {
		t.Errorf("open path should be skipped by VCarve, got %d paths", len(tps))
	}
}

func TestVCarveCollapse(t *testing.T) {
	// Small square with large offset should collapse
	paths := []InputPath{square(2)}
	params := VCarveParams{
		Tool:     VBit{Angle: 90, TipWidth: 0},
		MaxDepth: 10, // offset = 10, much larger than 2x2 square
		StepDown: 10,
	}
	tps := VCarve(paths, params)
	if len(tps) != 0 {
		t.Errorf("expected collapse (no toolpaths), got %d", len(tps))
	}
}

func TestVCarveGCode(t *testing.T) {
	tps := []Toolpath{{
		Points: []Point3D{{0, 0, -1}, {10, 0, -1}, {10, 10, -1}},
		Closed: false,
	}}
	params := VCarveParams{
		Tool:       VBit{Angle: 90, TipWidth: 0.1},
		FeedRate:   1000,
		PlungeRate: 300,
		SafeZ:      5,
		CO2Assist:  true,
	}
	gcode := VCarveGCode(tps, params)
	joined := strings.Join(gcode, "\n")

	checks := []string{
		"G90 G21",
		"M7",
		"G0 Z5.000",
		"G0 X0.000 Y0.000",
		"G1 Z-1.000 F300",
		"G1 X10.000 Y0.000 Z-1.000 F1000",
		"G1 X10.000 Y10.000 Z-1.000 F1000",
		"M9",
		"M2",
	}
	for _, c := range checks {
		if !strings.Contains(joined, c) {
			t.Errorf("G-code missing: %q", c)
		}
	}
}

func TestVCarveGCodeClosed(t *testing.T) {
	tps := []Toolpath{{
		Points: []Point3D{{0, 0, -1}, {10, 0, -1}, {10, 10, -1}},
		Closed: true,
	}}
	params := VCarveParams{
		Tool:       VBit{Angle: 90},
		FeedRate:   1000,
		PlungeRate: 300,
		SafeZ:      5,
	}
	gcode := VCarveGCode(tps, params)
	joined := strings.Join(gcode, "\n")

	// Closed path should return to first point
	if !strings.Contains(joined, "G1 X0.000 Y0.000 Z-1.000") {
		t.Error("closed path should have closing move back to first point")
	}
	// No CO2 commands when disabled
	if strings.Contains(joined, "M7") {
		t.Error("M7 should not be present when CO2Assist is false")
	}
}

func TestVCarveGCodeNoCO2(t *testing.T) {
	tps := []Toolpath{{
		Points: []Point3D{{0, 0, -1}},
	}}
	params := VCarveParams{
		Tool:      VBit{Angle: 90},
		SafeZ:     5,
		CO2Assist: false,
	}
	gcode := VCarveGCode(tps, params)
	joined := strings.Join(gcode, "\n")
	if strings.Contains(joined, "M7") || strings.Contains(joined, "M9") {
		t.Error("should not have coolant commands without CO2Assist")
	}
}

func TestOffsetPolygonSquare(t *testing.T) {
	pts := []Point2D{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	result := offsetPolygon(pts, 1.0)
	if len(result) == 0 {
		t.Fatal("expected offset result for 10x10 square with 1mm offset")
	}
	// Offset 1mm inward on a 10x10 square → 8x8 square
	r := result[0]
	minX, maxX := r[0].X, r[0].X
	minY, maxY := r[0].Y, r[0].Y
	for _, p := range r {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	if !approx(maxX-minX, 8.0) {
		t.Errorf("offset width = %f, want 8.0", maxX-minX)
	}
	if !approx(maxY-minY, 8.0) {
		t.Errorf("offset height = %f, want 8.0", maxY-minY)
	}
}

func TestOffsetPolygonCollapse(t *testing.T) {
	pts := []Point2D{{0, 0}, {4, 0}, {4, 4}, {0, 4}}
	result := offsetPolygon(pts, 3.0) // 4x4 square, 3mm offset → collapses
	if len(result) != 0 {
		t.Errorf("expected collapse for 4x4 square with 3mm offset")
	}
}

func TestOffsetPolygonTriangle(t *testing.T) {
	pts := []Point2D{{0, 0}, {20, 0}, {10, 17.32}}
	result := offsetPolygon(pts, 1.0)
	if len(result) == 0 {
		t.Fatal("expected offset result for triangle")
	}
	// Should be smaller than original
	origArea := math.Abs(signedArea(pts))
	resultArea := math.Abs(signedArea(result[0]))
	if resultArea >= origArea {
		t.Errorf("offset area %f should be less than original %f", resultArea, origArea)
	}
}

func TestSignedArea(t *testing.T) {
	// CCW square: positive area
	ccw := []Point2D{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
	if signedArea(ccw) <= 0 {
		t.Errorf("CCW square area = %f, want positive", signedArea(ccw))
	}

	// CW square: negative area
	cw := []Point2D{{0, 0}, {0, 10}, {10, 10}, {10, 0}}
	if signedArea(cw) >= 0 {
		t.Errorf("CW square area = %f, want negative", signedArea(cw))
	}
}

func TestVCarveGCodeString(t *testing.T) {
	tps := []Toolpath{{Points: []Point3D{{0, 0, -1}}}}
	params := VCarveParams{
		Tool:  VBit{Angle: 90},
		SafeZ: 5,
	}
	s := VCarveGCodeString(tps, params)
	if !strings.HasSuffix(s, "\n") {
		t.Error("GCodeString should end with newline")
	}
	if !strings.Contains(s, "M2") {
		t.Error("GCodeString missing M2")
	}
}
