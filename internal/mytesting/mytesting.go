package mytesting

import (
	"bufio"
	"net/http"

	bodypkg "github.com/rickb777/httpclient/body"
)

type StubHttp struct {
	Captured http.Request
	ReqBody  string
	Err      error
	Res      *http.Response
}

func StubHttpWithBody(body string) *StubHttp {
	rdr := bufio.NewReader(bodypkg.NewBodyString(body))
	res, err := http.ReadResponse(rdr, nil)
	must(err)

	return &StubHttp{Res: res}
}

func (h *StubHttp) Do(req *http.Request) (*http.Response, error) {
	h.Captured = *req
	if req.Body != nil {
		b, _ := bodypkg.Copy(req.Body)
		h.ReqBody = b.String()
		req.Body = b
	}
	if h.Err != nil {
		return nil, h.Err
	}
	return h.Res, nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
