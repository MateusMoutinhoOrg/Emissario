package privacy

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// Neither the LGPD nor the GDPR sets a number of days: both ask for "no
// longer than necessary" (LGPD arts. 15–16; GDPR art. 5(1)(e)). The numbers
// below come from regulators' references and from other laws. Every retention
// is a number of days from 1 up, or Forever; none is refused for being long,
// but one past its warning asks for the retention notice of its rule, which
// lists the reference that was crossed.

// Forever is a retention that keeps the data until a root deletes it: a
// blank field on a form, null in JSON.
const Forever int64 = 0

// MaxRetentionDays bounds a retention so its arithmetic never overflows: a
// century, which no warning lets pass unnoticed.
const MaxRetentionDays int64 = 36500

// The data a retention applies to.
const (
	DataContact   = "contact"
	DataAccessLog = "access-log"
	DataIp        = "ip"
	DataView      = "view"
	DataSummary   = "summary"
	DataBackup    = "backup"
)

// RetentionRule is one row of the retention table: the data it applies to,
// how it is named, its default, the bounds past which it asks for the
// retention notice — WarnBelow 0 for none, WarnAbove 0 for the longest of the
// other retentions (backups) — and the references behind them.
type RetentionRule struct {
	Data      string
	Label     string
	Default   int64
	WarnBelow int64
	WarnAbove int64
	Reference string
}

// RetentionRules is the retention table of docs/Privacy, in the order the
// Privacy page lists it.
var RetentionRules = []RetentionRule{
	{
		Data:      DataContact,
		Label:     "contact requests",
		Default:   180,
		WarnAbove: 365,
		Reference: "No law sets a number of days: keep a request while you answer it and while the relationship lasts (LGPD arts. 15–16; GDPR art. 5(1)(e)). The CNIL keeps prospects' data at most 3 years after the last contact. Five years, the Brazilian consumer limitation period (CDC art. 27), is the outer bound for a legal defence (LGPD art. 7 VI).",
	},
	{
		Data:      DataAccessLog,
		Label:     "access log rows",
		Default:   184,
		WarnBelow: 184,
		WarnAbove: 365,
		Reference: "Marco Civil art. 15 asks a for-profit company in Brazil to keep access logs at least 6 months, then delete them (Decreto 8.771/2016 art. 13 §2); 184 days is the longest 6-month span. The CNIL (rec. 2021-122) keeps security logs 6 months to 1 year, up to 3 years only with a documented need.",
	},
	{
		Data:      DataIp,
		Label:     "IPs in the records, hashed or full",
		Default:   184,
		WarnAbove: 365,
		Reference: "The CNIL (rec. 2021-122) keeps security logs 6 months to 1 year. A hashed IP is still personal data — pseudonymous, since this server holds the key (GDPR art. 4(5), recital 26; CJEU C-413/23 P SRB) — so it has a limit like any other.",
	},
	{
		Data:      DataView,
		Label:     "page views",
		Default:   90,
		WarnAbove: 760,
		Reference: "The CNIL keeps audience measurement statistics at most 25 months.",
	},
	{
		Data:      DataSummary,
		Label:     "agent summaries",
		Default:   90,
		WarnAbove: 365,
		Reference: "As contact requests: no law sets a number of days, keep them only while needed (LGPD arts. 15–16; GDPR art. 5(1)(e)).",
	},
	{
		Data:      DataBackup,
		Label:     "backup snapshots",
		Default:   Forever,
		Reference: "Data a retention removes from the databases lives on in the snapshots taken before: a snapshot kept longer than the longest retention above keeps it past that retention.",
	},
}

// Rule is the rule of data.
func Rule(data string) RetentionRule {
	for _, rule := range RetentionRules {
		if rule.Data == data {
			return rule
		}
	}
	return RetentionRule{Data: data, Label: data}
}

// RetentionNoticeId is the id of the retention notice of data.
func RetentionNoticeId(data string) string {
	return "retention-" + data + "-v1"
}

// Warns tells whether days — Forever included — is past the warning of
// rule. longest is the longest of the other retentions, which bounds a
// backup's.
func (rule RetentionRule) Warns(days int64, longest int64) bool {
	if days == Forever {
		return true
	}
	if rule.WarnBelow > 0 && days < rule.WarnBelow {
		return true
	}
	above := rule.WarnAbove
	if above == 0 {
		above = longest
	}
	return above > 0 && days > above
}

// CheckRetention refuses a retention of days that is not one: below 1 and not
// Forever, or above MaxRetentionDays.
func CheckRetention(sandbox *api.Sandbox, field string, days int64) *Refusal {
	if days == Forever {
		return nil
	}
	if days < 1 || days > MaxRetentionDays {
		return refuse(field, "A retention is a number of days from 1 to 36500, or blank to keep the data until a root deletes it.")
	}
	return nil
}

// ParseRetention reads a retention as a form sends it: blank is Forever, and
// a number of days from 1 up anything else; 0, a negative number or text is
// refused.
func ParseRetention(sandbox *api.Sandbox, field string, text string) (int64, *Refusal) {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if text == "" {
		return Forever, nil
	}
	days, err := sandbox.Deps.StringsDeps.ParseInt(text, 10, 64)
	if err != nil || days < 1 {
		return 0, refuse(field, "A retention is a number of days from 1 to 36500, or blank to keep the data until a root deletes it.")
	}
	return days, CheckRetention(sandbox, field, days)
}

// FormatRetention spells a retention for a form: the number of days, or
// blank for Forever.
func FormatRetention(sandbox *api.Sandbox, days int64) string {
	if days == Forever {
		return ""
	}
	return sandbox.Deps.StringsDeps.FormatInt(days, 10)
}

// DescribeRetention spells a retention for a page: "180 days", or "kept
// until a root deletes it".
func DescribeRetention(sandbox *api.Sandbox, days int64) string {
	if days == Forever {
		return "kept until a root deletes it"
	}
	if days == 1 {
		return "1 day"
	}
	return sandbox.Deps.StringsDeps.FormatInt(days, 10) + " days"
}

// RetentionNotice adds to notices the retention notice of data when a
// retention changes from before to after and after is past its warning.
func RetentionNotice(notices []string, data string, before int64, after int64, longest int64) []string {
	if before == after || !Rule(data).Warns(after, longest) {
		return notices
	}
	return append(notices, RetentionNoticeId(data))
}
