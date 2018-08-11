package e2e

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
)

const (
	PanicFailNow = "failnow"
	PanicSkipNow = "skipnow"
)

func runTest(name string, testFn Test) (t *T) {
	t = &T{name: name}
	defer func() {
		if r := recover(); r != nil {
			switch r {
			case PanicFailNow:
				t.Fail()
			case PanicSkipNow:
				t.mu.Lock()
				t.skipped = true
				t.mu.Unlock()
			default:
				panic(r)
			}
			if r == PanicFailNow {
				t.Fail()
			} else {
				panic(r)
			}
		}
	}()
	testFn(t)
	return
}

type Test func(t *T)

type T struct {
	name string

	mu       sync.RWMutex
	failed   bool
	skipped  bool
	output   []byte
	subTests []*T
	runner   string
	helpers  map[string]struct{}
}

func (t *T) Name() string {
	return t.name
}

func (t *T) log(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.output = append(t.output, t.decorate(s)...)
}

func (t *T) Log(args ...interface{}) {
	t.log(fmt.Sprint(args...))
}

func (t *T) Logf(f string, v ...interface{}) {
	t.log(fmt.Sprintf(f, v...))
}

func (t *T) Error(args ...interface{}) {
	t.Fail()
	t.log(fmt.Sprint(args...))
}

func (t *T) Errorf(f string, v ...interface{}) {
	t.Fail()
	t.log(fmt.Sprintf(f, v...))
}

func (t *T) Fatal(args ...interface{}) {
	t.log(fmt.Sprint(args...))
	t.FailNow()
}

func (t *T) Fatalf(f string, v ...interface{}) {
	t.log(fmt.Sprintf(f, v...))
	t.FailNow()
}

func (t *T) Fail() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failed = true
}

func (t *T) FailNow() {
	t.Fail()
	panic(PanicFailNow)
}

func (t *T) Failed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.failed
}

func (t *T) Output() []byte {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for i := range t.subTests {
		if t.subTests[i].Failed() {
			t.output = append(t.output, t.subTests[i].Output()...)
			t.output = append(t.output, '\n')
		}
	}
	if t.Skipped() {
		return append(t.output, "skipped\n"...)
	}
	if t.Failed() {
		return append(t.output, "FAIL\n"...)
	}
	return append(t.output, "PASS\n"...)
}

func (t *T) Run(name string, testFn Test) {
	tt := runTest(name, testFn)
	if tt.Failed() {
		t.Fail()
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	tt.output = append([]byte(fmt.Sprintf("- %s/%s\n", t.name, name)), tt.output...)
	t.subTests = append(t.subTests, tt)
}

func (t *T) Skip(args ...interface{}) {
	t.Log(args...)
	t.SkipNow()
}

func (t *T) Skipf(s string, v ...interface{}) {
	t.Logf(s, v...)
	t.SkipNow()
}

func (t *T) SkipNow() {
	panic(PanicSkipNow)
}

func (t *T) Skipped() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.skipped
}

func (t *T) Helper() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.helpers == nil {
		t.helpers = make(map[string]struct{})
	}
	t.helpers[callerName(1)] = struct{}{}
}

// decorate prefixes the string with the file and line of the call site
// and inserts the final newline if needed and indentation tabs for formatting.

// decorate is lifted verbatim from https://golang.org/src/testing/testing.go#L365
func (t *T) decorate(s string) string {
	skip := t.frameSkip(3) // decorate + log + public function.
	_, file, line, ok := runtime.Caller(skip)
	if ok {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
		line = 1
	}

	buf := new(bytes.Buffer)

	// Every line is indented at least one tab.
	buf.WriteByte('\t')
	fmt.Fprintf(buf, "%s:%d: ", file, line)
	lines := strings.Split(s, "\n")
	if l := len(lines); l > 1 && lines[l-1] == "" {
		lines = lines[:l-1]
	}

	for i, line := range lines {
		if i > 0 {
			// Second and subsequent lines are indented an extra tab.
			buf.WriteString("\n\t\t")
		}
		buf.WriteString(line)
	}
	buf.WriteByte('\n')
	return buf.String()

}

// frameSkip searches, starting after skip frames, for the first caller frame

// in a function not marked as a helper and returns the frames to skip

// to reach that site. The search stops if it finds a tRunner function that

// was the entry point into the test.

// This function must be called with c.mu held.
func (t *T) frameSkip(skip int) int {
	if t.helpers == nil {
		return skip
	}
	var pc [50]uintptr

	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(skip+2, pc[:])
	if n == 0 {
		panic("testing: zero callers found")
	}
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	more := true
	for i := 0; more; i++ {
		frame, more = frames.Next()
		if frame.Function == t.runner {
			// We've gone up all the way to the tRunner calling
			// the test function (so the user must have
			// called tb.Helper from inside that test function).
			// Only skip up to the test function itself.
			return skip + i - 1
		}
		if _, ok := t.helpers[frame.Function]; !ok {
			// Found a frame that wasn't inside a helper function.
			return skip + i
		}
	}
	return skip
}

// callerName gives the function name (qualified with a package path)
// for the caller after skip frames (where 0 means the current function).
func callerName(skip int) string {
	// Make room for the skip PC.
	var pc [2]uintptr
	n := runtime.Callers(skip+2, pc[:]) // skip + runtime.Callers + callerName
	if n == 0 {
		panic("e2e: zero callers found")
	}

	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return frame.Function
}
