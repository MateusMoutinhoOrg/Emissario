package backofficeapi

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	serializabledeps "github.com/MateusMoutinhoOrg/Emissario/sandbox/deps/serializabledeps"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/blocker"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/blocker_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
)

// IpBlockJSON is one block: {id, value, kind, reason, created_at, expires_at,
// expired, stale, hits}. expires_at is 0 for a block until a root removes it;
// stale is true for a hashed ip made under an IP key rotated since; hits is
// how many requests it turned away since the server started.
func IpBlockJSON(sandbox *api.Sandbox, block blocker_db.IpBlockRecord, now int64, rotatedAt int64) *serializabledeps.SerializableObject {
	item := sandbox.Deps.SerializableDeps.CreateObject()
	item.AddItemToObject("id", block.Id)
	item.AddItemToObject("value", block.Value)
	item.AddItemToObject("kind", block.Kind)
	item.AddItemToObject("reason", block.Reason)
	item.AddItemToObject("created_at", block.CreatedAt)
	item.AddItemToObject("expires_at", block.ExpiresAt)
	item.AddItemToObject("expired", blocker.Expired(block, now))
	item.AddItemToObject("stale", blocker.Stale(sandbox, block, rotatedAt))
	item.AddItemToObject("hits", blocker.Hits(sandbox, block.Value))
	return item
}

// IpBlocksJSON is every block: {ip_blocks, total}.
func IpBlocksJSON(sandbox *api.Sandbox, blocks []blocker_db.IpBlockRecord) *serializabledeps.SerializableObject {
	now := blocker.NowSeconds(sandbox)
	rotatedAt, _ := privacy.IpKeyRotatedAt(sandbox)
	list := sandbox.Deps.SerializableDeps.CreateArray()
	for _, block := range blocks {
		list.AddItemToArray(IpBlockJSON(sandbox, block, now, rotatedAt))
	}
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("ip_blocks", list)
	document.AddItemToObject("total", len(blocks))
	return document
}

// IpBlockResponseJSON is {ip_block}.
func IpBlockResponseJSON(sandbox *api.Sandbox, block blocker_db.IpBlockRecord) *serializabledeps.SerializableObject {
	rotatedAt, _ := privacy.IpKeyRotatedAt(sandbox)
	document := sandbox.Deps.SerializableDeps.CreateObject()
	document.AddItemToObject("ip_block", IpBlockJSON(sandbox, block, blocker.NowSeconds(sandbox), rotatedAt))
	return document
}
