package rx

import "time"

// DebounceTime receives function and will postpone its execution until
// after wait milliseconds have elapsed since the last time it was invoked.
//
// Useful for implementing behavior that should only happen after the input has stopped arriving.
func DebounceTime(interval time.Duration, input chan interface{}, cb func(arg interface{})) {
	var item interface{}
	timer := time.NewTimer(interval)
	for {
		select {
		case item = <-input:
			timer.Reset(interval)
		case <-timer.C:
			if item != nil {
				cb(item)
			}
		}
	}
}
