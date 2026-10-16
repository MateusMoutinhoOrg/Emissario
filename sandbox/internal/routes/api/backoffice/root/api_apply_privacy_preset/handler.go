package api_apply_privacy_preset

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/apply-privacy-preset: every setting of
// the preset is saved, after the notices it needs — 409
// acknowledgement-required listing them all until every id is in
// acknowledge, and nothing is written. Whether data is collected is never
// changed. Answers {preset}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	preset, refusal, err := privacyadmin.ApplyPreset(sandbox, input.Body.Preset, input.Body.Acknowledge)
	if err != nil {
		return err
	}
	if refusal != nil && len(refusal.Notices) > 0 {
		notices, err := privacyadmin.LoadPresetNotices(sandbox, preset.Values, refusal.Notices)
		if err != nil {
			return err
		}
		return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusConflict, backofficeapi.AcknowledgementRequiredJSON(sandbox, notices))
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.PresetRefJSON(sandbox, preset))
}
