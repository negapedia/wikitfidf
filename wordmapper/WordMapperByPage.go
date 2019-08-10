package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/MarcoChilese/Wikipedia-Conflict-Analyzer/structures"
	"github.com/MarcoChilese/Wikipedia-Conflict-Analyzer/utils"
)

func getMappedPage(page *structures.StemmedPageJson) structures.PageElement {
	var mappedText = make(map[string]uint32)

	for _, rev := range page.Revision {
		for _, word := range rev.Text {
			if _, ok := mappedText[word]; ok {
				mappedText[word]++
			} else {
				mappedText[word] = 1
			}
		}
	}
	return structures.PageElement{PageID: page.PageID, TopicID: page.TopicID, Word: mappedText}
}

// WordMapperByPage given the result dir, generate a global file containing all the processed pages
func WordMapperByPage(resultDir string) {
	fileList := utils.FilesInDir(resultDir, "S[0-9]*")
	nFile := len(fileList)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)
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
	fmt.Println()
}
