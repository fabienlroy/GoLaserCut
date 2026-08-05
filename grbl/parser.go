package grbl

import (
	"strconv"
	"strings"
)

func Parse(line string) Response {
	switch {
	case line == "ok":
		return &OK{}
	case strings.HasPrefix(line, "error:"):
		return parseError(line)
	case strings.HasPrefix(line, "ALARM:"):
		return parseAlarm(line)
	case len(line) > 2 && line[0] == '<' && line[len(line)-1] == '>':
		return parseStatus(line)
	case strings.HasPrefix(line, "Grbl "):
		return parseWelcome(line)
	case strings.HasPrefix(line, "[MSG:"):
		return parseFeedbackMsg(line)
	case strings.HasPrefix(line, "[GC:"):
		return parseParserState(line)
	case strings.HasPrefix(line, "[PRB:"):
		return parseProbe(line)
	case strings.HasPrefix(line, "[VER:"):
		return parseBuildVersion(line)
	case strings.HasPrefix(line, "[OPT:"):
		return parseCompileOptions(line)
	case strings.HasPrefix(line, "[echo:"):
		return parseEcho(line)
	case len(line) > 1 && line[0] == '$' && line[1] != '$':
		return parseSetting(line)
	case strings.HasPrefix(line, ">"):
		return parseStartupLine(line)
	default:
		return &Unknown{Data: line}
	}
}

func parseError(line string) *Error {
	code, _ := strconv.Atoi(strings.TrimPrefix(line, "error:"))
	return &Error{Code: code, Message: errorMessages[code]}
}

func parseAlarm(line string) *Alarm {
	code, _ := strconv.Atoi(strings.TrimPrefix(line, "ALARM:"))
	return &Alarm{Code: code, Message: alarmMessages[code]}
}

func parseWelcome(line string) *Welcome {
	parts := strings.SplitN(line, " ", 3)
	version := ""
	if len(parts) >= 2 {
		version = parts[1]
	}
	return &Welcome{Version: version}
}

func parseFeedbackMsg(line string) *FeedbackMsg {
	text := strings.TrimSuffix(strings.TrimPrefix(line, "[MSG:"), "]")
	return &FeedbackMsg{Text: text}
}

func parseParserState(line string) *ParserState {
	modes := strings.TrimSuffix(strings.TrimPrefix(line, "[GC:"), "]")
	return &ParserState{Modes: modes}
}

func parseProbe(line string) *ProbeResult {
	inner := strings.TrimSuffix(strings.TrimPrefix(line, "[PRB:"), "]")
	parts := strings.Split(inner, ":")
	if len(parts) < 2 {
		return &ProbeResult{}
	}
	r := &ProbeResult{Success: parts[1] == "1"}
	coords := strings.Split(parts[0], ",")
	if len(coords) >= 3 {
		r.X, _ = strconv.ParseFloat(coords[0], 64)
		r.Y, _ = strconv.ParseFloat(coords[1], 64)
		r.Z, _ = strconv.ParseFloat(coords[2], 64)
	}
	return r
}

func parseBuildVersion(line string) *BuildVersion {
	inner := strings.TrimSuffix(strings.TrimPrefix(line, "[VER:"), "]")
	return &BuildVersion{Version: inner}
}

func parseCompileOptions(line string) *CompileOptions {
	inner := strings.TrimSuffix(strings.TrimPrefix(line, "[OPT:"), "]")
	opts := &CompileOptions{Raw: inner}
	parts := strings.Split(inner, ",")
	if len(parts) >= 3 {
		opts.BlockBuffer, _ = strconv.Atoi(parts[1])
		opts.RxBuffer, _ = strconv.Atoi(parts[2])
	}
	return opts
}

func parseEcho(line string) *Echo {
	text := strings.TrimSuffix(strings.TrimPrefix(line, "[echo:"), "]")
	return &Echo{Text: text}
}

func parseSetting(line string) *Setting {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return &Setting{}
	}
	key, _ := strconv.Atoi(strings.TrimPrefix(parts[0], "$"))
	return &Setting{Key: key, Value: parts[1]}
}

func parseStartupLine(line string) *StartupLine {
	inner := strings.TrimPrefix(line, ">")
	parts := strings.SplitN(inner, ":", 2)
	sl := &StartupLine{Line: parts[0]}
	if len(parts) >= 2 {
		sl.OK = parts[1] == "ok"
	}
	return sl
}

func parseStatus(line string) *Status {
	inner := line[1 : len(line)-1]
	fields := strings.Split(inner, "|")
	s := &Status{LineNum: -1}

	if len(fields) == 0 {
		return s
	}
	s.State, s.SubState = parseState(fields[0])

	for _, field := range fields[1:] {
		kv := strings.SplitN(field, ":", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "MPos":
			s.MPos = parsePosition(kv[1])
		case "WPos":
			s.WPos = parsePosition(kv[1])
		case "FS":
			vals := strings.Split(kv[1], ",")
			if len(vals) >= 1 {
				s.Feed, _ = strconv.ParseFloat(vals[0], 64)
			}
			if len(vals) >= 2 {
				s.Spindle, _ = strconv.ParseFloat(vals[1], 64)
			}
		case "F":
			s.Feed, _ = strconv.ParseFloat(kv[1], 64)
		case "Ov":
			s.Overrides = parseOverrides(kv[1])
		case "WCO":
			s.WCO = parsePosition(kv[1])
		case "Bf":
			s.Buffer = parseBufferState(kv[1])
		case "Ln":
			s.LineNum, _ = strconv.Atoi(kv[1])
		case "Pn":
			s.Pins = kv[1]
		case "A":
			s.Accessory = kv[1]
		}
	}
	return s
}

func parseState(s string) (MachineState, int) {
	parts := strings.SplitN(s, ":", 2)
	var state MachineState
	switch parts[0] {
	case "Idle":
		state = StateIdle
	case "Run":
		state = StateRun
	case "Hold":
		state = StateHold
	case "Jog":
		state = StateJog
	case "Alarm":
		state = StateAlarm
	case "Door":
		state = StateDoor
	case "Check":
		state = StateCheck
	case "Home":
		state = StateHome
	case "Sleep":
		state = StateSleep
	default:
		state = StateUnknown
	}
	sub := 0
	if len(parts) == 2 {
		sub, _ = strconv.Atoi(parts[1])
	}
	return state, sub
}

func parsePosition(s string) *Position {
	vals := strings.Split(s, ",")
	if len(vals) < 3 {
		return nil
	}
	p := &Position{}
	p.X, _ = strconv.ParseFloat(vals[0], 64)
	p.Y, _ = strconv.ParseFloat(vals[1], 64)
	p.Z, _ = strconv.ParseFloat(vals[2], 64)
	return p
}

func parseOverrides(s string) *Overrides {
	vals := strings.Split(s, ",")
	if len(vals) < 3 {
		return nil
	}
	o := &Overrides{}
	o.Feed, _ = strconv.Atoi(vals[0])
	o.Rapids, _ = strconv.Atoi(vals[1])
	o.Spindle, _ = strconv.Atoi(vals[2])
	return o
}

func parseBufferState(s string) *BufferState {
	vals := strings.Split(s, ",")
	if len(vals) < 2 {
		return nil
	}
	b := &BufferState{}
	b.PlannerBlocks, _ = strconv.Atoi(vals[0])
	b.SerialBytes, _ = strconv.Atoi(vals[1])
	return b
}
