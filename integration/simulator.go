package integration

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

// GRBLSimulator emulates a GRBL 1.1 controller over an io.ReadWriter.
// It processes G-code lines and responds with ok/error, handles
// realtime commands (?, !, ~, 0x18), and tracks machine position.
type GRBLSimulator struct {
	pw     *io.PipeWriter
	pr     *io.PipeReader
	input  *io.PipeReader
	output *io.PipeWriter

	mu       sync.Mutex
	state    string
	x, y, z  float64
	feed     float64
	spindle  float64
	laser    bool
	lines    int
	errors   int
	done     chan struct{}
}

// NewGRBLSimulator creates a simulator. Returns (portForClient, simulator).
// The client reads/writes the portIO; the simulator processes commands internally.
func NewGRBLSimulator() (*SimPort, *GRBLSimulator) {
	clientRead, simWrite := io.Pipe()
	simRead, clientWrite := io.Pipe()

	sim := &GRBLSimulator{
		pw:    simWrite,
		pr:    simRead,
		input: simRead,
		output: simWrite,
		state: "Idle",
		done:  make(chan struct{}),
	}

	port := &SimPort{
		reader: clientRead,
		writer: clientWrite,
	}

	go sim.run()

	return port, sim
}

// SimPort implements the portIO interface for use with serial.Connection.
type SimPort struct {
	reader *io.PipeReader
	writer *io.PipeWriter
	closed bool
	mu     sync.Mutex
}

func (p *SimPort) Read(buf []byte) (int, error)  { return p.reader.Read(buf) }
func (p *SimPort) Write(buf []byte) (int, error) { return p.writer.Write(buf) }
func (p *SimPort) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	p.reader.Close()
	p.writer.Close()
	return nil
}

func (sim *GRBLSimulator) run() {
	defer close(sim.done)

	sim.send("Grbl 1.1h ['$' for help]\r\n")

	reader := bufio.NewReader(sim.pr)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return
		}

		switch b {
		case '?':
			sim.handleStatusQuery()
		case '!':
			sim.mu.Lock()
			if sim.state == "Run" {
				sim.state = "Hold"
			}
			sim.mu.Unlock()
		case '~':
			sim.mu.Lock()
			if sim.state == "Hold" {
				sim.state = "Run"
			}
			sim.mu.Unlock()
		case 0x18: // soft reset
			sim.mu.Lock()
			sim.state = "Idle"
			sim.x, sim.y, sim.z = 0, 0, 0
			sim.feed = 0
			sim.spindle = 0
			sim.laser = false
			sim.mu.Unlock()
			sim.send("Grbl 1.1h ['$' for help]\r\n")
		case '\n':
			// empty line, ignore
		default:
			reader.UnreadByte()
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line != "" {
				sim.processLine(line)
			}
		}
	}
}

func (sim *GRBLSimulator) processLine(line string) {
	upper := strings.ToUpper(strings.TrimSpace(line))

	if upper == "" {
		return
	}

	// Settings query
	if upper == "$$" {
		sim.send("$0=10\r\n")
		sim.send("$30=1000\r\n")
		sim.send("$32=1\r\n")
		sim.send("ok\r\n")
		return
	}

	// Check mode
	if upper == "$C" {
		sim.mu.Lock()
		if sim.state == "Check" {
			sim.state = "Idle"
		} else {
			sim.state = "Check"
		}
		sim.mu.Unlock()
		sim.send("ok\r\n")
		return
	}

	sim.mu.Lock()
	sim.state = "Run"
	sim.mu.Unlock()

	if !sim.executeGCode(upper) {
		sim.mu.Lock()
		sim.errors++
		sim.mu.Unlock()
		sim.send("error:20\r\n") // unsupported command
		return
	}

	sim.mu.Lock()
	sim.lines++
	sim.state = "Idle"
	sim.mu.Unlock()
	sim.send("ok\r\n")
}

func (sim *GRBLSimulator) executeGCode(line string) bool {
	sim.mu.Lock()
	defer sim.mu.Unlock()

	tokens := tokenize(line)
	for _, tok := range tokens {
		if len(tok) < 2 {
			continue
		}
		letter := tok[0]
		value := tok[1:]

		switch letter {
		case 'G':
			// G0, G1, G4, G17, G21, G28, G90, G91 — all accepted
		case 'M':
			switch value {
			case "2", "30":
				sim.state = "Idle"
				sim.spindle = 0
			case "3", "4":
				sim.laser = true
			case "5":
				sim.laser = false
				sim.spindle = 0
			case "7", "8", "9":
				// coolant
			default:
				return false
			}
		case 'X':
			fmt.Sscanf(value, "%f", &sim.x)
		case 'Y':
			fmt.Sscanf(value, "%f", &sim.y)
		case 'Z':
			fmt.Sscanf(value, "%f", &sim.z)
		case 'F':
			fmt.Sscanf(value, "%f", &sim.feed)
		case 'S':
			fmt.Sscanf(value, "%f", &sim.spindle)
		case 'P':
			// dwell parameter
		case 'N':
			// line number
		default:
			return false
		}
	}
	return true
}

func (sim *GRBLSimulator) handleStatusQuery() {
	sim.mu.Lock()
	status := fmt.Sprintf("<%s|MPos:%.3f,%.3f,%.3f|FS:%.0f,%.0f>\r\n",
		sim.state, sim.x, sim.y, sim.z, sim.feed, sim.spindle)
	sim.mu.Unlock()
	sim.send(status)
}

func (sim *GRBLSimulator) send(s string) {
	sim.pw.Write([]byte(s))
}

// State returns the current machine state.
func (sim *GRBLSimulator) State() string {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	return sim.state
}

// Position returns current XYZ.
func (sim *GRBLSimulator) Position() (x, y, z float64) {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	return sim.x, sim.y, sim.z
}

// LinesProcessed returns the count of successfully processed lines.
func (sim *GRBLSimulator) LinesProcessed() int {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	return sim.lines
}

// Errors returns the count of error responses sent.
func (sim *GRBLSimulator) Errors() int {
	sim.mu.Lock()
	defer sim.mu.Unlock()
	return sim.errors
}

// Close shuts down the simulator.
func (sim *GRBLSimulator) Close() {
	sim.pr.Close()
	sim.pw.Close()
	<-sim.done
}

func tokenize(line string) []string {
	var tokens []string
	current := ""
	for _, c := range line {
		if c == ' ' {
			if current != "" {
				tokens = append(tokens, current)
				current = ""
			}
			continue
		}
		if c >= 'A' && c <= 'Z' && current != "" {
			tokens = append(tokens, current)
			current = string(c)
		} else {
			current += string(c)
		}
	}
	if current != "" {
		tokens = append(tokens, current)
	}
	return tokens
}
