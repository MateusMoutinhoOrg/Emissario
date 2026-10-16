package get_contact_request_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/get-contact-request?id=, open to every backoffice
// user: one contact request whole, its JSON laid out as it was sent. An id no
// contact request holds answers the inbox under a 404.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}

	request, ok := contact.Find(sandbox, input.Id)
	if !ok {
		return backofficerender.RenderContactRequestsRefusal(sandbox, response, api.StatusNotFound, props.User, "No contact request has this id.")
	}
	return backofficerender.RenderContactRequestPage(sandbox, response, props.User, request, input.Notice, input.Count, input.Reason)
}
