package job

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/support/test"
	"github.com/stretchr/testify/assert"
)

func TestRunContext_Result(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, &test.Log{T: t})
	errs := make(chan error)
	expected := errors.New("test error")
	wg := &sync.WaitGroup{}
	rtx.SetErrorChannel(errs)
	rtx.SetWaitGroup(wg)
	assert.Equal(t, true, rtx.IsAlive())
	go func(e error) {
		wg.Add(1)
		rtx.Result(e)
	}(expected)

	wg.Wait()
	result := <-rtx.Errors()
	assert.Equal(t, expected, result)
}

func TestRunContext_Success(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, &test.Log{T: t})
	wg := &sync.WaitGroup{}
	rtx.SetWaitGroup(wg)
	go func() {
		wg.Add(1)
		time.Sleep(time.Millisecond * 100)
		rtx.Success()
	}()

	wg.Wait()
	result := <-rtx.Errors()
	assert.Nil(t, result)
}

func TestRunContext_ChildContext(t *testing.T) {
	vars := sdk.Vars{"foo": "bar"}
	baseTx := NewRunContext(context.Background(), nil, &test.Log{T: t})
	baseTx.SetVars(vars)
	rtx := baseTx.ChildContext().(*RunContext)

	assert.Equal(t, true, rtx.IsChild())
	assert.Equal(t, true, rtx.IsAlive())
	assert.NotNil(t, rtx.cancelFn)
	assert.Equal(t, vars, rtx.Vars())
	close(rtx.Errors())
}

func TestRunContext_ForkContext(t *testing.T) {
	vars := sdk.Vars{"foo": "bar"}
	ctx := NewRunContext(context.Background(), vars, &test.Log{T: t})
	child := ctx.ForkContext().(*RunContext)

	assert.Equal(t, ctx.context, child.context)
	assert.Equal(t, ctx.RootVars, child.RootVars)
	assert.Equal(t, ctx.Error, child.Error)
	assert.Equal(t, ctx.wg, child.wg)
	assert.Equal(t, true, child.IsChild())
}

func TestRunContext_Cancel(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, &test.Log{T: t})
	canceled := false
	rtx.cancelFn = func() {
		canceled = true
	}

	rtx.Cancel()
	assert.Equal(t, true, canceled)
	rtx.finished = false
	rtx.cancelFn = nil
	rtx.Cancel()
	assert.Equal(t, false, rtx.IsAlive())
	close(rtx.Errors())
}

func TestRunContext_Log(t *testing.T) {
	l := &test.Log{T: t}
	rtx := NewRunContext(context.Background(), nil, l)
	assert.Equal(t, l, rtx.Log())
}

func TestRunContext_Timeout(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, &test.Log{T: t})
	canceled := false
	rtx.cancelFn = func() {
		canceled = true
	}

	rtx.Timeout(time.Millisecond * 100)
	time.Sleep(time.Millisecond * 150)

	assert.Equal(t, true, canceled)
	close(rtx.Errors())
}

func TestRunContext_Context(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, nil)
	assert.NotNil(t, rtx.Context())
	close(rtx.Errors())
}
