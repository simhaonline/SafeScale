/*
 * Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package concurrency

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/CS-SI/SafeScale/lib/utils/temporal"
)

// Tracer ...
type Tracer struct {
	taskSig      string
	fileName     string
	funcName     string
	callerParams string
	enabled      bool
	inDone       bool
	outDone      bool
	sw           temporal.Stopwatch
}

var NullTracer = &Tracer{}

// IsLogActive ... FIXME:
func IsLogActive(key string) bool {
	if logs := os.Getenv("SAFESCALE_OPTIONAL_LOGS"); logs != "" {
		return strings.Contains(logs, key)
	}
	return false
}

// NewTracer creates a new Tracer instance
func NewTracer(t Task, message string, enabled bool) *Tracer {
	tracer := Tracer{}
	if t != nil {
		tracer.taskSig = t.GetSignature()
	}
	tracer.enabled = enabled

	tracer.callerParams = message
	if message == "" {
		tracer.callerParams = "()"
	}

	if pc, file, _, ok := runtime.Caller(1); ok {
		tracer.fileName = strings.Replace(file, getPartToRemove(), "", 1)
		if f := runtime.FuncForPC(pc); f != nil {
			tracer.funcName = filepath.Base(f.Name())
		}
	}
	if tracer.funcName == "" {
		tracer.funcName = "<unknown function>"
	}
	if tracer.fileName == "" {
		tracer.funcName = "<unknown file>"
	}

	return &tracer
}

// IsNull returns true if the instance is NullTracer
func (t *Tracer) IsNull() bool {
	return t == nil || t == NullTracer
}

// WithStopwatch will add a measure of duration between GoingIn and GoingOut.
// GoingOut will add the elapsed time in the log message (if it has to be logged...).
func (t *Tracer) WithStopwatch() *Tracer {
	if !t.IsNull() && t.sw == nil {
		t.sw = temporal.NewStopwatch()
	}
	return t
}

// GoingInMessage returns the content of the message when entering the function
func (t *Tracer) GoingInMessage() string {
	if t.IsNull() {
		return ""
	}
	return ">>> " + t.buildMessage()
}

// GoingIn logs the input message (signifying we are going in) using TRACE level
func (t *Tracer) GoingIn() *Tracer {
	if !t.IsNull() && !t.inDone {
		if t.sw != nil {
			t.sw.Start()
		}
		if t.enabled {
			t.inDone = true
			msg := t.GoingInMessage()
			if msg != "" {
				logrus.Tracef(msg)
			}
		}
	}
	return t
}

// GoingOutMessage returns the content of the message when exiting the function
func (t *Tracer) GoingOutMessage() string {
	if t.IsNull() {
		return ""
	}
	return "<<< " + t.buildMessage()
}

// GoingOut logs the output message (signifying we are going out) using TRACE level and adds duration if WithStopwatch() has been called.
func (t *Tracer) GoingOut() *Tracer {
	if !t.IsNull() && !t.outDone {
		if t.sw != nil {
			t.sw.Stop()
		}
		if t.enabled {
			t.outDone = true
			msg := t.GoingOutMessage()
			if t.sw != nil {
				msg += " (duration: " + t.sw.String() + ")"
			}
			if msg != "" {
				logrus.Tracef(msg)
			}
		}
	}
	return t
}

// OnExitTrace returns a function that will log the output message using TRACE level.
func (t *Tracer) OnExitTrace() func() {
	if t.IsNull() || t.outDone {
		return func() {}
	}
	return func() { t.GoingOut() }
}

// buildMessage builds the message with available information from stack trace
func (t *Tracer) buildMessage() string {
	if t.IsNull() {
		return ""
	}

	message := t.taskSig
	if _, _, line, ok := runtime.Caller(1); ok {
		message += " " + t.funcName + t.callerParams + " [" + t.fileName + ":" + strconv.Itoa(line) + "]"
	}
	return message
}

// TraceMessage returns a string containing a trace message
func (t *Tracer) TraceMessage(format string, a ...interface{}) string {
	return "--- " + t.buildMessage() + ": " + fmt.Sprintf(format, a...)
}

// Trace traces a message
func (t *Tracer) Trace(format string, a ...interface{}) *Tracer {
	if !t.IsNull() && t.enabled {
		msg := t.TraceMessage(format, a...)
		if msg != "" {
			logrus.Tracef(msg)
		}
	}
	return t
}

// TraceAsError traces a message with error level
func (t *Tracer) TraceAsError(format string, a ...interface{}) *Tracer {
	if !t.IsNull() && t.enabled {
		msg := t.TraceMessage(format, a...)
		if msg != "" {
			logrus.Errorf(msg)
		}
	}
	return t
}

// Stopwatch returns the stopwatch used (if a stopwatch has been asked with WithStopwatch() )
func (t *Tracer) Stopwatch() temporal.Stopwatch {
	return t.sw
}

// removePart contains the basedir to remove from file pathes
var removePart atomic.Value

func getPartToRemove() string {
	if anon := removePart.Load(); anon != nil {
		return anon.(string)
	}
	return "github.com/CS-SI/SafeScale/"
}

func init() {
	var rootPath string
	if _, file, _, ok := runtime.Caller(0); ok {
		rootPath = strings.Split(file, "lib/utils")[0]
	}
	removePart.Store(rootPath)
}
