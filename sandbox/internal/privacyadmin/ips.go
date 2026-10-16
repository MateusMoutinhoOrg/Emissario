package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// pageSize is how many records a sweep over a table reads at a time.
const pageSize = 500

// HashedIps is how many full IPs HashStoredIps turned into hashes, per table.
type HashedIps struct {
	Views    int
	Contacts int
	Usages   int
}

// Total is every IP HashStoredIps turned into a hash.
func (hashed HashedIps) Total() int {
	return hashed.Views + hashed.Contacts + hashed.Usages
}

// HashStoredIps turns every full IP visualization-view, contact-request and
// agent-usage keep into its hash under the current IP key, and blanks the
// full user agent of every view it hashes: the hashed mode keeps neither.
// It cannot be undone. The access log keeps its full IPs: it is kept under
// its own legal obligation.
func HashStoredIps(sandbox *api.Sandbox) (HashedIps, error) {
	hashed := HashedIps{}
	privacy.HoldForBackups(sandbox)

	views := visualization_db.New(sandbox)
	for offset := 0; ; offset += pageSize {
		page, err := views.ListVisualizationViewsPage(offset, pageSize)
		if err != nil {
			return hashed, err
		}
		for _, view := range page {
			if view.Ip == "" || privacy.IsHash(sandbox, view.Ip) {
				continue
			}
			hash, err := privacy.HashIp(sandbox, view.Ip)
			if err != nil {
				return hashed, err
			}
			if err := views.SetVisualizationViewIp(view.Id, hash); err != nil {
				return hashed, err
			}
			if view.UserAgent != "" {
				if err := views.SetVisualizationViewUserAgent(view.Id, ""); err != nil {
					return hashed, err
				}
			}
			hashed.Views++
		}
		if len(page) < pageSize {
			break
		}
	}

	contacts := contact_db.New(sandbox)
	for offset := 0; ; offset += pageSize {
		page, err := contacts.ListContactRequestsPage(offset, pageSize)
		if err != nil {
			return hashed, err
		}
		for _, request := range page {
			if request.Ip == "" || privacy.IsHash(sandbox, request.Ip) {
				continue
			}
			hash, err := privacy.HashIp(sandbox, request.Ip)
			if err != nil {
				return hashed, err
			}
			if err := contacts.SetContactRequestIp(request.Id, hash); err != nil {
				return hashed, err
			}
			hashed.Contacts++
		}
		if len(page) < pageSize {
			break
		}
	}

	usages := agent_db.New(sandbox)
	for offset := 0; ; offset += pageSize {
		page, err := usages.ListAgentUsagesPage(offset, pageSize)
		if err != nil {
			return hashed, err
		}
		for _, usage := range page {
			if usage.Ip == "" || privacy.IsHash(sandbox, usage.Ip) {
				continue
			}
			hash, err := privacy.HashIp(sandbox, usage.Ip)
			if err != nil {
				return hashed, err
			}
			if err := usages.SetAgentUsageIp(usage.Id, hash); err != nil {
				return hashed, err
			}
			hashed.Usages++
		}
		if len(page) < pageSize {
			break
		}
	}
	return hashed, nil
}

// RotateIpKey replaces the IP key with a new random one. The hashes made
// before it stay, but no new visit and no IP typed in a filter or in Find
// and erase matches them anymore: the way to forget the linkage early, after
// a server compromise say. It cannot be undone.
func RotateIpKey(sandbox *api.Sandbox) error {
	return privacy.RotateIpKey(sandbox)
}
