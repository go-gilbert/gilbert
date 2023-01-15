package log

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/hcl/v2"
	"go.uber.org/zap/buffer"
)

var (
	buffPool = buffer.NewPool()
)

type Encoder interface {
	EncodeEvent(event Event, buff *buffer.Buffer) error
	EncodeDiagnostics(src []byte, diags hcl.Diagnostics, buff *buffer.Buffer) error
}

type OutputPrinter struct {
	Name    string
	level   Level
	writer  Writer
	encoder Encoder
}

func NewOutputPrinter(level Level, writer Writer, encoder Encoder) *OutputPrinter {
	return &OutputPrinter{
		level:   level,
		writer:  writer,
		encoder: encoder,
	}
}

func (l OutputPrinter) write(level Level, msg string) {
	l.writeWithStyle(level, StyleDefault, msg)
}

func (l OutputPrinter) writeWithStyle(level Level, style MessageStyle, msg string) {
	event := Event{
		Level:      level,
		Style:      style,
		Time:       time.Now(),
		Message:    msg,
		LoggerName: l.Name,
	}
	buff := buffPool.Get()
	defer buff.Free()

	err := l.encoder.EncodeEvent(event, buff)
	if err != nil {
		fmt.Println("ERROR: log.Encoder.EncodeEvent:", err)
	}

	err = l.writer.Write(level, bytes.NewReader(buff.Bytes()))
	if err != nil {
		fmt.Println("ERROR: log.Writer.Write:", err)
	}

	switch level {
	case PanicLevel:
		panic(event.Message)
	case FatalLevel:
		os.Exit(1)
	}
}

func (l OutputPrinter) Debug(args ...any) {
	l.write(DebugLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Debugf(msg string, args ...any) {
	l.write(DebugLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Info(args ...any) {
	l.write(InfoLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Infof(msg string, args ...any) {
	l.write(InfoLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Warn(args ...any) {
	l.write(WarnLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Warnf(msg string, args ...any) {
	l.write(WarnLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Error(args ...any) {
	l.write(ErrorLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Errorf(msg string, args ...any) {
	l.write(ErrorLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Fatal(args ...any) {
	l.write(FatalLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Fatalf(msg string, args ...any) {
	l.write(FatalLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Panic(args ...any) {
	l.write(PanicLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Panicf(msg string, args ...any) {
	l.write(PanicLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Success(args ...any) {
	l.writeWithStyle(InfoLevel, StyleSuccess, fmt.Sprint(args...))
}

func (l OutputPrinter) Successf(msg string, args ...any) {
	l.writeWithStyle(InfoLevel, StyleSuccess, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Trace(args ...any) {
	l.writeWithStyle(InfoLevel, StyleStep, fmt.Sprint(args...))
}

func (l OutputPrinter) Tracef(msg string, args ...any) {
	l.writeWithStyle(InfoLevel, StyleStep, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) ReportDiagnostics(src []byte, diags hcl.Diagnostics) {
	buff := buffPool.Get()
	defer buff.Free()

	err := l.encoder.EncodeDiagnostics(src, diags, buff)
	if err != nil {
		fmt.Println("ERROR: log.Encoder.EncodeEvent:", err)
	}

	err = l.writer.Write(ErrorLevel, bytes.NewReader(buff.Bytes()))
	if err != nil {
		fmt.Println("ERROR: log.Writer.Write:", err)
	}
}
