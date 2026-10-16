package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
)

// LoginPage is what backoffice/login.html is rendered with.
type LoginPage struct {
	// Error is shown above the form, "" for none.
	Error string
	// Username refills the login field after a failed attempt.
	Username string
	// Pending is the token of a sign-in whose password was accepted and
	// whose two-factor code is awaited: the page asks for the code alone.
	Pending string
}

// RenderLoginPage answers the login page under status, with message above the form.
func RenderLoginPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, message string, username string) error {
	return RenderHTML(sandbox, response, status, "backoffice/login.html", LoginPage{Error: message, Username: username})
}

// RenderLoginCodePage answers the step of the sign-in that asks for the
// two-factor code under status, with message above the form and pending —
// the token of the password accepted — in the form.
func RenderLoginCodePage(sandbox *api.Sandbox, response *serverdeps.Response, status int, message string, pending string) error {
	return RenderHTML(sandbox, response, status, "backoffice/login.html", LoginPage{Error: message, Pending: pending})
}
