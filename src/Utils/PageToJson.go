package Utils

import (
	"bufio"
	"encoding/json"
	"os"
	"../DataStructure"
)

func WritePage(resultPath string, page *DataStructure.Page) {
	outFile, err := os.Create(resultPath + page.PageID+".json")
	if err == nil {
		writer := bufio.NewWriter(outFile)
		defer outFile.Close()

		var dictPage, err = json.Marshal(page)
		if err == nil{
			_, _ = writer.Write(dictPage)
			_ = writer.Flush()
		}
	}
}
