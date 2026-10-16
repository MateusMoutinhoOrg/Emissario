package api_set_totp

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// Handle answers POST /api/admin/set-totp for the owner of the token. With no
// code it begins the setup: {"enabled":false,"secret","otpauth_uri"}, or
// {"enabled":true,"enabled_at"} when it is on already. With a code of that
// secret it turns two-factor sign-in on — every API token of the user
// revoked, this one included, and their sessions closed — and answers
// {"enabled":true}; a wrong code answers 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	serializable := sandbox.Deps.SerializableDeps
	document := serializable.CreateObject()
	if totp.Enabled(*props.User) {
		document.AddItemToObject("enabled", true)
		document.AddItemToObject("enabled_at", props.User.TotpEnabledAt)
		return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(serializable, *response, api.StatusOK, document)
	}
	if input.Body.Code == "" {
		setup, err := totp.Begin(sandbox, *props.User)
		if err != nil {
			return err
		}
		document.AddItemToObject("enabled", false)
		document.AddItemToObject("secret", setup.Secret)
		document.AddItemToObject("otpauth_uri", setup.Uri)
		return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(serializable, *response, api.StatusOK, document)
	}
	ok, err := totp.Confirm(sandbox, *props.User, nil, input.Body.Code)
	if err != nil {
		return err
	}
	if !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, "code", "the code does not match the secret, or no setup was begun")
	}
	privacyadmin.InvalidateBanners(sandbox)
	document.AddItemToObject("enabled", true)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(serializable, *response, api.StatusOK, document)
}
