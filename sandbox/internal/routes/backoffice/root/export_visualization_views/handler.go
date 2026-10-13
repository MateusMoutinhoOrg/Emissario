package export_visualization_views

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/root/export-visualization-views with the views
// the filters of the query string keep, as text/csv downloaded as
// visualization-views-<from>-<to>.csv: newest first, 100 000 at most, every ip
// whole. A filter that does not read is answered 400. Only a root reaches it:
// the backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	filter, refusal := visualization.ParseFilter(sandbox, visualization.RawFilter{
		From:    input.From,
		To:      input.To,
		Path:    input.Path,
		Prefix:  input.Prefix,
		Ip:      input.Ip,
		Country: input.Country,
		Device:  input.Device,
		Browser: input.Browser,
		Os:      input.Os,
		Status:  input.Status,
		Bots:    input.Bots,
	}, true)
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	selection, err := visualization.Select(sandbox, filter)
	if err != nil {
		return err
	}
	response.SetHeader("Content-Type", "text/csv; charset=utf-8")
	response.SetHeader("Content-Disposition", "attachment; filename=\""+visualization.ExportName(sandbox, filter)+"\"")
	response.SetHeader("Cache-Control", "no-store")
	response.SetStatus(api.StatusOK)
	return response.Write(visualization.ExportCsv(sandbox, selection))
}
