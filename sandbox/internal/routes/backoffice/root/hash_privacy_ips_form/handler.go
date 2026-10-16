package hash_privacy_ips_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/hash-privacy-ips: every full IP the views,
// contact requests and agent usage keep becomes its hash under the current
// IP key, and the full user agents of the views are blanked. It cannot be
// undone; the access log keeps its full IPs. It redirects to the Privacy
// page with how many it hashed.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	hashed, err := privacyadmin.HashStoredIps(sandbox)
	if err != nil {
		return err
	}
	count := sandbox.Deps.StringsDeps.FormatInt(int64(hashed.Total()), 10)
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, privacy.SettingsPath+"?notice="+backofficerender.PrivacyNoticeHashed+"&count="+count)
}
