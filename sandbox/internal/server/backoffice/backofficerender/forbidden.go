package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
)

// ForbiddenPage is what backoffice/forbidden.html is rendered with.
type ForbiddenPage struct {
	Viewer Viewer
	// Title and Message say why, "" for the page that asks for a root.
	Title   string
	Message string
}

// RenderForbiddenPage answers, under a 403, the page telling user that what they asked
// for needs the root role.
func RenderForbiddenPage(sandbox *api.Sandbox, response *serverdeps.Response, user *backoffice_db.BackofficeUserRecord) error {
	return RenderHTML(sandbox, response, api.StatusForbidden, "backoffice/forbidden.html", ForbiddenPage{Viewer: viewerOf(sandbox, user)})
}

// RenderForbiddenMessage answers, under a 403, the page telling user why
// they may not see what they asked for: title, and message.
func RenderForbiddenMessage(sandbox *api.Sandbox, response *serverdeps.Response, user *backoffice_db.BackofficeUserRecord, title string, message string) error {
	return RenderHTML(sandbox, response, api.StatusForbidden, "backoffice/forbidden.html", ForbiddenPage{Viewer: viewerOf(sandbox, user), Title: title, Message: message})
}
