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
	"log"
	"os"

	"github.com/negapedia/Wikipedia-Conflict-Analyzer/structures"
)

// WriteCleanPage write a json containing the page after first clean
func WriteCleanPage(resultDir string, page *structures.Page) {
	outFile, err := os.Create(resultDir + fmt.Sprint(page.TopicID) + "_" + fmt.Sprint(page.PageID) + ".json")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	writer := bufio.NewWriter(outFile)
	defer outFile.Close()

	dictPage, err := json.Marshal(page)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	_, _ = writer.Write(dictPage)
	_ = writer.Flush()
}

// WriteMappedPage write a json containing the mapped page with term frequency
func WriteMappedPage(resultPath string, page *structures.PageElement) {
	outFile, err := os.Create(resultPath + "M" + fmt.Sprint(page.TopicID) + "_" + fmt.Sprint(page.PageID) + ".json")
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
func WriteGlobalWord(resultPath string, gloabalWord *map[string]map[string]uint32) {
	outFile, err := os.Create(resultPath + "GlobalWords.json")
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
