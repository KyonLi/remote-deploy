package util

import (
	"github.com/mholt/archiver"
	"os"
	"path"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func Compress(pathStr string) (string, error) {
	z := archiver.NewTarGz()
	fileName := path.Base(pathStr) + ".tar.gz"
	_ = os.Remove(fileName)
	if err := z.Archive([]string{pathStr}, fileName); err != nil {
		return "", err
	}
	return fileName, nil
}
