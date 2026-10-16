package privacyadmin

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/agent"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/contact"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// The privacy notice draft is the text the website's own privacy notice
// starts from, built from the settings in use: what is kept, why, for how
// long, the access log and its legal obligation, the host of the model
// provider, that no cookie is set and the chat history stays in the
// visitor's browser, the rights and the time to answer them. The controller,
// the contact and the encarregado or DPO are left as [placeholders], and the
// text says it is a draft to review. Its templates live in
// assets/privacy/notice-drafts: br.md (pt-BR) and eu.md (en); br-eu is both.
//
// A template holds {{name}} values and [[if name]] … [[end]] blocks, each on
// lines of their own or inside one line; a block is kept when its condition
// holds.

// draftsDir is where the templates of the draft live, under assets/.
const draftsDir = "privacy/notice-drafts/"

// NoticeDraft answers the privacy notice draft of the settings in use, in the
// template of privacy-setting.notice-draft.
func NoticeDraft(sandbox *api.Sandbox) (string, error) {
	settings, err := privacy.ReadSettings(sandbox)
	if err != nil {
		return "", err
	}
	values, conditions, err := draftValues(sandbox, settings)
	if err != nil {
		return "", err
	}
	parts := []string{privacy.DraftBr, privacy.DraftEu}
	switch settings.NoticeDraft {
	case privacy.DraftBr:
		parts = []string{privacy.DraftBr}
	case privacy.DraftEu:
		parts = []string{privacy.DraftEu}
	}
	texts := []string{}
	for _, part := range parts {
		source, err := sandbox.Deps.EmbedDeps.ReadFile(draftsDir + part + ".md")
		if err != nil {
			return "", sandbox.Deps.StdDeps.Errorf("no notice draft %s", part)
		}
		texts = append(texts, fillDraft(sandbox, string(source), values, conditions))
	}
	return sandbox.Deps.StringsDeps.Join(texts, "\n---\n\n"), nil
}

// draftValues are the values and the conditions a draft is filled with.
func draftValues(sandbox *api.Sandbox, settings privacy.Settings) (map[string]string, map[string]bool, error) {
	retentions, err := CurrentRetentions(sandbox)
	if err != nil {
		return nil, nil, err
	}
	contacts, err := contact.ReadSettings(sandbox)
	if err != nil {
		return nil, nil, err
	}
	agentSettings, err := agent.ReadSettings(sandbox)
	if err != nil {
		return nil, nil, err
	}
	host := agent.ProviderHost(sandbox, agentSettings)
	transfer := host != "" && !agent.ProviderIsLocal(sandbox, host)
	values := map[string]string{
		"ip_form_br":        "pseudonimizado",
		"ip_note_br":        ", um código que não revela o endereço",
		"ip_form_en":        "in pseudonymised form",
		"ip_note_en":        ", a code that does not reveal the address",
		"agent_host":        host,
		"agent_transfer_br": "",
		"agent_transfer_en": "",
	}
	if settings.IpMode == privacy.IpModeFull {
		values["ip_form_br"] = "completo"
		values["ip_form_en"] = "in full"
		values["ip_note_br"] = ""
		values["ip_note_en"] = ""
	}
	if transfer {
		values["agent_transfer_br"] = ", sem usar as conversas para treinar os seus modelos; [país do provedor e garantias da transferência internacional (LGPD, art. 33)]"
		values["agent_transfer_en"] = ", without using the conversations to train its models; [the provider's country and the safeguards of the transfer (GDPR arts. 44–49)]"
	}
	for _, retention := range []struct {
		name string
		days int64
	}{
		{"view_days", retentions.View},
		{"ip_days", retentions.Ip},
		{"contact_days", retentions.Contact},
		{"summary_days", retentions.Summary},
		{"access_days", retentions.AccessLog},
		{"backup_days", retentions.Backup},
	} {
		values[retention.name+"_br"] = daysBr(sandbox, retention.days)
		values[retention.name+"_en"] = daysEn(sandbox, retention.days)
	}
	conditions := map[string]bool{
		"contacts":   contacts.Enabled,
		"agent":      agentSettings.Enabled,
		"summaries":  agentSettings.Enabled && agentSettings.SummariesEnabled,
		"access_log": settings.AccessLog,
	}
	return values, conditions, nil
}

// daysBr spells a retention in Portuguese.
func daysBr(sandbox *api.Sandbox, days int64) string {
	switch days {
	case privacy.Forever:
		return "[prazo: hoje sem limite — defina um]"
	case 1:
		return "1 dia"
	}
	return sandbox.Deps.StringsDeps.FormatInt(days, 10) + " dias"
}

// daysEn spells a retention in English.
func daysEn(sandbox *api.Sandbox, days int64) string {
	switch days {
	case privacy.Forever:
		return "[period: no limit today — set one]"
	case 1:
		return "1 day"
	}
	return sandbox.Deps.StringsDeps.FormatInt(days, 10) + " days"
}

// fillDraft keeps the [[if name]] blocks of source whose condition holds,
// drops the others, and replaces every {{name}} with its value. A line left
// blank by a dropped block is dropped with it.
func fillDraft(sandbox *api.Sandbox, source string, values map[string]string, conditions map[string]bool) string {
	strings := sandbox.Deps.StringsDeps
	out := []byte{}
	keep := []bool{}
	kept := func() bool {
		for _, on := range keep {
			if !on {
				return false
			}
		}
		return true
	}
	for index := 0; index < len(source); {
		rest := source[index:]
		switch {
		case strings.HasPrefix(rest, "[[if "):
			end := len(strings.Split(rest, "]]")[0])
			name := strings.TrimSpace(rest[len("[[if "):end])
			keep = append(keep, conditions[name])
			index += end + 2
		case strings.HasPrefix(rest, "[[end]]"):
			if len(keep) > 0 {
				keep = keep[:len(keep)-1]
			}
			index += len("[[end]]")
		default:
			if kept() {
				out = append(out, source[index])
			}
			index++
		}
	}
	text := string(out)
	for name, value := range values {
		text = strings.ReplaceAll(text, "{{"+name+"}}", value)
	}
	lines := []string{}
	blank := false
	for _, line := range strings.Split(text, "\n") {
		empty := strings.TrimSpace(line) == ""
		if empty && blank {
			continue
		}
		blank = empty
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n")) + "\n"
}
