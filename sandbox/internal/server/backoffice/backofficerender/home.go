package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
)

// HomePage is what backoffice/home.html is rendered with.
type HomePage struct {
	Id       string
	Username string
	// Initial is the first letter of Username, shown in the avatar.
	Initial        string
	Email          string
	Role           string
	SessionMinutes int
	// IsRoot shows what only a root may do.
	IsRoot bool
	// VisualizationToday and VisualizationWeek are the views of the public
	// website today and in the last 7 days, bots left out.
	VisualizationToday string
	VisualizationWeek  string
	// ContactInbox and ContactToday are the contact requests in the inbox
	// and the ones received today.
	ContactInbox string
	ContactToday string
}

// initialOf is the first letter of name, "?" for an empty one.
func initialOf(sandbox *api.Sandbox, name string) string {
	for _, letter := range name {
		return string(letter)
	}
	return "?"
}

// RenderHomePage answers the home page for user, whose session lasts
// sessionMinutes, with the views of the public website today and in the last 7
// days, and the contact requests in the inbox and received today.
func RenderHomePage(sandbox *api.Sandbox, response *serverdeps.Response, user *backoffice_db.BackofficeUserRecord, sessionMinutes int, viewsToday int, viewsWeek int, contactInbox int, contactToday int) error {
	return RenderHTML(sandbox, response, api.StatusOK, "backoffice/home.html", HomePage{
		Id:             sandbox.Deps.StringsDeps.FormatInt(user.Id, 10),
		Username:       user.Username,
		Initial:        initialOf(sandbox, user.Username),
		Email:          user.Email,
		Role:           backofficeauth.RoleName(sandbox, backofficeauth.Role(user.Role)),
		SessionMinutes: sessionMinutes,
		IsRoot:         backofficeauth.Role(user.Role) == backofficeauth.RoleRoot,

		VisualizationToday: formatCount(sandbox, viewsToday),
		VisualizationWeek:  formatCount(sandbox, viewsWeek),

		ContactInbox: formatCount(sandbox, contactInbox),
		ContactToday: formatCount(sandbox, contactToday),
	})
}
