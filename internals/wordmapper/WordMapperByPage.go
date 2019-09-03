package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internals/structures"
	"github.com/negapedia/wikitfidf/internals/utils"
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

	var skipped int
	for _, file := range fileList {
		jsonFile, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "Error while opening file: "+file)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		var page structures.StemmedPageJSON

		err = json.Unmarshal(byteValue, &page)
		if err != nil {
			//return errors.Wrapf(err, "Error while unmarshalling json."+file)
			skipped++
			_ = os.Remove(file)
			continue
		}

		mappedPage := getMappedPage(&page)
		_ = os.Remove(file)
		if len(mappedPage.Word) > 0 {
			utils.WriteMappedPage(resultDir, &mappedPage)
		}
	}
	fmt.Println()
	fmt.Println("SKIPPED: ", skipped)
	return nil
}
