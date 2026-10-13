package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// excludedRoots are the paths no view is ever recorded on, with everything
// under each: the backoffice, its api, its assets and the two routes the
// server answers itself.
var excludedRoots = []string{"/admin", "/api", "/backoffice", "/health", "/openapi.json"}

// Counted tells whether an answer of the front route is a view: a GET — never
// a HEAD, which the server answers as the GET it asks the headers of — that
// was answered with an HTML page under a 200, or with the not-found page under
// a 404, on a path of the public website. An asset (css, js, an image, a
// font: anything whose Content-Type is not text/html) is no view.
func Counted(sandbox *api.Sandbox, method string, status int, contentType string, path string) bool {
	strings := sandbox.Deps.StringsDeps
	if method != "GET" {
		return false
	}
	if status != StatusFound && status != StatusNotFound {
		return false
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "text/html") {
		return false
	}
	for _, root := range excludedRoots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return false
		}
	}
	return true
}
