package set_website_source_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/websitesource"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/set-website-source. The repository url,
// branch and sync time are written on the frontend source, a sync is asked of
// the loop for its next tick, and the browser is sent back to the page. A
// refused value answers the form again, filled with what was sent, under a 400
// with the reason above it. Only a root reaches it: the backoffice-root-guard
// answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings := websitesource.Settings{
		RepoUrl:  input.Body.RepoUrl,
		Branch:   input.Body.Branch,
		SyncTime: input.Body.SyncTime,
	}
	message, err := websitesource.Save(sandbox, websitesource.FrontendName, settings)
	if err != nil {
		return err
	}
	if message != "" {
		return backofficerender.RenderWebsiteSourcePage(sandbox, response, api.StatusBadRequest, props.User, settings, "", message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, websitesource.PagePath+"?notice="+websitesource.NoticeSaved)
}
