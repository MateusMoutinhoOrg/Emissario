package nethttprequest

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
	requestdeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/requestdeps"
)

// requestTimeout bounds one whole round trip — connect, send, and read the
// response — when the caller set no SetTimeout of its own. Without it net/http
// waits forever, so a hung server would hang the calling program with no way
// for the sandbox to intervene: the contract exposes no cancellation, which
// makes the bound this adapter's job.
const requestTimeout = 30 * time.Second

// noRedirect keeps a 3xx as the response: a redirect is never followed, so a
// request never reaches a host its caller did not name.
func noRedirect(request *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// Bind fills deps.Deps.RequestDeps.NewRequest with the HTTP request
// implementation built on the standard library's net/http package.
func Bind(deps *deps.Deps) {
	deps.RequestDeps.NewRequest = newRequest
}

// newRequest builds one HTTP request bound to url.
func newRequest(url string) requestdeps.Request {
	headers := make(map[string]string)
	method := "GET"
	var reqBody []byte
	timeout := requestTimeout

	return requestdeps.Request{
		AddHeader: func(key string, value string) {
			headers[key] = value
		},
		SetMethod: func(m string) {
			method = m
		},
		SetBody: func(body []byte) {
			reqBody = body
		},
		SetTimeout: func(milliseconds int64) {
			if milliseconds <= 0 {
				timeout = requestTimeout
				return
			}
			timeout = time.Duration(milliseconds) * time.Millisecond
		},
		Fetch: func() (requestdeps.Response, error) {
			var bodyReader io.Reader
			if reqBody != nil {
				bodyReader = bytes.NewReader(reqBody)
			}

			req, err := http.NewRequest(method, url, bodyReader)
			if err != nil {
				return requestdeps.Response{}, err
			}

			for k, v := range headers {
				req.Header.Add(k, v)
			}

			client := &http.Client{Timeout: timeout, CheckRedirect: noRedirect}
			resp, err := client.Do(req)
			if err != nil {
				return requestdeps.Response{}, err
			}

			return requestdeps.Response{
				GetStatusCode: func() int {
					return resp.StatusCode
				},
				GetHeader: func(key string) string {
					return resp.Header.Get(key)
				},
				ReadBody: func(size int) ([]byte, error) {
					if size == -1 {
						return io.ReadAll(resp.Body)
					}
					buf := make([]byte, size)
					n, err := io.ReadFull(resp.Body, buf)
					// A body shorter than size is the contract's normal
					// case, not a failure: io.ReadFull reports it as EOF
					// or ErrUnexpectedEOF, and both mean "that was all
					// of it".
					if err == io.EOF || err == io.ErrUnexpectedEOF {
						return buf[:n], nil
					}
					return buf[:n], err
				},
				Close: func() error {
					return resp.Body.Close()
				},
			}, nil
		},
	}
}
