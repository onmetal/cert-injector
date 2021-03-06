/*
Copyright (c) 2021 T-Systems International GmbH, SAP SE or an SAP affiliate company. All right reserved
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

package errors

import (
	"errors"
	"fmt"
	"strings"
)

const (
	StatusReasonAlreadyExist StatusReason = "already exist"
	StatusReasonNotExist     StatusReason = "not exist"
	StatusReasonNotRequired  StatusReason = "not required"
	StatusReasonNotFound     StatusReason = "not found"
	StatusReasonUnknown      StatusReason = "unknown"
)

type Reason struct {
	Message      string
	StatusReason StatusReason
}

type StatusReason string

type Error struct {
	ErrStatus Reason
}

type IStatus interface {
	Status() Reason
}

func (e *Error) Error() string { return e.ErrStatus.Message }

func (e *Error) Status() Reason { return e.ErrStatus }

func IsNotExist(err error) bool { return ReasonForError(err) == StatusReasonNotExist }

func IsAlreadyExists(err error) bool { return ReasonForError(err) == StatusReasonAlreadyExist }

func IsNotRequired(err error) bool { return ReasonForError(err) == StatusReasonNotRequired }

func IsNotFound(err error) bool { return ReasonForError(err) == StatusReasonNotFound }

func IsRateLimited(err error) bool { return strings.Contains(err.Error(), "rateLimited") }

func ReasonForError(err error) StatusReason {
	if reason := IStatus(nil); errors.As(err, &reason) {
		return reason.Status().StatusReason
	}
	return StatusReasonUnknown
}

func AlreadyExist(s string) *Error {
	return &Error{
		ErrStatus: Reason{
			Message:      fmt.Sprintf("name: %s, already exist ", s),
			StatusReason: StatusReasonAlreadyExist,
		},
	}
}

func NotExist(s string) *Error {
	return &Error{
		ErrStatus: Reason{
			Message:      fmt.Sprintf("name: %s, not exist", s),
			StatusReason: StatusReasonNotExist,
		},
	}
}

func NotRequired() *Error {
	return &Error{
		ErrStatus: Reason{
			Message:      "component is under deletion",
			StatusReason: StatusReasonNotRequired,
		},
	}
}

func NotFound() *Error {
	return &Error{
		ErrStatus: Reason{
			Message:      "not found",
			StatusReason: StatusReasonNotRequired,
		},
	}
}
