package log

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap/buffer"
)

var (
	buffPool = buffer.NewPool()
)

type Encoder interface {
	EncodeEvent(event Event, buff *buffer.Buffer) error
}

type AbstractLogger interface {
	Debug(args ...any)
	Debugf(msg string, args ...any)
	Info(args ...any)
	Infof(msg string, args ...any)
	Warn(args ...any)
	Warnf(msg string, args ...any)
	Error(args ...any)
	Errorf(msg string, args ...any)
	Fatal(args ...any)
	Fatalf(msg string, args ...any)
	Panic(args ...any)
	Panicf(msg string, args ...any)
	Success(args ...any)
	Successf(msg string, args ...any)
	Trace(args ...any)
	Tracef(msg string, args ...any)
}

type Logger struct {
	Name    string
	level   Level
	writer  Writer
	encoder Encoder
}

func NewLogger(level Level, writer Writer, encoder Encoder) *Logger {
	return &Logger{
		level:   level,
		writer:  writer,
		encoder: encoder,
	}
}

func (l Logger) write(level Level, msg string) {
	l.writeWithStyle(level, StyleDefault, msg)
}

func (l Logger) writeWithStyle(level Level, style MessageStyle, msg string) {
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

func (l Logger) Debug(args ...any) {
	l.write(DebugLevel, fmt.Sprint(args...))
}

func (l Logger) Debugf(msg string, args ...any) {
	l.write(DebugLevel, fmt.Sprintf(msg, args...))
}

func (l Logger) Info(args ...any) {
	l.write(InfoLevel, fmt.Sprint(args...))
}

func (l Logger) Infof(msg string, args ...any) {
	l.write(InfoLevel, fmt.Sprintf(msg, args...))
}

func (l Logger) Warn(args ...any) {
	l.write(WarnLevel, fmt.Sprint(args...))
}

func (l Logger) Warnf(msg string, args ...any) {
	l.write(WarnLevel, fmt.Sprintf(msg, args...))
}

func (l Logger) Error(args ...any) {
	l.write(ErrorLevel, fmt.Sprint(args...))
}

func (l Logger) Errorf(msg string, args ...any) {
	l.write(ErrorLevel, fmt.Sprintf(msg, args...))
}

func (l Logger) Fatal(args ...any) {
	l.write(FatalLevel, fmt.Sprint(args...))
}

func (l Logger) Fatalf(msg string, args ...any) {
	l.write(FatalLevel, fmt.Sprintf(msg, args...))
}

func (l Logger) Panic(args ...any) {
	l.write(PanicLevel, fmt.Sprint(args...))
}

func (l Logger) Panicf(msg string, args ...any) {
	l.write(PanicLevel, fmt.Sprintf(msg, args...))
}

func (l Logger) Success(args ...any) {
	l.writeWithStyle(InfoLevel, StyleSuccess, fmt.Sprint(args...))
}

func (l Logger) Successf(msg string, args ...any) {
	l.writeWithStyle(InfoLevel, StyleSuccess, fmt.Sprintf(msg, args...))
}

func (l Logger) Trace(args ...any) {
	l.writeWithStyle(InfoLevel, StyleStep, fmt.Sprint(args...))
}

func (l Logger) Tracef(msg string, args ...any) {
	l.writeWithStyle(InfoLevel, StyleStep, fmt.Sprintf(msg, args...))
}
