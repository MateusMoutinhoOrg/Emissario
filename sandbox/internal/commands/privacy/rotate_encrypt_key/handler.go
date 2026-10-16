package rotate_encrypt_key

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/encryption"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// Handle answers `rotate-encrypt-key`: it reseals every database under
// --database, and every file every snapshot keeps, from the key in the
// variable --old-key-env names to the one in EMISSARIO_ENCRYPT_KEY, the path
// after a key leaked. Both variables are removed from the environment once
// read. It takes a snapshot first, works file by file, and writes
// encryption.json last, so one interrupted is run again and reseals only
// what is left. The database-encryption middleware skips this command: it
// reads both keys itself. Run it with the server stopped.
func Handle(sandbox *api.Sandbox, props *commandprops.CommandProps, input *Input, response *api.CommandResponse) error {
	cli := sandbox.Deps.OpinionatedAgnosCli
	enc := sandbox.Deps.EncryptionDeps
	newEnv := encryption.KeyEnv(sandbox)

	oldKeyId, oldPresent, oldErr := enc.LoadOldKey(input.OldKeyEnv)
	newKeyId, newPresent, newErr := enc.LoadKey(newEnv)
	if oldErr != nil {
		return cli.Fail(api.ExitFailure, "old-key-env", oldErr.Error())
	}
	if newErr != nil {
		return cli.Fail(api.ExitFailure, "", newErr.Error())
	}
	if !oldPresent {
		return cli.Fail(api.ExitFailure, "old-key-env", input.OldKeyEnv+" is not set: set it to the key the databases are encrypted with now")
	}
	if !newPresent {
		return cli.Fail(api.ExitFailure, "", newEnv+" is not set: set it to the new key (`"+sandbox.Config.ProjectName+" generate-encrypt-key`)")
	}
	if oldKeyId == newKeyId {
		return cli.Fail(api.ExitFailure, "", "the old key and the new one are the same key")
	}

	folder, info, err := encryption.Folder(sandbox)
	if err != nil {
		return cli.Fail(api.ExitFailure, "", err.Error())
	}
	if folder != encryption.FolderEncrypted {
		return cli.Fail(api.ExitFailure, "", "the databases under "+sandbox.Config.DatabaseDir+" are not encrypted: there is no key to rotate")
	}
	if info.KeyId == newKeyId {
		response.Printf("the databases under %s are already encrypted with the new key (key_id %s): nothing to do\n", sandbox.Config.DatabaseDir, newKeyId)
		return nil
	}
	if info.KeyId != oldKeyId {
		return cli.Fail(api.ExitFailure, "old-key-env", "the key in "+input.OldKeyEnv+" does not match the databases under "+sandbox.Config.DatabaseDir+" (key_id "+info.KeyId+")")
	}

	if err := enc.Rotate(encryption.Skip(sandbox)); err != nil {
		return cli.Fail(api.ExitFailure, "", err.Error())
	}
	sandbox.Config.Encrypted = true
	sandbox.Config.EncryptionKeyId = oldKeyId
	result, err := snapshots.Rotate(sandbox, newKeyId)
	if err != nil {
		return cli.Fail(api.ExitFailure, "", "the rotation stopped: "+err.Error()+"; run it again with both keys, it reseals only what is left")
	}
	if err := encryption.WriteInfo(sandbox, newKeyId); err != nil {
		return cli.Fail(api.ExitFailure, "", "every file is resealed, but "+encryption.FileName+" could not be written: "+err.Error()+"; run it again with both keys")
	}
	sandbox.Config.EncryptionKeyId = newKeyId

	response.Printf("rotated %s from key_id %s to %s: %d database files and %d files of %d snapshots resealed; snapshot %s holds what it held before\n", sandbox.Config.DatabaseDir, oldKeyId, newKeyId, result.Files, result.Contents, result.Snapshots, result.Snapshot)
	response.Printf("archives downloaded before the rotation stay sealed under the old key: keep it to read them, or download them again\n")
	return nil
}
