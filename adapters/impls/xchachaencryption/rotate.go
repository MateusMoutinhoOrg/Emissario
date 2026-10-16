package xchachaencryption

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// rotation is the old key and the new one, read once off the state.
func rotation() (oldRing *keyring, newRing *keyring, err error) {
	state.RLock()
	defer state.RUnlock()
	if state.mode != rotating || state.old == nil || state.current == nil {
		return nil, nil, errors.New("the storage is not armed for a rotation")
	}
	return state.old, state.current, nil
}

// rotateFiles fills encryptiondeps.Contract.RotateFiles.
func rotateFiles(dir string) (int, error) {
	oldRing, newRing, err := rotation()
	if err != nil {
		return 0, err
	}
	resealed := 0
	err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || strings.Contains(entry.Name(), ".") {
			return nil
		}
		relative, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		value, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		changed, done, err := resealWith(value, relative, oldRing, newRing)
		if err != nil || !done {
			return err
		}
		if err := replace(path, changed); err != nil {
			return err
		}
		resealed++
		return nil
	})
	return resealed, err
}

// reseal fills encryptiondeps.Contract.Reseal.
func reseal(value []byte, path string) ([]byte, bool, error) {
	oldRing, newRing, err := rotation()
	if err != nil {
		return nil, false, err
	}
	return resealWith(value, path, oldRing, newRing)
}

// resealWith answers value, the file at path, sealed under newRing, and
// whether that changed it.
func resealWith(value []byte, path string, oldRing *keyring, newRing *keyring) ([]byte, bool, error) {
	if len(value) >= overhead && value[0] == header && bytes.Equal(value[1:1+keyIdBytes], newRing.id[:]) {
		return value, false, nil
	}
	opened, err := openWith(path, value, oldRing)
	if err != nil {
		return nil, false, err
	}
	sealed, err := sealWith(newRing, path, opened)
	clear(opened)
	if err != nil {
		return nil, false, err
	}
	return sealed, true, nil
}

// replace writes value to a temporary file beside path, flushes it and
// renames it over path, so a crash leaves the old value or the new one,
// never a torn one. The temporary file's name starts with a '.', as the
// store's own do, so neither a snapshot nor a rotation takes it for a value.
func replace(path string, value []byte) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".emissario-rotate-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if _, err := temp.Write(value); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, 0o600); err != nil {
		return err
	}
	return os.Rename(name, path)
}
