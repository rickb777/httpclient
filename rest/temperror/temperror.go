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
