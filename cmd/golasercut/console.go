package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fabienlroy/GoLaserCut/grbl"
	"github.com/fabienlroy/GoLaserCut/serial"
)

func console(conn *serial.Connection) error {
	go func() {
		for line := range conn.Lines {
			printResponse(grbl.Parse(line), line)
		}
		fmt.Println("connection closed")
		os.Exit(0)
	}()

	fmt.Println("type commands, '?' for status, 'quit' to exit")
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("grbl> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			return nil
		}

		switch input {
		case "?":
			conn.WriteRealtime(serial.CmdStatusReport)
		case "!":
			conn.WriteRealtime(serial.CmdFeedHold)
		case "~":
			conn.WriteRealtime(serial.CmdCycleResume)
		default:
			if err := conn.WriteLine(input); err != nil {
				fmt.Fprintf(os.Stderr, "send error: %v\n", err)
			}
		}
	}
	return scanner.Err()
}

func printResponse(resp grbl.Response, raw string) {
	switch r := resp.(type) {
	case *grbl.OK:
		fmt.Println("ok")
	case *grbl.Error:
		if r.Message != "" {
			fmt.Printf("error:%d (%s)\n", r.Code, r.Message)
		} else {
			fmt.Printf("error:%d\n", r.Code)
		}
	case *grbl.Alarm:
		if r.Message != "" {
			fmt.Printf("ALARM:%d (%s)\n", r.Code, r.Message)
		} else {
			fmt.Printf("ALARM:%d\n", r.Code)
		}
	case *grbl.Status:
		fmt.Printf("[%s]", r.State)
		if r.MPos != nil {
			fmt.Printf(" MPos:%.3f,%.3f,%.3f", r.MPos.X, r.MPos.Y, r.MPos.Z)
		} else if r.WPos != nil {
			fmt.Printf(" WPos:%.3f,%.3f,%.3f", r.WPos.X, r.WPos.Y, r.WPos.Z)
		}
		fmt.Printf(" F:%.0f S:%.0f", r.Feed, r.Spindle)
		if r.Overrides != nil {
			fmt.Printf(" Ov:%d,%d,%d", r.Overrides.Feed, r.Overrides.Rapids, r.Overrides.Spindle)
		}
		fmt.Println()
	case *grbl.FeedbackMsg:
		fmt.Printf("[MSG: %s]\n", r.Text)
	case *grbl.Setting:
		fmt.Printf("$%d=%s\n", r.Key, r.Value)
	case *grbl.Welcome:
		fmt.Printf("Grbl %s\n", r.Version)
	default:
		fmt.Println(raw)
	}
}
