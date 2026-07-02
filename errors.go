package grout

import (
	"errors"
	"fmt"
	"syscall"
)

var (
	// ErrConnection indicates the UNIX socket connection failed.
	ErrConnection = errors.New("grout: connection failed")

	// ErrVersionMismatch indicates an incompatible API version.
	ErrVersionMismatch = errors.New("grout: API version mismatch")

	// ErrUnsupported indicates the feature is not available in the connected version.
	ErrUnsupported = errors.New("grout: feature not supported in this version")

	// ErrPayloadTooLarge indicates the payload exceeds the 128 KiB limit.
	ErrPayloadTooLarge = errors.New("grout: payload too large")

	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("grout: not found")
)

// GrError wraps an errno returned by the grout daemon.
type GrError struct {
	Errno  syscall.Errno
	Method string
}

func (e *GrError) Error() string {
	return fmt.Sprintf("grout: %s: %s", e.Method, e.Errno)
}

func (e *GrError) Unwrap() error {
	return e.Errno
}

// statusToError converts a grout response status (negated errno) to a Go error.
// A zero status means success (nil error).
func statusToError(status uint32, method string) error {
	if status == 0 {
		return nil
	}
	errno := syscall.Errno(-int32(status))
	if errno == syscall.ENOENT {
		return fmt.Errorf("%w: %s", ErrNotFound, method)
	}
	return &GrError{Errno: errno, Method: method}
}
