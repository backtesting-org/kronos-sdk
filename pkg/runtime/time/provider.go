package time

import (
	"time"

	"github.com/backtesting-org/kronos-sdk/pkg/types/temporal"
)

// TimeProvider is a wrapper around time operations
type timeProvider struct{}

// NewTimeProvider creates the production time provider
func NewTimeProvider() temporal.TimeProvider {
	return &timeProvider{}
}

func (tp *timeProvider) Now() time.Time {
	return time.Now()
}

func (tp *timeProvider) Since(t time.Time) time.Duration {
	return time.Since(t)
}

func (tp *timeProvider) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (tp *timeProvider) NewTimer(d time.Duration) temporal.Timer {
	return &timer{timer: time.NewTimer(d)}
}

func (tp *timeProvider) NewTicker(d time.Duration) temporal.Ticker {
	return &ticker{ticker: time.NewTicker(d)}
}

func (tp *timeProvider) Sleep(d time.Duration) {
	time.Sleep(d)
}

// timer wraps time.Timer to implement our Timer interface
type timer struct {
	timer *time.Timer
}

func (t *timer) C() <-chan time.Time {
	return t.timer.C
}

func (t *timer) Reset(d time.Duration) bool {
	return t.timer.Reset(d)
}

func (t *timer) Stop() bool {
	return t.timer.Stop()
}

// ticker wraps time.Ticker to implement our Ticker interface
type ticker struct {
	ticker *time.Ticker
}

func (t *ticker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *ticker) Reset(d time.Duration) {
	t.ticker.Reset(d)
}

func (t *ticker) Stop() {
	t.ticker.Stop()
}
