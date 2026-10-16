package reset_backoffice_totp

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// Handle answers `reset-backoffice-totp`: the user --username names — a
// username or an email — signs in with their password alone until they set
// two-factor sign-in up again. It is the way back in for a root who lost
// their phone.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	user, found, err := backofficeauth.FindUserByUsernameOrEmail(sandbox, input.Username)
	if err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "failed to read the backoffice users: "+err.Error())
	}
	if !found {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "username", "no backoffice user "+input.Username)
	}
	if err := totp.Reset(sandbox, user.Id); err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "failed to reset the two-factor sign-in: "+err.Error())
	}
	response.Printf("two-factor sign-in of %s turned off: they sign in with their password until they set it up again on /admin/set-totp\n", user.Username)
	return nil
}
