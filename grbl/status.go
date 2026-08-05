package grbl

type MachineState int

const (
	StateIdle    MachineState = iota
	StateRun
	StateHold
	StateJog
	StateAlarm
	StateDoor
	StateCheck
	StateHome
	StateSleep
	StateUnknown
)

func (s MachineState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateRun:
		return "Run"
	case StateHold:
		return "Hold"
	case StateJog:
		return "Jog"
	case StateAlarm:
		return "Alarm"
	case StateDoor:
		return "Door"
	case StateCheck:
		return "Check"
	case StateHome:
		return "Home"
	case StateSleep:
		return "Sleep"
	default:
		return "Unknown"
	}
}

type Position struct {
	X, Y, Z float64
}

type Overrides struct {
	Feed    int
	Rapids  int
	Spindle int
}

type BufferState struct {
	PlannerBlocks int
	SerialBytes   int
}

type Status struct {
	State     MachineState
	SubState  int
	MPos      *Position
	WPos      *Position
	Feed      float64
	Spindle   float64
	Overrides *Overrides
	WCO       *Position
	Buffer    *BufferState
	LineNum   int
	Pins      string
	Accessory string
}

func (*Status) grblResponse() {}
