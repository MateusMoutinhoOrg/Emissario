package routeprops

// Privacy is what the privacy mechanic adds to every request: the source
// port of the visitor's connection, which the access log keeps beside the
// ip — behind carrier-grade NAT many customers share one IPv4 address, and
// only the port tells them apart.
type Privacy struct {
	// ClientPort is the source port of the visitor's connection — the
	// connection's own, or, with --client-port-header, the one the reverse
	// proxy in front set — "" when none reads as a port.
	ClientPort string
}
