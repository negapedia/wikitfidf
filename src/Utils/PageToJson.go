package Utils

import (
	"../DataStructure"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func WritePage(resultPath string, page *DataStructure.Page) {
	outFile, err := os.Create(resultPath + fmt.Sprint(page.PageID) + ".json")
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

func WriteMappedPage(resultPath string, page *DataStructure.PageElement) {
	outFile, err := os.Create(resultPath + "M" + page.PageId + ".json")
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