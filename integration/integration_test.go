package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/fabienlroy/GoLaserCut/grbl"
	"github.com/fabienlroy/GoLaserCut/serial"
)

func newTestSetup(t *testing.T) (*serial.Connection, *serial.Sender, *GRBLSimulator) {
	t.Helper()
	port, sim := NewGRBLSimulator()
	conn := serial.NewConnection(port, "simulator")
	sender := serial.NewSender(conn)

	// Drain welcome message
	select {
	case line := <-conn.Lines:
		resp := grbl.Parse(line)
		if _, ok := resp.(*grbl.Welcome); !ok {
			t.Fatalf("expected welcome, got %T: %s", resp, line)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for welcome")
	}

	return conn, sender, sim
}

func drainAcks(t *testing.T, conn *serial.Connection, sender *serial.Sender, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		select {
		case line := <-conn.Lines:
			resp := grbl.Parse(line)
			if _, ok := resp.(*grbl.OK); ok {
				sender.Ack()
			} else if e, ok := resp.(*grbl.Error); ok {
				sender.Ack()
				t.Logf("error response at line %d: %s", i, e.Error())
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for ack %d/%d", i+1, count)
		}
	}
}

// --- Tests ---

func TestWelcomeMessage(t *testing.T) {
	port, sim := NewGRBLSimulator()
	defer sim.Close()
	conn := serial.NewConnection(port, "test")
	defer conn.Close()

	select {
	case line := <-conn.Lines:
		resp := grbl.Parse(line)
		w, ok := resp.(*grbl.Welcome)
		if !ok {
			t.Fatalf("expected Welcome, got %T", resp)
		}
		if w.Version != "1.1h" {
			t.Errorf("version = %q, want 1.1h", w.Version)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestSendSingleLine(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	if err := sender.Send("G0 X10 Y20"); err != nil {
		t.Fatal(err)
	}
	drainAcks(t, conn, sender, 1)

	x, y, _ := sim.Position()
	if x != 10 || y != 20 {
		t.Errorf("position = (%.1f, %.1f), want (10, 20)", x, y)
	}
	if sim.LinesProcessed() != 1 {
		t.Errorf("lines processed = %d, want 1", sim.LinesProcessed())
	}
}

func TestSendMultipleLines(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	lines := []string{
		"G90 G21",
		"M4",
		"S1000",
		"G0 X0 Y0",
		"G1 X50 Y0 F1000",
		"G1 X50 Y30",
		"G1 X0 Y30",
		"G1 X0 Y0",
		"S0",
		"M5",
		"M2",
	}

	done := make(chan struct{})
	go func() {
		for _, l := range lines {
			if err := sender.Send(l); err != nil {
				return
			}
		}
		close(done)
	}()

	// Drain all responses
	acked := 0
	timeout := time.After(10 * time.Second)
	for acked < len(lines) {
		select {
		case line, ok := <-conn.Lines:
			if !ok {
				t.Fatal("channel closed")
			}
			resp := grbl.Parse(line)
			if grbl.IsAck(resp) {
				sender.Ack()
				acked++
			}
		case <-timeout:
			t.Fatalf("timeout after %d/%d acks", acked, len(lines))
		}
	}

	<-done

	if sim.LinesProcessed() != len(lines) {
		t.Errorf("processed = %d, want %d", sim.LinesProcessed(), len(lines))
	}
}

func TestStatusReport(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	sender.Send("G0 X25.5 Y10.0 Z-3.0")
	drainAcks(t, conn, sender, 1)

	conn.WriteRealtime(serial.CmdStatusReport)

	select {
	case line := <-conn.Lines:
		resp := grbl.Parse(line)
		status, ok := resp.(*grbl.Status)
		if !ok {
			t.Fatalf("expected Status, got %T: %s", resp, line)
		}
		if status.State != grbl.StateIdle {
			t.Errorf("state = %v, want Idle", status.State)
		}
		if status.MPos == nil {
			t.Fatal("MPos is nil")
		}
		if status.MPos.X != 25.5 || status.MPos.Y != 10 || status.MPos.Z != -3 {
			t.Errorf("position = (%.1f, %.1f, %.1f), want (25.5, 10.0, -3.0)",
				status.MPos.X, status.MPos.Y, status.MPos.Z)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for status")
	}

	_ = sim
}

func TestFeedHoldResume(t *testing.T) {
	port, sim := NewGRBLSimulator()
	conn := serial.NewConnection(port, "test")
	defer func() { conn.Close(); sim.Close() }()

	// Drain welcome
	<-conn.Lines

	// Send a command to set state to Run then Idle
	conn.WriteLine("G0 X10")
	<-conn.Lines // ok

	// Feed hold
	conn.WriteRealtime(serial.CmdFeedHold)
	// State might not change since we're already Idle, but the sim handles it
	// Let's query status
	conn.WriteRealtime(serial.CmdStatusReport)
	<-conn.Lines // status

	// Resume
	conn.WriteRealtime(serial.CmdCycleResume)
}

func TestSoftReset(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	sender.Send("G0 X50 Y50")
	drainAcks(t, conn, sender, 1)

	x, y, _ := sim.Position()
	if x != 50 || y != 50 {
		t.Errorf("pre-reset position = (%.0f, %.0f)", x, y)
	}

	// Soft reset
	conn.WriteRealtime(serial.CmdSoftReset)
	sender.Reset()

	// Should get new welcome
	select {
	case line := <-conn.Lines:
		resp := grbl.Parse(line)
		if _, ok := resp.(*grbl.Welcome); !ok {
			t.Fatalf("expected Welcome after reset, got %T: %s", resp, line)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for welcome after reset")
	}

	x, y, _ = sim.Position()
	if x != 0 || y != 0 {
		t.Errorf("post-reset position = (%.0f, %.0f), want (0, 0)", x, y)
	}
}

func TestErrorResponse(t *testing.T) {
	conn, _, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	conn.WriteLine("INVALID_COMMAND")

	select {
	case line := <-conn.Lines:
		resp := grbl.Parse(line)
		e, ok := resp.(*grbl.Error)
		if !ok {
			t.Fatalf("expected Error, got %T: %s", resp, line)
		}
		if e.Code != 20 {
			t.Errorf("error code = %d, want 20", e.Code)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	if sim.Errors() != 1 {
		t.Errorf("errors = %d, want 1", sim.Errors())
	}
}

func TestGCodeFileStream(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	// Simulate streaming a small G-code file
	gcode := `G90 G21
G17
M4
S800
G0 X0 Y0
G1 X100 Y0 F1000
G1 X100 Y50 F1000
G1 X0 Y50 F1000
G1 X0 Y0 F1000
S0
M5
G0 X0 Y0
M2`

	lines := strings.Split(gcode, "\n")

	// Send all lines with proper flow control
	done := make(chan error, 1)
	go func() {
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			if err := sender.Send(l); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	// Process responses
	acked := 0
	timeout := time.After(10 * time.Second)
	for acked < len(lines) {
		select {
		case line, ok := <-conn.Lines:
			if !ok {
				goto finished
			}
			resp := grbl.Parse(line)
			if grbl.IsAck(resp) {
				sender.Ack()
				acked++
			}
		case <-timeout:
			t.Fatalf("timeout after %d/%d acks", acked, len(lines))
		}
	}

finished:
	err := <-done
	if err != nil {
		t.Fatalf("send error: %v", err)
	}

	if sim.LinesProcessed() != len(lines) {
		t.Errorf("processed = %d, want %d", sim.LinesProcessed(), len(lines))
	}
	if sim.Errors() != 0 {
		t.Errorf("errors = %d, want 0", sim.Errors())
	}
}

func TestPauseResumeDuringSend(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	// Send first line
	sender.Send("G0 X10")
	drainAcks(t, conn, sender, 1)

	// Pause
	sender.Pause()

	// Try to send — should block
	sent := make(chan error, 1)
	go func() {
		sent <- sender.Send("G0 X20")
	}()

	select {
	case <-sent:
		t.Fatal("Send should block while paused")
	case <-time.After(200 * time.Millisecond):
	}

	// Resume
	sender.Resume()

	select {
	case err := <-sent:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("still blocked after resume")
	}

	drainAcks(t, conn, sender, 1)

	x, _, _ := sim.Position()
	if x != 20 {
		t.Errorf("X = %.0f, want 20", x)
	}
}

func TestCharacterCountingProtocol(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	// Send multiple short lines that fit in buffer
	for i := 0; i < 10; i++ {
		go func() {
			sender.Send("G0 X1")
		}()
	}

	// Drain all acks
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 10; i++ {
		select {
		case line := <-conn.Lines:
			resp := grbl.Parse(line)
			if grbl.IsAck(resp) {
				sender.Ack()
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout at ack %d/10", i+1)
		}
	}

	if sim.LinesProcessed() != 10 {
		t.Errorf("processed = %d, want 10", sim.LinesProcessed())
	}
}

func TestSettingsQuery(t *testing.T) {
	conn, _, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	conn.WriteLine("$$")

	// Should get settings + ok
	settings := []string{}
	timeout := time.After(2 * time.Second)
	for {
		select {
		case line := <-conn.Lines:
			resp := grbl.Parse(line)
			if _, ok := resp.(*grbl.OK); ok {
				goto done
			}
			if s, ok := resp.(*grbl.Setting); ok {
				settings = append(settings, line)
				_ = s
			}
		case <-timeout:
			t.Fatal("timeout")
		}
	}
done:
	if len(settings) < 2 {
		t.Errorf("expected at least 2 settings, got %d", len(settings))
	}
}

func TestLaserModeCommands(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	cmds := []string{"M4", "S500", "G1 X10 F1000", "S0", "M5"}
	for _, cmd := range cmds {
		sender.Send(cmd)
	}
	drainAcks(t, conn, sender, len(cmds))

	if sim.Errors() != 0 {
		t.Errorf("errors = %d, want 0", sim.Errors())
	}
}

func TestCoolantCommands(t *testing.T) {
	conn, sender, sim := newTestSetup(t)
	defer func() { conn.Close(); sim.Close() }()

	cmds := []string{"M7", "G0 X10", "M9"}
	for _, cmd := range cmds {
		sender.Send(cmd)
	}
	drainAcks(t, conn, sender, len(cmds))

	if sim.Errors() != 0 {
		t.Errorf("errors = %d", sim.Errors())
	}
}
