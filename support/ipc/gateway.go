package ipc

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
)

type Gateway struct {
	MessageEmitter
	Messages chan *Message
	writer   io.Writer
}

// Write implements io.Writer and used to receive messages from plugin process
func (g *Gateway) Write(p []byte) (int, error) {
	l := len(p)
	msg := new(Message)
	err := json.Unmarshal(p, msg)
	if err == nil {
		return l, g.pushMessage(msg.SessionID, msg)
	}

	return l, err
}

func (g *Gateway) pushMessage(id uuid.UUID, msg *Message) error {
	if g.MessageEmitter.HasSubscriber(id) {
		return g.MessageEmitter.Emit(id, msg)
	}

	// Post to common channel if nobody subscribed to it
	g.Messages <- msg
	return nil
}

func (g *Gateway) Send(m *Message) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to send message '%s', %s", m.ID, err)
		}
	}()

	var out []byte
	out, err = json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to send message '%s', %s", m.ID, err)
	}

	_, err = g.writer.Write(out)
	return err
}

func (g *Gateway) Close() {
	g.MessageEmitter.RemoveAll()
	close(g.Messages)
}

func NewGateway(w io.Writer, poolSize int) *Gateway {
	return &Gateway{
		MessageEmitter: NewMessageEmitter(poolSize),
		Messages:       make(chan *Message, poolSize),
		writer:         w,
	}
}
