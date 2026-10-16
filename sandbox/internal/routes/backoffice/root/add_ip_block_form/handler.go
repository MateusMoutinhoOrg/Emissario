package add_ip_block_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/blocker"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/add-ip-block: the block is checked and
// added, and the browser sent back to the blocks. A refused value answers the
// page again, the form filled with what was sent, under its status with the
// reason above it. Only a root reaches it: the backoffice-root-guard answers
// anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	sent := backofficerender.IpBlockForm{Value: input.Body.Value, Reason: input.Body.Reason, Days: input.Body.Days}
	days, refusal := blocker.ParseDays(sandbox, input.Body.Days)
	if refusal == nil {
		var err error
		_, refusal, err = blocker.Add(sandbox, blocker.Input{Value: input.Body.Value, Reason: input.Body.Reason, Days: days}, props.ClientIp)
		if err != nil {
			return err
		}
	}
	if refusal != nil {
		return backofficerender.RenderIpBlocksPage(sandbox, response, refusal.Status, props.User, props.ClientIp, sent, "", 0, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, blocker.BlocksPath+"?notice="+blocker.NoticeAdded)
}
