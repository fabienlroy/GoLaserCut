package grbl

import (
	"math"
	"testing"
)

func TestParseOK(t *testing.T) {
	r := Parse("ok")
	if _, ok := r.(*OK); !ok {
		t.Fatalf("expected *OK, got %T", r)
	}
	if !IsAck(r) {
		t.Error("OK should be an ack")
	}
}

func TestParseError(t *testing.T) {
	r := Parse("error:22")
	e, ok := r.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", r)
	}
	if e.Code != 22 {
		t.Errorf("Code = %d, want 22", e.Code)
	}
	if e.Message != "Feed rate not set" {
		t.Errorf("Message = %q, want 'Feed rate not set'", e.Message)
	}
	if !IsAck(r) {
		t.Error("Error should be an ack")
	}
	if e.Error() != "error:22 (Feed rate not set)" {
		t.Errorf("Error() = %q", e.Error())
	}
}

func TestParseErrorUnknownCode(t *testing.T) {
	e := Parse("error:99").(*Error)
	if e.Code != 99 {
		t.Errorf("Code = %d, want 99", e.Code)
	}
	if e.Message != "" {
		t.Errorf("Message = %q, want empty", e.Message)
	}
}

func TestParseAlarm(t *testing.T) {
	r := Parse("ALARM:1")
	a, ok := r.(*Alarm)
	if !ok {
		t.Fatalf("expected *Alarm, got %T", r)
	}
	if a.Code != 1 {
		t.Errorf("Code = %d, want 1", a.Code)
	}
	if a.Message != "Hard limit triggered" {
		t.Errorf("Message = %q", a.Message)
	}
	if IsAck(r) {
		t.Error("Alarm should not be an ack")
	}
}

func TestParseWelcome(t *testing.T) {
	r := Parse("Grbl 1.1h ['$' for help]")
	w, ok := r.(*Welcome)
	if !ok {
		t.Fatalf("expected *Welcome, got %T", r)
	}
	if w.Version != "1.1h" {
		t.Errorf("Version = %q, want 1.1h", w.Version)
	}
}

func TestParseFeedbackMsg(t *testing.T) {
	r := Parse("[MSG:Reset to continue]")
	m, ok := r.(*FeedbackMsg)
	if !ok {
		t.Fatalf("expected *FeedbackMsg, got %T", r)
	}
	if m.Text != "Reset to continue" {
		t.Errorf("Text = %q", m.Text)
	}
}

func TestParseParserState(t *testing.T) {
	r := Parse("[GC:G0 G54 G17 G21 G90 G94 M5 M9 T0 F0 S0]")
	ps, ok := r.(*ParserState)
	if !ok {
		t.Fatalf("expected *ParserState, got %T", r)
	}
	if ps.Modes != "G0 G54 G17 G21 G90 G94 M5 M9 T0 F0 S0" {
		t.Errorf("Modes = %q", ps.Modes)
	}
}

func TestParseProbe(t *testing.T) {
	r := Parse("[PRB:1.000,2.500,-3.200:1]")
	p, ok := r.(*ProbeResult)
	if !ok {
		t.Fatalf("expected *ProbeResult, got %T", r)
	}
	if !p.Success {
		t.Error("expected Success=true")
	}
	assertFloat(t, "X", p.X, 1.0)
	assertFloat(t, "Y", p.Y, 2.5)
	assertFloat(t, "Z", p.Z, -3.2)
}

func TestParseProbeFail(t *testing.T) {
	p := Parse("[PRB:0.000,0.000,0.000:0]").(*ProbeResult)
	if p.Success {
		t.Error("expected Success=false")
	}
}

func TestParseBuildVersion(t *testing.T) {
	r := Parse("[VER:1.1h.20190825:]")
	v, ok := r.(*BuildVersion)
	if !ok {
		t.Fatalf("expected *BuildVersion, got %T", r)
	}
	if v.Version != "1.1h.20190825:" {
		t.Errorf("Version = %q", v.Version)
	}
}

func TestParseCompileOptions(t *testing.T) {
	r := Parse("[OPT:V,15,128]")
	o, ok := r.(*CompileOptions)
	if !ok {
		t.Fatalf("expected *CompileOptions, got %T", r)
	}
	if o.Raw != "V,15,128" {
		t.Errorf("Raw = %q", o.Raw)
	}
	if o.BlockBuffer != 15 {
		t.Errorf("BlockBuffer = %d, want 15", o.BlockBuffer)
	}
	if o.RxBuffer != 128 {
		t.Errorf("RxBuffer = %d, want 128", o.RxBuffer)
	}
}

func TestParseEcho(t *testing.T) {
	r := Parse("[echo:hello world]")
	e, ok := r.(*Echo)
	if !ok {
		t.Fatalf("expected *Echo, got %T", r)
	}
	if e.Text != "hello world" {
		t.Errorf("Text = %q", e.Text)
	}
}

func TestParseSetting(t *testing.T) {
	r := Parse("$32=1")
	s, ok := r.(*Setting)
	if !ok {
		t.Fatalf("expected *Setting, got %T", r)
	}
	if s.Key != 32 {
		t.Errorf("Key = %d, want 32", s.Key)
	}
	if s.Value != "1" {
		t.Errorf("Value = %q, want 1", s.Value)
	}
}

func TestParseSettingFloat(t *testing.T) {
	s := Parse("$100=800.000").(*Setting)
	if s.Key != 100 || s.Value != "800.000" {
		t.Errorf("got Key=%d Value=%q", s.Key, s.Value)
	}
}

func TestParseStartupLineOK(t *testing.T) {
	r := Parse(">G54G20:ok")
	sl, ok := r.(*StartupLine)
	if !ok {
		t.Fatalf("expected *StartupLine, got %T", r)
	}
	if sl.Line != "G54G20" {
		t.Errorf("Line = %q", sl.Line)
	}
	if !sl.OK {
		t.Error("expected OK=true")
	}
}

func TestParseStartupLineError(t *testing.T) {
	sl := Parse(">G54G20:error:1").(*StartupLine)
	if sl.OK {
		t.Error("expected OK=false")
	}
}

func TestParseUnknown(t *testing.T) {
	r := Parse("some random text")
	u, ok := r.(*Unknown)
	if !ok {
		t.Fatalf("expected *Unknown, got %T", r)
	}
	if u.Data != "some random text" {
		t.Errorf("Data = %q", u.Data)
	}
}

func TestParseStatusIdle(t *testing.T) {
	r := Parse("<Idle|MPos:0.000,0.000,0.000|FS:0,0>")
	s, ok := r.(*Status)
	if !ok {
		t.Fatalf("expected *Status, got %T", r)
	}
	if s.State != StateIdle {
		t.Errorf("State = %v, want Idle", s.State)
	}
	if s.MPos == nil {
		t.Fatal("MPos is nil")
	}
	assertFloat(t, "MPos.X", s.MPos.X, 0)
	assertFloat(t, "MPos.Y", s.MPos.Y, 0)
	assertFloat(t, "MPos.Z", s.MPos.Z, 0)
	assertFloat(t, "Feed", s.Feed, 0)
	assertFloat(t, "Spindle", s.Spindle, 0)
	if s.WPos != nil {
		t.Error("WPos should be nil")
	}
}

func TestParseStatusRun(t *testing.T) {
	s := Parse("<Run|WPos:10.500,20.300,-1.000|FS:1000,500|Ov:100,100,100>").(*Status)
	if s.State != StateRun {
		t.Errorf("State = %v, want Run", s.State)
	}
	if s.MPos != nil {
		t.Error("MPos should be nil when WPos is reported")
	}
	if s.WPos == nil {
		t.Fatal("WPos is nil")
	}
	assertFloat(t, "WPos.X", s.WPos.X, 10.5)
	assertFloat(t, "WPos.Y", s.WPos.Y, 20.3)
	assertFloat(t, "WPos.Z", s.WPos.Z, -1.0)
	assertFloat(t, "Feed", s.Feed, 1000)
	assertFloat(t, "Spindle", s.Spindle, 500)
	if s.Overrides == nil {
		t.Fatal("Overrides is nil")
	}
	if s.Overrides.Feed != 100 {
		t.Errorf("Overrides.Feed = %d", s.Overrides.Feed)
	}
	if s.Overrides.Rapids != 100 {
		t.Errorf("Overrides.Rapids = %d", s.Overrides.Rapids)
	}
	if s.Overrides.Spindle != 100 {
		t.Errorf("Overrides.Spindle = %d", s.Overrides.Spindle)
	}
}

func TestParseStatusHoldSubState(t *testing.T) {
	s := Parse("<Hold:1|MPos:5.000,10.000,0.000|FS:0,0>").(*Status)
	if s.State != StateHold {
		t.Errorf("State = %v, want Hold", s.State)
	}
	if s.SubState != 1 {
		t.Errorf("SubState = %d, want 1", s.SubState)
	}
}

func TestParseStatusDoorSubState(t *testing.T) {
	s := Parse("<Door:2|MPos:0.000,0.000,0.000|F:0>").(*Status)
	if s.State != StateDoor {
		t.Errorf("State = %v, want Door", s.State)
	}
	if s.SubState != 2 {
		t.Errorf("SubState = %d, want 2", s.SubState)
	}
	assertFloat(t, "Feed", s.Feed, 0)
	assertFloat(t, "Spindle", s.Spindle, 0)
}

func TestParseStatusFeedOnly(t *testing.T) {
	s := Parse("<Idle|MPos:0.000,0.000,0.000|F:500>").(*Status)
	assertFloat(t, "Feed", s.Feed, 500)
	assertFloat(t, "Spindle", s.Spindle, 0)
}

func TestParseStatusAllFields(t *testing.T) {
	s := Parse("<Run|MPos:1.0,2.0,3.0|FS:800,400|Ov:110,100,90|WCO:10.0,20.0,0.0|Bf:14,127|Ln:42|Pn:XYZ|A:SF>").(*Status)
	if s.State != StateRun {
		t.Errorf("State = %v", s.State)
	}
	if s.WCO == nil {
		t.Fatal("WCO is nil")
	}
	assertFloat(t, "WCO.X", s.WCO.X, 10)
	assertFloat(t, "WCO.Y", s.WCO.Y, 20)
	assertFloat(t, "WCO.Z", s.WCO.Z, 0)
	if s.Buffer == nil {
		t.Fatal("Buffer is nil")
	}
	if s.Buffer.PlannerBlocks != 14 {
		t.Errorf("PlannerBlocks = %d", s.Buffer.PlannerBlocks)
	}
	if s.Buffer.SerialBytes != 127 {
		t.Errorf("SerialBytes = %d", s.Buffer.SerialBytes)
	}
	if s.LineNum != 42 {
		t.Errorf("LineNum = %d, want 42", s.LineNum)
	}
	if s.Pins != "XYZ" {
		t.Errorf("Pins = %q, want XYZ", s.Pins)
	}
	if s.Accessory != "SF" {
		t.Errorf("Accessory = %q, want SF", s.Accessory)
	}
}

func TestParseStatusLineNumDefault(t *testing.T) {
	s := Parse("<Idle|MPos:0.000,0.000,0.000|FS:0,0>").(*Status)
	if s.LineNum != -1 {
		t.Errorf("LineNum = %d, want -1 (not reported)", s.LineNum)
	}
}

func TestMachineStateString(t *testing.T) {
	tests := []struct {
		state MachineState
		want  string
	}{
		{StateIdle, "Idle"},
		{StateRun, "Run"},
		{StateHold, "Hold"},
		{StateAlarm, "Alarm"},
		{StateSleep, "Sleep"},
		{StateUnknown, "Unknown"},
		{MachineState(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("%d.String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestIsAck(t *testing.T) {
	if !IsAck(&OK{}) {
		t.Error("OK should be ack")
	}
	if !IsAck(&Error{Code: 1}) {
		t.Error("Error should be ack")
	}
	if IsAck(&Alarm{Code: 1}) {
		t.Error("Alarm should not be ack")
	}
	if IsAck(&Status{}) {
		t.Error("Status should not be ack")
	}
	if IsAck(&Welcome{}) {
		t.Error("Welcome should not be ack")
	}
}

func TestParseSettingNotDollarDollar(t *testing.T) {
	r := Parse("$$")
	if _, ok := r.(*Unknown); !ok {
		t.Errorf("$$ should be Unknown, got %T", r)
	}
}

func assertFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.001 {
		t.Errorf("%s = %f, want %f", name, got, want)
	}
}
