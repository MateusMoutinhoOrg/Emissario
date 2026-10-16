package privacy

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
)

// A notice is the text a root reads before a risky choice: it names the law
// and the obligations the choice brings. The backoffice runs no inline
// script, so the server renders it: a page that needs one comes back under a
// 409 with the notice inside its form and a required checkbox, and its JSON
// twin answers 409 {"error":"acknowledgement-required","field":"acknowledge",
// "notices":[…]} until the request sends the ids in "acknowledge". The server
// keeps nothing about it — not who confirmed, nor when, nor that it was shown
// — so the notice comes back every time the setting is turned to that value
// again, and at every use of an action.
//
// A notice lives in assets/privacy/notices/<id>.md, versioned in its id
// (contacts-v1): its first line is "# Title", then paragraphs and "- " lists
// whose text may hold **bold**, *italic* and `code`. The retention notices are
// generated from RetentionRules.

// The notices of the settings and actions.
const (
	NoticeContacts          = "contacts-v1"
	NoticeAccessLog         = "access-log-v1"
	NoticeIpFull            = "ip-full-v1"
	NoticeViewerIp          = "viewer-ip-v1"
	NoticeAccessLogNoTotp   = "access-log-no-totp-v1"
	NoticeAccessLogNoInvent = "access-log-no-inventory-v1"
	NoticeEraseKept         = "erase-kept-v1"
	NoticeSummaries         = "summaries-v1"
	NoticeAgent             = "agent-v1"
)

// AcknowledgementRequired is the error of the 409 a request that has not
// acknowledged its notices answers.
const AcknowledgementRequired = "acknowledgement-required"

// AcknowledgeField is the field a request acknowledges notices in.
const AcknowledgeField = "acknowledge"

// Span is a run of text of one style.
type Span struct {
	Text   string
	Bold   bool
	Italic bool
	Code   bool
}

// Block is a paragraph — its spans — or a list — one run of spans per item.
type Block struct {
	List  bool
	Spans []Span
	Items [][]Span
}

// Notice is one notice ready to show: its id, its title, its text as blocks
// for a page and as plain text for JSON.
type Notice struct {
	Id     string
	Title  string
	Blocks []Block
	Text   string
}

// NoticeContext is what a notice's text is filled with: the host of the
// model provider, and the retention a retention notice is about with the
// longest of the others.
type NoticeContext struct {
	Host      string
	Retention int64
	Longest   int64
}

// LoadNotices answers the notices of ids, in that order, each once.
func LoadNotices(sandbox *api.Sandbox, ids []string, context NoticeContext) ([]Notice, error) {
	notices := []Notice{}
	seen := map[string]bool{}
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		notice, err := LoadNotice(sandbox, id, context)
		if err != nil {
			return nil, err
		}
		notices = append(notices, notice)
	}
	return notices, nil
}

// LoadNotice answers the notice of id.
func LoadNotice(sandbox *api.Sandbox, id string, context NoticeContext) (Notice, error) {
	strings := sandbox.Deps.StringsDeps
	var source string
	if strings.HasPrefix(id, "retention-") && strings.HasSuffix(id, "-v1") {
		data := strings.TrimSuffix(strings.TrimPrefix(id, "retention-"), "-v1")
		source = retentionNoticeSource(sandbox, Rule(data), context)
	} else {
		content, err := sandbox.Deps.EmbedDeps.ReadFile("privacy/notices/" + id + ".md")
		if err != nil {
			return Notice{}, sandbox.Deps.StdDeps.Errorf("no notice %s", id)
		}
		source = string(content)
	}
	host := context.Host
	if host == "" {
		host = "the model provider"
	}
	source = strings.ReplaceAll(source, "<host>", host)
	return parseNotice(sandbox, id, source), nil
}

// retentionNoticeSource is the text of the retention notice of rule, for the
// retention of context.
func retentionNoticeSource(sandbox *api.Sandbox, rule RetentionRule, context NoticeContext) string {
	format := sandbox.Deps.StringsDeps.FormatInt
	text := "# Retention of " + rule.Label + "\n\n"
	above := rule.WarnAbove
	if above == 0 {
		above = context.Longest
	}
	switch {
	case context.Retention == Forever:
		text += "You keep " + rule.Label + " **until a root deletes them**. The laws ask you to keep personal data no longer than needed (LGPD arts. 15–16; GDPR art. 5(1)(e)): keeping it with no limit needs a documented reason.\n\n"
	case rule.WarnBelow > 0 && context.Retention < rule.WarnBelow:
		text += "You set **" + format(context.Retention, 10) + " days**, below " + format(rule.WarnBelow, 10) + " days.\n\n"
	default:
		text += "You set **" + format(context.Retention, 10) + " days**. Above " + format(above, 10) + " days, the references below ask for a documented reason.\n\n"
	}
	return text + rule.Reference + "\n"
}

// Missing is every id of required that acknowledged does not hold, each
// once, in the order of required.
func Missing(required []string, acknowledged []string) []string {
	given := map[string]bool{}
	for _, id := range acknowledged {
		given[id] = true
	}
	missing := []string{}
	seen := map[string]bool{}
	for _, id := range required {
		if given[id] || seen[id] {
			continue
		}
		seen[id] = true
		missing = append(missing, id)
	}
	return missing
}

// NeedsAcknowledgement is the 409 refusal of a request that has not
// acknowledged every one of required, nil when it has.
func NeedsAcknowledgement(required []string, acknowledged []string) *Refusal {
	missing := Missing(required, acknowledged)
	if len(missing) == 0 {
		return nil
	}
	return &Refusal{Status: api.StatusConflict, Field: AcknowledgeField, Message: AcknowledgementRequired, Notices: missing}
}

// parseNotice reads source — "# Title", then paragraphs and "- " lists —
// into a Notice.
func parseNotice(sandbox *api.Sandbox, id string, source string) Notice {
	strings := sandbox.Deps.StringsDeps
	notice := Notice{Id: id}
	plain := []string{}
	var paragraph []string
	var items []string
	flush := func() {
		if len(paragraph) > 0 {
			text := strings.Join(paragraph, " ")
			notice.Blocks = append(notice.Blocks, Block{Spans: parseSpans(text)})
			plain = append(plain, stripMarks(text))
			paragraph = nil
		}
		if len(items) > 0 {
			block := Block{List: true}
			for _, item := range items {
				block.Items = append(block.Items, parseSpans(item))
				plain = append(plain, "- "+stripMarks(item))
			}
			notice.Blocks = append(notice.Blocks, block)
			items = nil
		}
	}
	for _, raw := range strings.Split(source, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case line == "":
			flush()
		case strings.HasPrefix(line, "# ") && notice.Title == "":
			notice.Title = strings.TrimSpace(line[2:])
		case strings.HasPrefix(line, "- "):
			if len(paragraph) > 0 {
				flush()
			}
			items = append(items, strings.TrimSpace(line[2:]))
		case len(items) > 0:
			items[len(items)-1] += " " + line
		default:
			paragraph = append(paragraph, line)
		}
	}
	flush()
	notice.Text = strings.Join(plain, "\n")
	return notice
}

// parseSpans splits text at its **, * and ` marks into styled runs. A mark
// left open runs to the end of the text.
func parseSpans(text string) []Span {
	spans := []Span{}
	current := Span{}
	run := []byte{}
	emit := func() {
		if len(run) > 0 {
			current.Text = string(run)
			spans = append(spans, current)
			run = []byte{}
		}
	}
	for index := 0; index < len(text); index++ {
		character := text[index]
		switch {
		case character == '`':
			emit()
			current.Code = !current.Code
		case current.Code:
			run = append(run, character)
		case character == '*' && index+1 < len(text) && text[index+1] == '*':
			emit()
			current.Bold = !current.Bold
			index++
		case character == '*':
			emit()
			current.Italic = !current.Italic
		default:
			run = append(run, character)
		}
	}
	emit()
	return spans
}

// stripMarks is text without its **, * and ` marks.
func stripMarks(text string) string {
	out := []byte{}
	for index := 0; index < len(text); index++ {
		if text[index] != '*' && text[index] != '`' {
			out = append(out, text[index])
		}
	}
	return string(out)
}
