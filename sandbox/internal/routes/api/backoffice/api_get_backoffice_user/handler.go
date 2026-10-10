package api_get_backoffice_user

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeusers"
)

// Handle answers POST /api/admin/get-backoffice-user, open to
// every backoffice user like the list, with the user whose id the body names,
// or a 404 when there is none.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	user, ok := backofficeusers.Find(sandbox, int64(input.Body.Id))
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusNotFound, "id", "that user does not exist")
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.UserResponseJSON(sandbox, backofficeusers.Masked(sandbox, props.User, user)))
}
