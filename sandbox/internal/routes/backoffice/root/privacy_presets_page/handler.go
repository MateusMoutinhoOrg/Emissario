package privacy_presets_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/privacy-presets: the preset applied, how
// this server drifts from it, its controller checklist, every preset with
// Preview, Export and Delete, and the forms that save the current settings
// as a preset and import one. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	return backofficerender.RenderPrivacyPresetsPage(sandbox, response, api.StatusOK, props.User, input.Notice, input.Preset, input.Reason)
}
