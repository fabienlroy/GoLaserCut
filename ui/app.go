package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/fabienlroy/GoLaserCut/gcode"
	"github.com/fabienlroy/GoLaserCut/grbl"
	"github.com/fabienlroy/GoLaserCut/serial"
)

type C = layout.Context
type D = layout.Dimensions

type serialEvent struct {
	kind int
	line string
	resp grbl.Response
	err  error
	conn *serial.Connection
}

const (
	evLine = iota
	evConnected
	evDisconnected
	evSendDone
)

type App struct {
	win   *app.Window
	theme *material.Theme

	conn      *serial.Connection
	sender    *serial.Sender
	connected bool
	mu        sync.Mutex

	serialCh chan serialEvent

	portEditor widget.Editor
	connectBtn widget.Clickable
	fileEditor widget.Editor
	loadBtn    widget.Clickable

	consoleList widget.List
	logLines    []string
	cmdEditor   widget.Editor
	sendCmdBtn  widget.Clickable

	status *grbl.Status

	jogXP, jogXM widget.Clickable
	jogYP, jogYM widget.Clickable
	jogZP, jogZM widget.Clickable
	jogHome      widget.Clickable
	jogStepBtns  [4]widget.Clickable
	jogStep      float64

	fileLines   []string
	sending     bool
	paused      bool
	acked       atomic.Int64
	total       int64
	errCount    atomic.Int64
	sendFileBtn widget.Clickable
	pauseBtn    widget.Clickable
	stopBtn     widget.Clickable
}

func Run() {
	go run()
	app.Main()
}

func run() {
	a := &App{
		serialCh: make(chan serialEvent, 128),
		jogStep:  1.0,
	}
	a.portEditor.SingleLine = true
	a.portEditor.Submit = true
	a.cmdEditor.SingleLine = true
	a.cmdEditor.Submit = true
	a.fileEditor.SingleLine = true
	a.consoleList.Axis = layout.Vertical

	if ports, err := serial.ListPorts(); err == nil {
		for _, p := range ports {
			if p.IsUSB {
				a.portEditor.SetText(p.Name)
				break
			}
		}
	}

	a.win = new(app.Window)
	a.win.Option(app.Title("GoLaserCut"), app.Size(unit.Dp(1200), unit.Dp(800)))
	a.theme = material.NewTheme()
	a.theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	a.log("GoLaserCut ready. Connect to a GRBL device to begin.")

	var ops op.Ops
	for {
		switch e := a.win.Event().(type) {
		case app.DestroyEvent:
			a.disconnect()
			os.Exit(0)
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.processSerialEvents()
			a.handleInputs(gtx)
			a.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (a *App) processSerialEvents() {
	for {
		select {
		case ev := <-a.serialCh:
			switch ev.kind {
			case evConnected:
				if ev.err != nil {
					a.log(fmt.Sprintf("connection error: %v", ev.err))
				} else {
					a.conn = ev.conn
					a.connected = true
					a.log(fmt.Sprintf("connected to %s", a.portEditor.Text()))
				}
			case evDisconnected:
				a.conn = nil
				a.connected = false
				if a.sending && a.sender != nil {
					a.sender.Stop()
					a.sending = false
				}
				a.status = nil
				a.log("disconnected")
			case evLine:
				switch r := ev.resp.(type) {
				case *grbl.OK:
					if a.sending && a.sender != nil {
						a.sender.Ack()
						a.acked.Add(1)
					} else {
						a.log("ok")
					}
				case *grbl.Error:
					if a.sending && a.sender != nil {
						a.sender.Ack()
						a.acked.Add(1)
						a.errCount.Add(1)
					}
					a.log(r.Error())
				case *grbl.Alarm:
					if a.sending && a.sender != nil {
						a.sender.Stop()
						a.sending = false
					}
					a.log(r.Error())
				case *grbl.Status:
					a.status = r
				case *grbl.Welcome:
					a.log(fmt.Sprintf("Grbl %s", r.Version))
				case *grbl.FeedbackMsg:
					a.log(fmt.Sprintf("[MSG: %s]", r.Text))
				case *grbl.Setting:
					a.log(fmt.Sprintf("$%d=%s", r.Key, r.Value))
				default:
					if ev.line != "" {
						a.log(ev.line)
					}
				}
			case evSendDone:
				a.sending = false
				n := a.acked.Load()
				e := a.errCount.Load()
				msg := fmt.Sprintf("send complete: %d/%d lines", n, a.total)
				if e > 0 {
					msg += fmt.Sprintf(" (%d errors)", e)
				}
				a.log(msg)
			}
		default:
			return
		}
	}
}

func (a *App) handleInputs(gtx C) {
	for a.connectBtn.Clicked(gtx) {
		if a.connected {
			a.disconnect()
		} else {
			a.connect(a.portEditor.Text())
		}
	}

	for {
		ev, ok := a.cmdEditor.Update(gtx)
		if !ok {
			break
		}
		if e, ok := ev.(widget.SubmitEvent); ok {
			a.sendCommand(e.Text)
			a.cmdEditor.SetText("")
		}
	}

	for a.sendCmdBtn.Clicked(gtx) {
		cmd := a.cmdEditor.Text()
		if cmd != "" {
			a.sendCommand(cmd)
			a.cmdEditor.SetText("")
		}
	}

	for a.loadBtn.Clicked(gtx) {
		path := strings.TrimSpace(a.fileEditor.Text())
		if path != "" {
			a.loadFile(path)
		}
	}

	for a.sendFileBtn.Clicked(gtx) {
		if a.connected && len(a.fileLines) > 0 && !a.sending {
			a.startSend()
		}
	}

	for a.pauseBtn.Clicked(gtx) {
		if a.sending && a.conn != nil {
			a.paused = !a.paused
			if a.paused {
				a.conn.WriteRealtime(serial.CmdFeedHold)
				a.sender.Pause()
			} else {
				a.conn.WriteRealtime(serial.CmdCycleResume)
				a.sender.Resume()
			}
		}
	}

	for a.stopBtn.Clicked(gtx) {
		if a.sending && a.conn != nil {
			a.conn.WriteRealtime(serial.CmdSoftReset)
			a.sender.Stop()
			a.sending = false
			a.paused = false
			a.log("send stopped")
		}
	}

	type jogCmd struct {
		btn  *widget.Clickable
		axis string
		dir  float64
	}
	for _, jc := range []jogCmd{
		{&a.jogXP, "X", 1}, {&a.jogXM, "X", -1},
		{&a.jogYP, "Y", 1}, {&a.jogYM, "Y", -1},
		{&a.jogZP, "Z", 1}, {&a.jogZM, "Z", -1},
	} {
		for jc.btn.Clicked(gtx) {
			if a.connected && a.conn != nil {
				dist := a.jogStep * jc.dir
				a.conn.WriteLine(fmt.Sprintf("$J=G91 %s%.3f F1000", jc.axis, dist))
			}
		}
	}
	for a.jogHome.Clicked(gtx) {
		if a.connected && a.conn != nil {
			a.conn.WriteLine("$H")
		}
	}
	steps := [4]float64{0.1, 1, 10, 100}
	for i := range a.jogStepBtns {
		for a.jogStepBtns[i].Clicked(gtx) {
			a.jogStep = steps[i]
		}
	}
}

func (a *App) log(s string) {
	a.logLines = append(a.logLines, s)
	a.consoleList.Position.First = len(a.logLines)
}

func (a *App) connect(portName string) {
	if portName == "" {
		a.log("no port specified")
		return
	}
	a.log(fmt.Sprintf("connecting to %s...", portName))
	go func() {
		conn, err := serial.Open(portName)
		if err != nil {
			a.serialCh <- serialEvent{kind: evConnected, err: err}
			a.win.Invalidate()
			return
		}
		a.serialCh <- serialEvent{kind: evConnected, conn: conn}
		a.win.Invalidate()

		for line := range conn.Lines {
			resp := grbl.Parse(line)
			a.serialCh <- serialEvent{kind: evLine, line: line, resp: resp}
			a.win.Invalidate()
		}
		a.serialCh <- serialEvent{kind: evDisconnected}
		a.win.Invalidate()
	}()
}

func (a *App) disconnect() {
	if a.sending && a.sender != nil {
		a.sender.Stop()
	}
	if a.conn != nil {
		a.conn.Close()
	}
}

func (a *App) sendCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	if !a.connected || a.conn == nil {
		a.log("not connected")
		return
	}
	switch cmd {
	case "?":
		a.conn.WriteRealtime(serial.CmdStatusReport)
		return
	case "!":
		a.conn.WriteRealtime(serial.CmdFeedHold)
		return
	case "~":
		a.conn.WriteRealtime(serial.CmdCycleResume)
		return
	}
	a.log(fmt.Sprintf("> %s", cmd))
	if err := a.conn.WriteLine(cmd); err != nil {
		a.log(fmt.Sprintf("send error: %v", err))
	}
}

func (a *App) loadFile(path string) {
	lines, err := gcode.ReadFile(path)
	if err != nil {
		a.log(fmt.Sprintf("load error: %v", err))
		return
	}
	a.fileLines = lines
	a.log(fmt.Sprintf("loaded %d lines from %s", len(lines), path))
}

func (a *App) startSend() {
	a.sender = serial.NewSender(a.conn)
	a.sending = true
	a.paused = false
	a.acked.Store(0)
	a.errCount.Store(0)
	a.total = int64(len(a.fileLines))
	a.log(fmt.Sprintf("sending %d lines...", a.total))

	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			a.mu.Lock()
			c := a.conn
			a.mu.Unlock()
			if !a.sending || c == nil {
				return
			}
			c.WriteRealtime(serial.CmdStatusReport)
			a.win.Invalidate()
		}
	}()

	go func() {
		for _, line := range a.fileLines {
			if err := a.sender.Send(line); err != nil {
				break
			}
		}
		for a.acked.Load() < a.total {
			a.mu.Lock()
			s := a.sending
			a.mu.Unlock()
			if !s {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		a.serialCh <- serialEvent{kind: evSendDone}
		a.win.Invalidate()
	}()
}
