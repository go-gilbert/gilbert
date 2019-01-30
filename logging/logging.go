package logging

import "io"

var (
	Log Logger
)

type Logger interface {
	SubLogger() Logger
	Sprintf(message string, args ...interface{}) string
	Log(message string, args ...interface{})
	Debug(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	Info(message string, args ...interface{})
	Success(message string, args ...interface{})
	Write(data []byte) (int, error)
	ErrorWriter() io.Writer
}
