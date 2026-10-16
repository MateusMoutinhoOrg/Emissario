package databasedeps

import (
	keepbinding "github.com/MateusMoutinhoOrg/Keep/adapters/bindings/standard"
	keeplib "github.com/MateusMoutinhoOrg/Keep/sandbox"
	storagedeps "github.com/MateusMoutinhoOrg/Keep/sandbox/deps/storagedeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// BindOver fills deps.DatabaseDeps as Bind does, with one change: the storage
// Keep writes every database through is wrap applied to the one Keep's own
// binding builds. It is how a project changes what Keep is built over without
// a release of Keep and without a second copy of the converters: this file is
// hand-written and lives beside generated.databasedeps.go, in its package, so
// it reuses convSandbox, which `agnos set-dep databasedeps` regenerates for
// every version of Keep while it leaves this file alone.
func BindOver(deps *deps.Deps, wrap func(storagedeps.Contract) storagedeps.Contract) {
	remote_deps := keepbinding.New()
	remote_deps.StorageDeps = wrap(remote_deps.StorageDeps)
	remote_sandbox := keeplib.New(&remote_deps)
	deps.DatabaseDeps = convSandbox(*remote_sandbox)
}
