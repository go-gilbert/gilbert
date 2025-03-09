package log

import (
	"encoding/json"
	"sync"
	"time"
)

const newLine = "\n"

type jsonLine struct {
	At      time.Time `json:"at"`
	Level   Level     `json:"level"`
	Tag     string    `json:"tag,omitempty"`
	Message string    `json:"msg,omitempty"`
}

type JSONWriter struct {
	lock sync.Mutex
}

// NewJSONWriter returns a writer that writes log messages in JSON format.
func NewJSONWriter() *JSONWriter {
	return &JSONWriter{}
}

func (w *JSONWriter) Write(level Level, tag, message string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	dst := writerForLevel(level)
	line := jsonLine{
		At:      time.Now(),
		Level:   level,
		Tag:     tag,
		Message: message,
	}

	_ = json.NewEncoder(dst).Encode(line)
	_, _ = dst.Write([]byte(newLine))
}
