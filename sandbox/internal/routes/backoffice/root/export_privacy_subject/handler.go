package export_privacy_subject

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers GET /admin/root/export-privacy-subject: every record of the
// person of ip and text, downloaded as JSON — what an access or a
// portability request asks for. A search that holds access log rows of a
// reader is inventoried.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	subject, refusal := privacyadmin.ParseSubject(sandbox, input.Ip, input.Text)
	if refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	found, err := privacyadmin.FindSubject(sandbox, props.User, nil, subject)
	if err != nil {
		return err
	}
	response.SetHeader("Content-Disposition", "attachment; filename=\"privacy-subject.json\"")
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.SubjectJSON(sandbox, found))
}
