package retry

import (
	"errors"
	"github.com/cenkalti/backoff/v4"
	"github.com/rs/zerolog"
	"net"
	"time"
)

// RetryConfig configures connection retry logic. The zero value will work fine.
// By default, the initial back-off is 500ms and this increases exponentially
// to a maximum of 1 minute.
type RetryConfig struct {
	ConnectTimeout time.Duration // give up after this
	ConnectTries   int
}

// NewExponentialBackOff runs some connect function repeatedly until it returns without
// error or returns with an error considered permanent (i.e. not a network error).
//
// It is based on github.com/cenkalti/backoff/v4 backoff.ExponentialBackOff
func NewExponentialBackOff(cfg RetryConfig, target string, lgr zerolog.Logger, connect func() error) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = cfg.ConnectTimeout

	var boff backoff.BackOff = exponentialBackOff
	if cfg.ConnectTries > 0 {
		boff = backoff.WithMaxRetries(exponentialBackOff, uint64(cfg.ConnectTries))
	}

	err := backoff.RetryNotify(
		func() error {
			e2 := connect()
			if e2 != nil {
				if opError, is := NetworkConnectionError(e2); is {
					// NewExponentialBackOff connection on this specific error
					return opError
				}

				// Other errors are considered permanent
				return backoff.Permanent(e2)
			}
			return nil
		},
		boff,
		func(err error, next time.Duration) {
			lgr.Warn().Str("target", target).
				Stringer("next_retry", next.Truncate(time.Millisecond)).
				AnErr("error", err).
				Msg("Failed to open connection")
		},
	)

	if err != nil {
		lgr.Error().Str("target", target).Err(err).Msg("Connection failed")
		return err
	}

	return nil
}

func NetworkConnectionError(err error) (opError *net.OpError, matched bool) {
	if errors.As(err, &opError) && opError.Op == "dial" {
		return opError, true
	}
	return nil, false
}
