package api_list_access_log

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /api/admin/root/list-access-log to a reader of the
// access log, with a token made after two-factor sign-in while it is
// required: one page of the rows the filter keeps, newest first. Every list
// writes one row of the inventory while it is on.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if refusal := privacyadmin.CanReadAccessLog(sandbox, props.User, props.ApiToken); refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	started := privacy.NowMilliseconds(sandbox)
	filter, refusal := privacyadmin.ParseAccessFilter(sandbox, privacyadmin.RawAccessFilter{Ip: input.Body.Ip, From: input.Body.From, To: input.Body.To, Kind: input.Body.Kind})
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	rows, err := privacyadmin.SelectAccess(sandbox, filter)
	if err != nil {
		return err
	}
	paging, from, to := visualization.Paginate(sandbox, len(rows), input.Body.Page, input.Body.Limit)
	if err := privacyadmin.Inventory(sandbox, props.User, privacyadmin.ActionAccessLogList, privacyadmin.FilterJSON(sandbox, filter), to-from, started); err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.AccessListJSON(sandbox, rows[from:to], paging))
}
