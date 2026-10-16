package detach_privacy_preset_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/detach-privacy-preset: the preset applied
// is forgotten, the settings stay as they are, and the server is Custom. It
// redirects to the Presets page.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if err := privacyadmin.DetachPreset(sandbox); err != nil {
		return err
	}
	return back(sandbox, response, backofficerender.PresetNoticeDetached, "", "")
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
