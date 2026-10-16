package backoffice_login

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficethrottle"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// Handle answers POST /admin/login. The username field takes a
// username or an email; a match sets a session cookie bound to the client ip
// the request came from and redirects to /admin/home, anything else answers
// the login page again under a 401. The attempt is counted before the
// password is checked, so sign-ins sent at the same time cannot all slip
// under the limit; once the client ip or the account passed its limit of
// failed sign-ins, the login page is answered under a 429 without the
// password being checked, until the window of backofficethrottle closes. A
// login or a password longer than backofficethrottle allows is refused
// unchecked.
//
// A user who turned two-factor sign-in on is answered, once the password is
// accepted, a step that asks for their code, carrying a short-lived token of
// the password accepted — never the password — bound to the client ip; the
// session is opened once the code matches. Each code is counted like a
// password.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if input.Body.Pending != "" {
		return confirmCode(sandbox, props, input, response)
	}
	username := input.Body.Username
	password := input.Body.Password
	tooLong := len(username) > backofficethrottle.MaxLoginLength || len(password) > backofficethrottle.MaxPasswordLength
	if tooLong {
		username = ""
	}

	user, found, err := backofficeauth.FindUserByUsernameOrEmail(sandbox, username)
	if err != nil {
		return err
	}
	if !found {
		user.Id = 0
	}
	reservation := backofficethrottle.ReserveLogin(sandbox, props.ClientIp, backofficethrottle.AccountOf(sandbox, username, user.Id))
	if !reservation.Allowed {
		response.SetHeader("Retry-After", backofficethrottle.RetryAfter(sandbox))
		return backofficerender.RenderLoginPage(sandbox, response, api.StatusTooManyRequests, "Too many failed sign-in attempts. Try again in 15 minutes.", username)
	}
	if tooLong {
		return backofficerender.RenderLoginPage(sandbox, response, api.StatusUnauthorized, "Invalid username or password.", "")
	}

	user, ok, err := backofficeauth.CheckPassword(sandbox, user, found, password)
	if err != nil {
		return err
	}
	if !ok {
		return backofficerender.RenderLoginPage(sandbox, response, api.StatusUnauthorized, "Invalid username or password.", username)
	}
	if totp.Enabled(user) {
		pending, err := totp.IssuePending(sandbox, user, props.ClientIp)
		if err != nil {
			return err
		}
		return backofficerender.RenderLoginCodePage(sandbox, response, api.StatusOK, "", pending)
	}
	backofficethrottle.LoginSucceeded(sandbox, reservation)
	return signIn(sandbox, props, user, response)
}

// confirmCode is the second step of a two-factor sign-in: the code of the
// user the pending token names, counted like a password. A token expired or
// from another ip starts the sign-in over.
func confirmCode(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	user, ok := totp.ResolvePending(sandbox, input.Body.Pending, props.ClientIp)
	if !ok {
		return backofficerender.RenderLoginPage(sandbox, response, api.StatusUnauthorized, "The sign-in expired: sign in again.", "")
	}
	reservation := backofficethrottle.ReserveLogin(sandbox, props.ClientIp, backofficethrottle.AccountOf(sandbox, user.Username, user.Id))
	if !reservation.Allowed {
		response.SetHeader("Retry-After", backofficethrottle.RetryAfter(sandbox))
		return backofficerender.RenderLoginPage(sandbox, response, api.StatusTooManyRequests, "Too many failed sign-in attempts. Try again in 15 minutes.", "")
	}
	matched, err := totp.Check(sandbox, user, input.Body.Code)
	if err != nil {
		return err
	}
	if !matched {
		return backofficerender.RenderLoginCodePage(sandbox, response, api.StatusUnauthorized, "Invalid code.", input.Body.Pending)
	}
	backofficethrottle.LoginSucceeded(sandbox, reservation)
	return signIn(sandbox, props, user, response)
}

// signIn opens a session for user, sets its cookie and redirects home.
func signIn(sandbox *api.Sandbox, props *routeprops.RouteProps, user backoffice_db.BackofficeUserRecord, response *serverdeps.Response) error {
	token, err := backofficeauth.IssueSessionJWT(sandbox, user, props.ClientIp)
	if err != nil {
		return err
	}
	response.AddHeader("Set-Cookie", backofficeauth.SessionCookie(sandbox, token))
	response.SetHeader("Location", "/admin/home")
	response.SetStatus(api.StatusSeeOther)
	return nil
}
