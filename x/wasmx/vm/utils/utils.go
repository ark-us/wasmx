package utils

import (
	"os"
	"path/filepath"
)

func SafeWriteFile(path string, content []byte) (err error) {
	dirPath, fileName := filepath.Split(path)

	file, err := os.CreateTemp(dirPath, fileName+".*.tmp")
	if err != nil {
		return
	}
	defer func() {
		file.Close()
		if err != nil {
			_ = os.Remove(file.Name())
		}
	}()
	if err = os.WriteFile(path, content, 0644); err != nil {
		return
	}
	if err = file.Sync(); err != nil {
		return
	}
	if err = file.Close(); err != nil {
		return
	}
	err = os.Rename(file.Name(), path)
	return
}
