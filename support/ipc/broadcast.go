package ipc

import (
	"fmt"
	"sync"

	"github.com/satori/go.uuid"
)

// Subscriptions is a set of subscribers
type Subscriptions map[uuid.UUID]chan *Message

// MessageEmitter broadcasts message to subscriber
type MessageEmitter struct {
	buffSize    int
	subscribers map[uuid.UUID]chan *Message
	mtx         *sync.RWMutex
}

// HasSubscriber checks if subscriber exists
func (m *MessageEmitter) HasSubscriber(id uuid.UUID) bool {
	m.mtx.RLock()
	_, ok := m.subscribers[id]
	m.mtx.RUnlock()
	return ok
}

// Emit sends message to subscriber
func (m *MessageEmitter) Emit(id uuid.UUID, msg *Message) error {
	if !m.HasSubscriber(id) {
		return fmt.Errorf("no subscriber for UUID '%s'", id.String())
	}

	m.subscribers[id] <- msg
	return nil
}

// Subscribe subscribes for messages and returns a subscription channel
func (m *MessageEmitter) Subscribe(id uuid.UUID) (chan *Message, error) {
	if m.HasSubscriber(id) {
		return nil, fmt.Errorf("subscriber '%s' already exists", id.String())
	}

	m.mtx.Lock()
	m.subscribers[id] = make(chan *Message, m.buffSize)
	m.mtx.Unlock()
	return m.subscribers[id], nil
}

// Unsubscribe removes subscription and closes subscription channel
func (m *MessageEmitter) Unsubscribe(id uuid.UUID) error {
	if !m.HasSubscriber(id) {
		return fmt.Errorf("no such subscriber '%s'", id.String())
	}

	m.mtx.Lock()
	close(m.subscribers[id])
	delete(m.subscribers, id)
	m.mtx.Unlock()
	return nil
}

// RemoveAll removes all subscriptions
func (m *MessageEmitter) RemoveAll() {
	m.mtx.Lock()
	for k := range m.subscribers {
		close(m.subscribers[k])
		delete(m.subscribers, k)
	}
	m.mtx.Unlock()
}

// NewMessageEmitter creates a new message emitter
func NewMessageEmitter(buffSize int) MessageEmitter {
	return MessageEmitter{
		buffSize:    buffSize,
		mtx:         &sync.RWMutex{},
		subscribers: make(Subscriptions),
	}
}
