package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/databasedeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/agent_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/backoffice_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/contact_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/privacy_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/visualization_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// Kept is one table of personal data this server keeps, as the Privacy page
// lists it: the base of the records of processing (LGPD art. 37; GDPR art.
// 30), which the controller completes in their own documents. The legal
// basis is a suggestion, not a finding: the controller chooses it.
type Kept struct {
	Table     string
	Fields    string
	Purpose   string
	Basis     string
	Retention string
	Count     int
	// Oldest is the at of the oldest row, in Unix seconds; 0 for none, or
	// for a table whose rows have no time.
	Oldest int64
}

// KeptTables answers every table of personal data, with how many rows it
// holds and the oldest one.
func KeptTables(sandbox *api.Sandbox) ([]Kept, error) {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return nil, err
	}
	retentions, err := CurrentRetentions(sandbox)
	if err != nil {
		return nil, err
	}
	describe := func(days int64) string { return privacy.DescribeRetention(sandbox, days) }
	ipNote := "the ip " + map[bool]string{true: "in full", false: "hashed"}[settings.IpMode == privacy.IpModeFull] + ", blanked after " + describe(retentions.Ip)

	kept := []Kept{}
	add := func(row Kept, records []databasedeps.Record, err error, milliseconds bool) error {
		if err != nil {
			return err
		}
		row.Count = len(records)
		oldest, err := oldestAt(sandbox, records)
		if err != nil {
			return err
		}
		if milliseconds {
			oldest /= 1000
		}
		row.Oldest = oldest
		kept = append(kept, row)
		return nil
	}

	records, err := contact_db.ContactRequestRecords(sandbox, contact_db.New(sandbox))
	if err := add(Kept{
		Table:     "contact-request",
		Fields:    "what the visitor sent (names, emails, phones, messages), the time, the country; " + ipNote,
		Purpose:   "answer the visitor",
		Basis:     "steps the visitor asked for before a contract (LGPD art. 7 V; GDPR art. 6(1)(b)), or legitimate interest (LGPD art. 7 IX; GDPR art. 6(1)(f))",
		Retention: describe(retentions.Contact),
	}, records, err, false); err != nil {
		return nil, err
	}

	records, err = visualization_db.VisualizationViewRecords(sandbox, visualization_db.New(sandbox))
	if err := add(Kept{
		Table:     "visualization-view",
		Fields:    "the page, the time, the status, the country, the device, browser and os, the referring site" + map[bool]string{true: ", the full user agent", false: ""}[settings.IpMode == privacy.IpModeFull] + "; " + ipNote,
		Purpose:   "audience statistics, with no cookie",
		Basis:     "legitimate interest, with a balancing test (LGPD art. 10; GDPR art. 6(1)(f))",
		Retention: describe(retentions.View),
	}, records, err, false); err != nil {
		return nil, err
	}

	records, err = agent_db.AgentSummaryRecords(sandbox, agent_db.New(sandbox))
	if err := add(Kept{
		Table:     "agent-summary",
		Fields:    "a summary a model wrote of a conversation, with no ip; it can still hold personal data",
		Purpose:   "follow up what the visitor asked the agent",
		Basis:     "consent of the visitor who asked for it, or legitimate interest (LGPD art. 7 I, IX; GDPR art. 6(1)(a), (f))",
		Retention: describe(retentions.Summary),
	}, records, err, false); err != nil {
		return nil, err
	}

	records, err = agent_db.AgentUsageRecords(sandbox, agent_db.New(sandbox))
	if err := add(Kept{
		Table:     "agent-usage",
		Fields:    "the time, the kind of call and its tokens; " + ipNote,
		Purpose:   "the limits of the agent and its cost",
		Basis:     "legitimate interest in security and cost control (LGPD art. 10; GDPR art. 6(1)(f))",
		Retention: "90 days",
	}, records, err, false); err != nil {
		return nil, err
	}

	records, err = privacy_db.AccessLogRecords(sandbox, privacy_db.New(sandbox))
	accessLog := Kept{
		Table:     "access-log",
		Fields:    "the full ip, the source port, the time in UTC, the kind of request and the record it wrote",
		Purpose:   "the access logs of Marco Civil art. 15, disclosed only on a court order",
		Basis:     "legal obligation (LGPD art. 7 II); in the EU, legitimate interest in security (GDPR art. 6(1)(f))",
		Retention: describe(retentions.AccessLog),
	}
	if !settings.AccessLog {
		accessLog.Purpose += " — off: no new row is written"
	}
	if err := add(accessLog, records, err, true); err != nil {
		return nil, err
	}

	records, err = privacy_db.PrivacyAccessRecords(sandbox, privacy_db.New(sandbox))
	if err := add(Kept{
		Table:     "privacy-access",
		Fields:    "who read the access log, when, for how long, what and with which filter",
		Purpose:   "the inventory of accesses of Decreto 8.771/2016 art. 13 III",
		Basis:     "legal obligation (LGPD art. 7 II)",
		Retention: "kept until a root deletes it",
	}, records, err, true); err != nil {
		return nil, err
	}

	users, err := backoffice_db.New(sandbox).CountBackofficeUser()
	if err != nil {
		return nil, err
	}
	kept = append(kept, Kept{
		Table:     "backoffice-user, api-token",
		Fields:    "the staff's usernames, emails, password hashes, two-factor secrets, sessions, and the ips their tokens were last used from",
		Purpose:   "sign-in and the security of the backoffice",
		Basis:     "the employment or service contract, and legitimate interest in security (LGPD art. 7 V, IX; GDPR art. 6(1)(b), (f))",
		Retention: "while the account exists",
		Count:     users,
	})
	return kept, nil
}

// oldestAt is the at of the oldest of records, 0 when there is none.
func oldestAt(sandbox *api.Sandbox, records []databasedeps.Record) (int64, error) {
	oldest := int64(0)
	for _, record := range records {
		at, err := sandbox.Deps.OpinionatedAgnosDatabase.ReadInt(record, "at")
		if err != nil {
			return 0, err
		}
		if at > 0 && (oldest == 0 || at < oldest) {
			oldest = at
		}
	}
	return oldest, nil
}
