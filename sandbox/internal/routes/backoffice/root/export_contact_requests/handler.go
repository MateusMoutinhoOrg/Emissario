package export_contact_requests

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers GET /admin/root/export-contact-requests: the contact requests
// of ids, or every one the filter keeps, downloaded as JSON — what an access
// or a portability request asks for.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	filter, refusal := contact.ParseFilter(sandbox, contact.RawFilter{From: input.From, To: input.To, Country: input.Country, Contains: input.Contains, Status: input.Status})
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	requests, refusal, err := contact.Export(sandbox, filter, input.Ids)
	if err != nil {
		return err
	}
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	response.SetHeader("Content-Disposition", "attachment; filename=\"contact-requests.json\"")
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.ContactRequestsExportJSON(sandbox, requests))
}
