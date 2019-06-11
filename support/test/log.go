package test

//nolint

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	sdk "github.com/go-gilbert/gilbert-sdk"
	"github.com/go-gilbert/gilbert/log"
)

////////////////////////
//   Test Fixtures    //
////////////////////////

type Log struct {
	T        *testing.T
	messages []string
	mtx      sync.Mutex
}

func (c *Log) log(level int, args ...interface{}) {
	c.T.Log(args...)
	c.mtx.Lock()
	c.messages = append(c.messages, fmt.Sprint(args...))
	c.mtx.Unlock()
}

func (c *Log) logf(level int, format string, args ...interface{}) {
	c.T.Logf(format, args...)
	c.mtx.Lock()
	c.messages = append(c.messages, fmt.Sprintf(format, args...))
	c.mtx.Unlock()
}

// AssertMessage checks if message has been logged
func (c *Log) AssertMessage(msg string) {
	for _, m := range c.messages {
		if strings.Contains(m, msg) {
			return
		}
	}

	c.T.Errorf("log: expected message '%s' was not logged", msg)
}

func (c *Log) SubLogger() sdk.Logger {
	return c
}

func (c *Log) Format(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func (c *Log) Log(args ...interface{}) {
	c.log(log.LevelMsg, args...)
}

func (c *Log) Logf(format string, args ...interface{}) {
	c.logf(log.LevelMsg, format, args...)
}

func (c *Log) Debug(args ...interface{}) {
	c.log(log.LevelDebug, args...)
}

func (c *Log) Debugf(format string, args ...interface{}) {
	c.logf(log.LevelDebug, format, args...)
}

func (c *Log) Warn(args ...interface{}) {
	c.log(log.LevelWarn, args...)
}

func (c *Log) Warnf(format string, args ...interface{}) {
	c.logf(log.LevelWarn, format, args...)
}

func (c *Log) Info(args ...interface{}) {
	c.log(log.LevelInfo, args...)
}

func (c *Log) Infof(format string, args ...interface{}) {
	c.logf(log.LevelInfo, format, args...)
}

func (c *Log) Success(args ...interface{}) {
	c.log(log.LevelSuccess, args...)
}

func (c *Log) Successf(format string, args ...interface{}) {
	c.logf(log.LevelSuccess, format, args...)
}

func (c *Log) Error(args ...interface{}) {
	c.log(log.LevelError, args...)
}

func (c *Log) Errorf(format string, args ...interface{}) {
	c.logf(log.LevelError, format, args...)
}

func (c *Log) Write(data []byte) (int, error) {
	return fmt.Fprint(os.Stdout, data)
}

func (c *Log) ErrorWriter() io.Writer {
	return os.Stderr
}
