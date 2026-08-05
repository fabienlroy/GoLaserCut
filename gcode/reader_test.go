package gcode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"G0 X10 Y20", "G0 X10 Y20"},
		{"g0 x10 y20", "G0 X10 Y20"},
		{"G0 X10 ; move to start", "G0 X10"},
		{"G0 X10 (rapid move) Y20", "G0 X10  Y20"},
		{"(full line comment)", ""},
		{"; full line comment", ""},
		{"  G1 X5  ", "G1 X5"},
		{"", ""},
		{"   ", ""},
		{"M4 S1000 ; laser on (full power)", "M4 S1000"},
		{"G0 (move) X10 (to) Y20", "G0  X10  Y20"},
	}
	for _, tt := range tests {
		got := CleanLine(tt.input)
		if got != tt.want {
			t.Errorf("CleanLine(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestReadFile(t *testing.T) {
	content := `; header comment
G90 ; absolute mode
G21 ; mm mode
(start cutting)
G0 X0 Y0
G1 X10 Y10 F1000 S500
G1 X20 Y0

M5 ; laser off
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.gcode")
	os.WriteFile(path, []byte(content), 0644)

	lines, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"G90", "G21", "G0 X0 Y0", "G1 X10 Y10 F1000 S500", "G1 X20 Y0", "M5"}
	if len(lines) != len(expected) {
		t.Fatalf("got %d lines, want %d", len(lines), len(expected))
	}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("line %d = %q, want %q", i, lines[i], want)
		}
	}
}

func TestReadFileNotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/file.gcode")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
