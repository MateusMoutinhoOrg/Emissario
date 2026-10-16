package privacy

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
)

// In the hashed mode — the default — the records hold an IP as HashPrefix and
// the hex of the first 16 bytes of HMAC-SHA256(ip-key, ip): IPv4 whole, IPv6
// by its /64, the grouping of the rate limits, so a visitor cannot dodge a
// match by rotating inside their /64. The key is 32 random bytes kept in
// privacy-setting, encrypted by the store like every value and never derived
// from EMISSARIO_ENCRYPT_KEY, so rotating that key keeps every hash. The hash
// is deterministic — it counts visitors over any period and matches an abuser
// later — and cannot be reversed without the key; whoever holds the key can
// still confirm one IP, so it stays personal data, blanked when the IP
// retention ends.
//
// What needs the raw IP — the country, the ignored IPs, the rate limits and
// the access log — runs before it is hashed, in memory.

// HashPrefix starts every hashed IP a record holds.
const HashPrefix = "h:"

// hashHexLength is how many hex digits of the HMAC a hash keeps: 16 bytes.
const hashHexLength = 32

// shownHashLength is how many hex digits of a hash a page shows.
const shownHashLength = 4

// GroupKey is the key ip counts and hashes under: an IPv4 ip whole, an IPv6
// one its /64 (2001:db8:1:2::/64). Something that is not an ip is itself.
func GroupKey(sandbox *api.Sandbox, ip string) string {
	address, ok := backofficehttp.ParseIp(sandbox, ip)
	if !ok {
		return ip
	}
	full := backofficehttp.FormatIp(sandbox, address)
	if !sandbox.Deps.StringsDeps.Contains(full, ":") {
		return full
	}
	for index := 8; index < 16; index++ {
		address[index] = 0
	}
	return backofficehttp.FormatIp(sandbox, address) + "/64"
}

// IsHash tells whether stored is a hashed IP.
func IsHash(sandbox *api.Sandbox, stored string) bool {
	return sandbox.Deps.StringsDeps.HasPrefix(stored, HashPrefix)
}

// HashIp is the hash of ip under the current IP key.
func HashIp(sandbox *api.Sandbox, ip string) (string, error) {
	key, err := IpKey(sandbox)
	if err != nil {
		return "", err
	}
	return hashWith(sandbox, key, ip), nil
}

// hashWith is the hash of ip under key.
func hashWith(sandbox *api.Sandbox, key []byte, ip string) string {
	sum := sandbox.Deps.HashDeps.HmacSha256Hex(key, []byte(GroupKey(sandbox, ip)))
	return HashPrefix + sum[:hashHexLength]
}

// StoredIp is what a record keeps of ip under the IP mode of settings: its
// hash in the hashed mode, the ip as it came in the full one; "" for no ip.
func StoredIp(sandbox *api.Sandbox, settings Settings, ip string) (string, error) {
	if ip == "" || IsHash(sandbox, ip) {
		return ip, nil
	}
	if settings.IpMode == IpModeFull {
		return ip, nil
	}
	return HashIp(sandbox, ip)
}

// Store is StoredIp under the current settings.
func Store(sandbox *api.Sandbox, ip string) (string, error) {
	settings, err := CurrentSettings(sandbox)
	if err != nil {
		return "", err
	}
	return StoredIp(sandbox, settings, ip)
}

// TallyKey is the key the tokens of a visitor at ip count under, the same
// one LoadKey reads back off a stored row: the hash in the hashed mode, the
// GroupKey in the full one. A store that fails falls back to the GroupKey.
func TallyKey(sandbox *api.Sandbox, ip string) string {
	if ip == "" {
		return ""
	}
	stored, err := Store(sandbox, ip)
	if err != nil || !IsHash(sandbox, stored) {
		return GroupKey(sandbox, ip)
	}
	return stored
}

// StoredKey is the key a stored ip counts under: a hash as it is, an ip its
// GroupKey.
func StoredKey(sandbox *api.Sandbox, stored string) string {
	if IsHash(sandbox, stored) {
		return stored
	}
	return GroupKey(sandbox, stored)
}

// MaskIp is ip as a viewer reads it: an IPv4 one without its last byte
// (203.0.113.xxx), an IPv6 one with its first 48 bits alone
// (2001:db8:1:xxxx:…).
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

// ShowHash is a hash as every page shows it: its prefix and first digits
// (h:3fa9…).
func ShowHash(sandbox *api.Sandbox, stored string) string {
	if len(stored) <= len(HashPrefix)+shownHashLength {
		return stored
	}
	return stored[:len(HashPrefix)+shownHashLength] + "…"
}

// ShowIp is stored — a hash, an ip, or a GroupKey ending in /64 — as the
// reader reads it: a hash shortened for everyone; an ip whole to a root, and
// to a viewer when the settings show viewers full IPs, masked otherwise.
func ShowIp(sandbox *api.Sandbox, stored string, root bool) string {
	if stored == "" {
		return ""
	}
	if IsHash(sandbox, stored) {
		return ShowHash(sandbox, stored)
	}
	if root {
		return stored
	}
	if settings, err := CurrentSettings(sandbox); err == nil && settings.ViewerIp == ViewerIpFull {
		return stored
	}
	strings := sandbox.Deps.StringsDeps
	return MaskIp(sandbox, strings.TrimSuffix(stored, "/64"))
}

// IpFilter is an IP filter a root typed, ready to match stored IPs: a hash
// exactly, one ip by its hash and as it is, or — with full IPs only — a
// CIDR range.
type IpFilter struct {
	// Text is the filter in its canonical spelling.
	Text string
	// Range is true for a CIDR range.
	Range bool
	// Hash is the hash the filter matches: the one typed, or the hash of the
	// ip typed under the current IP key.
	Hash string
}

// ParseIpFilter reads text — a hash, an ip or a CIDR range — refusing a range
// in the hashed mode, where a range cannot be hashed.
func ParseIpFilter(sandbox *api.Sandbox, field string, text string) (IpFilter, *Refusal) {
	strings := sandbox.Deps.StringsDeps
	text = strings.TrimSpace(text)
	if IsHash(sandbox, text) {
		if len(text) != len(HashPrefix)+hashHexLength {
			return IpFilter{}, refuse(field, "A hashed IP is h: followed by 32 hex digits.")
		}
		return IpFilter{Text: strings.ToLower(text), Hash: strings.ToLower(text)}, nil
	}
	canonical, ok := backofficehttp.CanonicalIpOrRange(sandbox, text)
	if !ok {
		return IpFilter{}, refuse(field, "ip is one ip (203.0.113.7), one CIDR range (203.0.113.0/24) with full IPs, or one hashed IP (h:…).")
	}
	filter := IpFilter{Text: canonical, Range: strings.Contains(canonical, "/")}
	settings, err := CurrentSettings(sandbox)
	if err != nil {
		return IpFilter{}, &Refusal{Status: api.StatusInternalServerError, Message: "the privacy settings could not be read"}
	}
	if filter.Range {
		if settings.IpMode == IpModeHashed {
			return IpFilter{}, refuse(field, "A CIDR range cannot be hashed: in the hashed IP mode filter by one ip, which is hashed to match.")
		}
		return filter, nil
	}
	hash, err := HashIp(sandbox, canonical)
	if err != nil {
		return IpFilter{}, &Refusal{Status: api.StatusInternalServerError, Message: "the IP key could not be read"}
	}
	filter.Hash = hash
	return filter, nil
}

// Matches tells whether stored — a hash or an ip — is one filter keeps.
func (filter IpFilter) Matches(sandbox *api.Sandbox, stored string) bool {
	if stored == "" || filter.Text == "" {
		return false
	}
	if IsHash(sandbox, stored) {
		return filter.Hash != "" && stored == filter.Hash
	}
	if IsHash(sandbox, filter.Text) {
		return false
	}
	return backofficehttp.IpMatches(sandbox, filter.Text, stored)
}
