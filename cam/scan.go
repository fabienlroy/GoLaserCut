package cam

import (
	"fmt"
	"strings"
)

// ScanParams configures raster engraving.
type ScanParams struct {
	FeedRate   float64 // engraving speed, mm/min
	MaxPower   float64 // max laser power S value (0–1000)
	LineSpace  float64 // line spacing, mm (e.g. 0.1)
	Overscan   float64 // overscan distance, mm (0 = auto from RampProfile)
	MaxAccel   float64 // max acceleration, mm/s² (for auto overscan)
	Bidirect   bool    // true = alternating direction per line
	LaserMode  bool    // true = M4, false = M3
	AirAssist  bool
}

// ScanLine is one horizontal raster line with per-pixel power values.
type ScanLine struct {
	Y      float64
	StartX float64
	Pixels []uint8 // power values 0–255, mapped to S0–SMaxPower
}

// RasterScan generates G-code for raster engraving a grayscale image.
// pixels is a row-major grayscale buffer (0=white/no burn, 255=black/full power).
// width/height are pixel dimensions, pixelMM is the size of one pixel in mm.
// originX/originY is the bottom-left corner position in machine coordinates.
func RasterScan(width, height int, pixels []uint8, pixelMM float64, originX, originY float64, params ScanParams) []string {
	if width <= 0 || height <= 0 || len(pixels) < width*height {
		return nil
	}

	lineSpace := params.LineSpace
	if lineSpace <= 0 {
		lineSpace = pixelMM
	}

	overscan := params.Overscan
	if overscan <= 0 && params.MaxAccel > 0 {
		ramp := NewRampProfile(params.FeedRate, params.MaxAccel)
		overscan = ramp.RampDistance
	}
	if overscan < 0 {
		overscan = 0
	}

	var scanLines []ScanLine
	yStep := lineSpace
	for row := height - 1; row >= 0; row-- {
		rowPixels := pixels[row*width : (row+1)*width]

		start, end := trimRow(rowPixels)
		if start < 0 {
			continue
		}

		y := originY + float64(height-1-row)*yStep
		sl := ScanLine{
			Y:      y,
			StartX: originX + float64(start)*pixelMM,
			Pixels: make([]uint8, end-start+1),
		}
		copy(sl.Pixels, rowPixels[start:end+1])
		scanLines = append(scanLines, sl)
	}

	return scanLinesToGCode(scanLines, pixelMM, overscan, params)
}

// trimRow finds the first and last non-zero pixel in a row.
// Returns -1, -1 if the row is entirely white.
func trimRow(row []uint8) (int, int) {
	start := -1
	end := -1
	for i, v := range row {
		if v > 0 {
			if start < 0 {
				start = i
			}
			end = i
		}
	}
	return start, end
}

func scanLinesToGCode(lines []ScanLine, pixelMM, overscan float64, params ScanParams) []string {
	if len(lines) == 0 {
		return nil
	}

	var gcode []string
	w := func(s string) { gcode = append(gcode, s) }

	w("; Raster scan")
	w("G90 G21")
	w("G17")
	if params.LaserMode {
		w("M4")
	} else {
		w("M3")
	}
	w("S0")
	if params.AirAssist {
		w("M7 ; air assist on")
	}

	leftToRight := true

	for _, sl := range lines {
		endX := sl.StartX + float64(len(sl.Pixels))*pixelMM

		if params.Bidirect && !leftToRight {
			w(fmt.Sprintf("G0 X%.3f Y%.3f", endX+overscan, sl.Y))
			w("S0")
			w(fmt.Sprintf("G1 X%.3f F%.0f", endX, params.FeedRate))

			for i := len(sl.Pixels) - 1; i >= 0; i-- {
				s := pixelToS(sl.Pixels[i], params.MaxPower)
				x := sl.StartX + float64(i)*pixelMM
				w(fmt.Sprintf("G1 X%.3f S%.0f", x, s))
			}

			w("S0")
			w(fmt.Sprintf("G1 X%.3f", sl.StartX-overscan))
		} else {
			w(fmt.Sprintf("G0 X%.3f Y%.3f", sl.StartX-overscan, sl.Y))
			w("S0")
			w(fmt.Sprintf("G1 X%.3f F%.0f", sl.StartX, params.FeedRate))

			for i, px := range sl.Pixels {
				s := pixelToS(px, params.MaxPower)
				x := sl.StartX + float64(i+1)*pixelMM
				w(fmt.Sprintf("G1 X%.3f S%.0f", x, s))
			}

			w("S0")
			w(fmt.Sprintf("G1 X%.3f", endX+overscan))
		}

		if params.Bidirect {
			leftToRight = !leftToRight
		}
	}

	if params.AirAssist {
		w("M9 ; air assist off")
	}
	w("S0")
	w("M5")
	w("G0 X0 Y0")
	w("M2")

	return gcode
}

func pixelToS(pixel uint8, maxPower float64) float64 {
	return float64(pixel) / 255.0 * maxPower
}

// RasterScanString returns the full G-code as a single string.
func RasterScanString(width, height int, pixels []uint8, pixelMM float64, originX, originY float64, params ScanParams) string {
	lines := RasterScan(width, height, pixels, pixelMM, originX, originY, params)
	if lines == nil {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
