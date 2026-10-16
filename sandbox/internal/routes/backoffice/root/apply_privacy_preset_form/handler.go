package apply_privacy_preset_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/apply-privacy-preset: every setting of the
// preset is saved, after the notices it needs. Until each is ticked the
// preview answers again under a 409, and nothing is written. Whether data
// is collected is never changed, and nothing records who applied it.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	preset, refusal, err := privacyadmin.ApplyPreset(sandbox, input.Body.Preset, input.Body.Acknowledge)
	if err != nil {
		return err
	}
	if refusal != nil && len(refusal.Notices) > 0 {
		return backofficerender.RenderPresetPreviewPage(sandbox, response, refusal.Status, props.User, preset, "Tick every notice to apply this preset.")
	}
	if refusal != nil {
		return back(sandbox, response, backofficerender.PresetNoticeRefused, input.Body.Preset, refusal.Message)
	}
	return back(sandbox, response, backofficerender.PresetNoticeApplied, preset.Ref(), "")
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
