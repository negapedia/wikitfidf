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

	globalWord := make(map[string]uint64)
	var totalWord uint64
	totalWord = 0

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i, nFile)

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
				globalWord[word] += freq
			} else {
				globalWord[word] = freq
			}
			totalWord += freq
		}
	}

	globalWord["@Total Word"] = totalWord

	Utils.WriteGlobalWord(resultDir, &globalWord)
}
