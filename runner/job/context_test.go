package job

import (
	"context"
	"errors"
	"fmt"
	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/log"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"sync"
	"testing"
	"time"
)

func TestRunContext_Result(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, &testLog{t: t})
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
	rtx := NewRunContext(context.Background(), nil, &testLog{t: t})
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
	baseTx := NewRunContext(context.Background(), nil, &testLog{t: t})
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
	ctx := NewRunContext(context.Background(), vars, &testLog{t: t})
	child := ctx.ForkContext().(*RunContext)

	assert.Equal(t, ctx.context, child.context)
	assert.Equal(t, ctx.RootVars, child.RootVars)
	assert.Equal(t, ctx.Error, child.Error)
	assert.Equal(t, ctx.wg, child.wg)
	assert.Equal(t, true, child.IsChild())
}

func TestRunContext_Cancel(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, &testLog{t: t})
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
	l := &testLog{t: t}
	rtx := NewRunContext(context.Background(), nil, l)
	assert.Equal(t, l, rtx.Log())
}

func TestRunContext_Timeout(t *testing.T) {
	rtx := NewRunContext(context.Background(), nil, &testLog{t: t})
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

////////////////////////
//   Test Fixtures    //
////////////////////////

type testLog struct {
	t        *testing.T
	messages []string
}

func (c *testLog) log(level int, args ...interface{}) {
	c.t.Log(args...)
}

func (c *testLog) logf(level int, format string, args ...interface{}) {
	c.t.Logf(format, args...)
}

func (c *testLog) SubLogger() sdk.Logger {
	return c
}

func (c *testLog) Format(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (c *testLog) Log(args ...interface{}) {
	c.log(log.LevelMsg, args...)
}

func (c *testLog) Logf(format string, args ...interface{}) {
	c.logf(log.LevelMsg, format, args...)
}

func (c *testLog) Debug(args ...interface{}) {
	c.log(log.LevelDebug, args...)
}

func (c *testLog) Debugf(format string, args ...interface{}) {
	c.logf(log.LevelDebug, format, args...)
}

func (c *testLog) Warn(args ...interface{}) {
	c.log(log.LevelWarn, args...)
}

func (c *testLog) Warnf(format string, args ...interface{}) {
	c.logf(log.LevelWarn, format, args...)
}

func (c *testLog) Info(args ...interface{}) {
	c.log(log.LevelInfo, args...)
}

func (c *testLog) Infof(format string, args ...interface{}) {
	c.logf(log.LevelInfo, format, args...)
}

func (c *testLog) Success(args ...interface{}) {
	c.log(log.LevelSuccess, args...)
}

func (c *testLog) Successf(format string, args ...interface{}) {
	c.logf(log.LevelSuccess, format, args...)
}

func (c *testLog) Error(args ...interface{}) {
	c.log(log.LevelError, args...)
}

func (c *testLog) Errorf(format string, args ...interface{}) {
	c.logf(log.LevelError, format, args...)
}

func (c *testLog) Write(data []byte) (int, error) {
	return fmt.Fprint(os.Stdout, data)
}

func (c *testLog) ErrorWriter() io.Writer {
	return os.Stderr
}
