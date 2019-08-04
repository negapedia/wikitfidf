package wordmapper

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"../structures"
	"../utils"
)

func getMappedPage(page *structures.StemmedPageJson) structures.PageElement {
	var mappedText = make(map[string]float64)

	for _, rev := range page.Revision {
		for _, word := range rev.Text {
			if _, ok := mappedText[word]; ok {
				mappedText[word] += 1
			} else {
				mappedText[word] = 1
			}
		}
	}
	return structures.PageElement{PageId: page.PageID, TopicID: page.TopicID, Word: mappedText}
}

// WordMapperByPage given the result dir, generate a global file containing all the processed pages
func WordMapperByPage(resultDir string) {
	fileList := utils.FilesInDir(resultDir, "S[0-9]*")

	for _, file := range fileList {
		jsonFile, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		var page structures.StemmedPageJson

		_ = json.Unmarshal(byteValue, &page)

		mappedPage := getMappedPage(&page)
		_ = os.Remove(file)
		if len(mappedPage.Word) > 0 {
			utils.WriteMappedPage(resultDir, &mappedPage)
		}
	}
}