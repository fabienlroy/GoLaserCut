package serial

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSenderSimple(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)
	if err := s.Send("G0 X10"); err != nil {
		t.Fatal(err)
	}
	if got := s.InFlight(); got != 7 {
		t.Errorf("InFlight() = %d, want 7", got)
	}
	if got := s.Pending(); got != 1 {
		t.Errorf("Pending() = %d, want 1", got)
	}

	s.Ack()
	if got := s.InFlight(); got != 0 {
		t.Errorf("after Ack, InFlight() = %d, want 0", got)
	}
	if got := s.Pending(); got != 0 {
		t.Errorf("after Ack, Pending() = %d, want 0", got)
	}
}

func TestSenderBufferFull(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)

	long := strings.Repeat("x", grblBufferSize-2)
	if err := s.Send(long); err != nil {
		t.Fatal(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- s.Send("G0 X0")
	}()

	select {
	case <-done:
		t.Fatal("Send should block when buffer is full")
	case <-time.After(100 * time.Millisecond):
	}

	s.Ack()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Send still blocked after Ack")
	}
}

func TestSenderLineTooLong(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)
	long := strings.Repeat("G", grblBufferSize)
	if err := s.Send(long); err == nil {
		t.Error("expected error for line exceeding buffer size")
	}
}

func TestSenderStop(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)
	long := strings.Repeat("x", grblBufferSize-2) // fills buffer
	s.Send(long)

	done := make(chan error, 1)
	go func() {
		done <- s.Send("G0 X0")
	}()

	time.Sleep(50 * time.Millisecond)
	s.Stop()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected error after Stop")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Send still blocked after Stop")
	}
}

func TestSenderPauseResume(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)
	s.Pause()

	done := make(chan error, 1)
	go func() {
		done <- s.Send("G0 X10")
	}()

	select {
	case <-done:
		t.Fatal("Send should block when paused")
	case <-time.After(100 * time.Millisecond):
	}

	s.Resume()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Send still blocked after Resume")
	}
}

func TestSenderReset(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)
	s.Send("G0 X10")
	s.Stop()

	s.Reset()

	if got := s.InFlight(); got != 0 {
		t.Errorf("after Reset, InFlight() = %d, want 0", got)
	}
	if err := s.Send("G0 X0"); err != nil {
		t.Errorf("Send after Reset failed: %v", err)
	}
}

func TestSenderMultipleLines(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)

	lines := []string{"G0 X10", "G1 Y20 F500", "M5"}
	var wg sync.WaitGroup
	for _, l := range lines {
		wg.Add(1)
		go func(line string) {
			defer wg.Done()
			s.Send(line)
		}(l)
	}

	wg.Wait()
	expected := len("G0 X10") + 1 + len("G1 Y20 F500") + 1 + len("M5") + 1
	if got := s.InFlight(); got != expected {
		t.Errorf("InFlight() = %d, want %d", got, expected)
	}
	if got := s.Pending(); got != 3 {
		t.Errorf("Pending() = %d, want 3", got)
	}
}

func TestSenderAckEmpty(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	s := NewSender(conn)
	s.Ack() // should not panic
	if got := s.InFlight(); got != 0 {
		t.Errorf("InFlight() = %d after empty Ack", got)
	}
}

