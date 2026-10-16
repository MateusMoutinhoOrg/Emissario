package snapshots

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backup"
)

// RotationResult is what Rotate did.
type RotationResult struct {
	// Snapshot is the name of the snapshot taken before anything was
	// resealed.
	Snapshot string
	// Files is how many files of the databases were resealed.
	Files int
	// Contents is how many files kept by snapshots were resealed.
	Contents int
	// Snapshots is how many snapshots hold only files sealed under the new
	// key now.
	Snapshots int
}

// preRotationPrefix starts the name of the snapshot a rotation takes first.
const preRotationPrefix = "pre-rotation"

// Rotate reseals everything under DataDir from the old key to the new one of
// newKeyId, the store already armed for a rotation (EncryptionDeps.Rotate):
// a snapshot of what DataDir holds first, then every file of every database
// — the backup database's own included — and then every file every snapshot
// keeps, each under the path it is restored at, so every snapshot stays
// restorable with the new key; the blobs no snapshot holds anymore are
// removed. Run again after an interruption, it reseals only what is left.
// The caller writes encryption.json once it returns, and runs it with no
// server writing DataDir.
func Rotate(sandbox *api.Sandbox, newKeyId string) (RotationResult, error) {
	result := RotationResult{}
	if !acquire(sandbox) {
		return result, sandbox.Deps.StdDeps.Errorf("a backup job is running")
	}
	defer release(sandbox)
	db := backup.New(sandbox)

	safety, err := create(sandbox, db, preRotationPrefix)
	if err != nil {
		return result, sandbox.Deps.StdDeps.Errorf("the snapshot taken before the rotation failed, nothing was resealed: %s", err.Error())
	}
	result.Snapshot = safety.Name

	result.Files, err = sandbox.Deps.EncryptionDeps.RotateFiles(DataDir(sandbox))
	if err != nil {
		return result, err
	}

	listed, err := db.ListSnapshots(backup.SnapshotFilter{})
	if err != nil {
		return result, err
	}
	resealed := map[string]string{}
	for _, snapshot := range listed {
		contents, err := db.ListSnapshotContents(snapshot.Id)
		if err != nil {
			return result, err
		}
		for _, content := range contents {
			memo := content.Sha + "\x00" + content.Path
			sha, done := resealed[memo]
			if !done {
				blob, ok := db.FindBlobBySha(content.Sha)
				if !ok {
					return result, sandbox.Deps.StdDeps.Errorf("snapshot %s: the content of %s is missing", snapshot.Name, content.Path)
				}
				value, changed, err := sandbox.Deps.EncryptionDeps.Reseal(blob.Value, content.Path)
				if err != nil {
					return result, sandbox.Deps.StdDeps.Errorf("snapshot %s: %s", snapshot.Name, err.Error())
				}
				sha = content.Sha
				if changed {
					if sha, err = storeBlob(sandbox, db, value); err != nil {
						return result, err
					}
					result.Contents++
				}
				resealed[memo] = sha
			}
			if sha != content.Sha {
				if err := backup.SetSnapshotContentSha(sandbox, db, snapshot.Id, content.Id, sha); err != nil {
					return result, err
				}
			}
		}
		if snapshot.KeyId != newKeyId {
			if err := db.SetSnapshotKeyId(snapshot.Id, newKeyId); err != nil {
				return result, err
			}
		}
		result.Snapshots++
	}

	if _, _, err := optimize(sandbox, db); err != nil {
		return result, err
	}
	return result, nil
}
