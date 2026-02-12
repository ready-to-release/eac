package tui

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// messagePump delivers tea.Msg values to a bubbletea program using a two-channel
// priority system: critical messages (state changes) are drained before regular
// messages (output lines) to maintain ordering guarantees.
type messagePump struct {
	program      *tea.Program
	msgChan      chan tea.Msg
	criticalChan chan tea.Msg
	wg           sync.WaitGroup
	closeOnce    sync.Once
}

// newMessagePump creates a message pump that delivers to the given program.
func newMessagePump(program *tea.Program) *messagePump {
	return &messagePump{
		program:      program,
		msgChan:      make(chan tea.Msg, 500),
		criticalChan: make(chan tea.Msg, 10000),
	}
}

// Start begins the message pump goroutine.
func (mp *messagePump) Start() {
	mp.wg.Add(1)
	go mp.run()
}

// run is the main pump loop that prioritizes critical messages over regular ones.
func (mp *messagePump) run() {
	defer mp.wg.Done()
	for {
		select {
		case msg, ok := <-mp.criticalChan:
			if !ok {
				for msg := range mp.msgChan {
					mp.program.Send(msg)
				}
				return
			}
			mp.program.Send(msg)
		default:
			select {
			case msg, ok := <-mp.criticalChan:
				if !ok {
					for msg := range mp.msgChan {
						mp.program.Send(msg)
					}
					return
				}
				mp.program.Send(msg)
			case msg, ok := <-mp.msgChan:
				if !ok {
					for msg := range mp.criticalChan {
						mp.program.Send(msg)
					}
					return
				}
				mp.program.Send(msg)
			}
		}
	}
}

// SendAsync queues a regular message (may be dropped if buffer full).
func (mp *messagePump) SendAsync(msg tea.Msg) {
	select {
	case mp.msgChan <- msg:
	default:
		// Buffer full - drop message
	}
}

// SendCritical queues a critical message that should not be dropped.
func (mp *messagePump) SendCritical(msg tea.Msg) {
	select {
	case mp.criticalChan <- msg:
	default:
		// Fall back to regular channel
		mp.SendAsync(msg)
	}
}

// Close closes both channels and waits for the pump to drain.
func (mp *messagePump) Close() {
	mp.closeOnce.Do(func() {
		close(mp.criticalChan)
		close(mp.msgChan)
	})
	mp.wg.Wait()
}
