package backofficerender

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serverdeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/blocker"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// The page of the IP blocks only a root reaches, under /admin/root: every
// block, the form that blocks an ip, a range or a hashed ip, and the one that
// blocks a country.

// IpBlockForm is what the add form is filled with.
type IpBlockForm struct {
	Value  string
	Reason string
	// Days is blank to block until a root removes it.
	Days string
}

// IpBlockLine is one block of the list.
type IpBlockLine struct {
	Id    string
	Value string
	// Kind is the kind of the block, and Label how the page names it.
	Kind  string
	Label string
	// Flag is the flag of a country block, "" for any other.
	Flag    string
	Reason  string
	Created string
	// Expires is when the block ends, "" for never; Expired is true once it
	// did.
	Expires string
	Expired bool
	// Stale is true for a hashed ip made under an IP key rotated since.
	Stale bool
	Hits  string
}

// IpBlockCountry is one country of the country form.
type IpBlockCountry struct {
	Code string
	Name string
	Flag string
}

// IpBlocksPage is what backoffice/ip_blocks.html is rendered with.
type IpBlocksPage struct {
	Viewer       Viewer
	Notice       Notice
	Error        string
	Form         IpBlockForm
	Rows         []IpBlockLine
	Countries    []IpBlockCountry
	AddAction    string
	RemoveAction string
	// ClientIp and ClientCountry are where the root is connected from: a
	// block that matches them is refused.
	ClientIp      string
	ClientCountry string
	// Geoip is false while no country database is open: a country block
	// then matches no one.
	Geoip bool
	// HashedMode is true when the records keep hashed IPs.
	HashedMode bool
	Max        int
	MaxDays    int
}

// ipBlockNotice is the sentence a notice code of the blocks page reads as.
func ipBlockNotice(sandbox *api.Sandbox, code string, count int) Notice {
	switch code {
	case blocker.NoticeAdded:
		return Notice{Text: "Blocked: requests from it are turned away from now on.", Kind: "ok"}
	case blocker.NoticeRemoved:
		if count > 0 {
			return Notice{Text: plural(sandbox, count, "block", "blocks") + " removed.", Kind: "ok"}
		}
		return Notice{Text: "Removed.", Kind: "ok"}
	}
	return Notice{}
}

// ipBlockLabel is how the page names a kind of block.
func ipBlockLabel(kind string) string {
	switch kind {
	case blocker.KindIp:
		return "IP"
	case blocker.KindRange:
		return "Range"
	case blocker.KindHash:
		return "Hashed IP"
	case blocker.KindCountry:
		return "Country"
	}
	return kind
}

// RenderIpBlocksPage answers, under status, every block beside the forms
// that add one, the value form filled with form and with message above it —
// or the sentence of notice, count being how many blocks it removed.
// clientIp is the ip of the root reading it.
func RenderIpBlocksPage(sandbox *api.Sandbox, response *serverdeps.Response, status int, user *backoffice_db.BackofficeUserRecord, clientIp string, form IpBlockForm, notice string, count int, message string) error {
	blocks, err := blocker.List(sandbox)
	if err != nil {
		return err
	}
	settings, err := privacy.CurrentSettings(sandbox)
	if err != nil {
		return err
	}
	rotatedAt, err := privacy.IpKeyRotatedAt(sandbox)
	if err != nil {
		return err
	}
	page := IpBlocksPage{
		Viewer:        viewerOf(sandbox, user),
		Notice:        ipBlockNotice(sandbox, notice, count),
		Error:         message,
		Form:          form,
		AddAction:     blocker.AddPath,
		RemoveAction:  blocker.RemovePath,
		ClientIp:      clientIp,
		ClientCountry: visualization.CountryName(sandbox, visualization.Country(sandbox, clientIp)),
		Geoip:         visualization.CurrentGeoip(sandbox).Open,
		HashedMode:    settings.IpMode == privacy.IpModeHashed,
		Max:           blocker.MaxBlocks,
		MaxDays:       blocker.MaxDays,
	}
	now := blocker.NowSeconds(sandbox)
	for _, block := range blocks {
		line := IpBlockLine{
			Id:      sandbox.Deps.StringsDeps.FormatInt(block.Id, 10),
			Value:   block.Value,
			Kind:    block.Kind,
			Label:   ipBlockLabel(block.Kind),
			Reason:  block.Reason,
			Created: formatInstant(sandbox, block.CreatedAt),
			Expired: blocker.Expired(block, now),
			Stale:   blocker.Stale(sandbox, block, rotatedAt),
			Hits:    formatCount(sandbox, int(blocker.Hits(sandbox, block.Value))),
		}
		if block.ExpiresAt > 0 {
			line.Expires = formatInstant(sandbox, block.ExpiresAt)
		}
		if block.Kind == blocker.KindCountry {
			line.Flag = visualization.CountryFlag(sandbox, block.Value)
			line.Value = block.Value + " · " + visualization.CountryName(sandbox, block.Value)
		}
		page.Rows = append(page.Rows, line)
	}
	for _, code := range visualization.Countries(sandbox) {
		page.Countries = append(page.Countries, IpBlockCountry{Code: code, Name: visualization.CountryName(sandbox, code), Flag: visualization.CountryFlag(sandbox, code)})
	}
	return RenderHTML(sandbox, response, status, "backoffice/ip_blocks.html", page)
}

// IpBlockHref is the blocks page with its add form filled with stored — an
// ip or a hash as a record keeps it — for the Block link of a record; "" for
// a record without one.
func IpBlockHref(sandbox *api.Sandbox, stored string) string {
	if stored == "" {
		return ""
	}
	return visualization.Href(sandbox, blocker.BlocksPath, []visualization.Param{{Key: "value", Value: stored}})
}
