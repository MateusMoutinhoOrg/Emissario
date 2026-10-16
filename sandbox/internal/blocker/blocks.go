package blocker

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/blocker_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// Input is what a form or a JSON body sends to add a block: the value — an
// ip, a CIDR range, a hashed ip or a country code —, why, and for how many
// days, 0 for until a root removes it.
type Input struct {
	Value  string
	Reason string
	Days   int64
}

// writeLock is held while a block is added or removed, so two never take one
// value nor pass MaxBlocks together.
var writeLock = make(chan struct{}, 1)

// List is every block, the newest first.
func List(sandbox *api.Sandbox) ([]blocker_db.IpBlockRecord, error) {
	blocks, err := blocker_db.New(sandbox).ListIpBlocks(blocker_db.IpBlockFilter{})
	if err != nil {
		return nil, err
	}
	sandbox.Deps.SortDeps.Slice(blocks, func(i int, j int) bool {
		if blocks[i].CreatedAt != blocks[j].CreatedAt {
			return blocks[i].CreatedAt > blocks[j].CreatedAt
		}
		return blocks[i].Id > blocks[j].Id
	})
	return blocks, nil
}

// Expired tells whether block ended, at now.
func Expired(block blocker_db.IpBlockRecord, now int64) bool {
	return block.ExpiresAt > 0 && block.ExpiresAt <= now
}

// Stale tells whether block is a hashed ip made under an IP key rotated
// since: it matches no one anymore.
func Stale(sandbox *api.Sandbox, block blocker_db.IpBlockRecord, rotatedAt int64) bool {
	return block.Kind == KindHash && rotatedAt > 0 && block.CreatedAt < rotatedAt
}

// Parse reads text as the value of a block, in its canonical spelling: two
// letters are a country code, h: and 32 hex digits a hashed ip, and anything
// else an ip or a CIDR range. An IPv6 address is blocked by its /64, the
// grouping of the hashes and the rate limits, so a visitor cannot dodge it by
// rotating inside their /64.
func Parse(sandbox *api.Sandbox, text string) (kind string, value string, refusal *Refusal) {
	strings := sandbox.Deps.StringsDeps
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", refuse("value", "Type an ip, a CIDR range, a hashed ip or a country code.")
	}
	if len(text) == 2 {
		code := strings.ToUpper(text)
		if code == visualization.UnknownCountry || !visualization.IsKnownCountry(sandbox, code) {
			return "", "", refuse("value", "This is no country code: a country is its ISO 3166-1 alpha-2 code, such as BR.")
		}
		return KindCountry, code, nil
	}
	if privacy.IsHash(sandbox, strings.ToLower(text)) {
		hash := strings.ToLower(text)
		if !isHash(hash) {
			return "", "", refuse("value", "A hashed IP is h: followed by 32 hex digits.")
		}
		return KindHash, hash, nil
	}
	canonical, ok := backofficehttp.CanonicalIpOrRange(sandbox, text)
	if !ok {
		return "", "", refuse("value", "Type one ip (203.0.113.7), one CIDR range (203.0.113.0/24), one hashed IP (h:…) or one country code (BR).")
	}
	if strings.Contains(canonical, "/") {
		return KindRange, canonical, nil
	}
	if strings.Contains(canonical, ":") {
		return KindRange, privacy.GroupKey(sandbox, canonical), nil
	}
	return KindIp, canonical, nil
}

// isHash tells whether text is privacy.HashPrefix and 32 lower-case hex
// digits.
func isHash(text string) bool {
	digits := text[len(privacy.HashPrefix):]
	if len(digits) != 32 {
		return false
	}
	for index := 0; index < len(digits); index++ {
		char := digits[index]
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

// Add checks input and adds it as a block, refused when it does not hold,
// when the value is blocked already, when MaxBlocks are there, or when it
// matches clientIp, the ip the root adding it came from: they would block
// themselves.
func Add(sandbox *api.Sandbox, input Input, clientIp string) (blocker_db.IpBlockRecord, *Refusal, error) {
	kind, value, refusal := Parse(sandbox, input.Value)
	if refusal != nil {
		return blocker_db.IpBlockRecord{}, refusal, nil
	}
	reason := sandbox.Deps.StringsDeps.TrimSpace(input.Reason)
	if len([]rune(reason)) > MaxReasonLength {
		return blocker_db.IpBlockRecord{}, refuse("reason", "The reason is at most 200 characters."), nil
	}
	if input.Days < 0 || input.Days > MaxDays {
		return blocker_db.IpBlockRecord{}, refuse("days", "The days are a number from 1 to 36500, or blank to block until a root removes it."), nil
	}
	if Matches(sandbox, kind, value, clientIp) {
		return blocker_db.IpBlockRecord{}, &Refusal{Status: api.StatusConflict, Field: "value", Message: "This block matches the IP you are connected from: you would block yourself."}, nil
	}

	writeLock <- struct{}{}
	defer func() { <-writeLock }()
	db := blocker_db.New(sandbox)
	if _, taken := db.FindIpBlockByValue(value); taken {
		return blocker_db.IpBlockRecord{}, &Refusal{Status: api.StatusConflict, Field: "value", Message: "This value is blocked already."}, nil
	}
	count, err := db.CountIpBlock()
	if err != nil {
		return blocker_db.IpBlockRecord{}, nil, err
	}
	if count >= MaxBlocks {
		return blocker_db.IpBlockRecord{}, refuse("value", "There are 10000 blocks already: remove one first."), nil
	}
	now := NowSeconds(sandbox)
	expiresAt := int64(0)
	if input.Days > 0 {
		expiresAt = now + input.Days*daySeconds
	}
	block, err := db.AddIpBlock(blocker_db.IpBlockInput{Value: value, Kind: kind, Reason: reason, CreatedAt: now, ExpiresAt: expiresAt})
	Invalidate(sandbox)
	return block, nil, err
}

// Remove removes the blocks of ids and answers how many it removed; ids no
// block holds are skipped, and a 404 answers when none of them did.
func Remove(sandbox *api.Sandbox, ids []int64) (int, *Refusal, error) {
	if len(ids) == 0 {
		return 0, refuse("id", "Select at least one block."), nil
	}
	if len(ids) > MaxIds {
		return 0, refuse("id", "Remove at most 100 blocks at once."), nil
	}
	writeLock <- struct{}{}
	defer func() { <-writeLock }()
	defer Invalidate(sandbox)
	db := blocker_db.New(sandbox)
	removed := 0
	for _, id := range ids {
		if _, ok := db.FindIpBlockById(id); !ok {
			continue
		}
		if err := db.RemoveIpBlock(id); err != nil {
			return removed, nil, err
		}
		removed++
	}
	if removed == 0 {
		return 0, &Refusal{Status: api.StatusNotFound, Field: "id", Message: "No block has this id."}, nil
	}
	return removed, nil, nil
}

// RemoveValue removes the block of value, typed as Parse reads it; false when
// no block holds it.
func RemoveValue(sandbox *api.Sandbox, text string) (bool, *Refusal, error) {
	_, value, refusal := Parse(sandbox, text)
	if refusal != nil {
		return false, refusal, nil
	}
	writeLock <- struct{}{}
	defer func() { <-writeLock }()
	db := blocker_db.New(sandbox)
	block, ok := db.FindIpBlockByValue(value)
	if !ok {
		return false, nil, nil
	}
	defer Invalidate(sandbox)
	return true, nil, db.RemoveIpBlock(block.Id)
}

// RemoveAll removes every block and answers how many there were.
func RemoveAll(sandbox *api.Sandbox) (int, error) {
	return removeWhere(sandbox, func(block blocker_db.IpBlockRecord) bool { return true })
}

// RemoveExpired removes every block that ended and answers how many: the
// retention cleanup calls it.
func RemoveExpired(sandbox *api.Sandbox) (int, error) {
	now := NowSeconds(sandbox)
	return removeWhere(sandbox, func(block blocker_db.IpBlockRecord) bool { return Expired(block, now) })
}

// removeWhere removes every block remove holds for and answers how many.
func removeWhere(sandbox *api.Sandbox, remove func(block blocker_db.IpBlockRecord) bool) (int, error) {
	writeLock <- struct{}{}
	defer func() { <-writeLock }()
	db := blocker_db.New(sandbox)
	blocks, err := db.ListIpBlocks(blocker_db.IpBlockFilter{})
	if err != nil {
		return 0, err
	}
	removed := 0
	defer func() {
		if removed > 0 {
			Invalidate(sandbox)
		}
	}()
	for _, block := range blocks {
		if !remove(block) {
			continue
		}
		if err := db.RemoveIpBlock(block.Id); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

// ParseDays reads the days of a form: blank is 0, until a root removes it.
func ParseDays(sandbox *api.Sandbox, text string) (int64, *Refusal) {
	text = sandbox.Deps.StringsDeps.TrimSpace(text)
	if text == "" {
		return 0, nil
	}
	days, err := sandbox.Deps.StringsDeps.ParseInt(text, 10, 64)
	if err != nil || days < 1 || days > MaxDays {
		return 0, refuse("days", "The days are a number from 1 to 36500, or blank to block until a root removes it.")
	}
	return days, nil
}
