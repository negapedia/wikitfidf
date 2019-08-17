package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/negapedia/wikiconflict/internals/structures"
	"github.com/negapedia/wikiconflict/internals/utils"
)

func getMappedPage(page *structures.StemmedPageJSON) structures.PageElement {
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

// ByPage given the result dir, generate a global file containing all the processed pages
func ByPage(resultDir string) error {
	fileList, err := utils.FilesInDir(resultDir, "S[0-9]*")
	if err != nil {
		return err
	}
	nFile := len(fileList)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)
		jsonFile, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "Error while opening file.")
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		var page structures.StemmedPageJSON

		err = json.Unmarshal(byteValue, &page)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json.")
		}

		mappedPage := getMappedPage(&page)
		_ = os.Remove(file)
		if len(mappedPage.Word) > 0 {
			utils.WriteMappedPage(resultDir, &mappedPage)
		}
	}
	fmt.Println()
	return nil
}
