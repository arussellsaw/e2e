package e2e

import "testing"

func TestFailNow(t *testing.T) {
	var ran bool
	Run(TesterFunc(func(e *E) {
		ran = true
		e.FailNow()
		t.Errorf("expected this to not run")
	}))
	if !ran {
		t.Errorf("expected ran, got !ran")
	}
}

func TestErrorf(t *testing.T) {
	var (
		ran bool
		ee  *E
	)
	Run(TesterFunc(func(e *E) {
		ee = e
		ran = true
		e.Errorf("fail this test")
	}))
	if !ran {
		t.Errorf("expected ran, got !ran")
	}
	if !ee.Failed() {
		t.Errorf("")
	}
}
