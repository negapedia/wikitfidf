package wordmapper

import (
	"../datastructure"
	"../utils"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func getTotalWordInPage(page *datastructure.PageElement) float64 {
	var tot float64
	tot = 0
	for _, wordFreq := range page.Word {
		tot += wordFreq
	}

	return tot
}

// The function, given the result dir, aggregate all the page files into a global file
func PageMapAggregator(resultDir string) {
	fileList := utils.FilesInDir(resultDir, ".json", "M")
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
		_ = jsonFile.Close()
		_ = os.Remove(file)

		var Page datastructure.PageElement

		_ = json.Unmarshal(byteValue, &Page)

		pageToWrite := make(map[uint32]datastructure.AggregatedPage)
		pageToWrite[Page.PageId] = datastructure.AggregatedPage{Tot: getTotalWordInPage(&Page), Words: Page.Word}

		if i == 0 {
			marshalledPage, _ := json.Marshal(pageToWrite)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[:len(pageAsString)-1] + ",\n"
			encWriter.Write([]byte(pageAsString))

		} else if /*i != nFile-1 && */i > 0 {
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
