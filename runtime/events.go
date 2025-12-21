package runtime

import (
	"iter"
	"time"
)

// A sequence that yields when work is to be done.
type Event iter.Seq[any]

// Returns a sequence that yields once immediately, then returns the passed in event's sequence.
func NowAnd(when Event) Event {
	return func(yield func(any) bool) {
		for range Now() {
			if !yield(nil) {
				break
			}
		}
		for range when {
			if !yield(nil) {
				return
			}
		}
	}
}

// Returns a single-event sequence that yields once, immediately.
func Now() Event {
	return func(yield func(_ any) bool) {
		_ = yield(nil)
	}
}

// Returns an unbounded sequence that yields after a [time.Ticker] tick.
func Every(duration time.Duration) Event {
	ticker := time.NewTicker(duration)
	return func(yield func(_ any) bool) {
		defer ticker.Stop()
		for range ticker.C {
			if !yield(nil) {
				return
			}
		}
	}
}
