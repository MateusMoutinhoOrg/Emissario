package backoffice_home

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/home with backoffice/home.html,
// rendered for the user the backoffice-session-auth middleware put on props.User.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	return backofficerender.RenderHomePage(sandbox, response, props.User, backofficeauth.SessionSeconds/60)
}
