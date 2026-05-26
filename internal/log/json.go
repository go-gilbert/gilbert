package log

import (
	"encoding/json"
	"io"
	"sync"
	"time"
	"unsafe"
)

const newLine = "\n"

var _ Writer = (*JSONWriter)(nil)

type jsonLine struct {
	At      time.Time `json:"at"`
	Level   Level     `json:"level"`
	Tag     string    `json:"tag,omitempty"`
	Message string    `json:"msg,omitempty"`
	Fields  []Field   `json:"fields,omitempty"`
}

type jsonIOWriter struct {
	tag   string
	level Level
	lock  *sync.Mutex
	dst   io.Writer
}

func (w jsonIOWriter) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()

	line := jsonLine{
		At:      time.Now(),
		Level:   w.level,
		Tag:     w.tag,
		Message: bytesAsString(p),
	}

	err := json.NewEncoder(w.dst).Encode(line)
	return len(p), err
}

func bytesAsString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

type JSONWriter struct {
	lock   sync.Mutex
	stdout io.Writer
	stderr io.Writer
}

// NewJSONWriter returns a writer that writes log messages in JSON format.
func NewJSONWriter(streams IOStreams) *JSONWriter {
	return &JSONWriter{
		stdout: streams.Stdout,
		stderr: streams.Stderr,
	}
}

func (w *JSONWriter) IOStreamWriters(tag string) IOStreamWriters {
	return IOStreamWriters{
		Stdout: jsonIOWriter{tag: tag, level: LevelInfo, lock: &w.lock, dst: w.stdout},
		Stderr: jsonIOWriter{tag: tag, level: LevelError, lock: &w.lock, dst: w.stderr},
	}
}

func (w *JSONWriter) Write(level Level, tag, message string, fields []Field) {
	w.lock.Lock()
	defer w.lock.Unlock()

	dst := writerForLevel(level, w.stdout, w.stderr)
	line := jsonLine{
		At:      time.Now(),
		Level:   level,
		Tag:     tag,
		Message: message,
		Fields:  fields,
	}

	_ = json.NewEncoder(dst).Encode(line)
}
