package privacy_start_server

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// Handle runs in front of start-server, after every other start-server
// middleware, and answers nothing. It adds the privacy settings with a new IP
// key on a new database, so the key exists before the first view is hashed;
// it trusts the header --client-port-header names for the source port of the
// visitor's connection — only behind the reverse proxy --allow-x-forwarded-for
// trusts, under the same rule — and starts the writer of the access log.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	if err := privacy.EnsureSettings(sandbox); err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "the privacy settings could not be written: "+err.Error())
	}
	header := sandbox.Deps.StringsDeps.TrimSpace(input.ClientPortHeader)
	switch {
	case header != "" && !sandbox.Config.AllowXForwardedFor:
		sandbox.Deps.StdDeps.Eprintf("warning: --client-port-header is ignored without --allow-x-forwarded-for: a header only tells the source port when a reverse proxy the server trusts sets it; the access log keeps the port of the connection itself\n")
	case header != "":
		sandbox.Deps.ServerDeps.TrustPortHeader(header)
		sandbox.Config.ClientPortHeader = header
	}
	privacy.StartAccessLog(sandbox)
	return nil
}
