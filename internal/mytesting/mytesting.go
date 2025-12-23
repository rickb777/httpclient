package mytesting

import (
	"bufio"
	"net/http"

	bodypkg "github.com/rickb777/httpclient/body"
)

type Outcome struct {
	Err error
	Res *http.Response
}

type StubHttp struct {
	Captured []http.Request
	Outcome  []Outcome
}

func StubHttpWithBody(body string) *StubHttp {
	rdr := bufio.NewReader(bodypkg.NewBodyString(body))
	res, err := http.ReadResponse(rdr, nil)
	must(err)

	return &StubHttp{Outcome: []Outcome{{Res: res}}}
}

func StubHttpWithError(err error) *StubHttp {
	return &StubHttp{Outcome: []Outcome{{Err: err}}}
}

func (h *StubHttp) ThenWithBody(body string) *StubHttp {
	rdr := bufio.NewReader(bodypkg.NewBodyString(body))
	res, err := http.ReadResponse(rdr, nil)
	must(err)

	h.Outcome = append(h.Outcome, Outcome{Res: res})
	return h
}

func (h *StubHttp) ThenWithError(err error) *StubHttp {
	h.Outcome = append(h.Outcome, Outcome{Err: err})
	return h
}

func (h *StubHttp) Do(req *http.Request) (*http.Response, error) {
	i := len(h.Captured)
	if req.Body != nil {
		b, _ := bodypkg.Copy(req.Body)
		req.Body = b // easier to debug now
	}

	h.Outcome[i].Res.Request = req
	h.Captured = append(h.Captured, *req)

	if h.Outcome[i].Err != nil {
		return nil, h.Outcome[i].Err
	} else {
		return h.Outcome[i].Res, nil
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
