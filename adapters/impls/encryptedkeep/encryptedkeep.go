package encryptedkeep

import (
	databasedeps "github.com/MateusMoutinhoOrg/Emissario/adapters/impls/databasedeps"
	keepencryption "github.com/MateusMoutinhoOrg/Emissario/adapters/impls/keepencryption"
	contract "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"

	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps"
)

// Bind fills deps.Deps.DatabaseDeps with Keep as the generated databasedeps
// shim builds it, and sets the EncryptKey of every Props it opens to the one
// keepencryption armed: Keep itself then encrypts every value of every
// database on its way to the disk and decrypts it on its way back. Until the
// database-encryption middleware arms it, through deps.Deps.EncryptionDeps,
// every database refuses to open. The EncryptKey is blanked on the Props a
// database hands back, so the key never reaches the sandbox. It stands in the
// binding for the generated shim, which stays installed and is what
// `agnos set-dep` keeps up to date.
func Bind(deps *deps.Deps) {
	databasedeps.Bind(deps)
	open := deps.DatabaseDeps.Databases.New
	deps.DatabaseDeps.Databases.New = func(props contract.Props) (contract.Database, *contract.Error) {
		encryptKey, err := keepencryption.EncryptKey()
		if err != nil {
			return contract.Database{}, &contract.Error{Type: contract.Internal, Message: err.Error()}
		}
		props.EncryptKey = encryptKey
		database, failure := open(props)
		database.Props.EncryptKey = ""
		return database, failure
	}
}
