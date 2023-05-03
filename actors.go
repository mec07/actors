package actors

import (
	"errors"
	"sync"
	"time"

	"github.com/edwingeng/deque"
	"github.com/mec07/async"
)

var ErrActorAlreadyStopped = errors.New("actor has already been stoped")

// Receiver is the interface that the actors must satisfy.
type Receiver interface {
	Receive(message interface{})
}

// ReceiverFunc satisfies the Receiver interface (very similar to
// http.HandlerFunc.
type ReceiverFunc func(message interface{})

// Receive makes ReceiverFunc implement the Receiver interface.
func (f ReceiverFunc) Receive(message interface{}) {
	f(message)
}

// PoisonPill is a message used to shut down the actor system. This shutdown
// only happens once the actor system has processed all the messages already in
// the queue.
type PoisonPill struct{}

// PID is a process identifier for an actor and is used to send messages to it.
type PID struct {
	mu       sync.RWMutex
	inbox    deque.Deque
	stopped  bool
	receiver Receiver
}

// Send is a way of passing messages to actors in a non-blocking way.
func (p *PID) Send(message interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return ErrActorAlreadyStopped
	}

	p.inbox.PushBack(message)
	return nil
}

func (p *PID) startMonitoringInbox() {
	for {
		message := p.inbox.PopFront()
		if message == nil {
			time.Sleep(time.Millisecond)
		}
		if _, ok := message.(PoisonPill); ok {
			p.setStopped()
			return
		}
		p.receiver.Receive(message)
	}
}

func (p *PID) setStopped() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stopped = true
}

func SpawnSystem(receiver Receiver) *PID {
	actor := &PID{
		inbox:    deque.NewDeque(),
		receiver: receiver,
	}
	async.Go(actor.startMonitoringInbox, nil)
	return actor
}
