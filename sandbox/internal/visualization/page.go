package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// CleanPath is the path a view is recorded under: the query string and the
// fragment dropped, at most MaxPathLength bytes, "/" for an empty one.
func CleanPath(sandbox *api.Sandbox, path string) string {
	for index := 0; index < len(path); index++ {
		if path[index] == '?' || path[index] == '#' {
			path = path[:index]
			break
		}
	}
	if path == "" {
		path = "/"
	}
	return truncate(path, MaxPathLength)
}

// CanonicalPage is the page a view of path is counted under, so the spellings
// the website answers one file by are one page:
//
//	/, /index.html                -> /
//	/blog/, /blog/index.html      -> /blog/
//	/pagina, /pagina.html         -> /pagina
//
// A not-found view keeps the path requested: it names no page.
func CanonicalPage(sandbox *api.Sandbox, path string, status int) string {
	strings := sandbox.Deps.StringsDeps
	path = CleanPath(sandbox, path)
	if status == StatusNotFound {
		return path
	}
	if path == "/"+IndexFile || strings.HasSuffix(path, "/"+IndexFile) {
		return strings.TrimSuffix(path, IndexFile)
	}
	if strings.HasSuffix(path, PageExtension) && len(path) > len(PageExtension)+1 {
		return strings.TrimSuffix(path, PageExtension)
	}
	return path
}

// IndexFile is the file a path ending in / is answered with.
const IndexFile = "index.html"

// PageExtension is what a path naming no file is answered with appended.
const PageExtension = ".html"
