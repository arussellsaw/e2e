package e2e

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/arussellsaw/lit"
	"github.com/avct/schedule"
	"github.com/gobuffalo/packr"
)

type Runner struct {
	s     schedule.Scheduler
	mu    sync.Mutex
	tests map[string]*testRunner
}

func (r *Runner) Schedule(name string, t Test, interval time.Duration) {
	r.mu.Lock()
	tr := &testRunner{
		Name: name,
		t:    t,
	}
	if r.tests == nil {
		r.tests = make(map[string]*testRunner)
	}
	r.tests[name] = tr
	r.mu.Unlock()
	r.s.Schedule(
		schedule.JobFunc(func() {
			tr.runJob()
		}),
		schedule.Every(interval),
	)
}

func (r *Runner) StatusHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.mu.Lock()
	json.NewEncoder(w).Encode(r.tests)
	r.mu.Unlock()
}

func (r *Runner) GetUIHandler() (http.Handler, error) {
	b := packr.NewBox("static")
	h, err := lit.LittleUI(lit.DefaultWrapper, b.String("index.html"), func(req *http.Request) (interface{}, error) {
		r.mu.Lock()
		defer r.mu.Unlock()
		return r.tests, nil
	})
	return h, err
}

type testRunner struct {
	Name string
	t    Test

	mu                sync.Mutex
	Failing           bool
	LastSuccessTime   time.Time
	LastFailureTime   time.Time
	LastFailureOutput string
	Failures          int
	Successes         int
}

func (tr *testRunner) runJob() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	e := runTest(tr.Name, tr.t)
	if e.Failed() {
		tr.Failing = true
		tr.Failures++
		tr.LastFailureTime = time.Now()
		tr.LastFailureOutput = string(e.Output())
	} else {
		tr.Failing = false
		tr.Successes++
		tr.LastSuccessTime = time.Now()
	}
}
