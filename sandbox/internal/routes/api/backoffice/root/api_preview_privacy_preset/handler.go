package api_preview_privacy_preset

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/preview-privacy-preset: what applying
// the preset would change — every setting that differs, the notices it asks
// for, what the next cleanup would delete — writing nothing. 404 for a
// preset no one is named.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	preset, ok, err := privacyadmin.FindPreset(sandbox, input.Body.Preset)
	if err != nil {
		return err
	}
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusNotFound, "preset", "No preset is named "+input.Body.Preset+".")
	}
	preview, err := privacyadmin.PreviewPreset(sandbox, preset)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.PreviewJSON(sandbox, preview))
}
