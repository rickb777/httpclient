package retry_test

import (
	"errors"
	. "github.com/onsi/gomega"
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
	g := NewGomegaWithT(t)

	lgr := zerolog.New(ioutil.Discard)
	stub := testhttpclient.New(t).
		AddResponse("GET", "http://localhost/foobar", testhttpclient.MockResponse(200, []byte("OK"), ""))

	r := retry.New(stub, retry.RetryConfig{}, lgr)

	_, err := r.Do(httptest.NewRequest("GET", "http://localhost/foobar", nil))
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(stub.RemainingOutcomes()).To(Equal(0))
}

func TestRetry_Get_error(t *testing.T) {
	g := NewGomegaWithT(t)

	lgr := zerolog.New(ioutil.Discard)
	e1 := &net.OpError{Op: "dial", Err: errors.New("bang2")}
	e2 := errors.New("bang1")

	stub := testhttpclient.New(t).
		AddError("GET", "http://localhost/foo", e1).
		AddError("GET", "http://localhost/foo", e2)

	r := retry.New(stub, retry.RetryConfig{}, lgr)

	_, err := r.Do(httptest.NewRequest("GET", "http://localhost/foo", nil))
	g.Expect(err).To(Equal(e2))

	g.Expect(stub.RemainingOutcomes()).To(Equal(0))
}

func TestNewExponentialBackOff(t *testing.T) {
	g := NewGomegaWithT(t)
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

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(count).To(Equal(2))
	msg := lgrBuf.String()
	g.Expect(msg).To(ContainSubstring(`"level":"warn"`))
	g.Expect(msg).To(ContainSubstring(`"target":"TGT"`))
	g.Expect(msg).To(ContainSubstring(`"error":"dial: bang"`))
	g.Expect(msg).To(ContainSubstring(`"message":"Failed to open connection"`))
}
