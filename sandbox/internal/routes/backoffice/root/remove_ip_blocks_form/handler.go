package remove_ip_blocks_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/blocker"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/remove-ip-blocks, the Remove button of the
// blocks: the blocks selected are removed and the browser sent back to the
// blocks. None found answers the page under a 404. Only a root reaches it:
// the backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	ids := []int64{}
	for _, text := range input.Body.Ids {
		id, err := sandbox.Deps.StringsDeps.ParseInt(sandbox.Deps.StringsDeps.TrimSpace(text), 10, 64)
		if err == nil {
			ids = append(ids, id)
		}
	}
	removed, refusal, err := blocker.Remove(sandbox, ids)
	if err != nil {
		return err
	}
	if refusal != nil {
		return backofficerender.RenderIpBlocksPage(sandbox, response, refusal.Status, props.User, props.ClientIp, backofficerender.IpBlockForm{}, "", 0, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, blocker.BlocksPath+"?notice="+blocker.NoticeRemoved+"&count="+sandbox.Deps.StringsDeps.FormatInt(int64(removed), 10))
}
