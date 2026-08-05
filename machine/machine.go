package machine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Origin is the machine home/origin corner.
type Origin int

const (
	FrontLeft Origin = iota
	FrontRight
	RearLeft
	RearRight
)

func (o Origin) String() string {
	switch o {
	case FrontLeft:
		return "front-left"
	case FrontRight:
		return "front-right"
	case RearLeft:
		return "rear-left"
	case RearRight:
		return "rear-right"
	}
	return "unknown"
}

func ParseOrigin(s string) Origin {
	switch s {
	case "front-right":
		return FrontRight
	case "rear-left":
		return RearLeft
	case "rear-right":
		return RearRight
	default:
		return FrontLeft
	}
}

func (o Origin) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *Origin) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*o = ParseOrigin(s)
	return nil
}

// GasAssist describes the gas assist capability.
type GasAssist int

const (
	NoAssist  GasAssist = iota
	AirAssist           // compressed air
	CO2Assist           // CO₂ gas
	O2Assist            // oxygen (thin steel cutting)
)

func (g GasAssist) String() string {
	switch g {
	case AirAssist:
		return "air"
	case CO2Assist:
		return "co2"
	case O2Assist:
		return "o2"
	}
	return "none"
}

func ParseGasAssist(s string) GasAssist {
	switch s {
	case "air":
		return AirAssist
	case "co2":
		return CO2Assist
	case "o2":
		return O2Assist
	default:
		return NoAssist
	}
}

func (g GasAssist) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

func (g *GasAssist) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*g = ParseGasAssist(s)
	return nil
}

// MachineType classifies the machine.
type MachineType int

const (
	DiodeLaser MachineType = iota
	CO2Laser
	FiberLaser
	CNCRouter
)

func (m MachineType) String() string {
	switch m {
	case CO2Laser:
		return "co2-laser"
	case FiberLaser:
		return "fiber-laser"
	case CNCRouter:
		return "cnc-router"
	}
	return "diode-laser"
}

func ParseMachineType(s string) MachineType {
	switch s {
	case "co2-laser":
		return CO2Laser
	case "fiber-laser":
		return FiberLaser
	case "cnc-router":
		return CNCRouter
	default:
		return DiodeLaser
	}
}

func (m MachineType) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *MachineType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*m = ParseMachineType(s)
	return nil
}

// Machine describes a GRBL-compatible laser or CNC machine.
type Machine struct {
	Name     string      `json:"name"`
	Brand    string      `json:"brand"`
	Type     MachineType `json:"type"`
	Custom   bool        `json:"custom,omitempty"`

	WorkAreaX float64 `json:"work_area_x"` // mm
	WorkAreaY float64 `json:"work_area_y"` // mm
	WorkAreaZ float64 `json:"work_area_z"` // mm (0 for 2-axis)

	MaxFeedX float64 `json:"max_feed_x"` // mm/min
	MaxFeedY float64 `json:"max_feed_y"` // mm/min
	MaxFeedZ float64 `json:"max_feed_z"` // mm/min (0 for 2-axis)

	AccelX float64 `json:"accel_x"` // mm/s²
	AccelY float64 `json:"accel_y"` // mm/s²
	AccelZ float64 `json:"accel_z"` // mm/s² (0 for 2-axis)

	Origin    Origin    `json:"origin"`
	LaserMode bool      `json:"laser_mode"` // $32=1
	MaxPower  float64   `json:"max_power"`  // $30 value
	GasAssist GasAssist `json:"gas_assist"`
}

// Axes returns the number of axes (2 or 3).
func (m *Machine) Axes() int {
	if m.WorkAreaZ > 0 {
		return 3
	}
	return 2
}

// GRBLSettings returns key GRBL $$ values for this machine.
func (m *Machine) GRBLSettings() map[string]float64 {
	s := map[string]float64{
		"$30":  m.MaxPower,
		"$110": m.MaxFeedX,
		"$111": m.MaxFeedY,
		"$120": m.AccelX,
		"$121": m.AccelY,
	}
	if m.WorkAreaZ > 0 {
		s["$112"] = m.MaxFeedZ
		s["$122"] = m.AccelZ
	}
	if m.LaserMode {
		s["$32"] = 1
	} else {
		s["$32"] = 0
	}
	return s
}

// Library holds a collection of machine profiles.
type Library struct {
	Machines []Machine `json:"machines"`
}

// Find returns a machine by name, or nil if not found.
func (lib *Library) Find(name string) *Machine {
	for i := range lib.Machines {
		if lib.Machines[i].Name == name {
			return &lib.Machines[i]
		}
	}
	return nil
}

// Add adds a machine to the library.
func (lib *Library) Add(m Machine) {
	lib.Machines = append(lib.Machines, m)
}

// Remove removes a machine by name. Returns true if found.
func (lib *Library) Remove(name string) bool {
	for i, m := range lib.Machines {
		if m.Name == name {
			lib.Machines = append(lib.Machines[:i], lib.Machines[i+1:]...)
			return true
		}
	}
	return false
}

// Names returns all machine names.
func (lib *Library) Names() []string {
	names := make([]string, len(lib.Machines))
	for i, m := range lib.Machines {
		names[i] = m.Name
	}
	return names
}

// Save writes the library to a JSON file.
func (lib *Library) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling library: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads a library from a JSON file.
func Load(path string) (*Library, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var lib Library
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &lib, nil
}
