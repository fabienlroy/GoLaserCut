package gcode

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := CleanLine(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func CleanLine(line string) string {
	if i := strings.Index(line, ";"); i >= 0 {
		line = line[:i]
	}
	for {
		start := strings.Index(line, "(")
		if start < 0 {
			break
		}
		end := strings.Index(line[start:], ")")
		if end < 0 {
			break
		}
		line = line[:start] + line[start+end+1:]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	return strings.ToUpper(line)
}
