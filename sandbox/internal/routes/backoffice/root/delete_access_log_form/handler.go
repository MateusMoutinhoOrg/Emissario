package delete_access_log_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/delete-access-log for a reader: the rows of
// ids are deleted, but those on a legal hold. Deleting rows younger than the
// access log retention answers the page under a 409 with the erase-kept-v1
// notice until it is acknowledged — every time. It redirects to the access
// log with how many rows went.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if refusal := privacyadmin.CanReadAccessLog(sandbox, props.User, nil); refusal != nil {
		if refusal.Status != api.StatusForbidden {
			return sandbox.Deps.OpinionatedAgnosServer.Fail(refusal.Status, refusal.Field, refusal.Message)
		}
		return backofficerender.RenderForbiddenMessage(sandbox, response, props.User, "Access log readers only", refusal.Message)
	}
	ids := []int64{}
	for _, id := range input.Body.Ids {
		ids = append(ids, int64(id))
	}
	raw := privacyadmin.RawAccessFilter{Ip: input.Body.Ip, From: input.Body.From, To: input.Body.To, Kind: input.Body.Kind}
	result, refusal, err := privacyadmin.DeleteAccess(sandbox, ids, input.Body.Acknowledge)
	if err != nil {
		return err
	}
	if refusal != nil {
		notices, err := privacy.LoadNotices(sandbox, refusal.Notices, privacy.NoticeContext{})
		if err != nil {
			return err
		}
		filter, ok := privacyadmin.ParseAccessFilter(sandbox, raw)
		return backofficerender.RenderAccessLogPage(sandbox, response, api.StatusConflict, props.User, raw, filter, ok == nil, nil, visualization.Paging{}, backofficerender.Notice{}, "", notices, ids)
	}
	params := []visualization.Param{{Key: "from", Value: raw.From}, {Key: "to", Value: raw.To}, {Key: "ip", Value: raw.Ip}, {Key: "kind", Value: raw.Kind}, {Key: "notice", Value: backofficerender.AccessNoticeDeleted}, {Key: "count", Value: sandbox.Deps.StringsDeps.FormatInt(int64(result.Deleted), 10)}}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, visualization.Href(sandbox, backofficerender.AccessLogPath, params))
}
