package api_remove_contact_requests

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/remove-contact-requests, the JSON twin
// of the Delete buttons: the contact requests named, or every one before a
// date, and how many it removed. Only a root reaches it: the
// backoffice-api-root-guard answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	removed, refusal, err := contact.Remove(sandbox, input.Body.Ids, input.Body.Before)
	if err != nil {
		return err
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.RemovedJSON(sandbox, removed))
}
