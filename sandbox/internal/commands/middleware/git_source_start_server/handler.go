package git_source_start_server

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/gitsource"
)

// Handle runs in front of `start-server` and answers nothing, so
// start-server runs after it. It points sandbox.Config.FontsDir at the folder
// --fonts names — the git sources are cloned there, the frontend one in
// <fonts>/frontend — and starts the goroutine that keeps that checkout at the
// head of the branch a root set in the backoffice, every sync time. The
// goroutine lives as long as the process: when start-server stops, it stops.
// With no sync time set it does nothing.
//
// It is a middleware rather than an edit to start-server's own handler, the
// same way backoffice-start-server is, so start-server stays as it was.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	if input.Fonts != "" {
		sandbox.Config.FontsDir = input.Fonts
	}
	gitsource.Start(sandbox)
	return nil
}
