package task

import (
	"os"
	"path"
)

const TmpVolumePath = "~/.elaina/volume"

func GetAbsTmpVolume() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return path.Join(pwd, TmpVolumePath)
}
