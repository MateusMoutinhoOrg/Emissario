package api_rotate_privacy_ip_key

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
)

// Handle answers POST /api/admin/root/rotate-privacy-ip-key: the IP key is
// replaced with a new random one, so the hashes made before stop matching
// new visits and any IP typed in a filter or in Find and erase. It cannot be
// undone. It answers {ip_key_made_at}.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if err := privacyadmin.RotateIpKey(sandbox); err != nil {
		return err
	}
	made, err := privacy.IpKeyRotatedAt(sandbox)
	if err != nil {
		return err
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("ip_key_made_at", made)
	return sandbox.Deps.OpinionatedAgnosServer.WriteJSON(sandbox.Deps.SerializableDeps, *response, api.StatusOK, document)
}
