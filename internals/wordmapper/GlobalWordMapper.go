package wordmapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"github.com/negapedia/wikiconflict/internals/structures"
	"github.com/negapedia/wikiconflict/internals/utils"
)

// GlobalWordMapper given the result dir, generate the file containing the global report about word frequency
func GlobalWordMapper(resultDir string) error {
	fileList, err := utils.FilesInDir(resultDir, "M[0-9]*")
	if err != nil {
		return err
	}
	nFile := len(fileList)

	globalWord := make(map[string]map[string]uint32)
	var totalWord uint32
	var totalPage uint32

	for i, file := range fileList {
		fmt.Printf("\rOn %d/%d", i+1, nFile)

		jsonFile, err := os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "Error happened while trying to open file:"+file)
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		jsonFile.Close()

		var page structures.PageElement

		err = json.Unmarshal(byteValue, &page)
		if err != nil {
			return errors.Wrapf(err, "Error while unmarshalling json."+file)
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
	return nil
}
