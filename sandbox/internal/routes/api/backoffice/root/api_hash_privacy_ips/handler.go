package api_hash_privacy_ips

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/hash-privacy-ips: every full IP the
// views, contact requests and agent usage keep becomes its hash under the
// current IP key, and the full user agents of the views are blanked. It
// cannot be undone; the access log keeps its full IPs. It answers how many
// it hashed.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	hashed, err := privacyadmin.HashStoredIps(sandbox)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.HashedIpsJSON(sandbox, hashed))
}
