package wordmapper

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/negapedia/wikitfidf/internal/structures"
	"github.com/negapedia/wikitfidf/internal/utils"
)

func getTotalWordInPage(page *structures.PageElement) uint32 {
	var tot uint32
	for _, wordFreq := range page.Word {
		tot += wordFreq
	}

	return tot
}

// PageMapAggregator given the result dir, aggregate all the page files into a global file
func PageMapAggregator(resultDir string) (err error) {
	fileList, err := utils.FilesInDir(resultDir, "M[0-9]*")
	if err != nil {
		return err
	}

	outFile, _ := os.Create(filepath.Join(resultDir, "GlobalPages.json"))
	defer outFile.Close()
	encWriter := bufio.NewWriter(outFile)

	for i, file := range fileList {
		jsonFile, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "Error happened while trying to open json %v", file)
		}

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			return errors.Wrapf(err, "Error while reading json %v", file)
		}

		if err = os.Remove(file); err != nil {
			return errors.Wrapf(err, "Error while removing %v", file)
		}

		if err = jsonFile.Close(); err != nil {
			return errors.Wrapf(err, "Error while closing json %v", file)
		}

		if err != nil {
			return errors.Wrapf(err, "Error while removing json %v", file)
		}

		var Page structures.PageElement

		if err = json.Unmarshal(byteValue, &Page); err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json %v", file)
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
