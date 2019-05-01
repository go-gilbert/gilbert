package ipc

import (
	"encoding/json"
	"fmt"
	"sync"
)

type MessageHandler func(args json.RawMessage) (interface{}, error)

// RequestListener listens for incoming RPC calls
type RequestListener struct {
	handlers map[string]MessageHandler
	mtx      *sync.RWMutex
}

// handleMessage handles message request
func (m *RequestListener) handleMessage(msg *Message) (resp *Message, err error) {
	method := *msg.Method
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	handler, ok := m.handlers[method]
	if !ok {
		return msg.Response(nil, fmt.Errorf("unknown method %s()", method))
	}

	return msg.Response(handler(msg.Params))
}

// HandleFunc registers message handler
func (m *RequestListener) HandleFunc(methodName string, h MessageHandler) {
	m.mtx.Lock()
	m.handlers[methodName] = h
	m.mtx.Unlock()
}

// RemoveHandler removes handler
func (m *RequestListener) RemoveHandler(methodName string) error {
	m.mtx.Lock()
	delete(m.handlers, methodName)
	m.mtx.Unlock()
	return nil
}

// RemoveAll removes all subscriptions
func (m *RequestListener) Close() {
	m.mtx.Lock()
	for k := range m.handlers {
		delete(m.handlers, k)
	}
	m.mtx.Unlock()
}

// NewRequestListener creates a new request handler
func NewRequestListener() RequestListener {
	return RequestListener{
		mtx:      &sync.RWMutex{},
		handlers: make(map[string]MessageHandler),
	}
}
