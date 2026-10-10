package list_backups_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// Handle answers GET /admin/root/list-backups, open to a
// root only, with backoffice/backup_snapshots.html: every snapshot of the
// databases whose name starts with the prefix query value — every one
// without it — newest first, with the controls that take, upload, download
// and restore one, and the outcome of the last action above them.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	listed, err := snapshots.List(sandbox, input.Prefix)
	if err != nil {
		return err
	}
	return backofficerender.RenderBackupSnapshotsPage(sandbox, response, api.StatusOK, props.User, listed, snapshots.Busy(sandbox), input.Prefix, input.Notice)
}
