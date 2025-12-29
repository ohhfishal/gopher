package runtime

import (
	"github.com/ohhfishal/nibbles/assert"
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
func AfterEvery(duration time.Duration) Event {
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

/*
Returns an Event that yields whenever a file of the matching extension is modified.
Interval is the minimum time between two events.
Syntax sugar for [FileCache.Event] that panics if there is an error. (Which signifies the os is probably suffering).
*/
func OnFileChange(interval time.Duration, extensions ...string) Event {
	// TODO: Support options
	cache := &FileCache{
		Interval:   interval,
		Extensions: extensions,
	}
	event, err := cache.Event()
	assert.Nil(err, "filecache getting an event: %w", err)
	assert.True(
		event != nil,
		"caught null pointer exception due to incorrect FileCache.Event()",
	)
	return event
}
