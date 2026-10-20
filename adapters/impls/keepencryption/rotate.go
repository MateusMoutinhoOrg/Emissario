package keepencryption

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// rotateFiles fills encryptiondeps.Contract.RotateFiles.
func rotateFiles(dir string, drop func(path string) bool) (int, int, error) {
	oldKey, newKey, err := rotation()
	if err != nil {
		return 0, 0, err
	}
	resealed, dropped := 0, 0
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
		if drop(relative) {
			if err := os.Remove(path); err != nil {
				return err
			}
			dropped++
			return nil
		}
		value, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sealed, changed, err := resealWith(value, oldKey, newKey)
		if err != nil {
			return errors.New("the file " + relative + " " + err.Error())
		}
		if !changed {
			return nil
		}
		if err := replace(path, sealed); err != nil {
			return err
		}
		resealed++
		return nil
	})
	return resealed, dropped, err
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
