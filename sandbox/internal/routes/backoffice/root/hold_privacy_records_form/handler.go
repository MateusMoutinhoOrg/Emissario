package hold_privacy_records_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /admin/root/hold-privacy-records: the rows of ids of
// table are held for reason until the end of the day until. It redirects to
// back — a backoffice page, the Privacy page by default — with
// ?notice=held&count=<n>, or ?notice=hold-refused and why.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	ids := []int64{}
	for _, id := range input.Body.Ids {
		ids = append(ids, int64(id))
	}
	back := input.Body.Back
	if !sandbox.Deps.StringsDeps.HasPrefix(back, "/admin/") || sandbox.Deps.StringsDeps.Contains(back, "?") {
		back = privacy.SettingsPath
	}
	hold, refusal, err := privacyadmin.HoldRecords(sandbox, props.User, nil, input.Body.Table, ids, input.Body.Reason, input.Body.Until)
	if err != nil {
		return err
	}
	params := []visualization.Param{{Key: "notice", Value: "held"}, {Key: "count", Value: sandbox.Deps.StringsDeps.FormatInt(int64(len(hold.Ids)), 10)}}
	if refusal != nil {
		params = []visualization.Param{{Key: "notice", Value: "hold-refused"}, {Key: "reason", Value: refusal.Message}}
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, visualization.Href(sandbox, back, params))
}
