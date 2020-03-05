package utils

import (
	"os"
	"path/filepath"
	"github.com/ricochet2200/go-disk-usage/du"
)

// https://stackoverflow.com/questions/32482673/how-to-get-directory-total-size
func DirSize(path string) (uint64, error) {
	var size uint64 = 0
	err := filepath.Walk(path, func(subpath string, info os.FileInfo, err error) error {
		GetLogger().Printf("[DEBUG] Walking path: %v.", subpath)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	return size, err
}

func AvailableDisk(path string) uint64 {
	// TODO: error wrapper in case this third party function flips
	usage := du.NewDiskUsage(path)
	return usage.Available()
}
