package set_access_log_readers_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers POST /admin/root/set-access-log-readers: the roots of ids
// read the access log from now on, and no one else. It redirects to the
// Privacy page with what happened.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	ids := []int64{}
	for _, id := range input.Body.Ids {
		ids = append(ids, int64(id))
	}
	notice := "readers-saved"
	if refusal := privacyadmin.SetReaders(sandbox, ids); refusal != nil {
		notice = "readers-refused"
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, privacy.SettingsPath+"?notice="+notice)
}
