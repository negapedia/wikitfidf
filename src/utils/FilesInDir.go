package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// The function return a list of string of the files in a directory filtered by extension and partial name
func FilesInDir(dir string, extension string, partialName string) []string {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if extension == "" && partialName != "" {
			if !info.IsDir() && strings.Contains(info.Name(), partialName){
				files = append(files, path)
			}
		} else if extension != "" && partialName == "" {
			if !info.IsDir() && filepath.Ext(path) == extension{
				files = append(files, path)
				return nil
			}
		} else if extension == "" && partialName == ""{
			if !info.IsDir() {
				files = append(files, path)
				return nil
			}
		} else if !info.IsDir() && filepath.Ext(path) == extension && strings.Contains(info.Name(), partialName){	// extension and partialName are valued
			files = append(files, path)
			return nil
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}
