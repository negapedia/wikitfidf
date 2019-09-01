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
	files, err := filepath.Glob(dir + pattern)

	if err != nil {
		return nil, errors.Wrapf(err, "Error while trying to list file in path: "+dir)
	}

	sort.Strings(files) // sort in increasing order

	for i := len(files)/2 - 1; i >= 0; i-- { // flip the slice --> decreasing order
		opp := len(files) - 1 - i
		files[i], files[opp] = files[opp], files[i]
	}
	return files, nil
}
