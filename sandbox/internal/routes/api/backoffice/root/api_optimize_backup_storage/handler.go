package api_optimize_backup_storage

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficesnapshots"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// Handle answers POST /api/admin/root/optimize-backup-storage: every
// stored content no snapshot holds anymore starts being removed in the
// background, and 202 is answered at once; list-backups tells when
// it ended, busy turning false. Another backup job running is answered 409.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	if !snapshots.StartOptimize(sandbox) {
		return sandbox.Deps.OpinionatedAgnosServer.WriteError(sandbox.Deps.SerializableDeps, *response, api.StatusConflict, "", backofficesnapshots.BusyMessage)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, backofficesnapshots.StatusAccepted, backofficeapi.JobJSON(sandbox, "optimizing"))
}
