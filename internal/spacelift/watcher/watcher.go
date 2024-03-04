package watcher

import "time"

const (
	DefaultTimeout     = 70 * time.Minute
	DefaultInterval    = 3 * time.Second
	DefaultErrInterval = 10 * time.Second
)

type Watcher struct {
	Interval, ErrInterval time.Duration
	Timeout               time.Duration
}

var DefaultWatcher = Watcher{
	ErrInterval: DefaultErrInterval,
	Interval:    DefaultInterval,
	Timeout:     DefaultTimeout,
}
