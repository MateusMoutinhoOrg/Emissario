package server

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	server "github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server"
)

// Constructor fills Sandbox.Server, building it with the
// NewServer of sandbox/internal/server. sandbox/generated.new.go calls it
// once, along with the Constructor of every other package under
// sandbox/constructors/.
//
// Written once by `agnos build` and then yours: wrap the
// implementation, decorate the contract, or build a different one entirely.
// No build rewrites this file once it is there.
func Constructor(sandbox *api.Sandbox) {
	sandbox.Server = server.NewServer(sandbox)
}
