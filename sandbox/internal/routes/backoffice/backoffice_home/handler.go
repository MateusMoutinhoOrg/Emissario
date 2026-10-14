package backoffice_home

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/home with backoffice/home.html,
// rendered for the user the backoffice-session-auth middleware put on props.User,
// with the views of the public website today and in the last 7 days, and the
// contact requests in the inbox and received today.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	today, week, err := visualization.HomeCounts(sandbox)
	if err != nil {
		return err
	}
	inbox, received, err := contact.HomeCounts(sandbox)
	if err != nil {
		return err
	}
	return backofficerender.RenderHomePage(sandbox, response, props.User, backofficeauth.SessionSeconds/60, today, week, inbox, received)
}
