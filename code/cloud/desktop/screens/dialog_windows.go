// +build windows

package screens

import (
	"github.com/sqweek/dialog"
)

func LoadFileDialog() (string, error) {
	return dialog.File().Load()
}

func SaveFileDialog() (string, error) {
	return dialog.File().Save()
}

func BrowseDirDialog() (string, error) {
	return dialog.Directory().Browse()
}
