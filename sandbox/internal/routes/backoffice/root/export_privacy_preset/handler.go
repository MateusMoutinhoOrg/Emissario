package export_privacy_preset

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers GET /admin/root/export-privacy-preset?preset=: the preset's
// document as a JSON download, to import on another server. 404 for a
// preset no one is named.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	preset, ok, err := privacyadmin.FindPreset(sandbox, input.Preset)
	if err != nil {
		return err
	}
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusNotFound, "preset", "No preset is named "+input.Preset+".")
	}
	response.SetHeader("Content-Disposition", "attachment; filename=\""+preset.Id+"-v"+sandbox.Deps.StringsDeps.FormatInt(preset.Version, 10)+".json\"")
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, privacyadmin.PresetJSON(sandbox, preset))
}
