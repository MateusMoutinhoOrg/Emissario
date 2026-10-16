package rotate_privacy_ip_key_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/rotate-privacy-ip-key: the IP key is
// replaced with a new random one, so the hashes made before stop matching
// new visits and any IP typed in a filter or in Find and erase. It cannot be
// undone. It redirects to the Privacy page.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if err := privacyadmin.RotateIpKey(sandbox); err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, privacy.SettingsPath+"?notice="+backofficerender.PrivacyNoticeRotated)
}
