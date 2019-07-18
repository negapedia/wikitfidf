package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"../datastructure"
	"../utils"
)

func getMappedPage(page *datastructure.StemmedPageJson) datastructure.PageElement {
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
	return datastructure.PageElement{PageId: page.PageID, Word: mappedText}
}

// WordMapperByPage, given the result dir, generate a global file containing all the processed pages
func WordMapperByPage(resultDir string) {
	fileList := utils.FilesInDir(resultDir, ".json", "S")
	nFile := len(fileList)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i, nFile)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			panic(err)
		}
		// defer the closing of our jsonFile so that we can parse it later on

		// read our opened xmlFile as a byte array.
		byteValue, _ := ioutil.ReadAll(jsonFile)

		_ = jsonFile.Close()

		var page datastructure.StemmedPageJson

		// we unmarshal our byteArray which contains our
		// jsonFile's content into 'users' which we defined above
		_ = json.Unmarshal(byteValue, &page)

		mappedPage := getMappedPage(&page)
		_ = os.Remove(file)
		if len(mappedPage.Word) > 0 {
			utils.WriteMappedPage(resultDir, &mappedPage)
		}
	}
}
