package api_delete_privacy_preset

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/delete-privacy-preset: the custom
// preset, <id>@<version>, is deleted. A built-in preset, or the one applied,
// is refused 400; one no one is named 404. Answers {removed: 1}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	refusal, err := privacyadmin.DeletePreset(sandbox, input.Body.Preset)
	if err != nil {
		return err
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.RemovedJSON(sandbox, 1))
}
