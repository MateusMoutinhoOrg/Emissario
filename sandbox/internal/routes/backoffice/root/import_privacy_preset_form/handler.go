package import_privacy_preset_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/import-privacy-preset: json, a preset
// document, is checked whole — its format, every setting in range, no
// unknown key — and stored, not applied. It redirects to the Presets page
// with what happened.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	preset, refusal, err := privacyadmin.ImportPreset(sandbox, input.Body.Json)
	if err != nil {
		return err
	}
	if refusal != nil {
		reason := refusal.Message
		if refusal.Field != "" {
			reason = refusal.Field + ": " + reason
		}
		return back(sandbox, response, backofficerender.PresetNoticeRefused, "", "Not imported. "+reason)
	}
	return back(sandbox, response, backofficerender.PresetNoticeImported, preset.Ref(), "")
}

// back sends the browser to the Presets page with the outcome code, about
// preset, and the reason it was refused.
func back(sandbox *api.Sandbox, response *serverdeps.Response, code string, preset string, reason string) error {
	params := []visualization.Param{{Key: "notice", Value: code}, {Key: "preset", Value: preset}}
	if reason != "" {
		params = append(params, visualization.Param{Key: "reason", Value: reason})
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, visualization.Href(sandbox, privacy.PresetsPath, params))
}
