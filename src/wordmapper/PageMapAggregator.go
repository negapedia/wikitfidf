package wordmapper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"../structures"
	"../utils"
)

func getTotalWordInPage(page *structures.PageElement) float64 {
	var tot float64
	tot = 0
	for _, wordFreq := range page.Word {
		tot += wordFreq
	}

	return tot
}

// PageMapAggregator given the result dir, aggregate all the page files into a global file
func PageMapAggregator(resultDir string) {
	fileList := utils.FilesInDir(resultDir, "M[0-9]*")
	nFile := len(fileList)

	outFile, _ := os.Create(resultDir + "GlobalPage.json")
	encWriter := bufio.NewWriter(outFile)

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			panic(err)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		_ = os.Remove(file)

		var Page structures.PageElement

		_ = json.Unmarshal(byteValue, &Page)

		pageToWrite := make(map[uint32]structures.AggregatedPage)
		pageToWrite[Page.PageId] = structures.AggregatedPage{Tot: getTotalWordInPage(&Page), Words: Page.Word}

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

	encWriter.Write([]byte("}"))
	_ = encWriter.Flush()
	outFile.Close()
}
