package temperror

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/rickb777/expect"
)

func TestDNS_lookup_failed(t *testing.T) {
	t.Parallel()

	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := d.DialContext(ctx, "tcp", "does-not-exist.local:80")

	matched := NetworkConnectionError(err)
	expect.Bool(matched).ToBeFalse(t)
}

func TestConnect_failed(t *testing.T) {
	t.Parallel()

	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := d.DialContext(ctx, "tcp", "127.0.0.1:1")

	matched := NetworkConnectionError(err)
	expect.Bool(matched).ToBeTrue(t)
}
