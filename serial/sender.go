package serial

import (
	"fmt"
	"sync"
)

const grblBufferSize = 127

type Sender struct {
	conn     *Connection
	mu       sync.Mutex
	cond     *sync.Cond
	inFlight int
	queue    []int
	paused   bool
	stopped  bool
}

func NewSender(conn *Connection) *Sender {
	s := &Sender{conn: conn}
	s.cond = sync.NewCond(&s.mu)
	return s
}

func (s *Sender) Send(line string) error {
	lineLen := len(line) + 1
	if lineLen > grblBufferSize {
		return fmt.Errorf("line too long (%d bytes, max %d)", lineLen, grblBufferSize)
	}

	s.mu.Lock()
	for s.inFlight+lineLen > grblBufferSize || s.paused {
		if s.stopped {
			s.mu.Unlock()
			return fmt.Errorf("sender stopped")
		}
		s.cond.Wait()
	}
	s.inFlight += lineLen
	s.queue = append(s.queue, lineLen)
	s.mu.Unlock()

	return s.conn.WriteLine(line)
}

func (s *Sender) Ack() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.queue) > 0 {
		s.inFlight -= s.queue[0]
		s.queue = s.queue[1:]
		s.cond.Broadcast()
	}
}

func (s *Sender) Pause() {
	s.mu.Lock()
	s.paused = true
	s.mu.Unlock()
}

func (s *Sender) Resume() {
	s.mu.Lock()
	s.paused = false
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *Sender) Stop() {
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *Sender) Reset() {
	s.mu.Lock()
	s.inFlight = 0
	s.queue = nil
	s.paused = false
	s.stopped = false
	s.mu.Unlock()
	s.cond.Broadcast()
}

func (s *Sender) InFlight() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.inFlight
}

func (s *Sender) Pending() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.queue)
}
