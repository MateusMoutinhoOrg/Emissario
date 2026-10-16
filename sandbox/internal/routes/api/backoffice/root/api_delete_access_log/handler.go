package api_delete_access_log

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/delete-access-log for a reader: the
// rows of ids are deleted, but those on a legal hold; rows younger than the
// access log retention answer 409 with the erase-kept-v1 notice until it is
// acknowledged, at every call. {deleted, held, missing}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if refusal := privacyadmin.CanReadAccessLog(sandbox, props.User, props.ApiToken); refusal != nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
	}
	ids := []int64{}
	for _, id := range input.Body.Ids {
		ids = append(ids, int64(id))
	}
	result, refusal, err := privacyadmin.DeleteAccess(sandbox, ids, input.Body.Acknowledge)
	if err != nil {
		return err
	}
	if refusal != nil {
		return backofficeapi.FailPrivacy(sandbox, response, refusal, privacy.NoticeContext{})
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("deleted", result.Deleted)
	document.AddItemToObject("held", result.Held)
	document.AddItemToObject("missing", result.Missing)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, document)
}
