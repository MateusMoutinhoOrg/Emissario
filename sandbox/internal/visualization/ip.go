package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
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
// (2001:db8:1:xxxx:…). A root reads every ip whole.
func MaskIp(sandbox *api.Sandbox, ip string) string {
	address, ok := backofficehttp.ParseIp(sandbox, ip)
	if !ok {
		return "xxx"
	}
	strings := sandbox.Deps.StringsDeps
	full := backofficehttp.FormatIp(sandbox, address)
	if !strings.Contains(full, ":") {
		return full[:strings.LastIndex(full, ".")] + ".xxx"
	}
	groups := []string{}
	for index := 0; index < 3; index++ {
		group := int64(address[2*index])<<8 | int64(address[2*index+1])
		groups = append(groups, strings.FormatInt(group, 16))
	}
	return strings.Join(groups, ":") + ":xxxx:…"
}

// ShownIp is ip as the reader reads it: whole to a root, masked to a viewer.
func ShownIp(sandbox *api.Sandbox, ip string, root bool) string {
	if root {
		return ip
	}
	return MaskIp(sandbox, ip)
}
