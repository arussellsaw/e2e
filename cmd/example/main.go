package main

import (
	"net/http"
	"time"

	"github.com/arussellsaw/e2e"
)

func main() {
	r := e2e.Runner{}

	r.Schedule("TestAlwaysPasses", TestAlwaysPasses, 10*time.Second)
	r.Schedule("TestAlwaysFails", TestAlwaysFails, 10*time.Second)
	r.Schedule("TestSubtests", TestSubtests, 10*time.Second)
	r.Schedule("TestSlow", TestSlow, 1*time.Minute)

	http.ListenAndServe(":8080", r.Mux())
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

func TestSlow(t *e2e.T) {
	t.Logf("this test is slow")
	time.Sleep(1 * time.Second)
	t.Logf("it logs periodically")
	time.Sleep(30 * time.Second)
	t.Logf("hopefully it's done soon")
	time.Sleep(30 * time.Second)
	t.Logf("done")
	t.Logf("")
}
