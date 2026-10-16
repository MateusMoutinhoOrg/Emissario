package api_reset_backoffice_totp

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// Handle answers POST /api/admin/root/reset-backoffice-totp: the user of id
// signs in with their password alone until they set two-factor sign-in up
// again. 404 for an id no user holds.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if _, ok := backoffice_db.New(sandbox).FindBackofficeUserById(int64(input.Body.Id)); !ok {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusNotFound, "id", "no backoffice user has this id")
	}
	if err := totp.Reset(sandbox, int64(input.Body.Id)); err != nil {
		return err
	}
	privacyadmin.InvalidateBanners(sandbox)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.OkJSON(sandbox))
}
