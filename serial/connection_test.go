package serial

import (
	"io"
	"sync"
	"testing"
	"time"
)

type mockPort struct {
	pr     *io.PipeReader
	sent   []byte
	sentMu sync.Mutex
}

func newMockPort() (*mockPort, *io.PipeWriter) {
	pr, pw := io.Pipe()
	return &mockPort{pr: pr}, pw
}

func (m *mockPort) Read(p []byte) (int, error)  { return m.pr.Read(p) }
func (m *mockPort) Write(p []byte) (int, error) {
	m.sentMu.Lock()
	defer m.sentMu.Unlock()
	m.sent = append(m.sent, p...)
	return len(p), nil
}
func (m *mockPort) Close() error { return m.pr.Close() }

func (m *mockPort) Sent() string {
	m.sentMu.Lock()
	defer m.sentMu.Unlock()
	return string(m.sent)
}

func TestReadLines(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")

	go func() {
		pw.Write([]byte("Grbl 1.1h ['$' for help]\r\n"))
		pw.Write([]byte("ok\r\n"))
		pw.Close()
	}()

	lines := drain(t, conn.Lines, 2)
	if lines[0] != "Grbl 1.1h ['$' for help]" {
		t.Errorf("line 0 = %q, want banner", lines[0])
	}
	if lines[1] != "ok" {
		t.Errorf("line 1 = %q, want ok", lines[1])
	}
	conn.Close()
}

func TestReadLinesCRLF(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")

	go func() {
		pw.Write([]byte("ok\r\n"))
		pw.Write([]byte("error:1\n"))
		pw.Close()
	}()

	lines := drain(t, conn.Lines, 2)
	if lines[0] != "ok" {
		t.Errorf("line 0 = %q, want ok", lines[0])
	}
	if lines[1] != "error:1" {
		t.Errorf("line 1 = %q, want error:1", lines[1])
	}
	conn.Close()
}

func TestWriteLine(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	if err := conn.WriteLine("G0 X10"); err != nil {
		t.Fatal(err)
	}
	if got := mock.Sent(); got != "G0 X10\n" {
		t.Errorf("sent = %q, want %q", got, "G0 X10\n")
	}
}

func TestWriteRealtime(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")
	defer func() { pw.Close(); conn.Close() }()

	if err := conn.WriteRealtime(CmdStatusReport); err != nil {
		t.Fatal(err)
	}
	if got := mock.Sent(); got != "?" {
		t.Errorf("sent = %q, want %q", got, "?")
	}
}

func TestWriteAfterClose(t *testing.T) {
	mock, _ := newMockPort()
	conn := NewConnection(mock, "mock")
	conn.Close()

	if err := conn.WriteLine("test"); err == nil {
		t.Error("expected error writing to closed connection")
	}
	if err := conn.WriteRealtime(CmdStatusReport); err == nil {
		t.Error("expected error writing realtime to closed connection")
	}
}

func TestCloseIdempotent(t *testing.T) {
	mock, _ := newMockPort()
	conn := NewConnection(mock, "mock")
	conn.Close()
	if err := conn.Close(); err != nil {
		t.Errorf("second Close() returned error: %v", err)
	}
}

func TestLinesClosedOnDisconnect(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "mock")

	pw.Write([]byte("ok\r\n"))
	pw.Close()

	count := 0
	for range conn.Lines {
		count++
	}
	if count != 1 {
		t.Errorf("received %d lines, want 1", count)
	}
}

func TestPortName(t *testing.T) {
	mock, pw := newMockPort()
	conn := NewConnection(mock, "/dev/ttyUSB0")
	defer func() { pw.Close(); conn.Close() }()

	if got := conn.PortName(); got != "/dev/ttyUSB0" {
		t.Errorf("PortName() = %q, want /dev/ttyUSB0", got)
	}
}

func drain(t *testing.T, ch <-chan string, n int) []string {
	t.Helper()
	var lines []string
	timeout := time.After(2 * time.Second)
	for len(lines) < n {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatalf("channel closed after %d lines, want %d", len(lines), n)
			}
			lines = append(lines, line)
		case <-timeout:
			t.Fatalf("timeout after %d lines, want %d", len(lines), n)
		}
	}
	return lines
}
