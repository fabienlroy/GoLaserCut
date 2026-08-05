package material

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func approx(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

// --- Builtin library tests ---

func TestBuiltinMaterialLibrary(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	if len(lib.Materials) == 0 {
		t.Fatal("library is empty")
	}
}

func TestBuiltinNoDuplicateNames(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	seen := map[string]bool{}
	for _, m := range lib.Materials {
		if seen[m.Name] {
			t.Errorf("duplicate: %q", m.Name)
		}
		seen[m.Name] = true
	}
}

func TestBuiltinValidSpecs(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	for _, m := range lib.Materials {
		if m.Name == "" {
			t.Error("empty name")
		}
		if m.Category == "" {
			t.Errorf("%s: empty category", m.Name)
		}
		if m.DefaultFeed <= 0 {
			t.Errorf("%s: invalid default feed", m.Name)
		}
		if m.MinThickness <= 0 {
			t.Errorf("%s: invalid min thickness", m.Name)
		}
		if m.MaxThickness < m.MinThickness {
			t.Errorf("%s: max thickness < min thickness", m.Name)
		}
	}
}

func TestBuiltinCategories(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	cats := map[string]bool{}
	for _, m := range lib.Materials {
		cats[m.Category] = true
	}
	for _, want := range []string{"hardwood", "softwood", "plywood", "leather", "metal"} {
		if !cats[want] {
			t.Errorf("missing category: %s", want)
		}
	}
}

func TestFindMaterial(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	m := lib.FindMaterial("Mahogany")
	if m == nil {
		t.Fatal("Mahogany not found")
	}
	if m.Category != "hardwood" {
		t.Errorf("category = %q, want hardwood", m.Category)
	}
	if m.MinThickness < 3 {
		t.Errorf("min thickness = %f, want >= 3", m.MinThickness)
	}
}

func TestFindMaterialNotFound(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	if lib.FindMaterial("Unobtainium") != nil {
		t.Error("should return nil")
	}
}

func TestMaterialNames(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	names := lib.MaterialNames()
	if len(names) != len(lib.Materials) {
		t.Errorf("names count %d != materials %d", len(names), len(lib.Materials))
	}
}

// --- Calibration pattern tests ---

func TestPowerLevels(t *testing.T) {
	levels := PowerLevels(10, 10, 90)
	if len(levels) != 10 {
		t.Fatalf("got %d levels, want 10", len(levels))
	}
	if !approx(levels[0], 10) {
		t.Errorf("first = %f, want 10", levels[0])
	}
	if !approx(levels[9], 90) {
		t.Errorf("last = %f, want 90", levels[9])
	}
	for i := 1; i < len(levels); i++ {
		if levels[i] <= levels[i-1] {
			t.Errorf("not monotonic at index %d", i)
		}
	}
}

func TestPowerLevelsSingle(t *testing.T) {
	levels := PowerLevels(1, 50, 50)
	if len(levels) != 1 || !approx(levels[0], 50) {
		t.Errorf("single level = %v", levels)
	}
}

func TestPatternWidth(t *testing.T) {
	params := CalibrationParams{
		PocketSize:  10,
		PocketCount: 10,
		Margin:      3,
	}
	w := PatternWidth(params)
	// 10 pockets * 10mm + 9 gaps * 3mm = 127mm
	if !approx(w, 127) {
		t.Errorf("width = %f, want 127", w)
	}
}

func TestGenerateCalibrationGCode(t *testing.T) {
	params := CalibrationParams{
		PocketSize:  10,
		PocketCount: 5,
		MinPowerPct: 20,
		MaxPowerPct: 100,
		FeedRate:    1000,
		LineSpace:   0.5,
		MaxSPower:   1000,
	}
	gcode, levels := GenerateCalibrationGCode("Mahogany", params)

	if len(levels) != 5 {
		t.Fatalf("got %d levels, want 5", len(levels))
	}
	if !approx(levels[0], 20) || !approx(levels[4], 100) {
		t.Errorf("levels range: %v", levels)
	}

	if !strings.Contains(gcode, "Mahogany") {
		t.Error("missing material name in header")
	}
	if !strings.Contains(gcode, "G90 G21") {
		t.Error("missing preamble")
	}
	if !strings.Contains(gcode, "M4") {
		t.Error("missing M4")
	}
	if !strings.Contains(gcode, "M2") {
		t.Error("missing M2")
	}
	if !strings.Contains(gcode, "Pocket 1") {
		t.Error("missing pocket 1 comment")
	}
	if !strings.Contains(gcode, "Pocket 5") {
		t.Error("missing pocket 5 comment")
	}

	// Check that S values increase
	if !strings.Contains(gcode, "S200") {
		t.Error("expected S200 for 20% of 1000")
	}
	if !strings.Contains(gcode, "S1000") {
		t.Error("expected S1000 for 100% of 1000")
	}
}

func TestGenerateCalibrationGCodeDefaults(t *testing.T) {
	params := CalibrationParams{FeedRate: 800}
	gcode, levels := GenerateCalibrationGCode("Oak", params)

	if len(levels) != 10 {
		t.Errorf("default count: got %d, want 10", len(levels))
	}
	if gcode == "" {
		t.Error("empty output")
	}
}

func TestGenerateCalibrationGCodeMultiPass(t *testing.T) {
	params := CalibrationParams{
		PocketSize:  5,
		PocketCount: 3,
		MinPowerPct: 30,
		MaxPowerPct: 90,
		FeedRate:    500,
		Passes:      2,
	}
	gcode, _ := GenerateCalibrationGCode("Pine", params)

	if !strings.Contains(gcode, "pass 1/2") {
		t.Error("missing pass 1/2")
	}
	if !strings.Contains(gcode, "pass 2/2") {
		t.Error("missing pass 2/2")
	}
}

func TestGenerateCalibrationGCodePocketSize(t *testing.T) {
	for _, size := range []float64{5, 10, 15, 20} {
		params := CalibrationParams{
			PocketSize:  size,
			PocketCount: 3,
			FeedRate:    1000,
		}
		gcode, _ := GenerateCalibrationGCode("Test", params)
		if gcode == "" {
			t.Errorf("empty output for size %f", size)
		}
	}
}

func TestGenerateCalibrationGCodeClampSize(t *testing.T) {
	params := CalibrationParams{PocketSize: 2, FeedRate: 1000} // below minimum 5
	_, levels := GenerateCalibrationGCode("Test", params)
	if len(levels) == 0 {
		t.Error("should still produce output with clamped size")
	}
}

// --- Interpolation tests ---

func TestDepthForPower(t *testing.T) {
	r := &CalibrationResult{
		Points: []CalibrationPoint{
			{10, 0.1},
			{30, 0.5},
			{50, 1.2},
			{70, 2.0},
			{90, 3.5},
		},
	}

	// Exact points
	if !approx(r.DepthForPower(10), 0.1) {
		t.Errorf("at 10%%: %f", r.DepthForPower(10))
	}
	if !approx(r.DepthForPower(90), 3.5) {
		t.Errorf("at 90%%: %f", r.DepthForPower(90))
	}

	// Interpolated
	d50 := r.DepthForPower(50)
	if !approx(d50, 1.2) {
		t.Errorf("at 50%%: %f, want 1.2", d50)
	}

	// Between 30 and 50: linear interpolation
	d40 := r.DepthForPower(40)
	if d40 <= 0.5 || d40 >= 1.2 {
		t.Errorf("at 40%%: %f, should be between 0.5 and 1.2", d40)
	}
	if !approx(d40, 0.85) { // (0.5+1.2)/2
		t.Errorf("at 40%%: %f, want ~0.85", d40)
	}

	// Below range → clamp
	if !approx(r.DepthForPower(5), 0.1) {
		t.Errorf("below range: %f", r.DepthForPower(5))
	}

	// Above range → clamp
	if !approx(r.DepthForPower(100), 3.5) {
		t.Errorf("above range: %f", r.DepthForPower(100))
	}
}

func TestPowerForDepth(t *testing.T) {
	r := &CalibrationResult{
		Points: []CalibrationPoint{
			{10, 0.1},
			{30, 0.5},
			{50, 1.2},
			{70, 2.0},
			{90, 3.5},
		},
	}

	// Exact
	if !approx(r.PowerForDepth(0.1), 10) {
		t.Errorf("depth 0.1: %f", r.PowerForDepth(0.1))
	}

	// Interpolated
	p := r.PowerForDepth(0.85)
	if !approx(p, 40) {
		t.Errorf("depth 0.85: %f, want ~40", p)
	}

	// Beyond max → -1
	p = r.PowerForDepth(5.0)
	if p != -1 {
		t.Errorf("beyond max: %f, want -1", p)
	}

	// At max within tolerance
	p = r.PowerForDepth(3.5)
	if !approx(p, 90) {
		t.Errorf("at max: %f, want 90", p)
	}

	// Below min
	if !approx(r.PowerForDepth(0.05), 10) {
		t.Errorf("below min: %f", r.PowerForDepth(0.05))
	}
}

func TestPowerForDepthEmpty(t *testing.T) {
	r := &CalibrationResult{}
	if r.PowerForDepth(1.0) != -1 {
		t.Error("empty should return -1")
	}
}

func TestMaxDepth(t *testing.T) {
	r := &CalibrationResult{
		Points: []CalibrationPoint{{10, 0.1}, {50, 2.0}, {90, 3.5}},
	}
	if !approx(r.MaxDepth(), 3.5) {
		t.Errorf("max depth = %f, want 3.5", r.MaxDepth())
	}
}

func TestMaxDepthEmpty(t *testing.T) {
	r := &CalibrationResult{}
	if r.MaxDepth() != 0 {
		t.Error("empty max depth should be 0")
	}
}

// --- NewCalibrationResult tests ---

func TestNewCalibrationResult(t *testing.T) {
	levels := []float64{10, 30, 50, 70, 90}
	depths := []float64{0.1, 0.5, 1.2, 2.0, 3.5}

	result, err := NewCalibrationResult("Mahogany", 5, 800, 10, levels, depths)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if result.Material != "Mahogany" {
		t.Errorf("material = %q", result.Material)
	}
	if len(result.Points) != 5 {
		t.Errorf("points = %d, want 5", len(result.Points))
	}
	if !approx(result.DepthForPower(50), 1.2) {
		t.Errorf("depth at 50%%: %f", result.DepthForPower(50))
	}
}

func TestNewCalibrationResultMismatchLength(t *testing.T) {
	_, err := NewCalibrationResult("X", 5, 800, 10, []float64{10, 20}, []float64{0.1})
	if err == nil {
		t.Error("expected error for mismatched lengths")
	}
}

func TestNewCalibrationResultNegativeDepth(t *testing.T) {
	_, err := NewCalibrationResult("X", 5, 800, 10, []float64{10}, []float64{-1})
	if err == nil {
		t.Error("expected error for negative depth")
	}
}

// --- Calibration storage tests ---

func TestAddFindCalibration(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	cal := CalibrationResult{
		Material:   "Mahogany",
		Thickness:  5,
		FeedRate:   800,
		PocketSize: 10,
		Points:     []CalibrationPoint{{10, 0.1}, {50, 1.0}, {90, 2.5}},
	}
	lib.AddCalibration(cal)

	found := lib.FindCalibration("Mahogany", 5, 800)
	if found == nil {
		t.Fatal("calibration not found")
	}
	if len(found.Points) != 3 {
		t.Errorf("points = %d, want 3", len(found.Points))
	}
}

func TestAddCalibrationReplace(t *testing.T) {
	lib := &MaterialLibrary{}
	lib.AddCalibration(CalibrationResult{
		Material: "Oak", Thickness: 5, FeedRate: 700,
		Points: []CalibrationPoint{{50, 1.0}},
	})
	lib.AddCalibration(CalibrationResult{
		Material: "Oak", Thickness: 5, FeedRate: 700,
		Points: []CalibrationPoint{{50, 1.5}, {90, 3.0}},
	})
	if len(lib.Calibrations) != 1 {
		t.Errorf("should replace, got %d calibrations", len(lib.Calibrations))
	}
	if len(lib.Calibrations[0].Points) != 2 {
		t.Error("should have updated points")
	}
}

func TestFindCalibrationNotFound(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	if lib.FindCalibration("Mahogany", 5, 800) != nil {
		t.Error("should return nil without calibration data")
	}
}

// --- Save/Load tests ---

func TestSaveLoadMaterialLibrary(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	lib.AddCalibration(CalibrationResult{
		Material: "Mahogany", Thickness: 5, FeedRate: 800, PocketSize: 10,
		Points: []CalibrationPoint{{10, 0.1}, {50, 1.0}, {90, 2.5}},
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "materials.json")

	if err := lib.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadMaterialLibrary(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Materials) != len(lib.Materials) {
		t.Errorf("materials: %d vs %d", len(loaded.Materials), len(lib.Materials))
	}

	cal := loaded.FindCalibration("Mahogany", 5, 800)
	if cal == nil {
		t.Fatal("calibration not found after load")
	}
	if !approx(cal.DepthForPower(50), 1.0) {
		t.Errorf("loaded depth at 50%%: %f", cal.DepthForPower(50))
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := LoadMaterialLibrary("/nonexistent/materials.json")
	if err == nil {
		t.Error("expected error")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("{invalid"), 0644)

	_, err := LoadMaterialLibrary(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// --- Integration: calibrate-then-use workflow ---

func TestCalibrationWorkflow(t *testing.T) {
	lib := BuiltinMaterialLibrary()
	mat := lib.FindMaterial("Mahogany")
	if mat == nil {
		t.Fatal("Mahogany not found")
	}

	params := CalibrationParams{
		PocketSize:  10,
		PocketCount: 5,
		MinPowerPct: 10,
		MaxPowerPct: 90,
		FeedRate:    mat.DefaultFeed,
	}

	gcode, levels := GenerateCalibrationGCode(mat.Name, params)
	if gcode == "" {
		t.Fatal("empty G-code")
	}
	if len(levels) != 5 {
		t.Fatalf("got %d levels", len(levels))
	}

	// Simulate user entering measured depths
	depths := []float64{0.1, 0.4, 0.9, 1.8, 3.2}
	result, err := NewCalibrationResult(mat.Name, 5, mat.DefaultFeed, params.PocketSize, levels, depths)
	if err != nil {
		t.Fatal(err)
	}

	lib.AddCalibration(*result)

	// Now use calibration to find power for 1mm depth
	cal := lib.FindCalibration("Mahogany", 5, mat.DefaultFeed)
	power := cal.PowerForDepth(1.0)
	if power < 10 || power > 90 {
		t.Errorf("power for 1mm = %f, expected between 10 and 90", power)
	}

	// Should be between 50% and 70% (between 0.9 and 1.8)
	if power < 50 || power > 70 {
		t.Errorf("power for 1mm = %f, expected ~53", power)
	}
}
