/*
 * Developed by Marco Chilese.
 * Last modified 7/18/19 6:02 PM
 *
 */

package utils

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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

// getSliceOfUintFilenames given a slice of filepath returns a slice with all the uint32 pageIDs
func getSliceOfUintFilenames(files []string, initialLetter, extension string) ([]uint32, error) {
	var uint32Filenames []uint32
	for i, _ := range files {
		_, file := filepath.Split(files[i])

		file = strings.ReplaceAll(file, initialLetter, "")
		file = strings.ReplaceAll(file, extension, "")

		pageId, err := strconv.ParseUint(file, 10, 32)
		if err != nil {
			return nil, errors.Wrapf(err, "Error while trying to convert string to uint32:"+files[i])
		}
		uint32Filenames = append(uint32Filenames, uint32(pageId))
	}

	return uint32Filenames, nil
}

func sortUint32Slice(mySlice []uint32) {
	sort.Slice(mySlice, func(i, j int) bool { return mySlice[i] < mySlice[j] })
}

// buildFullSortedFileSlice build a []string slice containing ordered files given the result dir, the desired inital letter, original extension
// and the slice with uint32 pageID
func buildFullSortedFileSlice(mySlice []uint32, dir, initialLetter, extension string) []string {
	var pathSlice []string
	for _, pageID := range mySlice {
		pathSlice = append(pathSlice, filepath.Join(dir, initialLetter+fmt.Sprint(pageID)+extension))
	}

	return pathSlice
}

// FilesInDirSorted return a list of string of the files in a directory filtered by pattern sorted by increasing order
func FilesInDirSorted(dir string, pattern string, initialLetter string, extension string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(dir, pattern))

	if err != nil {
		return nil, errors.Wrapf(err, "Error while trying to list file in path: "+dir)
	}

	uint32Filenames, err := getSliceOfUintFilenames(files, initialLetter, extension)
	if err != nil {
		return nil, err
	}
	sortUint32Slice(uint32Filenames)
	toReturn := buildFullSortedFileSlice(uint32Filenames, dir, initialLetter, extension)

	return toReturn, err
}
