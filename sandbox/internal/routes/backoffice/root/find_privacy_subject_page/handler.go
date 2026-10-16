package find_privacy_subject_page

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers GET /admin/root/find-privacy-subject: with no ip and no
// text, the search form; else every record of that person, each marked
// deleted or kept by Delete all. A search that shows access log rows to a
// reader is inventoried. Only a root reaches it: the backoffice-root-guard
// answers anyone else under /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	if input.Ip == "" && input.Text == "" {
		return backofficerender.RenderPrivacySubjectPage(sandbox, response, api.StatusOK, props.User, "", "", nil, nil, false, input.Notice, input.Count, input.Kept, "")
	}
	subject, refusal := privacyadmin.ParseSubject(sandbox, input.Ip, input.Text)
	if refusal != nil {
		return backofficerender.RenderPrivacySubjectPage(sandbox, response, refusal.Status, props.User, input.Ip, input.Text, nil, nil, false, "", 0, 0, refusal.Message)
	}
	found, err := privacyadmin.FindSubject(sandbox, props.User, nil, subject)
	if err != nil {
		return err
	}
	return backofficerender.RenderPrivacySubjectPage(sandbox, response, api.StatusOK, props.User, input.Ip, input.Text, &found, nil, false, input.Notice, input.Count, input.Kept, "")
}
