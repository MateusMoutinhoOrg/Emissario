package release_privacy_hold_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers POST /admin/root/release-privacy-hold: the hold of id is
// released and deleted, its rows following their retention again. It
// redirects to the Privacy page with what happened.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	found, refusal, err := privacyadmin.ReleaseHold(sandbox, props.User, nil, int64(input.Body.Id))
	if err != nil {
		return err
	}
	notice := "released"
	if !found {
		notice = "hold-not-found"
	} else if refusal != nil {
		notice = "hold-refused"
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, privacy.SettingsPath+"?notice="+notice)
}
