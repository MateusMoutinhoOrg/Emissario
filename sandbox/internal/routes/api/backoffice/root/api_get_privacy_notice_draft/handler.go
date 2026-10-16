package api_get_privacy_notice_draft

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/root/get-privacy-notice-draft: the privacy
// notice draft of the settings in use, {template, text}, in the template the
// privacy settings name. Only a root reaches it: the
// backoffice-api-root-guard answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	text, err := privacyadmin.NoticeDraft(sandbox)
	if err != nil {
		return err
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("template", settings.NoticeDraft)
	document.AddItemToObject("text", text)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, document)
}
