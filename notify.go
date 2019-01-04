package e2e

import "time"

type Notification struct {
	Name     string
	Failed   bool
	Output   []byte
	Duration time.Duration
}

type Notifier interface {
	Notify(n Notification)
}

type noopNotifier struct{}

func (nn noopNotifier) Notify(n Notification) {
}

var defaultNotifier = noopNotifier{}
