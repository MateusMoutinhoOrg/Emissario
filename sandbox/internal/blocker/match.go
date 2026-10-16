package blocker

import (
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/api"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/databases/blocker_db"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/privacy"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/server/backoffice/backofficehttp"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/snapshots"
	"github.com/MateusMoutinhoOrg/Emissario/sandbox/internal/visualization"
)

// index is the blocks every request is matched against, by kind: an ip, a
// hash and a country by a lookup, a range by a scan. Each value maps to when
// its block ends, 0 for never.
type index struct {
	ips       map[string]int64
	hashes    map[string]int64
	countries map[string]int64
	ranges    []rangeBlock
	// readAt is when the blocks were read, restores how many restores had
	// ended then.
	readAt   int64
	restores int64
	valid    bool
}

// rangeBlock is one range block of index.
type rangeBlock struct {
	value     string
	expiresAt int64
}

// cache is the index the requests go by, guarded by cacheLock.
var cache index

// cacheLock is held while cache is read or written.
var cacheLock = make(chan struct{}, 1)

// hits is how many requests each value blocked since the process started,
// guarded by hitsLock.
var hits = map[string]int64{}

// hitsLock is held while hits is read or written.
var hitsLock = make(chan struct{}, 1)

// Invalidate makes the next request read the blocks from the store again.
func Invalidate(sandbox *api.Sandbox) {
	cacheLock <- struct{}{}
	cache.valid = false
	<-cacheLock
}

// current is the index read within cacheSeconds and since the last restore,
// else read again. A store that fails keeps the last index read, and blocks
// nothing when none was: a broken store never turns every visitor away.
func current(sandbox *api.Sandbox) index {
	cacheLock <- struct{}{}
	defer func() { <-cacheLock }()
	now := NowSeconds(sandbox)
	restores := snapshots.RestoresEnded(sandbox)
	if cache.valid && cache.restores == restores && now-cache.readAt < cacheSeconds && now >= cache.readAt {
		return cache
	}
	blocks, err := blocker_db.New(sandbox).ListIpBlocks(blocker_db.IpBlockFilter{})
	if err != nil {
		sandbox.Deps.StdDeps.Eprintf("warning: the ip blocks could not be read: %s\n", err.Error())
		return cache
	}
	cache = indexOf(sandbox, blocks, now)
	cache.restores = restores
	return cache
}

// indexOf is blocks as an index read at now; a block that ended is left out.
func indexOf(sandbox *api.Sandbox, blocks []blocker_db.IpBlockRecord, now int64) index {
	built := index{ips: map[string]int64{}, hashes: map[string]int64{}, countries: map[string]int64{}, readAt: now, valid: true}
	for _, block := range blocks {
		if Expired(block, now) {
			continue
		}
		switch block.Kind {
		case KindIp:
			built.ips[block.Value] = block.ExpiresAt
		case KindHash:
			built.hashes[block.Value] = block.ExpiresAt
		case KindCountry:
			built.countries[block.Value] = block.ExpiresAt
		case KindRange:
			built.ranges = append(built.ranges, rangeBlock{value: block.Value, expiresAt: block.ExpiresAt})
		}
	}
	return built
}

// Blocked answers the value of the block ip falls under, false when none:
// its ip, a range holding it, its hash under the current IP key, or its
// country. A block that ended since the blocks were read is skipped. An ip
// that is not one is never blocked.
func Blocked(sandbox *api.Sandbox, ip string) (string, bool) {
	blocks := current(sandbox)
	if len(blocks.ips)+len(blocks.hashes)+len(blocks.countries)+len(blocks.ranges) == 0 {
		return "", false
	}
	address, ok := backofficehttp.ParseIp(sandbox, ip)
	if !ok {
		return "", false
	}
	now := NowSeconds(sandbox)
	live := func(expiresAt int64) bool { return expiresAt == 0 || expiresAt > now }
	canonical := backofficehttp.FormatIp(sandbox, address)
	if expiresAt, ok := blocks.ips[canonical]; ok && live(expiresAt) {
		return hit(canonical), true
	}
	for _, block := range blocks.ranges {
		if live(block.expiresAt) && backofficehttp.IpMatches(sandbox, block.value, canonical) {
			return hit(block.value), true
		}
	}
	if len(blocks.hashes) > 0 {
		if hash, err := privacy.HashIp(sandbox, canonical); err == nil {
			if expiresAt, ok := blocks.hashes[hash]; ok && live(expiresAt) {
				return hit(hash), true
			}
		}
	}
	if len(blocks.countries) > 0 {
		country := visualization.Country(sandbox, canonical)
		if expiresAt, ok := blocks.countries[country]; ok && live(expiresAt) {
			return hit(country), true
		}
	}
	return "", false
}

// Matches tells whether a block of kind and value would turn ip away.
func Matches(sandbox *api.Sandbox, kind string, value string, ip string) bool {
	address, ok := backofficehttp.ParseIp(sandbox, ip)
	if !ok {
		return false
	}
	canonical := backofficehttp.FormatIp(sandbox, address)
	switch kind {
	case KindIp:
		return canonical == value
	case KindRange:
		return backofficehttp.IpMatches(sandbox, value, canonical)
	case KindHash:
		hash, err := privacy.HashIp(sandbox, canonical)
		return err == nil && hash == value
	case KindCountry:
		return visualization.Country(sandbox, canonical) == value
	}
	return false
}

// hit counts one request value blocked, and answers value.
func hit(value string) string {
	hitsLock <- struct{}{}
	hits[value]++
	<-hitsLock
	return value
}

// Hits is how many requests the block of value turned away since the
// process started.
func Hits(sandbox *api.Sandbox, value string) int64 {
	hitsLock <- struct{}{}
	defer func() { <-hitsLock }()
	return hits[value]
}
