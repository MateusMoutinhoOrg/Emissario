package websitesource

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	opinionatedagnosfront "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/OpinionatedAgnosFront"
)

const (
	// IndexPage is the file a path ending in / is answered with, / included.
	IndexPage = "index.html"
	// PageExtension is what a path naming no file is tried again with, so
	// /pagina answers pagina.html.
	PageExtension = ".html"
	// NotFoundPage is the file a path naming no file is answered with, under
	// a 404: the one of the site, else the default one embedded at
	// assets/front/not-found.html.
	NotFoundPage = "not-found.html"
	// BackofficeAssets is the folder of the embedded assets/front tree the
	// backoffice pages load their script from. A file of it is always the
	// embedded one, so no repository can put its own script in the pages of
	// the backoffice.
	BackofficeAssets = "backoffice/"
)

// Page is one answer of the public website.
type Page struct {
	// Relative is the path of the file answered, relative to the folder it
	// was read from; its extension is the Content-Type.
	Relative string
	// Content is the whole file.
	Content []byte
	// Status is api.StatusOK, or api.StatusNotFound for a not-found page.
	Status int
}

// Resolve answers the request path of the public website out of SiteDir,
// <fonts>/frontend/frontend, the frontend folder of the checkout:
//
//	/           -> <fonts>/frontend/frontend/index.html
//	/blog/      -> <fonts>/frontend/frontend/blog/index.html
//	/blog/a.css -> <fonts>/frontend/frontend/blog/a.css
//	/blog       -> <fonts>/frontend/frontend/blog, else .../blog.html
//
// A path naming no file is answered under a 404 with the not-found.html of
// SiteDir, else with the not-found.html embedded at assets/front; false when
// that one is gone too. The path goes through OpinionatedAgnosFront.SafePath
// first, so nothing outside SiteDir is ever read: no .. segment, no dotfile.
// A file of BackofficeAssets is read out of the binary.
func Resolve(sandbox *api.Sandbox, requested string) (Page, bool) {
	strings := sandbox.Deps.StringsDeps
	relative, safe := sandbox.Deps.OpinionatedAgnosFront.SafePath(requested)
	if safe {
		if strings.HasPrefix(relative, BackofficeAssets) {
			if content, ok := readEmbedded(sandbox, relative); ok {
				return Page{Relative: relative, Content: content, Status: api.StatusOK}, true
			}
		}
		file := relative
		if relative == "" {
			file = IndexPage
		} else if strings.HasSuffix(requested, "/") {
			file = relative + "/" + IndexPage
		}
		if content, ok := readSite(sandbox, file); ok {
			return Page{Relative: file, Content: content, Status: api.StatusOK}, true
		}
		if file == relative {
			page := relative + PageExtension
			if content, ok := readSite(sandbox, page); ok {
				return Page{Relative: page, Content: content, Status: api.StatusOK}, true
			}
		}
	}

	if content, ok := readSite(sandbox, NotFoundPage); ok {
		return Page{Relative: NotFoundPage, Content: content, Status: api.StatusNotFound}, true
	}
	if content, ok := readEmbedded(sandbox, NotFoundPage); ok {
		return Page{Relative: NotFoundPage, Content: content, Status: api.StatusNotFound}, true
	}
	return Page{}, false
}

// readSite reads the file at the slash-separated relative, a path SafePath
// already kept inside SiteDir, false when it is no file.
func readSite(sandbox *api.Sandbox, relative string) ([]byte, bool) {
	io := sandbox.Deps.IoDeps
	path := io.Join(append([]string{SiteDir(sandbox)}, sandbox.Deps.StringsDeps.Split(relative, "/")...)...)
	if !io.IsFile(path) {
		return nil, false
	}
	content, err := io.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return content, true
}

// readEmbedded reads relative out of the embedded assets/front tree, false
// when there is no such asset.
func readEmbedded(sandbox *api.Sandbox, relative string) ([]byte, bool) {
	content, err := sandbox.Deps.EmbedDeps.ReadFile(opinionatedagnosfront.Root + "/" + relative)
	if err != nil {
		return nil, false
	}
	return content, true
}
