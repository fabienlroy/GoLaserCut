package machine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinLibrary(t *testing.T) {
	lib := BuiltinLibrary()
	if len(lib.Machines) == 0 {
		t.Fatal("builtin library is empty")
	}

	names := lib.Names()
	if len(names) != len(lib.Machines) {
		t.Errorf("Names() length %d != Machines length %d", len(names), len(lib.Machines))
	}
}

func TestBuiltinLibraryNoDuplicateNames(t *testing.T) {
	lib := BuiltinLibrary()
	seen := map[string]bool{}
	for _, m := range lib.Machines {
		if seen[m.Name] {
			t.Errorf("duplicate machine name: %q", m.Name)
		}
		seen[m.Name] = true
	}
}

func TestBuiltinLibraryValidSpecs(t *testing.T) {
	lib := BuiltinLibrary()
	for _, m := range lib.Machines {
		if m.WorkAreaX <= 0 || m.WorkAreaY <= 0 {
			t.Errorf("%s: invalid work area %.0fx%.0f", m.Name, m.WorkAreaX, m.WorkAreaY)
		}
		if m.MaxFeedX <= 0 || m.MaxFeedY <= 0 {
			t.Errorf("%s: invalid max feed", m.Name)
		}
		if m.AccelX <= 0 || m.AccelY <= 0 {
			t.Errorf("%s: invalid acceleration", m.Name)
		}
		if m.MaxPower <= 0 {
			t.Errorf("%s: invalid max power", m.Name)
		}
		if m.Brand == "" {
			t.Errorf("%s: missing brand", m.Name)
		}
	}
}

func TestFindMachine(t *testing.T) {
	lib := BuiltinLibrary()
	m := lib.Find("FoxAlien XE-Pro")
	if m == nil {
		t.Fatal("FoxAlien XE-Pro not found")
	}
	if m.Brand != "FoxAlien" {
		t.Errorf("brand = %q, want FoxAlien", m.Brand)
	}
	if m.GasAssist != CO2Assist {
		t.Errorf("gas assist = %v, want CO2", m.GasAssist)
	}
}

func TestFindMachineNotFound(t *testing.T) {
	lib := BuiltinLibrary()
	if lib.Find("NonExistent") != nil {
		t.Error("should return nil for unknown machine")
	}
}

func TestMachineAxes(t *testing.T) {
	lib := BuiltinLibrary()

	diode := lib.Find("Sculpfun S9")
	if diode == nil {
		t.Fatal("Sculpfun S9 not found")
	}
	if diode.Axes() != 2 {
		t.Errorf("diode axes = %d, want 2", diode.Axes())
	}

	cnc := lib.Find("FoxAlien Masuter Pro")
	if cnc == nil {
		t.Fatal("FoxAlien Masuter Pro not found")
	}
	if cnc.Axes() != 3 {
		t.Errorf("CNC axes = %d, want 3", cnc.Axes())
	}
}

func TestGRBLSettings(t *testing.T) {
	m := Machine{
		WorkAreaX: 400, WorkAreaY: 400, WorkAreaZ: 65,
		MaxFeedX: 2000, MaxFeedY: 2000, MaxFeedZ: 1200,
		AccelX: 300, AccelY: 300, AccelZ: 30,
		LaserMode: false, MaxPower: 10000,
	}
	s := m.GRBLSettings()

	if s["$30"] != 10000 {
		t.Errorf("$30 = %f, want 10000", s["$30"])
	}
	if s["$110"] != 2000 {
		t.Errorf("$110 = %f, want 2000", s["$110"])
	}
	if s["$112"] != 1200 {
		t.Errorf("$112 = %f, want 1200", s["$112"])
	}
	if s["$32"] != 0 {
		t.Errorf("$32 = %f, want 0", s["$32"])
	}
}

func TestGRBLSettings2Axis(t *testing.T) {
	m := Machine{
		MaxFeedX: 6000, MaxFeedY: 6000,
		AccelX: 1000, AccelY: 1000,
		LaserMode: true, MaxPower: 1000,
	}
	s := m.GRBLSettings()

	if _, ok := s["$112"]; ok {
		t.Error("2-axis machine should not have $112")
	}
	if s["$32"] != 1 {
		t.Errorf("$32 = %f, want 1", s["$32"])
	}
}

func TestAddRemoveMachine(t *testing.T) {
	lib := &Library{}
	lib.Add(Machine{Name: "Custom 1", Brand: "Me", Custom: true})
	lib.Add(Machine{Name: "Custom 2", Brand: "Me", Custom: true})

	if len(lib.Machines) != 2 {
		t.Fatalf("got %d machines, want 2", len(lib.Machines))
	}

	if !lib.Remove("Custom 1") {
		t.Error("Remove should return true")
	}
	if len(lib.Machines) != 1 {
		t.Fatalf("got %d machines after remove, want 1", len(lib.Machines))
	}
	if lib.Machines[0].Name != "Custom 2" {
		t.Errorf("remaining machine = %q, want Custom 2", lib.Machines[0].Name)
	}

	if lib.Remove("NonExistent") {
		t.Error("Remove should return false for unknown")
	}
}

func TestSaveLoad(t *testing.T) {
	lib := BuiltinLibrary()
	lib.Add(Machine{
		Name: "My Custom Laser", Brand: "DIY", Type: DiodeLaser,
		WorkAreaX: 500, WorkAreaY: 300,
		MaxFeedX: 10000, MaxFeedY: 10000,
		AccelX: 2000, AccelY: 2000,
		Origin: FrontLeft, LaserMode: true, MaxPower: 1000,
		GasAssist: AirAssist, Custom: true,
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "machines.json")

	if err := lib.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Machines) != len(lib.Machines) {
		t.Fatalf("loaded %d machines, want %d", len(loaded.Machines), len(lib.Machines))
	}

	custom := loaded.Find("My Custom Laser")
	if custom == nil {
		t.Fatal("custom machine not found after load")
	}
	if custom.GasAssist != AirAssist {
		t.Errorf("gas assist = %v, want air", custom.GasAssist)
	}
	if !custom.Custom {
		t.Error("custom flag should be true")
	}
}

func TestSaveCreatesDir(t *testing.T) {
	lib := &Library{Machines: []Machine{{Name: "Test"}}}
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "machines.json")

	if err := lib.Save(path); err != nil {
		t.Fatalf("Save with nested dir: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file should exist: %v", err)
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := Load("/nonexistent/machines.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestOriginJSON(t *testing.T) {
	tests := []struct {
		origin Origin
		json   string
	}{
		{FrontLeft, `"front-left"`},
		{FrontRight, `"front-right"`},
		{RearLeft, `"rear-left"`},
		{RearRight, `"rear-right"`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.origin)
		if err != nil {
			t.Fatalf("marshal %v: %v", tt.origin, err)
		}
		if string(data) != tt.json {
			t.Errorf("marshal %v = %s, want %s", tt.origin, data, tt.json)
		}

		var got Origin
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %s: %v", data, err)
		}
		if got != tt.origin {
			t.Errorf("unmarshal %s = %v, want %v", data, got, tt.origin)
		}
	}
}

func TestGasAssistJSON(t *testing.T) {
	tests := []struct {
		gas  GasAssist
		json string
	}{
		{NoAssist, `"none"`},
		{AirAssist, `"air"`},
		{CO2Assist, `"co2"`},
		{O2Assist, `"o2"`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.gas)
		if err != nil {
			t.Fatalf("marshal %v: %v", tt.gas, err)
		}
		if string(data) != tt.json {
			t.Errorf("marshal %v = %s, want %s", tt.gas, data, tt.json)
		}

		var got GasAssist
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %s: %v", data, err)
		}
		if got != tt.gas {
			t.Errorf("unmarshal %s = %v, want %v", data, got, tt.gas)
		}
	}
}

func TestMachineTypeJSON(t *testing.T) {
	tests := []struct {
		mt   MachineType
		json string
	}{
		{DiodeLaser, `"diode-laser"`},
		{CO2Laser, `"co2-laser"`},
		{FiberLaser, `"fiber-laser"`},
		{CNCRouter, `"cnc-router"`},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt.mt)
		if err != nil {
			t.Fatalf("marshal %v: %v", tt.mt, err)
		}
		if string(data) != tt.json {
			t.Errorf("marshal %v = %s, want %s", tt.mt, data, tt.json)
		}

		var got MachineType
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("unmarshal %s: %v", data, err)
		}
		if got != tt.mt {
			t.Errorf("unmarshal %s = %v, want %v", data, got, tt.mt)
		}
	}
}

func TestGasAssistTypes(t *testing.T) {
	lib := BuiltinLibrary()

	hasAir := false
	hasCO2 := false
	hasO2 := false
	hasNone := false

	for _, m := range lib.Machines {
		switch m.GasAssist {
		case AirAssist:
			hasAir = true
		case CO2Assist:
			hasCO2 = true
		case O2Assist:
			hasO2 = true
		case NoAssist:
			hasNone = true
		}
	}

	if !hasAir {
		t.Error("library should contain at least one machine with air assist")
	}
	if !hasCO2 {
		t.Error("library should contain at least one machine with CO2 assist")
	}
	if !hasO2 {
		t.Error("library should contain at least one machine with O2 assist")
	}
	if !hasNone {
		t.Error("library should contain machines without assist")
	}
}

func TestMachineTypes(t *testing.T) {
	lib := BuiltinLibrary()

	types := map[MachineType]bool{}
	for _, m := range lib.Machines {
		types[m.Type] = true
	}

	for _, mt := range []MachineType{DiodeLaser, CO2Laser, FiberLaser, CNCRouter} {
		if !types[mt] {
			t.Errorf("library missing machine type: %v", mt)
		}
	}
}

func TestOriginTypes(t *testing.T) {
	lib := BuiltinLibrary()

	origins := map[Origin]bool{}
	for _, m := range lib.Machines {
		origins[m.Origin] = true
	}

	if !origins[FrontLeft] {
		t.Error("library should have front-left origin machines")
	}
	if !origins[RearLeft] {
		t.Error("library should have rear-left origin machines (CO2)")
	}
}
