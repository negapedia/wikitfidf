package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internal/structures"
	"github.com/negapedia/wikitfidf/internal/utils"
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

	for _, file := range fileList {
		jsonFile, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "Error while opening json %v", file)
		}

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return errors.Wrapf(err, "Error while reading json %v", file)
		}

		if err = jsonFile.Close(); err != nil {
			return errors.Wrapf(err, "Error while closing json %v", file)
		}

		var page structures.StemmedPageJSON

		err = json.Unmarshal(byteValue, &page)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json %v", file)
		}

		mappedPage := getMappedPage(&page)
		err = os.Remove(file)
		if err != nil {
			return errors.Wrapf(err, "Error while removing json %v", file)
		}

		if len(mappedPage.Word) == 0 {
			continue
		}

		err = utils.Write2JSON(filepath.Join(resultDir, "M"+fmt.Sprint(page.PageID)+".json"), mappedPage)
		if err != nil {
			return err
		}
	}
	return nil
}
