package cam

import (
	"fmt"
	"strings"
	"testing"
)

func TestLineCutBasic(t *testing.T) {
	paths := []InputPath{{
		Points: []Point2D{{0, 0}, {100, 0}, {100, 50}},
	}}
	params := LineParams{
		FeedRate:  1000,
		Power:     800,
		Passes:    1,
		LaserMode: true,
	}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	checks := []string{
		"G90 G21",
		"M4",
		"S800",
		"G0 X0.000 Y0.000",
		"G1 X100.000 Y0.000 F1000",
		"G1 X100.000 Y50.000 F1000",
		"S0",
		"M5",
		"M2",
	}
	for _, c := range checks {
		if !strings.Contains(joined, c) {
			t.Errorf("missing: %q", c)
		}
	}
}

func TestLineCutM3Mode(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}, {10, 0}}}}
	params := LineParams{FeedRate: 500, Power: 500, LaserMode: false}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	if !strings.Contains(joined, "M3") {
		t.Error("expected M3 for constant power mode")
	}
	if strings.Contains(joined, "M4") {
		t.Error("should not have M4 in constant power mode")
	}
}

func TestLineCutClosedPath(t *testing.T) {
	paths := []InputPath{{
		Points: []Point2D{{0, 0}, {10, 0}, {10, 10}},
		Closed: true,
	}}
	params := LineParams{FeedRate: 1000, Power: 1000, LaserMode: true}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	if !strings.Contains(joined, "G1 X0.000 Y0.000 F1000") {
		t.Error("closed path should return to start")
	}
}

func TestLineCutMultiPass(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}, {50, 0}}}}
	params := LineParams{FeedRate: 1000, Power: 800, Passes: 3, LaserMode: true}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	for i := 1; i <= 3; i++ {
		marker := fmt.Sprintf("; pass %d/3", i)
		if !strings.Contains(joined, marker) {
			t.Errorf("missing %q", marker)
		}
	}

	count := strings.Count(joined, "G0 X0.000 Y0.000")
	if count < 3 {
		t.Errorf("expected 3 rapid moves to start, got %d", count)
	}
}

func TestLineCutZeroPassesDefaultsToOne(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}, {10, 0}}}}
	params := LineParams{FeedRate: 500, Power: 500, Passes: 0, LaserMode: true}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	if strings.Contains(joined, "; pass") {
		t.Error("single pass should not have pass markers")
	}
	if !strings.Contains(joined, "G1 X10.000") {
		t.Error("should still produce cutting moves")
	}
}

func TestLineCutWithSafeZ(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}, {10, 0}}}}
	params := LineParams{FeedRate: 500, Power: 500, SafeZ: 5, LaserMode: true}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	if !strings.Contains(joined, "G0 Z5.000") {
		t.Error("expected safe Z retract")
	}
	if !strings.Contains(joined, "G1 Z0 F500") {
		t.Error("expected Z plunge to 0")
	}
}

func TestLineCutAirAssist(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}, {10, 0}}}}
	params := LineParams{FeedRate: 500, Power: 500, AirAssist: true, LaserMode: true}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	if !strings.Contains(joined, "M7") {
		t.Error("expected M7 air assist on")
	}
	if !strings.Contains(joined, "M9") {
		t.Error("expected M9 air assist off")
	}
}

func TestLineCutNoAirAssist(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}, {10, 0}}}}
	params := LineParams{FeedRate: 500, Power: 500, AirAssist: false, LaserMode: true}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	if strings.Contains(joined, "M7") || strings.Contains(joined, "M9") {
		t.Error("should not have air assist commands")
	}
}

func TestLineCutEmptyPath(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{5, 5}}}}
	params := LineParams{FeedRate: 500, Power: 500, LaserMode: true}
	gcode := LineCut(paths, params)
	joined := strings.Join(gcode, "\n")

	if strings.Contains(joined, "G1 X") {
		t.Error("single-point path should produce no cutting moves")
	}
}

func TestLineCutReturnToOrigin(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{10, 10}, {50, 50}}}}
	params := LineParams{FeedRate: 500, Power: 500, LaserMode: true}
	gcode := LineCut(paths, params)

	found := false
	for _, line := range gcode {
		if line == "G0 X0 Y0" {
			found = true
		}
	}
	if !found {
		t.Error("should return to origin at end")
	}
}

func TestLineCutString(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{0, 0}, {10, 0}}}}
	params := LineParams{FeedRate: 500, Power: 500, LaserMode: true}
	s := LineCutString(paths, params)

	if !strings.HasSuffix(s, "\n") {
		t.Error("should end with newline")
	}
	if !strings.Contains(s, "M2") {
		t.Error("missing M2")
	}
}

func TestOptimizePathsOrder(t *testing.T) {
	paths := []InputPath{
		{Points: []Point2D{{100, 100}, {110, 100}}},
		{Points: []Point2D{{10, 10}, {20, 10}}},
		{Points: []Point2D{{50, 50}, {60, 50}}},
	}
	result := OptimizePaths(paths)

	if len(result) != 3 {
		t.Fatalf("got %d paths, want 3", len(result))
	}
	if result[0].Points[0].X != 10 {
		t.Errorf("first path should be nearest to origin, got start X=%f", result[0].Points[0].X)
	}
	if result[1].Points[0].X != 50 {
		t.Errorf("second path should be (50,50), got start X=%f", result[1].Points[0].X)
	}
}

func TestOptimizePathsSingle(t *testing.T) {
	paths := []InputPath{{Points: []Point2D{{5, 5}, {10, 10}}}}
	result := OptimizePaths(paths)
	if len(result) != 1 {
		t.Fatalf("got %d paths, want 1", len(result))
	}
}

func TestOptimizePathsEmpty(t *testing.T) {
	result := OptimizePaths(nil)
	if len(result) != 0 {
		t.Errorf("got %d paths, want 0", len(result))
	}
}

func TestOptimizePathsClosedRotation(t *testing.T) {
	paths := []InputPath{
		{
			Points: []Point2D{{100, 100}, {110, 100}, {110, 110}, {100, 110}},
			Closed: true,
		},
	}
	cur := Point2D{108, 100}
	best := nearestPointIndex(cur, paths[0].Points)

	if best != 1 {
		t.Errorf("nearest to (108,100) should be index 1, got %d", best)
	}

	rotated := rotatePath(paths[0], 1)
	if rotated.Points[0].X != 110 || rotated.Points[0].Y != 100 {
		t.Errorf("rotated start = (%f,%f), want (110,100)", rotated.Points[0].X, rotated.Points[0].Y)
	}
	if !rotated.Closed {
		t.Error("rotated path should still be closed")
	}
}

func TestDist2D(t *testing.T) {
	d := dist2D(Point2D{0, 0}, Point2D{3, 4})
	if !approx(d, 5.0) {
		t.Errorf("dist = %f, want 5.0", d)
	}
}
