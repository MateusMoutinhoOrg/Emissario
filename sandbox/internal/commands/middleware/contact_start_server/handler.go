package contact_start_server

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
)

// Handle runs in front of `start-server` and answers nothing, so start-server
// runs after it. It starts the writer of the contact requests, which turns
// what POST /api/contact-requests queues into records of contact-db (holding
// them while a backup is taken or restored). It lives as long as the process:
// when start-server stops, it stops.
//
// The country of each request is looked up in the database
// visualization-start-server opened, on the rung before.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	contact.Start(sandbox)
	return nil
}
