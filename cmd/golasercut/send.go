package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fabienlroy/GoLaserCut/gcode"
	"github.com/fabienlroy/GoLaserCut/grbl"
	"github.com/fabienlroy/GoLaserCut/serial"
)

func sendFile(conn *serial.Connection, filename string) error {
	lines, err := gcode.ReadFile(filename)
	if err != nil {
		return err
	}
	total := int64(len(lines))
	if total == 0 {
		return fmt.Errorf("no G-code lines in %s", filename)
	}
	fmt.Printf("loaded %d lines from %s\n", len(lines), filename)

	sender := serial.NewSender(conn)
	var acked atomic.Int64
	var errCount atomic.Int64
	var lastStatus atomic.Pointer[grbl.Status]

	stop := make(chan struct{})
	var stopOnce sync.Once
	doStop := func() { stopOnce.Do(func() { close(stop) }) }

	allAcked := make(chan struct{})
	var ackOnce sync.Once

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		select {
		case <-sigCh:
			fmt.Fprintf(os.Stderr, "\ninterrupted, stopping machine...\n")
			conn.WriteRealtime(serial.CmdFeedHold)
			time.Sleep(100 * time.Millisecond)
			conn.WriteRealtime(serial.CmdSoftReset)
			sender.Stop()
			doStop()
		case <-stop:
		}
	}()

	responseDone := make(chan struct{})
	go func() {
		defer close(responseDone)
		for line := range conn.Lines {
			switch r := grbl.Parse(line).(type) {
			case *grbl.OK:
				sender.Ack()
				if acked.Add(1) >= total {
					ackOnce.Do(func() { close(allAcked) })
				}
			case *grbl.Error:
				sender.Ack()
				errCount.Add(1)
				fmt.Fprintf(os.Stderr, "\nerror:%d %s\n", r.Code, r.Message)
				if acked.Add(1) >= total {
					ackOnce.Do(func() { close(allAcked) })
				}
			case *grbl.Alarm:
				fmt.Fprintf(os.Stderr, "\nALARM:%d %s\n", r.Code, r.Message)
				sender.Stop()
				doStop()
			case *grbl.Status:
				lastStatus.Store(r)
			}
		}
		doStop()
	}()

	pollDone := make(chan struct{})
	go func() {
		defer close(pollDone)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				conn.WriteRealtime(serial.CmdStatusReport)
			case <-stop:
				return
			}
		}
	}()

	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				n := acked.Load()
				pct := int(n * 100 / total)
				bar := progressBar(pct, 30)
				info := statusInfo(lastStatus.Load())
				fmt.Printf("\r\033[K%s %3d%% (%d/%d)%s", bar, pct, n, total, info)
			case <-stop:
				return
			case <-allAcked:
				return
			}
		}
	}()

	start := time.Now()
	for _, line := range lines {
		if err := sender.Send(line); err != nil {
			break
		}
	}

	select {
	case <-allAcked:
	case <-stop:
	}

	signal.Stop(sigCh)
	doStop()
	<-pollDone
	<-progressDone
	conn.Close()
	<-responseDone

	elapsed := time.Since(start)
	n := acked.Load()
	e := errCount.Load()
	fmt.Printf("\r\033[K")
	fmt.Printf("done: %d/%d lines in %s", n, total, elapsed.Round(time.Millisecond))
	if e > 0 {
		fmt.Printf(" (%d errors)", e)
	}
	fmt.Println()

	return nil
}


func progressBar(pct, width int) string {
	filled := pct * width / 100
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("=", filled)
	if filled < width {
		bar += ">"
		bar += strings.Repeat(" ", width-filled-1)
	}
	return "[" + bar + "]"
}

func statusInfo(s *grbl.Status) string {
	if s == nil {
		return ""
	}
	info := fmt.Sprintf(" | %s", s.State)
	if s.MPos != nil {
		info += fmt.Sprintf(" X:%.1f Y:%.1f Z:%.1f", s.MPos.X, s.MPos.Y, s.MPos.Z)
	} else if s.WPos != nil {
		info += fmt.Sprintf(" X:%.1f Y:%.1f Z:%.1f", s.WPos.X, s.WPos.Y, s.WPos.Z)
	}
	info += fmt.Sprintf(" F:%.0f", s.Feed)
	return info
}
