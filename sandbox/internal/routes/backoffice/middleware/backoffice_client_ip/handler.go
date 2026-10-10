package backoffice_client_ip

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
)

// Handle runs in front of every ANY /*, on the lowest rung of the
// chain: it is a middleware. It puts the ip the request came from on
// props.ClientIp — the connection's own, or, when start-server trusts
// X-Forwarded-For, the one the reverse proxy in front appended — and declines,
// so every route after it reads one ip, worked out once.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	props.ClientIp = backofficehttp.ClientIp(sandbox, input.XClientIp, input.XForwardedFor)
	return nil
}
