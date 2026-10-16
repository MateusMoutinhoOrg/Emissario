package api_list_privacy_access

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Handle answers POST /api/admin/root/list-privacy-access: one page of the
// inventory of the access log reads, newest first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	accesses, err := privacy_db.New(sandbox).ListPrivacyAccesses(privacy_db.PrivacyAccessFilter{})
	if err != nil {
		return err
	}
	sandbox.Deps.SortDeps.SliceStable(accesses, func(i int, j int) bool {
		if accesses[i].At != accesses[j].At {
			return accesses[i].At > accesses[j].At
		}
		return accesses[i].Id > accesses[j].Id
	})
	paging, from, to := visualization.Paginate(sandbox, len(accesses), input.Body.Page, input.Body.Limit)
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, access := range accesses[from:to] {
		list.AddItemToArray(backofficeapi.PrivacyAccessJSON(sandbox, access))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("accesses", list)
	document.AddItemToObject("total", paging.Total)
	document.AddItemToObject("page", paging.Page)
	document.AddItemToObject("pages", paging.Pages)
	document.AddItemToObject("limit", paging.Limit)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, document)
}
