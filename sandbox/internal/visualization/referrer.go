package visualization

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// Referrer is the site a view was reached from: the host of its Referer
// header, lower case, without port ("google.com", "www.google.com"). It is ""
// for a direct visit, for a link inside the site itself — a Referer naming
// host, the host the request was sent to — and for anything that is not an
// http or https url.
func Referrer(sandbox *api.Sandbox, referer string, host string) string {
	strings := sandbox.Deps.StringsDeps
	text := strings.ToLower(strings.TrimSpace(referer))
	switch {
	case strings.HasPrefix(text, "https://"):
		text = text[len("https://"):]
	case strings.HasPrefix(text, "http://"):
		text = text[len("http://"):]
	default:
		return ""
	}
	for index := 0; index < len(text); index++ {
		if text[index] == '/' || text[index] == '?' || text[index] == '#' {
			text = text[:index]
			break
		}
	}
	if cut := strings.LastIndex(text, "@"); cut >= 0 {
		text = text[cut+1:]
	}
	referred := hostName(sandbox, text)
	if referred == "" || referred == hostName(sandbox, strings.ToLower(strings.TrimSpace(host))) {
		return ""
	}
	return truncate(referred, 255)
}

// hostName is authority — host, or host:port, or [ipv6]:port — without its
// port.
func hostName(sandbox *api.Sandbox, authority string) string {
	strings := sandbox.Deps.StringsDeps
	if strings.HasPrefix(authority, "[") {
		if end := strings.LastIndex(authority, "]"); end > 0 {
			return authority[1:end]
		}
		return ""
	}
	if strings.Count(authority, ":") == 1 {
		authority = authority[:strings.LastIndex(authority, ":")]
	}
	return strings.TrimSuffix(authority, ".")
}
