package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"../datastructure"
)

// The function write a json containing the mapped page with term frequency
func WriteMappedPage(resultPath string, page *datastructure.PageElement) {
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

// The function write a json of global word map
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

// The function write a json of global stemming dictionary
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
