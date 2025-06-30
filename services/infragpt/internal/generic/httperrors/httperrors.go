package httperrors

import (
	"errors"
	"strings"
)

type Error struct {
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Fields     []string `json:"fields"`
	HttpStatus int      `json:"-"`
}

func (e Error) Error() string {
	return e.Message
}

// From converts an error to an Error using errors.As
func From(err error) Error {
	var e Error
	if ok := errors.As(err, &e); !ok {
		httpStatus := 400
		msg := ""
		if err != nil {
			msg = err.Error()
		}
		if err == nil {
			httpStatus = 200
		}
		if err != nil && (strings.Contains(strings.ToLower(msg), "internal") ||
			strings.Contains(err.Error(), "panic") ||
			strings.Contains(err.Error(), "nil pointer dereference") ||
			strings.Contains(err.Error(), "goroutine")) {
			httpStatus = 500
		}
		return Error{
			HttpStatus: httpStatus,
			Code:       "",
			Fields:     []string{},
			Message:    msg,
		}
	}
	return e
}

func New(httpStatus int, code string, message string, fields []string) error {
	return Error{
		Code:       code,
		Message:    message,
		Fields:     fields,
		HttpStatus: httpStatus,
	}
}

func (e Error) Is(target error) bool {
	var err Error
	if ok := errors.As(target, &err); !ok {
		return false
	}

	if e.Code == err.Code {
		return true
	}

	if e.Message == err.Message {
		return true
	}

	return false
}
