package preview_privacy_preset_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/preview-privacy-preset?preset=: every
// setting now and in the preset, with its reference, what happens to the
// data stored, and the notices applying it asks for, beside Apply. Nothing
// is written. A preset no one is named answers the Presets page under a 404.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	preset, ok, err := privacyadmin.FindPreset(sandbox, input.Preset)
	if err != nil {
		return err
	}
	if !ok {
		return backofficerender.RenderPrivacyPresetsPage(sandbox, response, api.StatusNotFound, props.User, backofficerender.PresetNoticeRefused, "", "No preset is named "+input.Preset+".")
	}
	return backofficerender.RenderPresetPreviewPage(sandbox, response, api.StatusOK, props.User, preset, "")
}
