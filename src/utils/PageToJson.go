/*
 * Developed by Marco Chilese.
 * Last modified 7/19/19 10:53 AM
 *
 */

package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"../structures"
)

// WriteMappedPage write a json containing the mapped page with term frequency
func WriteMappedPage(resultPath string, page *structures.PageElement) {
	outFile, err := os.Create(resultPath + "M" + fmt.Sprint(page.PageId) + ".json")
	if err == nil {
		writer := bufio.NewWriter(outFile)
		defer outFile.Close()

		var dictPage, err = json.Marshal(page)
		if err == nil {
			_, _ = writer.Write(dictPage)
			_ = writer.Flush()
		}
	}
}

// WriteGlobalWord write a json of global word map
func WriteGlobalWord(resultPath string, gloabalWord *map[string]map[string]float64) {
	outFile, err := os.Create(resultPath + "GlobalWord.json")
	if err == nil {
		writer := bufio.NewWriter(outFile)
		defer outFile.Close()

		var dictPage, err = json.Marshal(gloabalWord)
		if err == nil {
			_, _ = writer.Write(dictPage)
			_ = writer.Flush()
		}
	}
}

// WriteGlobalStem write a json of global stemming dictionary
func WriteGlobalStem(resultPath string, gloabaStem *map[string]string) {
	outFile, err := os.Create(resultPath + "GlobalStem.json")
	if err == nil {
		writer := bufio.NewWriter(outFile)
		defer outFile.Close()

		var dictPage, err = json.Marshal(gloabaStem)
		if err == nil {
			_, _ = writer.Write(dictPage)
			_ = writer.Flush()
		}
	}
}
