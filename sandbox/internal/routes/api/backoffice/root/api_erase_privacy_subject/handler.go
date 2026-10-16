package api_erase_privacy_subject

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/erase-privacy-subject: every record of
// the person of ip and text is deleted but the held ones and — unless
// also_access_log — the access log rows. Deleting access log rows answers
// 409 acknowledgement-required with erase-kept-v1, every time, until it is
// in acknowledge. Nothing records who erased.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	subject, refusal := privacyadmin.ParseSubject(sandbox, input.Body.Ip, input.Body.Text)
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	erased, refusal, err := privacyadmin.EraseSubject(sandbox, props.User, props.ApiToken, subject, input.Body.AlsoAccessLog, input.Body.Acknowledge)
	if err != nil {
		return err
	}
	if refusal != nil {
		return backofficeapi.FailPrivacy(sandbox, response, refusal, privacy.NoticeContext{})
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.ErasedJSON(sandbox, erased))
}
