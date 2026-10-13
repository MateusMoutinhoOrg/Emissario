package set_visualization_settings_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/set-visualization-settings: tracking on
// (enabled=1) or off, the retention — blank for forever — and the ignored ips,
// one per line, are saved and the browser is sent back to the page with
// ?notice=saved. A refused value answers the form again, filled with what was
// sent, under a 400 with the reason above it. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	form := backofficerender.VisualizationSettingsForm{
		Enabled:    input.Body.Enabled == "1" || input.Body.Enabled == "on",
		Retention:  input.Body.RetentionDays,
		IgnoredIps: input.Body.IgnoredIps,
	}
	retention, message := visualization.ParseRetention(sandbox, input.Body.RetentionDays)
	if message != "" {
		return backofficerender.RenderVisualizationSettingsPage(sandbox, response, api.StatusBadRequest, props.User, form, "", 0, message, props.ClientIp)
	}
	settings, _, message := visualization.CheckSettings(sandbox, visualization.Settings{
		Enabled:       form.Enabled,
		RetentionDays: retention,
		IgnoredIps:    visualization.SplitIgnoredIps(sandbox, input.Body.IgnoredIps),
	})
	if message != "" {
		return backofficerender.RenderVisualizationSettingsPage(sandbox, response, api.StatusBadRequest, props.User, form, "", 0, message, props.ClientIp)
	}
	if err := visualization.SaveSettings(sandbox, settings); err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, visualization.SettingsPath+"?notice="+visualization.NoticeSaved)
}
