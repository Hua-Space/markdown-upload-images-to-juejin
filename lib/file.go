package lib

import (
	"os"
	"path/filepath"
)

func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {

		return false
	}
	return s.IsDir()
}

func GetFileName(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !IsDir(path) {
			if filepath.Ext(path) == ".md" {
				files = append(files, path)
			}
		}
		return nil
	})

	return files, err
}
