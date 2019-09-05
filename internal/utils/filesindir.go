/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 6:02 PM
 *
 */

package utils

import (
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

// FilesInDir return a list of string of the files in a directory filtered by pattern
func FilesInDir(dir string, pattern string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, pattern))

	if err != nil {
		return nil, errors.Wrapf(err, "Error while trying to list file in path: "+dir)
	}

	sort.Strings(files) // sort in increasing order

	return files, nil
}
