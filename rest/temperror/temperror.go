package temperror

import (
	"errors"
	"net"
	"os"
)

// NetworkConnectionError tests an error to see whether it
// was caused by a network connection failing. Depending on context,
// the network operation could be retried.
func NetworkConnectionError(err error) (matched bool) {
	var sce = &os.SyscallError{}
	if errors.As(err, &sce) && sce.Syscall == "connect" {
		return true
	}

	var opError = &net.OpError{}
	if errors.As(err, &opError) && opError.Op == "dial" {
		var sce = &os.SyscallError{}
		if errors.As(opError.Err, &sce) {
			return true
		}
	}

	return false
}

//-------------------------------------------------------------------------------------------------

// TransientError signals that the operation can be retried.
// This is the opposite to PermanentError in github.com/cenkalti/backoff/v4
type TransientError struct {
	Err error
}

// Wrap wraps the given error in a *TransientError.
func Wrap(err error) error {
	if err == nil {
		return nil
	}
	if already := IsTransient(err); already {
		return err
	}
	return &TransientError{
		Err: err,
	}
}

// Error implements the error interface.
func (e *TransientError) Error() string {
	return "transient error: " + e.Err.Error()
}

// Unwrap provides access to the contained error.
func (e *TransientError) Unwrap() error {
	return e.Err
}

// Is performs a shallow comparison and is used by [errors.Is].
func (e *TransientError) Is(target error) bool {
	_, ok := target.(*TransientError)
	return ok
}

//-------------------------------------------------------------------------------------------------

// IsTransient tests whether an error is a *TransientError.
func IsTransient(target error) bool {
	return errors.Is(target, &TransientError{})
}

// IsPermanent tests whether an error is not a *TransientError.
func IsPermanent(target error) bool {
	return !IsTransient(target)
}
