package api_detach_privacy_preset

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/detach-privacy-preset: the preset
// applied is forgotten and the server is Custom, its settings as they are.
// Answers the state check-privacy-preset answers.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if err := privacyadmin.DetachPreset(sandbox); err != nil {
		return err
	}
	state, err := privacyadmin.CheckPreset(sandbox)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.PresetStateJSON(sandbox, state))
}
