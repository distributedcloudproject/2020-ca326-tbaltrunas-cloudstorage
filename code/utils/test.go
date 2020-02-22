package utils

import (
	"errors"
	"io/ioutil"
	"os"
)

// Shared test functions.
// This file should not import any cloud packages to avoid any import cycles.

// GetTestDirs creates n temporary directories using the given prefix.
// It returns the path of all the created directories.
// It is the caller's responsibility to remove the temporary directories.
// The caller may call the GetTestDirsCleanup with defer for appropriate clean up.
func GetTestDirs(prefix string, n int) ([]string, error) {
	if n < 1 {
		return nil, errors.New("Number of directories is less than one.")
	}
	tmpDirs := make([]string, 0)
	for i := 0; i < n; i++ {
		dir, err := ioutil.TempDir("", prefix)
		if err != nil {
			return nil, err
		}
		tmpDirs = append(tmpDirs, dir)
	}
	return tmpDirs, nil
}

// GetTestDirsCleanup performs clean up for GetTestDirs.
func GetTestDirsCleanup(dirs []string) {
	RemoveDirs(dirs)
}

// RemoveDirs removes all the directories in the list.
func RemoveDirs(dirs []string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}

// GetTestFile returns a temporary file with the given byte contents and a filename prefix.
// It is the caller's responsibility to remove the temporary file.
// The caller may call the GetTestFileCleanup with defer for appropriate clean up.
func GetTestFile(prefix string, contents []byte) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", prefix)
	if err != nil {
		return nil, err
	}
	GetLogger().Printf("Temporary filepath: %s", tmpfile.Name())
	GetLogger().Printf("Writing contents to temporary file: %s", contents)
	err = ioutil.WriteFile(tmpfile.Name(), contents, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return tmpfile, nil
}

// GetTestFileCleanup performs clean up for GetTestFile.
func GetTestFileCleanup(file *os.File) {
	file.Close()
	os.Remove(file.Name())
}
