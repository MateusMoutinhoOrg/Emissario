package front

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	opinionatedagnosfront "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/OpinionatedAgnosFront"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers GET /<rest...> with the public website: the frontend/ folder
// of the website source a root set in the backoffice, cloned into
// <--fonts>/frontend, read through websitesource.Resolve. A path ending in / is
// answered with the index.html of that folder, any other with the file it
// names, else with that file plus .html (/pagina -> pagina.html); one naming no
// file with the site's not-found.html under a 404, else
// with the default assets/front/not-found.html embedded in the binary. Only
// when that one is gone too is the path declined — nil without answering — so
// the chain goes on and HandleNotFound answers it.
//
// input.FullRoute is read rather than input.Rest because the trailing / is
// what tells /blog/ (blog/index.html) from /blog (a file named blog).
//
// This route is declared with the highest priority number of the project, so
// every api route runs first and this one is the fallback in front of the 404.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	page, ok := websitesource.Resolve(sandbox, input.FullRoute)
	if !ok {
		return nil
	}

	response.SetHeader("Content-Type", sandbox.Deps.OpinionatedAgnosFront.ContentTypeOf(page.Relative))
	response.SetHeader("Cache-Control", opinionatedagnosfront.RevalidateCache)
	response.SetStatus(page.Status)
	response.Write(page.Content)

	return nil
}
