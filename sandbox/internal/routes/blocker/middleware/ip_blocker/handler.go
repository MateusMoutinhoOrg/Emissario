package ip_blocker

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/blocker"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle runs in front of every ANY /*, after the backoffice authentication:
// it is a middleware. A request from an ip a root blocked — the ip, a range
// holding it, its hash or its country — is answered 403 {"error":
// "blocked"}, before any route reads it; any other declines. A request a
// backoffice session or API token authenticated is never blocked: the
// session-auth middleware already sent an /admin request without one to the
// sign-in, which is blocked.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User != nil {
		return nil
	}
	if _, blocked := blocker.Blocked(sandbox, props.ClientIp); blocked {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusForbidden, "", blocker.CodeBlocked)
	}
	return nil
}
