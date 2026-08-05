package serial

import (
	"bufio"
	"fmt"
	"strings"
	"sync"

	"go.bug.st/serial"
)

const DefaultBaudRate = 115200

// PortIO is the interface for a serial port or simulator.
type PortIO interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
}

type Connection struct {
	port     PortIO
	portName string
	reader   *bufio.Reader
	mu       sync.Mutex
	closed   bool
	Lines    chan string
	done     chan struct{}
	cancel   chan struct{}
}

func Open(portName string) (*Connection, error) {
	mode := &serial.Mode{
		BaudRate: DefaultBaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", portName, err)
	}
	return NewConnection(port, portName), nil
}

func NewConnection(port PortIO, name string) *Connection {
	c := &Connection{
		port:     port,
		portName: name,
		reader:   bufio.NewReader(port),
		Lines:    make(chan string, 64),
		done:     make(chan struct{}),
		cancel:   make(chan struct{}),
	}
	go c.readLoop()
	return c
}

func (c *Connection) PortName() string {
	return c.portName
}

func (c *Connection) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.cancel)
	c.mu.Unlock()

	err := c.port.Close()
	<-c.done
	return err
}

func (c *Connection) readLoop() {
	defer close(c.done)
	defer close(c.Lines)
	for {
		line, err := c.reader.ReadString('\n')
		if len(line) > 0 {
			trimmed := strings.TrimRight(line, "\r\n")
			if trimmed != "" {
				select {
				case c.Lines <- trimmed:
				case <-c.cancel:
					return
				}
			}
		}
		if err != nil {
			return
		}
	}
}

func (c *Connection) WriteLine(s string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	_, err := c.port.Write([]byte(s + "\n"))
	if err != nil {
		return fmt.Errorf("writing to %s: %w", c.portName, err)
	}
	return nil
}

func (c *Connection) WriteRealtime(cmd byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	_, err := c.port.Write([]byte{cmd})
	if err != nil {
		return fmt.Errorf("writing realtime to %s: %w", c.portName, err)
	}
	return nil
}
