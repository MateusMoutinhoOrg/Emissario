package database_encryption

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/encryption"
)

// Handle runs in front of every command line that may open a database —
// every one but help, version, generate-encrypt-key, rotate-encrypt-key,
// drop-front-skills and drop-agents-skills — right after database-dir has
// set the folder, and answers nothing. It reads the key out of the
// environment, removes it from there, checks it against the folder and arms
// the databases: Keep opens every one under the EncryptKey derived from the
// key, or in clear with --notencrypt (encryption.Open); until then, every
// database refuses to open. A command line that cannot run — no key, another
// key, a folder in clear, a rotation that did not finish — ends here with
// exit 1. An empty command line opens nothing.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	if input.Verb == "" {
		return nil
	}
	if err := encryption.Open(sandbox, input.Notencrypt); err != nil {
		return sandbox.Deps.OpinionatedAgnosCli.Fail(api.ExitFailure, "", err.Error())
	}
	if input.Notencrypt {
		sandbox.Deps.StdDeps.Eprintf("%s", encryption.Warning(sandbox))
	}
	return nil
}
