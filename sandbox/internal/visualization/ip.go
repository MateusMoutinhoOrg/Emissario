package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
)

// unroutedRanges are the ips no country is ever looked up for: private,
// shared, loopback, link-local, multicast and unspecified addresses. No
// country database places them, and none of them is a visitor's own.
var unroutedRanges = []string{
	"0.0.0.0/8", "10.0.0.0/8", "100.64.0.0/10", "127.0.0.0/8", "169.254.0.0/16", "172.16.0.0/12",
	"192.168.0.0/16", "224.0.0.0/4", "240.0.0.0/4",
	"::/128", "::1/128", "fc00::/7", "fe80::/10", "ff00::/8",
}

// Unrouted tells whether ip is not an ip, or is one of unroutedRanges.
func Unrouted(sandbox *api.Sandbox, ip string) bool {
	if _, ok := backofficehttp.ParseIp(sandbox, ip); !ok {
		return true
	}
	for _, network := range unroutedRanges {
		if backofficehttp.IpMatches(sandbox, network, ip) {
			return true
		}
	}
	return false
}

// MaskIp is ip as a viewer reads it: an IPv4 one without its last byte
// (203.0.113.xxx), an IPv6 one with its first 48 bits alone
// (2001:db8:1:xxxx:…).
func MaskIp(sandbox *api.Sandbox, ip string) string {
	return privacy.MaskIp(sandbox, ip)
}

// ShownIp is ip — an ip or a hash, as a view stores it — as the reader gets
// it from a listing: as stored to a root, who filters by it, and as
// privacy.ShowIp shows it to a viewer — a hash shortened, an ip masked unless
// the settings show viewers full IPs. A page shows a root's through
// privacy.ShowIp too.
func ShownIp(sandbox *api.Sandbox, ip string, root bool) string {
	if root {
		return ip
	}
	return privacy.ShowIp(sandbox, ip, false)
}
