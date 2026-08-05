package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fabienlroy/GoLaserCut/grbl"
	"github.com/fabienlroy/GoLaserCut/serial"
)

func main() {
	port := flag.String("port", "", "serial port (auto-detect if empty)")
	list := flag.Bool("list", false, "list available serial ports")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: golasercut [flags] [file.gcode]\n\n")
		fmt.Fprintf(os.Stderr, "GRBL laser cutter controller.\n\n")
		fmt.Fprintf(os.Stderr, "Without a file, starts an interactive console.\n")
		fmt.Fprintf(os.Stderr, "With a .gcode file, sends it with progress tracking.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *list {
		listPorts()
		return
	}

	portName := *port
	if portName == "" {
		fmt.Println("detecting GRBL device...")
		detected, err := serial.DetectGRBL()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			fmt.Fprintln(os.Stderr, "use -port to specify, or -list to see available ports")
			os.Exit(1)
		}
		portName = detected
		fmt.Printf("found GRBL on %s\n", portName)
	}

	conn, err := serial.Open(portName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	w, err := waitBanner(conn, 5*time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	} else {
		fmt.Printf("Grbl %s\n", w.Version)
	}

	if flag.NArg() > 0 {
		err = sendFile(conn, flag.Arg(0))
	} else {
		err = console(conn)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func listPorts() {
	ports, err := serial.ListPorts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(ports) == 0 {
		fmt.Println("no serial ports found")
		return
	}
	for _, p := range ports {
		extra := ""
		if p.IsUSB {
			extra = fmt.Sprintf(" [USB %s:%s]", p.VID, p.PID)
		}
		fmt.Printf("  %s%s\n", p.Name, extra)
	}
}

func waitBanner(conn *serial.Connection, timeout time.Duration) (*grbl.Welcome, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case line, ok := <-conn.Lines:
			if !ok {
				return nil, fmt.Errorf("connection closed")
			}
			if w, ok := grbl.Parse(line).(*grbl.Welcome); ok {
				return w, nil
			}
		case <-timer.C:
			return nil, fmt.Errorf("no GRBL banner within %s", timeout)
		}
	}
}
