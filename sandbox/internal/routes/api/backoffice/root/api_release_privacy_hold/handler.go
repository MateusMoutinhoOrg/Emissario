package api_release_privacy_hold

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/release-privacy-hold: the hold of id is
// released and deleted. 404 for an id no hold has.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	found, refusal, err := privacyadmin.ReleaseHold(sandbox, props.User, props.ApiToken, int64(input.Body.Id))
	if err != nil {
		return err
	}
	if !found {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusNotFound, "id", "no hold has this id")
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.OkJSON(sandbox))
}
