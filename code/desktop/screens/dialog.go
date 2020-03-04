// +build !windows

package screens

import (
	"errors"
	"os"
	"os/exec"
)

func LoadFileDialog() (string, error) {
	cmd := exec.Command(os.Args[0] + " --dialog=file-load")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 && out[0] == '!' {
		return "", errors.New(string(out[1:]))
	}
	return string(out), nil
}

func SaveFileDialog() (string, error) {
	cmd := exec.Command(os.Args[0] + " --dialog=file-save")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 && out[0] == '!' {
		return "", errors.New(string(out[1:]))
	}
	return string(out), nil
}

func BrowseDirDialog() (string, error) {
	cmd := exec.Command(os.Args[0] + " --dialog=dir")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(out) > 0 && out[0] == '!' {
		return "", errors.New(string(out[1:]))
	}
	return string(out), nil
}
