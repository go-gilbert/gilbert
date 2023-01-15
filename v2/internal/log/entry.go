package log

import "time"

type Event struct {
	Style      MessageStyle `json:"style,omitempty"`
	Level      Level        `json:"level"`
	Time       time.Time    `json:"time"`
	Message    string       `json:"message"`
	LoggerName string       `json:"logger,omitempty"`
}
