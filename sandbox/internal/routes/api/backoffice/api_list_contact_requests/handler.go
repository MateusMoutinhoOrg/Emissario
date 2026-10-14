package api_list_contact_requests

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/list-contact-requests, the JSON twin of the
// contact list, open to every backoffice user: one page of the contact
// requests the filters keep, newest first. Filters that do not read are
// answered 400.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	filter, refusal := contact.ParseFilter(sandbox, contact.RawFilter{
		From:     input.Body.From,
		To:       input.Body.To,
		Country:  input.Body.Country,
		Contains: input.Body.Contains,
		Status:   input.Body.Status,
	})
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	requests, paging, err := contact.List(sandbox, filter, input.Body.Page, input.Body.Limit)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.ContactRequestsJSON(sandbox, requests, paging))
}
