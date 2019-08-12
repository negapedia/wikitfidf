package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/structures"
	"github.com/negapedia/Wikipedia-Conflict-Analyzer/internals/utils"
)

// GlobalWordMapper given the result dir, generate the file containing the global report about word frequency
func GlobalWordMapper(resultDir string) {
	fileList := utils.FilesInDir(resultDir, "M[0-9]*")
	nFile := len(fileList)

	globalWord := make(map[string]map[string]uint32)
	var totalWord uint32
	totalWord = 0
	var totalPage uint32
	totalPage = 0

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)

		jsonFile, err := os.Open(file)
		if err != nil {
			log.Fatal("Error happened while trying to open file:", file, "Error:",err)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		var page structures.PageElement

		err = json.Unmarshal(byteValue, &page)
		if err != nil {
			log.Fatal("Error while unmarshalling json.",err)
		}

		totalPage++

		for word, freq := range page.Word {
			if _, ok := globalWord[word]; ok {
				globalWord[word]["a"] += uint32(freq) // a --> abs, i --> in
				globalWord[word]["i"]++
			} else {
				globalWord[word] = make(map[string]uint32)
				globalWord[word]["a"] = uint32(freq)
				globalWord[word]["i"] = 1
			}
			totalWord += uint32(freq)
		}
	}
	fmt.Println()

	globalWord["@Total Word"] = make(map[string]uint32)
	globalWord["@Total Word"]["tot"] = totalWord
	globalWord["@Total Page"] = make(map[string]uint32)
	globalWord["@Total Page"]["tot"] = totalPage

	utils.WriteGlobalWord(resultDir, &globalWord)
}
