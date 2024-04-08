/*
Copyright 2024 Georgia Institute of Technology

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Based on zgrab2/status.go
package aliasv6

import (
	"io"
	"runtime/debug"

	log "github.com/sirupsen/logrus"
)

// LookUpStatus is the enum value that states how the lookup ended.
type LookUpStatus string

// TODO: Conform to standard string const format (names, capitalization, hyphens/underscores, etc)
// TODO: Enumerate further status types
const (
	LOOKUP_SUCCESS       = LookUpStatus("success")       // The protocol in question was positively identified and the lookup encountered no errors
	LOOKUP_NO_MATCH      = LookUpStatus("no-match")      // No positive match on the aliased lookup table
	LOOKUP_UNKNOWN_ERROR = LookUpStatus("unknown-error") // Catch-all for unrecognized errors
)

// LookUpError an error that also includes a LookUpStatus.
type LookUpError struct {
	Status LookUpStatus
	Err    error
}

// Error is an implementation of the builtin.error interface -- just forward the wrapped error's Error() method
func (err *LookUpError) Error() string {
	if err.Err == nil {
		return "<nil>"
	}
	return err.Err.Error()
}

func (err *LookUpError) Unpack(results interface{}) (LookUpStatus, interface{}, error) {
	return err.Status, results, err.Err
}

// NewLookUpError returns a LookUpError with the given status and error.
func NewLookUpError(status LookUpStatus, err error) *LookUpError {
	return &LookUpError{Status: status, Err: err}
}

// DetectLookUpError returns a LookUpError that attempts to detect the status from the given error.
func DetectLookUpError(err error) *LookUpError {
	return &LookUpError{Status: TryGetLookUpStatus(err), Err: err}
}

// TryGetLookUpStatus attempts to get the LookUpStatus enum value corresponding to the given error.
// A nil error is interpreted as LOOKUP_SUCCESS.
// An unrecognized error is interpreted as LOOKUP_UNKNOWN_ERROR.
func TryGetLookUpStatus(err error) LookUpStatus {
	if err == nil {
		return LOOKUP_SUCCESS
	}
	if err == io.EOF {
		// Presumably the caller did not call TryGetLookUpStatus if the EOF was expected
		return LOOKUP_UNKNOWN_ERROR
	}
	switch e := err.(type) {
	case *LookUpError:
		return e.Status
	default:
		log.Debugf("Failed to detect error from %v at %s", e, string(debug.Stack()))
		return LOOKUP_UNKNOWN_ERROR
	}
}
