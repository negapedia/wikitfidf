/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 6:02 PM
 *
 */

package utils

import (
	"github.com/pkg/errors"
	"path/filepath"
)

// FilesInDir return a list of string of the files in a directory filtered by pattern
func FilesInDir(dir string, pattern string) ([]string, error) {
	files, err := filepath.Glob(dir + pattern)

	if err != nil {
		return nil, errors.Wrapf(err, "Error while trying to list file in path: "+dir)
	}
	return files, nil
}
