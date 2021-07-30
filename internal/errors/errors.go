package errors

import (
	"errors"
	"fmt"
)

const (
	StatusReasonAlreadyExist StatusReason = "already exist"
	StatusReasonNotExist     StatusReason = "not exist"
	StatusReasonNotRequired  StatusReason = "not required"
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
