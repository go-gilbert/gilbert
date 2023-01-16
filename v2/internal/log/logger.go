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
	EncodeDiagnostics(diags hcl.Diagnostics, buff *buffer.Buffer) error
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
	if l.level < DebugLevel {
		return
	}
	l.write(DebugLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Debugf(msg string, args ...any) {
	if l.level < DebugLevel {
		return
	}
	l.write(DebugLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Info(args ...any) {
	if l.level < InfoLevel {
		return
	}
	l.write(InfoLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Infof(msg string, args ...any) {
	if l.level < InfoLevel {
		return
	}
	l.write(InfoLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Warn(args ...any) {
	if l.level < WarnLevel {
		return
	}
	l.write(WarnLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Warnf(msg string, args ...any) {
	if l.level < WarnLevel {
		return
	}
	l.write(WarnLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Error(args ...any) {
	if l.level < ErrorLevel {
		return
	}
	l.write(ErrorLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Errorf(msg string, args ...any) {
	if l.level < ErrorLevel {
		return
	}
	l.write(ErrorLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Fatal(args ...any) {
	if l.level < FatalLevel {
		return
	}
	l.write(FatalLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Fatalf(msg string, args ...any) {
	if l.level < FatalLevel {
		return
	}
	l.write(FatalLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Panic(args ...any) {
	if l.level < PanicLevel {
		return
	}
	l.write(PanicLevel, fmt.Sprint(args...))
}

func (l OutputPrinter) Panicf(msg string, args ...any) {
	if l.level < PanicLevel {
		return
	}
	l.write(PanicLevel, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Success(args ...any) {
	if l.level < InfoLevel {
		return
	}
	l.writeWithStyle(InfoLevel, StyleSuccess, fmt.Sprint(args...))
}

func (l OutputPrinter) Successf(msg string, args ...any) {
	if l.level < InfoLevel {
		return
	}
	l.writeWithStyle(InfoLevel, StyleSuccess, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) Trace(args ...any) {
	if l.level < InfoLevel {
		return
	}
	l.writeWithStyle(InfoLevel, StyleStep, fmt.Sprint(args...))
}

func (l OutputPrinter) Tracef(msg string, args ...any) {
	if l.level < InfoLevel {
		return
	}
	l.writeWithStyle(InfoLevel, StyleStep, fmt.Sprintf(msg, args...))
}

func (l OutputPrinter) ReportDiagnostics(diags hcl.Diagnostics) {
	buff := buffPool.Get()
	defer buff.Free()

	err := l.encoder.EncodeDiagnostics(diags, buff)
	if err != nil {
		fmt.Println("ERROR: log.Encoder.EncodeEvent:", err)
	}

	err = l.writer.Write(ErrorLevel, bytes.NewReader(buff.Bytes()))
	if err != nil {
		fmt.Println("ERROR: log.Writer.Write:", err)
	}
}
