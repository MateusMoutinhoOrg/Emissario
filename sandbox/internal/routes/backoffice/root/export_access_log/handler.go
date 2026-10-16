package export_access_log

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers GET /admin/root/export-access-log to a reader of the access
// log: every row the filter keeps, at most privacyadmin.MaxAccessExportRows,
// downloaded as JSON or CSV. Every export writes one row of the inventory
// while it is on.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if refusal := privacyadmin.CanReadAccessLog(sandbox, props.User, nil); refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	started := privacy.NowMilliseconds(sandbox)
	format := input.Format
	if format == "" {
		format = privacyadmin.FormatJson
	}
	if format != privacyadmin.FormatJson && format != privacyadmin.FormatCsv {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, "format", "format is json or csv")
	}
	filter, refusal := privacyadmin.ParseAccessFilter(sandbox, privacyadmin.RawAccessFilter{Ip: input.Ip, From: input.From, To: input.To, Kind: input.Kind})
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
	name := "access-log-" + filter.From + "-" + filter.To
	if format == privacyadmin.FormatCsv {
		response.SetHeader("Content-Type", "text/csv; charset=utf-8")
		response.SetHeader("Content-Disposition", "attachment; filename=\""+name+".csv\"")
		response.SetHeader("Cache-Control", "no-store")
		response.SetStatus(api.StatusOK)
		return response.Write([]byte(privacyadmin.AccessCsv(sandbox, rows)))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("rows", backofficeapi.AccessRowsJSON(sandbox, rows))
	response.SetHeader("Content-Disposition", "attachment; filename=\""+name+".json\"")
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, document)
}
