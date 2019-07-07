package WordMapper

import (
	"../DataStructure"
	"../Utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func GlobalWordMapper(resultDir string) {
	fileList := Utils.FilesInDir(resultDir, ".json", "M")
	nFile := len(fileList)

	globalWord := make(map[string]map[string]float64)
	var totalWord float64
	totalWord = 0

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)

		jsonFile, err := os.Open(file)
		// if we os.Open returns an error then handle it
		if err != nil {
			panic(err)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		_ = jsonFile.Close()

		var page DataStructure.PageElement

		_ = json.Unmarshal(byteValue, &page)

		for word, freq := range page.Word {
			if _, ok := globalWord[word]; ok {
				globalWord[word]["abs"] += float64(freq)
				globalWord[word]["in"] += 1
			} else {
				globalWord[word] = make(map[string]float64)
				globalWord[word]["abs"] = float64(freq)
				globalWord[word]["in"] = 1
			}
			totalWord += float64(freq)
		}
	}

	globalWord["@Total Word"] = make(map[string]float64)
	globalWord["@Total Word"]["tot"] = totalWord

	Utils.WriteGlobalWord(resultDir, &globalWord)
}
