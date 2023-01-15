package log

// MessageStyle is event message style. Reflects on how it's displayed.
type MessageStyle int

const (
	// StyleDefault is plain event style.
	StyleDefault MessageStyle = iota

	// StyleSuccess marks event as result success event.
	StyleSuccess

	// StyleStep marks event as process step event.
	StyleStep
)
