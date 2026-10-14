package remove_contact_requests_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/remove-contact-requests, the Delete buttons
// of the contact requests: the ones selected, or every one before a date. A
// deletion sends the browser back to the page it came from — the list, when
// it was one of the requests deleted — with how many it removed; a refusal
// answers the page it came from under its status. Only a root reaches it: the
// backoffice-root-guard answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	removed, refusal, err := contact.Remove(sandbox, input.Body.Ids, input.Body.Before)
	if err != nil {
		return err
	}
	if refusal != nil {
		if sandbox.Deps.StringsDeps.HasPrefix(input.Body.Back, contact.SettingsPath) {
			settings, err := contact.ReadSettings(sandbox)
			if err != nil {
				return err
			}
			return backofficerender.RenderContactSettingsPage(sandbox, response, refusal.Status, props.User, backofficerender.ContactSettingsFormOf(sandbox, settings), "", 0, refusal.Message)
		}
		return backofficerender.RenderContactRequestsRefusal(sandbox, response, refusal.Status, props.User, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, contact.BackTo(sandbox, input.Body.Back, true, contact.NoticeRemoved, removed))
}
