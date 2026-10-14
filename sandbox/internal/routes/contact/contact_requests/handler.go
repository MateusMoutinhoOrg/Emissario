package contact_requests

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers POST /api/contact-requests, the contact form of the website:
// the body — any JSON value, at most the largest body of the contact
// settings — is kept exactly as sent, with the time, the ip and its country,
// and its id answered 201. It is refused 503 while the route is off or a
// restore runs, 403 from another site's page, 429 past the limits of the ip,
// 413 for a body too large and 400 for one that is not JSON.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	settings, refusal := contact.Admit(sandbox, props.ClientIp, input.Origin, input.Host)
	if refusal != nil {
		return contact.Fail(sandbox, response.SetHeader, refusal)
	}
	ref, refusal := contact.Receive(sandbox, settings, props.ClientIp, input.Body)
	if refusal != nil {
		return contact.Fail(sandbox, response.SetHeader, refusal)
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusCreated, contact.CreatedJSON(sandbox, ref))
}
