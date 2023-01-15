package log

import (
	"io"
	"os"
)

type Writer interface {
	Write(level Level, data io.Reader) error
}

// SplitWriter writes error and info logs to separate sinks.
type SplitWriter struct {
	logWriter   io.Writer
	errorWriter io.Writer
}

// NewSplitWriter constructs a new writer that writes errors and logs to separate sinks.
func NewSplitWriter(logWriter io.Writer, errorWriter io.Writer) *SplitWriter {
	return &SplitWriter{logWriter: logWriter, errorWriter: errorWriter}
}

func (w SplitWriter) Write(level Level, data io.Reader) error {
	writer := w.logWriter
	if level.IsError() {
		writer = w.errorWriter
	}

	_, err := io.Copy(writer, data)
	return err
}

// NewStdoutWriter returns a new SplitWriter that writes to io.Stdout and io.Stderr.
func NewStdoutWriter() *SplitWriter {
	return NewSplitWriter(os.Stderr, os.Stderr)
}
