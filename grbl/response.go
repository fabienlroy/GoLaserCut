package grbl

import "fmt"

type Response interface {
	grblResponse()
}

type OK struct{}

func (*OK) grblResponse() {}

type Error struct {
	Code    int
	Message string
}

func (*Error) grblResponse() {}

func (e *Error) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("error:%d (%s)", e.Code, e.Message)
	}
	return fmt.Sprintf("error:%d", e.Code)
}

type Alarm struct {
	Code    int
	Message string
}

func (*Alarm) grblResponse() {}

func (a *Alarm) Error() string {
	if a.Message != "" {
		return fmt.Sprintf("ALARM:%d (%s)", a.Code, a.Message)
	}
	return fmt.Sprintf("ALARM:%d", a.Code)
}

type Welcome struct {
	Version string
}

func (*Welcome) grblResponse() {}

type FeedbackMsg struct {
	Text string
}

func (*FeedbackMsg) grblResponse() {}

type ParserState struct {
	Modes string
}

func (*ParserState) grblResponse() {}

type ProbeResult struct {
	X, Y, Z float64
	Success bool
}

func (*ProbeResult) grblResponse() {}

type BuildVersion struct {
	Version string
}

func (*BuildVersion) grblResponse() {}

type CompileOptions struct {
	Raw         string
	BlockBuffer int
	RxBuffer    int
}

func (*CompileOptions) grblResponse() {}

type Echo struct {
	Text string
}

func (*Echo) grblResponse() {}

type Setting struct {
	Key   int
	Value string
}

func (*Setting) grblResponse() {}

type StartupLine struct {
	Line string
	OK   bool
}

func (*StartupLine) grblResponse() {}

type Unknown struct {
	Data string
}

func (*Unknown) grblResponse() {}

func IsAck(r Response) bool {
	switch r.(type) {
	case *OK, *Error:
		return true
	}
	return false
}
