package totp

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/jwtdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapitokens"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
)

// Two-factor sign-in for the backoffice: TOTP (RFC 6238, 6 digits, 30 s),
// which a user turns on from their own account page by scanning a QR code
// the server draws, or typing its base32 secret, and confirming one code.
// From then on signing in asks for a code after the password. The readers of
// the access log need it by default (Decreto 8.771/2016 art. 13 II).
//
// Turning it on revokes the user's API tokens and closes their other
// sessions, so every token and session of a reader was made after a
// two-factor sign-in. A code is accepted once: the step it matched is kept,
// and a code of that step or an earlier one is refused. A root resets
// another user's TOTP, and `emissario reset-backoffice-totp` does it from the
// host.

// skewSteps is how many 30-second steps a code may be early or late: a
// phone's clock drifts.
const skewSteps = 1

// pendingSeconds is how long the step that asks for the code waits after the
// password was accepted.
const pendingSeconds = 300

// pendingId is the `jti` of a pending sign-in token, which never names a
// session.
const pendingId = "totp-pending"

// Setup is what a user scans or types to turn TOTP on: the secret and the
// otpauth uri the QR code draws.
type Setup struct {
	Secret string
	Uri    string
}

// Enabled tells whether user signs in with TOTP.
func Enabled(user backoffice_db.BackofficeUserRecord) bool {
	return user.TotpEnabledAt > 0 && user.TotpSecret != ""
}

// nowSeconds is the current instant in Unix seconds.
func nowSeconds(sandbox *api.Sandbox) int64 {
	return sandbox.Deps.StdDeps.Now() / 1000000000
}

// issuer is the name an authenticator app lists the account under.
func issuer(sandbox *api.Sandbox) string {
	name := sandbox.Config.ProjectName
	if name == "" {
		return "Emissario"
	}
	return sandbox.Deps.StringsDeps.ToUpper(name[:1]) + name[1:]
}

// Begin answers the setup of user: the secret not confirmed yet, a new one
// when there is none. A user whose TOTP is on has none to begin.
func Begin(sandbox *api.Sandbox, user backoffice_db.BackofficeUserRecord) (Setup, error) {
	secret := user.TotpSecret
	if Enabled(user) {
		return Setup{}, sandbox.Deps.StdDeps.Errorf("two-factor sign-in is already on")
	}
	if secret == "" {
		created, err := sandbox.Deps.TotpDeps.NewSecret()
		if err != nil {
			return Setup{}, err
		}
		if err := backoffice_db.New(sandbox).SetBackofficeUserTotpSecret(user.Id, created); err != nil {
			return Setup{}, err
		}
		secret = created
	}
	return Setup{Secret: secret, Uri: sandbox.Deps.TotpDeps.Uri(issuer(sandbox), user.Username, secret)}, nil
}

// Qr answers the setup of user drawn as a PNG QR code; false when they have
// no setup pending — TOTP on, or no secret drawn yet.
func Qr(sandbox *api.Sandbox, user backoffice_db.BackofficeUserRecord) ([]byte, bool, error) {
	if Enabled(user) || user.TotpSecret == "" {
		return nil, false, nil
	}
	png, err := sandbox.Deps.TotpDeps.QrPng(sandbox.Deps.TotpDeps.Uri(issuer(sandbox), user.Username, user.TotpSecret))
	return png, err == nil, err
}

// Confirm turns TOTP on for user when code is one of the secret Begin gave
// them: their API tokens are revoked and every session but keep is closed.
// ok is false for a code that does not match.
func Confirm(sandbox *api.Sandbox, user backoffice_db.BackofficeUserRecord, keep *backoffice_db.BackofficeUserSessionRecord, code string) (bool, error) {
	if Enabled(user) || user.TotpSecret == "" {
		return false, nil
	}
	now := nowSeconds(sandbox)
	step, ok := sandbox.Deps.TotpDeps.Verify(user.TotpSecret, code, now, skewSteps)
	if !ok {
		return false, nil
	}
	db := backoffice_db.New(sandbox)
	if err := db.SetBackofficeUserTotpLastStep(user.Id, step); err != nil {
		return false, err
	}
	if err := db.SetBackofficeUserTotpEnabledAt(user.Id, now); err != nil {
		return false, err
	}
	if err := backofficeapitokens.RemoveOfOwner(sandbox, user.Id); err != nil {
		return true, err
	}
	return true, backofficeauth.CloseSessions(sandbox, user.Id, keep)
}

// Check tells whether code signs user in: one of their secret for a step
// later than the last one accepted, which it then keeps.
func Check(sandbox *api.Sandbox, user backoffice_db.BackofficeUserRecord, code string) (bool, error) {
	if !Enabled(user) {
		return false, nil
	}
	step, ok := sandbox.Deps.TotpDeps.Verify(user.TotpSecret, code, nowSeconds(sandbox), skewSteps)
	if !ok || step <= user.TotpLastStep {
		return false, nil
	}
	return true, backoffice_db.New(sandbox).SetBackofficeUserTotpLastStep(user.Id, step)
}

// Reset turns TOTP off for the user of userId and forgets their secret: they
// sign in with their password alone until they turn it on again.
func Reset(sandbox *api.Sandbox, userId int64) error {
	db := backoffice_db.New(sandbox)
	if _, ok := db.FindBackofficeUserById(userId); !ok {
		return sandbox.Deps.StdDeps.Errorf("no backoffice user %d", userId)
	}
	for _, write := range []func() error{
		func() error { return db.SetBackofficeUserTotpSecret(userId, "") },
		func() error { return db.SetBackofficeUserTotpEnabledAt(userId, 0) },
		func() error { return db.SetBackofficeUserTotpLastStep(userId, 0) },
	} {
		if err := write(); err != nil {
			return err
		}
	}
	return nil
}

// pendingSecret is the secret a pending sign-in token is signed with: the
// session secret bound to this purpose, so a session token never reads as a
// pending one, nor the other way round.
func pendingSecret(sandbox *api.Sandbox) string {
	return sandbox.Config.SessionSecret + "\x00totp-pending"
}

// IssuePending signs the token of a sign-in whose password was accepted and
// whose code is awaited: it names user, is bound to the client ip and lasts
// pendingSeconds. It carries no password.
func IssuePending(sandbox *api.Sandbox, user backoffice_db.BackofficeUserRecord, ip string) (string, error) {
	now := nowSeconds(sandbox)
	return sandbox.Deps.JwtDeps.Sign(jwtdeps.Claims{
		Id:        pendingId,
		Subject:   sandbox.Deps.StringsDeps.FormatInt(user.Id, 10),
		IssuedAt:  now,
		ExpiresAt: now + pendingSeconds,
		Ip:        ip,
	}, pendingSecret(sandbox))
}

// ResolvePending answers the user a pending sign-in token names, false when
// it is not one, expired, or was issued to another ip.
func ResolvePending(sandbox *api.Sandbox, token string, ip string) (backoffice_db.BackofficeUserRecord, bool) {
	if token == "" || ip == "" {
		return backoffice_db.BackofficeUserRecord{}, false
	}
	claims, err := sandbox.Deps.JwtDeps.Parse(token, pendingSecret(sandbox))
	if err != nil || claims.Id != pendingId || claims.Ip != ip {
		return backoffice_db.BackofficeUserRecord{}, false
	}
	userId, err := sandbox.Deps.StringsDeps.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return backoffice_db.BackofficeUserRecord{}, false
	}
	user, ok := backoffice_db.New(sandbox).FindBackofficeUserById(userId)
	if !ok || !Enabled(user) {
		return backoffice_db.BackofficeUserRecord{}, false
	}
	return user, true
}
