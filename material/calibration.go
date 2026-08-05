package material

import (
	"fmt"
	"strings"
)

// CalibrationParams configures the depth calibration pattern.
type CalibrationParams struct {
	PocketSize  float64 // pocket width/height, mm (5–20, default 10)
	PocketCount int     // number of pockets (default 10)
	MinPowerPct float64 // minimum power percentage (default 10)
	MaxPowerPct float64 // maximum power percentage (default 90)
	FeedRate    float64 // engraving feed rate, mm/min
	LineSpace   float64 // raster line spacing, mm (default 0.1)
	Margin      float64 // gap between pockets, mm (default 3)
	MaxSPower   float64 // machine $30 value (default 1000)
	Passes      int     // passes per pocket (default 1)
}

func (p *CalibrationParams) defaults() {
	if p.PocketSize < 5 {
		p.PocketSize = 10
	}
	if p.PocketSize > 20 {
		p.PocketSize = 20
	}
	if p.PocketCount <= 0 {
		p.PocketCount = 10
	}
	if p.MinPowerPct <= 0 {
		p.MinPowerPct = 10
	}
	if p.MaxPowerPct <= 0 {
		p.MaxPowerPct = 90
	}
	if p.MaxPowerPct <= p.MinPowerPct {
		p.MaxPowerPct = p.MinPowerPct + 10
	}
	if p.FeedRate <= 0 {
		p.FeedRate = 1000
	}
	if p.LineSpace <= 0 {
		p.LineSpace = 0.1
	}
	if p.Margin <= 0 {
		p.Margin = 3
	}
	if p.MaxSPower <= 0 {
		p.MaxSPower = 1000
	}
	if p.Passes <= 0 {
		p.Passes = 1
	}
}

// PowerLevels returns the power percentages for each pocket.
func PowerLevels(count int, minPct, maxPct float64) []float64 {
	if count <= 1 {
		return []float64{minPct}
	}
	levels := make([]float64, count)
	step := (maxPct - minPct) / float64(count-1)
	for i := range levels {
		levels[i] = minPct + float64(i)*step
	}
	return levels
}

// PatternWidth returns the total width of the calibration pattern in mm.
func PatternWidth(params CalibrationParams) float64 {
	params.defaults()
	return float64(params.PocketCount)*params.PocketSize + float64(params.PocketCount-1)*params.Margin
}

// GenerateCalibrationGCode creates G-code for the depth calibration pattern.
// Returns the G-code string and the power percentages used per pocket.
func GenerateCalibrationGCode(materialName string, params CalibrationParams) (string, []float64) {
	params.defaults()

	levels := PowerLevels(params.PocketCount, params.MinPowerPct, params.MaxPowerPct)

	var lines []string
	w := func(s string) { lines = append(lines, s) }

	w(fmt.Sprintf("; Depth calibration pattern: %s", materialName))
	w(fmt.Sprintf("; %d pockets, %.0f%% to %.0f%% power", params.PocketCount, params.MinPowerPct, params.MaxPowerPct))
	w(fmt.Sprintf("; Pocket size: %.0fx%.0fmm, feed: %.0fmm/min, passes: %d",
		params.PocketSize, params.PocketSize, params.FeedRate, params.Passes))
	w(fmt.Sprintf("; Line spacing: %.2fmm", params.LineSpace))
	w("")
	w("G90 G21")
	w("G17")
	w("M4")
	w("S0")

	for i, pct := range levels {
		sValue := pct / 100.0 * params.MaxSPower
		x0 := float64(i) * (params.PocketSize + params.Margin)
		y0 := 0.0

		w("")
		w(fmt.Sprintf("; Pocket %d: %.1f%% power (S%.0f)", i+1, pct, sValue))

		for pass := 0; pass < params.Passes; pass++ {
			if params.Passes > 1 {
				w(fmt.Sprintf("; pass %d/%d", pass+1, params.Passes))
			}
			fillPocket(&lines, x0, y0, params.PocketSize, params.LineSpace, params.FeedRate, sValue)
		}
	}

	w("")
	w("S0")
	w("M5")
	w("G0 X0 Y0")
	w("M2")

	return strings.Join(lines, "\n") + "\n", levels
}

func fillPocket(lines *[]string, x0, y0, size, lineSpace, feedRate, sValue float64) {
	w := func(s string) { *lines = append(*lines, s) }

	leftToRight := true
	numLines := int(size/lineSpace) + 1

	for i := 0; i < numLines; i++ {
		y := y0 + float64(i)*lineSpace
		if y > y0+size {
			y = y0 + size
		}

		if leftToRight {
			w(fmt.Sprintf("G0 X%.3f Y%.3f", x0, y))
			w(fmt.Sprintf("G1 X%.3f S%.0f F%.0f", x0+size, sValue, feedRate))
		} else {
			w(fmt.Sprintf("G0 X%.3f Y%.3f", x0+size, y))
			w(fmt.Sprintf("G1 X%.3f S%.0f F%.0f", x0, sValue, feedRate))
		}
		w("S0")

		leftToRight = !leftToRight
	}
}

// NewCalibrationResult creates a CalibrationResult from pocket measurements.
// depths must have the same length as powerLevels.
func NewCalibrationResult(materialName string, thickness, feedRate, pocketSize float64,
	powerLevels []float64, depths []float64) (*CalibrationResult, error) {
	if len(powerLevels) != len(depths) {
		return nil, fmt.Errorf("power levels (%d) and depths (%d) must have same length",
			len(powerLevels), len(depths))
	}
	points := make([]CalibrationPoint, len(powerLevels))
	for i := range powerLevels {
		if depths[i] < 0 {
			return nil, fmt.Errorf("depth at pocket %d cannot be negative", i+1)
		}
		points[i] = CalibrationPoint{
			PowerPct: powerLevels[i],
			Depth:    depths[i],
		}
	}
	return &CalibrationResult{
		Material:   materialName,
		Thickness:  thickness,
		FeedRate:   feedRate,
		PocketSize: pocketSize,
		Points:     points,
	}, nil
}
