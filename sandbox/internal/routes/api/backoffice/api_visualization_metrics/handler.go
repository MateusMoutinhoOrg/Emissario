package api_visualization_metrics

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /api/admin/visualization-metrics, open to every
// backoffice user, with the metrics of the views the filters of the body keep:
// {summary, timeline, top}. Every field of the body is optional: the period is
// the last 7 days, the bucket the one the period calls for, top 10. A filter
// that does not read is answered 400, an ip sent by a viewer 403; the ips of
// top.visitors are masked to a viewer.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	root := backofficeauth.Role(props.User.Role) == backofficeauth.RoleRoot
	filter, refusal := visualization.ParseFilter(sandbox, visualization.RawFilter{
		From:    input.Body.From,
		To:      input.Body.To,
		Path:    input.Body.Path,
		Prefix:  input.Body.Prefix,
		Ip:      input.Body.Ip,
		Country: input.Body.Country,
		Device:  input.Body.Device,
		Browser: input.Body.Browser,
		Os:      input.Body.Os,
		Status:  visualization.StatusParam(sandbox, input.Body.Status),
		Bots:    visualization.BotsParam(sandbox, input.Body.Bots),
	}, root)
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	selection, err := visualization.Select(sandbox, filter)
	if err != nil {
		return err
	}
	bucket, refusal := visualization.ChooseBucket(sandbox, filter, input.Body.Bucket)
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	top := input.Body.Top
	if top < 1 {
		top = visualization.DefaultTop
	}
	metrics := visualization.Measure(sandbox, selection, bucket, min(top, visualization.MaxTop), root)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.VisualizationMetricsJSON(sandbox, metrics))
}
