package erase_privacy_subject_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/erase-privacy-subject: every record of the
// person of ip and text is deleted but the held ones and — unless
// also-access-log — the access log rows. Deleting access log rows answers
// the page under a 409 with the erase-kept-v1 notice, every time, until it
// is acknowledged. It redirects back to the search with how many rows it
// deleted and kept. Nothing records who erased.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	alsoAccess := input.Body.AlsoAccessLog == "1" || input.Body.AlsoAccessLog == "on"
	subject, refusal := privacyadmin.ParseSubject(sandbox, input.Body.Ip, input.Body.Text)
	if refusal != nil {
		return backofficerender.RenderPrivacySubjectPage(sandbox, response, refusal.Status, props.User, input.Body.Ip, input.Body.Text, nil, nil, false, "", 0, 0, refusal.Message)
	}
	erased, refusal, err := privacyadmin.EraseSubject(sandbox, props.User, nil, subject, alsoAccess, input.Body.Acknowledge)
	if err != nil {
		return err
	}
	if refusal != nil && len(refusal.Notices) > 0 {
		found, err := privacyadmin.FindSubject(sandbox, props.User, nil, subject)
		if err != nil {
			return err
		}
		notices, err := privacy.LoadNotices(sandbox, refusal.Notices, privacy.NoticeContext{})
		if err != nil {
			return err
		}
		return backofficerender.RenderPrivacySubjectPage(sandbox, response, refusal.Status, props.User, input.Body.Ip, input.Body.Text, &found, notices, true, "", 0, 0, "")
	}
	if refusal != nil {
		return backofficerender.RenderPrivacySubjectPage(sandbox, response, refusal.Status, props.User, input.Body.Ip, input.Body.Text, nil, nil, false, "", 0, 0, refusal.Message)
	}
	format := sandbox.Deps.StringsDeps.FormatInt
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, visualization.Href(sandbox, privacy.FindPath, []visualization.Param{
		{Key: "ip", Value: input.Body.Ip},
		{Key: "text", Value: input.Body.Text},
		{Key: "notice", Value: backofficerender.PrivacyNoticeErased},
		{Key: "count", Value: format(int64(erased.Total()), 10)},
		{Key: "kept", Value: format(int64(erased.Held+erased.KeptByLaw), 10)},
	}))
}
