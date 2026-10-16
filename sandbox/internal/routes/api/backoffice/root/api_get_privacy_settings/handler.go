package api_get_privacy_settings

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeapi"
)

// Handle answers POST /api/admin/root/get-privacy-settings: the privacy
// settings, the encryption at rest and every table of personal data this
// server keeps. Only a root reaches it: the backoffice-api-root-guard
// answers anyone else under /api/admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return err
	}
	made, err := privacy.IpKeyRotatedAt(sandbox)
	if err != nil {
		return err
	}
	kept, err := privacyadmin.KeptTables(sandbox)
	if err != nil {
		return err
	}
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, backofficeapi.PrivacyStateJSON(sandbox, settings, made, kept))
}
