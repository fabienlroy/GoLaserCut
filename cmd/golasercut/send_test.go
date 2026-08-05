package main

import "testing"

func TestProgressBar(t *testing.T) {
	tests := []struct {
		pct  int
		want string
	}{
		{0, "[>                             ]"},
		{50, "[===============>              ]"},
		{100, "[==============================]"},
	}
	for _, tt := range tests {
		got := progressBar(tt.pct, 30)
		if got != tt.want {
			t.Errorf("progressBar(%d, 30) = %q, want %q", tt.pct, got, tt.want)
		}
	}
}
