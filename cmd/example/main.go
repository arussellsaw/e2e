package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/arussellsaw/e2e"
	"github.com/gorilla/mux"
)

func main() {
	r := e2e.Runner{}

	r.Schedule("TestAlwaysPasses", TestAlwaysPasses, 10*time.Second)
	r.Schedule("TestAlwaysFails", TestAlwaysFails, 10*time.Second)
	r.Schedule("TestSubtests", TestSubtests, 10*time.Second)

	m := mux.NewRouter()
	h, err := r.GetUIHandler()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	m.Handle("/ui", h)
	m.Handle("/status", http.HandlerFunc(r.StatusHandler))

	http.ListenAndServe(":8080", m)
}

func TestAlwaysPasses(t *e2e.T) {
	t.Logf("this is fine!")
}

func TestAlwaysFails(t *e2e.T) {
	t.Logf("this always fails")
	t.Errorf("expected nil, got some error")
	t.Errorf("expected 0, got 1")
}

func TestSubtests(t *e2e.T) {
	t.Logf("this test uses subtests")
	tc := []struct {
		name string
		log  string
		fail bool
	}{
		{
			name: "foo",
			log:  "run foo",
			fail: false,
		},
		{
			name: "bar",
			log:  "run bar",
			fail: true,
		},
	}
	for i := range tc {
		t.Run(tc[i].name, func(t *e2e.T) {
			t.Logf("this test logs: %s", tc[i].log)
			if tc[i].fail {
				t.Errorf("fail")
			}
		})
	}
}
