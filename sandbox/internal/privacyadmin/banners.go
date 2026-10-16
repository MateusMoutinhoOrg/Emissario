package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficeauth"
)

// The privacy administration is what reaches across every database holding
// personal data: the banners every root page shows, the retention cleanup,
// find and erase, legal holds and presets. The privacy package is its core,
// which the databases' own packages call; this one calls them.

// The kinds of a banner.
const (
	BannerDanger = "danger"
	BannerWarn   = "warn"
)

// Banner is one line every backoffice page shows a root above its content:
// what is off in the privacy setup, and where to fix it.
type Banner struct {
	Kind string
	Text string
	Href string
	Link string
}

// bannersCacheSeconds is how long the banners last computed are shown: every
// page shows them, and what they check changes rarely.
const bannersCacheSeconds = 30

// bannersCache is the banners last computed, guarded by bannersLock.
var bannersCache struct {
	banners []Banner
	at      int64
	valid   bool
}

// bannersLock guards bannersCache.
var bannersLock = make(chan struct{}, 1)

// InvalidateBanners makes the next page compute the banners again: a save
// that changes what they check calls it.
func InvalidateBanners(sandbox *api.Sandbox) {
	bannersLock <- struct{}{}
	bannersCache.valid = false
	<-bannersLock
}

// Banners answers the banners of a root's pages, nil for a viewer's. A check
// that fails is left out, never shown as an error.
func Banners(sandbox *api.Sandbox, user *backoffice_db.BackofficeUserRecord) []Banner {
	if user == nil || backofficeauth.Role(user.Role) != backofficeauth.RoleRoot {
		return nil
	}
	bannersLock <- struct{}{}
	defer func() { <-bannersLock }()
	now := privacy.NowSeconds(sandbox)
	if bannersCache.valid && now-bannersCache.at < bannersCacheSeconds && now >= bannersCache.at {
		return bannersCache.banners
	}
	banners := computeBanners(sandbox)
	bannersCache.banners = banners
	bannersCache.at = now
	bannersCache.valid = true
	return banners
}

// bannerChecks is every check a banner may come out of, in the order they
// are shown.
var bannerChecks = []func(sandbox *api.Sandbox) []Banner{
	notEncryptedBanner,
	presetBanners,
	pastRetentionBanner,
	accessLogBanners,
	rootsWithoutTotpBanner,
}

// computeBanners runs every check of bannerChecks.
func computeBanners(sandbox *api.Sandbox) []Banner {
	banners := []Banner{}
	for _, check := range bannerChecks {
		banners = append(banners, check(sandbox)...)
	}
	return banners
}

// notEncryptedBanner warns of a server running --notencrypt.
func notEncryptedBanner(sandbox *api.Sandbox) []Banner {
	if !sandbox.Config.NotEncrypt {
		return nil
	}
	return []Banner{{
		Kind: BannerDanger,
		Text: "Database not encrypted — debug only. This server runs with --notencrypt: every IP, email, contact request and summary, the access log and the key that hashes the IPs sit in clear on the disk and in every backup.",
	}}
}

// presetBanners says when the settings differ from the preset applied, when
// that preset is gone, and when a newer version of it is available — which
// is never applied by itself: the law changes, and the root decides.
func presetBanners(sandbox *api.Sandbox) []Banner {
	state, err := CheckPreset(sandbox)
	if err != nil || state.Preset == "" {
		return nil
	}
	if !state.Found {
		return []Banner{{Kind: BannerWarn, Text: "The preset applied, " + state.Preset + ", no longer exists: apply another one or detach it.", Href: privacy.PresetsPath, Link: "Presets"}}
	}
	banners := []Banner{}
	if !state.Conform {
		names := []string{}
		for _, difference := range state.Differences {
			names = append(names, difference.Setting)
		}
		count := sandbox.Deps.StringsDeps.FormatInt(int64(len(names)), 10)
		text := count + " settings differ from " + state.Preset + ": " + sandbox.Deps.StringsDeps.Join(names, ", ") + "."
		if len(names) == 1 {
			text = "1 setting differs from " + state.Preset + ": " + names[0] + "."
		}
		banners = append(banners, Banner{Kind: BannerWarn, Text: text, Href: privacy.PresetsPath, Link: "Re-apply or detach"})
	}
	if state.Newer != nil {
		banners = append(banners, Banner{Kind: BannerWarn, Text: state.Newer.Name + " v" + sandbox.Deps.StringsDeps.FormatInt(state.Newer.Version, 10) + " is available: " + state.Newer.Changes, Href: privacy.PresetPreviewPth + "?preset=" + state.Newer.Ref(), Link: "Preview it"})
	}
	return banners
}

// BackupsPath is the Backups page, which marks the snapshots past retention.
const BackupsPath = "/admin/root/list-backups"

// pastRetentionBanner warns of the snapshots holding data the databases
// already deleted by retention.
func pastRetentionBanner(sandbox *api.Sandbox) []Banner {
	past, err := PastRetention(sandbox)
	if err != nil || len(past) == 0 {
		return nil
	}
	count := sandbox.Deps.StringsDeps.FormatInt(int64(len(past)), 10)
	text := count + " snapshots hold data the databases already deleted by retention: it lives on until they are deleted."
	if len(past) == 1 {
		text = "1 snapshot holds data the databases already deleted by retention: it lives on until it is deleted."
	}
	return []Banner{{Kind: BannerWarn, Text: text, Href: BackupsPath, Link: "Backups"}}
}

// accessLogBanners warns of an access log on without what Decreto 8.771/2016
// art. 13 asks: the visitor's source port, two-factor sign-in for its
// readers, the inventory of its reads.
func accessLogBanners(sandbox *api.Sandbox) []Banner {
	settings, err := privacy.CurrentSettings(sandbox)
	if err != nil || !settings.AccessLog {
		return nil
	}
	banners := []Banner{}
	if sandbox.Config.AllowXForwardedFor && sandbox.Config.ClientPortHeader == "" {
		banners = append(banners, Banner{Kind: BannerWarn, Text: "The access log has no source port: behind a reverse proxy the port it reads is the proxy's own. Set --client-port-header to the header the proxy sets the visitor's port in (nginx: proxy_set_header X-Client-Port $remote_port).", Href: privacy.SettingsPath, Link: "Privacy"})
	}
	if !settings.AccessLogTotp {
		banners = append(banners, Banner{Kind: BannerWarn, Text: "The access log is read without two-factor sign-in (Decreto 8.771/2016 art. 13 II).", Href: privacy.SettingsPath, Link: "Privacy"})
	}
	if !settings.AccessLogInventory {
		banners = append(banners, Banner{Kind: BannerWarn, Text: "The access log keeps no inventory of its reads (Decreto 8.771/2016 art. 13 III): keep one by other means.", Href: privacy.SettingsPath, Link: "Privacy"})
	}
	return banners
}

// rootsWithoutTotpBanner lists the roots who sign in without two-factor.
func rootsWithoutTotpBanner(sandbox *api.Sandbox) []Banner {
	users, err := backoffice_db.New(sandbox).ListBackofficeUsers(backoffice_db.BackofficeUserFilter{})
	if err != nil {
		return nil
	}
	names := []string{}
	for _, user := range users {
		if backofficeauth.Role(user.Role) == backofficeauth.RoleRoot && user.TotpEnabledAt == 0 {
			names = append(names, user.Username)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return []Banner{{
		Kind: BannerWarn,
		Text: "Roots without two-factor sign-in: " + sandbox.Deps.StringsDeps.Join(names, ", ") + ".",
		Href: privacy.TotpPath,
		Link: "Set up two-factor sign-in",
	}}
}
