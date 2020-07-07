package utils

import (
	"errors"
	"fmt"
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func randomBoolean() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(2) == 1
}

func errOriginator() (err error) {
	tracer := concurrency.NewTracer(nil, "(errOriginator)", true).GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	rb := randomBoolean()
	if rb {
		return errors.New(fmt.Sprintf("[%s, %d] This is the first error", scerr.GetCurrentFileName(), scerr.GetCurrentFileLine()))
	}

	return errors.New(fmt.Sprintf("[%s, %d] This is the last error", scerr.GetCurrentFileName(), scerr.GetCurrentFileLine()))
}

func errBuiltin() (err error) {
	tracer := concurrency.NewTracer(nil, "(errBuiltin)", true).GoingIn()
	defer tracer.OnExitTrace()()
	defer scerr.OnExitLogError(tracer.TraceMessage(""), &err)()

	rb := randomBoolean()
	if rb {
		return scerr.Errorf("This is the first error", nil)
	}

	err = errors.New("fake error")
	return scerr.Wrap(err, "This is the last error")
}

func errCaller() error {
	that := errOriginator()
	return that
}

func secondErrorCaller() error {
	return errCaller()
}

func TestLineInfo(t *testing.T) {
	err := secondErrorCaller()
	if err == nil {
		t.FailNow()
	}

	content := err.Error()
	if !(strings.Contains(content, "29") || strings.Contains(content, "32")) {
		t.FailNow()
	}

	if !(strings.Contains(content, "errorinfo_test.go")) {
		t.FailNow()
	}
}

func TestLineInfoBuiltin(t *testing.T) {
	err := errBuiltin()
	if err == nil {
		t.FailNow()
	}

	content := err.Error()
	if !(strings.Contains(content, "42") || strings.Contains(content, "46")) {
		t.FailNow()
	}

	if !(strings.Contains(content, "errorinfo_test.go")) {
		t.FailNow()
	}
}

func TestErrorTraceInformationBuiltin(t *testing.T) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logrus.SetOutput(os.Stdout)

	err := errBuiltin()
	fmt.Print(err)
	_ = err

	_ = w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	outString := string(out)

	fmt.Println(outString)

	// file and line where the TRACER is created
	if !(strings.Contains(outString, "errorinfo_test.go:36]:  ERROR OCCURRED")) {
		t.FailNow()
	}

	if !(strings.Contains(outString, "Builtin")) {
		t.FailNow()
	}
}

func TestErrorTraceInformationOriginator(t *testing.T) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logrus.SetOutput(os.Stdout)

	err := errOriginator()
	fmt.Print(err)
	_ = err

	_ = w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	outString := string(out)

	fmt.Println(outString)

	// file and line where the TRACER is created
	if !(strings.Contains(outString, "errorinfo_test.go:23]:  ERROR OCCURRED")) {
		t.FailNow()
	}

	if !(strings.Contains(outString, "Originator")) {
		t.FailNow()
	}
}
