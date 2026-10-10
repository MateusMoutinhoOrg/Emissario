package list_backoffice_users_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeusers"
)

// Handle answers GET /admin/list-backoffice-users, open to every
// backoffice user, with backoffice/backoffice_users.html: one page of the users
// the search and role filters keep, and, for a root, the controls that add,
// edit and remove them.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	listing, err := backofficeusers.List(sandbox, backofficeusers.Query{
		Search: input.Search,
		Role:   input.Role,
		Page:   input.Page,
		Limit:  input.Limit,
		Viewer: props.User,
	})
	if err != nil {
		return err
	}
	return backofficerender.RenderUsersPage(sandbox, response, props.User, listing, input.Notice)
}
