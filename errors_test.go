package grout

import (
	"errors"
	"syscall"
	"testing"
)

func TestGrError_Error(t *testing.T) {
	e := &GrError{Errno: syscall.ENOENT, Method: "InterfaceGet"}
	got := e.Error()
	want := "grout: InterfaceGet: no such file or directory"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestGrError_Unwrap(t *testing.T) {
	e := &GrError{Errno: syscall.ENOENT, Method: "InterfaceGet"}
	if !errors.Is(e, syscall.ENOENT) {
		t.Error("expected GrError to unwrap to syscall.ENOENT")
	}
}

func TestGrError_As(t *testing.T) {
	var err error = &GrError{Errno: syscall.EPERM, Method: "InterfaceAdd"}
	var grErr *GrError
	if !errors.As(err, &grErr) {
		t.Error("expected errors.As to succeed for *GrError")
	}
	if grErr.Errno != syscall.EPERM {
		t.Errorf("Errno = %v, want EPERM", grErr.Errno)
	}
}

func TestStatusToError_Success(t *testing.T) {
	if err := statusToError(0, "test"); err != nil {
		t.Errorf("expected nil, got: %v", err)
	}
}

func TestStatusToError_NotFound(t *testing.T) {
	// -ENOENT = -2 → 0xfffffffe as uint32
	err := statusToError(0xfffffffe, "InterfaceGet")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStatusToError_OtherErrno(t *testing.T) {
	// -EPERM = -1 → 0xffffffff as uint32
	err := statusToError(0xffffffff, "InterfaceAdd")
	var grErr *GrError
	if !errors.As(err, &grErr) {
		t.Fatalf("expected *GrError, got: %v", err)
	}
	if grErr.Errno != syscall.EPERM {
		t.Errorf("Errno = %v, want EPERM", grErr.Errno)
	}
	if grErr.Method != "InterfaceAdd" {
		t.Errorf("Method = %q, want %q", grErr.Method, "InterfaceAdd")
	}
}

func TestSentinelErrors(t *testing.T) {
	sentinels := []error{
		ErrConnection,
		ErrVersionMismatch,
		ErrUnsupported,
		ErrPayloadTooLarge,
		ErrNotFound,
	}
	for _, s := range sentinels {
		if s == nil {
			t.Error("sentinel error is nil")
		}
		if s.Error() == "" {
			t.Error("sentinel error has empty message")
		}
	}
}
