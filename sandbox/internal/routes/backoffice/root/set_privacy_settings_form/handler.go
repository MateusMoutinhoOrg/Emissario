package set_privacy_settings_form

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacyadmin"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/routeprops"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficerender"
)

// Handle answers POST /admin/root/set-privacy-settings: the privacy settings
// are checked and saved, and the browser is sent back to the page with
// ?notice=saved. A refused value answers the page again, filled with what
// was sent, under a 400 with the reason above it; a save that needs notices
// answers it under a 409 with the notices inside the form, to confirm. Only
// a root reaches it: the backoffice-root-guard answers anyone else under
// /admin/root first.
func Handle(sandbox *api.Sandbox, props *routeprops.RouteProps, input *Input, response *serverdeps.Response) error {
	if props.User == nil {
		return sandbox.Deps.OpinionatedAgnosServer.Fail(api.StatusUnauthorized, "", "no authenticated user")
	}
	checked := func(value string) bool { return value == "1" || value == "on" }
	sent := backofficerender.PrivacySettingsForm{
		IpMode:                 input.Body.IpMode,
		ViewerIp:               input.Body.ViewerIp,
		AccessLog:              checked(input.Body.AccessLog),
		AccessLogRetentionDays: input.Body.AccessLogRetentionDays,
		AccessLogTotp:          checked(input.Body.AccessLogTotp),
		AccessLogInventory:     checked(input.Body.AccessLogInventory),
		IpRetentionDays:        input.Body.IpRetentionDays,
		BackupRetentionDays:    input.Body.BackupRetentionDays,
		NoticeDraft:            input.Body.NoticeDraft,
	}
	settingsInput, refusal := privacyadmin.InputOfForm(sandbox, privacyadmin.SettingsForm{
		IpMode:                 input.Body.IpMode,
		ViewerIp:               input.Body.ViewerIp,
		AccessLog:              input.Body.AccessLog,
		AccessLogRetentionDays: input.Body.AccessLogRetentionDays,
		AccessLogTotp:          input.Body.AccessLogTotp,
		AccessLogInventory:     input.Body.AccessLogInventory,
		IpRetentionDays:        input.Body.IpRetentionDays,
		BackupRetentionDays:    input.Body.BackupRetentionDays,
		NoticeDraft:            input.Body.NoticeDraft,
		Acknowledge:            input.Body.Acknowledge,
	})
	if refusal == nil {
		var err error
		var settings privacy.Settings
		settings, refusal, err = privacyadmin.ApplySettings(sandbox, settingsInput)
		if err != nil {
			return err
		}
		if refusal != nil && len(refusal.Notices) > 0 {
			notices, err := privacyadmin.LoadSettingsNotices(sandbox, settings, refusal.Notices)
			if err != nil {
				return err
			}
			sent.Notices = notices
			return backofficerender.RenderPrivacySettingsPage(sandbox, response, refusal.Status, props.User, sent, "", 0, "", "")
		}
	}
	if refusal != nil {
		return backofficerender.RenderPrivacySettingsPage(sandbox, response, refusal.Status, props.User, sent, "", 0, "", refusal.Message)
	}
	return sandbox.Deps.OpinionatedAgnosServer.Redirect(*response, api.StatusSeeOther, privacy.SettingsPath+"?notice="+backofficerender.PrivacyNoticeSaved)
}
