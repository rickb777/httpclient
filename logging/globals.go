package logging

import (
	"time"
)

// Now provides the current time. It can be stubbed for testing.
var Now = func() time.Time {
	return time.Now().UTC()
}
