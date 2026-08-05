package gcode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriterBasic(t *testing.T) {
	w := NewWriter()
	w.Add("G90 G21")
	w.Add("M5")

	lines := w.Lines()
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0] != "G90 G21" || lines[1] != "M5" {
		t.Errorf("unexpected lines: %v", lines)
	}
}

func TestWriterAddf(t *testing.T) {
	w := NewWriter()
	w.Addf("G1 X%.3f Y%.3f F%.0f", 10.5, 20.0, 1000.0)

	if w.Lines()[0] != "G1 X10.500 Y20.000 F1000" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterComment(t *testing.T) {
	w := NewWriter()
	w.Comment("test comment")
	if w.Lines()[0] != "; test comment" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterBlank(t *testing.T) {
	w := NewWriter()
	w.Add("G90")
	w.Blank()
	w.Add("M5")
	if w.Lines()[1] != "" {
		t.Errorf("blank line should be empty string, got %q", w.Lines()[1])
	}
}

func TestWriterString(t *testing.T) {
	w := NewWriter()
	w.Add("G90")
	w.Add("M5")
	s := w.String()
	if s != "G90\nM5\n" {
		t.Errorf("got %q", s)
	}
}

func TestWriterCount(t *testing.T) {
	w := NewWriter()
	w.Add("A")
	w.Add("B")
	w.Add("C")
	if w.Count() != 3 {
		t.Errorf("Count() = %d, want 3", w.Count())
	}
}

func TestWriterHeader(t *testing.T) {
	w := NewWriter()
	w.Header(true, 800)
	s := w.String()

	checks := []string{"GoLaserCut", "G90 G21", "G17", "M4", "S800"}
	for _, c := range checks {
		if !strings.Contains(s, c) {
			t.Errorf("header missing %q", c)
		}
	}
}

func TestWriterHeaderM3(t *testing.T) {
	w := NewWriter()
	w.Header(false, 500)
	s := w.String()

	if !strings.Contains(s, "M3") {
		t.Error("expected M3 for non-laser mode")
	}
	if strings.Contains(s, "M4") {
		t.Error("should not have M4")
	}
}

func TestWriterFooter(t *testing.T) {
	w := NewWriter()
	w.Footer()
	s := w.String()

	checks := []string{"S0", "M5", "G0 X0 Y0", "M2"}
	for _, c := range checks {
		if !strings.Contains(s, c) {
			t.Errorf("footer missing %q", c)
		}
	}
}

func TestWriterRapid(t *testing.T) {
	w := NewWriter()
	w.Rapid(10.5, 20.0)
	if w.Lines()[0] != "G0 X10.500 Y20.000" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterRapidZ(t *testing.T) {
	w := NewWriter()
	w.RapidZ(5.0)
	if w.Lines()[0] != "G0 Z5.000" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterLinearMove(t *testing.T) {
	w := NewWriter()
	w.LinearMove(50, 30, 1000)
	if w.Lines()[0] != "G1 X50.000 Y30.000 F1000" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterLinearMoveZ(t *testing.T) {
	w := NewWriter()
	w.LinearMoveZ(10, 20, -1.5, 300)
	if w.Lines()[0] != "G1 X10.000 Y20.000 Z-1.500 F300" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterSetPower(t *testing.T) {
	w := NewWriter()
	w.SetPower(500)
	if w.Lines()[0] != "S500" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterAirAssist(t *testing.T) {
	w := NewWriter()
	w.AirAssistOn()
	w.AirAssistOff()
	if w.Lines()[0] != "M7" || w.Lines()[1] != "M9" {
		t.Errorf("got %v", w.Lines())
	}
}

func TestWriterDwell(t *testing.T) {
	w := NewWriter()
	w.Dwell(500)
	if w.Lines()[0] != "G4 P500.0" {
		t.Errorf("got %q", w.Lines()[0])
	}
}

func TestWriterMultiPass(t *testing.T) {
	w := NewWriter()
	w.MultiPass(3, func(pass int) {
		w.Addf("G1 X%d", pass*10)
	})
	s := w.String()

	for i := 1; i <= 3; i++ {
		marker := "; pass " + strings.Replace("N/3", "N", string(rune('0'+i)), 1)
		_ = marker
	}

	if !strings.Contains(s, "; pass 1/3") {
		t.Error("missing pass 1/3")
	}
	if !strings.Contains(s, "; pass 3/3") {
		t.Error("missing pass 3/3")
	}
	if !strings.Contains(s, "G1 X0") || !strings.Contains(s, "G1 X20") {
		t.Error("missing pass content")
	}
}

func TestWriterMultiPassSingle(t *testing.T) {
	w := NewWriter()
	w.MultiPass(1, func(pass int) {
		w.Add("G1 X10")
	})
	s := w.String()
	if strings.Contains(s, "; pass") {
		t.Error("single pass should not have pass markers")
	}
}

func TestWriterMultiPassZero(t *testing.T) {
	w := NewWriter()
	w.MultiPass(0, func(pass int) {
		w.Add("G1 X10")
	})
	if w.Count() != 1 {
		t.Errorf("zero passes should default to 1, got %d lines", w.Count())
	}
}

func TestWriterWriteFile(t *testing.T) {
	w := NewWriter()
	w.Header(true, 1000)
	w.Rapid(10, 20)
	w.LinearMove(50, 30, 1000)
	w.Footer()

	dir := t.TempDir()
	path := filepath.Join(dir, "output.gcode")

	if err := w.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "G90 G21") {
		t.Error("file missing header")
	}
	if !strings.Contains(content, "M2") {
		t.Error("file missing footer")
	}
}

func TestWriterWriteFileInvalidPath(t *testing.T) {
	w := NewWriter()
	w.Add("G90")
	err := w.WriteFile("/nonexistent/dir/out.gcode")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestWriterFullWorkflow(t *testing.T) {
	w := NewWriter()
	w.Header(true, 800)
	w.AirAssistOn()
	w.RapidZ(5)

	w.MultiPass(2, func(pass int) {
		w.Rapid(0, 0)
		w.LinearMove(100, 0, 1000)
		w.LinearMove(100, 50, 1000)
		w.LinearMove(0, 0, 1000)
		w.RapidZ(5)
	})

	w.AirAssistOff()
	w.Footer()

	s := w.String()
	if !strings.Contains(s, "; pass 1/2") || !strings.Contains(s, "; pass 2/2") {
		t.Error("missing pass markers")
	}
	if !strings.Contains(s, "M7") || !strings.Contains(s, "M9") {
		t.Error("missing air assist")
	}

	lines, err := ReadFile(writeTemp(t, s))
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Error("round-trip read should produce lines")
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.gcode")
	os.WriteFile(path, []byte(content), 0644)
	return path
}
