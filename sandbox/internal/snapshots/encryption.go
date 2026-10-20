package snapshots

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// A snapshot copies the files of every database as they are on the disk, so
// in an encrypted --database folder it holds them as Keep sealed them: it
// records the id of the key they are sealed under (KeyId, blank for a folder
// in clear), the archive of it carries that id in its manifest, and only a
// server running with the same key restores or uploads it. A key rotation
// reseals every file of every snapshot and leaves out its index entries —
// named by a digest of the old key — so they stay restorable with the new
// key: a restore reindexes every database once its files are written.

// currentKeyId is the id of the key the databases here are encrypted with,
// "" when they are in clear (--notencrypt).
func currentKeyId(sandbox *api.Sandbox) string {
	return sandbox.Config.EncryptionKeyId
}

// keyRefusal answers why a snapshot sealed under keyId — "" for one taken in
// clear — cannot be restored here, "" when it can.
func keyRefusal(sandbox *api.Sandbox, keyId string) string {
	current := currentKeyId(sandbox)
	switch {
	case current == "" && keyId != "":
		return "the snapshot is encrypted (key_id " + keyId + ") and the databases here are in clear: restore it on a server started with its key"
	case current != "" && keyId == "":
		return "the snapshot was taken in clear (--notencrypt, or an earlier version) and the databases here are encrypted: a database in clear is never restored into an encrypted folder"
	case current != keyId:
		return "the snapshot is encrypted with another key (key_id " + keyId + ", this server's is " + current + ")"
	}
	return ""
}

// contentRefusal answers why the file at path, holding content, cannot be
// restored here, "" when it can: in an encrypted folder every file Keep wrote
// decrypts under the current key.
func contentRefusal(sandbox *api.Sandbox, path string, content []byte) string {
	if currentKeyId(sandbox) == "" {
		return ""
	}
	if !sandbox.Deps.EncryptionDeps.Opens(content) {
		return "the file " + path + " does not decrypt with this server's key: it is in clear, encrypted with another key, or was changed"
	}
	return ""
}
