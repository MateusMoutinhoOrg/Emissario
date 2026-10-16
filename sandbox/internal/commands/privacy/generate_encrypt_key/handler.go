package generate_encrypt_key

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/encryption"
)

// Handle answers `generate-encrypt-key`: a fresh key, 32 random bytes in
// base64, on stdout alone — so `EMISSARIO_ENCRYPT_KEY=$(emissario
// generate-encrypt-key)` captures it — and on stderr the warning that losing
// it loses every database and every backup. It opens no database.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	key, err := sandbox.Deps.EncryptionDeps.GenerateKey()
	if err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", "failed to generate a key: "+err.Error())
	}
	response.Printf("%s\n", key)
	sandbox.Deps.StdDeps.Eprintf("warning: keep this key safe, outside the server and its backups (a secret manager): losing it loses every database and every backup encrypted with it. Set it as %s for every command that opens a database.\n", encryption.KeyEnv(sandbox))
	return nil
}
