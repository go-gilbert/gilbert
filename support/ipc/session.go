package ipc

import (
	"context"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
)

const msgBuffSz = 10

type Session struct {
	emitter  MessageEmitter
	id       uuid.UUID
	gw       *Gateway
	ctx      context.Context
	cancelFn context.CancelFunc
}

func (s *Session) Open() error {
	if s == nil {
		return errors.New("session is nil")
	}

	ch, err := s.gw.Subscribe(s.id)
	if err != nil {
		return fmt.Errorf("failed to open session, %s", err)
	}

	go s.listen(ch)
	return nil
}

func (s *Session) listen(ch chan *Message) {
	select {
	case msg := <-ch:
		// TODO: decide what to do with the error
		go func() {
			_ = s.emitter.Emit(s.id, msg)
		}()
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
		emitter:  NewMessageEmitter(msgBuffSz),
		id:       uuid.NewV4(),
		gw:       gw,
		ctx:      ctx,
		cancelFn: cancelFn,
	}
}
