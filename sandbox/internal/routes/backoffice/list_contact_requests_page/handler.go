package list_contact_requests_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers GET /admin/list-contact-requests, open to every backoffice
// user: one page of the contact requests the filters keep — the inbox, the
// archived or all, by period, country and a text of the JSON — newest first.
// Filters that do not read answer the page under a 400 with the reason.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	raw := contact.RawFilter{From: input.From, To: input.To, Country: input.Country, Contains: input.Contains, Status: input.Status}
	filter, refusal := contact.ParseFilter(sandbox, raw)
	if refusal != nil {
		return backofficerender.RenderContactRequestsPage(sandbox, response, refusal.Status, props.User, raw, filter, refusal.Message, nil, visualization.Paging{}, "", 0)
	}
	requests, paging, err := contact.List(sandbox, filter, input.Page, input.Limit)
	if err != nil {
		return err
	}
	return backofficerender.RenderContactRequestsPage(sandbox, response, api.StatusOK, props.User, raw, filter, "", requests, paging, input.Notice, input.Count)
}
