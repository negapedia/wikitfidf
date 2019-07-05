package WordMapper

import (
	"../DataStructure"
	"../Utils"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type PageData struct {
	Title string
	Words *map[string]uint64
}

func PageMapAggregator(resultDir string) {
	fileList := Utils.FilesInDir(resultDir, ".json", "M")
	nFile := len(fileList)

	outFile, _ := os.Create(resultDir+"GlobalPage.json")
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

		var Page DataStructure.PageElement

		_ = json.Unmarshal(byteValue, &Page)

		pageToWrite := make(map[string]PageData)
		pageToWrite[Page.PageId] = PageData{Title: Page.Title, Words: &Page.Word}


		if i == 0{
			marshalledPage, _ := json.Marshal(pageToWrite)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[:len(pageAsString)-1] + ",\n"
			encWriter.Write([]byte(pageAsString))

		} else if i != nFile-1 && i > 0 {
			marshalledPage, _ := json.Marshal(pageToWrite)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[1:len(pageAsString)-1] + ",\n"
			encWriter.Write([]byte(pageAsString))

		} else if i == nFile-1 {
			marshalledPage, _ := json.Marshal(pageToWrite)
			pageAsString := string(marshalledPage)
			pageAsString = pageAsString[1:]
			encWriter.Write([]byte(pageAsString))
		}

		_ = encWriter.Flush()
	}
}
