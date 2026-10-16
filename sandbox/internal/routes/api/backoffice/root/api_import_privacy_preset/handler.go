package api_import_privacy_preset

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/import-privacy-preset: json, a preset
// document as export-privacy-preset answers it, is checked whole — its
// format, every setting in range, no unknown key — and stored, not applied.
// 400 names the part refused, 409 a version already stored. Answers
// {preset}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	preset, refusal, err := privacyadmin.ImportPreset(sandbox, input.Body.Json)
	if err != nil {
		return err
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusCreated, backofficeapi.PresetRefJSON(sandbox, preset))
}
