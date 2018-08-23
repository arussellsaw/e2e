package e2e

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/avct/schedule"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/mux"
)

type Runner struct {
	s       schedule.Scheduler
	mu      sync.Mutex
	tests   map[string]*testRunner
	history map[string][]testRunner
}

func (r *Runner) Mux() http.Handler {
	isDev := os.Getenv("E2E_DEV") == "1"

	m := mux.NewRouter()
	m.PathPrefix("/ui").Handler(r.GetUIHandler(isDev))

	m.HandleFunc("/api/status", r.StatusHandler)
	m.HandleFunc("/api/force/{name}", r.ForceRunHandler)
	m.HandleFunc("/api/log/{name}", r.LiveOutputHandler)

	return m
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
			r.addPastTest(tr)
			tr.runJob()
		}),
		schedule.Every(interval),
	)
}

func (r *Runner) addPastTest(tr *testRunner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// if this is the first run, don't add history
	if tr.Successes == 0 && tr.Failures == 0 {
		return
	}
	if r.history == nil {
		r.history = make(map[string][]testRunner)
	}
	r.history[tr.Name] = append(r.history[tr.Name], *tr)
}

func (r *Runner) StatusHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.mu.Lock()
	json.NewEncoder(w).Encode(r.tests)
	r.mu.Unlock()
}

func (r *Runner) GetUIHandler(dev bool) http.Handler {
	if dev {
		u := &url.URL{
			Host:   "localhost:3000",
			Scheme: "http",
		}
		rp := httputil.NewSingleHostReverseProxy(u)
		return rp
	}
	b := packr.NewBox("frontend/dist")
	return http.StripPrefix("/ui", http.FileServer(&loggingFileSystem{FileSystem: b}))
}

func (r *Runner) ForceRunHandler(w http.ResponseWriter, req *http.Request) {
	name, ok := mux.Vars(req)["name"]
	if !ok {
		http.Error(w, "400 bad request (missing test name param)", http.StatusBadRequest)
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	tr, ok := r.tests[name]
	if !ok {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	go func() {
		r.addPastTest(tr)
		tr.runJob()
	}()
}

func (r *Runner) LiveOutputHandler(w http.ResponseWriter, req *http.Request) {
	name, ok := mux.Vars(req)["name"]
	if !ok {
		http.Error(w, "400 bad request (missing test name param)", http.StatusBadRequest)
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if tr, ok := r.tests[name]; ok {
		if tr.currentT == nil {
			http.Error(w, "404 not found (no *e2e.T for test)", http.StatusNotFound)
			return
		}
		w.Write(tr.currentT.Output())
		return
	}
	http.Error(w, "404 not found", http.StatusNotFound)
}

type TestState string

const (
	TestStateUnknown TestState = ""
	TestStateRunning TestState = "RUNNING"
	TestStatePassed  TestState = "PASSED"
	TestStateFailed  TestState = "FAILED"
)

type testRunner struct {
	Name string
	t    Test

	mu                sync.Mutex
	currentT          *T
	State             TestState
	LastSuccessTime   time.Time
	LastFailureTime   time.Time
	LastFailureOutput string
	Failures          int
	Successes         int
}

func (tr *testRunner) runJob() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.State = TestStateRunning
	t := &T{}
	tr.currentT = t
	doRun(tr.Name, tr.t, t)
	if t.Failed() {
		tr.State = TestStateFailed
		tr.Failures++
		tr.LastFailureTime = time.Now()
		tr.LastFailureOutput = string(t.Output())
	} else {
		tr.State = TestStatePassed
		tr.Successes++
		tr.LastSuccessTime = time.Now()
	}
}

type loggingFileSystem struct {
	http.FileSystem
}

func (l *loggingFileSystem) Open(name string) (http.File, error) {
	log.Println(name)
	return l.FileSystem.Open(name)
}
