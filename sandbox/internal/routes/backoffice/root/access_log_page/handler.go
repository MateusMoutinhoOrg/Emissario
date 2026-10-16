package access_log_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/root/access-log to a reader of the access log:
// one page of the rows the filter keeps, newest first. Every list writes one
// row of the inventory while it is on. Anyone else is answered 403.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if refusal := privacyadmin.CanReadAccessLog(sandbox, props.User, nil); refusal != nil {
		if refusal.Status != api.StatusForbidden {
			return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
		}
		return backofficerender.RenderForbiddenMessage(sandbox, response, props.User, "Access log readers only", refusal.Message)
	}
	started := privacy.NowMilliseconds(sandbox)
	raw := privacyadmin.RawAccessFilter{Ip: input.Ip, From: input.From, To: input.To, Kind: input.Kind}
	notice := backofficerender.AccessLogNotice(sandbox, input.Notice, input.Count)
	filter, refusal := privacyadmin.ParseAccessFilter(sandbox, raw)
	if refusal != nil {
		return backofficerender.RenderAccessLogPage(sandbox, response, refusal.Status, props.User, raw, filter, false, nil, visualization.Paging{Limit: input.Limit}, notice, refusal.Message, nil, nil)
	}
	rows, err := privacyadmin.SelectAccess(sandbox, filter)
	if err != nil {
		return err
	}
	paging, from, to := visualization.Paginate(sandbox, len(rows), input.Page, input.Limit)
	shown := rows[from:to]
	if err := privacyadmin.Inventory(sandbox, props.User, privacyadmin.ActionAccessLogList, privacyadmin.FilterJSON(sandbox, filter), len(shown), started); err != nil {
		return err
	}
	return backofficerender.RenderAccessLogPage(sandbox, response, api.StatusOK, props.User, raw, filter, true, shown, paging, notice, "", nil, nil)
}
