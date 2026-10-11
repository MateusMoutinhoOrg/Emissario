package config

import (
	api "github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	config "github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/config"
)

// Constructor fills Sandbox.Config, building it with the
// NewConfig of sandbox/internal/config, then starts the project's own part
// at its defaults: the git sources live under api.DefaultFontsDir until a
// command line's --fonts says otherwise. sandbox/generated.new.go calls it
// once, along with the Constructor of every other package under
// sandbox/constructors/.
//
// Written once by `agnos build` and then yours: wrap the
// implementation, decorate the contract, or build a different one entirely.
// No build rewrites this file once it is there.
func Constructor(sandbox *api.Sandbox) {
	sandbox.Config = config.NewConfig(sandbox)
	sandbox.Config.FontsDir = api.DefaultFontsDir
}
