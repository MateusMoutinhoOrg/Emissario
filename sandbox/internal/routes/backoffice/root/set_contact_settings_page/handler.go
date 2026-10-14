package set_contact_settings_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/set-contact-settings: the settings of the
// contact requests, beside how many are stored and the danger zone. Only a
// root reaches it: the backoffice-root-guard answers anyone else under
// /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, err := contact.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	return backofficerender.RenderContactSettingsPage(sandbox, response, api.StatusOK, props.User, backofficerender.ContactSettingsFormOf(sandbox, settings), input.Notice, input.Count, "")
}
