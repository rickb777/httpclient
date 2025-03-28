package retry_test

import (
	"errors"
	"github.com/rickb777/expect"
	"github.com/rickb777/httpclient/retry"
	"github.com/rickb777/httpclient/testhttpclient"
	"github.com/rs/zerolog"
	"io/ioutil"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRetry_Get_200_error_200(t *testing.T) {
	lgr := zerolog.New(ioutil.Discard)
	stub := testhttpclient.New(t).
		AddResponse("GET", "http://localhost/foobar", testhttpclient.MockResponse(200, []byte("OK"), ""))

	r := retry.New(stub, retry.RetryConfig{}, lgr)

	_, err := r.Do(httptest.NewRequest("GET", "http://localhost/foobar", nil))
	expect.Error(err).Not().ToHaveOccurred(t)

	expect.Slice(stub.RemainingOutcomes()).ToBeEmpty(t)
}

func TestRetry_Get_error(t *testing.T) {
	lgr := zerolog.New(ioutil.Discard)
	e1 := &net.OpError{Op: "dial", Err: errors.New("bang2")}
	e2 := errors.New("bang1")

	stub := testhttpclient.New(t).
		AddError("GET", "http://localhost/foo", e1).
		AddError("GET", "http://localhost/foo", e2)

	r := retry.New(stub, retry.RetryConfig{}, lgr)

	_, err := r.Do(httptest.NewRequest("GET", "http://localhost/foo", nil))
	expect.Any(err).ToBe(t, e2)

	expect.Slice(stub.RemainingOutcomes()).ToBeEmpty(t)
}

func TestNewExponentialBackOff(t *testing.T) {
	count := 0
	lgrBuf := &strings.Builder{}
	lgr := zerolog.New(lgrBuf)
	config := retry.RetryConfig{
		ConnectTimeout: time.Hour,
		ConnectTries:   10,
	}

	err := retry.NewExponentialBackOff(config, "TGT", lgr,
		func() error {
			count++
			if count == 1 {
				return &net.OpError{Op: "dial", Err: errors.New("bang")}
			}
			return nil
		})

	expect.Error(err).Not().ToHaveOccurred(t)
	expect.Number(count).ToBe(t, 2)
	msg := lgrBuf.String()
	expect.String(msg).ToContain(t, `"level":"warn"`)
	expect.String(msg).ToContain(t, `"target":"TGT"`)
	expect.String(msg).ToContain(t, `"error":"dial: bang"`)
	expect.String(msg).ToContain(t, `"message":"Failed to open connection"`)
}
