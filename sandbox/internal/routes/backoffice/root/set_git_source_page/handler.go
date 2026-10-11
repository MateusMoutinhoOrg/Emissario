package set_git_source_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/gitsource"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/set-git-source with the form that
// POST /admin/root/set-git-source reads, filled with the repository url,
// branch and sync time of the frontend source — or the defaults, before a root
// saved any — beside what the last sync did. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, ok, err := gitsource.Read(sandbox)
	if err != nil {
		return err
	}
	if !ok {
		settings = gitsource.Defaults(sandbox)
	}
	return backofficerender.RenderGitSourcePage(sandbox, response, api.StatusOK, props.User, settings, input.Notice, "")
}
