package set_privacy_settings_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/set-privacy-settings: the Privacy page,
// its form filled with the settings saved, with the outcome of the last
// action above it. Only a root reaches it: the backoffice-root-guard answers
// anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	return backofficerender.RenderPrivacySettingsPage(sandbox, response, api.StatusOK, props.User, backofficerender.PrivacySettingsFormOf(sandbox, settings), input.Notice, input.Count, input.Reason, "")
}
