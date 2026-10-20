package rotate_encrypt_key

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/commandprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/encryption"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
)

// Handle answers `rotate-encrypt-key`: it moves every database under
// --database, and every file every snapshot keeps, from the key in the
// variable --old-key-env names to the one in EMISSARIO_ENCRYPT_KEY, the path
// after a key leaked. Both variables are removed from the environment once
// read. It runs in three steps, and encryption.json records how far it got:
//
//  1. a snapshot of every database, opened under the old key, and then
//     rotating_to, naming the new key, in encryption.json;
//  2. every file of every database resealed under the new key, one by one,
//     and every index entry — named by a digest of the old key — removed;
//  3. under the new key, every database reindexed and every snapshot
//     resealed, and only then encryption.json naming the new key alone.
//
// One interrupted is run again with both keys: it skips step 1 and reseals
// only what is left. The database-encryption middleware skips this command,
// which reads both keys itself, and refuses every other one while
// rotating_to is set. Run it with the server stopped.
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
	if info.RotatingTo == "" && info.KeyId == newKeyId {
		response.Printf("the databases under %s are already encrypted with the new key (key_id %s): nothing to do\n", sandbox.Config.DatabaseDir, newKeyId)
		return nil
	}
	if info.KeyId != oldKeyId {
		return cli.Fail(api.ExitFailure, "old-key-env", "the key in "+input.OldKeyEnv+" does not match the databases under "+sandbox.Config.DatabaseDir+" (key_id "+info.KeyId+")")
	}
	if info.RotatingTo != "" && info.RotatingTo != newKeyId {
		return cli.Fail(api.ExitFailure, "", "a rotation of the databases under "+sandbox.Config.DatabaseDir+" to key_id "+info.RotatingTo+" did not finish: run it again with that key in "+newEnv)
	}
	again := "; run it again with both keys, it reseals only what is left"
	sandbox.Config.Encrypted = true

	safety := ""
	if info.RotatingTo == "" {
		if err := enc.EncryptOld(); err != nil {
			return cli.Fail(api.ExitFailure, "", err.Error())
		}
		sandbox.Config.EncryptionKeyId = oldKeyId
		snapshot, err := snapshots.PreRotation(sandbox)
		if err != nil {
			return cli.Fail(api.ExitFailure, "", "the snapshot taken before the rotation failed, nothing was resealed: "+err.Error())
		}
		safety = snapshot.Name
		if err := encryption.WriteInfo(sandbox, oldKeyId, newKeyId); err != nil {
			return cli.Fail(api.ExitFailure, "", encryption.FileName+" could not be written, nothing was resealed: "+err.Error())
		}
	}

	files, dropped, err := enc.RotateFiles(snapshots.DataDir(sandbox), func(path string) bool {
		return encryption.IndexEntry(sandbox, path)
	})
	if err != nil {
		return cli.Fail(api.ExitFailure, "", "the rotation stopped: "+err.Error()+again)
	}

	if err := enc.Encrypt(); err != nil {
		return cli.Fail(api.ExitFailure, "", err.Error())
	}
	sandbox.Config.EncryptionKeyId = newKeyId
	if err := encryption.Reindex(sandbox); err != nil {
		return cli.Fail(api.ExitFailure, "", "the rotation stopped: "+err.Error()+again)
	}
	result, err := snapshots.Reseal(sandbox, newKeyId)
	if err != nil {
		return cli.Fail(api.ExitFailure, "", "the rotation stopped: "+err.Error()+again)
	}
	if err := encryption.WriteInfo(sandbox, newKeyId, ""); err != nil {
		return cli.Fail(api.ExitFailure, "", "every file is resealed, but "+encryption.FileName+" could not be written: "+err.Error()+again)
	}

	response.Printf("rotated %s from key_id %s to %s: %d database files resealed, %d index entries removed and every index rebuilt under the new key; %d files of %d snapshots resealed and %d index entries left out of them\n", sandbox.Config.DatabaseDir, oldKeyId, newKeyId, files, dropped, result.Contents, result.Snapshots, result.Dropped)
	if safety != "" {
		response.Printf("snapshot %s holds what the databases held before\n", safety)
	}
	response.Printf("archives downloaded before the rotation stay sealed under the old key: keep it to read them, or download them again\n")
	return nil
}
