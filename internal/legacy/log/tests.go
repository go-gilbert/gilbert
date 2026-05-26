package log

import (
	"testing"
)

type testWriter struct {
	t *testing.T
}

func (t *testWriter) Write(level int, message string) {
	t.t.Log("log: ", message)
}

// UseTestLogger inits test logger
func UseTestLogger(t *testing.T) {
	Default = &logger{
		level:     LevelDebug,
		formatter: &paddingFormatter{},
		writer:    &testWriter{t: t},
	}
}
