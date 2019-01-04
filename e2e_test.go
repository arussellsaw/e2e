package e2e

import "testing"

func TestFailNow(t *testing.T) {
	var ran bool
	Run("FailNow", func(t *T) {
		ran = true
		t.FailNow()
		t.Errorf("expected this to not run")
	})
	if !ran {
		t.Errorf("expected ran, got !ran")
	}
}

func TestErrorf(t *testing.T) {
	var (
		ran bool
		tt  *T
	)
	Run("Errorf", func(t *T) {
		tt = t
		ran = true
		t.Errorf("fail this test")
	})
	if !ran {
		t.Errorf("expected ran, got !ran")
	}
	if !tt.Failed() {
		t.Errorf("Test should have failed")
	}
	out := string(tt.output)
	expected := "\te2e_test.go:25: fail this test\n"
	if out != expected {
		t.Errorf("Expected output %q, got %q", expected, out)
	}
}

type testNotifier struct {
	n *Notification
}

func (tn *testNotifier) Notify(n Notification) {
	tn.n = &n
}

func TestNotificationOnFailure(t *testing.T) {
	n := &testNotifier{}
	runner := &testRunner{
		Name: "test",
		t: func(t *T) {
			t.Errorf("output")
			t.FailNow()
		},
		n: n,
	}
	runner.runJob()
	if n.n == nil {
		t.Fatal("should have received a notification")
	}
	if !n.n.Failed {
		t.Error("should have failed")
	}
	if n.n.Name != "test" {
		t.Errorf("Expected %q, got %q", "test", n.n.Name)
	}
	if n.n.Duration == 0 {
		t.Error("Duration should be postitive")
	}
	expectedOutput := "\te2e_test.go:53: output\n"
	if string(n.n.Output) != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, string(n.n.Output))
	}
}

func TestNotificationOnSuccess(t *testing.T) {
	n := &testNotifier{}
	runner := &testRunner{
		Name: "test",
		t: func(t *T) {
			// No failure
		},
		n: n,
	}
	runner.runJob()
	if n.n == nil {
		t.Fatal("should have received a notification")
	}
	if n.n.Failed {
		t.Error("should not have failed")
	}
	if n.n.Name != "test" {
		t.Errorf("Expected %q, got %q", "test", n.n.Name)
	}
	if n.n.Duration == 0 {
		t.Error("Duration should be postitive")
	}
	expectedOutput := ""
	if string(n.n.Output) != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, string(n.n.Output))
	}
}
