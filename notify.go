package e2e

type Notifier interface {
	Notify(*T) error
}
