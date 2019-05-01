package ipc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/satori/go.uuid"
)

const msgBuffSz = 10

type Session struct {
	RequestListener
	emitter  MessageEmitter
	id       uuid.UUID
	gw       *Gateway
	ctx      context.Context
	cancelFn context.CancelFunc
}

func (s *Session) ID() uuid.UUID {
	return s.id
}

func (s *Session) Open() error {
	if s == nil {
		return errors.New("session is nil")
	}

	ch, err := s.gw.Subscribe(s.id)
	if err != nil {
		return fmt.Errorf("failed to open session '%s', %s", s.id, err)
	}

	go s.listen(ch)
	return nil
}

func (s *Session) Notify(methodName string, args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("call: panic - %s", r)
		}
	}()

	msg, err := s.newMsgRequest(false, methodName, args...)
	if err != nil {
		return err
	}

	// send message
	return s.gw.Send(msg)
}

// Call calls a method and returns response
func (s *Session) Call(out interface{}, methodName string, args ...interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("call: panic - %s", r)
		}
	}()

	msg, err := s.newMsgRequest(false, methodName, args...)
	if err != nil {
		return fmt.Errorf("call: failed to construct message: %s", err)
	}

	// subscribe for response
	ch, err := s.emitter.Subscribe(msg.ID)
	defer s.emitter.Unsubscribe(msg.ID) // nolint:errcheck
	if err != nil {
		return err
	}

	// send message
	if err := s.gw.Send(msg); err != nil {
		return fmt.Errorf("call: error on message send: %s", err)
	}

	result, ok := <-ch
	if !ok {
		return errors.New("call: cannot get response, result channel was closed")
	}

	if result.Error != nil {
		return errors.New(*result.Error)
	}

	if err = json.Unmarshal(result.Result, out); err != nil {
		return fmt.Errorf("call: failed to unmarshal response, %s", err)
	}

	return nil
}

func (s *Session) newMsgRequest(async bool, methodName string, args ...interface{}) (msg *Message, err error) {
	msg = &Message{
		ID:        uuid.NewV4(),
		SessionID: s.id,
		Method:    &methodName,
	}

	if async {
		msg.Type = Notify
	} else {
		msg.Type = Request
	}

	if len(args) == 1 {
		msg.Params, err = json.Marshal(args[1])
	} else {
		msg.Params, err = json.Marshal(args)
	}

	return msg, err
}

func (s *Session) listen(ch chan *Message) {
	select {
	case msg := <-ch:
		// TODO: handle errors
		switch msg.Type {
		case Request:
			// request - process and send response
			resp, _ := s.handleMessage(msg)
			_ = s.gw.Send(resp)
		case Notify:
			// notify - just handle the action
			_, _ = s.handleMessage(msg)
		case Response:
			// response - broadcast response
			_ = s.emitter.Emit(s.id, msg)
		}
	case <-s.ctx.Done():
		_ = s.gw.Unsubscribe(s.id)
		s.emitter.RemoveAll()
	}
}

func (s *Session) Close() {
	if s.cancelFn != nil {
		s.cancelFn()
	}
}

func NewSession(gw *Gateway, parentCtx context.Context) *Session {
	ctx, cancelFn := context.WithCancel(parentCtx)
	return &Session{
		RequestListener: NewRequestListener(),
		emitter:         NewMessageEmitter(msgBuffSz),
		id:              uuid.NewV4(),
		gw:              gw,
		ctx:             ctx,
		cancelFn:        cancelFn,
	}
}
