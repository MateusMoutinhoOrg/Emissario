package encryption

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// Every database under the --database folder is encrypted at rest: every
// value the store writes is sealed with XChaCha20-Poly1305 under a key read
// once from the environment (deps.Deps.EncryptionDeps, which never hands the
// key to the sandbox). The folder records which key in FileName, at its root:
// {"version":1,"algorithm":"XChaCha20-Poly1305","key_id":"<8 hex>"}, which
// reveals nothing about the key. Open decides, in front of every command that
// opens a database, whether the command runs and how the store is armed.

// FileName is the file at the root of the --database folder naming the key
// its databases are encrypted with.
const FileName = "encryption.json"

// Algorithm is the cipher every value is sealed with.
const Algorithm = "XChaCha20-Poly1305"

// Version is the format of FileName and of every sealed value.
const Version = 1

// Info is what FileName holds.
type Info struct {
	Version   int64
	Algorithm string
	KeyId     string
}

// The states a --database folder is found in.
const (
	// FolderEmpty is a folder that holds no database yet, or does not exist.
	FolderEmpty = "empty"
	// FolderEncrypted is a folder holding FileName.
	FolderEncrypted = "encrypted"
	// FolderPlaintext is a folder holding a database and no FileName: one
	// created with --notencrypt, or by a version before encryption.
	FolderPlaintext = "plaintext"
)

// KeyEnv is the environment variable the key is read from:
// EMISSARIO_ENCRYPT_KEY, named after the project like the backoffice secret.
func KeyEnv(sandbox *api.Sandbox) string {
	return envPrefix(sandbox) + "_ENCRYPT_KEY"
}

// envPrefix is the project name upper-cased, every character outside A-Z
// and 0-9 an underscore.
func envPrefix(sandbox *api.Sandbox) string {
	name := []byte(sandbox.Deps.StringsDeps.ToUpper(sandbox.Config.ProjectName))
	for i, char := range name {
		if !(char >= 'A' && char <= 'Z') && !(char >= '0' && char <= '9') {
			name[i] = '_'
		}
	}
	return string(name)
}

// Path is FileName inside the --database folder.
func Path(sandbox *api.Sandbox) string {
	return sandbox.Deps.IoDeps.Join(sandbox.Config.DatabaseDir, FileName)
}

// Skip is how many leading segments of every key the store writes name the
// --database folder: they are left out of what a value is bound to, so a
// snapshot restored into another folder still opens.
func Skip(sandbox *api.Sandbox) int {
	count := 0
	for _, segment := range sandbox.Deps.StringsDeps.Split(sandbox.Config.DatabaseDir, "/") {
		if segment != "" && segment != "." {
			count++
		}
	}
	return count
}

// ReadInfo reads FileName. found is false when the folder holds none.
func ReadInfo(sandbox *api.Sandbox) (info Info, found bool, err error) {
	io := sandbox.Deps.IoDeps
	path := Path(sandbox)
	if !io.IsFile(path) {
		return Info{}, false, nil
	}
	content, err := io.ReadFile(path)
	if err != nil {
		return Info{}, true, err
	}
	document, err := sandbox.Deps.SerializableDeps.ParseJson(string(content))
	if err != nil || !document.IsObject() {
		return Info{}, true, sandbox.Deps.StdDeps.Errorf("%s is not the JSON object a database folder records its key in", path)
	}
	if item, err := document.GetObjectItem("version"); err == nil && item != nil && item.IsInt() {
		info.Version, _ = item.GetInt()
	}
	if item, err := document.GetObjectItem("algorithm"); err == nil && item != nil && item.IsString() {
		info.Algorithm, _ = item.GetString()
	}
	if item, err := document.GetObjectItem("key_id"); err == nil && item != nil && item.IsString() {
		info.KeyId, _ = item.GetString()
	}
	if info.KeyId == "" {
		return info, true, sandbox.Deps.StdDeps.Errorf("%s names no key_id", path)
	}
	if info.Version != Version || info.Algorithm != Algorithm {
		return info, true, sandbox.Deps.StdDeps.Errorf("%s records version %d of %s, and this emissario reads version %d of %s", path, info.Version, info.Algorithm, Version, Algorithm)
	}
	return info, true, nil
}

// WriteInfo writes FileName for the key of keyId.
func WriteInfo(sandbox *api.Sandbox, keyId string) error {
	serializable := sandbox.Deps.SerializableDeps
	document := serializable.CreateObject()
	document.AddItemToObject("version", int64(Version))
	document.AddItemToObject("algorithm", Algorithm)
	document.AddItemToObject("key_id", keyId)
	sandbox.Deps.IoDeps.CreateDir(sandbox.Config.DatabaseDir)
	return sandbox.Deps.IoDeps.WriteFile(Path(sandbox), []byte(serializable.SerializeToJson(document)+"\n"))
}

// Folder is the state the --database folder is in, and what its FileName
// holds when it has one.
func Folder(sandbox *api.Sandbox) (string, Info, error) {
	info, found, err := ReadInfo(sandbox)
	if err != nil {
		return "", info, err
	}
	if found {
		return FolderEncrypted, info, nil
	}
	io := sandbox.Deps.IoDeps
	if !io.IsDir(sandbox.Config.DatabaseDir) || len(io.ListAll(sandbox.Config.DatabaseDir)) == 0 {
		return FolderEmpty, info, nil
	}
	return FolderPlaintext, info, nil
}

// Open reads the key, removes it from the environment, checks it against the
// --database folder and arms the store, as the table of docs/Privacy says:
//
//	folder              key  --notencrypt  result
//	empty               yes  -             encrypted, FileName written
//	empty or plaintext  no   yes           in clear
//	any                 no   no            refused: no key
//	any                 yes  yes           refused: contradictory
//	encrypted           its  -             encrypted
//	encrypted           other -            refused: another key
//	encrypted           no   yes           refused: the key is needed
//	plaintext           yes  -             refused: start a new folder
//
// The error, when the command may not run, says why and what to do; it
// never quotes the key. On success it fills sandbox.Config.Encrypted,
// EncryptionKeyId and NotEncrypt.
func Open(sandbox *api.Sandbox, notEncrypt bool) error {
	encryption := sandbox.Deps.EncryptionDeps
	std := sandbox.Deps.StdDeps
	env := KeyEnv(sandbox)
	keyId, present, err := encryption.LoadKey(env)
	if err != nil {
		return err
	}
	if present && notEncrypt {
		return std.Errorf("%s is set and --notencrypt was passed: choose one — the key to encrypt the databases, or --notencrypt to keep them in clear for debugging", env)
	}
	if !present && !notEncrypt {
		return std.Errorf("set %s to the key the databases are encrypted with (generate one with `%s generate-encrypt-key`, and keep it safe: losing it loses every database and every backup), or pass --notencrypt to keep them in clear for local debugging", env, sandbox.Config.ProjectName)
	}
	folder, info, err := Folder(sandbox)
	if err != nil {
		return err
	}
	switch folder {
	case FolderEncrypted:
		if notEncrypt {
			return std.Errorf("the databases under %s are encrypted: they cannot be opened without their key; set %s instead of passing --notencrypt", sandbox.Config.DatabaseDir, env)
		}
		if keyId != info.KeyId {
			return std.Errorf("%s does not match the databases under %s: they were encrypted with another key (key_id %s, this one is %s)", env, sandbox.Config.DatabaseDir, info.KeyId, keyId)
		}
	case FolderPlaintext:
		if present {
			return std.Errorf("the databases under %s are not encrypted: they were created with --notencrypt, or by an earlier version; start a new folder (--database) to encrypt them, a folder in clear is never encrypted in place", sandbox.Config.DatabaseDir)
		}
	case FolderEmpty:
		if present {
			if err := WriteInfo(sandbox, keyId); err != nil {
				return err
			}
		}
	}
	if notEncrypt {
		encryption.Plaintext()
		sandbox.Config.NotEncrypt = true
		return nil
	}
	if err := encryption.Encrypt(Skip(sandbox)); err != nil {
		return err
	}
	sandbox.Config.Encrypted = true
	sandbox.Config.EncryptionKeyId = keyId
	return nil
}

// Warning is what --notencrypt prints on stderr at every start.
func Warning(sandbox *api.Sandbox) string {
	return "warning: --notencrypt: the databases under " + sandbox.Config.DatabaseDir + " are not encrypted — debug only. Every IP, email, contact request, summary, the access log and the key that hashes the IPs sit in clear on the disk and in every backup; never run a website this way.\n"
}
