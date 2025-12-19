package runner

import (
	"iter"
	"time"
)

type RunEvent iter.Seq[any]

func NowAnd(when RunEvent) RunEvent {
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

func Now() RunEvent {
	return func(yield func(_ any) bool) {
		_ = yield(nil)
	}
}

func Every(duration time.Duration) RunEvent {
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
