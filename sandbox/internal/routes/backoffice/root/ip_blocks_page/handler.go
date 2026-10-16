package ip_blocks_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/ip-blocks: every block, beside the form
// that adds one, filled with ?value= when a Block link sent it. Only a root
// reaches it: the backoffice-root-guard answers anyone else under
// /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	form := backofficerender.IpBlockForm{Value: input.Value}
	return backofficerender.RenderIpBlocksPage(sandbox, response, api.StatusOK, props.User, props.ClientIp, form, input.Notice, input.Count, "")
}
