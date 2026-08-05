package material

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Material describes a laser-cuttable material with default parameters.
type Material struct {
	Name          string  `json:"name"`
	Category      string  `json:"category"`
	MinThickness  float64 `json:"min_thickness"`  // mm
	MaxThickness  float64 `json:"max_thickness"`  // mm
	DefaultFeed   float64 `json:"default_feed"`   // mm/min (starting estimate for ~10W diode)
	Notes         string  `json:"notes,omitempty"`
}

// CalibrationPoint is one power-vs-depth measurement.
type CalibrationPoint struct {
	PowerPct float64 `json:"power_pct"` // 0–100
	Depth    float64 `json:"depth"`     // mm (measured)
}

// CalibrationResult stores depth calibration data for a material.
type CalibrationResult struct {
	Material   string             `json:"material"`
	Thickness  float64            `json:"thickness"`
	FeedRate   float64            `json:"feed_rate"`
	PocketSize float64            `json:"pocket_size"`
	Points     []CalibrationPoint `json:"points"`
}

// DepthForPower returns interpolated depth for a given power percentage.
func (r *CalibrationResult) DepthForPower(pct float64) float64 {
	if len(r.Points) == 0 {
		return 0
	}
	if pct <= r.Points[0].PowerPct {
		return r.Points[0].Depth
	}
	last := r.Points[len(r.Points)-1]
	if pct >= last.PowerPct {
		return last.Depth
	}
	for i := 0; i < len(r.Points)-1; i++ {
		a, b := r.Points[i], r.Points[i+1]
		if pct >= a.PowerPct && pct <= b.PowerPct {
			t := (pct - a.PowerPct) / (b.PowerPct - a.PowerPct)
			return a.Depth + t*(b.Depth-a.Depth)
		}
	}
	return last.Depth
}

// PowerForDepth returns the power percentage needed for a target depth.
// Returns -1 if the depth exceeds the calibrated range.
func (r *CalibrationResult) PowerForDepth(depth float64) float64 {
	if len(r.Points) == 0 {
		return -1
	}
	if depth <= r.Points[0].Depth {
		return r.Points[0].PowerPct
	}
	last := r.Points[len(r.Points)-1]
	if depth >= last.Depth {
		if depth > last.Depth*1.05 {
			return -1
		}
		return last.PowerPct
	}
	for i := 0; i < len(r.Points)-1; i++ {
		a, b := r.Points[i], r.Points[i+1]
		if depth >= a.Depth && depth <= b.Depth {
			t := (depth - a.Depth) / (b.Depth - a.Depth)
			return a.PowerPct + t*(b.PowerPct-a.PowerPct)
		}
	}
	return -1
}

// MaxDepth returns the deepest calibrated depth.
func (r *CalibrationResult) MaxDepth() float64 {
	if len(r.Points) == 0 {
		return 0
	}
	max := r.Points[0].Depth
	for _, p := range r.Points[1:] {
		if p.Depth > max {
			max = p.Depth
		}
	}
	return max
}

// MaterialLibrary holds materials and their calibration data.
type MaterialLibrary struct {
	Materials    []Material          `json:"materials"`
	Calibrations []CalibrationResult `json:"calibrations,omitempty"`
}

// FindMaterial returns a material by name, or nil.
func (lib *MaterialLibrary) FindMaterial(name string) *Material {
	for i := range lib.Materials {
		if lib.Materials[i].Name == name {
			return &lib.Materials[i]
		}
	}
	return nil
}

// FindCalibration returns calibration data for a material/thickness/feed combo.
func (lib *MaterialLibrary) FindCalibration(material string, thickness, feedRate float64) *CalibrationResult {
	for i := range lib.Calibrations {
		c := &lib.Calibrations[i]
		if c.Material == material && c.Thickness == thickness && c.FeedRate == feedRate {
			return c
		}
	}
	return nil
}

// AddCalibration stores calibration results, replacing existing if present.
func (lib *MaterialLibrary) AddCalibration(result CalibrationResult) {
	for i, c := range lib.Calibrations {
		if c.Material == result.Material && c.Thickness == result.Thickness && c.FeedRate == result.FeedRate {
			lib.Calibrations[i] = result
			return
		}
	}
	lib.Calibrations = append(lib.Calibrations, result)
}

// MaterialNames returns all material names.
func (lib *MaterialLibrary) MaterialNames() []string {
	names := make([]string, len(lib.Materials))
	for i, m := range lib.Materials {
		names[i] = m.Name
	}
	return names
}

// Save writes the library to a JSON file.
func (lib *MaterialLibrary) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadMaterialLibrary reads a material library from JSON.
func LoadMaterialLibrary(path string) (*MaterialLibrary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var lib MaterialLibrary
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &lib, nil
}
