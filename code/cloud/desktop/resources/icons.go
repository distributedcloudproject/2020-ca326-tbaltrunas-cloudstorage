package resources

import (
	"fyne.io/fyne/theme"
)

var FileIcons map[string]*theme.ThemedResource

func init() {
	FileIcons = make(map[string]*theme.ThemedResource)

	FileIcons["file"] = theme.NewThemedResource(fileIconRes, nil)
}
