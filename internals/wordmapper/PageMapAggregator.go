package wordmapper

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internals/structures"
	"github.com/negapedia/wikitfidf/internals/utils"
)

func getTotalWordInPage(page *structures.PageElement) uint32 {
	var tot uint32
	for _, wordFreq := range page.Word {
		tot += wordFreq
	}

	return tot
}

// PageMapAggregator given the result dir, aggregate all the page files into a global file
func PageMapAggregator(resultDir string) error {
	fileList, err := utils.FilesInDir(resultDir, "M[0-9]*")
	if err != nil {
		return err
	}

	outFile, _ := os.Create(resultDir + "GlobalPages.json")
	defer outFile.Close()
	encWriter := bufio.NewWriter(outFile)

	for i, file := range fileList {
		jsonFile, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "Error happened while trying to open file:"+file)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		_ = os.Remove(file)

		var Page structures.PageElement

		err = json.Unmarshal(byteValue, &Page)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json.")
		}

		pageToWrite := make(map[uint32]structures.AggregatedPage)
		pageToWrite[Page.PageID] = structures.AggregatedPage{TopicID: Page.TopicID, Tot: getTotalWordInPage(&Page), Words: Page.Word}

		if i == 0 {
			marshalledPage, _ := json.Marshal(pageToWrite)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[:len(pageAsString)-1] + ",\n"
			_, err = encWriter.Write([]byte(pageAsString))

		} else if /*i != nFile-1 && */ i > 0 {
			marshalledPage, _ := json.Marshal(pageToWrite)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[1:len(pageAsString)-1] + ",\n"
			_, err = encWriter.Write([]byte(pageAsString))
		}
		if err != nil {
			return errors.Wrap(err, "Error while trying to write to file")
		}

		err = encWriter.Flush()
		if err != nil {
			return errors.Wrap(err, "Error while trying to flush to file")
		}
	}

	_, err = encWriter.Write([]byte("}"))
	if err != nil {
		return errors.Wrap(err, "Error while trying to write to file")
	}
	err = encWriter.Flush()
	if err != nil {
		return errors.Wrap(err, "Error while trying to flush to file")
	}
	return nil
}
