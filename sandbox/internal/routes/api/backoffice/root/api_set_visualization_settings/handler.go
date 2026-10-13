package api_set_visualization_settings

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /api/admin/root/set-visualization-settings, the JSON
// twin of the settings form: tracking on or off, the retention in days — 0
// keeps the views forever — and the ignored ips, each an ip or a CIDR range.
// A value refused is answered 400 naming its field; a save answers {settings}
// as they were stored. Only a root reaches it: the backoffice-api-root-guard
// answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	settings, field, message := visualization.CheckSettings(sandbox, visualization.Settings{
		Enabled:       input.Body.Enabled,
		RetentionDays: int64(input.Body.RetentionDays),
		IgnoredIps:    input.Body.IgnoredIps,
	})
	if message != "" {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusBadRequest, field, message)
	}
	if err := visualization.SaveSettings(sandbox, settings); err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.VisualizationSettingsResponseJSON(sandbox, settings))
}
