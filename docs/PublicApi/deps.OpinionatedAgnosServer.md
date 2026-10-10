# `deps.OpinionatedAgnosServer`

`sandbox/deps/OpinionatedAgnosServer`

| Constant | Value | Description |
| --- | --- | --- |
| `PathString` | `iota` | PathString takes any slice, bound as a string. |
| `PathInteger` |  | PathInteger takes one segment reading as a whole number, bound as an int. |
| `PathNumber` |  | PathNumber takes one segment reading as a number, bound as a float64. |
| `PathUuid` |  | PathUuid takes one segment reading as a canonical uuid, bound as a string. |
| `SourceHeader` | `iota` | SourceHeader reads a request header, matched without regard to case. |
| `SourceQuery` |  | SourceQuery reads a query-string parameter. |
| `SourceCookie` |  | SourceCookie reads a request cookie. |
| `ParameterString` | `iota` | ParameterString is bound as a string. |
| `ParameterNumber` |  | ParameterNumber is bound as a float64. |
| `ParameterBoolean` |  | ParameterBoolean is bound as a bool: true/1 or false/0. |
| `ParameterDateTime` |  | ParameterDateTime is bound as a string that has to read as RFC 3339. |
| `ParameterStringArray` |  | ParameterStringArray is bound as a []string: every occurrence of a query key, or a header's comma-separated values. |
| `ParameterInteger` |  | ParameterInteger is bound as an int. |
| `ParameterIntegerArray` |  | ParameterIntegerArray is bound as a []int, read the way a ParameterStringArray is. |
| `AnyMethod` | `"ANY"` | AnyMethod is the one entry of Methods that accepts every http method. |
| `StatusOK` | `200` | StatusOK reports that the route did what it was asked to do. |
| `StatusCreated` | `201` | StatusCreated reports that the route created what it was asked for. |
| `StatusNoContent` | `204` | StatusNoContent reports success with nothing to send back. |
| `StatusMovedPermanently` | `301` | StatusMovedPermanently sends the caller to Location for good. |
| `StatusFound` | `302` | StatusFound sends the caller to Location this once. |
| `StatusSeeOther` | `303` | StatusSeeOther sends the caller to Location with a GET, after a form was handled. |
| `StatusNotModified` | `304` | StatusNotModified reports that the caller's cached copy is current. |
| `StatusTemporaryRedirect` | `307` | StatusTemporaryRedirect sends the caller to Location this once, keeping its method. |
| `StatusPermanentRedirect` | `308` | StatusPermanentRedirect sends the caller to Location for good, keeping its method. |
| `StatusBadRequest` | `400` | StatusBadRequest reports a request the dispatch could not bind: a missing required field, an unparsable value, one out of range, or a body the declared schema rejects. |
| `StatusUnauthorized` | `401` | StatusUnauthorized reports a request that carries no valid credentials. |
| `StatusForbidden` | `403` | StatusForbidden reports credentials that are valid and not enough. |
| `StatusNotFound` | `404` | StatusNotFound reports that no declared route matches the path. |
| `StatusMethodNotAllowed` | `405` | StatusMethodNotAllowed reports a path a route matches under another method. |
| `StatusConflict` | `409` | StatusConflict reports a well-formed request the current state refuses. |
| `StatusPayloadTooLarge` | `413` | StatusPayloadTooLarge reports a body longer than the route's max-bytes. |
| `StatusUnsupportedMediaType` | `415` | StatusUnsupportedMediaType reports a content type the route does not declare. |
| `StatusUnprocessableEntity` | `422` | StatusUnprocessableEntity reports a well-formed request whose content breaks a rule of the domain. |
| `StatusTooManyRequests` | `429` | StatusTooManyRequests reports a caller over its rate. |
| `StatusInternalServerError` | `500` | StatusInternalServerError reports a well-formed request the route could not carry out. |
| `StatusServiceUnavailable` | `503` | StatusServiceUnavailable reports a server that cannot answer right now. |

## `PathType`

PathType is what one segment a Path reads has to convert to. A segment that will not is a non-match: the url is for some other route.

`type PathType int`

## `Path`

Path is one entry of `paths` in route.yaml: the slice of request segments from Start to End, both inclusive, End -1 standing for the last segment. The slice reads as "/" followed by its segments joined by "/", and is bound to the Input field tagged with its Id.

| Field | Type | Description |
| --- | --- | --- |
| `Id` | `string` | Id is the Input field the slice is bound to. |
| `Start` | `int` | Start is the index of the first segment of the slice. |
| `End` | `int` | End is the index of the last segment of the slice, -1 for the last segment of the request. |
| `Type` | `PathType` | Type is what the slice converts to; anything but PathString reads one segment alone. |
| `Description` | `string` | Description is the one-line help text. |
| `Trigger` | `opinionatedagnoscli.Trigger` | Trigger is what the slice has to match for the route to run. |

## `ParameterSource`

ParameterSource is one place of the request a Parameter is read from.

`type ParameterSource int`

## `ParameterType`

ParameterType is the type a Parameter is converted to before it reaches Input.

`type ParameterType int`

## `Parameter`

Parameter is one entry of `parameters` in route.yaml: one value read off the request under Key, from the first of Sources that carries it, and bound to the Input field tagged with its Id.

| Field | Type | Description |
| --- | --- | --- |
| `Id` | `string` | Id is the Input field the value is bound to. |
| `Key` | `string` | Key is the query key or the header name the value is read under. |
| `Sources` | `[]ParameterSource` | Sources are the places the value is read from, in order: the first one that brings a value wins. |
| `Required` | `bool` | Required reports that the request is answered 400 without it. |
| `Type` | `ParameterType` | Type is what the value is converted to. |
| `Default` | `string` | Default is the value bound when the request brings none, spelled as route.yaml writes it; HasDefault tells an empty default from none. |
| `HasDefault` | `bool` |  |
| `Trigger` | `opinionatedagnoscli.Trigger` | Trigger is what the value has to match for the route to run. |
| `Description` | `string` | Description is the one-line help text. |

## `RouteBody`

RouteBody is the request body a route declares — the parsed form of `body:` in its route.yaml. The dispatch enforces ContentType and MaxBytes before the handler runs; reading the body itself is the handler's to ask for, through the ReadBody its own package generates.

| Field | Type | Description |
| --- | --- | --- |
| `Type` | `string` | Type is the declared type: "none", "raw", "text", "json" or "form". |
| `Required` | `bool` | Required reports that an absent or empty body is rejected. |
| `MaxBytes` | `int` | MaxBytes is the largest body the route reads; a longer one is 413. |
| `ContentType` | `string` | ContentType is the media type the route accepts; a divergent one is 415. It is empty when the route accepts any. |
| `Schema` | `string` | Schema is the declared json-schema in canonical form, "" when the route declares none. |

## `RouteFailure`

RouteFailure is one way a request did not get answered: what the dispatch would have written, and why. Every failure the server layer raises — a header that will not bind, a body the schema rejected, a handler that returned an error, a path nothing matched — arrives at one of the project's own Handle* files as this, read off the Failure of the route it is handed. It is an error too: a Handle refuses a request by returning one, built by Fail, and the dispatch answers it through the Handle* file of its Status.

| Field | Type | Description |
| --- | --- | --- |
| `Status` | `int` | Status is the http status the failure carries: one of the Status* constants of server.go. |
| `Field` | `string` | Field is the header, query parameter, path segment or json path that failed, "" when the failure names no single value. |
| `Message` | `string` | Message is the one-line reason, written for whoever called the route. |
| `Cause` | `string` | Cause is what went wrong underneath — an error's text, or the value a handler panicked with — "" when there is nothing below the message. It is for the log, not for the caller. |

## `Route`

Route is one http route of the project, as the sandbox offers it: the whole of what its route.yaml declares, plus the handler behind it. Server.Routes holds one per directory under sandbox/internal/routes holding a route.yaml, at any depth, each built by that package's generated NewRoute, in run order — lowest Priority first. What Server.Routes holds is the declaration alone: nothing is ever bound onto it. The dispatch copies it with BindRoute, puts the request and the response on the copy, and hands the copy to Matches and Run.

| Field | Type | Description |
| --- | --- | --- |
| `Name` | `string` | Name is the package directory of the route, snake_case. |
| `Methods` | `[]string` | Methods are the http methods it answers to ("GET", "POST"), or AnyMethod alone for every one. |
| `Priority` | `int` | Priority is the rung this route runs on when several match one request: the dispatch runs them from the lowest upwards and stops at the first handler that sets a status. |
| `ResponseType` | `string` | ResponseType is the Content-Type set on the response before the handler runs; the handler may set another. |
| `Segments` | `int` | Segments is how many segments the request path has to have for the route to run, 0 for any count. |
| `Pattern` | `string` | Pattern is its path as it reads in docs and messages. |
| `Category` | `string` | Category groups it on the generated Routes page. |
| `Summary` | `string` | Summary is the one-line description. |
| `Description` | `string` | Description is the paragraph the Routes page prints. |
| `Examples` | `[]string` | Examples are whole requests the Routes page prints. |
| `Hidden` | `bool` | Hidden keeps it off the Routes page without disabling it. |
| `Paths` | `[]Path` | Paths are the slices of the request path it reads, in declaration order. |
| `Parameters` | `[]Parameter` | Parameters are the headers and query parameters it reads, in declaration order. |
| `Body` | `RouteBody` | Body is the request body declaration. |
| `ReadBody` | `func(bound *Route) (any, error)` | ReadBody reads, validates and converts the request body of one bound copy — the route package's own generated ReadBody, closed over the sandbox. Run calls it before the handler runs and binds what it returns onto Input.Body; it is nil on a route whose body is `none`. A body that fails comes back as a *RouteFailure, which the dispatch raises. |
| `Handle` | `any` | Handle is the route package's own Handle, closed over the sandbox: a func(props *routeprops.RouteProps, entries *Input, response *serverdeps.Response) error whose Input is that package's generated struct. It is held as any because every route's Input is a type of its own; the dispatch builds and fills one by reflection and calls it. |
| `Request` | `serverdeps.Request` | Request is the http request this copy was bound from and Response the one being written, the response a handler is handed a pointer to. |
| `Response` | `serverdeps.Response` |  |
| `Props` | `any` | Props is one request's *routeprops.RouteProps, shared by every route of the chain that runs for it and handed to each Handle as its first argument: what a middleware sets on it, the routes after it read. It is held as any because the project types it under sandbox/internal, which no contract may name; a Handle* file reads it back with route.Props.(*routeprops.RouteProps). |
| `Failure` | `*RouteFailure` | Failure is why this route is being handed to one of the project's Handle* files, nil on a normal run. The dispatch sets it as it raises, which is the one way any part of the server layer raises a failure. |
| `Matches` | `func(bound *Route) bool` | Matches reports whether one bound copy — its Request set — is for this route: the method is accepted, and every path slice and every parameter declaring a trigger matches it. Nil reads the copy with the lib's own matcher; set it to decide by hand. |
| `MatchesPath` | `func(bound *Route) bool` | MatchesPath reports whether one bound copy's request path is for this route whatever its method — what tells a 405 from a 404. Nil reads it with the lib's own matcher. |
| `Run` | `func(bound *Route) error` | Run binds one bound copy's request onto a fresh Input and runs Handle with it. It returns the failure the handler did not answer itself, nil otherwise; what it answered with is the status it wrote, and a handler writing none hands the request to the next route of the chain. Nil binds with the lib's own binder. |

## `Server`

Server is the http surface of the sandbox: every route the project declares, and the dispatch that reads a request against them. It is built by sandbox/internal/server/generated.new.go, generated by the build.

| Field | Type | Description |
| --- | --- | --- |
| `Serve` | `func(props ServeProps) error` | Serve opens the server: it hands the props to the lib's Main against these Routes, and blocks until the server stops. |
| `Routes` | `[]*Route` | Routes is every http route the project declares, in run order — lowest `priority` first, then by name — each built by the generated NewRoute of its own package. The dispatch reads a request against these declarations; a caller holding the sandbox reads the same surface without one. |
| `Fail` | `func(route *Route) error` | Fail answers one failure with the project's own handler for it: it reads route.Failure and calls the matching Handle* of sandbox/internal/server/errors/. The dispatch raises through it; a handler returns a failure built by Fail rather than calling it directly. |

## `ServeProps`

ServeProps describes one run of the http server: the address to listen on — host:port (":8080"), a bare port ("8080"), or a range of ports tried in turn, with or without a host ("3000:4000", "127.0.0.1:3000:4000") — the two timeouts, in milliseconds, a request and a response are held to, and how long the requests in flight get to finish once the process is asked to stop (0 waits for them).

| Field | Type |
| --- | --- |
| `Addr` | `string` |
| `ReadTimeoutMs` | `int` |
| `WriteTimeoutMs` | `int` |
| `ShutdownTimeoutMs` | `int` |

## `MainProps`

MainProps is everything one run of Main reads: the surface, the address and timeouts, and the deps the server reaches the outside world through. The generated registry builds it per call, so nothing the lib holds goes stale.

| Field | Type | Description |
| --- | --- | --- |
| `Server` | `*Server` | Server is the surface every request is read against. It is a pointer, read as each request runs, so a caller that replaced a field of sandbox.Server is followed. |
| `Serve` | `ServeProps` | Serve is the address and timeouts the server runs with. |
| `NewProps` | `func() any` | NewProps builds the one *routeprops.RouteProps every route of one request's chain is handed — a type of the project, which is why the lib is handed a constructor rather than naming it. |
| `MatchTrigger` | `func(trigger opinionatedagnoscli.Trigger, text string, segmented bool) bool` | MatchTrigger is how a path or a parameter is held to its trigger: the OpinionatedAgnosCli lib's own, shared with the cli layer. |
| `StdDeps` | `*stddeps.Contract` | StdDeps is the channel the server reports on, read at the moment of each print. |
| `ServerDeps` | `serverdeps.Contract` | ServerDeps opens the port; it routes nothing. |
| `SignalDeps` | `signaldeps.Contract` | SignalDeps hears the request to stop, which shuts the server down. |

## `Contract`

Contract is the server lib injected whole as the Deps.OpinionatedAgnosServer field.

| Field | Type | Description |
| --- | --- | --- |
| `Main` | `func(props MainProps) error` | Main opens the port through props.ServerDeps — which routes nothing — and runs every request through the chain of props.Server.Routes: each route the request is for runs, lowest `priority` first, until one of them answers — sets a status, or writes a byte, which sends a 200. Every way a request can end without an answer — nothing matched, matched under another method, a value that will not bind, a panic — is raised through props.Server.Fail. Addr may name a range of ports ("3000:4000"), tried in turn; the address it landed on is printed to stdout. The first request to stop the process shuts it down gracefully, after which it returns nil. |
| `NewRoute` | `func() *Route` | NewRoute returns an empty Route with every slice open, the base a route's generated NewRoute fills its declaration on. |
| `BindRoute` | `func(route *Route) *Route` | BindRoute copies one declaration into the route a single request runs on: the same declared fields — the slices are read-only and shared — with no request, response or failure yet. |
| `Fail` | `func(status int, field string, message string) error` | Fail is how a Handle — or a generated ReadBody — refuses a request: it builds the failure and returns it as an error, which the dispatch raises on the route, reaching the project's own Handle* file for that status: return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "authorization", "invalid token") A message left empty is filled by that file's own wording. |
| `FailWithCause` | `func(status int, field string, message string, cause string) error` | FailWithCause is Fail carrying what went wrong underneath — an error's text, meant for the log, never for the caller. |
| `FailureOf` | `func(route *Route, status int, message string) RouteFailure` | FailureOf is the failure one of the project's Handle* files is answering. A file stands for one status and one wording, and passes them as the fallback: a failure that carries its own — a field that would not bind, a body the schema rejected — answers with that, and one that carries none — "nothing matched" is raised with no message at all — answers with the file's. |
| `WriteError` | `func(serializables serializabledeps.Contract, response serverdeps.Response, status int, field string, message string) error` | WriteError writes one failure onto the response: the body is always the same JSON object — {"error": "...", "field": "..."} — so a client parses one shape whatever went wrong. It is the writer, not the policy: what decides whether a failure is answered this way is the project's own Handle* file. It returns the message as an error, so a Handle* file answers and reports in one line. |
| `WriteJSON` | `func(serializables serializabledeps.Contract, response serverdeps.Response, status int, document *serializabledeps.SerializableObject) error` | WriteJSON answers with one document serialized as JSON. |
| `WriteText` | `func(response serverdeps.Response, status int, text string) error` | WriteText answers with one text as text/plain. |
| `Redirect` | `func(response serverdeps.Response, status int, location string) error` | Redirect answers by sending the caller to location, with one of the redirect statuses: StatusFound, StatusSeeOther, StatusMovedPermanently, StatusTemporaryRedirect or StatusPermanentRedirect. |
| `ValidateSchema` | `func(serializables serializabledeps.Contract, schemaJson string, body []byte) (*serializabledeps.SerializableObject, string, string, bool)` | ValidateSchema parses body as JSON and checks it against schemaJson, the canonical form of a route's declared json-schema — a subset: type, nullable, const, enum, required, properties, additionalProperties, items, min/maxItems, uniqueItems, min/maxLength, pattern, format (email, uuid, date-time, uri) and the four numeric bounds. It returns the parsed document, the field path of the first violation, that violation's message, and whether the body passed. |
| `ValidateForm` | `func(serializables serializabledeps.Contract, schemaJson string, form map[string][]string) (*serializabledeps.SerializableObject, string, string, bool)` | ValidateForm is ValidateSchema for a body that arrives as `key=value` pairs: each declared key is converted to what its property declares — an integer or a number parsed, a boolean read as true/1/on or false/0/off, an array taking every occurrence — before the schema runs. An empty value counts as absent. |
| `ReadString` | `func(object *serializabledeps.SerializableObject, key string) string` | ReadString returns the named property of a validated document as text; every reader below returns the zero value for a property that is absent or of another kind, since the schema has been enforced by then. |
| `ReadInt` | `func(object *serializabledeps.SerializableObject, key string) int` | ReadInt returns the named property as an int. |
| `ReadFloat` | `func(object *serializabledeps.SerializableObject, key string) float64` | ReadFloat returns the named property as a float64. |
| `ReadBool` | `func(object *serializabledeps.SerializableObject, key string) bool` | ReadBool returns the named property as a bool. |
| `ReadObject` | `func(object *serializabledeps.SerializableObject, key string) *serializabledeps.SerializableObject` | ReadObject returns the named property as a document to read further, nil when it is absent. |
| `ReadItems` | `func(object *serializabledeps.SerializableObject, key string) []*serializabledeps.SerializableObject` | ReadItems returns the named property's items, in order. |
| `ItemString` | `func(item *serializabledeps.SerializableObject) string` | ItemString reads one array item as text. |
| `ItemInt` | `func(item *serializabledeps.SerializableObject) int` | ItemInt reads one array item as an int. |
| `ItemFloat` | `func(item *serializabledeps.SerializableObject) float64` | ItemFloat reads one array item as a float64. |
| `ItemBool` | `func(item *serializabledeps.SerializableObject) bool` | ItemBool reads one array item as a bool. |

| Function | Description |
| --- | --- |
| `Error() string` | Error is the failure's Message, which is what makes a RouteFailure an error a handler can return. |

[every contract](doc.md)
