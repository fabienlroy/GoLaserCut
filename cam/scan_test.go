package cam

import (
	"fmt"
	"strings"
	"testing"
)

func makeImage(w, h int, val uint8) []uint8 {
	pixels := make([]uint8, w*h)
	for i := range pixels {
		pixels[i] = val
	}
	return pixels
}

func TestRasterScanBasic(t *testing.T) {
	pixels := makeImage(4, 2, 255)
	params := ScanParams{
		FeedRate:  6000,
		MaxPower:  1000,
		LineSpace: 0.1,
		Overscan:  5,
		LaserMode: true,
	}
	gcode := RasterScan(4, 2, pixels, 0.1, 0, 0, params)
	if len(gcode) == 0 {
		t.Fatal("expected G-code output")
	}
	joined := strings.Join(gcode, "\n")

	checks := []string{"G90 G21", "M4", "S0", "M5", "M2"}
	for _, c := range checks {
		if !strings.Contains(joined, c) {
			t.Errorf("missing: %q", c)
		}
	}

	if !strings.Contains(joined, "S1000") {
		t.Error("full black pixels should produce S1000")
	}
}

func TestRasterScanWhiteRowsSkipped(t *testing.T) {
	// 4x3, middle row all black, top and bottom white
	pixels := make([]uint8, 12)
	for i := 4; i < 8; i++ {
		pixels[i] = 255
	}
	params := ScanParams{
		FeedRate: 3000, MaxPower: 1000, LineSpace: 0.1,
		Overscan: 2, LaserMode: true,
	}
	gcode := RasterScan(4, 3, pixels, 0.1, 0, 0, params)

	// Only one scan line Y coordinate (excluding the return G0 X0 Y0)
	scanYCount := 0
	for _, line := range gcode {
		if strings.HasPrefix(line, "G0 X") && strings.Contains(line, "Y") && line != "G0 X0 Y0" {
			scanYCount++
		}
	}
	if scanYCount != 1 {
		t.Errorf("expected 1 scan line rapid, got %d", scanYCount)
	}
}

func TestRasterScanOverscan(t *testing.T) {
	pixels := makeImage(10, 1, 128)
	params := ScanParams{
		FeedRate: 6000, MaxPower: 1000, LineSpace: 0.1,
		Overscan: 5, LaserMode: true,
	}
	gcode := RasterScan(10, 1, pixels, 0.1, 10, 0, params)
	joined := strings.Join(gcode, "\n")

	if !strings.Contains(joined, "X5.000") {
		t.Error("should rapid to startX - overscan = 10 - 5 = 5")
	}
	endX := 10 + 10*0.1 + 5 // 16.0
	expected := fmt.Sprintf("X%.3f", endX)
	if !strings.Contains(joined, expected) {
		t.Errorf("should overshoot to endX + overscan = %s", expected)
	}
}

func TestRasterScanAutoOverscan(t *testing.T) {
	pixels := makeImage(4, 1, 200)
	params := ScanParams{
		FeedRate: 6000, MaxPower: 1000, LineSpace: 0.1,
		Overscan: 0, MaxAccel: 1000, LaserMode: true,
	}
	gcode := RasterScan(4, 1, pixels, 0.1, 0, 0, params)
	if len(gcode) == 0 {
		t.Fatal("expected output with auto overscan")
	}
	joined := strings.Join(gcode, "\n")

	// Auto overscan: v²/a = (100)²/1000 = 10
	if !strings.Contains(joined, "X-10.000") {
		t.Error("auto overscan should produce X-10.000 for 6000mm/min, 1000mm/s²")
	}
}

func TestRasterScanBidirectional(t *testing.T) {
	pixels := makeImage(4, 3, 128)
	params := ScanParams{
		FeedRate: 3000, MaxPower: 1000, LineSpace: 0.1,
		Overscan: 2, Bidirect: true, LaserMode: true,
	}
	gcode := RasterScan(4, 3, pixels, 0.1, 0, 0, params)
	joined := strings.Join(gcode, "\n")

	// First line left-to-right, second right-to-left
	// In right-to-left, the rapid should go to endX + overscan
	if !strings.Contains(joined, "X-2.000") {
		t.Error("left-to-right lines should start at -2 (overscan)")
	}

	endX := 4*0.1 + 2 // 2.4
	expected := fmt.Sprintf("X%.3f", endX)
	if !strings.Contains(joined, expected) {
		t.Errorf("right-to-left line should rapid to %s", expected)
	}
}

func TestRasterScanAirAssist(t *testing.T) {
	pixels := makeImage(2, 1, 100)
	params := ScanParams{
		FeedRate: 3000, MaxPower: 1000, Overscan: 1,
		AirAssist: true, LaserMode: true,
	}
	gcode := RasterScan(2, 1, pixels, 0.1, 0, 0, params)
	joined := strings.Join(gcode, "\n")

	if !strings.Contains(joined, "M7") {
		t.Error("expected M7")
	}
	if !strings.Contains(joined, "M9") {
		t.Error("expected M9")
	}
}

func TestRasterScanEmptyImage(t *testing.T) {
	gcode := RasterScan(0, 0, nil, 0.1, 0, 0, ScanParams{})
	if gcode != nil {
		t.Error("empty image should return nil")
	}
}

func TestRasterScanAllWhite(t *testing.T) {
	pixels := makeImage(10, 10, 0)
	params := ScanParams{
		FeedRate: 3000, MaxPower: 1000, Overscan: 2, LaserMode: true,
	}
	gcode := RasterScan(10, 10, pixels, 0.1, 0, 0, params)
	if gcode != nil {
		t.Error("all-white image should produce nil (no scan lines)")
	}
}

func TestRasterScanPixelTrim(t *testing.T) {
	// Row with leading/trailing zeros: [0, 0, 128, 200, 0]
	pixels := []uint8{0, 0, 128, 200, 0}
	params := ScanParams{
		FeedRate: 3000, MaxPower: 1000, Overscan: 1, LaserMode: true,
	}
	gcode := RasterScan(5, 1, pixels, 0.1, 0, 0, params)
	joined := strings.Join(gcode, "\n")

	// Should start at pixel 2 (X=0.2) minus overscan
	if !strings.Contains(joined, "X-0.800") {
		t.Error("trimmed row should start at 0.2 - 1.0 = -0.8")
	}
}

func TestPixelToS(t *testing.T) {
	tests := []struct {
		pixel    uint8
		maxPower float64
		want     float64
	}{
		{0, 1000, 0},
		{255, 1000, 1000},
		{128, 1000, 128.0 / 255.0 * 1000},
		{255, 500, 500},
	}
	for _, tt := range tests {
		got := pixelToS(tt.pixel, tt.maxPower)
		if !approx(got, tt.want) {
			t.Errorf("pixelToS(%d, %f) = %f, want %f", tt.pixel, tt.maxPower, got, tt.want)
		}
	}
}

func TestTrimRow(t *testing.T) {
	tests := []struct {
		row        []uint8
		wantStart  int
		wantEnd    int
	}{
		{[]uint8{0, 0, 0}, -1, -1},
		{[]uint8{255, 0, 0}, 0, 0},
		{[]uint8{0, 0, 255}, 2, 2},
		{[]uint8{0, 128, 0, 200, 0}, 1, 3},
		{[]uint8{100, 200, 50}, 0, 2},
	}
	for _, tt := range tests {
		s, e := trimRow(tt.row)
		if s != tt.wantStart || e != tt.wantEnd {
			t.Errorf("trimRow(%v) = (%d,%d), want (%d,%d)", tt.row, s, e, tt.wantStart, tt.wantEnd)
		}
	}
}

func TestRasterScanString(t *testing.T) {
	pixels := makeImage(2, 2, 100)
	params := ScanParams{
		FeedRate: 3000, MaxPower: 1000, Overscan: 1, LaserMode: true,
	}
	s := RasterScanString(2, 2, pixels, 0.1, 0, 0, params)
	if !strings.HasSuffix(s, "\n") {
		t.Error("should end with newline")
	}
	if !strings.Contains(s, "M2") {
		t.Error("missing M2")
	}
}

func TestRasterScanStringEmpty(t *testing.T) {
	s := RasterScanString(0, 0, nil, 0.1, 0, 0, ScanParams{})
	if s != "" {
		t.Error("empty should return empty string")
	}
}
