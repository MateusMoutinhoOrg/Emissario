package set_totp_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// Handle answers POST /admin/set-totp: a code of the secret the page showed
// turns two-factor sign-in on for the user signed in — their API tokens
// revoked, their other sessions closed — and redirects to the page; a wrong
// code answers the page again under a 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	ok, err := totp.Confirm(sandbox, *props.User, props.Session, input.Body.Code)
	if err != nil {
		return err
	}
	if !ok {
		return backofficerender.RenderTotpPage(sandbox, response, api.StatusBadRequest, props.User, "", "That code does not match. Check that the phone's clock is right, and type the code the app shows now.")
	}
	privacyadmin.InvalidateBanners(sandbox)
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, backofficerender.TotpPath+"?notice="+backofficerender.TotpNoticeEnabled)
}
