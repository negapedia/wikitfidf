package wordmapper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/negapedia/wikiconflict/internals/structures"
	"github.com/negapedia/wikiconflict/internals/utils"
)

func getTotalWordInPage(page *structures.PageElement) uint32 {
	var tot uint32
	tot = 0
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
	nFile := len(fileList)

	outFile, _ := os.Create(resultDir + "GlobalPages.json")
	encWriter := bufio.NewWriter(outFile)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)

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
			encWriter.Write([]byte(pageAsString))

		} else if /*i != nFile-1 && */ i > 0 {
			marshalledPage, _ := json.Marshal(pageToWrite)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[1:len(pageAsString)-1] + ",\n"
			encWriter.Write([]byte(pageAsString))

		}

		_ = encWriter.Flush()
	}
	fmt.Println()

	encWriter.Write([]byte("}"))
	_ = encWriter.Flush()
	outFile.Close()
	return nil
}
