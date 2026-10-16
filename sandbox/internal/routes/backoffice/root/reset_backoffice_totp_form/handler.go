package reset_backoffice_totp_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeusers"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// Handle answers POST /admin/root/reset-backoffice-totp: the user of id signs
// in with their password alone until they set two-factor sign-in up again.
// It redirects to the user list with what happened.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	notice := backofficerender.TotpNoticeReset
	if _, ok := backoffice_db.New(sandbox).FindBackofficeUserById(int64(input.Body.Id)); !ok {
		notice = backofficeusers.NoticeNotFound
	} else if err := totp.Reset(sandbox, int64(input.Body.Id)); err != nil {
		return err
	}
	privacyadmin.InvalidateBanners(sandbox)
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, "/admin/list-backoffice-users?notice="+notice)
}
