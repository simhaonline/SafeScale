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

package retry

import (
	"time"

	"github.com/CS-SI/SafeScale/lib/utils"
	"github.com/CS-SI/SafeScale/lib/utils/retry/enums/verdict"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
	"github.com/CS-SI/SafeScale/lib/utils/temporal"
)

// Arbiter sleeps or selects any amount of time for each attempt.
type Arbiter func(Try) (verdict.Enum, error)

// DefaultArbiter allows 10 retries, with a maximum duration of 30 seconds
var DefaultArbiter = PrevailDone(Max(10), Timeout(temporal.GetBigDelay()))

// PrevailRetry aggregates verdicts from Arbiters for a try :
// - Returns Abort and the error as soon as an arbiter decides for an Abort.
// - If at least one arbiter returns Retry without any Abort from others, returns Retry with nil error.
// - Otherwise returns Done with nil error.
func PrevailRetry(arbiters ...Arbiter) Arbiter {
	return func(t Try) (verdict.Enum, error) {
		final := verdict.Done
		for _, a := range arbiters {
			v, err := a(t)
			if err != nil {
				return 0, err
			}
			switch v {
			case verdict.Retry:
				final = verdict.Retry
			case verdict.Abort:
				return verdict.Abort, err
			}
		}
		return final, nil
	}
}

// PrevailDone aggregates verdicts from Arbiters for a try :
// - Returns Abort and the error as soon as an Abort is decided.
// - If at least one arbiter return Done without any Abort, returns Done with nil error.
// - Otherwise returns Retry with nil error.
func PrevailDone(arbiters ...Arbiter) Arbiter {
	return func(t Try) (verdict.Enum, error) {
		final := verdict.Retry
		for _, a := range arbiters {
			v, err := a(t)
			if err != nil {
				return 0, err
			}
			switch v {
			case verdict.Done:
				final = verdict.Done
			case verdict.Abort:
				return verdict.Abort, err
			}
		}
		return final, nil
	}
}

// Unsuccessful returns Retry when the try produced an error.
func Unsuccessful() Arbiter {
	return func(t Try) (verdict.Enum, error) {
		if t.Err != nil {
			if _, ok := t.Err.(ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			return verdict.Retry, nil
		}
		return verdict.Done, nil
	}
}

// UnsuccessfulWhereRetcode255 returns Retry when the try produced an error with code 255,
// all other codes are considered as successful and returns Done.
func UnsuccessfulWhereRetcode255() Arbiter {
	return func(t Try) (verdict.Enum, error) {
		if t.Err != nil {
			if _, ok := t.Err.(ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			_, retCode, _ := utils.ExtractRetCode(t.Err)
			if retCode == 255 {
				return verdict.Retry, nil
			}
		}
		return verdict.Done, nil
	}
}

// Successful returns Retry when the try produced no error.
func Successful() Arbiter {
	return func(t Try) (verdict.Enum, error) {
		if t.Err == nil {
			if _, ok := t.Err.(ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			return verdict.Retry, nil
		}
		return verdict.Done, nil
	}
}

// Timeout returns Abort after a duration of time passes since the first try.
func Timeout(limit time.Duration) Arbiter {
	return func(t Try) (verdict.Enum, error) {
		if t.Err != nil {
			if _, ok := t.Err.(ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			if time.Since(t.Start) >= limit {
				return verdict.Abort, TimeoutError(limit, t.Err)
			}
			return verdict.Retry, nil
		}
		return verdict.Done, nil
	}
}

// Max errors after a limited number of tries
func Max(limit uint) Arbiter {
	return func(t Try) (verdict.Enum, error) {
		if t.Err != nil {
			if _, ok := t.Err.(ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*ErrAborted); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			if _, ok := t.Err.(*scerr.ErrRuntimePanic); ok {
				return verdict.Done, t.Err
			}

			if t.Count >= limit {
				return verdict.Abort, LimitError(limit, t.Err)
			}
			return verdict.Retry, nil
		}
		return verdict.Done, nil
	}
}
