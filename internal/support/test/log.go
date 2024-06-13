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
	"github.com/go-gilbert/gilbert/internal/log"
)

////////////////////////
//   Test Fixtures    //
////////////////////////

// Log implements sdk.Log
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

// SubLogger implements Log.SubLogger
func (c *Log) SubLogger() sdk.Logger {
	return c
}

// Format implements Log.Format
func (c *Log) Format(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// Log implements Log.Log
func (c *Log) Log(args ...interface{}) {
	c.log(log.LevelMsg, args...)
}

// Logf implements Log.Logf
func (c *Log) Logf(format string, args ...interface{}) {
	c.logf(log.LevelMsg, format, args...)
}

// Debug implements Log.Debug
func (c *Log) Debug(args ...interface{}) {
	c.log(log.LevelDebug, args...)
}

// Debugf implements Log.Debugf
func (c *Log) Debugf(format string, args ...interface{}) {
	c.logf(log.LevelDebug, format, args...)
}

// Warn implements Log.Warn
func (c *Log) Warn(args ...interface{}) {
	c.log(log.LevelWarn, args...)
}

// Warnf implements Log.Warnf
func (c *Log) Warnf(format string, args ...interface{}) {
	c.logf(log.LevelWarn, format, args...)
}

// Info implements Log.Info
func (c *Log) Info(args ...interface{}) {
	c.log(log.LevelInfo, args...)
}

// Infof implements Log.Infof
func (c *Log) Infof(format string, args ...interface{}) {
	c.logf(log.LevelInfo, format, args...)
}

// Success implements Log.Success
func (c *Log) Success(args ...interface{}) {
	c.log(log.LevelSuccess, args...)
}

// Successf implements Log.Successf
func (c *Log) Successf(format string, args ...interface{}) {
	c.logf(log.LevelSuccess, format, args...)
}

// Error implements Log.Error
func (c *Log) Error(args ...interface{}) {
	c.log(log.LevelError, args...)
}

// Errorf implements Log.Errorf
func (c *Log) Errorf(format string, args ...interface{}) {
	c.logf(log.LevelError, format, args...)
}

// Write implements Log.Write
func (c *Log) Write(data []byte) (int, error) {
	return fmt.Fprint(os.Stdout, data)
}

// ErrorWriter implements Log.ErrorWriter
func (c *Log) ErrorWriter() io.Writer {
	return os.Stderr
}
