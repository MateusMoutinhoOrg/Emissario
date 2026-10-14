package archive_contact_requests_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/archive-contact-requests, the Archive and
// Move to inbox buttons: the contact requests selected are archived, or
// brought back to the inbox, and the browser sent back to the page it came
// from with how many changed. A refusal answers the inbox under its status.
// Only a root reaches it: the backoffice-root-guard answers anyone else under
// /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	if input.Body.Archived != "1" && input.Body.Archived != "0" {
		return backofficerender.RenderContactRequestsRefusal(sandbox, response, api.StatusBadRequest, props.User, "archived is 1 to archive, or 0 to move back to the inbox.")
	}
	archived := input.Body.Archived == "1"
	changed, refusal, err := contact.SetArchived(sandbox, input.Body.Ids, archived)
	if err != nil {
		return err
	}
	if refusal != nil {
		return backofficerender.RenderContactRequestsRefusal(sandbox, response, refusal.Status, props.User, refusal.Message)
	}
	notice := contact.NoticeArchived
	if !archived {
		notice = contact.NoticeUnarchived
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, contact.BackTo(sandbox, input.Body.Back, false, notice, changed))
}
