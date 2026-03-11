package services

import "sync"

const defaultBroadcasterBufferSize = 16

type EventBroadcaster struct {
	mu         sync.RWMutex
	nextID     uint64
	bufferSize int
	clients    map[uint64]chan Event
}

func NewEventBroadcaster(bufferSize int) *EventBroadcaster {
	if bufferSize <= 0 {
		bufferSize = defaultBroadcasterBufferSize
	}
	return &EventBroadcaster{
		bufferSize: bufferSize,
		clients:    make(map[uint64]chan Event),
	}
}

func (b *EventBroadcaster) Subscribe() (uint64, <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := b.nextID
	ch := make(chan Event, b.bufferSize)
	b.clients[id] = ch
	return id, ch
}

func (b *EventBroadcaster) Unsubscribe(id uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch, ok := b.clients[id]
	if !ok {
		return
	}
	delete(b.clients, id)
	close(ch)
}

func (b *EventBroadcaster) Emit(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, ch := range b.clients {
		select {
		case ch <- event:
		default:
			delete(b.clients, id)
			close(ch)
		}
	}
}
