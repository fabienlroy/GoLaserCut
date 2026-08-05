package serial

import (
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type PortInfo struct {
	Name         string
	IsUSB        bool
	VID          string
	PID          string
	SerialNumber string
}

func ListPorts() ([]*PortInfo, error) {
	details, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("listing serial ports: %w", err)
	}
	var ports []*PortInfo
	for _, d := range details {
		ports = append(ports, &PortInfo{
			Name:         d.Name,
			IsUSB:        d.IsUSB,
			VID:          d.VID,
			PID:          d.PID,
			SerialNumber: d.SerialNumber,
		})
	}
	return ports, nil
}

func DetectGRBL() (string, error) {
	ports, err := ListPorts()
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, p := range ports {
		if p.IsUSB {
			candidates = append(candidates, p.Name)
		}
	}
	for _, p := range ports {
		if !p.IsUSB {
			candidates = append(candidates, p.Name)
		}
	}

	for _, name := range candidates {
		if probeGRBL(name) {
			return name, nil
		}
	}
	return "", fmt.Errorf("no GRBL device found on %d port(s)", len(candidates))
}

func probeGRBL(portName string) bool {
	mode := &serial.Mode{
		BaudRate: DefaultBaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return false
	}
	defer port.Close()

	port.SetReadTimeout(2 * time.Second)

	port.Write([]byte{CmdSoftReset})
	time.Sleep(500 * time.Millisecond)

	buf := make([]byte, 256)
	n, err := port.Read(buf)
	if err != nil || n == 0 {
		return false
	}

	return strings.Contains(string(buf[:n]), "Grbl")
}
