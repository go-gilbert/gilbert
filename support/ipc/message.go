package ipc

import (
	"encoding/json"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

const (
	Request = iota
	Notify
	Response
)

type Values map[string]interface{}

// Message is IPC message
type Message struct {
	ID        uuid.UUID       `json:"id"`
	SessionID uuid.UUID       `json:"sid"`
	Type      int             `json:"type"`
	Error     *string         `json:"error,omitempty"`
	Method    *string         `json:"method,omitempty"`
	Params    json.RawMessage `json:"params,omitempty"`
	Result    json.RawMessage `json:"result, omitempty"`
}

// Response creates a response for this message message
func (m *Message) Response(result interface{}, err error) (*Message, error) {
	out := &Message{
		ID:        m.ID,
		SessionID: m.SessionID,
		Type:      Response,
	}

	if result != nil {
		r, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to construct message response. %s", err)
		}

		out.Result = r
	}

	if err != nil {
		e := err.Error()
		out.Error = &e
	}

	return out, nil
}
