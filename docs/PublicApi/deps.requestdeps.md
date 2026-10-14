# `deps.RequestDeps`

`sandbox/deps/requestdeps`

## `Contract`

Contract is the HTTP-request constructor injected whole as the Deps.RequestDeps field — the same mechanic as iodeps.Contract.

| Field | Type | Description |
| --- | --- | --- |
| `NewRequest` | `func(url string) Request` | NewRequest builds an HTTP request bound to the given url. |

## `Request`

Request is one HTTP request under construction, handed back by Deps.RequestDeps.NewRequest already bound to a url. The setters mutate the pending request and may be called in any order; nothing leaves the machine until Fetch is called, and a Request may be sent more than once. The method defaults to GET and the body to none, so a plain read is NewRequest followed by Fetch.

| Field | Type | Description |
| --- | --- | --- |
| `AddHeader` | `func(key string, value string)` | AddHeader sets one header on the pending request, replacing whatever value that key carried before. |
| `SetMethod` | `func(method string)` | SetMethod sets the HTTP method the request is sent with — "POST", "PUT", "DELETE". It defaults to "GET". |
| `SetBody` | `func(body []byte)` | SetBody sets the bytes sent as the request body. It defaults to none. |
| `SetTimeout` | `func(milliseconds int64)` | SetTimeout bounds the whole round trip of the next Fetch — connect, send and read the response headers and body — to milliseconds. 0, or a value never set, keeps the adapter's default of 30 s. |
| `Fetch` | `func() (Response, error)` | Fetch sends the request and returns the response. The error reports a request that could not be built or a round trip that failed, a timeout included; an HTTP error status is *not* an error here, it is reported by Response.GetStatusCode. A redirect is never followed: a 3xx is the response handed back. A Response returned without an error holds an open body the caller must Close. |

## `Response`

Response is one HTTP response, handed back by Request.Fetch with its body still open. The caller must Close it, whether or not the body is read.

| Field | Type | Description |
| --- | --- | --- |
| `GetStatusCode` | `func() int` | GetStatusCode returns the response's HTTP status code. |
| `GetHeader` | `func(key string) string` | GetHeader returns the first value of one response header, or "" when the response carries no such header. |
| `ReadBody` | `func(size int) ([]byte, error)` | ReadBody reads at most size bytes of the response body, or the whole body when size is -1. A body shorter than size is returned whole rather than reported as an error, so a short read is not a failure. |
| `Close` | `func() error` | Close releases the response body. It must be called for every Response returned without an error, or the underlying connection leaks. |

[every contract](doc.md)
