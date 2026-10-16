package hold_privacy_records_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/hold-privacy-records: the rows of ids of
// table are held for reason until the end of the day until. It redirects to
// back — a backoffice page, with no query or one ?id= of a record, the
// Privacy page by default — with notice=held&count=<n>, or
// notice=hold-refused and why.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	ids := []int64{}
	for _, id := range input.Body.Ids {
		ids = append(ids, int64(id))
	}
	back, params := privacyadmin.BackOf(sandbox, input.Body.Back)
	hold, refusal, err := privacyadmin.HoldRecords(sandbox, props.User, nil, input.Body.Table, ids, input.Body.Reason, input.Body.Until)
	if err != nil {
		return err
	}
	if refusal != nil {
		params = append(params, visualization.Param{Key: "notice", Value: privacyadmin.NoticeHoldRefused}, visualization.Param{Key: "reason", Value: refusal.Message})
	} else {
		params = append(params, visualization.Param{Key: "notice", Value: privacyadmin.NoticeHeld}, visualization.Param{Key: "count", Value: sandbox.Deps.StringsDeps.FormatInt(int64(len(hold.Ids)), 10)})
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, visualization.Href(sandbox, back, params))
}
