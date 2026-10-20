package snapshots

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backup"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/encryption"
)

// RotationResult is what Reseal did to the snapshots.
type RotationResult struct {
	// Contents is how many files kept by snapshots were resealed.
	Contents int
	// Dropped is how many index entries were left out of snapshots.
	Dropped int
	// Snapshots is how many snapshots hold only files sealed under the new
	// key now.
	Snapshots int
}

// preRotationPrefix starts the name of the snapshot a rotation takes first.
const preRotationPrefix = "pre-rotation"

// PreRotation takes the snapshot a rotation takes before it reseals
// anything: what DataDir holds, the databases still armed under the old key.
func PreRotation(sandbox *api.Sandbox) (backup.SnapshotRecord, error) {
	if !acquire(sandbox) {
		return backup.SnapshotRecord{}, sandbox.Deps.StdDeps.Errorf("a backup job is running")
	}
	defer release(sandbox)
	return create(sandbox, backup.New(sandbox), preRotationPrefix)
}

// Reseal moves every file every snapshot keeps from the old key to the new
// one of newKeyId, once the files of every database — the backup database's
// own included — are resealed and the databases armed under the new key: the
// content of each file is resealed, and an index entry is left out of its
// snapshot, since its name is a digest of the old key — a restore reindexes
// every database once it has written the files. The blobs no snapshot holds
// anymore are removed. Run again after an interruption, it reseals only what
// is left. The caller runs it with no server writing DataDir.
func Reseal(sandbox *api.Sandbox, newKeyId string) (RotationResult, error) {
	result := RotationResult{}
	if !acquire(sandbox) {
		return result, sandbox.Deps.StdDeps.Errorf("a backup job is running")
	}
	defer release(sandbox)
	db := backup.New(sandbox)

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
			if encryption.IndexEntry(sandbox, content.Path) {
				if err := backup.RemoveSnapshotContent(sandbox, db, snapshot.Id, content.Id); err != nil {
					return result, err
				}
				result.Dropped++
				continue
			}
			sha, done := resealed[content.Sha]
			if !done {
				blob, ok := db.FindBlobBySha(content.Sha)
				if !ok {
					return result, sandbox.Deps.StdDeps.Errorf("snapshot %s: the content of %s is missing", snapshot.Name, content.Path)
				}
				value, changed, err := sandbox.Deps.EncryptionDeps.Reseal(blob.Value)
				if err != nil {
					return result, sandbox.Deps.StdDeps.Errorf("snapshot %s: the file %s %s", snapshot.Name, content.Path, err.Error())
				}
				sha = content.Sha
				if changed {
					if sha, err = storeBlob(sandbox, db, value); err != nil {
						return result, err
					}
					result.Contents++
				}
				resealed[content.Sha] = sha
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
