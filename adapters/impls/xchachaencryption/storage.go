package xchachaencryption

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	storagedeps "github.com/MateusMoutinhoOrg/Keep/sandbox/deps/storagedeps"
)

// casAttempts bounds how often a read-modify-write retries when another
// writer changed the value between its read and its conditional write.
const casAttempts = 64

// maxSegment and hexDigits are the ones of Keep's filestorage: encodeSegment
// spells a key segment the way that store names its file, so the additional
// data of a value is the path of its file relative to the --database folder.
const maxSegment = 200

const hexDigits = "0123456789ABCDEF"

// encodeSegment is Keep's filestorage encodeSegment: a-z, 0-9, '-' and '_'
// as they are, every other byte %XX, the empty segment "%", and an encoding
// longer than maxSegment "%%" followed by the SHA-256 of the segment. A test
// checks it against the files filestorage writes.
func encodeSegment(segment string) string {
	if segment == "" {
		return "%"
	}
	var encoded strings.Builder
	for index := 0; index < len(segment); index++ {
		character := segment[index]
		if character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '-' || character == '_' {
			encoded.WriteByte(character)
			continue
		}
		encoded.WriteByte('%')
		encoded.WriteByte(hexDigits[character>>4])
		encoded.WriteByte(hexDigits[character&15])
	}
	if encoded.Len() > maxSegment {
		sum := sha256.Sum256([]byte(segment))
		return "%%" + hex.EncodeToString(sum[:])
	}
	return encoded.String()
}

// relativePath is key below its first skip segments, spelled as the path of
// its file relative to the --database folder.
func relativePath(key []string, skip int) string {
	if skip > len(key) {
		skip = len(key)
	}
	parts := make([]string, 0, len(key)-skip)
	for _, segment := range key[skip:] {
		parts = append(parts, encodeSegment(segment))
	}
	return strings.Join(parts, "/")
}

// aad is the additional data of the value at path.
func aad(path string) []byte {
	return []byte(aadPrefix + path)
}

// sealWith seals plaintext under ring, for the file at path.
func sealWith(ring *keyring, path string, plaintext []byte) ([]byte, error) {
	sealed := make([]byte, 1+keyIdBytes+chacha20NonceSize, overhead+len(plaintext))
	sealed[0] = header
	copy(sealed[1:], ring.id[:])
	nonce := sealed[1+keyIdBytes:]
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return ring.aead.Seal(sealed, nonce, plaintext, aad(path)), nil
}

// openWith opens value, sealed for the file at path, under whichever of
// rings its header names. The error names the path, never the value.
func openWith(path string, value []byte, rings ...*keyring) ([]byte, error) {
	if len(value) < overhead || value[0] != header {
		return nil, errors.New("the value of " + path + " is not encrypted, in a database that is")
	}
	for _, ring := range rings {
		if ring == nil || !bytes.Equal(value[1:1+keyIdBytes], ring.id[:]) {
			continue
		}
		nonce := value[1+keyIdBytes : 1+keyIdBytes+chacha20NonceSize]
		opened, err := ring.aead.Open(nil, nonce, value[1+keyIdBytes+chacha20NonceSize:], aad(path))
		if err != nil {
			return nil, errors.New("the value of " + path + " does not open: it was changed, or moved from another place")
		}
		return opened, nil
	}
	return nil, errors.New("the value of " + path + " was encrypted with another key")
}

// chacha20NonceSize is the nonce of XChaCha20-Poly1305.
const chacha20NonceSize = 24

// codec is what one call of the wrapped storage does to the values it moves,
// read off the state once, so a call never mixes two modes.
type codec struct {
	mode mode
	skip int
	seal *keyring
	open []*keyring
}

// current is the codec of the state as it stands.
func current() (codec, error) {
	state.RLock()
	defer state.RUnlock()
	switch state.mode {
	case plain:
		return codec{mode: plain}, nil
	case sealed:
		return codec{mode: sealed, skip: state.skip, seal: state.current, open: []*keyring{state.current}}, nil
	case rotating:
		return codec{mode: rotating, skip: state.skip, seal: state.current, open: []*keyring{state.current, state.old}}, nil
	}
	return codec{}, errUnarmed
}

// clear tells whether values pass as they are.
func (c codec) clear() bool {
	return c.mode == plain
}

// seal answers the value to write at key for plaintext.
func (c codec) sealValue(key []string, plaintext []byte) ([]byte, error) {
	if c.clear() {
		return plaintext, nil
	}
	return sealWith(c.seal, relativePath(key, c.skip), plaintext)
}

// open answers the plaintext of value, read at key.
func (c codec) openValue(key []string, value []byte) ([]byte, error) {
	if c.clear() {
		return value, nil
	}
	return openWith(relativePath(key, c.skip), value, c.open...)
}

// Wrap answers inner with every value sealed on its way in and opened on its
// way out, as the state armed by EncryptionDeps says. Every one of its
// functions fails until the state is armed. Exists, Delete and the leases
// move no value, so they pass through once it is.
func Wrap(inner storagedeps.Contract) storagedeps.Contract {
	return storagedeps.Contract{
		Write: func(key []string, value []byte) error {
			c, err := current()
			if err != nil {
				return err
			}
			sealed, err := c.sealValue(key, value)
			if err != nil {
				return err
			}
			return inner.Write(key, sealed)
		},
		WriteIfAbsent: func(key []string, value []byte) (bool, error) {
			c, err := current()
			if err != nil {
				return false, err
			}
			sealed, err := c.sealValue(key, value)
			if err != nil {
				return false, err
			}
			return inner.WriteIfAbsent(key, sealed)
		},
		WriteIfValueEquals: func(key []string, value []byte, oldValue []byte) (bool, error) {
			c, err := current()
			if err != nil {
				return false, err
			}
			if c.clear() {
				return inner.WriteIfValueEquals(key, value, oldValue)
			}
			for attempt := 0; attempt < casAttempts; attempt++ {
				stored, found, err := inner.Read(key)
				if err != nil || !found {
					return false, err
				}
				opened, err := c.openValue(key, stored)
				if err != nil {
					return false, err
				}
				if !bytes.Equal(opened, oldValue) {
					return false, nil
				}
				sealed, err := c.sealValue(key, value)
				if err != nil {
					return false, err
				}
				written, err := inner.WriteIfValueEquals(key, sealed, stored)
				if err != nil || written {
					return written, err
				}
			}
			return false, nil
		},
		Append: func(key []string, value []byte) error {
			c, err := current()
			if err != nil {
				return err
			}
			if c.clear() {
				return inner.Append(key, value)
			}
			return modify(inner, c, key, func(opened []byte, found bool) ([]byte, error) {
				return append(append([]byte{}, opened...), value...), nil
			})
		},
		InsertAt: func(key []string, position int64, value []byte) error {
			c, err := current()
			if err != nil {
				return err
			}
			if c.clear() {
				return inner.InsertAt(key, position, value)
			}
			return modify(inner, c, key, func(opened []byte, found bool) ([]byte, error) {
				if position < 0 || position > int64(len(opened)) {
					return nil, errors.New("InsertAt: position past the end of the value")
				}
				spliced := make([]byte, 0, len(opened)+len(value))
				spliced = append(spliced, opened[:position]...)
				spliced = append(spliced, value...)
				return append(spliced, opened[position:]...), nil
			})
		},
		Exists: func(key []string) (bool, error) {
			if _, err := current(); err != nil {
				return false, err
			}
			return inner.Exists(key)
		},
		Read: func(key []string) ([]byte, bool, error) {
			c, err := current()
			if err != nil {
				return nil, false, err
			}
			stored, found, err := inner.Read(key)
			if err != nil || !found {
				return nil, found, err
			}
			opened, err := c.openValue(key, stored)
			if err != nil {
				return nil, false, err
			}
			return opened, true, nil
		},
		ReadAt: func(key []string, position int64, size int64) ([]byte, bool, error) {
			c, err := current()
			if err != nil {
				return nil, false, err
			}
			if c.clear() {
				return inner.ReadAt(key, position, size)
			}
			stored, found, err := inner.Read(key)
			if err != nil || !found {
				return nil, found, err
			}
			opened, err := c.openValue(key, stored)
			if err != nil {
				return nil, false, err
			}
			if position < 0 {
				position = 0
			}
			if position > int64(len(opened)) {
				return []byte{}, true, nil
			}
			end := int64(len(opened))
			if size >= 0 && position+size < end {
				end = position + size
			}
			return opened[position:end], true, nil
		},
		Delete: func(key []string) error {
			if _, err := current(); err != nil {
				return err
			}
			return inner.Delete(key)
		},
		Lock: func(key []string, seconds int) (bool, error) {
			if _, err := current(); err != nil {
				return false, err
			}
			return inner.Lock(key, seconds)
		},
		Unlock: func(key []string) error {
			if _, err := current(); err != nil {
				return err
			}
			return inner.Unlock(key)
		},
	}
}

// modify replaces the value at key with what change makes of its plaintext
// — empty and found false when the key holds nothing — sealing the result
// whole: a sealed value cannot be appended to in place. It writes through
// the inner conditional writes, so a concurrent writer is retried rather
// than overwritten.
func modify(inner storagedeps.Contract, c codec, key []string, change func(opened []byte, found bool) ([]byte, error)) error {
	for attempt := 0; attempt < casAttempts; attempt++ {
		stored, found, err := inner.Read(key)
		if err != nil {
			return err
		}
		opened := []byte{}
		if found {
			opened, err = c.openValue(key, stored)
			if err != nil {
				return err
			}
		}
		changed, err := change(opened, found)
		if err != nil {
			return err
		}
		sealed, err := c.sealValue(key, changed)
		if err != nil {
			return err
		}
		var written bool
		if found {
			written, err = inner.WriteIfValueEquals(key, sealed, stored)
		} else {
			written, err = inner.WriteIfAbsent(key, sealed)
		}
		if err != nil || written {
			return err
		}
	}
	return errors.New("the value of " + relativePath(key, c.skip) + " kept changing while it was being written")
}
