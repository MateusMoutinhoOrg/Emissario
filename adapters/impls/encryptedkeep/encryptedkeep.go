package encryptedkeep

import (
	databasedeps "github.com/MateusMoutinhoOrg/Emissario/adapters/impls/databasedeps"
	xchachaencryption "github.com/MateusMoutinhoOrg/Emissario/adapters/impls/xchachaencryption"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// Bind fills deps.Deps.DatabaseDeps with Keep built over its own file storage
// wrapped by xchachaencryption.Wrap: every value of every database is sealed
// on its way to the disk and opened on its way back, as the
// database-encryption middleware arms it through deps.Deps.EncryptionDeps.
// Until it does, every read and write fails. It stands in the binding for
// the generated databasedeps shim, which stays installed and is what
// `agnos set-dep` keeps up to date.
func Bind(deps *deps.Deps) {
	databasedeps.BindOver(deps, xchachaencryption.Wrap)
}
