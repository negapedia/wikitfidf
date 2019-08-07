/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 6:02 PM
 *
 */

package utils

import (
	"path/filepath"
)

// FilesInDir return a list of string of the files in a directory filtered by pattern
func FilesInDir(dir string, pattern string) []string {
	files, err := filepath.Glob(dir + pattern)

	if err != nil {
		panic(err)
	}
	return files
}
