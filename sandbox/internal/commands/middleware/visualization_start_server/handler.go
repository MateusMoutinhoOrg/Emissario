package visualization_start_server

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle runs in front of `start-server` and answers nothing, so
// start-server runs after it. It points sandbox.Config.VisualizationGeoipFile
// at the country database --visualization-geoip names — empty for the DB-IP
// Lite one embedded in the binary — and opens it: a file that does not open
// stops the server here, before it listens. It then starts the goroutine that
// writes the views the front route queues; it lives as long as the process:
// when start-server stops, it stops. The views older than the retention are
// removed by the retention cleanup privacy-start-server starts.
//
// It is a middleware rather than an edit to start-server's own handler, the
// same way website-source-start-server is, so start-server stays as it was.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	sandbox.Config.VisualizationGeoipFile = input.VisualizationGeoip
	if err := visualization.OpenGeoip(sandbox); err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", err.Error())
	}
	visualization.Start(sandbox)
	return nil
}
