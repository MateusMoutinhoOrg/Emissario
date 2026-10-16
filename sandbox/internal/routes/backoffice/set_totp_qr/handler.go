package set_totp_qr

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// Handle answers GET /admin/set-totp-qr with the QR code of the secret the
// two-factor page of the user signed in shows, as a PNG never cached: 404
// once two-factor sign-in is on, or before the page drew a secret.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	png, ok, err := totp.Qr(sandbox, *props.User)
	if err != nil {
		return err
	}
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusNotFound, "", "no two-factor setup is pending")
	}
	response.SetHeader("Content-Type", "image/png")
	response.SetHeader("Cache-Control", "no-store")
	response.SetStatus(api.StatusOK)
	return response.Write(png)
}
