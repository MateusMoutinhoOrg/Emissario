package agent

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
)

// A url is an endpoint a root lets the agent call, by its name: the model
// never writes a url, it picks one of these. Its headers are added by the
// server and are never shown to the model nor to a viewer; like the api key,
// their values are write-only — a page or a JSON answer shows their names
// alone. Whatever the agent can call, any anonymous visitor can call, with
// these headers.

const (
	// MaxUrls is the most urls the agent may hold.
	MaxUrls = 50
	// MaxUrlNameLength is the longest name of a url.
	MaxUrlNameLength = 64
	// MaxUrlDescriptionLength is the longest description of a url.
	MaxUrlDescriptionLength = 1000
	// MaxUrlLength is the longest url.
	MaxUrlLength = 2048
	// MaxHeaders is the most headers a url carries.
	MaxHeaders = 20
	// MaxHeadersLength is the longest text of headers.
	MaxHeadersLength = 8192
)

// UrlInput is what a form or a JSON body sends to add or change a url; the
// headers are one Name: value per line. Changing a url, blank headers keep the
// ones saved, unless ClearHeaders removes them.
type UrlInput struct {
	Name         string
	Description  string
	Method       string
	Url          string
	Headers      string
	ClearHeaders bool
}

// Header is one header a url is called with.
type Header struct {
	Name  string
	Value string
}

// urlLock is held while a url is added or changed, so two never take one name.
var urlLock = make(chan struct{}, 1)

// ListUrls is every url, by name.
func ListUrls(sandbox *api.Sandbox) ([]agent_db.AgentUrlRecord, error) {
	urls, err := agent_db.New(sandbox).ListAgentUrls(agent_db.AgentUrlFilter{})
	if err != nil {
		return nil, err
	}
	sandbox.Deps.SortDeps.Slice(urls, func(i int, j int) bool { return urls[i].Name < urls[j].Name })
	return urls, nil
}

// FindUrl is the url of id, false when there is none.
func FindUrl(sandbox *api.Sandbox, id int64) (agent_db.AgentUrlRecord, bool) {
	return agent_db.New(sandbox).FindAgentUrlById(id)
}

// AddUrl checks input and adds it as a url, refused when it does not hold,
// when its name is taken or when MaxUrls are there already.
func AddUrl(sandbox *api.Sandbox, input UrlInput) (agent_db.AgentUrlRecord, *Refusal, error) {
	input, refusal := CheckUrl(sandbox, input)
	if refusal != nil {
		return agent_db.AgentUrlRecord{}, refusal, nil
	}
	urlLock <- struct{}{}
	defer func() { <-urlLock }()
	db := agent_db.New(sandbox)
	count, err := db.CountAgentUrl()
	if err != nil {
		return agent_db.AgentUrlRecord{}, nil, err
	}
	if count >= MaxUrls {
		return agent_db.AgentUrlRecord{}, refuse("name", sandbox.Deps.StdDeps.Sprintf("The agent holds at most %d urls.", MaxUrls)), nil
	}
	if _, taken := db.FindAgentUrlByName(input.Name); taken {
		return agent_db.AgentUrlRecord{}, nameTaken(sandbox, input.Name), nil
	}
	record, err := db.AddAgentUrl(agent_db.AgentUrlInput{
		Name:        input.Name,
		Description: input.Description,
		Method:      input.Method,
		Url:         input.Url,
		Headers:     input.Headers,
	})
	return record, nil, err
}

// SetUrl checks input and writes it over the url of id — its headers kept
// when input sends none and does not clear them: 404 when there is none,
// refused when it does not hold or its name is another url's.
func SetUrl(sandbox *api.Sandbox, id int64, input UrlInput) (agent_db.AgentUrlRecord, *Refusal, error) {
	input, refusal := CheckUrl(sandbox, input)
	if refusal != nil {
		return agent_db.AgentUrlRecord{}, refusal, nil
	}
	urlLock <- struct{}{}
	defer func() { <-urlLock }()
	db := agent_db.New(sandbox)
	record, ok := db.FindAgentUrlById(id)
	if !ok {
		return record, urlNotFound(), nil
	}
	if other, taken := db.FindAgentUrlByName(input.Name); taken && other.Id != id {
		return record, nameTaken(sandbox, input.Name), nil
	}
	headers := input.Headers
	if headers == "" && !input.ClearHeaders {
		headers = record.Headers
	}
	for _, write := range []func() error{
		func() error { return db.SetAgentUrlName(id, input.Name) },
		func() error { return db.SetAgentUrlDescription(id, input.Description) },
		func() error { return db.SetAgentUrlMethod(id, input.Method) },
		func() error { return db.SetAgentUrlUrl(id, input.Url) },
		func() error { return db.SetAgentUrlHeaders(id, headers) },
	} {
		if err := write(); err != nil {
			return record, nil, err
		}
	}
	record, _ = db.FindAgentUrlById(id)
	return record, nil, nil
}

// RemoveUrl removes the url of id, 404 when there is none.
func RemoveUrl(sandbox *api.Sandbox, id int64) (*Refusal, error) {
	urlLock <- struct{}{}
	defer func() { <-urlLock }()
	db := agent_db.New(sandbox)
	if _, ok := db.FindAgentUrlById(id); !ok {
		return urlNotFound(), nil
	}
	return nil, db.RemoveAgentUrl(id)
}

// urlNotFound is the 404 of an id no url holds.
func urlNotFound() *Refusal {
	return &Refusal{Status: api.StatusNotFound, Field: "id", Message: "No url has this id."}
}

// nameTaken is the refusal of a name another url holds.
func nameTaken(sandbox *api.Sandbox, name string) *Refusal {
	return refuse("name", "Another url is named "+name+".")
}

// CheckUrl answers input trimmed — the method upper-cased, the headers one
// Name: value per line — or the refusal of the first value that does not hold.
func CheckUrl(sandbox *api.Sandbox, input UrlInput) (UrlInput, *Refusal) {
	strings := sandbox.Deps.StringsDeps
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Method = strings.ToUpper(strings.TrimSpace(input.Method))
	input.Url = strings.TrimSpace(input.Url)

	if len(input.Name) > MaxUrlNameLength || !matches(sandbox, `^[a-z0-9][a-z0-9_-]*$`, input.Name) {
		return input, refuse("name", "The name is a slug of lower-case letters, digits, - and _, at most 64 characters, such as order-status.")
	}
	if input.Description == "" || len([]rune(input.Description)) > MaxUrlDescriptionLength {
		return input, refuse("description", "The description says when the agent should call the url, in at most 1000 characters.")
	}
	if input.Method != "GET" && input.Method != "POST" {
		return input, refuse("method", "The method is GET or POST.")
	}
	if len(input.Url) > MaxUrlLength || hasControl(input.Url, false) ||
		!matches(sandbox, `^https?://[^/?#@\s]+([/?][^#\s]*)?$`, input.Url) {
		return input, refuse("url", "The url is a whole http:// or https:// address, with no user, password nor #fragment.")
	}
	headers, refusal := ParseHeaders(sandbox, input.Headers)
	if refusal != nil {
		return input, refusal
	}
	lines := []string{}
	for _, header := range headers {
		lines = append(lines, header.Name+": "+header.Value)
	}
	input.Headers = strings.Join(lines, "\n")
	return input, nil
}

// forbiddenHeaders are the headers the server sets itself on a call.
var forbiddenHeaders = []string{"host", "content-length", "transfer-encoding", "connection", "content-type"}

// ParseHeaders reads text, one Name: value per line — blank lines skipped —
// as headers, or the refusal of the first line that is not one.
func ParseHeaders(sandbox *api.Sandbox, text string) ([]Header, *Refusal) {
	strings := sandbox.Deps.StringsDeps
	headers := []Header{}
	if len(text) > MaxHeadersLength {
		return nil, refuse("headers", "The headers hold at most 8192 characters.")
	}
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		colon := -1
		for index, char := range line {
			if char == ':' {
				colon = index
				break
			}
		}
		if colon <= 0 {
			return nil, refuse("headers", "Each header is one line Name: value, such as Authorization: Bearer abc.")
		}
		name := strings.TrimSpace(line[:colon])
		value := strings.TrimSpace(line[colon+1:])
		if !matches(sandbox, "^[A-Za-z0-9!#$%&'*+.^_`|~-]+$", name) || hasControl(value, true) {
			return nil, refuse("headers", "The header "+cut(name, 64)+" is not a name of letters, digits and - followed by a value on one line.")
		}
		for _, forbidden := range forbiddenHeaders {
			if strings.ToLower(name) == forbidden {
				return nil, refuse("headers", "The header "+name+" is set by the server.")
			}
		}
		headers = append(headers, Header{Name: name, Value: value})
	}
	if len(headers) > MaxHeaders {
		return nil, refuse("headers", sandbox.Deps.StdDeps.Sprintf("A url carries at most %d headers.", MaxHeaders))
	}
	return headers, nil
}

// HeaderNames is the names of the headers of text, what a list shows.
func HeaderNames(sandbox *api.Sandbox, text string) []string {
	headers, refusal := ParseHeaders(sandbox, text)
	if refusal != nil {
		return nil
	}
	names := []string{}
	for _, header := range headers {
		names = append(names, header.Name)
	}
	return names
}
