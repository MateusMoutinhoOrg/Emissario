package api

// PrivacyConfig is the part of the Config the privacy mechanic reads — the
// encryption at rest of every database, and how the access log learns the
// source port of a visitor — embedded in api.Config so each field is read as
// sandbox.Config.<Field>. The database-encryption middleware fills the first
// three in front of every command that opens a database, and the
// privacy-start-server middleware the last one in front of start-server.
type PrivacyConfig struct {
	// Encrypted is true when every database under DatabaseDir is encrypted
	// at rest with the key of EMISSARIO_ENCRYPT_KEY, false when the command
	// runs with --notencrypt.
	Encrypted bool

	// EncryptionKeyId is the id of that key: the hex of four bytes derived
	// from it, which every value and <database>/encryption.json carry. It
	// reveals nothing about the key. Empty with --notencrypt.
	EncryptionKeyId string

	// NotEncrypt is true when the command runs with --notencrypt: the
	// databases are in clear, for local debugging only, and every backoffice
	// page says so.
	NotEncrypt bool

	// ClientPortHeader is the header a reverse proxy sets to the source port
	// of the visitor's connection — `--client-port-header`, X-Client-Port in
	// the nginx example — trusted only with --allow-x-forwarded-for. Empty,
	// the access log records the port of the connection the server itself
	// accepted.
	ClientPortHeader string
}
