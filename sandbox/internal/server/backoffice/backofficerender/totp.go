package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/totp"
)

// The paths of the two-factor sign-in of the user signed in.
const (
	TotpPath   = privacy.TotpPath
	TotpQrPath = "/admin/set-totp-qr"
)

// TotpNoticeEnabled is the outcome code of a confirmation that turned
// two-factor sign-in on, TotpNoticeReset of a root's reset of it.
const (
	TotpNoticeEnabled = "enabled"
	TotpNoticeReset   = "totp-reset"
)

// TotpPage is what backoffice/set_totp.html is rendered with.
type TotpPage struct {
	Viewer Viewer
	Notice Notice
	Error  string
	// Enabled is true once the user signs in with a code.
	Enabled   bool
	EnabledAt string
	// Secret is the base32 secret in groups of four, to type by hand.
	Secret string
	// QrSrc draws the same secret as a QR code.
	QrSrc  string
	Action string
	// Reader is true for a reader of the access log, who needs it.
	Reader bool
}

// groupSecret is secret in groups of four characters.
func groupSecret(secret string) string {
	out := []byte{}
	for index := 0; index < len(secret); index++ {
		if index > 0 && index%4 == 0 {
			out = append(out, ' ')
		}
		out = append(out, secret[index])
	}
	return string(out)
}

// RenderTotpPage answers the two-factor page of user under status: since
// when it is on, or the secret and QR code to turn it on, which it draws
// when there is none.
func RenderTotpPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, notice string, message string) error {
	page := TotpPage{
		Viewer: viewerOf(sandbox, user),
		Error:  message,
		Action: TotpPath,
		Reader: user.AccessLogReader == 1,
	}
	if notice == TotpNoticeEnabled {
		page.Notice = Notice{Text: "Two-factor sign-in is on: every sign-in asks for a code from now on. Your API tokens were revoked and your other sessions closed.", Kind: "ok"}
	}
	if totp.Enabled(*user) {
		page.Enabled = true
		page.EnabledAt = formatInstant(sandbox, user.TotpEnabledAt)
		return RenderHTML(sandbox, response, status, "backoffice/set_totp.html", page)
	}
	setup, err := totp.Begin(sandbox, *user)
	if err != nil {
		return err
	}
	page.Secret = groupSecret(setup.Secret)
	page.QrSrc = TotpQrPath
	return RenderHTML(sandbox, response, status, "backoffice/set_totp.html", page)
}
