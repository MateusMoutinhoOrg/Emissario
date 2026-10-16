package api_export_access_log

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/export-access-log to a reader of the
// access log: every row the filter keeps, at most
// privacyadmin.MaxAccessExportRows, as {rows} or {csv}. Every export writes
// one row of the inventory while it is on.
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
	if len(rows) > privacyadmin.MaxAccessExportRows {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, "", "the filter keeps more than 100000 rows: narrow its period or its ip")
	}
	if err := privacyadmin.Inventory(sandbox, props.User, privacyadmin.ActionAccessLogExport, privacyadmin.FilterJSON(sandbox, filter), len(rows), started); err != nil {
		return err
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	if input.Body.Format == privacyadmin.FormatCsv {
		document.AddItemToObject("csv", privacyadmin.AccessCsv(sandbox, rows))
	} else {
		document.AddItemToObject("rows", backofficeapi.AccessRowsJSON(sandbox, rows))
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, document)
}
