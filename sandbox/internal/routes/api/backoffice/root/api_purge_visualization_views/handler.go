package api_purge_visualization_views

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /api/admin/root/purge-visualization-views, the JSON twin
// of the Delete views form: {before: "YYYY-MM-DD"} removes every view before
// that local day, {all: true} every view, and {removed} says how many went. A
// body naming neither, or a date that does not read, is answered 400. Only a
// root reaches it: the backoffice-api-root-guard answers anyone else under
// /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	removed, field, message, err := visualization.Purge(sandbox, input.Body.Before, input.Body.All)
	if err != nil {
		return err
	}
	if message != "" {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, field, message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.VisualizationPurgeJSON(sandbox, removed))
}
