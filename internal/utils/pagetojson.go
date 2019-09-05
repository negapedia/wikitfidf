/*
 * Developed by Marco Chilese.
 * Last modified 7/19/19 10:53 AM
 *
 */

package utils

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/pkg/errors"
)

// Write2JSON writes a json containing the data
func Write2JSON(filename string, data interface{}) (err error) {
	outFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "Error while creating file %v", filename)
	}

	writer := bufio.NewWriter(outFile)
	defer func() {
		if e := outFile.Close(); e != nil && err == nil {
			err = errors.Wrapf(e, "Error while closing file %v", filename)
		}
	}()

	dictPage, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "Error while marshalling data for file %v", filename)
	}

	if _, err = writer.Write(dictPage); err != nil {
		return errors.Wrapf(err, "Error while writing to file %v", filename)
	}

	if err = writer.Flush(); err != nil {
		return errors.Wrapf(err, "Error while flushing file %v", filename)
	}

	return
}
