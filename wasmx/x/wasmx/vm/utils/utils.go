package utils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
)

func SafeWriteFile(path string, data []byte) (err error) {
	content := bytes.NewReader(data)
	return SafeWriteReader(path, content)
}

func SafeWriteReader(path string, content io.Reader) (err error) {
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
	if _, err = io.Copy(file, content); err != nil {
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
