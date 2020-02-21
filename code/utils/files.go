package utils

import (
	"os"
	"path/filepath"
)

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		GetLogger().Printf("[DEBUG] Walking path: %v.", subpath)
        if err != nil {
            return err
        }
        if !info.IsDir() {
            size += info.Size()
        }
        return nil
	})
	return size, err
}
